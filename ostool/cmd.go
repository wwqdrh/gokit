package ostool

import (
	"os"
	"runtime"
	"strings"

	ps "github.com/mitchellh/go-ps"
)

// IsCmd check running in windows cmd shell
func IsCmd() bool {
	proc, _ := ps.FindProcess(os.Getppid())
	if proc != nil && !strings.Contains(proc.Executable(), "cmd.exe") {
		return false
	}
	return true
}

// IsWindows check runtime is windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsMacos check runtime is macos
func IsMacos() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux check runtime is linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}
