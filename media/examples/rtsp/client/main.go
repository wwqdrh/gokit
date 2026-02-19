package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	// Create RTSP streamer
	config := stream.RTSPConfig{
		URL: "rtsp://localhost:554/" + streamName,
	}
	rtspStream := stream.NewRTSPStream(config)

	// Start streamer
	if err := rtspStream.Start(); err != nil {
		log.Printf("Failed to start RTSP streamer: %v", err)
		http.Error(w, "Failed to start streamer", http.StatusInternalServerError)
		return
	}
	defer rtspStream.Stop()

	// Send FLV header
	flvHeader := []byte{0x46, 0x4C, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09}
	w.Write(flvHeader)

	// Send metadata tag
	sendMetadataTag(w)

	// Get RTP packet channel
	packetChan := rtspStream.GetPacketChan()

	// Process RTP packets and convert to FLV
	processRTPPackets(w, packetChan)
}

// sendMetadataTag sends FLV metadata tag
func sendMetadataTag(w http.ResponseWriter) {
	// Create metadata tag
	// Simplified metadata for demonstration
	metadata := `{"encoder":"GoRTSP-FLV-Converter","width":1920,"height":1080,"framerate":25,"videocodecid":7,"audiocodecid":10}`

	// Calculate tag size
	tagSize := len(metadata) + 1

	// Create tag header
	tagHeader := make([]byte, 11)
	tagHeader[0] = 0x12 // Script data tag

	// Set tag size (big-endian)
	tagHeader[1] = byte(tagSize >> 16)
	tagHeader[2] = byte(tagSize >> 8)
	tagHeader[3] = byte(tagSize)

	// Set timestamp (0 for metadata)
	tagHeader[4] = 0
	tagHeader[5] = 0
	tagHeader[6] = 0
	tagHeader[7] = 0

	// Set stream ID (0)
	tagHeader[8] = 0
	tagHeader[9] = 0
	tagHeader[10] = 0

	// Create tag body
	tagBody := make([]byte, 0, len(tagHeader)+len(metadata)+1+4)
	tagBody = append(tagBody, tagHeader...)
	tagBody = append(tagBody, 0x02) // String type

	// Add string length (big-endian)
	strLen := len(metadata)
	tagBody = append(tagBody, byte(strLen>>8), byte(strLen))
	tagBody = append(tagBody, []byte(metadata)...)

	// Add tag size at the end (big-endian)
	totalSize := len(tagHeader) + len(tagBody) - len(tagHeader)
	tagBody = append(tagBody, byte(totalSize>>16), byte(totalSize>>8), byte(totalSize))

	// Write metadata tag
	w.Write(tagBody)
}

// processRTPPackets processes RTP packets and converts to FLV
func processRTPPackets(w http.ResponseWriter, packetChan chan stream.RTPInfo) {
	var lastTimestamp uint32
	var sequenceNum uint16

	for {
		select {
		case packet, ok := <-packetChan:
			if !ok {
				// Channel closed
				return
			}

			// Process H.264 video packets
			if packet.PayloadType == 96 { // H.264
				// Calculate timestamp delta
				timestampDelta := uint32(0)
				if lastTimestamp > 0 {
					timestampDelta = packet.Timestamp - lastTimestamp
				}
				lastTimestamp = packet.Timestamp

				// Create FLV video tag
				flvTag := createFLVVideoTag(packet.Payload, packet.Marker, timestampDelta, sequenceNum)
				sequenceNum++

				// Write FLV tag
				w.Write(flvTag)

				// Flush response to client
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			}
		}
	}
}

// createFLVVideoTag creates a FLV video tag from H.264 payload
func createFLVVideoTag(payload []byte, isKeyFrame bool, timestampDelta uint32, sequenceNum uint16) []byte {
	// Calculate tag size
	tagSize := len(payload) + 5 // 5 bytes for video header

	// Create tag header
	tagHeader := make([]byte, 11)
	tagHeader[0] = 0x09 // Video tag

	// Set tag size (big-endian)
	tagHeader[1] = byte(tagSize >> 16)
	tagHeader[2] = byte(tagSize >> 8)
	tagHeader[3] = byte(tagSize)

	// Set timestamp (big-endian)
	tagHeader[4] = byte(timestampDelta >> 16)
	tagHeader[5] = byte(timestampDelta >> 8)
	tagHeader[6] = byte(timestampDelta)
	tagHeader[7] = byte(timestampDelta >> 24) // Extended timestamp

	// Set stream ID (0)
	tagHeader[8] = 0
	tagHeader[9] = 0
	tagHeader[10] = 0

	// Create video header
	videoHeader := make([]byte, 5)

	// Set frame type and codec ID
	if isKeyFrame {
		videoHeader[0] = 0x17 // Key frame + H.264
	} else {
		videoHeader[0] = 0x27 // Inter frame + H.264
	}

	// Set AVC packet type
	videoHeader[1] = 0x01 // AVC NALU

	// Set composition time (0 for simplicity)
	videoHeader[2] = 0
	videoHeader[3] = 0
	videoHeader[4] = 0

	// Create tag body
	tagBody := make([]byte, 0, len(tagHeader)+len(videoHeader)+len(payload)+4)
	tagBody = append(tagBody, tagHeader...)
	tagBody = append(tagBody, videoHeader...)
	tagBody = append(tagBody, payload...)

	// Add tag size at the end (big-endian)
	totalSize := len(tagHeader) + len(videoHeader) + len(payload)
	tagBody = append(tagBody, byte(totalSize>>16), byte(totalSize>>8), byte(totalSize))

	return tagBody
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
