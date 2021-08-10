package cmd

import (
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

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
	Version: constants.Version,
	Use:     constants.Name,
	Short:   "A FTP server that persists all data to Beyond Storage.",
	Long:    "A FTP server that persists all data to Beyond Storage.",
	Run: func(cmd *cobra.Command, args []string) {
		if cfgDebugFlag {
			pprof.StartPP()
		}
		c := config.LoadConfigFromFilepath(cfgFileFlag)
		s, err := server.NewFTPServer(c)
		utils.MustNil(err)
		utils.SetUpLog()
		defer zap.L().Sync()
		StartServer(s)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		utils.MustNil(err)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&cfgDebugFlag, "debug", "d", false, "Enter debug mode")
	RootCmd.PersistentFlags().StringVarP(&cfgFileFlag, "config", "c", "./config/config.example.toml", "Specify config file")
}

func StartServer(s server.Server) {
	s.Start()
	go signalHandler(s)
	for {
		connection, addr, err := s.AcceptClient()
		if err != nil {
			zap.L().Info("Server client error", zap.Error(err))
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

	count := atomic.AddInt32(&clientCount, 1)
	zap.L().Info("FTP Client connected",
		zap.String("id", id),
		zap.String("remote address", addr),
		zap.Int32("connection count", count),
	)
	c.WriteMessage(client.StatusServiceReady, "Welcome to BeyondFTP Server")
	c.HandleCommands()

	count = atomic.AddInt32(&clientCount, -1)
	zap.L().Info("FTP Client connected",
		zap.String("id", id),
		zap.String("remote address", addr),
		zap.Int32("connection count", count),
	)
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
