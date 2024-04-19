package clitool

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

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
	if c.Cmd == nil {
		c.Cmd = &cobra.Command{}
	}

	if !c.inited {
		if len(c.Options) == 0 && len(c.Persistent) == 0 {
			if options, err := c.optionByValues(c.Values); err != nil {
				logger.DefaultLogger.Error(err.Error())
			} else {
				for _, item := range options {
					if item.Persistent {
						c.Persistent = append(c.Persistent, *item)
					} else {
						c.Options = append(c.Options, *item)
					}
				}
			}
		}

		if len(c.Options) > 0 {
			SetOptions(c.Cmd, c.Cmd.Flags(), c.Values, c.Options)
		}
		if len(c.Persistent) > 0 {
			SetOptions(c.Cmd, c.Cmd.PersistentFlags(), c.Values, c.Persistent)
		}
		if c.Cmd != nil {
			fn := c.Cmd.PreRun
			c.Cmd.PreRun = func(cmd *cobra.Command, args []string) {
				c.echo()
				if fn != nil {
					fn(cmd, args)
				}
			}
		}

		c.inited = true
	}
}

func (c *Command) optionByValues(values interface{}) (map[string]*OptionConfig, error) {
	options := map[string]*OptionConfig{}

	if targets, err := GetTagValue(values, "name"); err == nil {
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

	if targets, err := GetTagValue(values, "alias"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				options[key] = &OptionConfig{
					Target: key,
				}
			}
			options[key].Alias = val
		}
	}

	if targets, err := GetTagValue(values, "desc"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				options[key] = &OptionConfig{
					Target: key,
				}
			}
			options[key].Description = val
		}
	}

	if targets, err := GetTagValue(values, "hidden"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				options[key] = &OptionConfig{
					Target: key,
				}
			}

			options[key].Hidden = val == "true"
		}
	}

	if targets, err := GetTagValue(values, "required"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				options[key] = &OptionConfig{
					Target: key,
				}
			}

			options[key].Required = val == "true"
		}
	}

	if targets, err := GetTagValue(values, "persistent"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				options[key] = &OptionConfig{
					Target: key,
				}
			}

			options[key].Persistent = val == "true"
		}
	}

	if targets, err := GetTagValue(values, "echo"); err == nil {
		for key, val := range targets {
			if _, ok := options[key]; !ok {
				options[key] = &OptionConfig{
					Target: key,
				}
			}

			options[key].ShouldEcho = val == "true"
		}
	}

	return options, nil
}

func (c *Command) Add(cmd *Command) {
	c.Builder()

	c.Cmd.AddCommand(cmd.Cmd)

	cmd.Builder()
}

func (cm *Command) echo() error {
	options, err := cm.optionByValues(cm.Values)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)
	for _, c := range options {
		if !c.ShouldEcho {
			continue
		}
		name := UnCapitalize(c.Target)
		if c.Name != "" {
			name = c.Name
		}
		field := reflect.ValueOf(cm.Values).Elem().FieldByName(c.Target)

		switch c.DefaultValue.(type) {
		case string:
			fieldPtr := (*string)(unsafe.Pointer(field.UnsafeAddr()))

			fmt.Printf("%s - %s default is: (%v)\n", name, c.Description, c.DefaultValue)
			scanner.Scan()
			if data := strings.TrimSpace(scanner.Text()); data != "" {
				*fieldPtr = data
			}
		case int:
			fieldPtr := (*int)(unsafe.Pointer(field.UnsafeAddr()))
			fmt.Printf("%s - %s default is: (%v)\n", name, c.Description, c.DefaultValue)
			scanner.Scan()
			if data := strings.TrimSpace(scanner.Text()); data != "" {
				if v64, err := strconv.ParseInt(data, 10, 64); err != nil {
					fmt.Println(err.Error())
				} else {
					*fieldPtr = int(v64)
				}
			}
		case bool:
			fieldPtr := (*bool)(unsafe.Pointer(field.UnsafeAddr()))
			fmt.Printf("%s - %s default is: (%v)\n", name, c.Description, c.DefaultValue)
			scanner.Scan()
			if data := strings.TrimSpace(scanner.Text()); data != "" {
				*fieldPtr = data == "true"
			}
		case []string:
			fieldPtr := (*[]string)(unsafe.Pointer(field.UnsafeAddr()))
			fmt.Printf("%s - %s default is: (%s)\n", c.Name, c.Description, strings.Trim(fmt.Sprint(c.DefaultValue), "[]"))
			scanner.Scan()
			if data := strings.TrimSpace(scanner.Text()); data != "" {
				*fieldPtr = strings.Split(data, ",")
			}
		case []int:
			fieldPtr := (*[]int)(unsafe.Pointer(field.UnsafeAddr()))
			fmt.Printf("%s - %s default is: (%v)\n", name, c.Description, c.DefaultValue)
			scanner.Scan()
			if data := strings.TrimSpace(scanner.Text()); data != "" {
				v := []int{}
				ok := true
				for _, item := range strings.Split(data, ",") {
					if v64, err := strconv.ParseInt(item, 10, 64); err != nil {
						fmt.Println(err.Error())
						ok = false
						break
					} else {
						v = append(v, int(v64))
					}
				}
				if ok {
					*fieldPtr = v
				}
			}
		case []bool:
			fieldPtr := (*[]bool)(unsafe.Pointer(field.UnsafeAddr()))
			fmt.Printf("%s - %s default is: (%v)\n", name, c.Description, c.DefaultValue)
			scanner.Scan()
			if data := strings.TrimSpace(scanner.Text()); data != "" {
				v := []bool{}
				for _, item := range strings.Split(data, ",") {
					v = append(v, item == "true")
				}
				*fieldPtr = v
			}
		}
	}
	return nil
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
