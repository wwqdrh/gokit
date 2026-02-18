package rtsp

import (
	"bufio"
	"bytes"
	"fmt"
	"net/textproto"
	"strconv"
	"strings"
)

// ParseRequest parses an RTSP request from a reader
func ParseRequest(reader *bufio.Reader) (*Request, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", line)
	}

	request := &Request{
		Method:  Method(parts[0]),
		URI:     parts[1],
		Version: parts[2],
		Header:  make(Header),
	}

	// Parse headers
	textReader := textproto.NewReader(reader)
	header, err := textReader.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	// Convert textproto.MIMEHeader to Header
	for key, values := range header {
		// textproto.MIMEHeader already has normalized keys
		request.Header[key] = values
	}
	
	// Debug: print headers
	// fmt.Printf("Parsed headers: %v\n", request.Header)

	// Parse body if present
	if contentLength := request.Header.Get("Content-Length"); contentLength != "" {
		length, err := strconv.Atoi(contentLength)
		if err == nil && length > 0 {
			request.Body = make([]byte, length)
			_, err = reader.Read(request.Body)
			if err != nil {
				return nil, err
			}
		}
	}

	return request, nil
}

// ParseResponse parses an RTSP response from a reader
func ParseResponse(reader *bufio.Reader) (*Response, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid response line: %s", line)
	}

	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %s", parts[1])
	}

	response := &Response{
		Version:   parts[0],
		StatusCode: StatusCode(statusCode),
		StatusText: strings.Join(parts[2:], " "),
		Header:     make(Header),
	}

	// Parse headers
	textReader := textproto.NewReader(reader)
	header, err := textReader.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	// Convert textproto.MIMEHeader to Header
	for key, values := range header {
		response.Header[key] = values
	}

	// Parse body if present
	if contentLength := response.Header.Get("Content-Length"); contentLength != "" {
		length, err := strconv.Atoi(contentLength)
		if err == nil && length > 0 {
			response.Body = make([]byte, length)
			_, err = reader.Read(response.Body)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, nil
}

// SerializeRequest serializes an RTSP request to bytes
func SerializeRequest(request *Request) ([]byte, error) {
	var buf bytes.Buffer

	// Write request line
	fmt.Fprintf(&buf, "%s %s %s\r\n", request.Method, request.URI, request.Version)

	// Write headers
	for key, values := range request.Header {
		for _, value := range values {
			fmt.Fprintf(&buf, "%s: %s\r\n", key, value)
		}
	}

	// Write Content-Length if body is present
	if len(request.Body) > 0 {
		fmt.Fprintf(&buf, "Content-Length: %d\r\n", len(request.Body))
	}

	// Write body separator
	buf.WriteString("\r\n")

	// Write body
	buf.Write(request.Body)

	return buf.Bytes(), nil
}

// SerializeResponse serializes an RTSP response to bytes
func SerializeResponse(response *Response) ([]byte, error) {
	var buf bytes.Buffer

	// Write response line
	fmt.Fprintf(&buf, "%s %d %s\r\n", response.Version, response.StatusCode, response.StatusText)

	// Write headers
	for key, values := range response.Header {
		for _, value := range values {
			fmt.Fprintf(&buf, "%s: %s\r\n", key, value)
		}
	}

	// Write Content-Length if body is present
	if len(response.Body) > 0 {
		fmt.Fprintf(&buf, "Content-Length: %d\r\n", len(response.Body))
	}

	// Write body separator
	buf.WriteString("\r\n")

	// Write body
	buf.Write(response.Body)

	return buf.Bytes(), nil
}
