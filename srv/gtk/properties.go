package gtksrv

import (
	"context"
	"fmt"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/lib/backend"
	"topotron/lib/fileinfo"
	"topotron/lib/format"
	"topotron/srv/log"
)

// PropertiesPage displays metadata for a file or directory.
type PropertiesPage struct {
	// gtk widgets
	page *adw.NavigationPage
}

// newPropertiesPage creates a new [PropertiesPage] for the given entry.
// It fetches fresh metadata from the backend in a goroutine.
func newPropertiesPage(backend libbackend.FileBackend, entry libfileinfo.FileInfo) *PropertiesPage {
	var receiver PropertiesPage

	icon := gtk.NewImageFromIconName(entry.Icon)
	icon.SetPixelSize(64)
	icon.SetMarginTop(24)
	icon.SetMarginBottom(12)
	icon.SetHAlign(gtk.AlignCenter)

	nameLabel := gtk.NewLabel(entry.Name)
	nameLabel.AddCSSClass("title-1")
	nameLabel.SetWrap(true)
	nameLabel.SetHAlign(gtk.AlignCenter)
	nameLabel.SetMarginBottom(24)

	// properties list
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionNone)
	listBox.AddCSSClass("boxed-list")
	listBox.SetMarginStart(12)
	listBox.SetMarginEnd(12)
	listBox.SetMarginBottom(12)

	// type
	typeValue := "File"
	if entry.IsDir {
		typeValue = "Folder"
	}
	listBox.Append(newPropertyRow("Type", typeValue))

	// path
	listBox.Append(newPropertyRow("Location", entry.Path))

	// size
	if !entry.IsDir {
		sizeText := fmt.Sprintf("%s (%d bytes)", libformat.Size(entry.Size), entry.Size)
		listBox.Append(newPropertyRow("Size", sizeText))
	}

	// modified
	modText := entry.ModTime.Format("2006-01-02 15:04:05")
	listBox.Append(newPropertyRow("Modified", modText))

	// for directories, compute total size in background
	sizeRow := newPropertyRow("Size", "Calculating...")
	if entry.IsDir {
		listBox.Append(sizeRow)
		go func() {
			totalSize := computeDirSize(backend, entry.Path)
			glib.IdleAdd(func() {
				sizeText := fmt.Sprintf("%s (%d bytes)", libformat.Size(totalSize), totalSize)
				sizeRow.SetSubtitle(sizeText)
			})
		}()
	}

	content := gtk.NewBox(gtk.OrientationVertical, 0)
	content.Append(icon)
	content.Append(nameLabel)
	content.Append(listBox)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	scrolled.SetChild(content)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(600)
	clamp.SetChild(scrolled)

	header := adw.NewHeaderBar()

	toolbar := adw.NewToolbarView()
	toolbar.AddTopBar(header)
	toolbar.SetContent(clamp)

	receiver.page = adw.NewNavigationPage(toolbar, "Properties")

	return &receiver
}

// newPropertyRow creates an [adw.ActionRow] displaying a property name and value.
func newPropertyRow(name, value string) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle(name)
	row.SetSubtitle(value)
	row.SetSubtitleSelectable(true)

	return row
}

// computeDirSize recursively computes the total size of a directory.
func computeDirSize(backend libbackend.FileBackend, dirPath string) int64 {
	log := logsrv.Begin()
	defer log.End()

	entries, err := backend.List(context.Background(), dirPath, true)
	if nil != err {
		return 0
	}

	var total int64
	for _, entry := range entries {
		if entry.IsDir {
			total += computeDirSize(backend, entry.Path)
		} else {
			total += entry.Size
		}
	}

	return total
}
