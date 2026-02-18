package main

import (
	"fmt"
	"log"

	"github.com/wwqdrh/gokit/media/rtsp/client"
)

func main() {
	// Create client
	c := client.NewClient()

	// Connect to server
	serverAddr := "localhost:554"
	fmt.Printf("Connecting to RTSP server at %s...\n", serverAddr)
	err := c.Connect(serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	// Send OPTIONS request
	fmt.Println("\nSending OPTIONS request...")
	response, err := c.Options()
	if err != nil {
		log.Fatalf("Failed to send OPTIONS: %v", err)
	}
	fmt.Printf("OPTIONS response: %d %s\n", response.StatusCode, response.StatusText)
	fmt.Printf("Public methods: %s\n", response.Header.Get("Public"))

	// Send DESCRIBE request
	fmt.Println("\nSending DESCRIBE request...")
	streamURI := "rtsp://localhost:554/test"
	response, err = c.Describe(streamURI)
	if err != nil {
		log.Fatalf("Failed to send DESCRIBE: %v", err)
	}
	fmt.Printf("DESCRIBE response: %d %s\n", response.StatusCode, response.StatusText)
	fmt.Printf("Content-Type: %s\n", response.Header.Get("Content-Type"))
	fmt.Printf("SDP:\n%s\n", string(response.Body))

	// Send SETUP request
	fmt.Println("\nSending SETUP request...")
	transport := "RTP/AVP;unicast;client_port=8000-8001"
	response, err = c.Setup(streamURI, transport)
	if err != nil {
		log.Fatalf("Failed to send SETUP: %v", err)
	}
	fmt.Printf("SETUP response: %d %s\n", response.StatusCode, response.StatusText)
	fmt.Printf("Session: %s\n", response.Header.Get("Session"))
	fmt.Printf("Transport: %s\n", response.Header.Get("Transport"))

	// Send PLAY request
	fmt.Println("\nSending PLAY request...")
	response, err = c.Play(streamURI)
	if err != nil {
		log.Fatalf("Failed to send PLAY: %v", err)
	}
	fmt.Printf("PLAY response: %d %s\n", response.StatusCode, response.StatusText)

	// Send PAUSE request
	fmt.Println("\nSending PAUSE request...")
	response, err = c.Pause(streamURI)
	if err != nil {
		log.Fatalf("Failed to send PAUSE: %v", err)
	}
	fmt.Printf("PAUSE response: %d %s\n", response.StatusCode, response.StatusText)

	// Send TEARDOWN request
	fmt.Println("\nSending TEARDOWN request...")
	response, err = c.Teardown(streamURI)
	if err != nil {
		log.Fatalf("Failed to send TEARDOWN: %v", err)
	}
	fmt.Printf("TEARDOWN response: %d %s\n", response.StatusCode, response.StatusText)

	fmt.Println("\nRTSP client example completed successfully!")
}
