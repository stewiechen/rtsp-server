package main

import (
	"rtsp/conf"
	"rtsp/protocol"
	rtsp "rtsp/socket"
)

func main() {
	config := conf.NewConfig("rtsp.json")
	server := rtsp.NewRtspServer("rtsp server", config.Port, config.FrameBuffer)
	server.Run(protocol.Handle)
}
