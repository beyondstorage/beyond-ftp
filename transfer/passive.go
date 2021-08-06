package transfer

import (
	"net"
	"time"

	"github.com/beyondstorage/beyond-ftp/utils"
)

// PassiveHandler handles passive connection.
type PassiveHandler struct {
	TCPListener *net.TCPListener // TCP Listener (only keeping it to define a deadline during the accept)
	Listener    net.Listener     // TCP or SSL Listener
	connection  net.Conn         // TCP Connection established
}

// Open opens connection.
func (p *PassiveHandler) Open() (utils.Conn, error) {
	return p.ConnectionWait(time.Minute)
}

// Close only the client connection is not supported at that time.
func (p *PassiveHandler) Close() error {
	if p.TCPListener != nil {
		_ = p.TCPListener.Close()
	}
	if p.connection != nil {
		_ = p.connection.Close()
	}
	return nil
}

// ConnectionWait wait for connection time out
func (p *PassiveHandler) ConnectionWait(wait time.Duration) (net.Conn, error) {
	if p.connection == nil {
		err := p.TCPListener.SetDeadline(time.Now().Add(wait))
		if err != nil {
			return nil, err
		}
		p.connection, err = p.Listener.Accept()
		if err != nil {
			return nil, err
		}
	}

	return p.connection, nil
}
