package main

import (
	"rtsp-server/conf"
	"rtsp-server/protocol"
	rtsp "rtsp-server/socket"
)

func main() {
	config := conf.NewConfig("rtsp.json")
	server := rtsp.NewRtspServer("rtsp server", config.Port, config.FrameBuffer)
	server.Run(protocol.Handle)
}
