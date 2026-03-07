package hls

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
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
}

// streamInfo 流信息
type streamInfo struct {
	// 输入文件路径
	inputFile string
	// 流名称
	streamName string
	// ffmpeg 进程
	ffmpegProcess *exec.Cmd
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

	return &HLSManager{
		streams:      make(map[string]*streamInfo),
		hlsOutputDir: hlsOutputDir,
	}
}

// StartServer 启动 HLS 服务器
func (hm *HLSManager) StartServer(address string, daemon bool) error {
	// 启动 HTTP 服务器提供播放页面和 HLS 流
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hm.servePlaybackPage(w, r)
	})

	// 提供 HLS 流访问
	http.Handle("/hls/", http.StripPrefix("/hls/", http.FileServer(http.Dir(hm.hlsOutputDir))))

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

	// 创建流的输出目录
	streamOutputDir := filepath.Join(hm.hlsOutputDir, streamName)
	if err := os.MkdirAll(streamOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create stream output directory: %v", err)
	}

	// 构建 ffmpeg 命令，直接生成 HLS 流
	cmd := exec.Command("ffmpeg",
		"-re",
		"-i", inputFile,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-hls_time", "10",
		"-hls_list_size", "6",
		"-hls_wrap", "10",
		"-start_number", "1",
		filepath.Join(streamOutputDir, "index.m3u8"),
	)

	// 启动 ffmpeg 进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	// 保存流信息
	hm.streams[streamName] = &streamInfo{
		inputFile:     inputFile,
		streamName:    streamName,
		ffmpegProcess: cmd,
		status:        "running",
	}

	log.Printf("Started HLS stream for %s as %s", inputFile, streamName)
	return nil
}

// StopStream 停止一个 HLS 流
func (hm *HLSManager) StopStream(streamName string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// 检查流是否存在
	stream, exists := hm.streams[streamName]
	if !exists {
		return fmt.Errorf("stream %s does not exist", streamName)
	}

	// 停止 ffmpeg 进程
	if stream.ffmpegProcess != nil {
		if err := stream.ffmpegProcess.Process.Kill(); err != nil {
			log.Printf("Error killing ffmpeg process: %v", err)
		}
		stream.ffmpegProcess.Wait()
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
	return fmt.Sprintf("http://localhost:8080/hls/%s/index.m3u8", streamName)
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
