package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// VideoServer 视频流服务器
type VideoServer struct {
	VideoDir string
}

// NewVideoServer 创建新的视频流服务器
func NewVideoServer(videoDir string) *VideoServer {
	if videoDir == "" {
		videoDir = "./videos"
	}
	// 确保视频目录存在
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		panic(fmt.Sprintf("创建视频目录失败: %v", err))
	}
	return &VideoServer{
		VideoDir: videoDir,
	}
}

// ServeHTTP 实现http.Handler接口
func (s *VideoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取请求的视频路径
	videoPath := strings.TrimPrefix(r.URL.Path, "/")
	if videoPath == "" {
		s.serveIndex(w, r)
		return
	}

	// 处理前端示例页面请求
	if strings.HasPrefix(videoPath, "examples/") {
		s.serveExamplePage(w, r, videoPath)
		return
	}

	// 构建完整的视频文件路径
	// 移除可能的video/前缀，避免路径重复
	cleanPath := strings.TrimPrefix(videoPath, "video/")
	fullPath := filepath.Join(s.VideoDir, cleanPath)

	// 检查文件是否存在
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "视频文件不存在", http.StatusNotFound)
		} else {
			http.Error(w, "服务器内部错误", http.StatusInternalServerError)
		}
		return
	}

	// 检查是否为目录
	if fileInfo.IsDir() {
		s.serveDirectory(w, r, fullPath, videoPath)
		return
	}

	// 处理视频文件请求
	s.serveVideoFile(w, r, fullPath, fileInfo)
}

// serveIndex 提供索引页面
func (s *VideoServer) serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>视频流服务</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 20px; }
			h1 { color: #333; }
			ul { list-style-type: none; padding: 0; }
			li { margin: 10px 0; }
			a { color: #0066cc; text-decoration: none; }
			a:hover { text-decoration: underline; }
		</style>
	</head>
	<body>
		<h1>视频流服务</h1>
		<p>当前视频目录: %s</p>
		<ul>
			<li><a href="/examples/">前端播放示例</a></li>
		</ul>
	</body>
	</html>
	`, s.VideoDir)
}

// serveDirectory 提供目录列表
func (s *VideoServer) serveDirectory(w http.ResponseWriter, r *http.Request, dirPath, urlPath string) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		http.Error(w, "读取目录失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>目录: %s</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 20px; }
			h1 { color: #333; }
			ul { list-style-type: none; padding: 0; }
			li { margin: 10px 0; }
			a { color: #0066cc; text-decoration: none; }
			a:hover { text-decoration: underline; }
		</style>
	</head>
	<body>
		<h1>目录: %s</h1>
		<ul>
			<li><a href="../">返回上一级</a></li>
		`, urlPath, urlPath)

	for _, entry := range entries {
		entryPath := filepath.Join(urlPath, entry.Name())
		if entry.IsDir() {
			fmt.Fprintf(w, "<li><a href=\"%s/\">%s/</a></li>\n", entryPath, entry.Name())
		} else {
			fmt.Fprintf(w, "<li><a href=\"%s\">%s</a></li>\n", entryPath, entry.Name())
		}
	}

	fmt.Fprintf(w, `
		</ul>
	</body>
	</html>
	`)
}

// serveVideoFile 处理视频文件请求，支持范围请求
func (s *VideoServer) serveVideoFile(w http.ResponseWriter, r *http.Request, filePath string, fileInfo os.FileInfo) {
	// 打开视频文件
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "打开视频文件失败", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 获取文件大小
	fileSize := fileInfo.Size()

	// 获取文件的MIME类型
	contentType := getContentType(filePath)

	// 处理范围请求
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		// 完整文件请求
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusOK)

		// 发送文件内容
		buf := make([]byte, 65536) // 64KB 缓冲区
		for {
			n, err := file.Read(buf)
			if n > 0 {
				if _, err := w.Write(buf[:n]); err != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
		return
	}

	// 解析范围请求
	var start, end int64
	// 更健壮的Range头解析方式
	rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		http.Error(w, "无效的范围请求", http.StatusBadRequest)
		return
	}

	// 解析起始位置
	if parts[0] != "" {
		if _, err := fmt.Sscanf(parts[0], "%d", &start); err != nil || start < 0 {
			http.Error(w, "无效的范围请求", http.StatusBadRequest)
			return
		}
	}

	// 解析结束位置
	if parts[1] != "" {
		if _, err := fmt.Sscanf(parts[1], "%d", &end); err != nil {
			http.Error(w, "无效的范围请求", http.StatusBadRequest)
			return
		}
	} else {
		// 如果结束位置为空，设置为文件末尾
		end = fileSize - 1
	}

	// 设置默认结束位置
	if end >= fileSize {
		end = fileSize - 1
	}

	// 检查范围有效性
	if start > end {
		http.Error(w, "无效的范围请求", http.StatusBadRequest)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusPartialContent)

	// 定位到请求的起始位置
	if _, err := file.Seek(start, 0); err != nil {
		http.Error(w, "读取文件失败", http.StatusInternalServerError)
		return
	}

	// 发送请求范围的内容
	remaining := end - start + 1
	buf := make([]byte, 65536) // 64KB 缓冲区
	for remaining > 0 {
		chunkSize := int64(len(buf))
		if chunkSize > remaining {
			chunkSize = remaining
		}

		n, err := file.Read(buf[:chunkSize])
		if n > 0 {
			if _, err := w.Write(buf[:n]); err != nil {
				break
			}
			remaining -= int64(n)
		}
		if err != nil {
			break
		}
	}
}

// serveExamplePage 提供前端示例页面
func (s *VideoServer) serveExamplePage(w http.ResponseWriter, r *http.Request, path string) {
	// 构建示例页面文件路径
	// 尝试从多个可能的位置查找examples目录
	relativePaths := []string{
		filepath.Join(".", "examples", strings.TrimPrefix(path, "examples/")),
		filepath.Join("..", "examples", strings.TrimPrefix(path, "examples/")),
		filepath.Join("..", "..", "http", "examples", strings.TrimPrefix(path, "examples/")),
		filepath.Join("..", "..", "examples", strings.TrimPrefix(path, "examples/")),
		filepath.Join("..", "..", "http", "examples", "player.html"),
	}

	var fullPath string
	var found bool

	// 如果路径以/结尾，默认返回player.html
	for _, relPath := range relativePaths {
		checkPath := relPath
		if strings.HasSuffix(path, "/") && !strings.HasSuffix(checkPath, "player.html") {
			checkPath = filepath.Join(relPath, "player.html")
		}

		// 检查文件是否存在
		if _, err := os.Stat(checkPath); err == nil {
			fullPath = checkPath
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "示例页面不存在", http.StatusNotFound)
		return
	}

	// 读取并返回文件内容
	content, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "读取示例页面失败", http.StatusInternalServerError)
		return
	}

	// 设置内容类型
	if strings.HasSuffix(fullPath, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else if strings.HasSuffix(fullPath, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasSuffix(fullPath, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}

	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// getContentType 根据文件扩展名获取MIME类型
func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".ogg":
		return "video/ogg"
	case ".avi":
		return "video/avi"
	case ".mov":
		return "video/quicktime"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".flv":
		return "video/x-flv"
	case ".mkv":
		return "video/x-matroska"
	default:
		return "application/octet-stream"
	}
}

// Start 启动视频流服务器
func (s *VideoServer) Start(addr string) error {
	fmt.Printf("视频流服务器启动在 %s\n", addr)
	fmt.Printf("视频目录: %s\n", s.VideoDir)
	fmt.Printf("访问 http://%s 查看视频列表\n", addr)
	fmt.Printf("访问 http://%s/examples/ 查看前端播放示例\n", addr)
	return http.ListenAndServe(addr, s)
}
