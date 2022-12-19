package clitool

import (
	"github.com/spf13/cobra"
	"github.com/wwqdrh/gokit/logger"
)

type Command struct {
	Cmd     *cobra.Command
	Values  interface{}
	Options []OptionConfig
}

func (c *Command) Builder() {
	SetOptions(c.Cmd, c.Cmd.Flags(), c.Values, c.Options)
}

func (c *Command) Add(cmd *Command) {
	cmd.Builder()

	c.Cmd.AddCommand(cmd.Cmd)
}

func (c *Command) Run() {
	c.Builder()

	c.Cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	c.Cmd.SilenceUsage = true
	c.Cmd.SilenceErrors = true
	// process will hang here
	if err := c.Cmd.Execute(); err != nil {
		logger.DefaultLogger.Error("Exit: " + err.Error())
	}
}
