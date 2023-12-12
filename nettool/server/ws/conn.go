package ws

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 解决跨域问题
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Conn interface {
	Read(context.Context) ([]byte, error)
	Write(ctx context.Context, msg []byte) error
	ReadMsg() ([]byte, error)
	WriteMsg(msg []byte) error
	LocalAddr() net.Addr
	ID() int32
	Close()
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	sync.Mutex
	closeFlag bool
	connid    int32

	Hub *WSHub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	recv chan []byte
}

func NewWSClient(rw http.ResponseWriter, req *http.Request, header http.Header) (*Client, error) {
	conn, err := upgrader.Upgrade(rw, req, header)
	if err != nil {
		return nil, err
	}
	client := &Client{conn: conn, send: make(chan []byte, 256), recv: make(chan []byte, 256)}
	if client.Hub != nil {
		client.Hub.register <- client
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
	return client, nil
}

func NewWSClientWithConn(conn *websocket.Conn) (*Client, error) {
	client := &Client{conn: conn, send: make(chan []byte, 256), recv: make(chan []byte, 256)}
	if client.Hub != nil {
		client.Hub.register <- client
	}
	go client.writePump()
	go client.readPump()
	return client, nil
}

func (c *Client) WithConnID(id int32) *Client {
	c.connid = id
	return c
}
func (c *Client) ID() int32 {
	return c.connid
}

func (c *Client) GetConn() *websocket.Conn {
	return c.conn
}

func (c *Client) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// 加入一个hub，在这里发送消息，其他的所有的连接也都会收到
func (c *Client) Join(key string) error {
	var hub *WSHub
	if val, ok := hubPool.Load(key); !ok {
		hub = NewHub()
		go hub.Run()
		hubPool.Store(key, hub)
	} else {
		var ok bool
		hub, ok = val.(*WSHub)
		if !ok {
			return errors.New("type is err")
		}
	}

	c.Hub = hub
	hub.register <- c
	return nil
}

func (c *Client) Read(ctx context.Context) ([]byte, error) {
	select {
	case msg := <-c.recv:
		return msg, nil
	case <-ctx.Done():
		return nil, errors.New("no value")
	}
}

func (c *Client) ReadMsg() ([]byte, error) {
	msg := <-c.recv
	return msg, nil
}

func (c *Client) Write(ctx context.Context, msg []byte) error {
	c.Lock()
	defer c.Unlock()

	if c.closeFlag {
		return errors.New("socket closeFlag is true")
	}
	select {
	case c.send <- msg:
		return nil
	case <-ctx.Done():
		return errors.New("time out")
	}
}

func (c *Client) WriteMsg(msg []byte) error {
	c.Lock()
	defer c.Unlock()

	if c.closeFlag {
		return errors.New("socket closeFlag is true")
	}
	c.send <- msg
	return nil
}

func (c *Client) Close() {
	c.Lock()
	defer c.Unlock()
	if c.closeFlag {
		return
	}
	close(c.send)
	c.closeFlag = true
}

func (c *Client) Broadcast(ctx context.Context, msg []byte) error {
	if c.Hub != nil {
		c.Hub.broadcast <- msg
		return nil
	} else {
		return errors.New("the hub is nil")
	}
}

func (c *Client) readPump() {
	defer func() {
		if c.Hub != nil {
			c.Hub.unregister <- c
		}
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.recv <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.Lock()
		c.closeFlag = true
		c.Unlock()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
