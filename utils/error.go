package utils

import (
	"runtime/debug"

	"go.uber.org/zap"
)

func MustNil(e error) {
	if e != nil {
		zap.L().Fatal("error occurred", zap.Error(e), zap.String("trace", string(debug.Stack())))
	}
}
