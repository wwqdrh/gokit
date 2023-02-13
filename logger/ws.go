package logger

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
	// space   = []byte{' '}
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// websocket read healper func
func WsRead(conn *websocket.Conn, done chan bool) {
	defer func() {
		conn.Close()
		select {
		case _, ok := <-done:
			if ok {
				close(done)
			}
		default:
			close(done)
		}
		fmt.Println("退出wsread")
	}()
	conn.SetReadLimit(maxMessageSize)
	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		DefaultLogger.Error(err.Error())
		return
	}
	conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(pongWait)) })
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func WsWrite(conn *websocket.Conn, send chan string, done chan bool) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		select {
		case _, ok := <-done:
			if ok {
				close(done)
			}
		default:
			close(done)
		}
		if err := conn.Close(); err != nil {
			DefaultLogger.Error(err.Error())
		}
		fmt.Println("退出wswrite")
	}()
	for {
		select {
		case message, ok := <-send:
			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				DefaultLogger.Error(err.Error())
				return
			}
			if !ok {
				// The hub closed the channel.
				if err := conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					DefaultLogger.Error(err.Error())
				}
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write([]byte(message)); err != nil {
				DefaultLogger.Error(err.Error())
				return
			}

			// Add queued chat messages to the current websocket message.
			n := len(send)
			for i := 0; i < n; i++ {
				if _, err := w.Write(newline); err != nil {
					DefaultLogger.Error(err.Error())
					continue
				}
				if _, err := w.Write([]byte(<-send)); err != nil {
					DefaultLogger.Error(err.Error())
					continue
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				DefaultLogger.Error(err.Error())
				return
			}
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}
