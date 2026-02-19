package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wwqdrh/gokit/media/rtsp/client"
	"github.com/wwqdrh/gokit/media/rtsp/stream"
)

// StreamInfo represents information about an available stream
type StreamInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func main() {
	// Create HTTP server
	http.HandleFunc("/api/streams", handleGetStreams)
	http.HandleFunc("/api/stream/", handleStream)
	http.HandleFunc("/", handleIndex)

	// Start HTTP server
	serverAddr := ":8080"
	fmt.Printf("Starting HTTP server at %s...\n", serverAddr)
	fmt.Printf("Available endpoints:\n")
	fmt.Printf("  - GET /api/streams - List available streams\n")
	fmt.Printf("  - GET /api/stream/{name} - Get FLV stream for a video\n")
	fmt.Printf("  - GET / - Frontend page for playing streams\n")

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// handleGetStreams handles the /api/streams endpoint
func handleGetStreams(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Get available streams from video directory
	videoDir := filepath.Join(".", "..", "server", "video")
	files, err := os.ReadDir(videoDir)
	if err != nil {
		log.Printf("Failed to read video directory: %v", err)
		http.Error(w, "Failed to get available streams", http.StatusInternalServerError)
		return
	}

	// Collect stream information
	streams := make([]StreamInfo, 0)
	for _, file := range files {
		if !file.IsDir() {
			name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			streams = append(streams, StreamInfo{
				Name: name,
				Path: file.Name(),
			})
		}
	}

	// Send JSON response
	fmt.Fprintf(w, "[")
	for i, stream := range streams {
		if i > 0 {
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, `{"name":"%s","path":"%s"}`, stream.Name, stream.Path)
	}
	fmt.Fprintf(w, "]")
}

// handleStream handles the /api/stream/{name} endpoint
func handleStream(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get stream name from URL
	streamName := strings.TrimPrefix(r.URL.Path, "/api/stream/")
	if streamName == "" {
		http.Error(w, "Stream name required", http.StatusBadRequest)
		return
	}

	// Set content type for FLV stream
	w.Header().Set("Content-Type", "video/x-flv")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Create RTSP client
	c := client.NewClient()
	defer c.Close()

	// Connect to RTSP server
	serverAddr := "localhost:554"
	if err := c.Connect(serverAddr); err != nil {
		log.Printf("Failed to connect to RTSP server: %v", err)
		http.Error(w, "Failed to connect to RTSP server", http.StatusInternalServerError)
		return
	}

	// Send DESCRIBE request
	streamURI := "rtsp://localhost:554/" + streamName
	_, err := c.Describe(streamURI)
	if err != nil {
		log.Printf("Failed to send DESCRIBE: %v", err)
		http.Error(w, "Failed to get stream information", http.StatusInternalServerError)
		return
	}

	// Send SETUP request
	transport := "RTP/AVP;unicast;client_port=8000-8001"
	_, err = c.Setup(streamURI, transport)
	if err != nil {
		log.Printf("Failed to send SETUP: %v", err)
		http.Error(w, "Failed to set up stream", http.StatusInternalServerError)
		return
	}

	// Send PLAY request
	_, err = c.Play(streamURI)
	if err != nil {
		log.Printf("Failed to send PLAY: %v", err)
		http.Error(w, "Failed to start stream", http.StatusInternalServerError)
		return
	}

	// Create RTSP streamer
	config := stream.RTSPConfig{
		URL: streamURI,
	}
	rtspStream := stream.NewRTSPStream(config)

	// Start streamer
	if err := rtspStream.Start(); err != nil {
		log.Printf("Failed to start RTSP streamer: %v", err)
		http.Error(w, "Failed to start streamer", http.StatusInternalServerError)
		return
	}
	defer rtspStream.Stop()

	// Simulate FLV streaming
	// Note: In a real implementation, we would need to:
	// 1. Receive RTP packets from the RTSP stream
	// 2. Convert them to FLV format
	// 3. Send the FLV data to the client

	// For demonstration purposes, send a simple FLV header
	flvHeader := []byte{0x46, 0x4C, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09}
	w.Write(flvHeader)

	// Send a metadata tag
	metadataTag := []byte{0x00, 0x00, 0x00, 0x12, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	w.Write(metadataTag)

	// Keep connection open
	select {}
}

// handleIndex handles the root endpoint (frontend page)
func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "text/html")

	// Send HTML response
	html := `<!DOCTYPE html>
<html>
<head>
	<title>RTSP to FLV Stream Player</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			max-width: 1200px;
			margin: 0 auto;
			padding: 20px;
		}
		.stream-list {
			margin-bottom: 20px;
		}
		.stream-item {
			padding: 10px;
			border: 1px solid #ddd;
			margin-bottom: 10px;
			cursor: pointer;
			border-radius: 4px;
		}
		.stream-item:hover {
			background-color: #f5f5f5;
		}
		.player-container {
			margin-top: 20px;
		}
		video {
			width: 100%;
			height: auto;
			border: 1px solid #ddd;
			border-radius: 4px;
		}
	</style>
</head>
<body>
	<h1>RTSP to FLV Stream Player</h1>
	
	<div class="stream-list">
		<h2>Available Streams</h2>
		<div id="streamList"></div>
	</div>
	
	<div class="player-container">
		<h2>Stream Player</h2>
		<video id="videoElement" controls autoplay muted style="width: 100%; height: auto;"></video>
	</div>
	
	<script src="https://cdn.jsdelivr.net/npm/flv.js@1.6.2/dist/flv.min.js"></script>
	<script>
		let flvPlayer = null;
		
		// Load available streams
		async function loadStreams() {
			try {
				const response = await fetch('/api/streams');
				const streams = await response.json();
				
				const streamList = document.getElementById('streamList');
				streamList.innerHTML = '';
				
				streams.forEach(stream => {
					const streamItem = document.createElement('div');
					streamItem.className = 'stream-item';
					streamItem.textContent = stream.name;
					streamItem.addEventListener('click', () => playStream(stream.name));
					streamList.appendChild(streamItem);
				});
			} catch (error) {
				console.error('Failed to load streams:', error);
			}
		}
		
		// Play a stream
		function playStream(streamName) {
			const videoElement = document.getElementById('videoElement');
			const flvUrl = "/api/stream/" + streamName;
			
			// Clean up existing player
			if (flvPlayer) {
				flvPlayer.pause();
				flvPlayer.unload();
				flvPlayer.detachMediaElement();
				flvPlayer.destroy();
				flvPlayer = null;
			}
			
			// Create new player
			if (flvjs.isSupported()) {
				flvPlayer = flvjs.createPlayer({
					type: 'flv',
					isLive: true,
					url: flvUrl,
					cors: true
				}, {
					enableStashBuffer: false,
					autoCleanupSourceBuffer: true
				});
				
				flvPlayer.attachMediaElement(videoElement);
				flvPlayer.load();
				flvPlayer.play().catch(error => {
					console.error('Auto play failed:', error);
				});
			} else {
				console.error('Current browser does not support flv.js');
			}
		}
		
		// Load streams on page load
		window.onload = loadStreams;
	</script>
</body>
</html>`

	fmt.Fprint(w, html)
}
