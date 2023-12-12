package iptables

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/wwqdrh/gokit/logger"
)

type iptablesRule struct{}

func NewIptablesRule() iptablesRule {
	return iptablesRule{}
}

func RunAndWait(cmd *exec.Cmd) (string, string, error) {
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	logger.DefaultLogger.Debugx("Task %s with args %+v", nil, cmd.Path, cmd.Args)
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// --dest: 127.0.0.1/32
// --dport: 53
// --to-ports: 10053
func (r iptablesRule) NatAddPortRedirect(dest, port, tport string) error {
	if _, _, err := RunAndWait(exec.Command("iptables",
		"--table",
		"nat",
		"--insert",
		"OUTPUT",
		"--proto",
		"udp",
		"--dest",
		dest,
		"--dport",
		port,
		"--jump",
		"REDIRECT",
		"--to-ports",
		tport,
	)); err != nil {
		logger.DefaultLogger.Errorx("%s: Failed to use local dns server", nil, err.Error())
		return err
	}
	return nil
}

// run command: iptables --table nat --delete OUTPUT --proto udp --dest 127.0.0.1/32 --dport 53 --jump REDIRECT --to-ports 10053
func (r iptablesRule) NatDelPortRedirect(dest, port, tport string) {
	// iptables -D INPUT 3
	for {
		_, _, err := RunAndWait(exec.Command("iptables",
			"--table",
			"nat",
			"--delete",
			"OUTPUT",
			"--proto",
			"udp",
			"--dest",
			dest,
			"--dport",
			port,
			"--jump",
			"REDIRECT",
			"--to-ports",
			tport,
		))
		if err != nil {
			return
		}
	}
}

func (r iptablesRule) ListRuleNumber(table string) error {
	// iptables -L -n --line-number
	if res, _, err := RunAndWait(exec.Command("iptables",
		"--table",
		table,
		"-L", "-n", "--line-number",
	)); err != nil {
		logger.DefaultLogger.Errorx("%s: Failed to use local dns server", nil, err.Error())
		return err
	} else {
		fmt.Println(res)
	}
	return nil
}
