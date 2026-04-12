package gui

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/cfg"
	"topotron/srv/log"
)

// Run creates the [adw.Application] and runs the GTK main loop.
// It returns the exit code for the process.
func Run() int {
	log := logsrv.Begin()
	defer log.End()

	app := adw.NewApplication(cfg.AppID, gio.ApplicationFlagsNone)

	app.ConnectActivate(func() {
		onActivate(app)
	})

	return app.Run(os.Args)
}

func onActivate(app *adw.Application) {
	win := app.ActiveWindow()
	if nil == win {
		w := newWindow(app)
		w.window.Present()
		return
	}
	win.Present()
}
