package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wwqdrh/gokit/media/rtsp"
	"github.com/wwqdrh/gokit/media/rtsp/server"
)

// CustomHandler implements a custom RTSP request handler
type CustomHandler struct {
	server.DefaultHandler
}

// HandleDESCRIBE overrides the default DESCRIBE handler
func (h *CustomHandler) HandleDESCRIBE(session *server.Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:   "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	response.Header.Set("Content-Type", "application/sdp")
	
	// Custom SDP for our test stream
	sdp := `v=0
` +
		`o=- 12345 12345 IN IP4 127.0.0.1
` +
		`s=GoRTSP Test Server
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
	
	fmt.Printf("Custom DESCRIBE handler called for URI: %s\n", request.URI)
	return response, nil
}

// HandlePLAY overrides the default PLAY handler
func (h *CustomHandler) HandlePLAY(session *server.Session, request *rtsp.Request) (*rtsp.Response, error) {
	response := &rtsp.Response{
		Version:   "RTSP/1.0",
		StatusCode: rtsp.StatusOK,
		StatusText: rtsp.StatusText(rtsp.StatusOK),
		Header:     make(rtsp.Header),
	}
	response.Header.Set("Session", session.ID)
	response.Header.Set("RTP-Info", "url=rtsp://localhost:554/test;seq=0;rtptime=0")
	
	fmt.Printf("Custom PLAY handler called for session: %s\n", session.ID)
	return response, nil
}

func main() {
	// Create server
	serverAddr := "0.0.0.0:554"
	s := server.NewServer(serverAddr)

	// Set custom handler
	customHandler := &CustomHandler{}
	s.SetHandler(customHandler)

	// Start server
	fmt.Printf("Starting RTSP server at %s...\n", serverAddr)
	err := s.Start()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer s.Stop()

	fmt.Println("RTSP server started successfully!")
	fmt.Println("Supported methods: OPTIONS, DESCRIBE, SETUP, PLAY, PAUSE, TEARDOWN, ANNOUNCE, RECORD, GET_PARAMETER, SET_PARAMETER")
	fmt.Println("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nStopping RTSP server...")
	s.Stop()
	fmt.Println("RTSP server stopped successfully!")
}
