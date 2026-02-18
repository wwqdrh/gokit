package rtsp

import (
	"net"
	"strconv"
	"strings"
)

// MessageType defines RTSP message types
type MessageType int

const (
	MessageTypeRequest MessageType = iota
	MessageTypeResponse
)

// Method defines RTSP methods
type Method string

const (
	MethodDESCRIBE   Method = "DESCRIBE"
	MethodANNOUNCE   Method = "ANNOUNCE"
	MethodGET_PARAMETER Method = "GET_PARAMETER"
	MethodOPTIONS    Method = "OPTIONS"
	MethodPAUSE      Method = "PAUSE"
	MethodPLAY       Method = "PLAY"
	MethodRECORD     Method = "RECORD"
	MethodREDIRECT   Method = "REDIRECT"
	MethodSETUP      Method = "SETUP"
	MethodSET_PARAMETER Method = "SET_PARAMETER"
	MethodTEARDOWN   Method = "TEARDOWN"
)

// StatusCode defines RTSP status codes
type StatusCode int

const (
	StatusContinue            StatusCode = 100
	StatusOK                  StatusCode = 200
	StatusCreated             StatusCode = 201
	StatusAccepted            StatusCode = 202
	StatusNonAuthoritative    StatusCode = 203
	StatusNoContent           StatusCode = 204
	StatusResetContent        StatusCode = 205
	StatusPartialContent      StatusCode = 206
	StatusMultipleChoices     StatusCode = 300
	StatusMovedPermanently    StatusCode = 301
	StatusFound               StatusCode = 302
	StatusSeeOther            StatusCode = 303
	StatusNotModified         StatusCode = 304
	StatusUseProxy            StatusCode = 305
	StatusTemporaryRedirect   StatusCode = 307
	StatusBadRequest          StatusCode = 400
	StatusUnauthorized        StatusCode = 401
	StatusPaymentRequired     StatusCode = 402
	StatusForbidden           StatusCode = 403
	StatusNotFound            StatusCode = 404
	StatusMethodNotAllowed    StatusCode = 405
	StatusNotAcceptable       StatusCode = 406
	StatusProxyAuthRequired   StatusCode = 407
	StatusRequestTimeout      StatusCode = 408
	StatusConflict            StatusCode = 409
	StatusGone                StatusCode = 410
	StatusLengthRequired      StatusCode = 411
	StatusPreconditionFailed  StatusCode = 412
	StatusRequestEntityTooLarge StatusCode = 413
	StatusRequestURITooLong   StatusCode = 414
	StatusUnsupportedMediaType StatusCode = 415
	StatusRangeNotSatisfiable StatusCode = 416
	StatusExpectationFailed   StatusCode = 417
	StatusUnsupportedTransport StatusCode = 461
	StatusInternalServerError StatusCode = 500
	StatusNotImplemented      StatusCode = 501
	StatusBadGateway          StatusCode = 502
	StatusServiceUnavailable  StatusCode = 503
	StatusGatewayTimeout      StatusCode = 504
	StatusRTSPVersionNotSupported StatusCode = 505
	StatusOptionNotSupported  StatusCode = 551
)

// StatusText returns the text representation of a status code
func StatusText(code StatusCode) string {
	switch code {
	case StatusContinue:
		return "Continue"
	case StatusOK:
		return "OK"
	case StatusCreated:
		return "Created"
	case StatusAccepted:
		return "Accepted"
	case StatusNonAuthoritative:
		return "Non-Authoritative Information"
	case StatusNoContent:
		return "No Content"
	case StatusResetContent:
		return "Reset Content"
	case StatusPartialContent:
		return "Partial Content"
	case StatusMultipleChoices:
		return "Multiple Choices"
	case StatusMovedPermanently:
		return "Moved Permanently"
	case StatusFound:
		return "Found"
	case StatusSeeOther:
		return "See Other"
	case StatusNotModified:
		return "Not Modified"
	case StatusUseProxy:
		return "Use Proxy"
	case StatusTemporaryRedirect:
		return "Temporary Redirect"
	case StatusBadRequest:
		return "Bad Request"
	case StatusUnauthorized:
		return "Unauthorized"
	case StatusPaymentRequired:
		return "Payment Required"
	case StatusForbidden:
		return "Forbidden"
	case StatusNotFound:
		return "Not Found"
	case StatusMethodNotAllowed:
		return "Method Not Allowed"
	case StatusNotAcceptable:
		return "Not Acceptable"
	case StatusProxyAuthRequired:
		return "Proxy Authentication Required"
	case StatusRequestTimeout:
		return "Request Timeout"
	case StatusConflict:
		return "Conflict"
	case StatusGone:
		return "Gone"
	case StatusLengthRequired:
		return "Length Required"
	case StatusPreconditionFailed:
		return "Precondition Failed"
	case StatusRequestEntityTooLarge:
		return "Request Entity Too Large"
	case StatusRequestURITooLong:
		return "Request-URI Too Long"
	case StatusUnsupportedMediaType:
		return "Unsupported Media Type"
	case StatusRangeNotSatisfiable:
		return "Range Not Satisfiable"
	case StatusExpectationFailed:
		return "Expectation Failed"
	case StatusUnsupportedTransport:
		return "Unsupported Transport"
	case StatusInternalServerError:
		return "Internal Server Error"
	case StatusNotImplemented:
		return "Not Implemented"
	case StatusBadGateway:
		return "Bad Gateway"
	case StatusServiceUnavailable:
		return "Service Unavailable"
	case StatusGatewayTimeout:
		return "Gateway Timeout"
	case StatusRTSPVersionNotSupported:
		return "RTSP Version Not Supported"
	case StatusOptionNotSupported:
		return "Option Not Supported"
	default:
		return "Unknown Status"
	}
}

// Header represents an RTSP header
type Header map[string][]string

// Get returns the first value for the given key (case-insensitive)
func (h Header) Get(key string) string {
	// First try exact match
	if values, ok := h[key]; ok && len(values) > 0 {
		return values[0]
	}
	
	// Try case-insensitive match
	lowerKey := strings.ToLower(key)
	for k, values := range h {
		if strings.ToLower(k) == lowerKey && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

// Set sets the value for the given key
func (h Header) Set(key, value string) {
	h[strings.Title(key)] = []string{value}
}

// Add adds a value for the given key
func (h Header) Add(key, value string) {
	titleKey := strings.Title(key)
	h[titleKey] = append(h[titleKey], value)
}

// Message represents an RTSP message
type Message struct {
	Type      MessageType
	Method    Method
	URI       string
	Version   string
	StatusCode StatusCode
	StatusText string
	Header    Header
	Body      []byte
}

// Request represents an RTSP request
type Request struct {
	Method  Method
	URI     string
	Version string
	Header  Header
	Body    []byte
}

// Response represents an RTSP response
type Response struct {
	Version   string
	StatusCode StatusCode
	StatusText string
	Header    Header
	Body      []byte
}

// Session represents an RTSP session
type Session struct {
	ID        string
	Conn      net.Conn
	CSeq      int
	Transport string
}

// ParseCSeq parses CSeq header value
func ParseCSeq(header Header) (int, error) {
	cseqStr := header.Get("CSeq")
	if cseqStr == "" {
		return 0, nil
	}
	return strconv.Atoi(cseqStr)
}

// BuildCSeq builds CSeq header value
func BuildCSeq(cseq int) string {
	return strconv.Itoa(cseq)
}

// DefaultPort returns the default RTSP port
func DefaultPort() int {
	return 554
}

// IsDefaultPort checks if the port is the default RTSP port
func IsDefaultPort(port int) bool {
	return port == DefaultPort()
}
