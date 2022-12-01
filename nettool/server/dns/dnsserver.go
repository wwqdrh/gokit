package dns

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/wwqdrh/gotoolkit/nettool/common"
	"github.com/wwqdrh/logger"
)

type DnsServer struct {
	dnsAddresses []string
	extraDomains map[string]string
}

func SetupLocalDns(remoteDnsPort, localDnsPort int, dnsOrder []string, ttl int64) error {
	var res = make(chan error)
	go func() {
		upstreamDnsAddresses := getDnsAddresses(dnsOrder, GetNameServer(), remoteDnsPort)
		// domain-name -> ip
		extraDomains := map[string]string{}
		logger.DefaultLogger.Infox("Setup local DNS with upstream %v", nil, upstreamDnsAddresses)
		HandleExtraDomainMapping(extraDomains, localDnsPort)
		res <- SetupDnsServer(&DnsServer{upstreamDnsAddresses, extraDomains}, localDnsPort, "udp")
	}()
	select {
	case err := <-res:
		return err
	case <-time.After(1 * time.Second):
		return nil
	}
}

func getDnsAddresses(dnsOrder []string, upstreamDns string, clusterDnsPort int) []string {
	upstreamPattern := fmt.Sprintf("^([cdptu]{3}:)?%s(:[0-9]+)?$", DnsOrderUpstream)
	var dnsAddresses []string
	for _, dnsAddr := range dnsOrder {
		if dnsAddr == DnsOrderCluster {
			dnsAddresses = append(dnsAddresses, fmt.Sprintf("tcp:%s:%d", common.Localhost, clusterDnsPort))
		} else if ok, err := regexp.MatchString(upstreamPattern, dnsAddr); err == nil && ok {
			upstreamParts := strings.Split(dnsAddr, ":")
			if upstreamDns != "" {
				switch strings.Count(dnsAddr, ":") {
				case 0:
					dnsAddresses = append(dnsAddresses, fmt.Sprintf("udp:%s:%d", upstreamDns, common.StandardDnsPort))
				case 1:
					if _, err = strconv.Atoi(upstreamParts[1]); err == nil {
						dnsAddresses = append(dnsAddresses, fmt.Sprintf("udp:%s:%s", upstreamDns, upstreamParts[1]))
					} else {
						dnsAddresses = append(dnsAddresses, fmt.Sprintf("%s:%s:%d", upstreamParts[0], upstreamDns, common.StandardDnsPort))
					}
				case 2:
					dnsAddresses = append(dnsAddresses, fmt.Sprintf("%s:%s:%s", upstreamParts[0], upstreamDns, upstreamParts[2]))
				default:
					logger.DefaultLogger.Warnx("Skip invalid upstream dns server %s", nil, dnsAddr)
				}
			}
		} else {
			switch strings.Count(dnsAddr, ":") {
			case 0:
				dnsAddresses = append(dnsAddresses, fmt.Sprintf("udp:%s:%d", dnsAddr, common.StandardDnsPort))
			case 1:
				if _, err = strconv.Atoi(strings.Split(dnsAddr, ":")[1]); err == nil {
					dnsAddresses = append(dnsAddresses, fmt.Sprintf("udp:%s", dnsAddr))
				} else {
					dnsAddresses = append(dnsAddresses, fmt.Sprintf("%s:%d", dnsAddr, common.StandardDnsPort))
				}
			case 2:
				dnsAddresses = append(dnsAddresses, dnsAddr)
			default:
				logger.DefaultLogger.Warn("Skip invalid dns server " + dnsAddr)
			}
		}
	}
	return dnsAddresses
}

// ServeDNS query DNS record
func (s *DnsServer) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	msg := (&dns.Msg{}).SetReply(req)
	msg.Authoritative = true
	msg.Answer = query(req, s.dnsAddresses, s.extraDomains, 10)
	if err := w.WriteMsg(msg); err != nil {
		logger.DefaultLogger.Warnx("%e: Failed to reply dns request", nil, err)
	}
}

func query(req *dns.Msg, dnsAddresses []string, extraDomains map[string]string, ttl int64) []dns.RR {
	domain := req.Question[0].Name
	qtype := req.Question[0].Qtype

	answer := ReadCache(domain, qtype, ttl)
	if answer != nil {
		logger.DefaultLogger.Debugx("Found domain %s (%d) in cache", nil, domain, qtype)
		return answer
	}

	for host, ip := range extraDomains {
		if wildcardMatch(host, domain) {
			return []dns.RR{toARecord(domain, ip)}
		}
	}

	for _, dnsAddr := range dnsAddresses {
		dnsParts := strings.SplitN(dnsAddr, ":", 3)
		protocol := dnsParts[0]
		ip := dnsParts[1]
		port, err := strconv.Atoi(dnsParts[2])
		if ip == "" || err != nil || (protocol != "tcp" && protocol != "udp") {
			// skip invalid dns address
			continue
		}
		res, err := NsLookup(domain, qtype, protocol, fmt.Sprintf("%s:%d", ip, port))
		if res != nil && len(res.Answer) > 0 {
			// only record none-empty result of cluster dns
			logger.DefaultLogger.Debugx("Found domain %s (%d) in dns (%s:%d)", nil, domain, qtype, ip, port)
			WriteCache(domain, qtype, res.Answer, time.Now().Unix())
			return res.Answer
		} else if err != nil && !IsDomainNotExist(err) {
			// usually io timeout error
			logger.DefaultLogger.Warnx("%e: Failed to lookup %s (%d) in dns (%s:%d)", nil, err, domain, qtype, ip, port)
		}
	}
	logger.DefaultLogger.Debugx("Empty answer for domain lookup %s (%d)", nil, domain, qtype)
	WriteCache(domain, qtype, []dns.RR{}, time.Now().Unix()-ttl/2)
	return []dns.RR{}
}

func wildcardMatch(pattenDomain, targetDomain string) bool {
	if !strings.HasSuffix(pattenDomain, ".") {
		pattenDomain = pattenDomain + "."
	}
	if strings.Contains(pattenDomain, "*") {
		ok, err := regexp.MatchString("^"+strings.ReplaceAll(strings.ReplaceAll(pattenDomain, ".", "\\."), "*", ".*")+"$", targetDomain)
		return ok && err == nil
	} else {
		return pattenDomain == targetDomain
	}
}

func getDnsOrder(dnsMode string) []string {
	if !strings.Contains(dnsMode, ":") {
		return []string{DnsOrderCluster, DnsOrderUpstream}
	}
	return strings.Split(strings.SplitN(dnsMode, ":", 2)[1], ",")
}
