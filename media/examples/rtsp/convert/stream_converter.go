package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wwqdrh/gokit/media/rtsp/stream"
)

func main() {
	// 1. 配置RTSP流
	rtspConfig := stream.RTSPConfig{
		URL:           "rtsp://192.168.1.100:554/stream1",
		Username:      "admin",
		Password:      "password",
		Transport:     "RTP/AVP",
		BufferSize:    1024,
		RetryInterval: 5 * time.Second,
		MaxRetries:    3,
	}

	// 2. 创建RTSP流
	rtspStream := stream.NewRTSPStream(rtspConfig)
	if err := rtspStream.Start(); err != nil {
		log.Fatalf("Failed to start RTSP stream: %v", err)
	}
	defer rtspStream.Stop()

	fmt.Println("RTSP stream started successfully!")
	fmt.Printf("Stream info: %+v\n", rtspStream.GetStreamInfo())

	// 3. 配置可选处理
	processorOptions := stream.StreamProcessorOptions{
		EnableScaling:     true,
		TargetResolution:  "1280x720",
		EnableWatermark:   false,
		WatermarkPath:     "/path/to/watermark.jpg",
		WatermarkPosition: "bottom-right",
		EnableFiltering:   false,
		FilterType:        "none",
	}

	// 4. 创建流处理器
	processor := stream.NewStreamProcessor(rtspStream, processorOptions)
	if err := processor.Start(); err != nil {
		log.Fatalf("Failed to start stream processor: %v", err)
	}
	defer processor.Stop()

	fmt.Println("Stream processor started successfully!")

	// 5. 创建转码器
	transcoder := stream.NewStreamTranscoder()
	if err := transcoder.Start(); err != nil {
		log.Fatalf("Failed to start transcoder: %v", err)
	}
	defer transcoder.Stop()

	// 6. 配置输出格式
	outputFormats := []stream.StreamType{
		stream.StreamTypeFLV,
		stream.StreamTypeHLS,
		stream.StreamTypeWebRTC,
	}

	// 7. 开始转码
	if err := transcoder.Transcode(rtspStream, outputFormats); err != nil {
		log.Fatalf("Failed to start transcoding: %v", err)
	}

	fmt.Println("Transcoding started successfully!")
	fmt.Println("Output streams:")
	for format, url := range transcoder.GetOutputURLs() {
		fmt.Printf("  %s: %s\n", format, url)
	}

	// 8. 创建分发器
	distributor := stream.NewStreamDistributor("0.0.0.0:8080")
	if err := distributor.Start(); err != nil {
		log.Fatalf("Failed to start distributor: %v", err)
	}
	defer distributor.Stop()

	// 9. 添加流到分发器
	for _, format := range outputFormats {
		if outputStream := transcoder.GetOutputStream(format); outputStream != nil {
			if stream := outputStream; stream != nil {
				if err := distributor.AddStream(stream); err != nil {
					log.Printf("Failed to add stream to distributor: %v", err)
				}
			}
		}
	}

	fmt.Println("\nStream distributor started successfully!")
	fmt.Println("Available endpoints:")
	fmt.Println("  FLV stream: http://localhost:8080/stream/flv")
	fmt.Println("  HLS stream: http://localhost:8080/hls/playlist.m3u8")
	fmt.Println("  WebRTC stream: http://localhost:8080/webrtc/")
	fmt.Println("  Status: http://localhost:8080/status")

	// 10. 等待中断信号
	fmt.Println("\nStreaming service is running...")
	fmt.Println("Press Ctrl+C to stop the service")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nStopping streaming service...")

	// 11. 清理资源
	fmt.Println("Streaming service stopped successfully!")
}

// Example with custom transcoding options
func exampleCustomOptions() {
	// 示例：使用自定义选项
	rtspConfig := stream.RTSPConfig{
		URL: "rtsp://192.168.1.100:554/camera",
	}

	rtspStream := stream.NewRTSPStream(rtspConfig)
	defer rtspStream.Stop()

	// 启用水印和缩放
	options := stream.StreamProcessorOptions{
		EnableScaling:     true,
		TargetResolution:  "854x480", // 720p
		EnableWatermark:   true,
		WatermarkPath:     "./watermark.jpg",
		WatermarkPosition: "top-left",
		EnableFiltering:   true,
		FilterType:        "denoise",
	}

	processor := stream.NewStreamProcessor(rtspStream, options)
	defer processor.Stop()

	fmt.Println("Custom stream processing started!")
}
