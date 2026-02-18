package rtsp

import (
	"bufio"
	"bytes"
	"testing"
)

func TestMessageSerialization(t *testing.T) {
	// Test request serialization
	request := &Request{
		Method:  MethodOPTIONS,
		URI:     "rtsp://localhost:554/test",
		Version: "RTSP/1.0",
		Header:  make(Header),
		Body:    nil,
	}
	request.Header.Set("CSeq", "1")
	request.Header.Set("User-Agent", "GoRTSP/1.0")

	data, err := SerializeRequest(request)
	if err != nil {
		t.Fatalf("Failed to serialize request: %v", err)
	}

	// Test request parsing
	reader := bufio.NewReader(bytes.NewReader(data))
	parsedRequest, err := ParseRequest(reader)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if parsedRequest.Method != request.Method {
		t.Errorf("Expected method %s, got %s", request.Method, parsedRequest.Method)
	}
	if parsedRequest.URI != request.URI {
		t.Errorf("Expected URI %s, got %s", request.URI, parsedRequest.URI)
	}
	if parsedRequest.Version != request.Version {
		t.Errorf("Expected version %s, got %s", request.Version, parsedRequest.Version)
	}
	if parsedRequest.Header.Get("CSeq") != request.Header.Get("CSeq") {
		t.Errorf("Expected CSeq %s, got %s", request.Header.Get("CSeq"), parsedRequest.Header.Get("CSeq"))
	}
}

func TestResponseSerialization(t *testing.T) {
	// Test response serialization
	response := &Response{
		Version:    "RTSP/1.0",
		StatusCode: StatusOK,
		StatusText: "OK",
		Header:     make(Header),
		Body:       nil,
	}
	response.Header.Set("CSeq", "1")
	response.Header.Set("Public", "OPTIONS, DESCRIBE, SETUP, PLAY, PAUSE, TEARDOWN")

	data, err := SerializeResponse(response)
	if err != nil {
		t.Fatalf("Failed to serialize response: %v", err)
	}

	// Test response parsing
	reader := bufio.NewReader(bytes.NewReader(data))
	parsedResponse, err := ParseResponse(reader)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if parsedResponse.Version != response.Version {
		t.Errorf("Expected version %s, got %s", response.Version, parsedResponse.Version)
	}
	if parsedResponse.StatusCode != response.StatusCode {
		t.Errorf("Expected status code %d, got %d", response.StatusCode, parsedResponse.StatusCode)
	}
	if parsedResponse.StatusText != response.StatusText {
		t.Errorf("Expected status text %s, got %s", response.StatusText, parsedResponse.StatusText)
	}
	if parsedResponse.Header.Get("Public") != response.Header.Get("Public") {
		t.Errorf("Expected Public %s, got %s", response.Header.Get("Public"), parsedResponse.Header.Get("Public"))
	}
}

func TestStatusText(t *testing.T) {
	// Test status text retrieval
	testCases := []struct {
		code     StatusCode
		expected string
	}{
		{StatusOK, "OK"},
		{StatusNotFound, "Not Found"},
		{StatusInternalServerError, "Internal Server Error"},
		{StatusMethodNotAllowed, "Method Not Allowed"},
	}

	for _, tc := range testCases {
		got := StatusText(tc.code)
		if got != tc.expected {
			t.Errorf("For code %d, expected status text %s, got %s", tc.code, tc.expected, got)
		}
	}
}

func TestHeaderOperations(t *testing.T) {
	header := make(Header)

	// Test Set
	header.Set("CSeq", "1")
	if header.Get("CSeq") != "1" {
		t.Errorf("Expected CSeq 1, got %s", header.Get("CSeq"))
	}

	// Test Add
	header.Add("Accept", "application/sdp")
	header.Add("Accept", "application/rtspl")
	if len(header["Accept"]) != 2 {
		t.Errorf("Expected 2 Accept values, got %d", len(header["Accept"]))
	}

	// Test Get (should return first value)
	if header.Get("Accept") != "application/sdp" {
		t.Errorf("Expected first Accept value to be application/sdp, got %s", header.Get("Accept"))
	}
}
