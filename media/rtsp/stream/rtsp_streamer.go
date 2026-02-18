package stream

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/wwqdrh/gokit/media/rtsp/client"
)

// RTPInfo contains RTP packet information
type RTPInfo struct {
	SequenceNumber uint16
	Timestamp      uint32
	SSRC           uint32
	PayloadType    uint8
	Marker         bool
	Payload        []byte
}

// RTSPStream implements RTSPStreamer interface
type RTSPStream struct {
	config     RTSPConfig
	client     *client.Client
	conn       net.Conn
	rtpConn    net.Conn
	rtcpConn   net.Conn
	transport  string
	streamInfo StreamInfo
	running    bool
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	packetChan chan RTPInfo
	clientCount int
}

// NewRTSPStream creates a new RTSP streamer
func NewRTSPStream(config RTSPConfig) *RTSPStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &RTSPStream{
		config:     config,
		client:     client.NewClient(),
		streamInfo: StreamInfo{
			URL:        config.URL,
			StreamType: StreamTypeRTSP,
			StartedAt:  time.Now(),
			LastActive: time.Now(),
		},
		ctx:        ctx,
		cancel:     cancel,
		packetChan: make(chan RTPInfo, 1024),
	}
}

// Start starts the RTSP streamer
func (s *RTSPStream) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	// Parse RTSP URL to get address
	addr := s.parseRTSPAddress(s.config.URL)
	if addr == "" {
		return fmt.Errorf("invalid RTSP URL: %s", s.config.URL)
	}

	// Connect to RTSP server
	if err := s.client.Connect(addr); err != nil {
		return fmt.Errorf("failed to connect to RTSP server: %v", err)
	}

	// Send OPTIONS request
	_, err := s.client.Options()
	if err != nil {
		s.client.Close()
		return fmt.Errorf("failed to send OPTIONS: %v", err)
	}

	// Send DESCRIBE request
	response, err := s.client.Describe(s.config.URL)
	if err != nil {
		s.client.Close()
		return fmt.Errorf("failed to send DESCRIBE: %v", err)
	}

	// Parse SDP to get stream information
	s.parseSDP(string(response.Body))

	// Start pulling the stream
	if err := s.Pull(); err != nil {
		s.client.Close()
		return fmt.Errorf("failed to start pulling stream: %v", err)
	}

	s.running = true
	s.streamInfo.LastActive = time.Now()
	return nil
}

// Stop stops the RTSP streamer
func (s *RTSPStream) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.cancel()

	// Send TEARDOWN request
	if s.client != nil {
		_, _ = s.client.Teardown(s.config.URL)
		_ = s.client.Close()
	}

	// Close RTP/RTCP connections
	if s.rtpConn != nil {
		_ = s.rtpConn.Close()
	}
	if s.rtcpConn != nil {
		_ = s.rtcpConn.Close()
	}

	close(s.packetChan)
	s.running = false
	s.streamInfo.LastActive = time.Now()
	return nil
}

// IsRunning returns whether the streamer is running
func (s *RTSPStream) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetStreamInfo returns stream information
func (s *RTSPStream) GetStreamInfo() StreamInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.streamInfo
}

// Pull starts pulling the RTSP stream
func (s *RTSPStream) Pull() error {
	// Find available client ports for RTP/RTCP
	rtpPort, rtcpPort, err := s.findAvailablePorts()
	if err != nil {
		return err
	}

	// Create transport string
	transport := fmt.Sprintf("RTP/AVP;unicast;client_port=%d-%d", rtpPort, rtcpPort)

	// Send SETUP request
	response, err := s.client.Setup(s.config.URL, transport)
	if err != nil {
		return err
	}

	// Parse transport response
	s.transport = response.Header.Get("Transport")

	// Create RTP and RTCP listeners
	rtpListener, err := net.ListenUDP("udp", &net.UDPAddr{Port: rtpPort})
	if err != nil {
		return fmt.Errorf("failed to create RTP listener: %v", err)
	}
	s.rtpConn = rtpListener

	rtcpListener, err := net.ListenUDP("udp", &net.UDPAddr{Port: rtcpPort})
	if err != nil {
		rtpListener.Close()
		return fmt.Errorf("failed to create RTCP listener: %v", err)
	}
	s.rtcpConn = rtcpListener

	// Send PLAY request
	_, err = s.client.Play(s.config.URL)
	if err != nil {
		rtpListener.Close()
		rtcpListener.Close()
		return fmt.Errorf("failed to send PLAY: %v", err)
	}

	// Start packet processing goroutines
	go s.processRTPPackets()
	go s.processRTCPPackets()

	return nil
}

// GetRTSPConfig returns RTSP configuration
func (s *RTSPStream) GetRTSPConfig() RTSPConfig {
	return s.config
}

// GetPacketChan returns the RTP packet channel
func (s *RTSPStream) GetPacketChan() chan RTPInfo {
	return s.packetChan
}

// IncrementClientCount increments the client count
func (s *RTSPStream) IncrementClientCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clientCount++
	s.streamInfo.ClientCount = s.clientCount
}

// DecrementClientCount decrements the client count
func (s *RTSPStream) DecrementClientCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.clientCount > 0 {
		s.clientCount--
		s.streamInfo.ClientCount = s.clientCount
	}
}

// processRTPPackets processes RTP packets
func (s *RTSPStream) processRTPPackets() {
	buffer := make([]byte, 1500)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			if s.rtpConn == nil {
				return
			}

			n, _, err := s.rtpConn.(*net.UDPConn).ReadFromUDP(buffer)
			if err != nil {
				continue
			}

			if n < 12 {
				continue
			}

			// Parse RTP header
			header := buffer[:12]
			version := (header[0] >> 6) & 0x03
			if version != 2 {
				continue
			}

			padding := (header[0] >> 5) & 0x01
			extension := (header[0] >> 4) & 0x01
			csrcCount := header[0] & 0x0f
			marker := (header[1] >> 7) & 0x01
			payloadType := header[1] & 0x7f

			sequenceNumber := uint16(header[2])<<8 | uint16(header[3])
			timestamp := uint32(header[4])<<24 | uint32(header[5])<<16 | uint32(header[6])<<8 | uint32(header[7])
			ssrc := uint32(header[8])<<24 | uint32(header[9])<<16 | uint32(header[10])<<8 | uint32(header[11])

			// Calculate payload start and end
			payloadStart := 12 + int(csrcCount)*4
			if extension != 0 {
				extensionLength := uint16(buffer[payloadStart])<<8 | uint16(buffer[payloadStart+1])
				payloadStart += 2 + int(extensionLength)*4
			}

			if payloadStart >= n {
				continue
			}

			payloadEnd := n
			if padding != 0 {
				paddingSize := int(buffer[n-1])
				if paddingSize > 0 && paddingSize < payloadEnd-payloadStart {
					payloadEnd -= paddingSize
				}
			}

			// Create RTP info
		rtpInfo := RTPInfo{
				SequenceNumber: sequenceNumber,
				Timestamp:      timestamp,
				SSRC:           ssrc,
				PayloadType:    payloadType,
				Marker:         marker != 0,
				Payload:        buffer[payloadStart:payloadEnd],
			}

			// Send to channel
			select {
			case s.packetChan <- rtpInfo:
			default:
				// Channel full, drop packet
			}

			s.mu.Lock()
			s.streamInfo.LastActive = time.Now()
			s.mu.Unlock()
		}
	}
}

// processRTCPPackets processes RTCP packets
func (s *RTSPStream) processRTCPPackets() {
	buffer := make([]byte, 1500)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			if s.rtcpConn == nil {
				return
			}

			n, _, err := s.rtcpConn.(*net.UDPConn).ReadFromUDP(buffer)
			if err != nil {
				continue
			}

			// RTCP packet processing (simplified)
			if n < 4 {
				continue
			}

			// Just acknowledge receipt for now
		}
	}
}

// parseRTSPAddress parses RTSP URL to get address
func (s *RTSPStream) parseRTSPAddress(url string) string {
	if !strings.HasPrefix(url, "rtsp://") {
		return ""
	}

	// Remove rtsp:// prefix
	url = strings.TrimPrefix(url, "rtsp://")

	// Extract address part before path
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}

	address := parts[0]

	// Check if port is specified
	if !strings.Contains(address, ":") {
		address += ":554"
	}

	return address
}

// findAvailablePorts finds available UDP ports for RTP/RTCP
func (s *RTSPStream) findAvailablePorts() (int, int, error) {
	for i := 8000; i < 65535; i += 2 {
		rtpAddr := &net.UDPAddr{Port: i}
		rtpListener, err := net.ListenUDP("udp", rtpAddr)
		if err != nil {
			continue
		}

		rtcpAddr := &net.UDPAddr{Port: i + 1}
		rtcpListener, err := net.ListenUDP("udp", rtcpAddr)
		if err != nil {
			rtpListener.Close()
			continue
		}

		rtpListener.Close()
		rtcpListener.Close()
		return i, i + 1, nil
	}

	return 0, 0, fmt.Errorf("no available UDP ports found")
}

// parseSDP parses SDP to get stream information
func (s *RTSPStream) parseSDP(sdp string) {
	lines := strings.Split(sdp, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if len(line) < 2 {
			continue
		}

		typeChar := line[0]
		value := line[2:]

		switch typeChar {
		case 'm':
			// Media description
			parts := strings.Split(value, " ")
			if len(parts) >= 4 && parts[0] == "video" {
				s.streamInfo.VideoCodec = parts[3]
			}
		case 'a':
			// Attribute
			if strings.HasPrefix(value, "rtpmap:") {
				parts := strings.Split(value, " ")
				if len(parts) >= 2 {
					codecInfo := strings.Split(parts[1], "/")
					if len(codecInfo) >= 1 {
						s.streamInfo.VideoCodec = codecInfo[0]
					}
				}
			} else if strings.HasPrefix(value, "fmtp:") {
				// Format parameters
			}
		case 's':
			// Session name
		case 't':
			// Timing
		}
	}

	// Set default values if not found
	if s.streamInfo.VideoCodec == "" {
		s.streamInfo.VideoCodec = "H264"
	}
	if s.streamInfo.AudioCodec == "" {
		s.streamInfo.AudioCodec = "AAC"
	}
	if s.streamInfo.Resolution == "" {
		s.streamInfo.Resolution = "1920x1080"
	}
	if s.streamInfo.Framerate == 0 {
		s.streamInfo.Framerate = 25.0
	}
	if s.streamInfo.Bitrate == 0 {
		s.streamInfo.Bitrate = 2048
	}
}
