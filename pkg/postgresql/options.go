package postgresql

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
	UserName        string
	Password        string
	Host            string
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

func WithUserName(username string) OptionFunc {
	return func(o *option) {
		o.UserName = username
	}
}

func WithPassword(password string) OptionFunc {
	return func(o *option) {
		o.Password = password
	}
}

func WithHost(host string) OptionFunc {
	return func(o *option) {
		o.Host = host
	}
}

func WithDbName(dbname string) OptionFunc {
	return func(o *option) {
		o.DbName = dbname
	}
}

func WithLoggerOut(w io.Writer) OptionFunc {
	return func(o *option) {
		o.loggerOut = w
	}
}

func WithLoggerLevel(l logger.LogLevel) OptionFunc {
	return func(o *option) {
		o.loggerLevel = l
	}
}
