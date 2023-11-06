package logger

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TODO 如何输出到控制台了会检测成测试失败
func TestDefaultLogger(t *testing.T) {
	// Default Logger
	DefaultLogger.Info("this is a info message")
	DefaultLogger.Error("this is a error message")

	DefaultLogger.Infox("this is a %s message", nil, "infox")
	DefaultLogger.Errorx("this is a %s message", nil, "errorx")
}

func TestNullLogger(t *testing.T) {
	// l := NewNullLogger()
	// l.Info("this is a info message")
	// l.Error("this is a error message")

	// l.Infox("this is a %s message", nil, "infox")
	// l.Errorx("this is a %s message", nil, "errorx")

	l2 := NewMiniLogger()
	l2.Info("this is a info message")
	l2.Error("this is a error message")

	l2.Infox("this is a %s message", nil, "infox")
	l2.Errorx("this is a %s message", nil, "errorx")
}

func TestExampleCustomLogger(t *testing.T) {
	// info级别 with name
	l := NewLogger(
		WithLevel(zapcore.DebugLevel),
		WithLogPath("./logs/info"),
		WithName("info"),
		WithCaller(true),
		WithConsole(false),
		WithSwitch(true, 2*time.Second),
	)
	l.Debug("this is a debug message")
	l.Info("this is a info message")
	l.Error("this is error")
	Get("info").Debug("this is a debug message")
	Get("info").Info("this is a info message")

	l.WithLabel("app1").Info("this is app1 logger")

	Get("info_app1").Info("this is app1 logger by get")

	// switch debug level
	s := switchLevel("info")
	s(zapcore.WarnLevel)

	l.Debug("this is a debug message")
	l.Info("this is a info message")
	Get("info").Debug("this is a debug message")
	Get("info").Info("this is a info message")

	time.Sleep(3 * time.Second)

	l.Debug("this is a debug message")
	l.Info("this is a info message")
	Get("info").Debug("this is a debug message")
	Get("info").Info("this is a info message")
}

func TestLoggerRotate(t *testing.T) {
	l := NewLogger(
		WithLevel(zapcore.DebugLevel),
		WithLogPath("./logs/rotate.log"),
		WithName("info"),
		WithCaller(true),
		WithConsole(false),
		WithSwitch(true, 2*time.Second),
	)

	chCnt := 0
	logline := "this is a test log"
	for {
		l.Info(logline)
		chCnt += len(logline)
		if chCnt > 1024*1024 {
			break
		}
	}
}

func TestLoggerWithTraceID(t *testing.T) {
	FA := func(ctx context.Context) {
		l := DefaultLogger.GetCtx(ctx)
		l.Info("FA", zap.Any("a", "a"))
	}

	FC := func(ctx context.Context) {
		l := DefaultLogger.GetCtx(ctx)
		l.Info("FC", zap.Any("c", "c"))
	}

	FB := func(ctx context.Context) {
		l := DefaultLogger.GetCtx(ctx)
		l.Info("FB", zap.Any("b", "b"))
		FC(ctx)
	}

	// 可以在中间件内赋值
	ctx, l := DefaultLogger.AddCtx(context.Background(), zap.String("traceId", uuid.New().String()))
	l.Info("TestGetLogger", zap.Any("t", "t"))
	FA(ctx)
	FB(ctx)

	// 可以在中间件内赋值
	ctx, l = DefaultLogger.AddCtx(context.Background(), zap.String("traceId", uuid.New().String()))
	l.Info("TestGetLogger", zap.Any("t", "t"))
	FA(ctx)
	FB(ctx)
}

func TestTraceID(t *testing.T) {
	wait := &sync.WaitGroup{}
	wait.Add(2)

	go func() {
		defer func() {
			wait.Done()
		}()
		DefaultLogger.Trace().Info("fun", zap.String("name", "func1"))
		DefaultLogger.Trace().Info("arg", zap.String("name", "arg1"))
	}()

	go func() {
		defer func() {
			wait.Done()
		}()
		DefaultLogger.Trace().Info("fun", zap.String("name", "func1"))
		DefaultLogger.Trace().Info("arg", zap.String("name", "arg1"))
	}()

	wait.Wait()
}
