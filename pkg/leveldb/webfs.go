package render

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/wwqdrh/gokit/logger"
)

// 实现了 fs.FS 接口,用于通过 HTTP 协议访问网络文件内容
// 1、添加目录结构的支持
// 2、添加缓存的支持
type WebFs struct {
	rootURL          string
	rootFileEntry    map[string][]string // 以#开头就表示为文件夹路径
	rootFileNotExist map[string]bool     // 判断文件是否存在
	dataCacheMaxSize int
	dataCache        map[string]*list.Element // []byte
	dataCacheIndex   *list.List               // lru, 把最左边的删除
}

type cacheEntry struct {
	key   string
	value []byte
}

func NewWebFs(prefix string) fs.FS {
	// get root path public.json
	f := &WebFs{
		rootURL:          prefix,
		rootFileEntry:    map[string][]string{},
		rootFileNotExist: map[string]bool{},
		dataCache:        make(map[string]*list.Element),
		dataCacheMaxSize: 50,
		dataCacheIndex:   list.New(),
	}
	if resp, err := http.Get(f.joinUri(f.rootURL, "public.json")); err != nil {
		logger.DefaultLogger.Warn(err.Error())
	} else {
		defer resp.Body.Close()
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.DefaultLogger.Warn(err.Error())
		} else {
			if err := json.Unmarshal(content, &f.rootFileEntry); err != nil {
				logger.DefaultLogger.Warn(err.Error())
			}
		}
	}
	return f
}

// Open 实现了 fs.FS 接口中的 Open 方法
// .: 根目录，需要列出这个目录下的文件以及文件夹
// [文件夹名称]/[文件夹名称]: 通过该名字寻找子目录下的文件内容
func (f *WebFs) Open(name string) (fs.File, error) {
	if _, ok := f.rootFileEntry[name]; ok {
		// 文件夹查询
		return &httpFile{
			name:        name,
			isDir:       true,
			getFileList: f.getFileList,
			getDirList:  f.getDirList,
		}, nil
	}
	buf, err := f.getDataOrCache(name)
	if err != nil {
		return nil, err
	}

	return &httpFile{
		Reader:      bytes.NewReader(buf),
		size:        int64(len(buf)),
		name:        name,
		isDir:       false,
		getFileList: f.getFileList,
		getDirList:  f.getDirList}, nil
}

func (f *WebFs) getDataOrCache(name string) ([]byte, error) {
	if f.rootFileNotExist[name] {
		return nil, os.ErrNotExist
	}

	if elem, ok := f.dataCache[name]; !ok {
		resp, err := http.Get(f.joinUri(f.rootURL, name))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			f.rootFileNotExist[name] = true
			return nil, os.ErrNotExist
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP status: %d", resp.StatusCode)
		}

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// 如果容量已满，则删除醉酒未使用的元素
		if len(f.dataCache) == f.dataCacheMaxSize {
			elem := f.dataCacheIndex.Back()
			delete(f.dataCache, elem.Value.(*cacheEntry).key)
			f.dataCacheIndex.Remove(elem)
		}

		newElem := f.dataCacheIndex.PushFront(&cacheEntry{name, content})
		f.dataCache[name] = newElem
		return content, nil
	} else {
		f.dataCacheIndex.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, nil
	}
}

func (f *WebFs) getFileList(name string) []fs.DirEntry {
	res := []fs.DirEntry{}
	if entrys, exist := f.rootFileEntry[name]; exist {
		for _, item := range entrys {
			res = append(res, &httpFileEntry{
				name:  item,
				isDir: false,
			})
		}
	}
	return res
}

func (f *WebFs) getDirList(name string) []fs.DirEntry {
	res := []fs.DirEntry{}
	if entrys, exist := f.rootFileEntry["#"+name]; exist {
		for _, item := range entrys {
			res = append(res, &httpFileEntry{
				name:  item,
				isDir: true,
			})
		}
	}
	return res
}

func (f *WebFs) joinUri(a, b string) string {
	if a[len(a)-1] == '/' || b[0] == '/' {
		return a + b
	} else {
		return a + "/" + b
	}
}

// httpFile 实现了 http.File 接口,用于读取 HTTP 响应体中的内容
type httpFile struct {
	*bytes.Reader
	size        int64
	name        string
	isDir       bool
	getFileList func(string) []fs.DirEntry
	getDirList  func(string) []fs.DirEntry
}

func (f *httpFile) Stat() (os.FileInfo, error) {
	return &httpFileEntry{
		name:  f.name,
		isDir: f.isDir,
		size:  f.size,
	}, nil
}

func (f *httpFile) Close() error {
	if f.Reader != nil {
		f.Reader.Reset(nil)
	}
	return nil
}

func (f *httpFile) ReadDir(n int) ([]fs.DirEntry, error) {
	// 由于我们的场景只涉及单个文件,所以返回一个空的切片
	res := []fs.DirEntry{}
	if f.getDirList != nil {
		res = append(res, f.getDirList(f.name)...)
	}
	if f.getFileList != nil {
		res = append(res, f.getFileList(f.name)...)
	}
	return res, nil
}

type httpFileEntry struct {
	name  string
	isDir bool
	size  int64
}

func (d *httpFileEntry) Name() string { return d.name }

func (d *httpFileEntry) IsDir() bool { return d.isDir }

func (d *httpFileEntry) Type() fs.FileMode {
	if d.isDir {
		return fs.ModeDir
	} else {
		return fs.ModePerm
	}
}

func (d *httpFileEntry) Info() (fs.FileInfo, error) {
	return &httpFileEntry{size: 1}, nil
}

func (d *httpFileEntry) Size() int64        { return d.size }
func (d *httpFileEntry) Mode() os.FileMode  { return 0 }
func (d *httpFileEntry) ModTime() time.Time { return time.Time{} }
func (d *httpFileEntry) Sys() interface{}   { return nil }
