package main

import (
	"rtsp/protocol"
	rtsp "rtsp/socket"
)

func main() {
	server := rtsp.NewRtspServer("rtsp server", 8080, 60)
	server.Run(protocol.Handle)
}
