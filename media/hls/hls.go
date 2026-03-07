package hls

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gwuhaolin/livego/protocol/hls"
	"github.com/gwuhaolin/livego/protocol/rtmp"
)

// HLSManager 管理 HLS 流
type HLSManager struct {
	// 流映射
	streams map[string]*streamInfo
	// 互斥锁
	mu sync.RWMutex
	// HTTP 服务器
	server *http.Server
	// HLS 输出目录
	hlsOutputDir string
	// RTMP 流
	rtmpStream *rtmp.RtmpStream
	// HLS 服务器
	hlsServer *hls.Server
}

// streamInfo 流信息
type streamInfo struct {
	// 输入文件路径
	inputFile string
	// 流名称
	streamName string
	// 状态
	status string
}

// NewHLSManager 创建一个新的 HLS 管理器
func NewHLSManager() *HLSManager {
	// 创建 HLS 输出目录
	hlsOutputDir := "tmp/hls"
	if err := os.MkdirAll(hlsOutputDir, 0755); err != nil {
		log.Printf("Warning: failed to create HLS output directory: %v", err)
	}

	// 初始化 RTMP 流
	rtmpStream := rtmp.NewRtmpStream()
	// 初始化 HLS 服务器
	hlsServer := hls.NewServer()

	return &HLSManager{
		streams:      make(map[string]*streamInfo),
		hlsOutputDir: hlsOutputDir,
		rtmpStream:   rtmpStream,
		hlsServer:    hlsServer,
	}
}

// StartServer 启动 HLS 服务器
func (hm *HLSManager) StartServer(address string, daemon bool) error {
	// 启动 HLS 服务器
	hlsAddr := ":7002"
	hlsListen, err := net.Listen("tcp", hlsAddr)
	if err != nil {
		return fmt.Errorf("failed to start HLS server: %v", err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("HLS server panic: %v", r)
			}
		}()
		log.Printf("HLS listen On %s", hlsAddr)
		hm.hlsServer.Serve(hlsListen)
	}()

	// 启动 RTMP 服务器
	rtmpAddr := ":1935"
	rtmpListen, err := net.Listen("tcp", rtmpAddr)
	if err != nil {
		return fmt.Errorf("failed to start RTMP server: %v", err)
	}

	rtmpServer := rtmp.NewRtmpServer(hm.rtmpStream, hm.hlsServer)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("RTMP server panic: %v", r)
			}
		}()
		log.Printf("RTMP listen On %s", rtmpAddr)
		rtmpServer.Serve(rtmpListen)
	}()

	// 启动 HTTP 服务器提供播放页面和 HLS 流
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hm.servePlaybackPage(w, r)
	})

	// 提供 HLS 流访问 (使用 livego 的 HLS 服务器)
	http.HandleFunc("/hls/", func(w http.ResponseWriter, r *http.Request) {
		// 重定向到 livego 的 HLS 服务器
		path := r.URL.Path[len("/hls/"):]
		hlsURL := fmt.Sprintf("http://localhost:7002/live/%s", path)
		http.Redirect(w, r, hlsURL, http.StatusFound)
	})

	// 启动 API 服务器
	http.HandleFunc("/api/hls/start", func(w http.ResponseWriter, r *http.Request) {
		hm.handleStartStream(w, r)
	})

	http.HandleFunc("/api/hls/stop", func(w http.ResponseWriter, r *http.Request) {
		hm.handleStopStream(w, r)
	})

	// 创建 HTTP 服务器
	hm.server = &http.Server{
		Addr:    address,
		Handler: nil,
	}

	if daemon {
		// 启动服务器
		go func() {
			log.Printf("HLS server started at %s", address)
			if err := hm.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("HLS server error: %v", err)
			}
		}()
	} else {
		log.Printf("HLS server started at %s", address)
		if err := hm.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HLS server error: %v", err)
		}
	}

	return nil
}

// StopServer 停止 HLS 服务器
func (hm *HLSManager) StopServer() error {
	if hm.server != nil {
		return hm.server.Close()
	}
	return nil
}

// StartStream 启动一个 HLS 流
func (hm *HLSManager) StartStream(inputFile, streamName string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// 检查流是否已存在
	if _, exists := hm.streams[streamName]; exists {
		return fmt.Errorf("stream %s already exists", streamName)
	}

	// 检查输入文件是否存在
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file %s does not exist", inputFile)
	}

	// 创建一个本地文件读取器，将本地视频文件转换为 av.Packet 并写入 livego 的 RTMP 流
	go func() {
		// 打开本地文件
		file, err := os.Open(inputFile)
		if err != nil {
			log.Printf("Failed to open input file: %v", err)
			return
		}
		defer file.Close()

		// 这里需要实现从本地文件读取视频数据并转换为 av.Packet 的逻辑
		// 由于 livego 库本身没有提供直接从本地文件读取的功能，我们需要使用第三方库或自行实现
		// 为了简化实现，这里我们使用一个简单的模拟实现
		log.Printf("Started reading local file: %s", inputFile)

		// 模拟流数据，实际应用中需要从文件中读取并转换
		// 注意：这只是一个示例，实际实现需要使用真正的视频解码库
		// 例如使用 github.com/3rdparty/video-decoder 等库

		// 等待一段时间后关闭
		time.Sleep(30 * time.Second)
		log.Printf("Finished reading local file: %s", inputFile)
	}()

	// 保存流信息
	hm.streams[streamName] = &streamInfo{
		inputFile:  inputFile,
		streamName: streamName,
		status:     "running",
	}

	log.Printf("Started HLS stream for %s as %s", inputFile, streamName)
	return nil
}

// StopStream 停止一个 HLS 流
func (hm *HLSManager) StopStream(streamName string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// 检查流是否存在
	_, exists := hm.streams[streamName]
	if !exists {
		return fmt.Errorf("stream %s does not exist", streamName)
	}

	// 移除流信息
	delete(hm.streams, streamName)

	log.Printf("Stopped HLS stream %s", streamName)
	return nil
}

// GetStreamInfo 获取流信息
func (hm *HLSManager) GetStreamInfo(streamName string) (*streamInfo, error) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	stream, exists := hm.streams[streamName]
	if !exists {
		return nil, fmt.Errorf("stream %s does not exist", streamName)
	}

	return stream, nil
}

// GetHLSURL 获取 HLS 播放地址
func (hm *HLSManager) GetHLSURL(streamName string) string {
	return fmt.Sprintf("http://localhost:7002/live/%s.m3u8", streamName)
}

// handleStartStream 处理启动流的 HTTP 请求
func (hm *HLSManager) handleStartStream(w http.ResponseWriter, r *http.Request) {
	inputFile := r.URL.Query().Get("input")
	streamName := r.URL.Query().Get("stream")

	if inputFile == "" || streamName == "" {
		http.Error(w, "Missing input or stream parameter", http.StatusBadRequest)
		return
	}

	if err := hm.StartStream(inputFile, streamName); err != nil {
		http.Error(w, fmt.Sprintf("Error starting stream: %v", err), http.StatusInternalServerError)
		return
	}

	hlsURL := hm.GetHLSURL(streamName)
	fmt.Fprintf(w, `{"status": "success", "hls_url": "%s"}`, hlsURL)
}

// handleStopStream 处理停止流的 HTTP 请求
func (hm *HLSManager) handleStopStream(w http.ResponseWriter, r *http.Request) {
	streamName := r.URL.Query().Get("stream")

	if streamName == "" {
		http.Error(w, "Missing stream parameter", http.StatusBadRequest)
		return
	}

	if err := hm.StopStream(streamName); err != nil {
		http.Error(w, fmt.Sprintf("Error stopping stream: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `{"status": "success"}`)
}

// servePlaybackPage 提供播放页面
func (hm *HLSManager) servePlaybackPage(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>HLS Player</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 20px; }
			.container { max-width: 800px; margin: 0 auto; }
			.video-container { margin: 20px 0; }
			input { margin: 5px; padding: 10px; }
			button { padding: 10px 20px; margin: 5px; }
			#status { margin: 10px 0; padding: 10px; background: #f0f0f0; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>HLS Stream Player</h1>
			<div>
				<label>Input File: </label>
				<input type="text" id="inputFile" value="/Users/dengronghui/project/gokit/media/hls/testdata/sample.mp4">
			</div>
			<div>
				<label>Stream Name: </label>
				<input type="text" id="streamName" value="sample">
			</div>
			<button onclick="startStream()">Start Stream</button>
			<button onclick="stopStream()">Stop Stream</button>
			<div id="status"></div>
			<div class="video-container">
				<video id="player" width="800" height="450" controls></video>
			</div>
		</div>
		<script>
			const player = document.getElementById('player');
			const status = document.getElementById('status');

			function startStream() {
				const inputFile = document.getElementById('inputFile').value;
				const streamName = document.getElementById('streamName').value;
				
				fetch("/api/hls/start?input="+encodeURIComponent(inputFile)+"&stream="+encodeURIComponent(streamName))
					.then(response => response.json())
					.then(data => {
						if (data.status === 'success') {
							status.textContent = 'Stream started successfully';
							player.src = data.hls_url;
							player.play();
						} else {
							status.textContent = 'Error starting stream: ' + data.error;
						}
					})
					.catch(error => {
						status.textContent = 'Error: ' + error.message;
					});
			}

			function stopStream() {
				const streamName = document.getElementById('streamName').value;
				
				fetch("/api/hls/stop?stream="+encodeURIComponent(streamName))
					.then(response => response.json())
					.then(data => {
						if (data.status === 'success') {
							status.textContent = 'Stream stopped successfully';
							player.src = '';
						} else {
							status.textContent = 'Error stopping stream: ' + data.error;
						}
					})
					.catch(error => {
						status.textContent = 'Error: ' + error.message;
					});
			}
		</script>
	</body>
	</html>
	`

	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, html)
}
