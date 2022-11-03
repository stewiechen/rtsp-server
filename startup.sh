rm -f rtsp-server.log
rm -f rtsp-server.pid
touch rtsp-server.log
touch rtsp-server.pid
go build
nohup ./rtsp-server > rtsp-server.log 2>&1 & echo $! > rtsp-server.pid