package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wwqdrh/gokit/logger"
	"github.com/wwqdrh/gotookit/clitool"
)

var connectFlags = []clitool.OptionConfig{
	{
		Target:       "Shadow",
		DefaultValue: "127.0.0.1:18080",
		Description:  "docker swarm集群中的shadow服务地址",
	},
	{
		Target:       "SshUser",
		DefaultValue: "root",
		Description:  "ssh用户名",
	},
	{
		Target:       "SshPass",
		DefaultValue: "",
		Description:  "ssh密码",
	},
	{
		Target:       "ProxyPort",
		DefaultValue: 2223,
		Description:  "(tun2socks mode only) Specify the local port which socks5 proxy should use",
	},
	{
		Target:       "DnsPort",
		DefaultValue: 53,
		Description:  "本地dns服务的端口",
	},
	{
		Target:       "DnsCacheTtl",
		DefaultValue: 60,
		Description:  "(local dns mode only) DNS cache refresh interval in seconds",
	},
}

type ConnectOptions struct {
	Shadow      string
	SshUser     string
	SshPass     string
	ProxyPort   int
	DnsPort     int
	DnsCacheTtl int
}

// DaemonOptions cli options
type DaemonOptions struct {
	Connect *ConnectOptions
}

var opt *DaemonOptions = &DaemonOptions{
	Connect: &ConnectOptions{},
}

// Get fetch options instance
func Opts() *DaemonOptions {
	return opt
}

func main() {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Create a network tunnel to docker swarm cluster",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("too many options specified (%s)", strings.Join(args, ","))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("do run", opt.Connect.Shadow, opt.Connect.SshPass)
			return nil
		},
		Example: "ktctl connect [command options]",
	}

	clitool.SetOptions(cmd, cmd.Flags(), Opts().Connect, connectFlags)

	var rootCmd = &cobra.Command{
		Use:     "simple",
		Version: "0.0.1",
		Short:   "A simple cli",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		Example: "dsctl <command> [command options]",
	}

	rootCmd.AddCommand(cmd)
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	// process will hang here
	if err := rootCmd.Execute(); err != nil {
		logger.DefaultLogger.Error("Exit: " + err.Error())
	}
}
