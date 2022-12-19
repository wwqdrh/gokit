package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wwqdrh/gokit/clitool"
)

var Opt = struct {
	Name string
}{}

var SubOpt = struct {
	Data string
}{}

func main() {
	cmd := clitool.Command{
		Cmd: &cobra.Command{
			Use:   "connect",
			Short: "Create a network tunnel to docker swarm cluster",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("do run", Opt.Name)
				return nil
			},
			Example: "ktctl connect [command options]",
		},
		Persistent: []clitool.OptionConfig{
			{
				Target:       "Name",
				Name:         "name",
				DefaultValue: "127.0.0.1:18080",
				Description:  "docker swarm集群中的shadow服务地址",
			},
		},
		Values: &Opt,
	}

	cmd.Add(&clitool.Command{
		Cmd: &cobra.Command{
			Use:   "sub",
			Short: "sub",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println(Opt.Name, SubOpt.Data)
				return nil
			},
		},
		Options: []clitool.OptionConfig{
			{
				Target:       "Data",
				Name:         "data",
				DefaultValue: "12data0",
				Description:  "docker swarm集群中的shadow服务地址",
			},
		},
		Values: &SubOpt,
	})

	cmd.Run()
}
