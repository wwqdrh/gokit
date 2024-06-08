//go:build linux
// +build linux

package fileindex

import (
	"os"
	"syscall"
	"time"
)

func GetFileCreateTime(path string) time.Time {
	fileInfo, _ := os.Stat(path)

	stat_t := fileInfo.Sys().(*syscall.Stat_t)
	tCreate := time.Unix(int64(stat_t.Ctim.Sec), int64(stat_t.Ctim.Nsec))
	return tCreate
}

func GetFileModTime(path string) time.Time {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return fileInfo.ModTime()
}
