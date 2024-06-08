package fileindex

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wwqdrh/gokit/logger"
)

type FileInfoTree struct {
	rootpath         string                // 扫描的根目录
	interval         int                   // 定时扫描路径
	ignores          map[string]struct{}   // 需要忽略的文件夹或者文件名
	onUpdate         OnFileInfoUpdate      // 文件更新回调函数
	tree             map[string]*FileIndex // 前缀树
	mu               sync.RWMutex          // 读写互斥锁
	stopCh           chan struct{}         // 停止通道
	wg               sync.WaitGroup        // 等待组
	running          bool
	defaultTimelines *sync.Map // map[string]int64
}

type FileIndex struct {
	BaseName   string
	Path       string
	UpdateTime int64 // 最后更新时间
	Size       int64 // 文件大小
}

type OnFileInfoUpdate func(FileIndex)

// 根据传递的p文件路径，递归遍历整个文件并构建一颗以文件路径为key的前缀树
func NewFileInfoTree(p string, interval int, ignores []string) *FileInfoTree {
	ignoreMap := map[string]struct{}{}
	for _, item := range ignores {
		ignoreMap[item] = struct{}{}
	}
	return &FileInfoTree{
		rootpath:         p,
		ignores:          ignoreMap,
		interval:         interval,
		tree:             make(map[string]*FileIndex),
		stopCh:           make(chan struct{}),
		defaultTimelines: &sync.Map{},
	}
}

func (i *FileInfoTree) SetOnFileInfoUpdate(fn OnFileInfoUpdate) {
	i.onUpdate = fn
}

func (i *FileInfoTree) SetDefaultTimeLines(timeslines map[string]int64) {
	for key, val := range timeslines {
		i.defaultTimelines.Store(key, val)
	}
}

// 定时扫描p路径下的文件，新增或者删除树节点，或者通过GetFileUpdateTime获取文件的上次更新时间并更新到树的节点中
// 扫描到文件的时候，执行i.onUpdate函数
func (i *FileInfoTree) Start() {
	i.wg.Add(1)
	go i.updateLoop()
}

func (i *FileInfoTree) updateLoop() {
	defer i.wg.Done()
	ticker := time.NewTicker(time.Duration(i.interval) * time.Millisecond)
	defer ticker.Stop()
	i.walk(i.rootpath)
	for {
		select {
		case <-ticker.C:
			if !i.running {
				i.running = true
				i.walk(i.rootpath)
				i.running = false
			}
		case <-i.stopCh:
			logger.DefaultLogger.Debug("quit fileinfo tree loop")
			return
		}
	}
}

// 如果为true，则需要更新
func (i *FileInfoTree) checkTimeline(path string, lasttime, currtime int64) bool {
	if lasttime == 0 {
		// 第一次调用，判断是否有设置的默认defaulttimeline
		if t, ok := i.defaultTimelines.Load(path); ok {
			lasttime = t.(int64)
		}
	} else {
		i.defaultTimelines.Store(path, currtime)
	}

	return lasttime < currtime
}

func (i *FileInfoTree) walk(path string) {
	if _, ok := i.ignores[filepath.Base(path)]; ok {
		return
	}
	fi, err := os.Stat(path)
	if err != nil {
		return
	}

	if !fi.IsDir() {
		var lastUpdate, lastSize int64
		if v, ok := i.tree[path]; ok {
			lastUpdate = v.UpdateTime
			lastSize = v.Size
		}

		idx := &FileIndex{
			Path:       path,
			BaseName:   fi.Name(),
			UpdateTime: fi.ModTime().UnixNano(),
			Size:       fi.Size(),
		}
		i.mu.Lock()
		i.tree[path] = idx
		i.mu.Unlock()
		// if update, maybe zero or actual data size
		if i.onUpdate != nil {
			// 空文件不做处理
			if idx.Size == 0 {
				return
			}
			// 如果更新时间发生了变化
			if i.checkTimeline(path, lastUpdate, idx.UpdateTime) {
				i.onUpdate(*idx)
			} else if lastSize != idx.Size && lastSize != 0 {
				i.onUpdate(*idx)
			}
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
	return *i.tree[p]
}

// 停止定时任务
func (i *FileInfoTree) Stop() {
	close(i.stopCh)
	i.wg.Wait()
}
