package logger

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestSwitchWatch(t *testing.T) {
	l := NewLogger(
		WithLevel(zapcore.WarnLevel),
		WithConsole(true),
		WithName("info"),
		WithCaller(true),
		WithConsole(false),
		WithSwitch(true, 2*time.Second),
	)
	l.Info("this is a info message")
	l.Error("this is error")
	fmt.Println("修改等级====")

	req := httptest.NewRequest("GET", switchAPI+"?"+url.Values{"key": []string{"info"}, "level": []string{"info"}}.Encode(), nil)
	res := httptest.NewRecorder()
	require.NotNil(t, handler)
	handler.Handler.ServeHTTP(res, req)
	fmt.Println(res.Result().StatusCode)
	require.Equal(t, 200, res.Result().StatusCode)
	fmt.Println("修改完成===")
	l.Info("this is a info message")
	l.Error("this is error")

	time.Sleep(3 * time.Second)
	fmt.Println("等待自动恢复===")
	l.Info("this is a info message")
	l.Error("this is error")
}
