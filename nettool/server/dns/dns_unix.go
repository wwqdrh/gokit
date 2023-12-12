//go:build !windows

package dns

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/wwqdrh/gokit/logger"
)

// listen address of systemd-resolved
const resolvedAddr = "127.0.0.53"
const resolvedConf = "/run/systemd/resolve/resolv.conf"

// GetLocalDomains get domain search postfixes
func GetLocalDomains() string {
	f, err := os.Open(ResolvConf)
	if err != nil {
		return ""
	}
	defer f.Close()

	var localDomains []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, FieldDomain) {
			localDomains = append(localDomains, strings.TrimSpace(strings.TrimPrefix(line, FieldDomain)))
		} else if strings.HasPrefix(line, FieldSearch) {
			for _, s := range strings.Split(strings.TrimPrefix(line, FieldSearch), " ") {
				if s != "" {
					localDomains = append(localDomains, s)
				}
			}
		}
	}
	return strings.Join(localDomains, ",")
}

// GetNameServer get primary dns server
func GetNameServer() string {
	ns := fetchNameServerInConf(ResolvConf)
	if runtime.GOOS == "linux" && ns == resolvedAddr {
		logger.DefaultLogger.Debug("Using systemd-resolved")
		return fetchNameServerInConf(resolvedConf)
	}
	return ns
}

func fetchNameServerInConf(resolvConf string) string {
	f, err := os.Open(resolvConf)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	pattern, _ := regexp.Compile(fmt.Sprintf("^%s[ \t]+"+IpAddrPattern, FieldNameserver))
	for scanner.Scan() {
		line := scanner.Text()
		if pattern.MatchString(line) {
			return strings.TrimSpace(strings.TrimPrefix(line, FieldNameserver))
		}
	}
	return ""
}
