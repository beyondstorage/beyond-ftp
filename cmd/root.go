package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/pengsrc/go-shared/check"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"

	"github.com/beyondstorage/beyond-ftp/client"
	"github.com/beyondstorage/beyond-ftp/config"
	"github.com/beyondstorage/beyond-ftp/constants"
	"github.com/beyondstorage/beyond-ftp/pprof"
	"github.com/beyondstorage/beyond-ftp/server"
	"github.com/beyondstorage/beyond-ftp/utils"
)

var (
	cfgFileFlag  string
	cfgDebugFlag bool

	clientCount int32
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   constants.Name,
	Short: "A FTP server that persists all data to Beyond Storage.",
	Long:  "A FTP server that persists all data to Beyond Storage.",
	Run: func(cmd *cobra.Command, args []string) {
		if cfgDebugFlag {
			pprof.StartPP()
		}
		c := config.LoadConfigFromFilepath(cfgFileFlag)
		s, err := server.NewFTPServer(c)
		check.ErrorForExit("server init error", err)
		StartServer(s)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Beyond-FTP version",
	Long:  "Print Beyond-FTP version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("BeyondFTP version %s", constants.Version)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		check.ErrorForExit(constants.Name, err)
	}
}

func init() {
	RootCmd.AddCommand(versionCmd)

	RootCmd.PersistentFlags().BoolVarP(&cfgDebugFlag, "debug", "d", false, "Enter debug mode")
	RootCmd.PersistentFlags().StringVarP(&cfgFileFlag, "config", "c", "./config/config.example.toml", "Specify config file")
}

func StartServer(s server.Server) {
	s.Start()
	go signalHandler(s)
	for {
		connection, addr, err := s.AcceptClient()
		if err != nil {
			utils.Logger.Errorf("Accept error: %v", err)
			return
		}

		id := strings.Replace(uuid.NewV4().String(), "-", "", -1)
		go serveClient(s, id, addr, connection)
	}
}

func serveClient(s server.Server, id, addr string, connection utils.Conn) {
	c := client.NewHandler(
		id, addr, connection, s.Setting(), s.Storager(), s.PassiveTransferFactory, s.ActiveTransferFactory,
	)

	atomic.AddInt32(&clientCount, 1)
	utils.Logger.Infof("FTP Client connected: ftp.connected, id: %s, RemoteAddr: %v, Total: %d", id, addr, clientCount)
	c.WriteMessage(client.StatusServiceReady, "Welcome to BeyondFTP Server")
	utils.Logger.Debugf("Accept client on: id: %s, IP: %v", id, addr)

	c.HandleCommands()

	utils.Logger.Debugf("Goodbye: id: %s, IP: %v", id, addr)
	atomic.AddInt32(&clientCount, -1)
	utils.Logger.Infof("FTP Client disconnected: ftp.disconnected, id: %s, RemoteAddr: %v, Total: %d", id, addr, clientCount)
}

func signalHandler(s server.Server) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)
	for {
		switch <-ch {
		case syscall.SIGTERM:
			s.Stop()
			return
		}
	}
}
