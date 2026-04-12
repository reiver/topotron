package gui

import (
	"context"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/lib/backend"
	"topotron/lib/fileinfo"
	"topotron/srv/log"
)

// showRenameDialog presents a dialog for renaming a file or directory.
func showRenameDialog(parent *gtk.Window, backend libbackend.FileBackend, entry libfileinfo.FileInfo, onRenamed func()) {
	dialog := gtk.NewWindow()
	dialog.SetTitle("Rename")
	dialog.SetDefaultSize(360, 0)
	dialog.SetModal(true)
	dialog.SetTransientFor(parent)

	nameEntry := adw.NewEntryRow()
	nameEntry.SetTitle("Name")
	nameEntry.SetText(entry.Name)

	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionNone)
	listBox.AddCSSClass("boxed-list")
	listBox.SetMarginTop(12)
	listBox.SetMarginBottom(12)
	listBox.SetMarginStart(12)
	listBox.SetMarginEnd(12)
	listBox.Append(nameEntry)

	renameBtn := gtk.NewButtonWithLabel("Rename")
	renameBtn.AddCSSClass("suggested-action")
	renameBtn.ConnectClicked(func() {
		newName := nameEntry.Text()
		if "" == newName || newName == entry.Name {
			dialog.Close()
			return
		}

		dir := filepath.Dir(entry.Path)
		newPath := filepath.Join(dir, newName)

		log := logsrv.Begin()
		defer log.End()

		err := backend.Rename(context.Background(), entry.Path, newPath)
		if nil != err {
			log.Highlightf("could not rename: %s", entry.Path)
			dialog.Close()
			return
		}

		dialog.Close()

		if nil != onRenamed {
			onRenamed()
		}
	})

	cancelBtn := gtk.NewButtonWithLabel("Cancel")
	cancelBtn.ConnectClicked(func() {
		dialog.Close()
	})

	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	btnBox.SetHAlign(gtk.AlignEnd)
	btnBox.SetMarginStart(12)
	btnBox.SetMarginEnd(12)
	btnBox.SetMarginBottom(12)
	btnBox.Append(cancelBtn)
	btnBox.Append(renameBtn)

	content := gtk.NewBox(gtk.OrientationVertical, 0)
	content.Append(listBox)
	content.Append(btnBox)

	header := adw.NewHeaderBar()

	toolbar := adw.NewToolbarView()
	toolbar.AddTopBar(header)
	toolbar.SetContent(content)

	dialog.SetChild(toolbar)
	dialog.Present()
}
