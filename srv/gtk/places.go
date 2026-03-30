package gtksrv

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/lib/place"
	"topotron/srv/settings"
)

// PlacesPage is the home screen showing browsable locations.
type PlacesPage struct {
	// gtk widgets
	page         *adw.NavigationPage
	contentBox   *gtk.Box
	localListBox *gtk.ListBox
	webdavGroup  *adw.PreferencesGroup

	// state
	settings   *settingsrv.Settings
	localPlaces []libplace.Place

	// callbacks
	OnActivated       func(place libplace.Place)
	OnWebDAVActivated func(bookmark settingsrv.Bookmark)
	OnAddWebDAV       func()
}

// newPlacesPage creates a new [PlacesPage] populated with the default places
// and any saved WebDAV bookmarks.
func newPlacesPage(settings *settingsrv.Settings) *PlacesPage {
	var receiver PlacesPage

	receiver.settings = settings

	// local places
	receiver.localPlaces = libplace.DefaultPlaces()

	receiver.localListBox = gtk.NewListBox()
	receiver.localListBox.SetSelectionMode(gtk.SelectionNone)
	receiver.localListBox.AddCSSClass("boxed-list")

	for _, p := range receiver.localPlaces {
		row := newPlaceRow(p)
		receiver.localListBox.Append(row)
	}

	receiver.localListBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		if nil == receiver.OnActivated {
			return
		}

		index := row.Index()
		if index < 0 || index >= len(receiver.localPlaces) {
			return
		}

		receiver.OnActivated(receiver.localPlaces[index])
	})

	localGroup := adw.NewPreferencesGroup()
	localGroup.SetTitle("Local")
	localGroup.Add(receiver.localListBox)

	// webdav section
	receiver.webdavGroup = adw.NewPreferencesGroup()
	receiver.webdavGroup.SetTitle("WebDAV")
	receiver.buildWebDAVList()

	// add webdav button
	addBtn := gtk.NewButtonWithLabel("Add WebDAV Server")
	addBtn.AddCSSClass("pill")
	addBtn.SetMarginTop(12)
	addBtn.SetHAlign(gtk.AlignCenter)
	addBtn.ConnectClicked(func() {
		if nil != receiver.OnAddWebDAV {
			receiver.OnAddWebDAV()
		}
	})

	// layout
	receiver.contentBox = gtk.NewBox(gtk.OrientationVertical, 12)
	receiver.contentBox.SetMarginTop(12)
	receiver.contentBox.SetMarginBottom(12)
	receiver.contentBox.SetMarginStart(12)
	receiver.contentBox.SetMarginEnd(12)
	receiver.contentBox.Append(localGroup)
	receiver.contentBox.Append(receiver.webdavGroup)
	receiver.contentBox.Append(addBtn)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	scrolled.SetChild(receiver.contentBox)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(600)
	clamp.SetChild(scrolled)

	header := adw.NewHeaderBar()

	// menu button with preferences
	menuBtn := buildMainMenu(settings)
	header.PackEnd(menuBtn)

	toolbar := adw.NewToolbarView()
	toolbar.AddTopBar(header)
	toolbar.SetContent(clamp)

	receiver.page = adw.NewNavigationPage(toolbar, "Topotron")

	return &receiver
}

// RebuildWebDAVList refreshes the WebDAV bookmarks section.
func (receiver *PlacesPage) RebuildWebDAVList() {
	receiver.buildWebDAVList()
}

// buildWebDAVList populates the WebDAV section with saved bookmarks.
func (receiver *PlacesPage) buildWebDAVList() {
	// clear existing children from the preferences group
	// AdwPreferencesGroup doesn't have RemoveAll, so we rebuild with a new ListBox
	webdavListBox := gtk.NewListBox()
	webdavListBox.SetSelectionMode(gtk.SelectionNone)
	webdavListBox.AddCSSClass("boxed-list")

	bookmarks := receiver.settings.Bookmarks()

	for _, bm := range bookmarks {
		row := adw.NewActionRow()
		row.SetTitle(bm.Name)
		row.SetSubtitle(bm.URL)
		row.SetActivatable(true)

		icon := gtk.NewImageFromIconName("network-server-symbolic")
		row.AddPrefix(icon)

		arrow := gtk.NewImageFromIconName("go-next-symbolic")
		row.AddSuffix(arrow)

		webdavListBox.Append(row)
	}

	webdavListBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		if nil == receiver.OnWebDAVActivated {
			return
		}

		index := row.Index()
		if index < 0 || index >= len(bookmarks) {
			return
		}

		receiver.OnWebDAVActivated(bookmarks[index])
	})

	if 0 == len(bookmarks) {
		receiver.webdavGroup.SetVisible(false)
	} else {
		receiver.webdavGroup.SetVisible(true)
	}

	receiver.webdavGroup.Add(webdavListBox)
}

// buildMainMenu creates the app menu with preferences toggles.
func buildMainMenu(settings *settingsrv.Settings) *gtk.MenuButton {
	hiddenCheck := gtk.NewCheckButton()
	hiddenCheck.SetActive(settings.ShowHidden())
	hiddenCheck.ConnectToggled(func() {
		settings.SetShowHidden(hiddenCheck.Active())
	})

	hiddenRow := adw.NewActionRow()
	hiddenRow.SetTitle("Show Hidden Files")
	hiddenRow.SetActivatable(true)
	hiddenRow.AddSuffix(hiddenCheck)
	hiddenRow.SetActivatableWidget(hiddenCheck)

	menuList := gtk.NewListBox()
	menuList.SetSelectionMode(gtk.SelectionNone)
	menuList.AddCSSClass("boxed-list")
	menuList.SetMarginTop(6)
	menuList.SetMarginBottom(6)
	menuList.SetMarginStart(6)
	menuList.SetMarginEnd(6)
	menuList.Append(hiddenRow)

	popover := gtk.NewPopover()
	popover.SetChild(menuList)

	btn := gtk.NewMenuButton()
	btn.SetIconName("open-menu-symbolic")
	btn.SetTooltipText("Menu")
	btn.SetPopover(popover)

	return btn
}

// newPlaceRow creates an [adw.ActionRow] for a single [libplace.Place].
func newPlaceRow(place libplace.Place) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle(place.Name)
	row.SetActivatable(true)

	icon := gtk.NewImageFromIconName(place.Icon)
	row.AddPrefix(icon)

	arrow := gtk.NewImageFromIconName("go-next-symbolic")
	row.AddSuffix(arrow)

	return row
}
