package tun

import (
	"context"
	"fmt"
	"io"

	"github.com/wwqdrh/logger"
	"github.com/xjasonlyu/tun2socks/v2/engine"
	tunLog "github.com/xjasonlyu/tun2socks/v2/log"
)

// Tunnel ...
type Tunnel interface {
	CheckContext() error
	ToSocks(ctx context.Context, sockAddr string) error
	SetRoute(ipRange []string, excludeIpRange []string) error
	CheckRoute(ipRange []string) []string
	RestoreRoute() error
	GetName() string
}

// Cli the singleton type
type Cli struct{}

var instance *Cli

// Ins get singleton instance
func Ins() Tunnel {
	if instance == nil {
		instance = &Cli{}
	}
	return instance
}

// ToSocks create a tun and connect to socks endpoint
func (s *Cli) ToSocks(ctx context.Context, sockAddr string) error {
	tunSignal := make(chan error)
	logLevel := "warning"
	// if opt.Get().Global.Debug {
	// 	logLevel = "debug"
	// }
	go func() {
		var key = new(engine.Key)
		key.Proxy = sockAddr
		key.Device = fmt.Sprintf("tun://%s", s.GetName())
		key.LogLevel = logLevel
		tunLog.SetOutput(io.Discard)
		engine.Insert(key)
		tunSignal <- engine.Start()

		defer func() {
			if err := engine.Stop(); err != nil {
				logger.DefaultLogger.Errorx("%s: Stop tun device %s failed", nil, err.Error(), key.Device)
			} else {
				logger.DefaultLogger.Infox("Tun device %s stopped", nil, key.Device)
			}
		}()

		<-ctx.Done()
	}()
	return <-tunSignal
}
