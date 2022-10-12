## 极简RTSP服务器

### 简介

`Golang`实现`RTSP`协议服务器

### 功能

直播推流、拉流

### 安装

1. 下载源码

```shell
# 需安装git
git clone https://gitee.com/chentong001/rtsp-server.git
```

2. 源码编译

```shell
# 需安装golang环境
# 进入项目目录执行以下命令
go build
```

### 使用

1. 启动服务。执行`rtsp-server`二进制文件启动服务
2. 推流。通过`RTSP`协议推送视频流到地址`rtsp://localhost:8554/{channel}`

```shell
# 示例
# ffmpeg推流屏幕画面
ffmpeg -f gdigrab -i desktop -vcodec libx264 -s 1280x720 -r 30 -b 1m -rtsp_transport tcp -f rtsp rtsp://127.0.0.1:8554/live
```

3. 拉流。通过`RTSP`协议从地址`rtsp://localhost:8554/{channel}`拉取视频流

```shell
# 示例
# ffmpeg拉取推送的视频流
ffplay -rtsp_transport tcp -f rtsp rtsp://127.0.0.1:8554/live
```

### 配置

**`rtsp.json`**

```json
{
  "port": 8554,
  "frame_buffer": 60
}
```

参数解释

| 参数名          | 解释   |
|--------------|------|
| port         | 端口号  |
| frame_buffer | 缓存帧数 |

