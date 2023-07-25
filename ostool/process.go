package ostool

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	ps "github.com/mitchellh/go-ps"
	"github.com/pkg/errors"
	"github.com/wwqdrh/gokit/logger"
)

// RunCmdStd run a command, and redirect to stdout stderr
func RunCmdStd(cmdStr string) error {
	cmd := exec.Command("sh", "-c", cmdStr)

	// Get the standard output and error pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		return err
	}

	// Create buffered readers for the pipes
	stdoutReader := bufio.NewReader(stdout)
	stderrReader := bufio.NewReader(stderr)

	// Read from the pipes until they are closed or an error occurs
	for {
		// Read a line from the standard output
		line, err := stdoutReader.ReadString('\n')
		if err != nil {
			break // Exit the loop if there is an error or EOF
		}
		// Print the line to the console
		fmt.Print(line)
	}

	iserr := false
	for {
		// Read a line from the standard error
		line, err := stderrReader.ReadString('\n')
		if err != nil {
			break // Exit the loop if there is an error or EOF
		}
		if line != "" {
			iserr = true
			fmt.Print(line)
		}
	}

	if err := cmd.Wait(); err != nil {
		return err
	} else if iserr {
		return errors.New("cmd error")
	} else {
		return nil
	}
}

// WaitInput 等待用户输入
func WaitInput(inputs []string, usage string) int {
	// Create a buffered reader for the standard input
	reader := bufio.NewReader(os.Stdin)

	// Prompt the user to enter y or n
	fmt.Print(fmt.Sprintf("%s %s\n", usage, strings.Join(inputs, " or ")))

	// Read a line from the standard input
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return -1
	}
	line = strings.TrimSpace(line)

	// Check the first character of the line
	for i, ch := range inputs {
		if ch == line {
			return i
		}
	}
	fmt.Println("Invalid input, please enter " + strings.Join(inputs, " or "))
	return WaitInput(inputs, usage)
}

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

func PipeCommands(commands ...*exec.Cmd) ([]byte, error) {
	for i, command := range commands[:len(commands)-1] {
		out, err := command.StdoutPipe()
		if err != nil {
			return nil, err
		}
		command.Start()
		commands[i+1].Stdin = out
	}
	final, err := commands[len(commands)-1].Output()
	if err != nil {
		return nil, err
	}
	return final, nil
}

// TODO: 为stdin添加eof，否则部分程序是会一直等待stdin
func RunAndWaitWithIn(cmd *exec.Cmd, in string) (string, string, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", "", err
	}
	defer stdin.Close() // the doc says subProcess.Wait will close it, but I'm not sure, so I kept this line

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	logger.DefaultLogger.Debugx("Task %s with args %+v", nil, cmd.Path, cmd.Args)
	if err = cmd.Start(); err != nil { //Use start, not run
		return "", "", fmt.Errorf("an error occured: %w", err)
	}
	io.WriteString(stdin, in)
	io.WriteString(stdin, "\n")
	if err := stdin.Close(); err != nil {
		return "", "", err
	}
	cmd.Wait()
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
