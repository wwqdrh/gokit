package ws

import (
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/wwqdrh/gokit/logger"
	"go.uber.org/zap"
)

type WSServer struct {
	sync.Mutex
	addr      string
	wg        sync.WaitGroup
	conns     map[*websocket.Conn]struct{}
	connid    int32
	ln        net.Listener
	onInvoker func(Conn) IConnInvoker

	close func()
}

type IConnInvoker interface {
	OnNew()
	OnClose()
}

func NewWSServer(addr string, invoker func(conn Conn) IConnInvoker) *WSServer {
	return &WSServer{
		addr:      addr,
		onInvoker: invoker,
		conns:     make(map[*websocket.Conn]struct{}),
	}
}

func (p *WSServer) Start() error {
	ln, err := net.Listen("tcp", p.addr)
	if err != nil {
		logger.DefaultLogger.Fatal("启动失败，端口被占用", zap.String("addr", p.addr))
		return err
	}

	p.ln = ln
	httpSvr := &http.Server{
		Addr:    p.addr,
		Handler: p,
	}

	p.close = func() {
		httpSvr.Close()
	}
	go func() {
		err := httpSvr.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			logger.DefaultLogger.Fatalx("WSServer Serve: %s", nil, err.Error())
			return
		}
	}()
	return nil
}

func (p *WSServer) Close() {
	p.close()
}

func (p *WSServer) ListenAddr() *net.TCPAddr {
	return p.ln.Addr().(*net.TCPAddr)
}

func (p *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.DefaultLogger.Errorx("ServerHttp upgrader.Upgrade", nil, err.Error())
		return
	}

	p.wg.Add(1)
	defer p.wg.Done()

	p.Lock()
	p.conns[conn] = struct{}{}
	p.Unlock()

	wsc, err := NewWSClientWithConn(conn)
	if err != nil {
		logger.DefaultLogger.Errorx("ServerHttp upgrader.Upgrade", nil, err.Error())
		return
	}
	wsc.WithConnID(p.NewConnID())

	var invoker IConnInvoker
	if p.onInvoker != nil {
		invoker = p.onInvoker(wsc)
	}
	if invoker != nil {
		invoker.OnNew()
	}
	wsc.Close()
	p.Lock()
	delete(p.conns, conn)
	p.Unlock()
	if invoker != nil {
		invoker.OnClose()
	}
}

func (p *WSServer) NewConnID() int32 {
	return atomic.AddInt32(&p.connid, 1)
}
