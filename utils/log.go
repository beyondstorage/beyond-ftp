package utils

import (
	"go.uber.org/zap"
)

func SetUpLog() {
	// Setup logger.
	logger, err := zap.NewProduction()
	MustNil(err)
	zap.ReplaceGlobals(logger)
}
