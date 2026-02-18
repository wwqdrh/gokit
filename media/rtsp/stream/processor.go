package stream

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"sync"
)

// StreamProcessorOptions contains options for stream processing
type StreamProcessorOptions struct {
	EnableScaling     bool
	TargetResolution  string
	EnableWatermark   bool
	WatermarkPath     string
	WatermarkPosition string
	EnableFiltering   bool
	FilterType        string
}

// StreamProcessor implements stream processing with optional features
type StreamProcessor struct {
	inputStream Streamer
	options     StreamProcessorOptions
	running     bool
	mu          sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	outputChan  chan []byte
	watermark   image.Image
}

// NewStreamProcessor creates a new stream processor
func NewStreamProcessor(input Streamer, options StreamProcessorOptions) *StreamProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	processor := &StreamProcessor{
		inputStream: input,
		options:     options,
		ctx:         ctx,
		cancel:      cancel,
		outputChan:  make(chan []byte, 1024),
	}

	// Load watermark if enabled
	if options.EnableWatermark && options.WatermarkPath != "" {
		if err := processor.loadWatermark(); err != nil {
			fmt.Printf("Failed to load watermark: %v\n", err)
		}
	}

	return processor
}

// Start starts the processor
func (p *StreamProcessor) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return nil
	}

	// Start processing loop
	go p.processLoop()

	p.running = true
	return nil
}

// Stop stops the processor
func (p *StreamProcessor) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.cancel()
	close(p.outputChan)
	p.running = false
	return nil
}

// IsRunning returns whether the processor is running
func (p *StreamProcessor) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

// GetStreamInfo returns stream information
func (p *StreamProcessor) GetStreamInfo() StreamInfo {
	return p.inputStream.GetStreamInfo()
}

// GetOutputChan returns the output channel
func (p *StreamProcessor) GetOutputChan() chan []byte {
	return p.outputChan
}

// loadWatermark loads watermark image
func (p *StreamProcessor) loadWatermark() error {
	file, err := os.Open(p.options.WatermarkPath)
	if err != nil {
		return fmt.Errorf("failed to open watermark file: %v", err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode watermark: %v", err)
	}

	p.watermark = img
	return nil
}

// processLoop processes the stream
func (p *StreamProcessor) processLoop() {
	if rtspStream, ok := p.inputStream.(*RTSPStream); ok {
		packetChan := rtspStream.GetPacketChan()
		for {
			select {
			case <-p.ctx.Done():
				return
			case packet, ok := <-packetChan:
				if !ok {
					return
				}

				// Process packet
				processedData := p.processPacket(packet)
				if len(processedData) > 0 {
					select {
					case p.outputChan <- processedData:
					default:
						// Channel full, drop packet
					}
				}
			}
		}
	}
}

// processPacket processes a single RTP packet
func (p *StreamProcessor) processPacket(packet RTPInfo) []byte {
	data := packet.Payload

	// Apply scaling if enabled
	if p.options.EnableScaling {
		data = p.applyScaling(data)
	}

	// Apply watermark if enabled
	if p.options.EnableWatermark && p.watermark != nil {
		data = p.applyWatermark(data)
	}

	// Apply filtering if enabled
	if p.options.EnableFiltering {
		data = p.applyFilter(data)
	}

	return data
}

// applyScaling applies resolution scaling
func (p *StreamProcessor) applyScaling(data []byte) []byte {
	// Simplified implementation
	// In real implementation, you would:
	// 1. Decode H264/NAL units
	// 2. Scale the frames
	// 3. Re-encode to H264
	
	// For now, just return original data
	fmt.Printf("Applying scaling to resolution: %s\n", p.options.TargetResolution)
	return data
}

// applyWatermark applies watermark to the video
func (p *StreamProcessor) applyWatermark(data []byte) []byte {
	// Simplified implementation
	// In real implementation, you would:
	// 1. Decode H264/NAL units
	// 2. Convert to RGB
	// 3. Draw watermark
	// 4. Re-encode to H264
	
	// For now, just return original data
	fmt.Printf("Applying watermark from: %s\n", p.options.WatermarkPath)
	return data
}

// applyFilter applies video filtering
func (p *StreamProcessor) applyFilter(data []byte) []byte {
	// Simplified implementation
	// In real implementation, you would apply various filters
	
	// For now, just return original data
	fmt.Printf("Applying filter: %s\n", p.options.FilterType)
	return data
}

// createWatermark creates a simple watermark
func (p *StreamProcessor) createWatermark(text string) image.Image {
	// Create a simple watermark with text
	// In real implementation, you would create a proper watermark
	bounds := image.Rect(0, 0, 200, 50)
	img := image.NewRGBA(bounds)

	// Fill with semi-transparent black
	black := color.RGBA{0, 0, 0, 128}
	draw.Draw(img, bounds, &image.Uniform{black}, image.Point{}, draw.Src)

	// TODO: Draw text on watermark

	return img
}

// GetOptions returns the processor options
func (p *StreamProcessor) GetOptions() StreamProcessorOptions {
	return p.options
}

// SetOptions updates the processor options
func (p *StreamProcessor) SetOptions(options StreamProcessorOptions) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.options = options

	// Reload watermark if needed
	if options.EnableWatermark && options.WatermarkPath != "" {
		if err := p.loadWatermark(); err != nil {
			return err
		}
	}

	return nil
}
