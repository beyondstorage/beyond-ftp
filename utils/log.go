package utils

import (
	"github.com/pengsrc/go-shared/check"
	"github.com/pengsrc/go-shared/log"
)

// Logger is the global logger for BeyondFTP
var Logger *log.ContextFreeLogger

func init() {
	// Setup logger.
	l, err := log.NewTerminalLogger("debug")
	check.ErrorForExit("log init error: ", err)
	Logger = log.NewContextFreeLogger(l)
}
