package gtksrv

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/lib/backend"
	"topotron/lib/place"
)

// Window is the main application window.
type Window struct {
	// gtk widgets
	window       *adw.ApplicationWindow
	navView      *adw.NavigationView
	toastOverlay *adw.ToastOverlay

	// pages
	placesPage *PlacesPage

	// state
	backend libbackend.FileBackend
}

// newWindow creates a new [Window] and attaches it to the given [adw.Application].
func newWindow(app *adw.Application) *Window {
	var receiver Window

	receiver.backend = libbackend.LocalBackend{}

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

// onPlaceActivated handles a place being tapped on the home screen.
func (receiver *Window) onPlaceActivated(place libplace.Place) {
	receiver.pushFileBrowser(place.Path)
}

// pushFileBrowser creates a new [FileBrowserPage] for the given path
// and pushes it onto the [adw.NavigationView].
func (receiver *Window) pushFileBrowser(path string) {
	page := newFileBrowserPage(receiver.backend, path)
	page.OnDirectoryActivated = receiver.pushFileBrowser
	receiver.navView.Push(page.page)
}
