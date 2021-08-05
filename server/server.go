// Package server provides all the tools to build your own FTP server: The core library and the driver.
package server

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	_ "github.com/beyondstorage/go-service-memory"
	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/beyondstorage/go-storage/v4/types"
	"github.com/pengsrc/go-shared/check"

	"github.com/beyondstorage/beyond-ftp/config"
	"github.com/beyondstorage/beyond-ftp/constants"
	"github.com/beyondstorage/beyond-ftp/transfer"
	"github.com/beyondstorage/beyond-ftp/utils"
)

// FTPServer is where everything is stored.
// We want to keep it as simple as possible.
type FTPServer struct {
	Listener  net.Listener // Listener used to receive files
	StartTime time.Time    // Time when the s was started

	setting  *config.ServerSettings
	storager types.Storager
}

func (s *FTPServer) Storager() types.Storager {
	return s.storager
}

func (s *FTPServer) Setting() *config.ServerSettings {
	return s.setting
}

func (s *FTPServer) AcceptClient() (utils.Conn, string, error) {
	conn, err := s.Listener.Accept()
	if err != nil {
		return nil, "", err
	}
	return conn, conn.RemoteAddr().String(), nil
}

func (s *FTPServer) Start() {
	var err error
	s.Listener, err = net.Listen("tcp", fmt.Sprintf(
		"%s:%d", s.setting.ListenHost, s.setting.ListenPort,
	))
	if err != nil {
		utils.Logger.Fatalf("Cannot listen: %v", err)
	}

	utils.Logger.Infof("Listening... %v", s.Listener.Addr())
	check.ErrorForExit(constants.Name, err)
}

func (s *FTPServer) PassiveTransferFactory(listenHost string, portRange *config.PortRange) (transfer.Handler, int, error) {
	var tcpListener *net.TCPListener
	var err error
	var localAddr *net.TCPAddr

	for start := portRange.Start; start < portRange.End; start++ {
		port := portRange.Start + rand.Intn(portRange.End-portRange.Start)
		localAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", listenHost, port))
		if err != nil {
			continue
		}

		tcpListener, err = net.ListenTCP("tcp", localAddr)
		if err == nil {
			break
		} else {
			continue
		}
	}

	if err != nil || tcpListener == nil {
		utils.Logger.Errorf("Could not listen: %v", err)
		return nil, 0, errors.New("cannot listen")
	}

	p := &transfer.PassiveHandler{
		TCPListener: tcpListener,
		Listener:    tcpListener,
	}

	return p, tcpListener.Addr().(*net.TCPAddr).Port, nil
}

func (s *FTPServer) ActiveTransferFactory(addr *net.TCPAddr) transfer.Handler {
	return &transfer.ActiveHandler{
		RemoteAddr: addr,
	}
}

// Stop closes the listener.
func (s *FTPServer) Stop() {
	if s.Listener != nil {
		l := s.Listener
		s.Listener = nil
		l.Close()
	}
}

// NewFTPServer creates a new FTPServer instance.
func NewFTPServer(c *config.Config) (*FTPServer, error) {
	setting := config.GetServerSetting(c)
	storager, err := services.NewStoragerFromString(c.Service)
	if err != nil {
		return nil, err
	}
	return &FTPServer{
		StartTime: time.Now().UTC(),
		setting:   setting,
		storager:  storager,
	}, nil
}
