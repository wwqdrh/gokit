package common

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func IpAndMask(cidr string) (string, string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}
	val := make([]byte, len(ipNet.Mask))
	copy(val, ipNet.Mask)

	var s []string
	for _, i := range val[:] {
		s = append(s, strconv.Itoa(int(i)))
	}
	return ipNet.IP.String(), strings.Join(s, "."), nil
}

func IpNetPart(cidr string) (string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	size, _ := ipNet.Mask.Size()
	return fmt.Sprintf("%s/%d", ipNet.IP.String(), size), nil
}

func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("x-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return ip
	}
	return ""
}

// 1、10. - 10.
// 2、172.16 - 172.31
// 3、192.168 - 192.168
// 4、127.0.0.1
func IsLocalIp(ip string) bool {
	if ip == "127.0.0.1" {
		return true
	}

	ipAddr := strings.Split(ip, ".")

	if strings.EqualFold(ipAddr[0], "10") {
		return true
	} else if strings.EqualFold(ipAddr[0], "172") {
		addr, _ := strconv.Atoi(ipAddr[1])
		if addr >= 16 && addr < 31 {
			return true
		}
	} else if strings.EqualFold(ipAddr[0], "192") && strings.EqualFold(ipAddr[1], "168") {
		return true
	}
	return false
}
