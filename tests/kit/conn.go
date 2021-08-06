package kit

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/beyondstorage/beyond-ftp/utils"
)

type connManager struct {
	mu   sync.Mutex
	pool map[int]*mockConn
	port int

	hooks []*Hook
}

func newConnManager() *connManager {
	return &connManager{
		mu:   sync.Mutex{},
		pool: make(map[int]*mockConn),
		port: 1024,
	}
}

func (m *connManager) new() (utils.Conn, int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.port++
	ctx, cancelFunc := context.WithCancel(context.Background())
	conn := newMockConn(make(chan byte), make(chan byte), ctx, cancelFunc, m.hooks...)
	m.pool[m.port] = conn

	return conn, m.port
}

func (m *connManager) connect(port int) utils.Conn {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn := m.pool[port]
	return newMockConn(conn.r, conn.w, conn.ctx, conn.cancelF, conn.hooks...)
}

func SetConnHooks(conn utils.Conn, hooks ...*Hook) {
	for i := range hooks {
		if hooks[i].OnRead == nil {
			hooks[i].OnRead = func() {}
		}

		if hooks[i].OnWrite == nil {
			hooks[i].OnWrite = func() {}
		}
	}
	conn.(*mockConn).SetHooks(hooks...)
}

type mockConn struct {
	w chan byte
	r chan byte

	hooks []*Hook

	ctx     context.Context
	cancelF context.CancelFunc
}

func newMockConn(w, r chan byte, ctx context.Context, cancelF context.CancelFunc, hooks ...*Hook) *mockConn {
	for i := range hooks {
		if hooks[i].OnRead == nil {
			hooks[i].OnRead = func() {}
		}

		if hooks[i].OnWrite == nil {
			hooks[i].OnWrite = func() {}
		}
	}
	return &mockConn{
		w:       w,
		r:       r,
		hooks:   hooks,
		ctx:     ctx,
		cancelF: cancelF,
	}
}

func (m *mockConn) Write(p []byte) (n int, err error) {
	for i, b := range p {
		for _, h := range m.hooks {
			h.OnWrite()
		}
		select {
		case m.w <- b:
		case <-m.ctx.Done():
			return i, errors.New("connection is closed")
		}
	}

	return len(p), nil
}

func (m *mockConn) Read(p []byte) (n int, err error) {
	for _, h := range m.hooks {
		h.OnRead()
	}
	var ok bool
	select {
	case p[0], ok = <-m.r:
		if !ok {
			return 0, io.EOF
		}
	case <-m.ctx.Done():
		return 0, io.EOF
	}

	return 1, nil
}

func (m *mockConn) Close() error {
	m.cancelF()
	close(m.w)
	return nil
}

func (m *mockConn) SetHooks(hooks ...*Hook) {
	m.hooks = hooks
}
