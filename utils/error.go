package utils

import (
	"log"
	"runtime/debug"
)

func MustNil(e error) {
	if e != nil {
		debug.PrintStack()
		log.Fatalf("error occured %v", e)
	}
}
