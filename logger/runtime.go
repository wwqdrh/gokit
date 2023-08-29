package logger

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

var (
	keyFn     = sync.Map{} // string, func(zapcore.Level)
	switchAPI = "/log/setlevel"
)

type switchFn func(zapcore.Level)

func enableSwitch(register func(string, http.HandlerFunc)) {
	// 设置路由规则
	register(switchAPI, func(w http.ResponseWriter, r *http.Request) {
		key, level := r.URL.Query().Get("key"), r.URL.Query().Get("level")

		fnI, ok := keyFn.Load(key)
		if !ok {
			w.WriteHeader(400)
			if _, err := w.Write([]byte("未注册")); err != nil {
				DefaultLogger.Error(err.Error())
			}
			return
		}
		fn, ok := fnI.(switchFn)
		if !ok {
			w.WriteHeader(500)
			if _, err := w.Write([]byte("未知错误")); err != nil {
				DefaultLogger.Error(err.Error())
			}
			return
		}
		if lev, ok := loggerLevelMap[level]; ok {
			fn(lev)
		}
	})
}

// Sync 保证日志刷入到磁盘不会丢失
func Sync() {
	loggerPool.Range(func(key, value interface{}) bool {
		l, ok := value.(*ZapX)
		if !ok {
			return true
		}

		if err := l.Core().Sync(); err != nil {
			fmt.Println(err.Error())
		}

		return true
	})
}

// 监听key的变化修改日志级别
func switchWatch(name string) {
	swit := switchLevel(name)
	keyFn.Store(name, swit)
}

// 动态切换logger的日志级别
func switchLevel(name string) switchFn {
	val, ok := atomicLevelPool.Load(name)
	if !ok {
		return nil
	}
	swit := val.(*switchContext)

	queue := make(chan zapcore.Level, 100)
	go func(l zapcore.Level) {
		timer := time.NewTicker(swit.duration)
		defer timer.Stop()

		for {
		next:
			<-timer.C

			// 通道中会有多个。使用最新的进行设置
			latestLvl := <-queue
			for {
				select {
				case latestLvl = <-queue:
				default:
					swit.l.SetLevel(latestLvl)
					goto next
				}
			}
		}
	}(swit.l.Level())

	return func(level zapcore.Level) {
		queue <- swit.l.Level() // 历史level
		swit.l.SetLevel(level)
	}
}
