package server

import (
	"net"

	"github.com/beyondstorage/go-storage/v4/types"

	"github.com/beyondstorage/beyond-ftp/config"
	"github.com/beyondstorage/beyond-ftp/transfer"
	"github.com/beyondstorage/beyond-ftp/utils"
)

type Server interface {
	// Start starts a server.
	Start()
	// Stop stops the server and release the resource.
	Stop()
	// AcceptClient return the connection and id when new client is arrived.
	AcceptClient() (utils.Conn, string, error)
	// PassiveTransferFactory return a passive transfer handler
	PassiveTransferFactory(listenHost string, portRange *config.PortRange) (transfer.Handler, int, error)
	// ActiveTransferFactory return a active transfer handler
	ActiveTransferFactory(addr *net.TCPAddr) transfer.Handler
	// Setting return the server setting
	Setting() *config.ServerSettings
	// Storager return the root storager of the server
	Storager() types.Storager
}
