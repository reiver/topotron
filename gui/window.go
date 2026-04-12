package gui

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/lib/backend"
	"topotron/lib/fileinfo"
	"topotron/lib/place"
	"topotron/srv/discover"
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

	// services
	discovery *discoversrv.Discovery

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
	receiver.placesPage.OnNetworkActivated = receiver.onNetworkActivated
	receiver.placesPage.OnWebDAVActivated = receiver.onWebDAVActivated
	receiver.placesPage.OnAddWebDAV = receiver.onAddWebDAV
	receiver.placesPage.OnAbout = func() {
		showAboutDialog(&receiver.window.Window)
	}
	receiver.placesPage.OnUnpin = func(path string) {
		receiver.settings.RemovePinnedDir(path)
		receiver.placesPage.RebuildLocalList()
		receiver.showToast("Unpinned from Places")
	}
	receiver.placesPage.OnResetPins = func() {
		receiver.settings.ResetPinnedDirs()
		receiver.placesPage.RebuildLocalList()
		receiver.showToast("Pinned folders reset")
	}
	receiver.placesPage.OnChangeIcon = func(path, currentIcon string) {
		showIconPicker(&receiver.window.Window, currentIcon, func(icon string) {
			receiver.settings.SetPinnedDirIcon(path, icon)
			receiver.placesPage.RebuildLocalList()
		})
	}

	receiver.navView = adw.NewNavigationView()
	receiver.navView.Add(receiver.placesPage.page)

	receiver.toastOverlay = adw.NewToastOverlay()
	receiver.toastOverlay.SetChild(receiver.navView)

	receiver.window = adw.NewApplicationWindow(&app.Application)
	receiver.window.SetTitle("Topotron")
	receiver.window.SetDefaultSize(360, 648)
	receiver.window.SetContent(receiver.toastOverlay)

	setupShortcuts(&receiver)

	// start mDNS discovery
	receiver.discovery = discoversrv.New()
	receiver.discovery.OnAdded = func(service discoversrv.DiscoveredService) {
		receiver.placesPage.AddNetworkService(service)
	}
	receiver.discovery.OnRemoved = func(name string) {
		receiver.placesPage.RemoveNetworkService(name)
	}

	receiver.window.ConnectCloseRequest(func() bool {
		receiver.discovery.Stop()
		return false
	})

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
	receiver.pushFileBrowserWithBackend(receiver.backend, place.Path)
}

// onNetworkActivated handles a discovered network service being tapped.
func (receiver *Window) onNetworkActivated(service discoversrv.DiscoveredService) {
	log := logsrv.Begin()
	defer log.End()

	webdavBackend := libbackend.NewWebDAVBackend(service.URL(), "", "", service.Name)

	err := webdavBackend.Connect()
	if nil != err {
		log.Highlightf("could not connect to network share: %s", service.URL())
		receiver.showToast("Could not connect to share")
		return
	}

	receiver.pushFileBrowserWithBackend(webdavBackend, "/")
}

// onWebDAVActivated handles a WebDAV bookmark being tapped.
func (receiver *Window) onWebDAVActivated(bookmark settingsrv.Bookmark) {
	log := logsrv.Begin()
	defer log.End()

	webdavBackend := libbackend.NewWebDAVBackend(bookmark.URL, bookmark.Username, bookmark.Password, bookmark.Name)

	err := webdavBackend.Connect()
	if nil != err {
		log.Highlightf("could not connect to WebDAV: %s", bookmark.URL)
		receiver.showToast("Could not connect to server")
		return
	}

	receiver.pushFileBrowserWithBackend(webdavBackend, "/")
}

// onAddWebDAV shows the dialog for adding a new WebDAV server.
func (receiver *Window) onAddWebDAV() {
	showAddWebDAVDialog(receiver.window, receiver.settings, func() {
		receiver.placesPage.RebuildWebDAVList()
	})
}

// pushFileBrowser creates a new [FileBrowserPage] for the given path
// using the local backend and pushes it onto the [adw.NavigationView].
func (receiver *Window) pushFileBrowser(path string) {
	receiver.pushFileBrowserWithBackend(receiver.backend, path)
}

// pushFileBrowserWithBackend creates a new [FileBrowserPage] for the given
// path and backend, and pushes it onto the [adw.NavigationView].
func (receiver *Window) pushFileBrowserWithBackend(backend libbackend.FileBackend, path string) {
	page := newFileBrowserPage(backend, path, receiver.sortOrder(), receiver.showHidden())
	page.OnDirectoryActivated = func(subPath string) {
		receiver.pushFileBrowserWithBackend(backend, subPath)
	}
	page.OnCut = receiver.onCut
	page.OnCopy = receiver.onCopy
	page.OnPaste = receiver.onPaste
	page.HasClipboard = receiver.hasClipboard
	page.OnRefreshNeeded = receiver.onRefreshNeeded
	page.OnSortChanged = receiver.onSortChanged
	page.OnProperties = func(entry libfileinfo.FileInfo) {
		receiver.showProperties(backend, entry)
	}
	page.OnRename = func(entry libfileinfo.FileInfo) {
		receiver.showRename(backend, entry, path)
	}
	page.OnPin = func(entry libfileinfo.FileInfo) {
		receiver.settings.AddPinnedDir(settingsrv.PinnedDir{
			Name: entry.Name,
			Path: entry.Path,
		})
		receiver.placesPage.RebuildLocalList()
		receiver.showToast("Pinned to Places")
	}
	page.UpdatePasteButton()
	receiver.navView.Push(page.page)
}

// showRename presents the rename dialog for a file or directory.
func (receiver *Window) showRename(backend libbackend.FileBackend, entry libfileinfo.FileInfo, dirPath string) {
	showRenameDialog(&receiver.window.Window, backend, entry, func() {
		receiver.pushFileBrowserWithBackend(backend, dirPath)
	})
}

// showProperties pushes a [PropertiesPage] onto the navigation view.
func (receiver *Window) showProperties(backend libbackend.FileBackend, entry libfileinfo.FileInfo) {
	props := newPropertiesPage(backend, entry)
	receiver.navView.Push(props.page)
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
		// TODO: show progress in loading overlay
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
