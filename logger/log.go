package logger

import (
	"go.uber.org/zap"
)

func SetUpLog() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}
