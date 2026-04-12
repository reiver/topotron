package main

import (
	"os"

	"topotron/gui"
	"topotron/srv/log"
)

func main() {
	log := logsrv.Begin()
	defer log.End()

	log.Highlightf("topotron ⚡")
	defer log.Highlightf("topotron 👻")

	os.Exit(gui.Run())
}
