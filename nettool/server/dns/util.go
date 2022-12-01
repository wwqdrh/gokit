package dns

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/wwqdrh/logger"
)

var (
	ResolvConf = "/etc/resolv.conf"
)

const (
	StandardDnsPort    = 53
	AlternativeDnsPort = 10053
)

func ArrayEquals(src, target []string) bool {
	if len(src) != len(target) {
		return false
	}
	for i := 0; i < len(src); i++ {
		found := false
		for j := 0; j < len(target); j++ {
			if src[i] == target[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// CreateDirIfNotExist create dir
func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// Contains check whether target exist in container, the type of container can be an array, slice or map
func Contains(container any, target any) bool {
	targetValue := reflect.ValueOf(container)
	switch reflect.TypeOf(container).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == target {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(target)).IsValid() {
			return true
		}
	}
	return false
}

const IpAddrPattern = "[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+"

func IsValidIp(ip string) bool {
	if ok, err := regexp.MatchString("^"+IpAddrPattern+"$", ip); ok && err == nil {
		return true
	}
	return false
}

// DomainNotExistError ...
type DomainNotExistError struct {
	name  string
	qtype uint16
}

func (e DomainNotExistError) Error() string {
	return fmt.Sprintf("domain %s (%d) not exist", e.name, e.qtype)
}

// IsDomainNotExist check the error type
func IsDomainNotExist(err error) bool {
	_, exists := err.(DomainNotExistError)
	return exists
}

// GetRandomTcpPort get pod random ssh port
func GetRandomTcpPort() int {
	for i := 0; i < 20; i++ {
		port := RandomPort()
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			logger.DefaultLogger.Debugx("Port %d not available", nil, port)
			_ = conn.Close()
		} else {
			logger.DefaultLogger.Debugx("Using port %d", nil, port)
			return port
		}
	}
	port := RandomPort()
	logger.DefaultLogger.Debugx("Using random port %d", nil, port)
	return port
}

func RandomPort() int {
	return rand.Intn(65535-1024) + 1024
}

func string2IP(ip string) [4]byte {
	res := [4]byte{}
	for i, item := range strings.SplitN(ip, ".", 4) {
		ch, _ := strconv.Atoi(item)
		res[i] = byte(ch)
	}
	return res
}
