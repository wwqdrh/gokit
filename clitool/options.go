package clitool

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unsafe"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/wwqdrh/gokit/logger"
	"gopkg.in/yaml.v2"
)

type OptionConfig struct {
	Target       string
	Name         string
	Alias        string
	DefaultValue any
	Description  string
	Hidden       bool
	Required     bool
}

func SetOptions(cmd *cobra.Command, flags *flag.FlagSet, optionStore any, config []OptionConfig) {
	cmd.Long = cmd.Short
	cmd.Flags().SortFlags = false
	cmd.InheritedFlags().SortFlags = false
	flags.SortFlags = false
	for _, c := range config {
		name := UnCapitalize(c.Target)
		if c.Name != "" {
			name = c.Name
		}
		field := reflect.ValueOf(optionStore).Elem().FieldByName(c.Target)
		switch c.DefaultValue.(type) {
		case string:
			fieldPtr := (*string)(unsafe.Pointer(field.UnsafeAddr()))
			defaultValue := c.DefaultValue.(string)
			if field.String() != "" {
				defaultValue = field.String()
			}
			if c.Alias != "" {
				flags.StringVarP(fieldPtr, name, c.Alias, defaultValue, c.Description)
			} else {
				flags.StringVar(fieldPtr, name, defaultValue, c.Description)
			}
		case int:
			defaultValue := c.DefaultValue.(int)
			if field.Int() != 0 {
				defaultValue = int(field.Int())
			}
			fieldPtr := (*int)(unsafe.Pointer(field.UnsafeAddr()))
			if c.Alias != "" {
				flags.IntVarP(fieldPtr, name, c.Alias, defaultValue, c.Description)
			} else {
				flags.IntVar(fieldPtr, name, defaultValue, c.Description)
			}
		case bool:
			defaultValue := c.DefaultValue.(bool)
			if field.Bool() {
				defaultValue = field.Bool()
			}
			fieldPtr := (*bool)(unsafe.Pointer(field.UnsafeAddr()))
			if c.Alias != "" {
				flags.BoolVarP(fieldPtr, name, c.Alias, defaultValue, c.Description)
			} else {
				flags.BoolVar(fieldPtr, name, defaultValue, c.Description)
			}
		}
		if c.Hidden {
			_ = flags.MarkHidden(name)
		}
		if c.Required {
			_ = cmd.MarkFlagRequired(name)
		}
	}
}

// opt, a config struct
func mergeOptions(opt interface{}, data []byte) {
	config := make(map[string]map[string]string)
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		logger.DefaultLogger.Warn("Invalid config content, skipping ...")
		return
	}
	for group, item := range config {
		for key, value := range item {
			groupField := reflect.ValueOf(opt).Elem().FieldByName(Capitalize(group))
			if groupField.IsValid() {
				itemField := groupField.Elem().FieldByName(Capitalize(key))
				if itemField.IsValid() {
					switch itemField.Kind() {
					case reflect.String:
						itemField.SetString(value)
					case reflect.Int:
						if v, err2 := strconv.Atoi(value); err2 == nil {
							itemField.SetInt(int64(v))
						} else {
							logger.DefaultLogger.Warnx("Config item '%s.%s' value is not integer: %s", nil, group, key, value)
						}
					case reflect.Bool:
						if v, err2 := strconv.ParseBool(value); err2 == nil {
							itemField.SetBool(v)
						} else {
							logger.DefaultLogger.Warnx("Config item '%s.%s' value is not bool: %s", nil, group, key, value)
						}
					default:
						logger.DefaultLogger.Warnx("Config item '%s.%s' of invalid type: %s", nil,
							group, key, itemField.Kind().String())
					}
					logger.DefaultLogger.Debugx("Loaded %s.%s = %s", nil, group, key, value)
				}
			}
		}
	}
}

// Capitalize convert dash separated string to capitalized string
func Capitalize(word string) string {
	prev := '-'
	capitalized := strings.Map(
		func(r rune) rune {
			if prev == '-' {
				prev = r
				return unicode.ToUpper(r)
			}
			prev = r
			return unicode.ToLower(r)
		},
		word)
	return strings.ReplaceAll(capitalized, "-", "")
}

func UnCapitalize(word string) string {
	firstLetter := true
	return strings.Map(
		func(r rune) rune {
			if firstLetter {
				firstLetter = false
				return unicode.ToLower(r)
			}
			return r
		},
		word)
}
