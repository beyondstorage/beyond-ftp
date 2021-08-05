package client

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/beyondstorage/go-storage/v4/pairs"
	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/beyondstorage/go-storage/v4/types"
)

func (c *Handler) absPath(p string) string {
	curPath := c.Path()

	p = path.Clean(p)
	if path.IsAbs(p) {
		return p
	}
	return path.Join(curPath, p)
}

func (c *Handler) handleCWD() {
	if c.param == ".." {
		c.handleCDUP()
		return
	}

	p := c.absPath(c.param)

	_, err := c.getDirInfo(p)
	if err != nil {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("CD issue: %v", err))
		return
	}
	c.SetPath(p)
	c.WriteMessage(StatusFileOK, fmt.Sprintf("CD worked on %s", p))
}

func (c *Handler) handleMKD() {
	p := c.absPath(c.param)
	_, err := c.getDirInfo(p)
	if err == nil {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("Dir already exists: %s", p))
		return
	}
	if !errors.Is(err, services.ErrObjectNotExist) {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("Could not create %s : %v", p, err))
		return
	}
	direr, ok := c.storager.(types.Direr)
	if !ok {
		c.WriteMessage(StatusCommandNotImplemented, fmt.Sprintf("This type of storage is not support create dir"))
		return
	}
	if _, err := direr.CreateDir(p); err != nil {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("Could not create %s : %v", p, err))
		return
	}

	c.WriteMessage(StatusPathCreated, fmt.Sprintf("Created dir %s", p))
}

func (c *Handler) getFileInfo(p string) (*fileInfo, error) {
	o, err := c.storager.Stat(p)
	return &fileInfo{o}, err
}

func (c *Handler) getDirInfo(p string) (*fileInfo, error) {
	o, err := c.storager.Stat(p, pairs.WithObjectMode(types.ModeDir))
	return &fileInfo{o}, err
}

func (c *Handler) handleRMD() {
	p := c.absPath(c.param)
	err := c.storager.DeleteWithContext(c.commandAbortCtx, p)
	if err != nil {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("Could not delete dir %s: %v", p, err))
		return
	}
	c.WriteMessage(StatusFileOK, fmt.Sprintf("Deleted dir %s", p))
}

func (c *Handler) handleCDUP() {
	if c.Path() == "/" {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("cannot CDUP"))
		return
	}
	parent := path.Dir(c.Path())
	c.SetPath(parent)
	c.WriteMessage(StatusFileOK, fmt.Sprintf("CDUP worked on %s", parent))
}

func (c *Handler) handlePWD() {
	c.WriteMessage(StatusPathCreated, "\""+c.Path()+"\" is the current directory")
}

func (c *Handler) handleLIST() {
	dir := c.absPath(c.param)

	fileInfos, err := c.listFile(dir)
	if err != nil {
		c.WriteMessage(StatusActionNotTaken, err.Error())
		return
	}

	tr, err := c.TransferOpen()
	if err != nil {
		c.WriteMessage(StatusCannotOpenDataConnection, err.Error())
		return
	}
	c.dirList(tr, fileInfos)

	select {
	case <-c.commandAbortCtx.Done():
		c.WriteMessage(StatusTransferAborted, "Connection closed; transfer aborted")
	default:
		c.TransferClose()
		c.WriteMessage(StatusClosingDataConn, "")
	}
}

func (c *Handler) listFile(p string) ([]*fileInfo, error) {
	iterator, err := c.storager.List(p)
	if err != nil {
		return nil, err
	}

	var files []*fileInfo
	for {
		o, err := iterator.Next()
		if err != nil {
			if errors.Is(err, types.IterateDone) {
				break
			} else {
				return nil, err
			}
		}
		files = append(files, &fileInfo{o})
	}
	return files, nil
}

func fileStat(file *fileInfo) string {
	return fmt.Sprintf(
		"%s 1 ftp ftp %12d %s %s",
		file.Mode(),
		file.Size(),
		file.ModTime().Format(" Jan _2 15:04 "),
		file.Name(),
	)
}

func (c *Handler) dirList(w io.Writer, files []*fileInfo) {
	for _, file := range files {
		stat := fileStat(file)
		if _, err := fmt.Fprintf(w, "%s\r\n", stat); err != nil {
			return
		}
	}
}

type fileInfo struct {
	*types.Object
}

func (f *fileInfo) Mode() os.FileMode {
	if f.GetMode().IsDir() {
		return os.ModeDir
	}
	return os.ModePerm
}

func (f *fileInfo) Size() int64 {
	return f.MustGetContentLength()
}

func (f *fileInfo) Name() string {
	return path.Base(f.GetPath())
}

func (f *fileInfo) ModTime() time.Time {
	modified, _ := f.GetLastModified()
	return modified
}
