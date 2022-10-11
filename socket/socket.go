package socket

import (
	"fmt"
	"log"
	"net"
	"rtsp/util"
	"strconv"
	"sync"
	"time"
)

const (
	PUSH = "Push"
	PLAY = "Play"
)

type handleFunc func(conn net.Conn, buf []byte, cli *RtspClient, server *RtspServer)

type RtspClient struct {
	Method      string
	Conn        net.Conn
	close       bool
	Addr        string
	LastRecvBuf []byte
	ConnectTime int64 //连接时间，单位秒
	DataNum     int64 //数据量

	//下面的变量都是RTSP的推流端和播放端用到的
	Session string //播放端和推流端 session信息
	Ready   bool   //播放端和推流端 RTSP通信是否已经完成
	Flag    string //播放端和推流端 客户端标志，Push或Play
	Channel string //播放端和推流端 对应的rtsp通道

	HasSend bool // 是否已经向拉流端推送服务器缓存的RTP信息

	Players     map[string]*RtspClient // 拉流端信息
	PlayersLock sync.RWMutex           // 读写锁
	RecordStart bool                   // 当推流端开始推流之后 会启动一个协程 用来给该频道的播放端推送RTP消息
	Sdp         string                 // 推流端 媒体SDP信息
	pushChan    chan []byte            // 推流端 每接受到推流端的一个RTP消息
}

func (cli *RtspClient) FromChan() <-chan []byte {
	return cli.pushChan
}

func (cli *RtspClient) ToChan() chan<- []byte {
	return cli.pushChan
}

type RtspServer struct {
	Port        int
	name        string
	Pushers     map[string]*RtspClient
	PushersLock sync.RWMutex
	FrameBuffer int
}

func NewRtspServer(name string, port int, frameBuffer int) *RtspServer {
	if frameBuffer < 0 {
		frameBuffer = 0
	}
	return &RtspServer{
		name:        name,
		Port:        port,
		FrameBuffer: frameBuffer,
		Pushers:     make(map[string]*RtspClient),
	}
}

func (server *RtspServer) handleConn(conn net.Conn, f handleFunc) {
	if conn == nil {
		log.Panic("conn is unresolved")
	}

	cli := &RtspClient{
		Conn:        conn,
		Ready:       false,
		Addr:        conn.RemoteAddr().String(),
		close:       false,
		HasSend:     false,
		Method:      "",
		Session:     util.RandomString(9),
		pushChan:    make(chan []byte, 1000),
		Players:     make(map[string]*RtspClient),
		RecordStart: false,
		ConnectTime: time.Now().Unix(),
		DataNum:     0,
	}

	// buffer
	buf := make([]byte, 2048)

	// 读取数据流
	for {
		if cli.close {
			break
		}

		count, err := conn.Read(buf)
		if count == 0 || err != nil {
			server.close(conn, cli)
			break
		}

		// count 数据帧大小
		cli.DataNum = cli.DataNum + int64(count)
		all := util.BytesCombine(cli.LastRecvBuf, buf[0:count])
		cli.LastRecvBuf = []byte{}

		f(conn, all, cli, server)
	}
}

func (server *RtspServer) close(conn net.Conn, cli *RtspClient) {
	_ = conn.Close()
	cli.close = true

	if cli.Flag == PLAY {
		server.PushersLock.Lock()
		for k, v := range server.Pushers {
			if k == cli.Channel {
				v.PlayersLock.Lock()
				_, ok := v.Players[cli.Addr]
				if ok {
					delete(v.Players, cli.Addr)
				}
				v.PlayersLock.Unlock()
			}
		}
		server.PushersLock.Unlock()
		log.Println("client", cli.Addr, "stop pull", cli.Channel)
	}

	if cli.Flag == PUSH {
		cli.ToChan() <- []byte("0")
		server.PushersLock.Lock()
		_, ok := server.Pushers[cli.Channel]
		if ok {
			delete(server.Pushers, cli.Channel)
		}
		server.PushersLock.Unlock()
		log.Println("client", cli.Addr, "stop push", cli.Channel)
	}
}

func (server *RtspServer) Run(f handleFunc) {
	cServer, err := net.Listen("tcp", ":"+strconv.Itoa(server.Port))
	if err != nil {
		log.Println("fail to start rtsp server in", server.Port, err)
		return
	}
	log.Println("started rtsp server in", cServer.Addr().String())
	for {
		// 接收来自 client 的连接,会阻塞
		conn, err := cServer.Accept()
		if err != nil {
			log.Println("connection error", err.Error())
			continue
		}

		go server.handleConn(conn, f)
	}
}

func SendTo(coon net.Conn, data []byte, cli *RtspClient, server *RtspServer) {
	fmt.Println("send to :", string(data))
	_, err := coon.Write(data)
	if err != nil {
		server.close(coon, cli)
	}
}
