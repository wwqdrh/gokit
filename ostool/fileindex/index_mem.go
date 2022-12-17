package fileindex

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/wwqdrh/gokit/logger"
)

func NewMemIndex(root string, id string, name string, filter Filter, scanInterval uint) (Index, error) {
	rootPath := filepath.Clean(root)
	fi, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("%v is not a directory", rootPath)
	}
	idx := &memIndex{id, name, rootPath, nil, make(chan struct{}), false}
	go func() {
		for {
			idx.update()
			<-time.After(time.Second * time.Duration(scanInterval))
		}
	}()
	return idx, nil
}

/* Index Entry */

func entryId(name string, path string, fi os.FileInfo) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%v\n", name)))
	h.Write([]byte(fmt.Sprintf("%v\n", path)))
	h.Write([]byte(fmt.Sprintf("%v\n", fi.Size())))
	h.Write([]byte(fmt.Sprintf("%v\n", fi.ModTime())))
	h.Write([]byte(fmt.Sprintf("%v\n", fi.IsDir())))
	return fmt.Sprintf("%x", h.Sum(nil))
}

type memEntry struct {
	index    *memIndex
	id       string
	path     string
	parentId string
	isDir    bool
}

func newMemEntry(index *memIndex, path string, fi os.FileInfo, parentId string) Entry {
	id := entryId(index.name, path, fi)
	return &memEntry{index, id, path, parentId, fi.IsDir()}
}

func (e *memEntry) Id() string {
	return e.id
}

func (e *memEntry) Name() string {
	return filepath.Base(e.path)
}

func (e *memEntry) IsDir() bool {
	return e.isDir
}

func (e *memEntry) ParentId() string {
	return e.parentId
}

func (e *memEntry) Path() string {
	return path.Join(e.index.root, e.path)
}

/* Index Entry */

type memIndex struct {
	id        string
	name      string
	root      string
	data      *memIndexData
	inited    chan struct{}
	firstinit bool
}

func (i *memIndex) Id() string {
	return i.id
}

func (i *memIndex) Name() string {
	return i.name
}

func (i *memIndex) Root() string {
	return i.id
}

func (i *memIndex) Get(id string) (Entry, error) {
	return i.data.GetEntry(id), nil
}

func (i *memIndex) WaitForReady() error {
	<-i.inited
	return nil
}

func (i *memIndex) List(parent string) ([]Entry, error) {
	return i.data.GetChildren(parent), nil
}

func (i *memIndex) updateDir(d *memIndexData, path string, parentId string) error {
	dir, err := readDirInfo(path)
	if err != nil {
		return err
	}

	dirEntry, err := i.entryFromInfo(dir.info, dir.path, parentId)
	if err != nil {
		return err
	}

	d.add(dirEntry.(*memEntry))

	return i.updateChildren(d, dir, dirEntry.Id())
}

func (i *memIndex) updateChildren(d *memIndexData, dir *dirInfo, parentId string) error {
	for _, fi := range dir.children {
		if fi.IsDir() {

			err := i.updateDir(d, filepath.Join(dir.path, fi.Name()), parentId)
			if err != nil {
				logger.DefaultLogger.Errorx("Error adding directory %v: %v", nil, filepath.Join(dir.path, fi.Name()), err)
				continue
			}

		} else {

			fileEntry, err := i.entryFromInfo(fi, filepath.Join(dir.path, fi.Name()), parentId)
			if err != nil {
				logger.DefaultLogger.Errorx("Error adding file %v: %v", nil, filepath.Join(dir.path, fi.Name()), err)
				continue
			}

			d.add(fileEntry.(*memEntry))
		}
	}
	return nil
}

func (i *memIndex) entryFromInfo(fi os.FileInfo, path string, parentId string) (Entry, error) {
	rp, err := filepath.Rel(i.root, path)
	if err != nil {
		return nil, fmt.Errorf("could not determine relative path of %v in %v", path, i)
	}
	e := newMemEntry(i, rp, fi, parentId)
	return e, nil
}

func (i *memIndex) update() {
	logger.DefaultLogger.Infox("Starting index scan for %v", nil, i)
	d := newMemIndexData()
	dir, err := readDirInfo(i.root)
	if err != nil {
		logger.DefaultLogger.Errorx("Error during index scan for %v: %v", nil, i, err)
		return
	}
	err = i.updateChildren(d, dir, "")
	if err != nil {
		logger.DefaultLogger.Errorx("Error during index scan for %v: %v", nil, i, err)
		return
	}
	i.data = d
	if !i.firstinit {
		i.firstinit = true
		close(i.inited)
	}
	logger.DefaultLogger.Infox("Finished index scan for %v. Found %v entries", nil, i, i.data.EntryLen())
}

func (i *memIndex) String() string {
	return fmt.Sprintf("filepath.MemIndex(%v)", i.root)
}

/* Index Data */

type memIndexData struct {
	mx       *sync.RWMutex
	entries  sync.Map // map[string]Entry
	children map[string][]Entry
}

func newMemIndexData() *memIndexData {
	// return &memIndexData{make(map[string]Entry), make(map[string][]Entry)}
	return &memIndexData{&sync.RWMutex{}, sync.Map{}, make(map[string][]Entry)}
}

func (d *memIndexData) add(e *memEntry) {
	logger.DefaultLogger.Debugx("Adding index entry %v", nil, e.Path())
	d.entries.Store(e.Id(), e)

	d.mx.Lock()
	defer d.mx.Unlock()
	d.children[e.ParentId()] = append(d.children[e.ParentId()], e)
}

func (d *memIndexData) EntryLen() int {
	length := 0
	d.entries.Range(func(key, value any) bool {
		length++
		return true
	})
	return length
}

func (d *memIndexData) GetEntry(id string) Entry {
	val, ok := d.entries.Load(id)
	if !ok {
		return nil
	}
	res, ok := val.(Entry)
	if !ok {
		return nil
	}
	return res
}

func (d *memIndexData) GetChildren(id string) []Entry {
	d.mx.RLock()
	defer d.mx.RUnlock()
	val, ok := d.children[id]
	if !ok {
		return nil
	}
	return val
}
