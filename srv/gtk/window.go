package gtksrv

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/lib/place"
	"topotron/srv/log"
)

// Window is the main application window.
type Window struct {
	// gtk widgets
	window       *adw.ApplicationWindow
	navView      *adw.NavigationView
	toastOverlay *adw.ToastOverlay

	// pages
	placesPage *PlacesPage
}

// newWindow creates a new [Window] and attaches it to the given [adw.Application].
func newWindow(app *adw.Application) *Window {
	var receiver Window

	receiver.placesPage = newPlacesPage()
	receiver.placesPage.OnActivated = receiver.onPlaceActivated

	receiver.navView = adw.NewNavigationView()
	receiver.navView.Add(receiver.placesPage.page)

	receiver.toastOverlay = adw.NewToastOverlay()
	receiver.toastOverlay.SetChild(receiver.navView)

	receiver.window = adw.NewApplicationWindow(&app.Application)
	receiver.window.SetTitle("Topotron")
	receiver.window.SetDefaultSize(360, 648)
	receiver.window.SetContent(receiver.toastOverlay)

	return &receiver
}

func (receiver *Window) onPlaceActivated(place libplace.Place) {
	log := logsrv.Begin()
	defer log.End()

	log.Highlightf("place activated: %s (%s)", place.Name, place.Path)

	// TODO: Phase 3 — push file browser page onto navView
}
