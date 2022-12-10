package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/wwqdrh/gokit/logger"
	"github.com/wwqdrh/gokit/nettool/server/socks5"
	"github.com/wwqdrh/gokit/nettool/server/ssh/client"
	"golang.org/x/net/proxy"
)

// Channel network channel
type IChannel interface {
	StartSocks5Proxy(ctx context.Context, local string) error
	Socks5HeartBeat(remoteIP, socks5Address string) *time.Ticker
	ForwardRemoteToLocal(privateKey, sshAddress, remoteEndpoint, localEndpoint string) error
	RunScript(sshAddress, script string) (string, error)
}

type channel struct {
	username   string
	password   string
	privateKey string
	sshAddress string
}

type channelOpt func(o *channel)

func ChannelWithAuth(username, password string) channelOpt {
	return func(o *channel) {
		o.username = username
		o.password = password
	}
}

func ChannelWithPrivateKey(privateKey string) channelOpt {
	return func(o *channel) {
		o.privateKey = privateKey
	}
}

func ChannelWithRemoteSSH(sshAddress string) channelOpt {
	return func(o *channel) {
		o.sshAddress = sshAddress
	}
}

func NewChannel(opts ...channelOpt) IChannel {
	c := &channel{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *channel) getSshTunnelAddress() string {
	if c.privateKey != "" {
		return fmt.Sprintf("ssh://%s@%s?identity_file=%s", c.username, c.sshAddress, c.privateKey)
	} else {
		return fmt.Sprintf("ssh://%s:%s@%s", c.username, c.password, c.sshAddress)
	}
}

// StartSocks5Proxy start socks5 proxy
func (c *channel) StartSocks5Proxy(ctx context.Context, local string) (err error) {
	dialer, err := client.NewDialer(c.getSshTunnelAddress())
	if err != nil {
		return err
	}
	defer dialer.Close()

	svc := &socks5.Server{
		Logger:    SocksLogger{},
		ProxyDial: dialer.DialContext,
	}
	go func() {
		if err := svc.ListenAndServe("tcp", local); err != nil {
			logger.DefaultLogger.Error(err.Error())
		}
	}()
	<-ctx.Done()
	return nil
}

func (c *channel) Socks5HeartBeat(remoteIP, socks5Address string) *time.Ticker {
	dialer, err := proxy.SOCKS5("tcp", socks5Address, nil, proxy.Direct)
	if err != nil {
		logger.DefaultLogger.Warnx("%e: Failed to create socks proxy heart beat ticker", nil, err)
	}
	ticker := time.NewTicker(60 * time.Second)
	go func() {
	TickLoop:
		for {
			select {
			case <-ticker.C:
				if c, err2 := dialer.Dial("tcp", fmt.Sprintf("%s:%d", remoteIP, StandardSshPort)); err2 != nil {
					logger.DefaultLogger.Debugx("%e: Socks proxy heartbeat interrupted", nil, err)
				} else {
					_ = c.Close()
					logger.DefaultLogger.Debugx("Heartbeat socks proxy ticked at %s", nil, time.Now().Format("2006-01-02 15:04:05"))
				}
			case <-time.After(2 * 60 * time.Second):
				logger.DefaultLogger.Debug("Socks proxy heartbeat stopped")
				break TickLoop
			}
		}
	}()
	return ticker
}

// RunScript run the script on remote host.
func (c *channel) RunScript(sshAddress, script string) (result string, err error) {
	dialer, err := client.NewDialer(c.getSshTunnelAddress())
	if err != nil {
		return "", err
	}
	defer dialer.Close()

	conn, err := dialer.SSHClient(context.Background())
	if err != nil {
		logger.DefaultLogger.Errorx("%s: Failed to create ssh tunnel", nil, err.Error())
		return "", err
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		logger.DefaultLogger.Errorx("%s: Failed to create ssh session", nil, err.Error())
		return "", err
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	err = session.Run(script)
	if err != nil {
		logger.DefaultLogger.Errorx("%s: Failed to run ssh script", nil, err.Error())
		return "", err
	}
	output := stdoutBuf.String()
	return output, nil
}

// ForwardRemoteToLocal forward remote request to local
func (c *channel) ForwardRemoteToLocal(privateKey, sshAddress, remoteEndpoint, localEndpoint string) error {
	// Handle incoming connections on reverse forwarded tunnel
	dialer, err := client.NewDialer(getSshTunnelAddress(privateKey, sshAddress))
	if err != nil {
		return err
	}
	defer dialer.Close()

	_, err = dialer.SSHClient(context.Background())
	if err != nil {
		logger.DefaultLogger.Errorx("%s: Failed to create ssh tunnel", nil, err.Error())
		return err
	}

	// Listen on remote server port of shadow pod, via ssh connection
	listener, err := dialer.Listen(context.Background(), "tcp", remoteEndpoint)
	if err != nil {
		logger.DefaultLogger.Errorx("%s: Failed to listen remote endpoint", nil, err.Error())
		disconnectRemotePort(privateKey, sshAddress, remoteEndpoint, c)
		return err
	}
	defer listener.Close()

	logger.DefaultLogger.Infox("Reverse tunnel %s -> %s established", nil, remoteEndpoint, localEndpoint)
	for {
		if err = handleRequest(listener, localEndpoint); errors.Is(err, io.EOF) {
			return err
		}
	}
}

func getSshTunnelAddress(privateKey string, sshAddress string) string {
	return fmt.Sprintf("ssh://root@%s?identity_file=%s", sshAddress, privateKey)
}

func disconnectRemotePort(privateKey, sshAddress, remoteEndpoint string, c IChannel) {
	remotePort := strings.Split(remoteEndpoint, ":")[1]
	out, err := c.RunScript(sshAddress, fmt.Sprintf("/disconnect.sh %s", remotePort))
	if out != "" {
		_, _ = BackgroundLogger.Write([]byte(out + Eol))
	}
	if err != nil {
		logger.DefaultLogger.Errorx("%s: Failed to disconnect remote port %s", nil, err.Error(), remotePort)
	}
}

func handleRequest(listener net.Listener, localEndpoint string) error {
	defer func() {
		if r := recover(); r != nil {
			logger.DefaultLogger.Errorx("Failed to handle request: %v", nil, r)
		}
	}()

	// Wait requests from remote endpoint
	client, err := listener.Accept()
	if err != nil {
		logger.DefaultLogger.Errorx("%s: Failed to accept remote request", nil, err.Error())
		if !errors.Is(err, io.EOF) {
			time.Sleep(2 * time.Second)
		}
		return err
	}

	// Open a (local) connection to localEndpoint whose content will be forwarded to remoteEndpoint
	local, err := net.Dial("tcp", localEndpoint)
	if err != nil {
		_ = client.Close()
		logger.DefaultLogger.Errorx("%s: Local service error", nil, err.Error())
		return err
	}

	// Handle request in individual coroutine, current coroutine continue to accept more requests
	go handleClient(client, local)
	return nil
}

func handleClient(client net.Conn, remote net.Conn) {
	done := make(chan int)

	// Start remote -> local data transfer
	remoteReader := NewInterpretableReader(remote)
	go func() {
		defer handleBrokenTunnel(done)
		if _, err := io.Copy(client, remoteReader); err != nil {
			logger.DefaultLogger.Errorx("%s: Error while copy remote->local", nil, err.Error())
		}
		done <- 1
	}()

	// Start local -> remote data transfer
	localReader := NewInterpretableReader(client)
	go func() {
		defer handleBrokenTunnel(done)
		if _, err := io.Copy(remote, localReader); err != nil {
			logger.DefaultLogger.Errorx("%s: Error while copy local->remote", nil, err.Error())
		}
		done <- 1
	}()

	<-done
	remoteReader.Cancel()
	localReader.Cancel()
	_ = remote.Close()
	_ = client.Close()
}

func handleBrokenTunnel(done chan int) {
	if r := recover(); r != nil {
		logger.DefaultLogger.Errorx("Ssh tunnel broken: %v", nil, r)
		done <- 1
	}
}
