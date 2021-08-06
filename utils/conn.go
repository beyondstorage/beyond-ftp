package utils

import (
	"io"
)

type Conn interface {
	io.Writer
	io.Reader
	io.Closer
}
