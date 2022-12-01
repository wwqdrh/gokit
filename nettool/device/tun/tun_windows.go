package tun

import (
	"fmt"
	"github.com/wwqdrh/logger"
	"os/exec"
	"strings"
)

// CheckContext check everything needed for tun setup
func (s *Cli) CheckContext() error {
	if !CanRun(exec.Command("which", "ip")) {
		return fmt.Errorf("failed to found 'ip' command")
	}
	return nil
}

// SetRoute let specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string, excludeIpRange []string) error {
	// run command: ip link set dev kt0 up
	_, _, err := RunAndWait(exec.Command("ip",
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
		logger.DefaultLogger.Info("Adding route to " + r)
		// run command: ip route add 10.96.0.0/16 dev kt0
		_, _, err = RunAndWait(exec.Command("ip",
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
	out, _, err := RunAndWait(exec.Command("ip",
		"route",
		"show",
	))
	if err != nil {
		logger.DefaultLogger.Warn("Failed to get route table")
		return []string{}
	}
	_, _ = BackgroundLogger.Write([]byte(">> Get route: " + out + Eol))

	nameWithPadding := fmt.Sprintf(" %s ", s.GetName())
	for _, ir := range ipRange {
		found := false
		for _, line := range strings.Split(out, Eol) {
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
	return TunNameLinux
}
