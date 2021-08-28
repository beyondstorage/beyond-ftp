package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/beyondstorage/go-storage/v4/pairs"
	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/beyondstorage/go-storage/v4/types"
	"github.com/beyondstorage/go-stream"

	"github.com/beyondstorage/beyond-ftp/utils"
)

func (c *Handler) handleSTOR() {
	path := c.absPath(c.param)

	c.storeOrAppend(path, false)
}

func (c *Handler) handleAPPE() {
	_, ok := c.storager.(types.Appender)
	if !ok {
		c.WriteMessage(StatusCommandNotImplemented, "this type of storage is not support append")
		return
	}
	path := c.absPath(c.param)
	c.storeOrAppend(path, true)
}

func (c *Handler) storeOrAppend(path string, append bool) {
	c.ctxRest = 0
	tr, err := c.TransferOpen()
	if err != nil {
		c.WriteMessage(StatusCannotOpenDataConnection, err.Error())
		return
	}

	if err := c.upload(path, tr, append); err != nil {
		c.TransferClose()
		c.WriteMessage(StatusFileActionNotTaken, err.Error())
		return
	}

	select {
	case <-c.commandAbortCtx.Done():
		c.WriteMessage(StatusTransferAborted, "Connection closed; transfer aborted")
	default:
		c.TransferClose()
		c.WriteMessage(StatusClosingDataConn, "transfer finished")
	}
}

func (c *Handler) upload(path string, tr utils.Conn, append bool) error {
	if appender, ok := c.storager.(types.Appender); ok {
		object, err := c.storager.Stat(path)
		if err != nil && !errors.Is(err, services.ErrObjectNotExist) {
			return err
		}

		if !append || errors.Is(err, services.ErrObjectNotExist) {
			object, err = appender.CreateAppendWithContext(c.commandAbortCtx, path)
			if err != nil {
				return err
			}
		}
		
		if err = writeToObject(c.commandAbortCtx, stream.PersistMethodAppend, path, tr, c.storager, object); err != nil {
			return err
		}
		return nil
	}

	if err := writeToObject(c.commandAbortCtx, stream.PersistMethodMultipart, path, tr, c.storager, nil); err != nil {
		return err
	}
	return nil
}

func writeToObject(ctx context.Context, label, path string, r io.Reader, storager types.Storager, o *types.Object) error {
	b := utils.Branch(label, path)
	if b != nil {
		_, err := b.ReadFrom(r)
		if err != nil {
			return err
		}
		return b.Complete()
	}

	file := new(bytes.Buffer)
	size, err := io.Copy(file, r)
	if err != nil {
		return err
	}
	switch label {
	case stream.PersistMethodAppend:
		appender := storager.(types.Appender)
		_, err = appender.WriteAppendWithContext(ctx, o, file, size)
		if err != nil {
			return err
		}
		err = appender.CommitAppendWithContext(ctx, o)
		if err != nil {
			return err
		}
	case stream.PersistMethodMultipart:
		_, err := storager.WriteWithContext(ctx, path, file, size)
		return err
	default:
		return fmt.Errorf("label %s not support", label)
	}

	return nil
}

func (c *Handler) handleRETR() {
	defer func() {
		c.ctxRest = 0
	}()
	path := c.absPath(c.param)
	tr, err := c.TransferOpen()
	if err != nil {
		c.WriteMessage(StatusCannotOpenDataConnection, err.Error())
		return
	}

	_, err = c.storager.ReadWithContext(c.commandAbortCtx, path, tr, pairs.WithOffset(c.ctxRest))
	if err != nil {
		c.TransferClose()
		c.WriteMessage(StatusActionNotTaken, err.Error())
		return
	}

	select {
	case <-c.commandAbortCtx.Done():
		c.WriteMessage(StatusTransferAborted, "Connection closed; transfer aborted")
	default:
		c.TransferClose()
		c.WriteMessage(StatusClosingDataConn, "transfer finished")
	}
}

func (c *Handler) handleDELE() {
	path := c.absPath(c.param)
	err := c.storager.DeleteWithContext(c.commandAbortCtx, path)
	if err != nil {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't delete %s: %v", path, err))
		return
	}
	c.WriteMessage(StatusFileOK, fmt.Sprintf("Removed file %s", path))
}

func (c *Handler) handleRNFR() {
	path := c.absPath(c.param)
	_, err := c.storager.Stat(path)
	if err != nil {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't access %s: %v", path, err))
		return
	}
	c.WriteMessage(StatusFileActionPending, "Sure, give me a target")
	c.ctxRnfr = path
}

func (c *Handler) handleRNTO() {
	path := c.absPath(c.param)
	mover, ok := c.storager.(types.Mover)
	if !ok {
		c.WriteMessage(StatusCommandNotImplemented, "this type of storage is not support rename")
		return
	}

	if c.ctxRnfr == "" {
		c.WriteMessage(StatusBadCommandSequence, "RNFR is expected before RNTO")
		return
	}

	err := mover.Move(c.ctxRnfr, path)
	if err != nil {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't rename file: %v", err))
		return
	}

	c.WriteMessage(StatusFileOK, "Done !")
	c.ctxRnfr = ""
}

func (c *Handler) handleSIZE() {
	path := c.absPath(c.param)
	object, err := c.storager.Stat(path)
	if err != nil {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't access %s: %v", path, err))
		return
	}
	length, ok := object.GetContentLength()
	if !ok {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't access %s: %v", path, err))
		return
	}
	c.WriteMessage(StatusFileStatus, fmt.Sprintf("%d", length))
}

func (c *Handler) handleSTATFile() {
	path := c.absPath(c.param)

	fileInfos, err := c.listFile(path)
	if err != nil {
		c.WriteMessage(StatusFileActionNotTaken, err.Error())
		return
	}
	c.dirList(c.writer, fileInfos)
	c.writeLine("213 End of status")
}

func (c *Handler) handleALLO() {
	c.WriteMessage(StatusNotImplemented, "OK, we have the free space")
}

func (c *Handler) handleREST() {
	if size, err := strconv.ParseInt(c.param, 10, 0); err == nil {
		c.ctxRest = size
		c.WriteMessage(StatusFileActionPending, "OK")
	} else {
		c.WriteMessage(StatusSyntaxErrorParameters, fmt.Sprintf("Couldn't parse size: %v", err))
	}
}

func (c *Handler) handleMDTM() {
	path := c.absPath(c.param)
	object, err := c.storager.Stat(path)
	if err != nil {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't access %s: %s", path, err.Error()))
		return
	}
	lastModified, ok := object.GetLastModified()
	if !ok {
		c.WriteMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't access %s", path))
		return
	}
	c.WriteMessage(StatusFileOK, lastModified.UTC().Format("20060102150405"))
}
