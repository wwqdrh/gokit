package ostool

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	fs "github.com/fsnotify/fsnotify"
	"github.com/wwqdrh/logger"
)

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
