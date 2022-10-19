rem 推送摄像头画面
ffmpeg -f dshow -i video="Integrated Camera" -vcodec libx264 -preset:v ultrafast -tune:v zerolatency -rtsp_transport tcp -f rtsp rtsp://127.0.0.1:8554/live

rem 推送桌面
ffmpeg -f gdigrab -i desktop -vcodec libx264 -s 1280x720 -r 30 -b 1m -rtsp_transport tcp -f rtsp rtsp://127.0.0.1:8554/live

rem 推流本地视频文件
ffmpeg -re -stream_loop -1 -i xxx.mp4 -vcodec libx264 -s 1280x720 -rtsp_transport tcp -f rtsp rtsp://39.101.140.244:8554/live