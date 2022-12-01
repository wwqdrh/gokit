package hosts

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/gofrs/flock"
	"github.com/wwqdrh/logger"
)

var (
	HostsFilePath = "/etc/hosts"
	UserHome      = os.Getenv("HOME")
	NettoolHome   = fmt.Sprintf("%s/.nettool", UserHome)
	HostLockDir   = fmt.Sprintf("%s/lock", NettoolHome)
	Eol           = "\n"
)

type hosts struct {
	escapeBegin string
	escapeEnd   string
}

type hostsOpt func(o *hosts)

func WithEscape(begin, end string) hostsOpt {
	return func(o *hosts) {
		o.escapeBegin = begin
		o.escapeEnd = end
	}
}

func NewHosts(opts ...hostsOpt) hosts {
	h := hosts{
		escapeBegin: "# Nettool Hosts Begin",
		escapeEnd:   "# Nettool Hosts End",
	}
	for _, opt := range opts {
		opt(&h)
	}
	return h
}

// TODO: this is a temporary solution to avoid dumping after cleanup triggered
var doNotDump = false

// DropHosts remove hosts domain record added by kt
func (h hosts) DropHosts() {
	doNotDump = true
	lines, err := h.loadHostsFile()
	if err != nil {
		logger.DefaultLogger.Errorx("%e: Failed to load hosts file", nil, err)
		return
	}
	linesAfterDrop, _, err := h.dropHosts(lines, "")
	if err != nil {
		logger.DefaultLogger.Errorx("%e: Failed to parse hosts file", nil, err)

		return
	}
	if len(linesAfterDrop) < len(lines) {
		err = h.updateHostsFile(linesAfterDrop)
		if err != nil {
			logger.DefaultLogger.Errorx("%e: Failed to drop hosts file", nil, err)
			return
		}
		logger.DefaultLogger.Info("Drop hosts successful")
	}
}

// DumpHosts dump service domain to hosts file
func (h hosts) DumpHosts(hostsMap map[string]string, namespaceToDrop string) error {
	if doNotDump {
		return nil
	}
	lines, err := h.loadHostsFile()
	if err != nil {
		logger.DefaultLogger.Errorx("%e: Failed to load hosts file", nil, err)

		return err
	}
	linesBeforeDump, linesToKeep, err := h.dropHosts(lines, namespaceToDrop)
	if err != nil {
		logger.DefaultLogger.Errorx("%e: Failed to parse hosts file", nil, err)

		return err
	}
	if err = h.updateHostsFile(h.mergeLines(linesBeforeDump, h.dumpHosts(hostsMap, linesToKeep))); err != nil {
		logger.DefaultLogger.Warn("Failed to dump hosts file")
		logger.DefaultLogger.Debug(err.Error())
		return err
	}
	logger.DefaultLogger.Debug("Dump hosts successful")

	return nil
}

func (h hosts) dropHosts(rawLines []string, namespaceToDrop string) ([]string, []string, error) {
	escapeBegin := -1
	escapeEnd := -1
	// midDomain := fmt.Sprintf(".%s", namespaceToDrop)
	// keepShortDomain := namespaceToDrop != opt.Get().Global.Namespace
	keepShortDomain := true
	recordsToKeep := make([]string, 0)
	for i, l := range rawLines {
		if l == h.escapeBegin {
			escapeBegin = i
		} else if l == h.escapeEnd {
			escapeEnd = i
		} else if escapeBegin >= 0 && escapeEnd < 0 && namespaceToDrop != "" {
			if ok, err := regexp.MatchString(".+ [^.]+$", l); ok && err == nil {
				if keepShortDomain {
					recordsToKeep = append(recordsToKeep, l)
				}
			} // else if !strings.HasSuffix(l, midDomain) && !strings.HasSuffix(l, fullDomain) {
			// 	recordsToKeep = append(recordsToKeep, l)
			// }
		}
	}
	if escapeEnd < escapeBegin {
		return nil, nil, fmt.Errorf("invalid hosts file: recordBegin=%d, recordEnd=%d", escapeBegin, escapeEnd)
	}

	if escapeBegin >= 0 && escapeEnd > 0 {
		linesAfterDrop := make([]string, len(rawLines)-(escapeEnd-escapeBegin+1))
		if escapeBegin > 0 {
			copy(linesAfterDrop[0:escapeBegin], rawLines[0:escapeBegin])
		}
		if escapeEnd < len(rawLines)-1 {
			copy(linesAfterDrop[escapeBegin:], rawLines[escapeEnd+1:])
		}
		return linesAfterDrop, recordsToKeep, nil
	} else if escapeBegin >= 0 || escapeEnd > 0 {
		return nil, nil, fmt.Errorf("invalid hosts file: recordBegin=%d, recordEnd=%d", escapeBegin, escapeEnd)
	} else {
		return rawLines, []string{}, nil
	}
}

func (h hosts) dumpHosts(hostsMap map[string]string, linesToKeep []string) []string {
	var lines []string
	lines = append(lines, h.escapeBegin)
	for host, ip := range hostsMap {
		if ip != "" {
			lines = append(lines, fmt.Sprintf("%s %s", ip, host))
		}
	}
	for _, l := range linesToKeep {
		lines = append(lines, l)
	}
	lines = append(lines, h.escapeEnd)
	return lines
}

func (h hosts) mergeLines(linesBefore []string, linesAfter []string) []string {
	lines := make([]string, len(linesBefore)+len(linesAfter)+2)
	posBegin := len(linesBefore)
	if posBegin > 0 {
		copy(lines[0:posBegin], linesBefore[:])
	}
	if len(linesAfter) > 0 {
		copy(lines[posBegin+1:len(lines)-1], linesAfter[:])
	}
	return lines
}

func (h hosts) loadHostsFile() ([]string, error) {
	var lines []string
	file, err := os.Open(h.getHostsPath())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func (h hosts) updateHostsFile(lines []string) error {
	lock := flock.New(fmt.Sprintf("%s/hosts.lock", HostLockDir))
	timeoutContext, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	defer cancel()
	if ok, err := lock.TryLockContext(timeoutContext, 100*time.Millisecond); !ok {
		return fmt.Errorf("failed to require hosts lock")
	} else if err != nil {
		logger.DefaultLogger.Errorx("%e: require hosts file failed with error", nil, err)
		return err
	}
	defer lock.Unlock()

	file, err := os.Create(h.getHostsPath())
	if err != nil {
		return err
	}

	// fix sometimes update windows system hosts file but not effective immediately on existed in linesBeforeDump
	// because hosts not closed immediately
	// for example 10.110.55.200 vip-apiserver.test.com, this domain not existed in dns
	defer file.Close()

	w := bufio.NewWriter(file)
	continualEmptyLine := false
	for _, l := range lines {
		if continualEmptyLine && l == "" {
			continue
		}
		continualEmptyLine = l == ""
		fmt.Fprintf(w, "%s%s", l, Eol)
	}

	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (h hosts) getHostsPath() string {
	if os.Getenv("HOSTS_PATH") == "" {
		return os.ExpandEnv(filepath.FromSlash(HostsFilePath))
	} else {
		return os.Getenv("HOSTS_PATH")
	}
}
