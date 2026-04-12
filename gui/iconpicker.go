package gui

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
)

// folderIcons is the list of available folder icons from the Adwaita theme.
var folderIcons = []struct {
	Name string
	Icon string
}{
	{"Folder", "folder-symbolic"},
	{"Home", "user-home-symbolic"},
	{"Documents", "folder-documents-symbolic"},
	{"Downloads", "folder-download-symbolic"},
	{"Music", "folder-music-symbolic"},
	{"Pictures", "folder-pictures-symbolic"},
	{"Videos", "folder-videos-symbolic"},
	{"Public", "folder-publicshare-symbolic"},
	{"Templates", "folder-templates-symbolic"},
	{"Desktop", "folder-desktop-symbolic"},
	{"Remote", "folder-remote-symbolic"},
	{"Disk", "drive-harddisk-symbolic"},
}

// showIconPicker presents a dialog for choosing a folder icon.
func showIconPicker(parent *gtk.Window, currentIcon string, onPicked func(icon string)) {
	dialog := gtk.NewWindow()
	dialog.SetTitle("Choose Icon")
	dialog.SetDefaultSize(360, 0)
	dialog.SetModal(true)
	dialog.SetTransientFor(parent)

	grid := gtk.NewFlowBox()
	grid.SetSelectionMode(gtk.SelectionNone)
	grid.SetMaxChildrenPerLine(6)
	grid.SetMinChildrenPerLine(4)
	grid.SetRowSpacing(6)
	grid.SetColumnSpacing(6)
	grid.SetMarginTop(12)
	grid.SetMarginBottom(12)
	grid.SetMarginStart(12)
	grid.SetMarginEnd(12)
	grid.SetHomogeneous(true)

	for _, fi := range folderIcons {
		iconName := fi.Icon

		btn := gtk.NewButton()
		btn.SetTooltipText(fi.Name)

		image := gtk.NewImageFromIconName(fi.Icon)
		image.SetPixelSize(32)
		btn.SetChild(image)

		if iconName == currentIcon {
			btn.AddCSSClass("suggested-action")
		}

		btn.ConnectClicked(func() {
			dialog.Close()
			if nil != onPicked {
				onPicked(iconName)
			}
		})

		grid.Append(btn)
	}

	header := adw.NewHeaderBar()

	toolbar := adw.NewToolbarView()
	toolbar.AddTopBar(header)
	toolbar.SetContent(grid)

	dialog.SetChild(toolbar)
	dialog.Present()
}
