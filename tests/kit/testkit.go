package kit

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"sort"
	"strconv"
	"strings"
	"testing"

	_ "github.com/beyondstorage/go-service-memory"
	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/beyondstorage/go-storage/v4/types"
	"github.com/stretchr/testify/assert"

	"github.com/beyondstorage/beyond-ftp/client"
	"github.com/beyondstorage/beyond-ftp/cmd"
	"github.com/beyondstorage/beyond-ftp/config"
	"github.com/beyondstorage/beyond-ftp/server"
	"github.com/beyondstorage/beyond-ftp/transfer"
	"github.com/beyondstorage/beyond-ftp/utils"
)

type mockPassiveHandler struct {
	conn utils.Conn
}

func (m *mockPassiveHandler) Open() (utils.Conn, error) {
	return m.conn, nil
}

func (m *mockPassiveHandler) Close() error {
	return m.conn.Close()
}

type mockActiveHandler struct {
	remoteAddr *net.TCPAddr

	conn utils.Conn
	cm   *connManager
}

func (m *mockActiveHandler) Open() (utils.Conn, error) {
	m.conn = m.cm.connect(m.remoteAddr.Port)
	return m.conn, nil
}

func (m *mockActiveHandler) Close() error {
	return m.conn.Close()
}

type Hook struct {
	OnWrite func()
	OnRead  func()
}

type MockServer struct {
	listener chan interface{}

	cm       *connManager
	setting  *config.ServerSettings
	storager types.Storager
}

func (m *MockServer) Storager() types.Storager {
	return m.storager
}

func (m *MockServer) Setting() *config.ServerSettings {
	return m.setting
}

func NewMockServer(listener chan interface{}, cm *connManager, setting *config.ServerSettings) (*MockServer, error) {
	storager, err := services.NewStoragerFromString(setting.Service)
	if err != nil {
		return nil, err
	}
	return &MockServer{listener: listener, cm: cm, setting: setting, storager: storager}, nil
}

func (m *MockServer) Start() {
}

func (m *MockServer) Stop() {
	close(m.listener)
}

func (m *MockServer) AcceptClient() (utils.Conn, string, error) {
	addr, ok := <-m.listener
	if !ok {
		return nil, "", errors.New("server stop")
	}
	conn, p := m.cm.new()
	m.listener <- p
	return conn, addr.(string), nil
}

func (m *MockServer) PassiveTransferFactory(listenHost string, portRange *config.PortRange) (transfer.Handler, int, error) {
	conn, i := m.cm.new()
	return &mockPassiveHandler{
		conn: conn,
	}, i, nil
}

func (m *MockServer) ActiveTransferFactory(addr *net.TCPAddr) transfer.Handler {
	return &mockActiveHandler{
		remoteAddr: addr,
	}
}

type Result struct {
	t *testing.T

	code int
	msg  string
}

func (r *Result) EqualCode(code int) {
	assert.Equal(r.t, code, r.code)
}

func (r *Result) EqualMsg(msg string) {
	assert.Equal(r.t, msg, r.msg)
}

var (
	DefaultServerSetting = &config.ServerSettings{
		Service:    "memory:///ftp",
		ListenHost: "127.0.0.1",
		ListenPort: 21,
		PublicHost: "127.0.0.1",
		DataPortRange: &config.PortRange{
			Start: 1204,
			End:   2048,
		},
		Users: map[string]string{"anonymous": ""},
	}
)

type TestKit struct {
	t *testing.T

	l  chan interface{}
	s  server.Server
	cm *connManager
}

func NewTestKit(t *testing.T) *TestKit {
	return NewTestKitWithConfig(t, DefaultServerSetting)
}

func NewTestKitWithConfig(t *testing.T, settings *config.ServerSettings) *TestKit {
	listener := make(chan interface{})
	cm := newConnManager()
	mockServer, err := NewMockServer(listener, cm, settings)
	utils.MustNil(err)
	go cmd.StartServer(mockServer)

	kit := &TestKit{
		t:  t,
		l:  listener,
		s:  mockServer,
		cm: cm,
	}
	return kit
}

func (k *TestKit) Dail() utils.Conn {
	k.l <- ""
	p := <-k.l
	conn := k.cm.connect(p.(int))
	response(bufio.NewReader(conn))
	return conn
}

func (k *TestKit) Stop() {
	k.s.Stop()
}

func (k *TestKit) SupportAppender() bool {
	_, ok := k.s.Storager().(types.Appender)
	return ok
}

func (k *TestKit) SupportDirer() bool {
	_, ok := k.s.Storager().(types.Direr)
	return ok
}

func (k TestKit) TransferConnReceive(r utils.Conn) []byte {
	bytes, err := ioutil.ReadAll(r)
	utils.MustNil(err)
	return bytes
}

func (k TestKit) TransferConnSend(conn utils.Conn, data []byte) {
	n, err := conn.Write(data)
	utils.MustNil(err)
	assert.Equal(k.t, len(data), n)
	utils.MustNil(conn.Close())
}

func (k *TestKit) PassiveConn(conn utils.Conn) utils.Conn {
	isNumber := func(b byte) bool {
		return byte('0') <= b && b <= byte('9')
	}
	port := k.Send(conn, client.PASV).Success().message()[0]
	addr := make([]int, 6)
	c := 0
	for i := 0; i < len(port); i++ {
		if isNumber(port[i]) {
			for i1 := i; i1 < len(port); i1++ {
				if isNumber(port[i1]) {
					addr[c] *= 10
					addr[c] += int(port[i1]) - int('0')
					continue
				}
				i = i1
				break
			}
			c++
		}
	}

	return k.cm.connect(addr[4]*256 + addr[5])
}

func (k *TestKit) ActiveConn(conn utils.Conn) utils.Conn {
	c, i := k.cm.new()
	k.Send(conn, fmt.Sprintf("PORT 127,0,0,1,%d,%d", i/256, i%256)).Auto().Success()
	return c
}

func (k *TestKit) MustSuccess(conn utils.Conn, cmd string) {
	k.sendWithExpectStates(conn, cmd, success, another)
}

func (k *TestKit) MustFailure(conn utils.Conn, cmd string) {
	k.sendWithExpectStates(conn, cmd, failure)
}

func (k *TestKit) MustError(conn utils.Conn, cmd string) {
	k.sendWithExpectStates(conn, cmd, erro)
}

func (k *TestKit) sendWithExpectStates(conn utils.Conn, cmd string, ss ...state) {
	k.Send(conn, cmd).Auto().Expect(ss...)
}

func (k *TestKit) AnonymousLogin() utils.Conn {
	conn := k.Dail()
	k.Send(conn, "user anonymous").Auto().Another()
	k.Send(conn, "pass").Auto().Success()
	return conn
}

func (k *TestKit) Size(conn utils.Conn, path string) int {
	size := k.Send(conn, fmt.Sprintf("SIZE %s", path)).Success().message()[0]
	fileSize, err := strconv.Atoi(size)
	utils.MustNil(err)
	return fileSize
}

func (k *TestKit) List(conn utils.Conn, path string) []string {
	passive := k.PassiveConn(conn)

	var data string

	k.Send(conn, fmt.Sprintf("LIST %s", path)).Expect(wait).TakeAction(func() {
		data = string(k.TransferConnReceive(passive))
	}).Success()
	data = data[:len(data)-2]
	fileList := strings.Split(data, "\r\n")
	sort.Strings(fileList)

	return fileList
}

func (k *TestKit) transferFile(conn utils.Conn, path string, data []byte, cmd string) {
	passive := k.PassiveConn(conn)
	k.Send(conn, fmt.Sprintf("%s %s", cmd, path)).Expect(wait).TakeAction(func() {
		k.TransferConnSend(passive, data)
	}).Success()
}

func (k *TestKit) Store(conn utils.Conn, path string, data []byte) {
	k.transferFile(conn, path, data, "STOR")
}

func (k *TestKit) Append(conn utils.Conn, path string, data []byte) {
	k.transferFile(conn, path, data, "APPE")
}

func (k *TestKit) Retrieve(conn utils.Conn, path string) []byte {
	passive := k.PassiveConn(conn)

	var data []byte
	k.Send(conn, fmt.Sprintf("RETR %s", path)).Expect(wait).TakeAction(func() {
		data = k.TransferConnReceive(passive)
	}).Success()

	return data
}

func (k *TestKit) Abort(conn utils.Conn) {
	k.Send(conn, "abot").Auto().Success()
}

func (k *TestKit) Send(conn utils.Conn, cmd string) *model {
	switch strings.ToUpper(strings.Split(cmd, " ")[0]) {
	case client.ABOR, client.ALLO, client.DELE, client.CWD, client.CDUP, client.SMNT, client.HELP,
		client.MODE, client.NOOP, client.PASV, client.QUIT, client.SITE, client.PORT, client.SYST,
		client.STAT, client.RMD, client.MKD, client.PWD, client.STRU, client.TYPE,
		client.MDTM, client.SIZE, client.FEAT:
		return replyModel(k.t, conn).Begin(cmd)
	case client.APPE, client.LIST, client.NLST, client.REIN, client.RETR, client.STOR, client.STOU:
		return waitReplyModel(k.t, conn).Begin(cmd)
	case client.USER, client.PASS:
		return loginModel(k.t, conn).Begin(cmd)
	case client.ACCT, client.RNTO:
		return acctOrRntoModel(k.t, conn).Begin(cmd)
	case client.RNFR, client.REST:
		return rnfrModel(k.t, conn).Begin(cmd)
	}
	panic(cmd + " is not support in testkit")
}

func send(writer *bufio.Writer, cmd string) {
	_, err := writer.WriteString(fmt.Sprintf("%s\n", cmd))
	utils.MustNil(err)
	err = writer.Flush()
	utils.MustNil(err)
}

func response(reader *bufio.Reader) (code, string) {
	respString, err := reader.ReadString('\n')
	utils.MustNil(err)

	if respString[3] != '-' {
		respString = respString[:len(respString)-2]
		resp := strings.SplitN(respString, " ", 2)
		c, err := strconv.Atoi(resp[0])
		utils.MustNil(err)
		msg := resp[1]
		return code(c), msg
	} else {
		respCode := respString[:3]
		msg := respString[4:]
		for {
			line, err := reader.ReadString('\n')
			utils.MustNil(err)
			if len(line) > 4 && line[3] == ' ' && line[:3] == respCode {
				c, err := strconv.Atoi(respCode)
				utils.MustNil(err)
				return code(c), msg
			}
			msg += line
		}
	}
}
