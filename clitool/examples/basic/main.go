package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	name string
	data string
)

func main() {
	rootcmd := &cobra.Command{
		Use: "basic",
	}

	subcmd := &cobra.Command{
		Use: "sub",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(name, data)
			return nil
		},
	}
	rootcmd.AddCommand(subcmd)

	subcmd.Flags().StringVar(&data, "data", "defaultdata", "set a data")
	rootcmd.PersistentFlags().StringVar(&name, "name", "defaultname", "set a name")

	rootcmd.Execute()
}
