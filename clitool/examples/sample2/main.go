package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wwqdrh/gokit/clitool"
)

var Opt = struct {
	Name string `name:"name" persistent:"true"`
}{}

var SubOpt = struct {
	Echo bool     `name:"echo" echo:"true" alias:"e"`
	Data string   `name:"data" echo:"true" desc:"docker swarm集群中的shadow服务地址"`
	Args []string `name:"args" echo:"true"`
}{
	Echo: false,
	Data: "hh213",
}

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
		Values: &Opt,
	}

	cmd.Add(&clitool.Command{
		Cmd: &cobra.Command{
			Use:   "sub",
			Short: "sub",
			PreRun: func(cmd *cobra.Command, args []string) {
				fmt.Println("try prerun")
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				if SubOpt.Echo {
					fmt.Println(Opt.Name, SubOpt.Data, SubOpt.Args)
				}
				return nil
			},
		},
		Values: &SubOpt,
	})

	cmd.Run()
}
