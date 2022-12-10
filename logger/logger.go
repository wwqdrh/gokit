package logger

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	loggerPool      = sync.Map{}
	atomicLevelPool = sync.Map{} // string=>*switchContext
	loggerLevel     = zap.InfoLevel
	loggerLevelMap  = map[string]zapcore.Level{
		"debug":  zap.DebugLevel,
		"info":   zap.InfoLevel,
		"warn":   zap.WarnLevel,
		"error":  zap.ErrorLevel,
		"dpanic": zap.DPanicLevel,
		"panic":  zap.PanicLevel,
		"fatal":  zap.FatalLevel,
	}
)

type switchContext struct {
	l        *zap.AtomicLevel
	duration time.Duration
}

var (
	DefaultLogger *ZapX
)

type ZapX struct {
	*zap.Logger

	opts *LoggerOptions
}

func NewZapX(l *zap.Logger) *ZapX {
	return &ZapX{
		Logger: l,
		opts:   NewLoggerOption(),
	}
}

// 从环境变量获取日志级别
func init() {
	logLevel := os.Getenv("LOGLEVEL")
	if val, ok := loggerLevelMap[logLevel]; !ok {
		loggerLevel = zapcore.InfoLevel // 默认info级别
	} else {
		loggerLevel = val
	}

	DefaultLogger = NewLogger(
		WithName("default"),
		WithLevel(loggerLevel),
		WithEncoderLevel(""),
		WithEncoderTime(""),
		WithEncoderOut("plain"),
	)
}

// 设置logger
func Set(name string, Logger *ZapX) {
	loggerPool.Store(name, Logger)
}

// 获取logger
func Get(name string) *ZapX {
	val, ok := loggerPool.Load(name)
	if !ok {
		return DefaultLogger
	}

	if v, ok := val.(*ZapX); !ok {
		return DefaultLogger
	} else {
		return v
	}
}

// 输出到日志中的不加颜色
// 控制台中的根据color属性判断
// l := NewLogger(
// 	WithLevel(zapcore.DebugLevel),
// 	WithLogPath("./logs/info.log"),
// 	WithName("info"),
// 	WithCaller(false),
// 	WithSwitchTime(2*time.Second),
// )
func NewLogger(options ...option) *ZapX {
	opt := NewLoggerOption()
	for _, item := range options {
		item(opt)
	}

	if val, ok := loggerPool.Load(opt.Name); ok {
		if l, ok := val.(*ZapX); ok {
			return l
		}
	}

	// encoder
	config := encoderConfig
	config.LevelKey = opt.EncoderLevel
	config.TimeKey = opt.EncoderTime

	var (
		basicEncoder zapcore.Encoder
		colorEncoder zapcore.Encoder
	)
	if opt.EncoderOut == "json" {
		basicEncoder = zapcore.NewJSONEncoder(config)
		colorEncoder = NewColorJsonEncoder(config)
	} else {
		basicEncoder = zapcore.NewConsoleEncoder(config)
		colorEncoder = NewColorConsoleEncoder(config)
	}

	// 构造zap
	var coreArr []zapcore.Core
	// 获取level
	var priority zap.LevelEnablerFunc
	if opt.Switch {
		atomicLevel := getAtomicPriority(opt.Level)
		atomicLevelPool.Store(opt.Name, &switchContext{
			l:        &atomicLevel,
			duration: opt.SwitchTime,
		})
		priority = atomicLevel.Enabled

		StartHandler()
		// 注册级别更改监听
		switchWatch(opt.Name)
	} else {
		priority = getPriority(opt.Level)
	}

	// 控制台输出是否添加颜色
	if opt.Console {
		if opt.Color {
			coreArr = append(coreArr, zapcore.NewCore(colorEncoder, zapcore.AddSync(os.Stdout), priority))
		} else {
			coreArr = append(coreArr, zapcore.NewCore(basicEncoder, zapcore.AddSync(os.Stdout), priority))
		}
	}
	// 是否保存到文件中
	if opt.LogPath != "" {
		r := &lumberjack.Logger{
			Filename:   opt.LogPath,       //日志文件存放目录，如果文件夹不存在会自动创建
			MaxSize:    opt.LogMaxSize,    //文件大小限制,单位MB
			MaxBackups: opt.LogMaxBackups, //最大保留日志文件数量
			MaxAge:     opt.LogMaxAge,     //日志文件保留天数
			LocalTime:  true,
			Compress:   opt.LogCompress, //是否压缩处理
		}
		fileWriteSyncer := zapcore.AddSync(r)
		coreArr = append(coreArr, zapcore.NewCore(basicEncoder, fileWriteSyncer, priority))
	}

	zapOpts := []zap.Option{}
	if opt.Caller {
		zapOpts = append(zapOpts, zap.AddCaller())
	}
	l := zap.New(zapcore.NewTee(coreArr...), zapOpts...)
	if opt.Name != "" {
		loggerPool.Store(opt.Name, l)
	}

	logx := &ZapX{
		Logger: l,
		opts:   opt,
	}
	loggerPool.Store(opt.Name, logx)
	return logx
}

func (l *ZapX) GetCtx(ctx context.Context) *ZapX {
	log, ok := ctx.Value(l.opts.CtxKey).(*ZapX)
	if ok {
		return log
	}
	return l
}

func (l *ZapX) AddCtx(ctx context.Context, field ...zap.Field) (context.Context, *ZapX) {
	log := &ZapX{Logger: l.With(field...), opts: l.opts}
	ctx = context.WithValue(ctx, l.opts.CtxKey, log)
	return ctx, log
}

func (l *ZapX) WithContext(ctx context.Context) *ZapX {
	log, ok := ctx.Value(l.opts.CtxKey).(*ZapX)
	if ok {
		return log
	}
	return l
}

func (l *ZapX) Debugx(format string, fields []zap.Field, value ...interface{}) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Debug(fmt.Sprintf(format, value...), fields...)
}

func (l *ZapX) Infox(format string, fields []zap.Field, value ...interface{}) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Info(fmt.Sprintf(format, value...), fields...)
}

func (l *ZapX) Warnx(format string, fields []zap.Field, value ...interface{}) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Warn(fmt.Sprintf(format, value...), fields...)
}

func (l *ZapX) Errorx(format string, fields []zap.Field, value ...interface{}) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Error(fmt.Sprintf(format, value...), fields...)
}

func (l *ZapX) DPanicx(format string, fields []zap.Field, value ...interface{}) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).DPanic(fmt.Sprintf(format, value...), fields...)
}

func (l *ZapX) Panicx(format string, fields []zap.Field, value ...interface{}) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Panic(fmt.Sprintf(format, value...), fields...)
}

func (l *ZapX) Fatalx(format string, fields []zap.Field, value ...interface{}) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Fatal(fmt.Sprintf(format, value...), fields...)
}
