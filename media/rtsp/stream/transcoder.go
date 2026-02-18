package stream

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TranscodedStream implements Streamer interface for transcoded streams
type TranscodedStream struct {
	inputStream   Streamer
	streamType    StreamType
	streamInfo    StreamInfo
	running       bool
	mu            sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
	outputChan    chan []byte
	clientCount   int
}

// NewTranscodedStream creates a new transcoded stream
func NewTranscodedStream(input Streamer, streamType StreamType) *TranscodedStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &TranscodedStream{
		inputStream: input,
		streamType:  streamType,
		streamInfo: StreamInfo{
			URL:        fmt.Sprintf("%s_%s", input.GetStreamInfo().URL, streamType),
			StreamType: streamType,
			StartedAt:  time.Now(),
			LastActive: time.Now(),
		},
		ctx:        ctx,
		cancel:     cancel,
		outputChan: make(chan []byte, 1024),
	}
}

// Start starts the transcoded stream
func (s *TranscodedStream) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.running = true
	s.streamInfo.LastActive = time.Now()
	return nil
}

// Stop stops the transcoded stream
func (s *TranscodedStream) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.cancel()
	close(s.outputChan)
	s.running = false
	s.streamInfo.LastActive = time.Now()
	return nil
}

// IsRunning returns whether the streamer is running
func (s *TranscodedStream) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetStreamInfo returns stream information
func (s *TranscodedStream) GetStreamInfo() StreamInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.streamInfo
}

// GetOutputChan returns the output channel
func (s *TranscodedStream) GetOutputChan() chan []byte {
	return s.outputChan
}

// IncrementClientCount increments the client count
func (s *TranscodedStream) IncrementClientCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clientCount++
	s.streamInfo.ClientCount = s.clientCount
}

// DecrementClientCount decrements the client count
func (s *TranscodedStream) DecrementClientCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.clientCount > 0 {
		s.clientCount--
		s.streamInfo.ClientCount = s.clientCount
	}
}

// StreamTranscoder implements Transcoder interface
type StreamTranscoder struct {
	inputStream   Streamer
	outputStreams map[StreamType]*TranscodedStream
	running       bool
	mu            sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
	outputURLs    map[StreamType]string
	hlsDir        string
}

// NewStreamTranscoder creates a new stream transcoder
func NewStreamTranscoder() *StreamTranscoder {
	ctx, cancel := context.WithCancel(context.Background())
	return &StreamTranscoder{
		outputStreams: make(map[StreamType]*TranscodedStream),
		outputURLs:    make(map[StreamType]string),
		ctx:           ctx,
		cancel:        cancel,
		hlsDir:        "/tmp/hls",
	}
}

// Start starts the transcoder
func (t *StreamTranscoder) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return nil
	}

	// Create HLS directory if needed
	if _, err := os.Stat(t.hlsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(t.hlsDir, 0755); err != nil {
			return fmt.Errorf("failed to create HLS directory: %v", err)
		}
	}

	t.running = true
	return nil
}

// Stop stops the transcoder
func (t *StreamTranscoder) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return nil
	}

	t.cancel()

	// Stop all output streams
	for _, stream := range t.outputStreams {
		stream.Stop()
	}

	t.running = false
	return nil
}

// IsRunning returns whether the transcoder is running
func (t *StreamTranscoder) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.running
}

// GetStreamInfo returns stream information
func (t *StreamTranscoder) GetStreamInfo() StreamInfo {
	if t.inputStream != nil {
		return t.inputStream.GetStreamInfo()
	}
	return StreamInfo{}
}

// Transcode starts the transcoding process
func (t *StreamTranscoder) Transcode(input Streamer, outputFormats []StreamType) error {
	t.inputStream = input

	// Create output streams
	for _, format := range outputFormats {
		stream := NewTranscodedStream(input, format)
		if err := stream.Start(); err != nil {
			return err
		}
		t.outputStreams[format] = stream
		t.outputURLs[format] = fmt.Sprintf("http://localhost:8080/stream/%s", format)
	}

	// Start transcoding process
	go t.transcodeLoop()

	return nil
}

// transcodeLoop processes the transcoding
func (t *StreamTranscoder) transcodeLoop() {
	if rtspStream, ok := t.inputStream.(*RTSPStream); ok {
		packetChan := rtspStream.GetPacketChan()
		for {
			select {
			case <-t.ctx.Done():
				return
			case packet, ok := <-packetChan:
				if !ok {
					return
				}
				
				// Transcode to different formats
				for format, stream := range t.outputStreams {
					switch format {
					case StreamTypeFLV:
						data := t.transcodeToFLV(packet)
						if len(data) > 0 {
							select {
							case stream.outputChan <- data:
							default:
								// Channel full, drop data
							}
						}
					case StreamTypeHLS:
						t.transcodeToHLS(packet)
					case StreamTypeWebRTC:
						data := t.transcodeToWebRTC(packet)
						if len(data) > 0 {
							select {
							case stream.outputChan <- data:
							default:
								// Channel full, drop data
							}
						}
					}
				}
			}
		}
	}
}

// transcodeToFLV transcodes RTP packet to FLV format
func (t *StreamTranscoder) transcodeToFLV(packet RTPInfo) []byte {
	// FLV header (9 bytes)
	flvHeader := []byte{0x46, 0x4C, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09}
	
	// FLV body
	var body bytes.Buffer
	
	// Create video tag
	tagType := byte(0x09) // Video tag
	dataSize := uint32(len(packet.Payload) + 5)
	timestamp := uint32(packet.Timestamp)
	timestampExtended := byte(timestamp >> 24)
	
	// Write tag header
	body.WriteByte(tagType)
	body.Write([]byte{byte(dataSize >> 16), byte(dataSize >> 8), byte(dataSize)})
	body.Write([]byte{byte(timestamp >> 16), byte(timestamp >> 8), byte(timestamp)})
	body.WriteByte(timestampExtended)
	body.Write([]byte{0x00, 0x00, 0x00}) // Stream ID
	
	// Write video data
	frameType := byte(0x10) // Key frame
	codecID := byte(0x07)   // H264
	body.WriteByte(frameType | codecID)
	
	// AVC packet type
	avcPacketType := byte(0x01) // NAL unit
	body.WriteByte(avcPacketType)
	
	// Composition time
	compositionTime := int32(0)
	body.Write([]byte{byte(compositionTime >> 16), byte(compositionTime >> 8), byte(compositionTime)})
	
	// Write NAL unit
	body.Write(packet.Payload)
	
	// Write tag size
	tagSize := uint32(body.Len())
	tagSizeBytes := []byte{byte(tagSize >> 24), byte(tagSize >> 16), byte(tagSize >> 8), byte(tagSize)}
	body.Write(tagSizeBytes)
	
	// Combine header and body
	result := append(flvHeader, body.Bytes()...)
	return result
}

// transcodeToHLS transcodes RTP packet to HLS format
func (t *StreamTranscoder) transcodeToHLS(packet RTPInfo) {
	// Simplified HLS implementation
	// In real implementation, you would:
	// 1. Collect packets into a TS segment
	// 2. When segment reaches desired duration, write to file
	// 3. Update m3u8 playlist
	
	// For now, just create dummy files
	streamName := "test"
	segmentName := fmt.Sprintf("%s_%d.ts", streamName, time.Now().Unix())
	segmentPath := filepath.Join(t.hlsDir, segmentName)
	
	// Write dummy TS file
	if err := os.WriteFile(segmentPath, packet.Payload, 0644); err != nil {
		return
	}
	
	// Update m3u8 playlist
	playlistPath := filepath.Join(t.hlsDir, "playlist.m3u8")
	playlistContent := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
`
	
	// Add segment
	playlistContent += fmt.Sprintf("#EXTINF:10.0,\n%s\n", segmentName)
	
	if err := os.WriteFile(playlistPath, []byte(playlistContent), 0644); err != nil {
		return
	}
}

// transcodeToWebRTC transcodes RTP packet to WebRTC format
func (t *StreamTranscoder) transcodeToWebRTC(packet RTPInfo) []byte {
	// WebRTC uses RTP packets directly
	// In real implementation, you would:
	// 1. Encrypt the packet
	// 2. Add WebRTC headers
	// 3. Send over peer connection
	
	// For now, just return the RTP payload
	return packet.Payload
}

// GetOutputURLs returns the output stream URLs
func (t *StreamTranscoder) GetOutputURLs() map[StreamType]string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.outputURLs
}

// SetHLSPath sets the HLS output path
func (t *StreamTranscoder) SetHLSPath(path string) {
	t.hlsDir = path
}

// GetOutputStream returns the output stream for a specific format
func (t *StreamTranscoder) GetOutputStream(streamType StreamType) Streamer {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.outputStreams[streamType]
}
