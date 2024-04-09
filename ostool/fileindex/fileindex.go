package fileindex

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileInfoTree struct {
	rootpath string               // 扫描的根目录
	interval int                  // 定时扫描路径
	onUpdate OnFileInfoUpdate     // 文件更新回调函数
	tree     map[string]FileIndex // 前缀树
	mu       sync.RWMutex         // 读写互斥锁
	stopCh   chan struct{}        // 停止通道
	wg       sync.WaitGroup       // 等待组
}

type FileIndex struct {
	UpdateTime int64 // 最后更新时间
	Size       int64 // 文件大小
}

type OnFileInfoUpdate func(FileIndex)

// 根据传递的p文件路径，递归遍历整个文件并构建一颗以文件路径为key的前缀树
func NewFileInfoTree(p string, interval int) *FileInfoTree {
	return &FileInfoTree{
		rootpath: p,
		interval: interval,
		tree:     make(map[string]FileIndex),
		stopCh:   make(chan struct{}),
	}
}

func (i *FileInfoTree) SetOnFileInfoUpdate(fn OnFileInfoUpdate) {
	i.onUpdate = fn
}

// 定时扫描p路径下的文件，新增或者删除树节点，或者通过GetFileUpdateTime获取文件的上次更新时间并更新到树的节点中
// 扫描到文件的时候，执行i.onUpdate函数
func (i *FileInfoTree) Start() {
	i.wg.Add(1)
	go i.updateLoop()
}

func (i *FileInfoTree) updateLoop() {
	defer i.wg.Done()
	ticker := time.NewTicker(time.Duration(i.interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i.walk(i.rootpath)
		case <-i.stopCh:
			return
		}
	}
}

func (i *FileInfoTree) walk(path string) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}

	if !fi.IsDir() {
		lastUpdate := i.tree[path].UpdateTime

		idx := FileIndex{
			UpdateTime: fi.ModTime().UnixNano(),
			Size:       fi.Size(),
		}
		i.mu.Lock()
		i.tree[path] = idx
		i.mu.Unlock()
		if i.onUpdate != nil && lastUpdate != idx.UpdateTime {
			i.onUpdate(idx)
		}
		return
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return
	}

	for _, file := range files {
		i.walk(filepath.Join(path, file.Name()))
	}
}

// 给定一个文件路径，获取该文件的节点信息
func (i *FileInfoTree) Get(p string) FileIndex {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.tree[p]
}

// 停止定时任务
func (i *FileInfoTree) Stop() {
	close(i.stopCh)
	i.wg.Wait()
}
