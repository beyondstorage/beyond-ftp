package tests

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/beyondstorage/beyond-ftp/pprof"
	"github.com/beyondstorage/beyond-ftp/tests/kit"
)

type ftpServerTestBase struct {
	suite.Suite
}

func (b *ftpServerTestBase) SetupSuite() {
	logger, err := zap.NewDevelopment()
	assert.Nil(b.T(), err)
	zap.ReplaceGlobals(logger)
	pprof.StartPP()
}

type ftpServerBaseCommandTest struct {
	ftpServerTestBase
}

type cmdTestCase struct {
	cmd     string
	msg     string
	success bool
}

func (t *ftpServerBaseCommandTest) TestMiscCommand() {
	tk := kit.NewTestKit(t.T())
	defer tk.Stop()

	conn := tk.AnonymousLogin()
	tk.Send(conn, "feat").Success()
	tk.Send(conn, "syst").Success()
	tk.Send(conn, "noop").Success()
}

func (t *ftpServerBaseCommandTest) TestMkdir() {
	tk := kit.NewTestKit(t.T())
	defer tk.Stop()

	tests := []cmdTestCase{
		{"mkd test", "", true},
		{"cwd test", "", true},
		{"mkd test", "", true},
		{"mkd test", "", false},
		{"pwd", `"/test" is the current directory`, true},
		{"cwd test", "", true},
		{"pwd", `"/test/test" is the current directory`, true},
		{"cwd ..", "", true},
		{"pwd", `"/test" is the current directory`, true},
		{"cwd /test/test", "", true},
		{"pwd", `"/test/test" is the current directory`, true},
		{"cwd /test/test1", "", false},
		{"mkd /test/test", "", false},
		{"mkd /test/test1", "", true},
		{"cwd /test/test1", "", true},
		{"pwd", `"/test/test1" is the current directory`, true},
		{"cdup", "", true},
		{"pwd", `"/test" is the current directory`, true},
	}

	conn := tk.AnonymousLogin()
	for _, t := range tests {
		if t.success {
			tk.Send(conn, t.cmd).Success(t.msg)
		} else {
			tk.Send(conn, t.cmd).Failure(t.msg)
		}
	}
}

func (t *ftpServerBaseCommandTest) TestCommandExecution() {
	tk := kit.NewTestKit(t.T())
	defer tk.Stop()

	conn := tk.AnonymousLogin()
	first := tk.Send(conn, "mkd test")
	second := tk.Send(conn, "mkd test1")

	monitor := tk.AnonymousLogin()

	// here, since the command will be executed one by one, it should be only one dir
	fileList := tk.List(monitor, "")
	assert.Equal(t.T(), 1, len(fileList))

	first.Success()
	second.Success()

	// now, the second mkd command has been executed, we will get 2 dirs
	fileList = tk.List(monitor, "")
	assert.Equal(t.T(), 2, len(fileList))

	// after quit, the connection is closed and cannot send any data
	tk.MustSuccess(monitor, "quit")
	_, err := monitor.Write([]byte(" "))
	assert.NotNil(t.T(), err)
}

func (t *ftpServerBaseCommandTest) TestUserCommand() {
	myConfig := *kit.DefaultServerSetting
	myConfig.Users = map[string]string{
		"test1": "test1",
	}
	tk := kit.NewTestKitWithConfig(t.T(), &myConfig)
	defer tk.Stop()

	conn := tk.Dail()
	tk.MustFailure(conn, "pass")
	tk.MustSuccess(conn, "user test")
	tk.MustFailure(conn, "pass test1")
	tk.MustSuccess(conn, "user anonymous")
	tk.MustFailure(conn, "pass test1")
	tk.MustSuccess(conn, "user test1")
	tk.MustSuccess(conn, "pass test1")
	tk.MustFailure(conn, "pass test1")
}

func (t *ftpServerBaseCommandTest) TestListFiles() {
	tk := kit.NewTestKit(t.T())
	defer tk.Stop()

	conn := tk.AnonymousLogin()
	tk.MustSuccess(conn, "mkd test")
	tk.MustSuccess(conn, "mkd test1")

	fileList := tk.List(conn, "")
	assert.Equal(t.T(), []string{
		"d--------- 1 ftp ftp            0  Jan  1 00:00 test",
		"d--------- 1 ftp ftp            0  Jan  1 00:00 test1",
	}, fileList)
}

func (t *ftpServerBaseCommandTest) TestDeleteFile() {
	tk := kit.NewTestKit(t.T())
	defer tk.Stop()

	conn := tk.AnonymousLogin()

	tk.MustSuccess(conn, "mkd test")
	tk.MustSuccess(conn, "mkd test1")
	fileList := tk.List(conn, "")
	assert.Equal(t.T(), []string{
		"d--------- 1 ftp ftp            0  Jan  1 00:00 test",
		"d--------- 1 ftp ftp            0  Jan  1 00:00 test1",
	}, fileList)
	tk.MustSuccess(conn, "RMD test1")
	fileList = tk.List(conn, "")
	assert.Equal(t.T(), []string{
		"d--------- 1 ftp ftp            0  Jan  1 00:00 test",
	}, fileList)
}

func (t *ftpServerBaseCommandTest) TestRenameFile() {
	tk := kit.NewTestKit(t.T())
	defer tk.Stop()

	conn := tk.AnonymousLogin()

	tk.Store(conn, "file1", []byte("file1 content"))
	fileList := tk.List(conn, "")
	assert.Equal(t.T(), []string{
		"-rwxrwxrwx 1 ftp ftp           13  Jan  1 00:00 file1",
	}, fileList)

	tk.MustSuccess(conn, "rnfr file1")
	tk.MustSuccess(conn, "rnto test")

	fileList = tk.List(conn, "")
	assert.Equal(t.T(), []string{
		"-rwxrwxrwx 1 ftp ftp           13  Jan  1 00:00 test",
	}, fileList)

	tk.MustFailure(conn, "rnto test1")
}

func (t *ftpServerBaseCommandTest) TestStoreFile() {
	tk := kit.NewTestKit(t.T())
	defer tk.Stop()

	conn := tk.AnonymousLogin()
	tk.MustSuccess(conn, "mkd test")
	tk.Store(conn, "file1", []byte("file1 content"))
	fileList := tk.List(conn, "")
	assert.Equal(t.T(), []string{
		"-rwxrwxrwx 1 ftp ftp           13  Jan  1 00:00 file1",
		"d--------- 1 ftp ftp            0  Jan  1 00:00 test",
	}, fileList)
	tk.Send(conn, "size file1").Success("13")

	file := tk.Retrieve(conn, "file1")
	assert.Equal(t.T(), []byte("file1 content"), file)

	size := 4 * 1024 * 1024
	content := make([]byte, size)
	rand.Read(content)
	path := "large-file"
	tk.Store(conn, path, content)
	file = tk.Retrieve(conn, path)
	assert.Equal(t.T(), content, file)
}

func (t *ftpServerBaseCommandTest) TestAbort() {
	h := &kit.Hook{
		OnWrite: func() {
			time.Sleep(time.Millisecond)
		},
	}

	tk := kit.NewTestKit(t.T())
	defer tk.Stop()

	conn := tk.AnonymousLogin()
	passiveConn := tk.PassiveConn(conn)
	kit.SetConnHooks(passiveConn, h)

	m := tk.Send(conn, "stor file1").Wait()
	content := "file1 content file1 content file1 content"

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		m.TakeAction(func() {
			write, err := passiveConn.Write([]byte(content))
			assert.NotNil(t.T(), err)
			assert.Less(t.T(), write, len(content))
		}).Failure()
		wg.Done()
	}()

	abort := tk.Send(conn, "abor")
	wg.Wait()
	abort.Success()

	assert.Less(t.T(), tk.Size(conn, "file1"), len(content))
}

func (t *ftpServerBaseCommandTest) TestReset() {
	tk := kit.NewTestKit(t.T())
	defer tk.Stop()

	conn := tk.AnonymousLogin()
	tk.Store(conn, "file", []byte("file content"))
	tk.Send(conn, "size file").Success("12")
	file := tk.Retrieve(conn, "file")
	assert.Equal(t.T(), []byte("file content"), file)

	tk.MustSuccess(conn, "REST 5")

	file = tk.Retrieve(conn, "file")
	assert.Equal(t.T(), []byte("content"), file)
}

func TestFTPServerTestSuite(t *testing.T) {
	suite.Run(t, new(ftpServerBaseCommandTest))
}
