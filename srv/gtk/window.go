package gtksrv

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/lib/backend"
	"topotron/lib/fileinfo"
	"topotron/lib/place"
	"topotron/srv/log"
	"topotron/srv/op"
	"topotron/srv/settings"
)

// ClipboardMode indicates whether entries were cut or copied.
type ClipboardMode int

const (
	ClipboardCut ClipboardMode = iota + 1
	ClipboardCopy
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
	backend       libbackend.FileBackend
	settings      *settingsrv.Settings
	clipboard     []libfileinfo.FileInfo
	clipboardMode ClipboardMode
}

// newWindow creates a new [Window] and attaches it to the given [adw.Application].
func newWindow(app *adw.Application) *Window {
	var receiver Window

	receiver.backend = libbackend.LocalBackend{}
	receiver.settings = settingsrv.New()

	receiver.placesPage = newPlacesPage(receiver.settings)
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

// sortOrder returns the current [SortOrder] from settings.
func (receiver *Window) sortOrder() SortOrder {
	return SortOrder(receiver.settings.SortOrder())
}

// showHidden returns whether hidden files should be shown.
func (receiver *Window) showHidden() bool {
	return receiver.settings.ShowHidden()
}

// onPlaceActivated handles a place being tapped on the home screen.
func (receiver *Window) onPlaceActivated(place libplace.Place) {
	receiver.pushFileBrowser(place.Path)
}

// pushFileBrowser creates a new [FileBrowserPage] for the given path
// and pushes it onto the [adw.NavigationView].
func (receiver *Window) pushFileBrowser(path string) {
	page := newFileBrowserPage(receiver.backend, path, receiver.sortOrder(), receiver.showHidden())
	page.OnDirectoryActivated = receiver.pushFileBrowser
	page.OnCut = receiver.onCut
	page.OnCopy = receiver.onCopy
	page.OnPaste = receiver.onPaste
	page.HasClipboard = receiver.hasClipboard
	page.OnRefreshNeeded = receiver.onRefreshNeeded
	page.OnSortChanged = receiver.onSortChanged
	page.UpdatePasteButton()
	receiver.navView.Push(page.page)
}

// hasClipboard returns whether the clipboard has entries.
func (receiver *Window) hasClipboard() bool {
	return len(receiver.clipboard) > 0
}

// onCut stores entries in the clipboard for moving.
func (receiver *Window) onCut(entries []libfileinfo.FileInfo) {
	receiver.clipboard = entries
	receiver.clipboardMode = ClipboardCut

	receiver.showToast("Files will be moved")
}

// onCopy stores entries in the clipboard for copying.
func (receiver *Window) onCopy(entries []libfileinfo.FileInfo) {
	receiver.clipboard = entries
	receiver.clipboardMode = ClipboardCopy

	receiver.showToast("Files will be copied")
}

// onPaste executes the clipboard operation into the given destination directory.
func (receiver *Window) onPaste(destPath string) {
	if 0 == len(receiver.clipboard) {
		return
	}

	log := logsrv.Begin()
	defer log.End()

	entries := receiver.clipboard
	mode := receiver.clipboardMode

	receiver.clipboard = nil
	receiver.clipboardMode = 0

	ctx := context.Background()

	onProgress := func(current, total int, name string) {
		// TODO: Phase 9 — show progress in loading overlay
	}

	onDone := func(err error) {
		if nil != err {
			log.Highlightf("paste failed: %v", err)
			receiver.showToast("Paste failed")
		} else {
			receiver.showToast("Paste complete")
		}
		receiver.onRefreshNeeded(destPath)
	}

	switch mode {
	case ClipboardCut:
		opsrv.MoveFiles(ctx, receiver.backend, entries, destPath, onProgress, onDone)
	case ClipboardCopy:
		opsrv.CopyFiles(ctx, receiver.backend, entries, destPath, onProgress, onDone)
	}
}

// onSortChanged persists the sort order when changed in a file browser page.
func (receiver *Window) onSortChanged(order SortOrder) {
	receiver.settings.SetSortOrder(string(order))
}

// onRefreshNeeded refreshes the file browser page for the given path.
func (receiver *Window) onRefreshNeeded(path string) {
	_ = path
}

// showToast displays a brief notification.
func (receiver *Window) showToast(message string) {
	toast := adw.NewToast(message)
	receiver.toastOverlay.AddToast(toast)
}
