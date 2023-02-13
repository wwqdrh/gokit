package logger

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hpcloud/tail"
)

var (
	tailHandler = map[string]*tailInfo{} // 文件名与channel的映射
)

type tailInfo struct {
	cmd *tail.Tail // 获取最新的日志数据
	chs []connNode // 多个连接进行复用
}

type connNode struct {
	ch   chan string
	done chan bool
}

func EnableMetrices(register func(string, http.HandlerFunc)) {
	register("/api/log/list", HandlerLogList)
	register("/api/log/label/list", HandlerLogLabelList)
	register("/api/log/tail", HandlerLogTail)
}

// 获取有哪些日志类型
func HandlerLogList(w http.ResponseWriter, r *http.Request) {
	names := []string{}
	loggerPool.Range(func(key, value any) bool {
		n := key.(string)
		if !strings.HasPrefix(n, "@") && n != "" {
			names = append(names, key.(string))
		}
		return true
	})

	if body, err := json.Marshal(names); err != nil {
		w.WriteHeader(500)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			DefaultLogger.Error(err.Error())
		}
	} else {
		w.WriteHeader(200)
		w.Header().Add("ContentType", "application/json")
		if _, err := w.Write(body); err != nil {
			DefaultLogger.Warn(err.Error())
		}
	}
}

// 获取日志有哪些label类型
// arg:
// 	1、name
func HandlerLogLabelList(w http.ResponseWriter, r *http.Request) {
	querys := r.URL.Query()
	name := querys.Get("name")
	if name == "" {
		w.WriteHeader(400)
		if _, err := w.Write([]byte("name为空")); err != nil {
			DefaultLogger.Error(err.Error())
		}
		return
	}

	l, ok := loggerPool.Load(name)
	if !ok {
		w.WriteHeader(400)
		if _, err := w.Write([]byte("该logger不存在")); err != nil {
			DefaultLogger.Error(err.Error())
		}
		return
	}

	logpath := l.(*ZapX).opts.LogPath
	DefaultLogger.Debugx("listdir %s", nil, logpath)
	if items, err := os.ReadDir(logpath); err != nil {
		w.WriteHeader(500)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			DefaultLogger.Error(err.Error())
		}
	} else {
		labels := []string{}
		for _, item := range items {
			if !item.IsDir() && strings.HasSuffix(item.Name(), ".txt") {
				labels = append(labels, strings.TrimSuffix(item.Name(), ".txt"))
			}
		}
		if body, err := json.Marshal(labels); err != nil {
			w.WriteHeader(500)
			if _, err := w.Write([]byte(err.Error())); err != nil {
				DefaultLogger.Error(err.Error())
			}
		} else {
			w.WriteHeader(200)
			w.Header().Add("ContentType", "application/json")
			if _, err := w.Write(body); err != nil {
				DefaultLogger.Warn(err.Error())
			}
		}
	}
}

// 获取具体的日志数据
// arg:
//  1、name
//  2、label
func HandlerLogTail(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		w.WriteHeader(500)
		if _, err := w.Write([]byte("upgrade error")); err != nil {
			DefaultLogger.Error(err.Error())
		}
		return
	}

	querys := r.URL.Query()
	name, label := querys.Get("name"), querys.Get("label")
	if name == "" {
		w.WriteHeader(400)
		if _, err := w.Write([]byte("name为空")); err != nil {
			DefaultLogger.Error(err.Error())
		}
		return
	}

	if label == "" {
		label = "default.txt"
	}

	// check log is exist?
	file := path.Join(name, label+".txt")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if err := ws.WriteMessage(websocket.TextMessage, []byte("log error: 日志文件不存在")); err != nil {
			DefaultLogger.Error(err.Error())
			return
		}

		if e := ws.Close(); e != nil {
			DefaultLogger.Error(e.Error())
		}
		return
	}

	close := make(chan bool)
	go WsRead(ws, close)
	go WsWrite(ws, tailLog(file, close), close)
}

func tailLog(fileName string, done chan bool) chan string {
	cur := make(chan string, 1000)
	if val, ok := tailHandler[fileName]; ok {
		val.chs = append(val.chs, connNode{
			ch:   cur,
			done: done,
		})
		return cur
	}

	config := tail.Config{
		ReOpen:    true,                                 // 重新打开
		Follow:    true,                                 // 是否跟随
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件的哪个地方开始读
		MustExist: false,                                // 文件不存在不报错
		Poll:      true,
	}
	tails, err := tail.TailFile(fileName, config)
	if err != nil {
		DefaultLogger.Error("tail file failed, err:" + err.Error())
		return nil
	}
	handler := &tailInfo{
		cmd: tails,
		chs: []connNode{
			{
				ch:   cur,
				done: done,
			},
		},
	}
	tailHandler[fileName] = handler
	go func() {
		var (
			line *tail.Line
			ok   bool
		)
		for {
			line, ok = <-tails.Lines
			if !ok {
				DefaultLogger.Info("tail file close reopen, filename:" + tails.Filename)
				time.Sleep(time.Second)
				continue
			}

			// 为所有的channel发送
			chs := handler.chs[:0]
			for _, item := range handler.chs {
				select {
				case <-item.done:
					continue
				case item.ch <- line.Text:
					chs = append(chs, item)
				default:
				}
			}
			handler.chs = chs // 剔除已经关闭的
		}
	}()
	return cur
}
