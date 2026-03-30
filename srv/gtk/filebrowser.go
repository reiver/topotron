package gtksrv

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/lib/backend"
	"topotron/lib/fileinfo"
	"topotron/lib/format"
	"topotron/srv/log"
)

// FileBrowserPage displays the contents of a directory.
type FileBrowserPage struct {
	// gtk widgets
	page    *adw.NavigationPage
	listBox *gtk.ListBox
	stack   *gtk.Stack

	// state
	backend libbackend.FileBackend
	path    string
	entries []libfileinfo.FileInfo

	// callbacks
	OnDirectoryActivated func(path string)
}

// newFileBrowserPage creates a new [FileBrowserPage] for the given path.
func newFileBrowserPage(backend libbackend.FileBackend, path string) *FileBrowserPage {
	var receiver FileBrowserPage

	receiver.backend = backend
	receiver.path = path

	// file list
	receiver.listBox = gtk.NewListBox()
	receiver.listBox.SetSelectionMode(gtk.SelectionNone)
	receiver.listBox.AddCSSClass("boxed-list")
	receiver.listBox.SetMarginTop(12)
	receiver.listBox.SetMarginBottom(12)
	receiver.listBox.SetMarginStart(12)
	receiver.listBox.SetMarginEnd(12)

	receiver.listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		receiver.onRowActivated(row)
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	scrolled.SetChild(receiver.listBox)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(600)
	clamp.SetChild(scrolled)

	// placeholder for empty directories
	placeholder := adw.NewStatusPage()
	placeholder.SetIconName("folder-open-symbolic")
	placeholder.SetTitle("Empty Folder")

	// loading spinner
	spinner := gtk.NewSpinner()
	spinner.SetSizeRequest(32, 32)
	spinner.SetHAlign(gtk.AlignCenter)
	spinner.SetVAlign(gtk.AlignCenter)
	spinner.Start()

	loadingPage := adw.NewStatusPage()
	loadingPage.SetTitle("Loading")
	loadingPage.SetChild(spinner)

	// stack switches between loading, content, and placeholder
	receiver.stack = gtk.NewStack()
	receiver.stack.AddNamed(loadingPage, "loading")
	receiver.stack.AddNamed(clamp, "content")
	receiver.stack.AddNamed(placeholder, "placeholder")
	receiver.stack.SetVisibleChildName("loading")

	header := adw.NewHeaderBar()

	toolbar := adw.NewToolbarView()
	toolbar.AddTopBar(header)
	toolbar.SetContent(receiver.stack)

	dirName := filepath.Base(path)
	if path == "/" {
		dirName = "Root"
	}

	receiver.page = adw.NewNavigationPage(toolbar, dirName)

	receiver.load()

	return &receiver
}

// load reads the directory contents in a goroutine and populates the list.
func (receiver *FileBrowserPage) load() {
	go func() {
		entries, err := receiver.backend.List(context.Background(), receiver.path, false)

		glib.IdleAdd(func() {
			if nil != err {
				log := logsrv.Begin()
				defer log.End()
				log.Highlightf("could not load directory: %s", receiver.path)

				receiver.stack.SetVisibleChildName("placeholder")
				return
			}

			receiver.entries = entries
			receiver.populate()
		})
	}()
}

// populate fills the [gtk.ListBox] with rows for each file entry.
func (receiver *FileBrowserPage) populate() {
	if 0 == len(receiver.entries) {
		receiver.stack.SetVisibleChildName("placeholder")
		return
	}

	for _, entry := range receiver.entries {
		row := newFileRow(entry)
		receiver.listBox.Append(row)
	}

	receiver.stack.SetVisibleChildName("content")
}

// onRowActivated handles a tap on a file or directory row.
func (receiver *FileBrowserPage) onRowActivated(row *gtk.ListBoxRow) {
	index := row.Index()
	if index < 0 || index >= len(receiver.entries) {
		return
	}

	entry := receiver.entries[index]

	if entry.IsDir {
		if nil != receiver.OnDirectoryActivated {
			receiver.OnDirectoryActivated(entry.Path)
		}
		return
	}

	receiver.openFile(entry.Path)
}

// openFile launches a file with the default system handler.
func (receiver *FileBrowserPage) openFile(path string) {
	log := logsrv.Begin()
	defer log.End()

	err := exec.Command("xdg-open", path).Start()
	if nil != err {
		log.Highlightf("could not open file: %s", path)
	}
}

// newFileRow creates an [adw.ActionRow] for a single [libfileinfo.FileInfo].
func newFileRow(entry libfileinfo.FileInfo) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle(entry.Name)
	row.SetActivatable(true)

	icon := gtk.NewImageFromIconName(entry.Icon)
	row.AddPrefix(icon)

	if entry.IsDir {
		arrow := gtk.NewImageFromIconName("go-next-symbolic")
		row.AddSuffix(arrow)
	} else {
		sizeLabel := gtk.NewLabel(libformat.Size(entry.Size))
		sizeLabel.AddCSSClass("dim-label")
		row.AddSuffix(sizeLabel)
	}

	return row
}
