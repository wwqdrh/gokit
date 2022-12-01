package ostool

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	ps "github.com/mitchellh/go-ps"
	"github.com/wwqdrh/logger"
)

// RunAndWait run cmd
func RunAndWait(cmd *exec.Cmd) (string, string, error) {
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	logger.DefaultLogger.Debugx("Task %s with args %+v", nil, cmd.Path, cmd.Args)
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// SetupProcess write pid file and set component type
func SetupProcess(workDir, name string) (chan os.Signal, error) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	return ch, WritePidFile(workDir, name, ch)
}

// GetDaemonRunning fetch daemon pid if exist
func GetDaemonRunning(workDir, name string) int {
	files, _ := ioutil.ReadDir(workDir)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), name) && strings.HasSuffix(f.Name(), ".pid") {
			from := len(name) + 1
			to := len(f.Name()) - len(".pid")
			pid, err := strconv.Atoi(f.Name()[from:to])
			if err == nil && IsProcessExist(pid, name) {
				return pid
			}
		}
	}
	return -1
}

// IsProcessExist check whether specified process still running
func IsProcessExist(pid int, name string) bool {
	proc, err := ps.FindProcess(pid)
	if proc == nil || err != nil {
		return false
	}
	return strings.Contains(proc.Executable(), name)
}
