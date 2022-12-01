package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/wwqdrh/gotoolkit/nettool/server/dns"
)

// address books
var (
	addressBookOfA = map[string][4]byte{
		"www.baidu.com.": [4]byte{220, 181, 38, 150},
	}
	addressBookOfPTR = map[string]string{
		"150.38.181.220.in-addr.arpa.": "www.baidu.com.",
	}
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dnsserver := dns.NewDnsServer(func(domain string) [4]byte {
		return addressBookOfA[domain]
	}, func(ip string) string {
		return addressBookOfPTR[ip]
	}, 5353)
	go dnsserver.Server(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
}
