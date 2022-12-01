package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wwqdrh/gotoolkit/nettool/server/ssh"
	"github.com/wwqdrh/logger"
)

func main() {
	StartServer()
	// StartSocks5Connection("", "", 22, false, 2000)
}

func StartScript() {
	channel := ssh.NewChannel(
		ssh.ChannelWithAuth("root", "123456"),
		ssh.ChannelWithRemoteSSH("ds-connect-shadow:22"),
	)

	res, err := channel.RunScript("ds-connect-shadow:22", "ping www.baidu.com -c 1")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(res)
	}
}

func StartServer() {
	channel := ssh.NewChannel(
		ssh.ChannelWithAuth("root", "123456"),
		ssh.ChannelWithRemoteSSH("ds-connect-shadow:22"),
	)

	channel.StartSocks5Proxy(context.TODO(), "0.0.0.0:8081")
}

func StartSocks5Connection(podIP, privateKey string, localSshPort int, isInitConnect bool, proxyPort int) error {
	var res = make(chan error)
	var ticker *time.Ticker
	// sshAddress := fmt.Sprintf("127.0.0.1:%d", localSshPort)
	localaddress := fmt.Sprintf("127.0.0.1:%d", proxyPort)
	gone := false

	channel := ssh.NewChannel(
		ssh.ChannelWithAuth("root", "123456"),
	)

	go func() {
		// will hang here if not error happen
		err := channel.StartSocks5Proxy(context.TODO(), localaddress)
		if !gone {
			res <- err
		}
		logger.DefaultLogger.Errorx("%e: Socks proxy interrupted", nil, err)
		if ticker != nil {
			ticker.Stop()
		}
		time.Sleep(10 * time.Second)
		logger.DefaultLogger.Debug("Socks proxy reconnecting ...")
		_ = StartSocks5Connection(podIP, privateKey, localSshPort, false, proxyPort)
	}()
	select {
	case err := <-res:
		if isInitConnect {
			logger.DefaultLogger.Warnx("%e: Failed to setup socks proxy connection", nil, err)
		}
		return err
	case <-time.After(1 * time.Second):
		ticker = channel.Socks5HeartBeat(podIP, "")
		logger.DefaultLogger.Info("Socks proxy established")
		gone = true
		return nil
	}
}
