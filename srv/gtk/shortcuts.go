package gtksrv

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// setupShortcuts registers keyboard shortcuts on the application window.
func setupShortcuts(window *Window) {
	controller := gtk.NewEventControllerKey()
	controller.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) bool {
		ctrl := state&gdk.ControlMask != 0

		if ctrl && keyval == gdk.KEY_q {
			window.window.Close()
			return true
		}

		return false
	})
	window.window.AddController(controller)
}
