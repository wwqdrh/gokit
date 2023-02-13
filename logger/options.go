package logger

import (
	"time"

	"go.uber.org/zap/zapcore"
)

type ctxKey string

type LoggerOptions struct {
	// 基础配置
	Name         string
	Color        bool
	Console      bool // 如非必要不输出到控制台，例如开启fluentd就不需要输出，除非是fluentd失败
	Switch       bool // 是否支持动态修改等级
	SwitchTime   time.Duration
	CtxKey       ctxKey
	DefaultName  string // 指定了日志存储目录之后，如果在执行日志操作时不指定使用哪个label的话，默认会使用的名字
	HttpMetrices bool   // 是否开启http接口，用于追踪当前日志
	HttpPort     int    // 运行http接口的端口号

	// encoder config
	Level             zapcore.Level
	Caller            bool   // 是否开启caller位置
	EncoderOut        string // json plain
	EncoderLevel      string
	EncoderTime       string
	EncoderTimeLayout string

	// tar
	LogPath       string // 保存的日志文件
	LogMaxSize    int    // 文件大小限制
	LogMaxBackups int    //最大保留日志文件数量
	LogMaxAge     int    //日志文件保留天数
	LogCompress   bool   //是否压缩处理
}

type option func(*LoggerOptions)

func NewLoggerOption() *LoggerOptions {
	return &LoggerOptions{
		Level:             zapcore.InfoLevel,
		DefaultName:       "default.txt",
		HttpMetrices:      false,
		CtxKey:            "logger",
		HttpPort:          5000,
		Color:             true,
		Console:           true,
		Switch:            false, // 默认不开启，因为会占用端口
		SwitchTime:        5 * time.Minute,
		Caller:            false,
		EncoderOut:        "json",
		EncoderLevel:      "level",
		EncoderTime:       "time",
		EncoderTimeLayout: "2006-01-02 15:04:05",
		LogPath:           "",
		LogMaxSize:        1,
		LogMaxBackups:     5,
		LogMaxAge:         1,
		LogCompress:       false,
	}
}

func WithName(name string) option {
	return func(lo *LoggerOptions) {
		lo.Name = name
	}
}

func WithDefaultLogName(name string) option {
	return func(lo *LoggerOptions) {
		lo.DefaultName = name
	}
}

func WithHTTPMetrices(enable bool) option {
	return func(lo *LoggerOptions) {
		lo.HttpMetrices = enable
	}
}

func WithCtxKey(key ctxKey) option {
	return func(lo *LoggerOptions) {
		lo.CtxKey = key
	}
}

func WithLevel(level zapcore.Level) option {
	return func(lo *LoggerOptions) {
		lo.Level = level
	}
}

func WithSwitch(isSwitch bool, switchTime time.Duration) option {
	return func(lo *LoggerOptions) {
		lo.Switch = isSwitch
		lo.SwitchTime = switchTime
	}
}

func WithLogPath(logPath string) option {
	return func(lo *LoggerOptions) {
		lo.LogPath = logPath
	}
}

func WithColor(color bool) option {
	return func(lo *LoggerOptions) {
		lo.Color = color
	}
}

func WithEncoderTime(timeKey string) option {
	return func(lo *LoggerOptions) {
		lo.EncoderTime = timeKey
	}
}

func WithEncoderTimeWithLayout(layout string) option {
	return func(lo *LoggerOptions) {
		lo.EncoderTimeLayout = layout
	}
}

func WithEncoderLevel(levelKey string) option {
	return func(lo *LoggerOptions) {
		lo.EncoderLevel = levelKey
	}
}

func WithEncoderOut(out string) option {
	return func(lo *LoggerOptions) {
		lo.EncoderOut = out
	}
}

func WithConsole(enable bool) option {
	return func(lo *LoggerOptions) {
		lo.Console = enable
	}
}

func WithCaller(enable bool) option {
	return func(lo *LoggerOptions) {
		lo.Caller = enable
	}
}
