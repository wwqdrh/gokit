package client

import (
	"bufio"
	"fmt"
	"net"

	"github.com/wwqdrh/gokit/media/rtsp"
)

// Client represents an RTSP client
type Client struct {
	conn       net.Conn
	cseq       int
	sessionID  string
	baseURI    string
	transport  string
	bufferedReader *bufio.Reader
}

// NewClient creates a new RTSP client
func NewClient() *Client {
	return &Client{
		cseq: 1,
	}
}

// Connect establishes a connection to the RTSP server
func (c *Client) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c.conn = conn
	c.bufferedReader = bufio.NewReader(conn)
	c.baseURI = fmt.Sprintf("rtsp://%s", addr)
	return nil
}

// Close closes the connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendRequest sends an RTSP request and returns the response
func (c *Client) SendRequest(method rtsp.Method, uri string, header rtsp.Header, body []byte) (*rtsp.Response, error) {
	// Create request
	request := &rtsp.Request{
		Method:  method,
		URI:     uri,
		Version: "RTSP/1.0",
		Header:  make(rtsp.Header),
		Body:    body,
	}

	// Set default headers
	request.Header.Set("CSeq", rtsp.BuildCSeq(c.cseq))
	request.Header.Set("User-Agent", "GoRTSP/1.0")

	// Add session ID if available
	if c.sessionID != "" {
		request.Header.Set("Session", c.sessionID)
	}

	// Add user-provided headers
	for key, values := range header {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	// Serialize and send request
	data, err := rtsp.SerializeRequest(request)
	if err != nil {
		return nil, err
	}

	_, err = c.conn.Write(data)
	if err != nil {
		return nil, err
	}

	// Increment CSeq
	c.cseq++

	// Read and parse response
	response, err := rtsp.ParseResponse(c.bufferedReader)
	if err != nil {
		return nil, err
	}

	// Update session ID if present
	if sessionID := response.Header.Get("Session"); sessionID != "" {
		c.sessionID = sessionID
	}

	return response, nil
}

// Options sends an OPTIONS request
func (c *Client) Options() (*rtsp.Response, error) {
	return c.SendRequest(rtsp.MethodOPTIONS, c.baseURI, nil, nil)
}

// Describe sends a DESCRIBE request
func (c *Client) Describe(uri string) (*rtsp.Response, error) {
	header := make(rtsp.Header)
	header.Set("Accept", "application/sdp")
	return c.SendRequest(rtsp.MethodDESCRIBE, uri, header, nil)
}

// Setup sends a SETUP request
func (c *Client) Setup(uri string, transport string) (*rtsp.Response, error) {
	header := make(rtsp.Header)
	header.Set("Transport", transport)
	response, err := c.SendRequest(rtsp.MethodSETUP, uri, header, nil)
	if err == nil {
		c.transport = transport
	}
	return response, nil
}

// Play sends a PLAY request
func (c *Client) Play(uri string) (*rtsp.Response, error) {
	return c.SendRequest(rtsp.MethodPLAY, uri, nil, nil)
}

// Pause sends a PAUSE request
func (c *Client) Pause(uri string) (*rtsp.Response, error) {
	return c.SendRequest(rtsp.MethodPAUSE, uri, nil, nil)
}

// Teardown sends a TEARDOWN request
func (c *Client) Teardown(uri string) (*rtsp.Response, error) {
	response, err := c.SendRequest(rtsp.MethodTEARDOWN, uri, nil, nil)
	if err == nil {
		c.sessionID = ""
	}
	return response, nil
}

// GetParameter sends a GET_PARAMETER request
func (c *Client) GetParameter(uri string, body []byte) (*rtsp.Response, error) {
	return c.SendRequest(rtsp.MethodGET_PARAMETER, uri, nil, body)
}

// SetParameter sends a SET_PARAMETER request
func (c *Client) SetParameter(uri string, body []byte) (*rtsp.Response, error) {
	return c.SendRequest(rtsp.MethodSET_PARAMETER, uri, nil, body)
}

// Announce sends an ANNOUNCE request
func (c *Client) Announce(uri string, body []byte) (*rtsp.Response, error) {
	header := make(rtsp.Header)
	header.Set("Content-Type", "application/sdp")
	return c.SendRequest(rtsp.MethodANNOUNCE, uri, header, body)
}

// Record sends a RECORD request
func (c *Client) Record(uri string) (*rtsp.Response, error) {
	return c.SendRequest(rtsp.MethodRECORD, uri, nil, nil)
}

// SessionID returns the current session ID
func (c *Client) SessionID() string {
	return c.sessionID
}

// CSeq returns the current CSeq value
func (c *Client) CSeq() int {
	return c.cseq
}

// Transport returns the current transport
func (c *Client) Transport() string {
	return c.transport
}
