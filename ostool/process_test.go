package ostool

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestRunWithStdin(t *testing.T) {
	out, er, err := RunAndWaitWithIn(exec.Command("bash", "-c", "read name && echo $name"), "hhh")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(out, er)
}
