package logger

import (
	"context"
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

func TestExampleCustomLogger(t *testing.T) {
	// info级别 with name
	l := NewLogger(
		WithLevel(zapcore.DebugLevel),
		WithLogPath("./logs/info.log"),
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
