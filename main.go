package main

import (
	"github.com/beyondstorage/beyond-ftp/cmd"
	"github.com/beyondstorage/beyond-ftp/pprof"
)

func main() {
	pprof.StartPP()
	cmd.Execute()
}
