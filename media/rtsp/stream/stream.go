package stream

import (
	"context"
	"net"
	"time"
)

// StreamType defines the type of stream format
type StreamType string

const (
	StreamTypeRTSP  StreamType = "rtsp"
	StreamTypeFLV   StreamType = "flv"
	StreamTypeHLS   StreamType = "hls"
	StreamTypeWebRTC StreamType = "webrtc"
)

// Streamer defines the interface for stream processing
type Streamer interface {
	// Start starts the stream processing
	Start() error
	// Stop stops the stream processing
	Stop() error
	// IsRunning returns whether the streamer is running
	IsRunning() bool
	// GetStreamInfo returns stream information
	GetStreamInfo() StreamInfo
}

// StreamInfo contains stream metadata
type StreamInfo struct {
	URL            string
	StreamType     StreamType
	VideoCodec     string
	AudioCodec     string
	Resolution     string
	Framerate      float64
	Bitrate        int
	StartedAt      time.Time
	LastActive     time.Time
	ClientCount    int
}

// RTSPStreamer defines the interface for RTSP stream processing
type RTSPStreamer interface {
	Streamer
	// Pull starts pulling the RTSP stream
	Pull() error
	// GetRTSPConfig returns RTSP configuration
	GetRTSPConfig() RTSPConfig
}

// RTSPConfig contains RTSP stream configuration
type RTSPConfig struct {
	URL               string
	Username          string
	Password          string
	Transport         string
	BufferSize        int
	RetryInterval     time.Duration
	MaxRetries        int
}

// Transcoder defines the interface for stream transcoding
type Transcoder interface {
	// Transcode starts the transcoding process
	Transcode(input Streamer, outputFormats []StreamType) error
	// GetOutputURLs returns the output stream URLs
	GetOutputURLs() map[StreamType]string
}

// Distributor defines the interface for stream distribution
type Distributor interface {
	// Distribute starts the distribution service
	Distribute() error
	// AddStream adds a stream to the distributor
	AddStream(stream Streamer) error
	// RemoveStream removes a stream from the distributor
	RemoveStream(streamID string) error
	// GetListenAddr returns the listen address
	GetListenAddr() net.Addr
}

// PipelineManager manages the entire stream processing pipeline
type PipelineManager struct {
	rtspStreamer RTSPStreamer
	transcoder   Transcoder
	distributor  Distributor
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
}

// NewPipelineManager creates a new pipeline manager
func NewPipelineManager(rtspConfig RTSPConfig) *PipelineManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &PipelineManager{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the entire stream processing pipeline
func (pm *PipelineManager) Start() error {
	// Start RTSP streamer
	if err := pm.rtspStreamer.Start(); err != nil {
		return err
	}

	// Start transcoder
	if err := pm.transcoder.Transcode(pm.rtspStreamer, []StreamType{StreamTypeFLV, StreamTypeHLS}); err != nil {
		pm.rtspStreamer.Stop()
		return err
	}

	// Start distributor
	if err := pm.distributor.Distribute(); err != nil {
		pm.rtspStreamer.Stop()
		pm.transcoder.(Streamer).Stop()
		return err
	}

	pm.running = true
	return nil
}

// Stop stops the entire stream processing pipeline
func (pm *PipelineManager) Stop() error {
	pm.cancel()

	// Stop components in reverse order
	if pm.distributor != nil {
		pm.distributor.(Streamer).Stop()
	}

	if pm.transcoder != nil {
		pm.transcoder.(Streamer).Stop()
	}

	if pm.rtspStreamer != nil {
		pm.rtspStreamer.Stop()
	}

	pm.running = false
	return nil
}

// IsRunning returns whether the pipeline manager is running
func (pm *PipelineManager) IsRunning() bool {
	return pm.running
}

// GetOutputURLs returns all output stream URLs
func (pm *PipelineManager) GetOutputURLs() map[StreamType]string {
	if pm.transcoder != nil {
		return pm.transcoder.GetOutputURLs()
	}
	return nil
}
