package logsrv

import (
	"io"
	"os"

	"codeberg.org/reiver/go-log"

	"topotron/cfg"
)

var writer io.Writer = os.Stdout

var logger log.Logger = log.CreateLogger(writer, cfg.LogLevel())
