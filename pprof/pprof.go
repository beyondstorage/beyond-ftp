package pprof

import (
	"net/http"
	_ "net/http/pprof"
)

func StartPP() {
	go func() {
		err := http.ListenAndServe("localhost:6060", nil)
		if err != nil {
			panic(err)
		}
	}()
}
