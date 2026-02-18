package stream

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// StreamDistributor implements Distributor interface
type StreamDistributor struct {
	addr        string
	server      *http.Server
	streams     map[string]Streamer
	mu          sync.Mutex
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
	hlsDir      string
	clientCount int
}

// NewStreamDistributor creates a new stream distributor
func NewStreamDistributor(addr string) *StreamDistributor {
	ctx, cancel := context.WithCancel(context.Background())
	return &StreamDistributor{
		addr:    addr,
		streams: make(map[string]Streamer),
		ctx:     ctx,
		cancel:  cancel,
		hlsDir:  "/tmp/hls",
	}
}

// Start starts the distributor
func (d *StreamDistributor) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return nil
	}

	// Create HTTP server
	handler := http.NewServeMux()
	handler.HandleFunc("/stream/", d.streamHandler)
	handler.HandleFunc("/hls/", d.hlsHandler)
	handler.HandleFunc("/webrtc/", d.webrtcHandler)
	handler.HandleFunc("/status", d.statusHandler)

	d.server = &http.Server{
		Addr:    d.addr,
		Handler: handler,
	}

	// Start server in goroutine
	go func() {
		if err := d.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	d.running = true
	return nil
}

// Stop stops the distributor
func (d *StreamDistributor) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	d.cancel()

	// Shutdown HTTP server
	if d.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := d.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %v", err)
		}
	}

	// Remove all streams
	for _, stream := range d.streams {
		stream.Stop()
	}

	d.running = false
	return nil
}

// IsRunning returns whether the distributor is running
func (d *StreamDistributor) IsRunning() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.running
}

// GetStreamInfo returns stream information (not implemented for distributor)
func (d *StreamDistributor) GetStreamInfo() StreamInfo {
	return StreamInfo{
		URL:        fmt.Sprintf("http://%s", d.addr),
		StreamType: "distributor",
		StartedAt:  time.Now(),
		LastActive: time.Now(),
	}
}

// Distribute starts the distribution service
func (d *StreamDistributor) Distribute() error {
	return d.Start()
}

// AddStream adds a stream to the distributor
func (d *StreamDistributor) AddStream(stream Streamer) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	streamID := stream.GetStreamInfo().URL
	d.streams[streamID] = stream
	return nil
}

// RemoveStream removes a stream from the distributor
func (d *StreamDistributor) RemoveStream(streamID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if stream, ok := d.streams[streamID]; ok {
		stream.Stop()
		delete(d.streams, streamID)
	}
	return nil
}

// GetListenAddr returns the listen address
func (d *StreamDistributor) GetListenAddr() net.Addr {
	if d.server != nil {
		// This is a placeholder, in real implementation we would get the actual address
		return &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 8080}
	}
	return nil
}

// streamHandler handles stream requests
func (d *StreamDistributor) streamHandler(w http.ResponseWriter, r *http.Request) {
	// Get stream path
	path := r.URL.Path
	if len(path) < 8 { // /stream/ prefix
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	streamType := path[8:]
	switch streamType {
	case "flv":
		d.flvHandler(w, r)
	case "hls":
		http.Redirect(w, r, "/hls/playlist.m3u8", http.StatusFound)
	case "webrtc":
		d.webrtcHandler(w, r)
	default:
		http.Error(w, "Unknown stream type", http.StatusBadRequest)
	}
}

// flvHandler handles FLV stream requests
func (d *StreamDistributor) flvHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for FLV streaming
	w.Header().Set("Content-Type", "video/x-flv")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Flush headers
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Find FLV stream
	var flvStream *TranscodedStream
	d.mu.Lock()
	for _, stream := range d.streams {
		if transcoded, ok := stream.(*TranscodedStream); ok && transcoded.streamType == StreamTypeFLV {
			flvStream = transcoded
			flvStream.IncrementClientCount()
			break
		}
	}
	d.mu.Unlock()

	if flvStream == nil {
		http.Error(w, "FLV stream not available", http.StatusNotFound)
		return
	}
	defer flvStream.DecrementClientCount()

	// Stream FLV data
	outputChan := flvStream.GetOutputChan()
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-r.Context().Done():
			return
		case data, ok := <-outputChan:
			if !ok {
				return
			}

			// Write data
			if _, err := w.Write(data); err != nil {
				return
			}

			// Flush data
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

// hlsHandler handles HLS stream requests
func (d *StreamDistributor) hlsHandler(w http.ResponseWriter, r *http.Request) {
	// Get file path
	path := r.URL.Path
	if len(path) < 5 { // /hls/ prefix
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	filename := path[5:]
	if filename == "" {
		filename = "playlist.m3u8"
	}

	hlsPath := filepath.Join(d.hlsDir, filename)

	// Set headers
	switch filepath.Ext(filename) {
	case ".m3u8":
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	case ".ts":
		w.Header().Set("Content-Type", "video/MP2T")
	default:
		http.Error(w, "Unknown file type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Serve file
	http.ServeFile(w, r, hlsPath)
}

// webrtcHandler handles WebRTC requests
func (d *StreamDistributor) webrtcHandler(w http.ResponseWriter, r *http.Request) {
	// WebRTC implementation would typically use WebSocket
	// For now, return a placeholder
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"status":"webrtc_available","message":"WebRTC streaming is available"}`)
}

// statusHandler handles status requests
func (d *StreamDistributor) statusHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Build status
	status := fmt.Sprintf(`{
	"status": "%s",
	"addr": "%s",
	"streams": [`,
		func() string { if d.running { return "running" } else { return "stopped" } }(),
		d.addr,
	)

	// Add stream statuses
	d.mu.Lock()
	for id, stream := range d.streams {
		info := stream.GetStreamInfo()
		status += fmt.Sprintf(`
		{
			"id": "%s",
			"type": "%s",
			"url": "%s",
			"clients": %d,
			"last_active": "%s"
		},`,
			id,
			info.StreamType,
			info.URL,
			info.ClientCount,
			info.LastActive.Format(time.RFC3339),
		)
	}
	d.mu.Unlock()

	// Remove trailing comma and close
	if len(d.streams) > 0 {
		status = status[:len(status)-1]
	}
	status += "\n	]\n}"

	// Write response
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, status)
}

// SetHLSPath sets the HLS directory path
func (d *StreamDistributor) SetHLSPath(path string) {
	d.hlsDir = path
}

// GetClientCount returns the total client count
func (d *StreamDistributor) GetClientCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.clientCount
}
