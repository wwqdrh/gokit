package fileindex

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wwqdrh/gokit/logger"
)

type Cache interface {
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, ey string) ([]byte, error)
}

type DirCache struct {
	path string
}

func NewDirCache(path string) *DirCache {
	return &DirCache{path}
}

func (d *DirCache) getCachePath(key string) string {
	return filepath.Join(d.path, key)
}

func (d *DirCache) Get(ctx context.Context, key string) ([]byte, error) {
	file, err := os.Open(d.getCachePath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()
	b := new(bytes.Buffer)
	if _, err = io.Copy(b, file); err != nil {
		logger.DefaultLogger.Errorx("Error copying file to cache value: %v", nil, err)
		return nil, err
	}
	return b.Bytes(), nil
}

func (d *DirCache) Set(ctx context.Context, key string, value []byte) error {
	logger.DefaultLogger.Debugx("Setting cache item %v", nil, key)
	if err := os.MkdirAll(d.path, 0777); err != nil {
		logger.DefaultLogger.Errorx("Could not create cache dir %v: %v", nil, d.path, err)
		return err
	}
	cacheTmpFile, err := ioutil.TempFile(d.path, key+".*")
	if err != nil {
		logger.DefaultLogger.Errorx("Could not create cache file %v: %v", nil, cacheTmpFile, err)
		return err
	}
	if _, err := io.Copy(cacheTmpFile, bytes.NewReader(value)); err != nil {
		logger.DefaultLogger.Errorx("Could not write cache file %v: %v", nil, cacheTmpFile, err)
		cacheTmpFile.Close()
		os.Remove(cacheTmpFile.Name())
		return err
	}
	if err = cacheTmpFile.Close(); err != nil {
		logger.DefaultLogger.Errorx("Could not close cache file %v: %v", nil, cacheTmpFile, err)
		os.Remove(cacheTmpFile.Name())
		return err
	}
	if err = os.Rename(cacheTmpFile.Name(), d.getCachePath(key)); err != nil {
		logger.DefaultLogger.Errorx("Could not move cache file %v: %v", nil, cacheTmpFile, err)
		os.Remove(cacheTmpFile.Name())
		return err
	}
	return nil
}
