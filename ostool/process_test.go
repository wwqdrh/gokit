package ostool

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestRunCmdStd(t *testing.T) {
	fmt.Println("Running 'ls -l'")
	err := RunCmdStd("ls -l")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Running 'echo hello'")
	err = RunCmdStd("echo hello")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Running 'invalid'")
	err = RunCmdStd("invalid")
	if err != nil {
		fmt.Println(err)
	}
}

func TestRunWithStdin(t *testing.T) {
	out, er, err := RunAndWaitWithIn(exec.Command("bash", "-c", "read name && echo $name"), "hhh")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(out, er)
}
