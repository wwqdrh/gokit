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
	inited     bool
}

func (c *Command) Builder() {
	if !c.inited {
		if len(c.Options) == 0 && len(c.Persistent) == 0 {
			if err := c.OptionByValues(); err != nil {
				logger.DefaultLogger.Error(err.Error())
			}
		}

		if len(c.Options) > 0 {
			SetOptions(c.Cmd, c.Cmd.Flags(), c.Values, c.Options)
		}
		if len(c.Persistent) > 0 {
			SetOptions(c.Cmd, c.Cmd.PersistentFlags(), c.Values, c.Persistent)
		}

		c.inited = true
	}
}

func (c *Command) OptionByValues() error {
	options := map[string]*OptionConfig{}

	if targets, err := GetTagValue(c.Values, "name"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				defaultvalue, err := GetValue(c.Values, key)
				if err != nil {
					continue
				}
				options[key] = &OptionConfig{
					Target:       key,
					DefaultValue: defaultvalue,
				}
			}
			options[key].Name = val
		}
	}

	if targets, err := GetTagValue(c.Values, "alias"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				options[key] = &OptionConfig{
					Target: key,
				}
			}
			options[key].Alias = val
		}
	}

	if targets, err := GetTagValue(c.Values, "desc"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				options[key] = &OptionConfig{
					Target: key,
				}
			}
			options[key].Description = val
		}
	}

	if targets, err := GetTagValue(c.Values, "hidden"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				options[key] = &OptionConfig{
					Target: key,
				}
			}

			options[key].Hidden = val == "true"
		}
	}

	if targets, err := GetTagValue(c.Values, "required"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				options[key] = &OptionConfig{
					Target: key,
				}
			}

			options[key].Required = val == "true"
		}
	}

	if targets, err := GetTagValue(c.Values, "persistent"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				options[key] = &OptionConfig{
					Target: key,
				}
			}

			options[key].Persistent = val == "true"
		}
	}

	for _, item := range options {
		if item.Persistent {
			c.Persistent = append(c.Persistent, *item)
		} else {
			c.Options = append(c.Options, *item)
		}
	}
	return nil
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
