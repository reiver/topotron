package gui

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/srv/settings"
)

// showAddWebDAVDialog presents a dialog for adding a new WebDAV server bookmark.
func showAddWebDAVDialog(parent *adw.ApplicationWindow, settings *settingsrv.Settings, onAdded func()) {
	dialog := gtk.NewWindow()
	dialog.SetTitle("Add WebDAV Server")
	dialog.SetDefaultSize(360, 0)
	dialog.SetModal(true)
	dialog.SetTransientFor(&parent.Window)

	nameEntry := adw.NewEntryRow()
	nameEntry.SetTitle("Name")
	nameEntry.SetText("My Server")

	urlEntry := adw.NewEntryRow()
	urlEntry.SetTitle("URL")
	urlEntry.SetText("https://")

	userEntry := adw.NewEntryRow()
	userEntry.SetTitle("Username")

	passEntry := adw.NewPasswordEntryRow()
	passEntry.SetTitle("Password")

	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionNone)
	listBox.AddCSSClass("boxed-list")
	listBox.SetMarginTop(12)
	listBox.SetMarginBottom(12)
	listBox.SetMarginStart(12)
	listBox.SetMarginEnd(12)
	listBox.Append(nameEntry)
	listBox.Append(urlEntry)
	listBox.Append(userEntry)
	listBox.Append(passEntry)

	addBtn := gtk.NewButtonWithLabel("Add")
	addBtn.AddCSSClass("suggested-action")
	addBtn.SetMarginStart(12)
	addBtn.SetMarginEnd(12)
	addBtn.SetMarginBottom(12)
	addBtn.ConnectClicked(func() {
		name := nameEntry.Text()
		url := urlEntry.Text()
		user := userEntry.Text()
		pass := passEntry.Text()

		if "" == name || "" == url {
			return
		}

		settings.AddBookmark(settingsrv.Bookmark{
			Name:     name,
			URL:      url,
			Username: user,
			Password: pass,
		})

		dialog.Close()

		if nil != onAdded {
			onAdded()
		}
	})

	cancelBtn := gtk.NewButtonWithLabel("Cancel")
	cancelBtn.SetMarginStart(12)
	cancelBtn.SetMarginEnd(12)
	cancelBtn.SetMarginBottom(12)
	cancelBtn.ConnectClicked(func() {
		dialog.Close()
	})

	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	btnBox.SetHAlign(gtk.AlignEnd)
	btnBox.SetMarginStart(12)
	btnBox.SetMarginEnd(12)
	btnBox.SetMarginBottom(12)
	btnBox.Append(cancelBtn)
	btnBox.Append(addBtn)

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
