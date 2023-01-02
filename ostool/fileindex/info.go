package fileindex

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/wwqdrh/gokit/logger"
)

// 获取文件的md5字符串
// 根据文件内容进行md5
func FileMd5ByContent(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("FileMd5: 文件打开失败 %w", err)
	}
	defer f.Close()
	md5h := md5.New()
	_, err = io.Copy(md5h, f)
	if err != nil {
		return "", fmt.Errorf("FileMd5: 生成md5失败 %w", err)
	}
	return hex.EncodeToString(md5h.Sum(nil)[:]), nil
}

// 获取文件的md5字符串
// 根据文件名字、文件大小 文件数据的前100byte+后100byte作为md5数据
func FileMd5BySpec(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("FileMd5: 文件打开失败 %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("FileMd5: 文件stat获取失败 %w", err)
	}

	md5h := md5.New()
	_, err = md5h.Write([]byte(stat.Name()))
	if err != nil {
		return "", fmt.Errorf("FileMd5: 生成md5失败 %w", err)
	}
	_, err = md5h.Write([]byte(fmt.Sprint(stat.Size())))
	if err != nil {
		return "", fmt.Errorf("FileMd5: 生成md5失败 %w", err)
	}
	frontBuf := make([]byte, 100)
	frontNum, err := f.Read(frontBuf)
	if err != nil {
		return "", fmt.Errorf("FileMd5: 生成md5失败 %w", err)
	}
	_, err = md5h.Write(frontBuf[:frontNum])
	if err != nil {
		return "", fmt.Errorf("FileMd5: 生成md5失败 %w", err)
	}
	backBuf := make([]byte, 100)
	_, err = f.Seek(100, os.SEEK_END)
	if err != nil {
		return "", fmt.Errorf("FileMd5: 生成md5失败 %w", err)
	}
	backNum, err := f.Read(backBuf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("FileMd5: 生成md5失败 %w", err)
	}
	_, err = md5h.Write(backBuf[:backNum])
	if err != nil {
		return "", fmt.Errorf("FileMd5: 生成md5失败 %w", err)
	}

	return hex.EncodeToString(md5h.Sum(nil)), nil
}

func IsSubDir(basePath, targetPath string) bool {
	f, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return false
	}
	if strings.Index(f, "../") == 0 {
		return false
	}
	return true
}

func GetAllDir(pathname string, s []string) ([]string, error) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return s, err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			s = append(s, fullDir)

			s, err = GetAllDir(fullDir, s)
			if err != nil {
				fmt.Println("read dir fail:", err)
				return s, err
			}
		}
	}
	return s, nil
}

func GetAllFile(source string, prefix bool) ([]string, error) {
	source = strings.TrimLeft(source, "./")
	dirStack := []string{source}

	res := []string{}
	for len(dirStack) > 0 {
		cur := dirStack[0]
		dirStack = dirStack[1:]

		files, err := os.ReadDir(cur)
		if err != nil {
			logger.DefaultLogger.Warn(cur + " 不是文件夹")
			continue
		}

		for _, item := range files {
			if item.IsDir() {
				dirStack = append(dirStack, path.Join(cur, item.Name()))
			} else {
				res = append(res, path.Join(cur, item.Name()))
			}
		}
	}

	if !prefix {
		for i := 0; i < len(res); i++ {
			cur := strings.TrimPrefix(res[i], source)
			cur = strings.TrimPrefix(cur, "/")
			res[i] = cur
		}
	}

	return res, nil
}
