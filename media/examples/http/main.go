package main

import (
	"fmt"
	"time"

	"github.com/wwqdrh/gokit/media/hls"
)

func main() {
	manager := hls.NewHLSManager()

	// 测试文件路径
	inputFile := "./video/sample.mp4"
	streamName := "test_sample"

	// 启动服务器
	err := manager.StartServer(":8080", false)
	if err != nil {
		fmt.Println(err.Error())
		// t.Errorf("StartServer() failed: %v", err)
		return
	}
	defer manager.StopServer()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 启动流
	err = manager.StartStream(inputFile, streamName)
	if err != nil {
		fmt.Println(err.Error())
		// t.Errorf("StartStream() failed: %v", err)
		return
	}

	// 等待流启动
	time.Sleep(5 * time.Second)

	// 检查流是否存在
	stream, err := manager.GetStreamInfo(streamName)
	if err != nil {
		fmt.Println(err.Error())
		// t.Errorf("GetStreamInfo() failed: %v", err)
		return
	}

	if stream == nil {
		fmt.Println(err.Error())
		// t.Errorf("GetStreamInfo() returned nil")
		return
	}

	// if stream.status != "running" {
	// 	return
	// 	// t.Errorf("Stream status is not running: %s", stream.status)
	// }

	// 停止流
	// err = manager.StopStream(streamName)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	// t.Errorf("StopStream() failed: %v", err)
	// 	return
	// }

	// // 检查流是否已停止
	// _, err = manager.GetStreamInfo(streamName)
	// if err == nil {
	// 	fmt.Println(err.Error())
	// 	return
	// 	// t.Errorf("GetStreamInfo() should have failed after stopping the stream")
	// }
}
