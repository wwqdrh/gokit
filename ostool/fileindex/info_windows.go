//go:build windows
// +build windows

package fileindex

import (
	"os"
	"runtime"
	"syscall"
	"time"
)

func GetFileCreateTime(path string) time.Time {
	osType := runtime.GOOS
	fileInfo, _ := os.Stat(path)
	if osType == "windows" {
		wFileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
		tNanSeconds := wFileSys.CreationTime.Nanoseconds() /// 返回的是纳秒
		tSec := tNanSeconds / 1e9                          ///秒
		return time.Unix(tSec, tNanSeconds%1e9)
	}

	return time.Now()
}

func GetFileModTime(path string) time.Time {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return fileInfo.ModTime()
}
