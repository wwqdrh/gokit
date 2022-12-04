package ostool

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"os"
	"path"
	"strconv"

	fs "github.com/fsnotify/fsnotify"
	"github.com/wwqdrh/logger"
)

type FileSize int64

func (f FileSize) B() string {
	return fmt.Sprintf("%dB", f)
}

func (f FileSize) KB() string {
	return fmt.Sprintf("%dKB", f/1024)
}

func (f FileSize) MB() string {
	return fmt.Sprintf("%dKB", f/1024/1024)
}

func (f FileSize) GB() string {
	return fmt.Sprintf("%dKB", f/1024/1024/1024)
}
func (f FileSize) TB() string {
	return fmt.Sprintf("%dKB", f/1024/1024/1024/1024)
}

type FileInfo struct {
	Dir  string
	Name string
	Size FileSize
}

func FileStore(dir string, file *multipart.FileHeader) (FileInfo, error) {
	var err error
	if err = CreateDirIfNotExist(dir); err != nil {
		return FileInfo{}, err
	}

	var name string
	if name, err = RandomName(dir, FileExt(file), 12); err != nil {
		return FileInfo{}, err
	}
	if err = SaveUploadedFile(file, path.Join(dir, name)); err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Dir:  dir,
		Name: name,
		Size: FileSize(file.Size),
	}, nil
}

func GetFileInfo(file string) (FileInfo, error) {
	if info, err := os.Stat(file); os.IsNotExist(err) {
		return FileInfo{}, err
	} else {
		dir, name := path.Split(file)
		return FileInfo{
			Dir:  dir,
			Name: name,
			Size: FileSize(info.Size()),
		}, nil
	}
}

func FileExt(file *multipart.FileHeader) string {
	return path.Ext(file.Filename)
}

func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func RandomName(dir, ext string, length int) (string, error) {
	for i := 0; i < 3; i++ {
		res := make([]byte, length)
		for i := 0; i < length; i++ {
			res[i] = Letters[rand.Intn(len(Letters))]
		}
		name := fmt.Sprintf("%s.%s", string(res), ext)
		if _, err := os.Stat(path.Join(dir, name)); os.IsNotExist(err) {
			return name, nil
		}
	}
	return "", errors.New("more 3 times")
}

// CreateDirIfNotExist create dir
func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// WritePidFile write pid to file
func WritePidFile(workDir, name string, ch chan os.Signal) error {
	pidFile := fmt.Sprintf("%s/%s-%d.pid", workDir, name, os.Getpid())
	if err := ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		return err
	}
	go watchPidFile(pidFile, ch)
	return nil
}

func watchPidFile(pidFile string, ch chan os.Signal) {
	watcher, err := fs.NewWatcher()
	if err != nil {
		logger.DefaultLogger.Warnx("%s: Failed to create pid file watcher", nil, err.Error())
	}
	defer watcher.Close()

	if err = watcher.Add(pidFile); err != nil {
		logger.DefaultLogger.Errorx("%s: Unable to watch pid file", nil, err.Error())

	}

	for event := range watcher.Events {
		logger.DefaultLogger.Debugx("Received event %s", nil, event)
		if event.Op&fs.Remove == fs.Remove || event.Op&fs.Rename == fs.Rename {
			logger.DefaultLogger.Info("Pid file was removed")
			ch <- os.Interrupt
		}
	}
}

// FixFileOwner set owner to original user when run with sudo
func FixFileOwner(path string) (err error) {
	var uid int
	var gid int
	sudoUid := os.Getenv("SUDO_UID")
	if sudoUid == "" {
		uid = os.Getuid()
	} else {
		if uid, err = strconv.Atoi(sudoUid); err != nil {
			return err
		}
	}
	sudoGid := os.Getenv("SUDO_GID")
	if sudoGid == "" {
		gid = os.Getuid()
	} else {
		if gid, err = strconv.Atoi(sudoGid); err != nil {
			return err
		}
	}
	return os.Chown(path, uid, gid)
}
