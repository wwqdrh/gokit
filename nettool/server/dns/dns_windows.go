package dns

import (
	"fmt"
	"github.com/wwqdrh/gotoolkit/nettool/common"
	"github.com/wwqdrh/logger"
	"os/exec"
	"regexp"
	"strings"
)

// SetNameServer set dns server records
func SetNameServer(dnsServer string) (err error) {
	// run command: netsh interface ip set interface KtConnectTunnel metric=2
	if _, _, err = RunAndWait(exec.Command("netsh",
		"interface",
		"ipv4",
		"set",
		"interface",
		common.TunNameWin,
		"metric=2",
	)); err != nil {
		logger.DefaultLogger.Error("Failed to set tun device order")
		return err
	}
	// run command: netsh interface ip set dnsservers name=KtConnectTunnel source=static address=8.8.8.8
	if _, _, err = RunAndWait(exec.Command("netsh",
		"interface",
		"ipv4",
		"set",
		"dnsservers",
		fmt.Sprintf("name=%s", common.TunNameWin),
		"source=static",
		fmt.Sprintf("address=%s", strings.Split(dnsServer, ":")[0]),
	)); err != nil {
		logger.DefaultLogger.Error("Failed to set dns server of tun device")
		return err
	}
	return nil
}

// HandleExtraDomainMapping handle extra domain change
func HandleExtraDomainMapping(extraDomains map[string]string, localDnsPort int) {
	// pass
}

// RestoreNameServer ...
func RestoreNameServer() {
	// Windows dns config is set on device, so explicit removal is unnecessary
}

// GetLocalDomains ...
func GetLocalDomains() string {
	return ""
}

// GetNameServer get dns server of the default interface
func GetNameServer() string {
	// run command: netsh interface ip show dnsservers
	out, _, err := RunAndWait(exec.Command("netsh",
		"interface",
		"ipv4",
		"show",
		"dnsservers",
	))
	if err != nil {
		logger.DefaultLogger.Error("Failed to get upstream dns server")
		return ""
	}
	_, _ = common.BackgroundLogger.Write([]byte(">> Get dns: " + out + common.Eol))

	r, _ := regexp.Compile(common.IpAddrPattern)
	nsAddresses := r.FindAllString(out, 10)
	if nsAddresses == nil {
		logger.DefaultLogger.Warn("No upstream dns server available")
		return ""
	}
	for _, addr := range nsAddresses {
		if addr != common.Localhost {
			return addr
		}
	}
	logger.DefaultLogger.Warn("No valid upstream dns server available")
	return ""
}
