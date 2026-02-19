package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wwqdrh/gokit/media/rtsp"
	"github.com/wwqdrh/gokit/media/rtsp/stream"
) // Handler defines the interface for RTSP request handlers
type Handler interface {
	HandleOPTIONS(*Session, *rtsp.Request) (*rtsp.Response, error)
	HandleDESCRIBE(*Session, *rtsp.Request) (*rtsp.Response, error)
	HandleSETUP(*Session, *rtsp.Request) (*rtsp.Response, error)
	HandlePLAY(*Session, *rtsp.Request) (*rtsp.Response, error)
	HandlePAUSE(*Session, *rtsp.Request) (*rtsp.Response, error)
	HandleTEARDOWN(*Session, *rtsp.Request) (*rtsp.Response, error)
	HandleANNOUNCE(*Session, *rtsp.Request) (*rtsp.Response, error)
	HandleRECORD(*Session, *rtsp.Request) (*rtsp.Response, error)
	HandleGET_PARAMETER(*Session, *rtsp.Request) (*rtsp.Response, error)
	HandleSET_PARAMETER(*Session, *rtsp.Request) (*rtsp.Response, error)
	HandleREDIRECT(*Session, *rtsp.Request) (*rtsp.Response, error)
}

// DefaultHandler provides a default implementation of Handler
type DefaultHandler struct{}

// HandleOPTIONS handles OPTIONS requests
func (h *DefaultHandler) HandleOPTIONS(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:    "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	response.Header.Set("Public", "OPTIONS, DESCRIBE, SETUP, PLAY, PAUSE, TEARDOWN, ANNOUNCE, RECORD, GET_PARAMETER, SET_PARAMETER")
	return response, nil
}

// HandleDESCRIBE handles DESCRIBE requests
func (h *DefaultHandler) HandleDESCRIBE(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:    "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	response.Header.Set("Content-Type", "application/sdp")

	// Extract video filename from URI
	videoFile := request.URI
	if videoFile == "/" {
		videoFile = "test"
	} else {
		// Remove leading slash
		videoFile = videoFile[1:]
	}

	// Example SDP with proper stream information
	sdp := `v=0
` +
		`o=- 12345 12345 IN IP4 127.0.0.1
` +
		`s=GoRTSP Server - ` + videoFile + `
` +
		`t=0 0
` +
		`m=video 0 RTP/AVP 96
` +
		`c=IN IP4 0.0.0.0
` +
		`a=rtpmap:96 H264/90000
` +
		`a=control:streamid=0
`
	response.Body = []byte(sdp)
	return response, nil
}

// HandleSETUP handles SETUP requests
func (h *DefaultHandler) HandleSETUP(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:    "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}

	// Parse transport header
	transport := request.Header.Get("Transport")
	if transport == "" {
		response.StatusCode = rtsp.StatusBadRequest
		response.StatusText = rtsp.StatusText(rtsp.StatusBadRequest)
		return response, nil
	}

	// Extract client ports from transport header
	// Example transport: RTP/AVP;unicast;client_port=5000-5001
	clientPort := ""
	parts := strings.Split(transport, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "client_port=") {
			clientPort = strings.TrimPrefix(part, "client_port=")
			break
		}
	}

	if clientPort == "" {
		response.StatusCode = rtsp.StatusBadRequest
		response.StatusText = rtsp.StatusText(rtsp.StatusBadRequest)
		return response, nil
	}

	// Extract client IP from connection
	// clientAddr := session.Conn.RemoteAddr().(*net.TCPAddr).IP.String()

	// Parse client ports
	portParts := strings.Split(clientPort, "-")
	if len(portParts) != 2 {
		response.StatusCode = rtsp.StatusBadRequest
		response.StatusText = rtsp.StatusText(rtsp.StatusBadRequest)
		return response, nil
	}

	// Store client address and ports in session
	session.Transport = transport

	// Set transport response
	response.Header.Set("Transport", transport+";server_port=8000-8001")
	response.Header.Set("Session", session.ID)

	return response, nil
}

// HandlePLAY handles PLAY requests
func (h *DefaultHandler) HandlePLAY(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:    "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	response.Header.Set("Session", session.ID)
	response.Header.Set("RTP-Info", "url="+request.URI+";seq=0;rtptime=0")

	// Start streaming if not already running
	if !session.StreamRunning {
		session.StreamRunning = true

		// Extract video filename from URI
		videoFile := request.URI
		if videoFile == "/" {
			videoFile = "test"
		} else {
			// Remove leading slash
			videoFile = videoFile[1:]
		}

		// Get server instance to access VideoDir
		// server := &Server{}
		// Note: In a real implementation, we would need to pass the server instance to the handler
		// For simplicity, we'll use a relative path
		videoPath := filepath.Join(".", "video", videoFile+".mp4")

		// Check if video file exists
		if _, err := os.Stat(videoPath); os.IsNotExist(err) {
			// Try without extension
			videoPath = filepath.Join(".", "video", videoFile)
			if _, err := os.Stat(videoPath); os.IsNotExist(err) {
				response.StatusCode = rtsp.StatusNotFound
				response.StatusText = rtsp.StatusText(rtsp.StatusNotFound)
				return response, nil
			}
		}

		session.VideoPath = videoPath

		// Start streaming goroutine
		go func() {
			// For demonstration purposes, we'll simulate H.264 video streaming
			fmt.Printf("Starting to stream video: %s\n", videoPath)

			// Parse client address and port from transport header
			clientAddr := session.Conn.RemoteAddr().(*net.TCPAddr).IP.String()
			clientPort := ""
			parts := strings.Split(session.Transport, ";")
			for _, part := range parts {
				if strings.HasPrefix(part, "client_port=") {
					clientPort = strings.TrimPrefix(part, "client_port=")
					break
				}
			}

			// Parse client ports
			portParts := strings.Split(clientPort, "-")
			if len(portParts) != 2 {
				fmt.Printf("Invalid client port format: %s\n", clientPort)
				session.StreamRunning = false
				return
			}

			// Create UDP connection for RTP packets
			rtpAddr, err := net.ResolveUDPAddr("udp", clientAddr+":"+portParts[0])
			if err != nil {
				fmt.Printf("Failed to resolve UDP address: %v\n", err)
				session.StreamRunning = false
				return
			}

			udpConn, err := net.DialUDP("udp", nil, rtpAddr)
			if err != nil {
				fmt.Printf("Failed to create UDP connection: %v\n", err)
				session.StreamRunning = false
				return
			}
			defer udpConn.Close()

			// Simulate video streaming
			sequenceNum := uint16(0)
			timestamp := uint32(0)
			ssrc := uint32(time.Now().UnixNano() % 0xffffffff)

			// Simulate 30 frames per second
			frameInterval := time.Second / 30

			// Stream for 30 seconds
			for i := 0; i < 900 && session.StreamRunning; i++ {
				// Create a dummy H.264 NALU
				// In a real implementation, we would decode the video file to get actual H.264 data
				naluType := byte(1) // Non-key frame
				if i%30 == 0 {
					naluType = byte(5) // Key frame every 30 frames
				}

				// Create dummy NALU data
				naluData := make([]byte, 1000)
				naluData[0] = naluType
				for j := 1; j < len(naluData); j++ {
					naluData[j] = byte(j % 256)
				}

				// Create RTP packet
				rtpPacket := createRTPPacket(naluData, sequenceNum, timestamp, ssrc, naluType == 5)

				// Send RTP packet to client
				n, err := udpConn.Write(rtpPacket)
				if err != nil {
					fmt.Printf("Failed to send RTP packet: %v\n", err)
					continue
				}

				fmt.Printf("Sent RTP packet: seq=%d, timestamp=%d, size=%d, bytes_sent=%d\n", sequenceNum, timestamp, len(rtpPacket), n)

				// Increment sequence number and timestamp
				sequenceNum++
				timestamp += 3000 // 90000 Hz / 30 fps = 3000

				// Sleep for frame interval
				time.Sleep(frameInterval)
			}

			// Stop streaming
			session.StreamRunning = false
			fmt.Printf("Stopped streaming video: %s\n", videoPath)
		}()
	}

	return response, nil
}

// createRTPPacket creates an RTP packet from H.264 NALU data
func createRTPPacket(naluData []byte, sequenceNum uint16, timestamp uint32, ssrc uint32, isKeyFrame bool) []byte {
	// RTP header size
	headerSize := 12

	// Total packet size
	totalSize := headerSize + len(naluData)

	// Create RTP packet
	rtpPacket := make([]byte, totalSize)

	// Set RTP header
	rtpPacket[0] = 0x80 // Version 2, padding 0, extension 0, CSRC count 0
	rtpPacket[1] = 96   // Payload type 96 (H.264)
	if isKeyFrame {
		rtpPacket[1] |= 0x80 // Set marker bit for key frames
	}

	// Set sequence number (big-endian)
	rtpPacket[2] = byte(sequenceNum >> 8)
	rtpPacket[3] = byte(sequenceNum)

	// Set timestamp (big-endian)
	rtpPacket[4] = byte(timestamp >> 24)
	rtpPacket[5] = byte(timestamp >> 16)
	rtpPacket[6] = byte(timestamp >> 8)
	rtpPacket[7] = byte(timestamp)

	// Set SSRC (big-endian)
	rtpPacket[8] = byte(ssrc >> 24)
	rtpPacket[9] = byte(ssrc >> 16)
	rtpPacket[10] = byte(ssrc >> 8)
	rtpPacket[11] = byte(ssrc)

	// Copy NALU data
	copy(rtpPacket[headerSize:], naluData)

	return rtpPacket
}

// HandlePAUSE handles PAUSE requests
func (h *DefaultHandler) HandlePAUSE(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:    "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	response.Header.Set("Session", session.ID)

	// Pause streaming
	if session.StreamRunning {
		session.StreamRunning = false
		fmt.Printf("Paused streaming for session: %s\n", session.ID)
	}

	return response, nil
}

// HandleTEARDOWN handles TEARDOWN requests
func (h *DefaultHandler) HandleTEARDOWN(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:    "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}

	// Stop streaming and clean up resources
	if session.StreamRunning {
		session.StreamRunning = false
		fmt.Printf("Stopped streaming for session: %s\n", session.ID)
	}

	// Close RTP/RTCP connections if they exist
	if session.RTPConn != nil {
		session.RTPConn.Close()
		session.RTPConn = nil
	}
	if session.RTCPConn != nil {
		session.RTCPConn.Close()
		session.RTCPConn = nil
	}

	// Stop streamer if it exists
	if session.Streamer != nil {
		session.Streamer.Stop()
		session.Streamer = nil
	}

	return response, nil
}

// HandleANNOUNCE handles ANNOUNCE requests
func (h *DefaultHandler) HandleANNOUNCE(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:    "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	return response, nil
}

// HandleRECORD handles RECORD requests
func (h *DefaultHandler) HandleRECORD(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:    "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	response.Header.Set("Session", session.ID)
	return response, nil
}

// HandleGET_PARAMETER handles GET_PARAMETER requests
func (h *DefaultHandler) HandleGET_PARAMETER(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:    "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	return response, nil
}

// HandleSET_PARAMETER handles SET_PARAMETER requests
func (h *DefaultHandler) HandleSET_PARAMETER(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:    "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	return response, nil
}

// HandleREDIRECT handles REDIRECT requests
func (h *DefaultHandler) HandleREDIRECT(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:    "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	return response, nil
}

// Session represents an RTSP session
type Session struct {
	ID         string
	Conn       net.Conn
	CSeq       int
	Transport  string
	CreatedAt  time.Time
	LastActive time.Time
	// Video streaming related fields
	VideoPath     string
	RTPConn       net.Conn
	RTCPConn      net.Conn
	StreamRunning bool
	Streamer      *stream.RTSPStream
	SequenceNum   uint16
	Timestamp     uint32
	SSRC          uint32
}

// Server represents an RTSP server
type Server struct {
	addr     string
	listener net.Listener
	handler  Handler
	sessions map[string]*Session
	sync.Mutex
	running bool
	// Video directory
	VideoDir string
}

// NewServer creates a new RTSP server
func NewServer(addr string) *Server {
	return &Server{
		addr:     addr,
		handler:  &DefaultHandler{},
		sessions: make(map[string]*Session),
		VideoDir: ".", // Default to current directory
	}
}

// SetVideoDir sets the directory where video files are stored
func (s *Server) SetVideoDir(dir string) {
	s.VideoDir = dir
}

// SetHandler sets the request handler
func (s *Server) SetHandler(handler Handler) {
	s.handler = handler
}

// Start starts the server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = listener
	s.running = true

	go s.acceptLoop()
	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	s.running = false
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// acceptLoop accepts incoming connections
func (s *Server) acceptLoop() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	session := &Session{
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		Conn:       conn,
		CSeq:       0,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		// Initialize video streaming fields
		VideoPath:     "",
		RTPConn:       nil,
		RTCPConn:      nil,
		StreamRunning: false,
		Streamer:      nil,
		SequenceNum:   0,
		Timestamp:     0,
		SSRC:          uint32(time.Now().UnixNano() % 0xffffffff),
	}

	// Add session
	s.Lock()
	s.sessions[session.ID] = session
	s.Unlock()

	defer func() {
		// Stop streaming if running
		if session.StreamRunning {
			session.StreamRunning = false
			if session.RTPConn != nil {
				session.RTPConn.Close()
			}
			if session.RTCPConn != nil {
				session.RTCPConn.Close()
			}
			if session.Streamer != nil {
				session.Streamer.Stop()
			}
		}
		// Remove session
		s.Lock()
		delete(s.sessions, session.ID)
		s.Unlock()
	}()

	for {
		// Parse request
		request, err := rtsp.ParseRequest(reader)
		if err != nil {
			break
		}

		// Update last active time
		session.LastActive = time.Now()

		// Get CSeq
		cseq, err := rtsp.ParseCSeq(request.Header)
		if err == nil {
			session.CSeq = cseq
		}

		// Handle request
		response, err := s.handleRequest(session, request)
		if err != nil {
			response = &rtsp.Response{
				Version:    "RTSP/1.0",
				StatusCode: rtsp.StatusInternalServerError,
				StatusText: rtsp.StatusText(rtsp.StatusInternalServerError),
				Header:     make(rtsp.Header),
			}
		}

		// Set CSeq in response
		response.Header.Set("CSeq", rtsp.BuildCSeq(session.CSeq))

		// Serialize and send response
		data, err := rtsp.SerializeResponse(response)
		if err != nil {
			break
		}

		_, err = conn.Write(data)
		if err != nil {
			break
		}
	}
}

// handleRequest handles an RTSP request
func (s *Server) handleRequest(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	switch request.Method {
	case rtsp.MethodOPTIONS:
		return s.handler.HandleOPTIONS(session, request)
	case rtsp.MethodDESCRIBE:
		return s.handler.HandleDESCRIBE(session, request)
	case rtsp.MethodSETUP:
		return s.handler.HandleSETUP(session, request)
	case rtsp.MethodPLAY:
		return s.handler.HandlePLAY(session, request)
	case rtsp.MethodPAUSE:
		return s.handler.HandlePAUSE(session, request)
	case rtsp.MethodTEARDOWN:
		return s.handler.HandleTEARDOWN(session, request)
	case rtsp.MethodANNOUNCE:
		return s.handler.HandleANNOUNCE(session, request)
	case rtsp.MethodRECORD:
		return s.handler.HandleRECORD(session, request)
	case rtsp.MethodGET_PARAMETER:
		return s.handler.HandleGET_PARAMETER(session, request)
	case rtsp.MethodSET_PARAMETER:
		return s.handler.HandleSET_PARAMETER(session, request)
	case rtsp.MethodREDIRECT:
		return s.handler.HandleREDIRECT(session, request)
	default:
		response := &rtsp.Response{
			Version:    "RTSP/1.0",
			StatusCode: rtsp.StatusMethodNotAllowed,
			StatusText: rtsp.StatusText(rtsp.StatusMethodNotAllowed),
			Header:     make(rtsp.Header),
		}
		return response, nil
	}
}

// GetSession returns a session by ID
func (s *Server) GetSession(id string) *Session {
	s.Lock()
	defer s.Unlock()
	return s.sessions[id]
}

// Sessions returns all active sessions
func (s *Server) Sessions() map[string]*Session {
	s.Lock()
	defer s.Unlock()
	copy := make(map[string]*Session)
	for k, v := range s.sessions {
		copy[k] = v
	}
	return copy
}
