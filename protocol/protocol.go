package protocol

import (
	"fmt"
	"log"
	"net"
	"rtsp/socket"
	"rtsp/util"
	"strconv"
	"strings"
)

// 策略模式
// 模板方法
type handleFunc func(conn net.Conn, buf []byte, cli *socket.RtspClient, server *socket.RtspServer)

// 协议字段
const (
	OPTIONS  = "OPTIONS"
	ANNOUNCE = "ANNOUNCE"
	SETUP    = "SETUP"
	RECORD   = "RECORD"
	DESCRIBE = "DESCRIBE"
	PLAY     = "PLAY"

	//SETUPv2 = "SETUPv2"
)

// 类型枚举
var protoType = []string{OPTIONS, ANNOUNCE, SETUP, RECORD, DESCRIBE, PLAY}

// 策略模式
// 方法映射
var handleMap = map[string]handleFunc{
	OPTIONS:  option,
	ANNOUNCE: announce,
	SETUP:    setup,
	RECORD:   record,
	DESCRIBE: describe,
	PLAY:     play,
}

// 解析协议正则表达式
var regMap = map[string]string{
	OPTIONS:  "^OPTIONS rtsp://[^:]+?:[\\d]+/([\\w\\W]+?) RTSP/([0-9.]+)[\\s]+CSeq:[\\s]*([0-9]+)[\\w\\W]+[\\s]+",
	ANNOUNCE: "^ANNOUNCE [\\w\\W]+?CSeq:[\\s]*([0-9]+)[\\w\\W]*[\\s]+Content-Length:[\\s]*([0-9]+)[^\\r^\\n]*?([\\w\\W]*)",
	//SETUP:    "^SETUP rtsp://[^:]+?:[\\d]+/([\\w\\W]+?) RTSP/[0-9.]+[\\s]+CSeq:[\\s]*([0-9]+)[\\w\\W]+Transport:[\\s]*([^\\n]+?)[\\s]+",
	RECORD:   "^RECORD [\\w\\W]+? RTSP/[0-9.]+[\\w\\W]+?CSeq:[\\s]*([0-9]+)",
	DESCRIBE: "^DESCRIBE rtsp://[\\w\\W]+?:[0-9]+/([\\w\\W]+?) RTSP/[0-9.]+[\\w\\W]+?CSeq:[\\s]*([0-9]+)",
	PLAY:     "^PLAY[\\w\\W]+?CSeq:[\\s]*([0-9]+)[\\w\\W]+",
	SETUP:    "^SETUP [\\w\\W]+? RTSP/[0-9.]+[\\s]+Transport:[\\s]*([^\\n]+?)[\\s]+CSeq:[\\s]*([0-9]+)[\\w\\W]+",
}

// 响应报文模板
var protoMap = map[string]string{
	OPTIONS:  "RTSP/1.0 200 OK\nCSeq: %s\nSession: %s\nPublic: DESCRIBE, SETUP, TEARDOWN, PLAY, PAUSE, OPTIONS, ANNOUNCE, RECORD\n\n",
	ANNOUNCE: "RTSP/1.0 200 OK\nCSeq: %s\nSession: %s\n\n",
	SETUP:    "RTSP/1.0 200 OK\nCSeq: %s\nSession: %s\nTransport: %s\n\n",
	RECORD:   "RTSP/1.0 200 OK\nSession: %s\nCSeq: %s\n\n",
	DESCRIBE: "RTSP/1.0 200 OK\nSession: %s\nContent-Length: %d\nCSeq: %s\n\n%s",
	PLAY:     "RTSP/1.0 200 OK\nSession: %s\nRange: npt=0.000-\nCSeq: %s\n\n",
}

// Handle 统一处理函数
func Handle(conn net.Conn, buf []byte, cli *socket.RtspClient, server *socket.RtspServer) {
	detail := detailData(buf, cli)
	if detail == 0 {
		for _, t := range protoType {
			if strings.HasPrefix(string(buf), t) {
				f, ok := handleMap[t]
				if ok {
					log.Println("recv :", string(buf))
					f(conn, buf, cli, server)
					break
				}
			}
		}
	}
}

func option(conn net.Conn, buf []byte, cli *socket.RtspClient, server *socket.RtspServer) {
	post := commonBefore(buf, OPTIONS, cli)

	if len(post) != 4 {
		return
	}

	cli.Channel = post[1]
	data := fmt.Sprintf(protoMap[OPTIONS], post[3], cli.Session)

	socket.SendTo(conn, []byte(data), cli, server)

	cli.LastRecvBuf = []byte{}
}

func announce(conn net.Conn, buf []byte, cli *socket.RtspClient, server *socket.RtspServer) {
	post := commonBefore(buf, ANNOUNCE, cli)
	if len(post) != 4 {
		return
	}
	ln, err := strconv.Atoi(post[2])
	if err != nil {
		return
	}

	if len(post[3]) <= 4 {
		cli.LastRecvBuf = buf
	}

	post[3] = post[3][4:]
	if ln > len(post[3]) {
		cli.LastRecvBuf = buf
		return
	}

	if ln < len(post[3]) {
		cli.LastRecvBuf = []byte(post[3][ln:])
	}
	cli.Sdp = post[3][0:ln]

	cli.Flag = socket.PUSH

	server.PushersLock.Lock()
	_, ok := server.Pushers[cli.Channel]
	if ok {
		server.PushersLock.Unlock()
		return
	}

	server.Pushers[cli.Channel] = cli
	server.PushersLock.Unlock()

	data := fmt.Sprintf(protoMap[ANNOUNCE], post[1], cli.Session)

	socket.SendTo(conn, []byte(data), cli, server)
	cli.LastRecvBuf = []byte{}
}

func setup(conn net.Conn, buf []byte, cli *socket.RtspClient, server *socket.RtspServer) {
	post := commonBefore(buf, SETUP, cli)
	//if post == nil || len(post) == 0 {
	//	post = commonBefore(buf, "SETUPv2", cli)
	//}

	data := fmt.Sprintf(protoMap[SETUP], post[2], cli.Session, post[1])

	socket.SendTo(conn, []byte(data), cli, server)
	cli.LastRecvBuf = []byte{}
}

func record(conn net.Conn, buf []byte, cli *socket.RtspClient, server *socket.RtspServer) {
	post := commonBefore(buf, RECORD, cli)
	if len(post) != 2 {
		return
	}

	cli.Ready = true

	data := fmt.Sprintf(protoMap[RECORD], cli.Session, post[1])
	socket.SendTo(conn, []byte(data), cli, server)

	if !cli.RecordStart {
		cli.RecordStart = true
		go pushData(cli, server)
	}

	cli.LastRecvBuf = []byte{}
	log.Println("client", cli.Addr, "start push", cli.Channel)
}

func describe(conn net.Conn, buf []byte, cli *socket.RtspClient, server *socket.RtspServer) {
	post := commonBefore(buf, DESCRIBE, cli)

	if len(post) != 3 {
		return
	}

	cli.Flag = socket.PLAY

	server.PushersLock.RLock()
	v, ok := server.Pushers[post[1]]
	server.PushersLock.RUnlock()

	if ok {
		data := fmt.Sprintf(protoMap[DESCRIBE], cli.Session, len(v.Sdp), post[2], v.Sdp)
		socket.SendTo(conn, []byte(data), cli, server)
	}

	cli.LastRecvBuf = []byte{}
}

func play(conn net.Conn, buf []byte, cli *socket.RtspClient, server *socket.RtspServer) {
	post := commonBefore(buf, PLAY, cli)
	if len(post) != 2 {
		return
	}

	cli.Ready = true

	server.PushersLock.Lock()
	v, ok := server.Pushers[cli.Channel]
	if ok {
		v.PlayersLock.Lock()
		v.Players[cli.Addr] = cli
		v.PlayersLock.Unlock()
	}
	server.PushersLock.Unlock()

	data := fmt.Sprintf(protoMap[PLAY], cli.Session, post[1])
	socket.SendTo(conn, []byte(data), cli, server)

	log.Println("client", cli.Addr, "will open this channel", cli.Channel)

	cli.LastRecvBuf = []byte{}
}

func commonBefore(buf []byte, doType string, cli *socket.RtspClient) []string {
	cli.LastRecvBuf = buf
	return util.RegTo(string(buf), regMap[doType])
}

func pushData(cli *socket.RtspClient, server *socket.RtspServer) {
	// 缓存帧数 可以给拉流端迅速响应
	var cache [][]byte

	// 循环处理数据
	for {
		// 从channel中获取数据
		d := <-cli.FromChan()

		// 读取到一个字符 代表关闭
		if len(d) == 1 {
			break
		}

		// 如果缓存帧数大于0 需要缓存
		if server.FrameBuffer > 0 {
			// 如果临时数组大小和缓存区大小相等 将头部缓存删去
			if len(cache) == server.FrameBuffer {
				cache = append(cache[:0], cache[1:]...)
			}
			// 临时数组存储缓存信息
			cache = append(cache, d)
		}

		// 表示要剔除的拉流端
		removes := []string{}
		rmflag := false

		cli.PlayersLock.RLock()
		// 遍历所有的拉流端
		for _, v := range cli.Players {
			// 如果有缓存且并未向拉流端发送缓存 则把缓存发送给对方
			if server.FrameBuffer > 0 && cli.HasSend == false {
				cli.HasSend = true
				// 将缓存中的数据发送给对方
				for _, data := range cache {
					_, e := v.Conn.Write(data)
					if e != nil {
						removes = append(removes, v.Addr)
						rmflag = true
						break
					}
				}
			}

			_, err := v.Conn.Write(d)

			// 发生异常则将该拉流端剔除
			if err != nil {
				removes = append(removes, v.Addr)
				rmflag = true
			}
		}
		cli.PlayersLock.RUnlock()

		if rmflag {
			cli.PlayersLock.Lock()
			for _, rm := range removes {
				delete(cli.Players, rm)
			}
			cli.PlayersLock.Unlock()
		}
	}
}

func detailData(buf []byte, cli *socket.RtspClient) int {
	if cli.Ready == false || cli.Flag != socket.PUSH {
		return 0
	}
	if len(buf) < 10 {
		cli.LastRecvBuf = buf
		return -1
	}

	cli.LastRecvBuf = []byte{}
	msgbytes0 := buf
	sendbytes := []byte{}
	for {
		if len(msgbytes0) < 4 {
			cli.LastRecvBuf = msgbytes0
			break
		}

		ln := util.BytesToInt16(msgbytes0[2:4])
		bl := util.BytesToInt8(msgbytes0[1:2])

		if bl >= 200 && bl <= 207 {
			// RTCP包 推送端推送会话质量信息
			log.Println(bl)
		}

		if ln < 0 || ln > 20000 {
			cli.LastRecvBuf = []byte{}
			break
		}
		if ln+4 > len(msgbytes0) {
			cli.LastRecvBuf = msgbytes0
			break
		}
		if ln+4 == len(msgbytes0) {
			sendbytes = util.BytesCombine(sendbytes, msgbytes0)
			cli.LastRecvBuf = []byte{}
			break
		}
		if ln+4 < len(msgbytes0) {
			sendbytes = util.BytesCombine(sendbytes, msgbytes0[0:4+ln])
			msgbytes0 = msgbytes0[4+ln:]
		}
	}

	if len(sendbytes) > 0 {
		cli.ToChan() <- sendbytes
	}

	return 1
}
