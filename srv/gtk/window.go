package gtksrv

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/cfg"
)

// Window is the main application window.
type Window struct {
	// gtk widgets
	window  *adw.ApplicationWindow
	header  *adw.HeaderBar
	toolbar *adw.ToolbarView
	content *gtk.Box
}

// newWindow creates a new [Window] and attaches it to the given [adw.Application].
func newWindow(app *adw.Application) *Window {
	var receiver Window

	receiver.header = adw.NewHeaderBar()

	receiver.content = gtk.NewBox(gtk.OrientationVertical, 0)
	receiver.content.SetVExpand(true)

	receiver.toolbar = adw.NewToolbarView()
	receiver.toolbar.AddTopBar(receiver.header)
	receiver.toolbar.SetContent(receiver.content)

	receiver.window = adw.NewApplicationWindow(&app.Application)
	receiver.window.SetTitle("Topotron")
	receiver.window.SetDefaultSize(360, 648)
	receiver.window.SetContent(receiver.toolbar)

	_ = cfg.Version

	return &receiver
}
