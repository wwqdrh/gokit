package dns

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/wwqdrh/gokit/logger"
)

var (
	FieldNameserver = "nameserver"
	FieldDomain     = "domain"
	FieldSearch     = "search"
)

type dnsResolve struct {
	commentAdd    string
	commentRemove string
	resolveConf   string
}

type dnsResolveOpt func(*dnsResolve)

func WithEscape(add, remove string) dnsResolveOpt {
	return func(dr *dnsResolve) {
		dr.commentAdd = add
		dr.commentRemove = remove
	}
}

func WithResolveConf(resolve string) dnsResolveOpt {
	return func(dr *dnsResolve) {
		dr.resolveConf = resolve
	}
}

func NewDnsResolve(opts ...dnsResolveOpt) dnsResolve {
	d := dnsResolve{
		commentAdd:    " # Added by Nettool",
		commentRemove: " # Removed by Nettool",
		resolveConf:   "/etc/resolv.conf",
	}
	for _, opt := range opts {
		opt(&d)
	}
	return d
}

func (dr dnsResolve) SetupResolvConf(dnsServer string) error {
	f, err := os.Open(dr.resolveConf)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer

	sample := fmt.Sprintf("%s %s ", FieldNameserver, strings.Split(dnsServer, ":")[0])
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, sample) {
			// required dns server already been added
			return nil
		} else if strings.HasPrefix(line, FieldNameserver) {
			buf.WriteString("#")
			buf.WriteString(line)
			buf.WriteString(dr.commentRemove)
			buf.WriteString("\n")
		} else {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	// Add nameserver and comment to resolv.conf
	nameserverIp := strings.Split(dnsServer, ":")[0]
	buf.WriteString(fmt.Sprintf("%s %s%s\n", FieldNameserver, nameserverIp, dr.commentAdd))

	stat, _ := f.Stat()
	return ioutil.WriteFile(dr.resolveConf, buf.Bytes(), stat.Mode())
}

func (dr dnsResolve) RestoreResolvConf() {
	f, err := os.Open(dr.resolveConf)
	if err != nil {
		logger.DefaultLogger.Errorx("%s: Failed to open resolve.conf during restoring", nil, err.Error())
		return
	}
	defer f.Close()

	var buf bytes.Buffer

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, dr.commentRemove) {
			line = strings.TrimSuffix(line, dr.commentRemove)
			line = strings.TrimPrefix(line, "#")
			buf.WriteString(line)
			buf.WriteString("\n")
		} else if strings.HasSuffix(line, dr.commentAdd) {
			logger.DefaultLogger.Debugx("remove line: %s ", nil, line)
		} else {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	stat, _ := f.Stat()
	if err = ioutil.WriteFile(dr.resolveConf, buf.Bytes(), stat.Mode()); err != nil {
		logger.DefaultLogger.Errorx("%s: Failed to write resolve.conf during restoring", nil, err.Error())

	}
}
