package server

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/wwqdrh/gokit/media/rtsp"
)

// Handler defines the interface for RTSP request handlers
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
		Version:   "RTSP/1.0",
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
		Version:   "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	response.Header.Set("Content-Type", "application/sdp")
	// Example SDP
	sdp := `v=0
` +
		`o=- 12345 12345 IN IP4 127.0.0.1
` +
		`s=RTSP Server
` +
		`t=0 0
` +
		`m=video 0 RTP/AVP 96
` +
		`c=IN IP4 0.0.0.0
` +
		`a=rtpmap:96 H264/90000
`
	response.Body = []byte(sdp)
	return response, nil
}

// HandleSETUP handles SETUP requests
func (h *DefaultHandler) HandleSETUP(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:   "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	response.Header.Set("Session", session.ID)
	response.Header.Set("Transport", request.Header.Get("Transport"))
	return response, nil
}

// HandlePLAY handles PLAY requests
func (h *DefaultHandler) HandlePLAY(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:   "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	response.Header.Set("Session", session.ID)
	return response, nil
}

// HandlePAUSE handles PAUSE requests
func (h *DefaultHandler) HandlePAUSE(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:   "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	response.Header.Set("Session", session.ID)
	return response, nil
}

// HandleTEARDOWN handles TEARDOWN requests
func (h *DefaultHandler) HandleTEARDOWN(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:   "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	return response, nil
}

// HandleANNOUNCE handles ANNOUNCE requests
func (h *DefaultHandler) HandleANNOUNCE(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:   "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	return response, nil
}

// HandleRECORD handles RECORD requests
func (h *DefaultHandler) HandleRECORD(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:   "RTSP/1.0",
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
		Version:   "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	return response, nil
}

// HandleSET_PARAMETER handles SET_PARAMETER requests
func (h *DefaultHandler) HandleSET_PARAMETER(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:   "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	return response, nil
}

// HandleREDIRECT handles REDIRECT requests
func (h *DefaultHandler) HandleREDIRECT(session *Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:   "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	return response, nil
}

// Session represents an RTSP session
type Session struct {
	ID        string
	Conn      net.Conn
	CSeq      int
	Transport string
	CreatedAt time.Time
	LastActive time.Time
}

// Server represents an RTSP server
type Server struct {
	addr     string
	listener net.Listener
	handler  Handler
	sessions map[string]*Session
	sync.Mutex
	running  bool
}

// NewServer creates a new RTSP server
func NewServer(addr string) *Server {
	return &Server{
		addr:     addr,
		handler:  &DefaultHandler{},
		sessions: make(map[string]*Session),
	}
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
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Conn:      conn,
		CSeq:      0,
		CreatedAt: time.Now(),
		LastActive: time.Now(),
	}

	// Add session
	s.Lock()
	s.sessions[session.ID] = session
	s.Unlock()

	defer func() {
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
				Version:   "RTSP/1.0",
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
			Version:   "RTSP/1.0",
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
