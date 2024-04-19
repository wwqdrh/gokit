package fileindex

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestFileInfoTree(t *testing.T) {
	tests := []struct {
		tmpDir      string
		interval    int
		ignores     []string
		expectCount int
	}{
		{tmpDir: "test", interval: 1000, ignores: nil, expectCount: 2},
		{tmpDir: "test2", interval: 500, ignores: []string{"test.txt"}, expectCount: 0},
	}
	for _, item := range tests {
		// 创建临时目录
		tempDir, err := os.MkdirTemp("", item.tmpDir)
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// 创建 FileInfoTree
		tree := NewFileInfoTree(tempDir, item.interval, item.ignores)

		// 设置回调函数
		var updateCalled int64 = 0
		tree.SetOnFileInfoUpdate(func(fi FileIndex) {
			fmt.Println(fi.UpdateTime, fi.Size, fi.BaseName, fi.Path)
			atomic.AddInt64(&updateCalled, 1)
		})

		// 启动更新
		tree.Start()

		// 创建测试文件
		testFilePath := filepath.Join(tempDir, "test.txt")
		err = os.WriteFile(testFilePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// 等待一段时间,
		// 与定时器的1000错开，不要使用倍数，否则刚好会遇到文件从有数据变为0，以及从0变为新的数据这两个阶段，会导致最终调用3次update
		time.Sleep(1500 * time.Millisecond)

		// 更新文件
		err = os.WriteFile(testFilePath, []byte("test content 2"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		// 这里必须等长一段时间
		time.Sleep(1500 * time.Millisecond)

		if atomic.LoadInt64(&updateCalled) != int64(item.expectCount) {
			t.Errorf("onUpdate callback not called")
		}
		tree.Stop()
	}
}
