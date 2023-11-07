package logger

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	handler *http.Server

	handlerPort = 5000

	mux = http.NewServeMux()

	handlerOnce = sync.Once{}
)

var (
	gidMap          = sync.Map{} // map[string]context.Context{} // name-gid child-zapx
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

func getGoroutineID() int64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.ParseInt(idField, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

type ZapX struct {
	*zap.Logger

	opts   *LoggerOptions
	rawOpt []option
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

	DefaultLogger = NewBasicLogger()
}

// [info]
func NewMiniLogger(options ...option) *ZapX {
	opt := append([]option{
		WithName("mini"),
		WithLevel(loggerLevel),
		WithEncoderLevel(""),
		WithEncoderTime(""),
		WithEncoderOut("plain"),
	}, options...)
	return NewLogger(
		opt...,
	)
}

// [time] [info]
func NewBasicLogger(options ...option) *ZapX {
	opt := append([]option{
		WithName("default"),
		WithLevel(loggerLevel),
		WithEncoderLevel(""),
		WithEncoderTime("at"),
		WithEncoderTimeWithLayout("2006-01-02 15:04:05.000"),
		WithEncoderOut("plain"),
		WithSwitch(false, 1*time.Second),
	}, options...)

	return NewLogger(
		opt...,
	)
}

// null
func NewNullLogger(options ...option) *ZapX {
	opt := append([]option{
		WithName("null"),
		WithConsole(false),
	}, options...)
	return NewLogger(
		opt...,
	)
}

// 设置logger
func Set(name string, Logger *ZapX) {
	if name == "default" {
		DefaultLogger = Logger
	}
	loggerPool.Store(name, Logger)
}

func SetDefaultWithOpt(options ...option) {
	loggerPool.Delete("default")

	DefaultLogger = NewLogger(append(options, func(lo *LoggerOptions) { lo.Name = "default" })...)
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
//
//	WithLevel(zapcore.DebugLevel),
//	WithLogPath("./logs/info.log"),
//	WithName("info"),
//	WithCaller(false),
//	WithSwitchTime(2*time.Second),
//
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
	config.EncodeTime = zapcore.TimeEncoderOfLayout(opt.EncoderTimeLayout)

	var (
		basicEncoder zapcore.Encoder
	)
	if opt.EncoderOut == "json" {
		basicEncoder = zapcore.NewJSONEncoder(config)
	} else {
		basicEncoder = zapcore.NewConsoleEncoder(config)
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

		// StartHandler()
		// 注册级别更改监听
		switchWatch(opt.Name)
	} else {
		priority = getPriority(opt.Level)
	}

	if opt.Console {
		coreArr = append(coreArr, zapcore.NewCore(basicEncoder, zapcore.AddSync(os.Stdout), priority))
	}
	// 是否保存到文件中
	if opt.LogPath != "" {
		r := &lumberjack.Logger{
			Filename:   path.Join(opt.LogPath, opt.DefaultName), //日志文件存放目录，如果文件夹不存在会自动创建
			MaxSize:    opt.LogMaxSize,                          //文件大小限制,单位MB
			MaxBackups: opt.LogMaxBackups,                       //最大保留日志文件数量
			MaxAge:     opt.LogMaxAge,                           //日志文件保留天数
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

	logx := &ZapX{
		Logger: zap.New(zapcore.NewTee(coreArr...), zapOpts...),
		opts:   opt,
		rawOpt: options,
	}
	loggerPool.Store(opt.Name, logx)

	logx.starthandler(logx.opts)

	return logx
}

// 写进到logpath下的[label].txt
func (l *ZapX) WithLabel(label string) *ZapX {
	name := fmt.Sprintf("@%s@%s", l.opts.Name, label)
	val, ok := loggerPool.Load(name)
	if ok {
		return val.(*ZapX)
	}

	childOpt := append([]option{}, l.rawOpt...)
	childOpt = append(childOpt, WithName(name), WithDefaultLogName(label+".txt"))

	childlogger := NewLogger(childOpt...)
	loggerPool.Store(name, childlogger)

	return childlogger
}

func (l *ZapX) starthandler(opt *LoggerOptions) {
	if !opt.Switch {
		return
	}

	handlerOnce.Do(func() {
		if opt.Switch {
			enableSwitch(opt.HttpEngine)
		}

		// 启用内部handler
		port := opt.SwitchPort
		if port == 0 {
			port = handlerPort
		}
		if opt.InternalEngine {
			handler = &http.Server{
				Addr:         fmt.Sprintf(":%d", port),
				WriteTimeout: time.Second * 3,
				Handler:      mux,
			}

			go func() {
				if err := handler.ListenAndServe(); err != nil {
					DefaultLogger.Error(err.Error())
				}
			}()
		}
	})
}

// Trace 根据当前gorotuine id分辨ctx
func (l *ZapX) Trace(fields ...zap.Field) *ZapX {
	curID := fmt.Sprintf("%s-%d", l.opts.Name, getGoroutineID())
	if v, exist := gidMap.Load(curID); exist {
		return l.GetCtx(v.(context.Context))
	} else {
		fields = append(fields, zap.String("traceid", uuid.New().String()))
		ctx, l := l.AddCtx(context.TODO(), fields...)
		gidMap.Store(curID, ctx)
		return l
	}
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
