package main

import (
	"fmt"
	"os"

	"github.com/wwqdrh/gokit/media/http"
)

func main() {
	// 获取视频目录参数
	videoDir := "./video"
	if len(os.Args) > 1 {
		videoDir = os.Args[1]
	}

	// 创建视频流服务器
	server := http.NewVideoServer(videoDir)

	// 启动服务器
	addr := ":8080"
	if len(os.Args) > 2 {
		addr = os.Args[2]
	}

	fmt.Println("=====================================")
	fmt.Println("视频流服务器启动中...")
	fmt.Println("=====================================")

	// 启动HTTP服务
	err := server.Start(addr)
	if err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}
