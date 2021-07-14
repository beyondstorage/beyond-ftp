package client

import (
	"fmt"
	"strings"
	"time"
)

func (c *Handler) handleSYST() {
	c.WriteMessage(StatusSystemType, "UNIX Type: L8")
}

func (c *Handler) handleSTAT() {
	if c.param == "" { // Without a file, it's the server stat.
		c.handleSTATServer()
	} else { // With a file/dir it's the file or the dir's files stat.
		c.handleSTATFile()
	}
}

func (c *Handler) handleSITE() {
	spl := strings.SplitN(c.param, " ", 2)
	if len(spl) > 1 {
		if strings.ToUpper(spl[0]) == "CHMOD" {
			c.handleCHMOD(spl[1])
		}
	}

	c.WriteMessage(StatusOK, "")
}

func (c *Handler) handleSTATServer() {
	c.writeLine("213- FTP server status:")
	duration := time.Now().UTC().Sub(c.connectedAt)
	duration -= duration % time.Second
	c.writeLine(fmt.Sprintf(
		"Connected to %s:%d from %s for %s",
		c.serverSetting.ListenHost, c.serverSetting.ListenPort,
		c.remoteAddr,
		duration,
	))
	c.writeLine(fmt.Sprintf("Logged in as %s", c.loginUser))
	c.writeLine("ftpserver - golang FTP server")
	c.WriteMessage(StatusFileStatus, "End")
}

func (c *Handler) handleOPTS() {
	args := strings.SplitN(c.param, " ", 2)
	if strings.ToUpper(args[0]) == "UTF8" {
		c.WriteMessage(StatusOK, "I'm in UTF8 only anyway")
	} else {
		c.WriteMessage(StatusSyntaxErrorNotRecognised, "Don't know this option")
	}
}

func (c *Handler) handleNOOP() {
	c.WriteMessage(StatusOK, "OK")
}

func (c *Handler) handleFEAT() {
	c.writeLine("211- These are my features")
	defer c.WriteMessage(StatusSystemStatus, "End")

	features := []string{
		"UTF8",
		"SIZE",
		"MDTM",
		"REST STREAM",
	}

	for _, f := range features {
		c.writeLine(" " + f)
	}
}

func (c *Handler) handleTYPE() {
	switch c.param {
	case "I":
		c.WriteMessage(StatusOK, "Type set to binary")
	case "A":
		c.WriteMessage(StatusOK, "Type set to ASCII")
	default:
		c.WriteMessage(StatusSyntaxErrorNotRecognised, "Not understood")
	}
}

func (c *Handler) handleQUIT() {
	c.WriteMessage(StatusClosingControlConn, "Goodbye")
	c.disconnect()
}

func (c *Handler) handleABOR() {
	c.commandAbortCancelFn()  // abort command
	c.TransferClose()         // close transfer connection
	c.commandRunningWg.Wait() // wait for command abort
	c.WriteMessage(StatusClosingDataConn, "abort command was successfully processed")
}
