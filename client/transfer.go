package client

import (
	"fmt"
	"strings"

	"github.com/beyondstorage/beyond-ftp/utils"
)

func (c *Handler) handlePASV() {
	p, port, err := c.passiveTransferFactory(c.serverSetting.ListenHost, c.serverSetting.DataPortRange)
	if err != nil {
		c.WriteMessage(StatusCannotOpenDataConnection, "Can't open data connection.")
		return
	}

	publicHost := c.serverSetting.PublicHost

	// We should rewrite this part.
	if c.command == PASV {
		p1 := port / 256
		p2 := port - (p1 * 256)

		quads := strings.Split(publicHost, ".")
		c.WriteMessage(StatusEnteringPASV, fmt.Sprintf("Entering Passive Mode (%s,%s,%s,%s,%d,%d)", quads[0], quads[1], quads[2], quads[3], p1, p2))
	} else {
		c.WriteMessage(StatusEnteringEPSV, fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", port))
	}

	c.transfer = p
}

func (c *Handler) handlePORT() {
	addr := utils.ParseRemoteAddr(c.param)
	c.transfer = c.activeTransferFactory(addr)
	c.WriteMessage(StatusOK, "PORT command successful")
}
