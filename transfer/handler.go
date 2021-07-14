package transfer

import (
	"github.com/beyondstorage/beyond-ftp/utils"
)

// Handler presents active/passive transfer connection handler.
type Handler interface {
	// Open the connection to transfer data on.
	Open() (utils.Conn, error)

	// Close the connection (and any associated resource).
	Close() error
}
