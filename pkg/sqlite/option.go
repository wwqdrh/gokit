package sqlite

import (
	"io"
	"os"
	"time"

	"gorm.io/gorm/logger"
)

var DefaultOption = option{
	charset:         "utf8",
	ParseTime:       true,
	loc:             "Asia/Shanghai",
	maxIdleConns:    10,
	maxOpenConns:    100,
	connMaxLifetime: time.Hour,
	loggerOut:       os.Stdout,
	loggerLevel:     logger.Warn,
}

type option struct {
	DbName          string
	charset         string // utf8
	ParseTime       bool   // true
	loc             string // Local
	maxIdleConns    int
	maxOpenConns    int
	connMaxLifetime time.Duration

	loggerOut   io.Writer
	loggerLevel logger.LogLevel
}

type OptionFunc = func(o *option)

func NewOption(options ...OptionFunc) *option {
	opt := DefaultOption
	for _, item := range options {
		item(&opt)
	}
	return &opt
}

func WithDbName(dbname string) OptionFunc {
	return func(o *option) {
		o.DbName = dbname
	}
}

func WithLoggerOut(w io.Writer) OptionFunc {
	return func(o *option) {
		if w == nil {
			w = os.Stdout
		}
		o.loggerOut = w
	}
}

func WithLoggerLevel(l logger.LogLevel) OptionFunc {
	return func(o *option) {
		o.loggerLevel = l
	}
}
