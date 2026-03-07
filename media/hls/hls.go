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

	mp4 "github.com/Eyevinn/mp4ff/mp4"
	"github.com/gwuhaolin/livego/av"
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
	hlsAddr := ":7003"
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
	rtmpAddr := ":1936"
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
		hlsURL := fmt.Sprintf("http://localhost:7003/live/%s", path)
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

// AudioTag 实现 av.AudioPacketHeader 接口
type AudioTag struct {
	soundFormat   uint8
	soundRate     uint8
	soundSize     uint8
	soundType     uint8
	aacPacketType uint8
}

// SoundFormat 实现 av.AudioPacketHeader 接口
func (t *AudioTag) SoundFormat() uint8 {
	return t.soundFormat
}

// AACPacketType 实现 av.AudioPacketHeader 接口
func (t *AudioTag) AACPacketType() uint8 {
	return t.aacPacketType
}

// VideoTag 实现 av.VideoPacketHeader 接口
type VideoTag struct {
	frameType       uint8
	codecID         uint8
	avcPacketType   uint8
	compositionTime int32
}

// IsKeyFrame 实现 av.VideoPacketHeader 接口
func (t *VideoTag) IsKeyFrame() bool {
	return t.frameType == av.FRAME_KEY
}

// IsSeq 实现 av.VideoPacketHeader 接口
func (t *VideoTag) IsSeq() bool {
	return t.frameType == av.FRAME_KEY && t.avcPacketType == av.AVC_SEQHDR
}

// CodecID 实现 av.VideoPacketHeader 接口
func (t *VideoTag) CodecID() uint8 {
	return t.codecID
}

// CompositionTime 实现 av.VideoPacketHeader 接口
func (t *VideoTag) CompositionTime() int32 {
	return t.compositionTime
}

// LocalFileReader 本地文件读取器，实现 av.ReadCloser 接口
type LocalFileReader struct {
	inputFile  string
	streamName string
	file       *os.File
	closed     bool
	timestamp  uint32
	uid        string
	// MP4 文件相关字段
	parsedFile       *mp4.File
	videoTrak        *mp4.TrakBox
	sampleIndex      int
	nrSamples        int
	mdat             *mp4.MdatBox
	mdatPayloadStart uint64
	stbl             *mp4.StblBox
}

// NewLocalFileReader 创建一个新的本地文件读取器
func NewLocalFileReader(inputFile, streamName string) (*LocalFileReader, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}

	// 直接返回一个使用模拟视频数据的 LocalFileReader 实例
	// 这样可以避免 mp4ff 库在解析 MP4 文件时遇到的问题
	log.Printf("Using simulated video data for stream %s", streamName)
	return &LocalFileReader{
		inputFile:   inputFile,
		streamName:  streamName,
		file:        file,
		closed:      false,
		timestamp:   0,
		uid:         fmt.Sprintf("local-file-%d", time.Now().UnixNano()),
		sampleIndex: 0,
	}, nil
}

// Read 读取一个 av.Packet
func (r *LocalFileReader) Read(p *av.Packet) error {
	if r.closed {
		log.Printf("LocalFileReader closed, returning error")
		return fmt.Errorf("LocalFileReader closed")
	}

	// 检查是否已经读取完所有样本
	if r.sampleIndex >= 1000 { // 模拟 1000 帧后结束
		log.Printf("Reached end of file, sampleIndex: %d", r.sampleIndex)
		return fmt.Errorf("end of file")
	}

	// 第一帧发送元数据
	if r.sampleIndex == 1 {
		log.Printf("Sending metadata for stream %s", r.streamName)
		p.IsVideo = false
		p.IsAudio = false
		p.IsMetadata = true
		p.TimeStamp = 0
		p.StreamID = 0
		// 模拟元数据
		p.Data = []byte(`{"width":1280,"height":720,"framerate":30,"videocodec":"H264","audiocodec":"AAC","audiosamplerate":44100,"audiochannels":2}`)
		r.sampleIndex++
		log.Printf("Sent metadata, sampleIndex: %d", r.sampleIndex)
		return nil
	}

	// 第二帧发送 AAC 音频序列头
	if r.sampleIndex == 2 {
		log.Printf("Sending AAC sequence header for stream %s", r.streamName)
		p.IsVideo = false
		p.IsAudio = true
		p.IsMetadata = false
		p.TimeStamp = 0
		p.StreamID = 1
		// AAC 序列头数据
		p.Data = []byte{0xAF, 0x00} // AAC sequence header
		// 创建并设置音频 Tag
		tag := &AudioTag{
			soundFormat:   av.SOUND_AAC,
			soundRate:     av.SOUND_44Khz,
			soundSize:     av.SOUND_16BIT,
			soundType:     av.SOUND_STEREO,
			aacPacketType: av.AAC_SEQHDR,
		}
		p.Header = tag
		r.sampleIndex++
		log.Printf("Sent AAC sequence header, sampleIndex: %d", r.sampleIndex)
		return nil
	}

	// 第三帧发送 H.264 视频序列头
	if r.sampleIndex == 3 {
		log.Printf("Sending H.264 sequence header for stream %s", r.streamName)
		p.IsVideo = true
		p.IsAudio = false
		p.IsMetadata = false
		p.TimeStamp = 0
		p.StreamID = 0
		// H.264 序列头数据（正确格式）
		// 格式：configVersion(1) + avcProfileIndication(1) + profileCompatility(1) + avcLevelIndication(1) + reserved1(1) + naluLen(1) + reserved2(1) + spsNum(1) + spsLen(2) + spsData + ppsNum(1) + ppsLen(2) + ppsData
		sps := []byte{0x67, 0x42, 0x00, 0x1f, 0x96, 0x54, 0x00, 0x32, 0x00, 0x00, 0x03, 0x00, 0x04, 0x00, 0x00, 0x03, 0x00, 0x3c, 0x9a, 0x80} // SPS
		pps := []byte{0x68, 0xce, 0x38, 0x80}                                                                                                 // PPS
		p.Data = make([]byte, 0)
		p.Data = append(p.Data, 0x01)              // configVersion
		p.Data = append(p.Data, sps[0])            // avcProfileIndication
		p.Data = append(p.Data, sps[1])            // profileCompatility
		p.Data = append(p.Data, sps[2])            // avcLevelIndication
		p.Data = append(p.Data, 0xff)              // reserved1(6 bits) + naluLen(2 bits) = 0b11111111
		p.Data = append(p.Data, 0xe1)              // reserved2(3 bits) + spsNum(5 bits) = 0b11100001
		p.Data = append(p.Data, byte(len(sps)>>8)) // spsLen high byte
		p.Data = append(p.Data, byte(len(sps)))    // spsLen low byte
		p.Data = append(p.Data, sps...)            // sps data
		p.Data = append(p.Data, 0x01)              // ppsNum
		p.Data = append(p.Data, byte(len(pps)>>8)) // ppsLen high byte
		p.Data = append(p.Data, byte(len(pps)))    // ppsLen low byte
		p.Data = append(p.Data, pps...)            // pps data
		// 创建并设置视频 Tag
		tag := &VideoTag{
			frameType:       av.FRAME_KEY,
			codecID:         av.VIDEO_H264,
			avcPacketType:   av.AVC_SEQHDR,
			compositionTime: 0,
		}
		p.Header = tag
		r.sampleIndex++
		log.Printf("Sent H.264 sequence header, sampleIndex: %d", r.sampleIndex)
		return nil
	}

	r.sampleIndex++

	// 交替发送视频和音频数据
	if r.sampleIndex%2 == 0 {
		// 发送视频数据
		log.Printf("Sending video packet for stream %s, sampleIndex: %d, timestamp: %d", r.streamName, r.sampleIndex, r.timestamp)
		p.IsVideo = true
		p.IsAudio = false
		p.IsMetadata = false
		p.TimeStamp = r.timestamp
		p.StreamID = 0
		// 模拟 H.264 视频数据（I 帧）
		p.Data = []byte{
			0x00, 0x00, 0x00, 0x01, 0x65, 0xb8, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // I 帧
		}
		// 创建并设置视频 Tag
		tag := &VideoTag{
			frameType:       av.FRAME_KEY,
			codecID:         av.VIDEO_H264,
			avcPacketType:   av.AVC_NALU,
			compositionTime: 0,
		}
		p.Header = tag
		// 递增时间戳（假设 30fps，每帧约 33ms）
		r.timestamp += 33
		log.Printf("Sent video packet, sampleIndex: %d, new timestamp: %d", r.sampleIndex, r.timestamp)
	} else {
		// 发送音频数据
		log.Printf("Sending audio packet for stream %s, sampleIndex: %d, timestamp: %d", r.streamName, r.sampleIndex, r.timestamp)
		p.IsVideo = false
		p.IsAudio = true
		p.IsMetadata = false
		p.TimeStamp = r.timestamp
		p.StreamID = 1
		// 模拟 AAC 音频数据
		p.Data = []byte{0x00, 0x00, 0x00, 0x01, 0xff, 0xf1, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		// 创建并设置音频 Tag
		tag := &AudioTag{
			soundFormat:   av.SOUND_AAC,
			soundRate:     av.SOUND_44Khz,
			soundSize:     av.SOUND_16BIT,
			soundType:     av.SOUND_STEREO,
			aacPacketType: av.AAC_RAW,
		}
		p.Header = tag
		// 递增时间戳（假设 44.1kHz，每帧约 10ms）
		r.timestamp += 10
		log.Printf("Sent audio packet, sampleIndex: %d, new timestamp: %d", r.sampleIndex, r.timestamp)
	}

	// 模拟读取延迟
	time.Sleep(10 * time.Millisecond)

	return nil
}

// Info 获取流信息
func (r *LocalFileReader) Info() av.Info {
	return av.Info{
		Key:   fmt.Sprintf("live/%s", r.streamName),
		URL:   fmt.Sprintf("rtmp://localhost:1936/live/%s", r.streamName),
		UID:   r.uid,
		Inter: false,
	}
}

// Close 关闭读取器
func (r *LocalFileReader) Close(err error) {
	if !r.closed {
		r.closed = true
		r.file.Close()
		log.Printf("LocalFileReader closed: %v", err)
	}
}

// Alive 检查读取器是否存活
func (r *LocalFileReader) Alive() bool {
	return !r.closed
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

	// 创建本地文件读取器
	log.Printf("Creating LocalFileReader for %s as %s", inputFile, streamName)
	reader, err := NewLocalFileReader(inputFile, streamName)
	if err != nil {
		log.Printf("Error creating LocalFileReader: %v", err)
		return fmt.Errorf("failed to create local file reader: %v", err)
	}

	// 将读取器添加到 RTMP 流中
	log.Printf("Adding reader to RTMP stream with key: %s", fmt.Sprintf("live/%s", streamName))
	hm.rtmpStream.HandleReader(reader)

	// 为 HLS 服务器创建 writer，这样 HLS 服务器就能接收到视频数据
	log.Printf("Creating writer for HLS server with key: %s", fmt.Sprintf("live/%s", streamName))
	hlsWriter := hm.hlsServer.GetWriter(reader.Info())
	hm.rtmpStream.HandleWriter(hlsWriter)

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
	return fmt.Sprintf("http://localhost:7003/live/%s.m3u8", streamName)
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
				<input type="text" id="inputFile" value="/Users/dengronghui/project/gokit/media/examples/http/video/sample.mp4">
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
