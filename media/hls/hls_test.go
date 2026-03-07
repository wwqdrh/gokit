package hls

import (
	"testing"
	"time"
)

// TestNewHLSManager 测试创建 HLS 管理器
func TestNewHLSManager(t *testing.T) {
	manager := NewHLSManager()
	if manager == nil {
		t.Errorf("NewHLSManager() returned nil")
	}
}

// TestGetHLSURL 测试获取 HLS URL
func TestGetHLSURL(t *testing.T) {
	manager := NewHLSManager()

	streamName := "test_sample"
	expectedURL := "http://localhost:7002/live/test_sample.m3u8"

	url := manager.GetHLSURL(streamName)
	if url != expectedURL {
		t.Errorf("GetHLSURL() returned %s, expected %s", url, expectedURL)
	}
}

// TestStartServer 测试启动服务器
func TestStartServer(t *testing.T) {
	manager := NewHLSManager()

	err := manager.StartServer(":8081", true)
	if err != nil {
		t.Errorf("StartServer() failed: %v", err)
		return
	}

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 停止服务器
	err = manager.StopServer()
	if err != nil {
		t.Errorf("StopServer() failed: %v", err)
		return
	}
}
