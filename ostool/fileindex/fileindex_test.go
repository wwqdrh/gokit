package fileindex

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestFileInfoTree(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建 FileInfoTree
	tree := NewFileInfoTree(tempDir, 1)

	// 设置回调函数
	var updateCalled int64 = 0
	tree.SetOnFileInfoUpdate(func(fi FileIndex) {
		atomic.AddInt64(&updateCalled, 1)
	})

	// 启动更新
	tree.Start()
	defer tree.Stop()

	// 创建测试文件
	testFilePath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFilePath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// 等待一段时间
	time.Sleep(3 * time.Second)
	// 检查文件是否被正确索引
	fi := tree.Get(testFilePath)
	if fi.Size == 0 {
		t.Errorf("expected non-zero file size, got %d", fi.Size)
	}
	if updateCalled != 1 {
		t.Errorf("onUpdate callback not called")
	}
}
