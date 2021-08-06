package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/beyondstorage/go-storage/v4/types"

	"github.com/beyondstorage/beyond-ftp/config"
	"github.com/beyondstorage/beyond-ftp/transfer"
	"github.com/beyondstorage/beyond-ftp/utils"
)

// Handler driver handles the file system access logic.
type Handler struct {
	id            string                 // id of the client
	conn          utils.Conn             // TCP connection
	writer        *bufio.Writer          // Writer on the TCP connection
	reader        *bufio.Reader          // Reader on the TCP connection
	storager      types.Storager         // The root storager
	user          string                 // Authenticated user
	loginUser     string                 // login in user name
	path          string                 // Current path
	command       string                 // Command received on the connection
	param         string                 // Param of the FTP command
	connectedAt   time.Time              // Date of connection
	remoteAddr    string                 // Remote address of the connection
	ctxRnfr       string                 // Rename from
	ctxRest       int64                  // Restart point
	transfer      transfer.Handler       // Transfer connection
	transferTLS   bool                   // Use TLS for transfer connection
	serverSetting *config.ServerSettings // serverSetting

	commandArrivedSignalCh chan *CommandDescription
	commandAbortCtx        context.Context
	commandAbortCancelFn   context.CancelFunc
	commandRunningWg       sync.WaitGroup

	passiveTransferFactory func(listenHost string, portRange *config.PortRange) (transfer.Handler, int, error)
	activeTransferFactory  func(*net.TCPAddr) transfer.Handler
}

// Path provides the current working directory of the client.
func (c *Handler) Path() string {
	return c.path
}

// SetPath changes the current working directory.
func (c *Handler) SetPath(path string) {
	c.path = path
}

// HandleCommands reads the stream of commands.
func (c *Handler) HandleCommands() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go c.handleCommand(ctx)
	defer func() {
		c.TransferClose()
		cancelFunc()
	}()
	for {
		line, err := c.reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				utils.Logger.Debugf("TCP disconnect: ftp.disconnect, ID: %s, Clean: %t", c.id, false)
			} else {
				utils.Logger.Errorf("Read error: ftp.read_error, ID: %s, Error: %v", c.id, err)
			}
			return
		}

		utils.Logger.Debugf("FTP RECV: ftp.cmd_recv, ID: %s, Line: %v", c.id, line)

		command, param := utils.ParseLine(line)
		command = strings.ToUpper(command)

		cmdDesc, ok := commandsMap[command]
		if !ok {
			c.WriteMessage(StatusSyntaxErrorNotRecognised, "Unknown command")
			continue
		}

		if cmdDesc == nil {
			c.WriteMessage(StatusCommandNotImplemented, command+" command not supported")
			continue
		}

		if c.loginUser == "" && !cmdDesc.Open {
			c.WriteMessage(StatusNotLoggedIn, "Please login with USER and PASS")
			continue
		}

		switch command {
		case ABOR:
			c.handleABOR()
		case QUIT:
			c.commandRunningWg.Wait()
			c.handleQUIT()
			return
		default:
			c.commandRunningWg.Wait()
			c.commandRunningWg.Add(1)
			c.commandAbortCtx, c.commandAbortCancelFn = context.WithCancel(context.Background())
			c.command = command
			c.param = param
			c.commandArrivedSignalCh <- cmdDesc
		}
	}
}

// TransferOpen opens transfer with handler
func (c *Handler) TransferOpen() (utils.Conn, error) {
	if c.transfer == nil {
		return nil, errors.New("no connection declared")
	}
	c.WriteMessage(StatusFileStatusOK, "Using transfer connection")
	conn, err := c.transfer.Open()
	if err == nil {
		utils.Logger.Debugf("FTP Transfer connection opened: ftp.transfer_open, ID: %s", c.id)
	} else {
		utils.Logger.Errorf("FTP Transfer connection open failed: %v: ", err)
	}

	return conn, err
}

// TransferClose closes transfer with handler
func (c *Handler) TransferClose() {
	if c.transfer != nil {
		c.transfer.Close()
		c.transfer = nil
		utils.Logger.Debugf("FTP Transfer connection closed: ftp.transfer_close. ID: %s", c.id)
	}
}

// handleCommand takes care of executing the received line.
func (c *Handler) handleCommand(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.Errorf("Internel error: %v, Trace: %s", r, debug.Stack())
			c.WriteMessage(StatusSyntaxErrorNotRecognised, fmt.Sprintf("Internal error: %s", r))
		}
	}()

	for {
		select {
		case cmdDesc := <-c.commandArrivedSignalCh:
			cmdDesc.Fn(c)
			c.commandRunningWg.Done()
		case <-ctx.Done():
			return
		}
	}
}

// WriteMessage writes server response
func (c *Handler) WriteMessage(code int, message string) {
	c.writeLine(fmt.Sprintf("%d %s", code, message))
}

func (c *Handler) disconnect() {
	if c.transfer != nil {
		c.transfer.Close()
		c.transfer = nil
	}
	c.conn.Close()
}

func (c *Handler) writeLine(line string) {
	utils.Logger.Debugf("FTP SEND: ftp.cmd_send, ID: %s, Line: %s", c.id, line)
	c.writer.Write([]byte(line))
	c.writer.Write([]byte("\r\n"))
	c.writer.Flush()
}

// NewHandler initializes a client handler when someone connects.
func NewHandler(id, remoteAddr string, connection utils.Conn, settings *config.ServerSettings,
	storager types.Storager,
	passive func(string, *config.PortRange) (transfer.Handler, int, error),
	active func(*net.TCPAddr) transfer.Handler,
) *Handler {
	p := &Handler{
		id:                     id,
		conn:                   connection,
		writer:                 bufio.NewWriter(connection),
		reader:                 bufio.NewReader(connection),
		storager:               storager,
		connectedAt:            time.Now().UTC(),
		remoteAddr:             remoteAddr,
		path:                   "/",
		serverSetting:          settings,
		commandArrivedSignalCh: make(chan *CommandDescription),
		commandRunningWg:       sync.WaitGroup{},
		passiveTransferFactory: passive,
		activeTransferFactory:  active,
	}

	return p
}
