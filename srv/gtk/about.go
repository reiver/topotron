package gtksrv

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/cfg"
)

// showAboutDialog presents the application about dialog.
func showAboutDialog(parent *gtk.Window) {
	about := adw.NewAboutWindow()
	about.SetTransientFor(parent)
	about.SetApplicationName("Topotron")
	about.SetApplicationIcon(cfg.AppID)
	about.SetVersion(cfg.Version)
	about.SetDeveloperName("Charles Iliya Krempeaux")
	about.SetWebsite("https://reiver.link")
	about.SetLicenseType(gtk.LicenseMITX11)
	about.SetCopyright("© 2026 Charles Iliya Krempeaux")
	about.SetComments("A mobile-optimized file manager for GNOME.\nSupports local filesystems, WebDAV, and network share discovery.")

	about.Present()
}
