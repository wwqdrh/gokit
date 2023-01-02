package ostool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"time"

	fs "github.com/fsnotify/fsnotify"
	"github.com/wwqdrh/gokit/logger"
	"github.com/wwqdrh/gokit/ostool/fileindex"
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

// interval 完全重新绑定监听的时间间隔，建议设置大一点比较好
func RegisterNotify(ctx context.Context, p string, interval time.Duration, cb func(fs.Event)) error {
	// Create new watcher.
	watcher, err := fs.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// Start listening for events.
	go func() {
		defer watcher.Close()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op == fs.Create {
					if fileindex.IsDir(event.Name) {
						if err := watcher.Add(event.Name); err != nil {
							logger.DefaultLogger.Warn(err.Error())
						}
						continue
					}
				} else if event.Op == fs.Remove {
					if fileindex.IsDir(event.Name) {
						if err := watcher.Remove(event.Name); err != nil {
							logger.DefaultLogger.Warn(err.Error())
						}
						continue
					}
				}
				cb(event)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-ctx.Done():
				logger.DefaultLogger.Info("停止监听")
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		updatesub := func() {
			for _, item := range watcher.WatchList() {
				watcher.Remove(item)
			}

			if err := watcher.Add(p); err != nil {
				logger.DefaultLogger.Warn(err.Error())
				return
			}
			res, err := fileindex.GetAllDir(p, []string{})
			if err != nil {
				logger.DefaultLogger.Warn(err.Error())
				return
			}
			for _, item := range res {
				if err := watcher.Add(item); err != nil {
					logger.DefaultLogger.Warn(err.Error())
				}
			}
		}

		updatesub()
		for {
			select {
			case <-ticker.C:
				updatesub()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
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
