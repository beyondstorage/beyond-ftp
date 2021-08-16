package transfer

import (
	"fmt"
	"net"
	"time"

	"github.com/beyondstorage/beyond-ftp/utils"
)

// ActiveHandler handles active connection.
type ActiveHandler struct {
	RemoteAddr *net.TCPAddr // remote address of the client

	conn net.Conn
}

// Open opens connection.
func (a *ActiveHandler) Open() (utils.Conn, error) {
	conn, err := net.DialTimeout("tcp", a.RemoteAddr.String(), 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("could not establish active connection: %v", err)
	}

	// Keep connection as it will be closed by Close().
	a.conn = conn

	return a.conn, nil
}

// Close closes only if connection is established.
func (a *ActiveHandler) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}
