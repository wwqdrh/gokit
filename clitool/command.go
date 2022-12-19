package clitool

import (
	"github.com/spf13/cobra"
	"github.com/wwqdrh/gokit/logger"
)

type Command struct {
	Cmd        *cobra.Command
	Values     interface{}
	Options    []OptionConfig
	Persistent []OptionConfig

	inited bool
}

func (c *Command) Builder() {
	if !c.inited {
		if len(c.Options) > 0 {
			SetOptions(c.Cmd, c.Cmd.Flags(), c.Values, c.Options)
		}
		if len(c.Persistent) > 0 {
			SetOptions(c.Cmd, c.Cmd.PersistentFlags(), c.Values, c.Options)
		}

		c.inited = true
	}
}

func (c *Command) Add(cmd *Command) {
	c.Builder()

	c.Cmd.AddCommand(cmd.Cmd)

	cmd.Builder()
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
