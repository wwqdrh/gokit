package tun

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/wwqdrh/gotoolkit/nettool/common"
	"github.com/wwqdrh/logger"
	"github.com/wwqdrh/ostool"
)

// CheckContext check everything needed for tun setup
// in Linux, should create the tun device by hand
// ip tuntap add mode tun dev tun0
// ip addr add 198.18.0.1/15 dev tun0
// ip link set dev tun0 up
func (s *Cli) CheckContext() error {
	if !ostool.CanRun(exec.Command("which", "ip")) {
		return fmt.Errorf("failed to found 'ip' command")
	}

	if _, errmsg, err := ostool.RunAndWait(exec.Command(
		"ip", "tuntap", "add", "mode", "tun", "dev", s.GetName()),
	); err != nil {
		return fmt.Errorf("%w: %s", err, errmsg)
	}

	// if _, errmsg, err := ostool.RunAndWait(exec.Command(
	// 	"ip", "addr", "add", "[ip]", "dev", s.GetName()),
	// ); err != nil {
	// 	return fmt.Errorf("%w: %s", err, errmsg)
	// }

	if _, errmsg, err := ostool.RunAndWait(exec.Command(
		"ip", "link", "set", "dev", s.GetName(), "up",
	)); err != nil {
		return fmt.Errorf("%w: %s", err, errmsg)
	}

	return nil
}

// SetRoute let specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string, excludeIpRange []string) error {
	// run command: ip link set dev kt0 up
	_, _, err := ostool.RunAndWait(exec.Command("ip",
		"link",
		"set",
		"dev",
		s.GetName(),
		"up",
	))
	if err != nil {
		logger.DefaultLogger.Error("Failed to set tun device up")
		return AllRouteFailError{err}
	}
	var lastErr error
	anyRouteOk := false
	for _, r := range ipRange {
		logger.DefaultLogger.Info("Adding route to %s" + r)
		// run command: ip route add 10.96.0.0/16 dev kt0
		_, _, err = ostool.RunAndWait(exec.Command("ip",
			"route",
			"add",
			r,
			"dev",
			s.GetName(),
		))
		if err != nil {
			logger.DefaultLogger.Warnx("Failed to set route %s to tun device", nil, r)
			lastErr = err
		} else {
			anyRouteOk = true
		}
	}
	if !anyRouteOk {
		return AllRouteFailError{lastErr}
	}
	return lastErr
}

// CheckRoute check whether all route rule setup properly
func (s *Cli) CheckRoute(ipRange []string) []string {
	var failedIpRange []string
	// run command: ip route show
	out, _, err := ostool.RunAndWait(exec.Command("ip",
		"route",
		"show",
	))
	if err != nil {
		logger.DefaultLogger.Warn("Failed to get route table")
		return []string{}
	}
	_, _ = io.Discard.Write([]byte(">> Get route: " + out + "\n"))

	nameWithPadding := fmt.Sprintf(" %s ", s.GetName())
	for _, ir := range ipRange {
		found := false
		for _, line := range strings.Split(out, "\n") {
			if strings.HasPrefix(line, ir) && strings.Contains(line, nameWithPadding) {
				found = true
				break
			}
		}
		if !found {
			failedIpRange = append(failedIpRange, ir)
		}
	}
	return failedIpRange
}

// RestoreRoute delete route rules made by kt
func (s *Cli) RestoreRoute() error {
	// Route will be auto removed when tun device destroyed
	return nil
}

func (s *Cli) GetName() string {
	return common.TunNameLinux
}
