package dns

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/wwqdrh/gokit/nettool/common"
	"github.com/wwqdrh/gokit/nettool/rules/iptables"
)

var (
	DnsResolve = NewDnsResolve(WithEscape(
		" # Added by NetTool", " # Removed by NetTool",
	))

	Iptables = iptables.NewIptablesRule()
)

// SetNameServer set dns server records
func SetNameServer(dnsServer string, dnsMode string) error {
	dnsSignal := make(chan error)
	go func() {
		defer func() {
			DnsResolve.RestoreResolvConf()
			if strings.HasPrefix(dnsMode, common.DnsModeLocalDns) {
				Iptables.NatDelPortRedirect(fmt.Sprintf("%s/32", common.Localhost), strconv.Itoa(common.StandardDnsPort), strconv.Itoa(common.AlternativeDnsPort))
			}
		}()
		if strings.HasPrefix(dnsMode, common.DnsModeLocalDns) {
			if err := Iptables.NatAddPortRedirect(fmt.Sprintf("%s/32", common.Localhost), strconv.Itoa(common.StandardDnsPort), strconv.Itoa(common.AlternativeDnsPort)); err != nil {
				dnsSignal <- err
				return
			}
		}
		dnsSignal <- DnsResolve.SetupResolvConf(dnsServer)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
	}()
	return <-dnsSignal
}

// HandleExtraDomainMapping handle extra domain change
func HandleExtraDomainMapping(extraDomains map[string]string, localDnsPort int) {
	// pass
}

// RestoreNameServer remove the nameservers added by ktctl
func RestoreNameServer() {
	DnsResolve.RestoreResolvConf()
	Iptables.NatDelPortRedirect(fmt.Sprintf("%s/32", common.Localhost), strconv.Itoa(common.StandardDnsPort), strconv.Itoa(common.AlternativeDnsPort))
}
