package gtksrv

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/lib/place"
	"topotron/srv/discover"
	"topotron/srv/settings"
)

// PlacesPage is the home screen showing browsable locations.
type PlacesPage struct {
	// gtk widgets
	page            *adw.NavigationPage
	contentBox      *gtk.Box
	localGroup      *adw.PreferencesGroup
	localListBox    *gtk.ListBox
	networkGroup    *adw.PreferencesGroup
	networkListBox  *gtk.ListBox
	webdavGroup     *adw.PreferencesGroup
	unpinRevealer   *gtk.Revealer
	unpinLabel      *gtk.Label
	unpinPath       string

	// state
	settings         *settingsrv.Settings
	localPlaces      []libplace.Place
	networkServices  []discoversrv.DiscoveredService

	// callbacks
	OnActivated        func(place libplace.Place)
	OnNetworkActivated func(service discoversrv.DiscoveredService)
	OnWebDAVActivated  func(bookmark settingsrv.Bookmark)
	OnAddWebDAV        func()
	OnAbout            func()
	OnUnpin            func(path string)
}

// newPlacesPage creates a new [PlacesPage] populated with the default places,
// discovered network services, and saved WebDAV bookmarks.
func newPlacesPage(settings *settingsrv.Settings) *PlacesPage {
	var receiver PlacesPage

	receiver.settings = settings

	// local places (from pinned dirs in settings)
	receiver.localListBox = gtk.NewListBox()
	receiver.localListBox.SetSelectionMode(gtk.SelectionNone)
	receiver.localListBox.AddCSSClass("boxed-list")

	receiver.localGroup = adw.NewPreferencesGroup()
	receiver.localGroup.SetTitle("Local")
	receiver.localGroup.Add(receiver.localListBox)

	// unpin revealer (bottom popup)
	receiver.unpinRevealer = receiver.buildUnpinBar()

	receiver.buildLocalList()

	// network section (auto-discovered services)
	receiver.networkListBox = gtk.NewListBox()
	receiver.networkListBox.SetSelectionMode(gtk.SelectionNone)
	receiver.networkListBox.AddCSSClass("boxed-list")

	receiver.networkListBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		if nil == receiver.OnNetworkActivated {
			return
		}

		index := row.Index()
		if index < 0 || index >= len(receiver.networkServices) {
			return
		}

		receiver.OnNetworkActivated(receiver.networkServices[index])
	})

	receiver.networkGroup = adw.NewPreferencesGroup()
	receiver.networkGroup.SetTitle("Network")
	receiver.networkGroup.Add(receiver.networkListBox)
	receiver.networkGroup.SetVisible(false)

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
	receiver.contentBox.Append(receiver.localGroup)
	receiver.contentBox.Append(receiver.networkGroup)
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
	menuBtn := buildMainMenu(settings, func() {
		if nil != receiver.OnAbout {
			receiver.OnAbout()
		}
	})
	header.PackEnd(menuBtn)

	toolbar := adw.NewToolbarView()
	toolbar.AddTopBar(header)
	toolbar.SetContent(clamp)
	toolbar.AddBottomBar(receiver.unpinRevealer)

	receiver.page = adw.NewNavigationPage(toolbar, "Topotron")

	return &receiver
}

// AddNetworkService adds a discovered service to the Network section.
// Called on the GTK main thread.
func (receiver *PlacesPage) AddNetworkService(service discoversrv.DiscoveredService) {
	receiver.networkServices = append(receiver.networkServices, service)

	row := adw.NewActionRow()
	row.SetTitle(service.Name)
	row.SetSubtitle(service.Host)
	row.SetActivatable(true)

	icon := gtk.NewImageFromIconName("network-workgroup-symbolic")
	row.AddPrefix(icon)

	arrow := gtk.NewImageFromIconName("go-next-symbolic")
	row.AddSuffix(arrow)

	receiver.networkListBox.Append(row)
	receiver.networkGroup.SetVisible(true)
}

// RemoveNetworkService removes a discovered service by name from the Network section.
// Called on the GTK main thread.
func (receiver *PlacesPage) RemoveNetworkService(name string) {
	// find and remove from slice
	found := false
	var updated []discoversrv.DiscoveredService
	for _, svc := range receiver.networkServices {
		if svc.Name == name {
			found = true
			continue
		}
		updated = append(updated, svc)
	}

	if !found {
		return
	}

	receiver.networkServices = updated
	receiver.rebuildNetworkList()
}

// rebuildNetworkList clears and repopulates the network list box.
func (receiver *PlacesPage) rebuildNetworkList() {
	// clear existing rows
	for {
		child := receiver.networkListBox.FirstChild()
		if nil == child {
			break
		}
		receiver.networkListBox.Remove(child)
	}

	// repopulate
	for _, svc := range receiver.networkServices {
		row := adw.NewActionRow()
		row.SetTitle(svc.Name)
		row.SetSubtitle(svc.Host)
		row.SetActivatable(true)

		icon := gtk.NewImageFromIconName("network-workgroup-symbolic")
		row.AddPrefix(icon)

		arrow := gtk.NewImageFromIconName("go-next-symbolic")
		row.AddSuffix(arrow)

		receiver.networkListBox.Append(row)
	}

	receiver.networkGroup.SetVisible(len(receiver.networkServices) > 0)
}

// RebuildLocalList refreshes the local pinned directories section.
func (receiver *PlacesPage) RebuildLocalList() {
	receiver.buildLocalList()
}

// buildLocalList populates the local section from pinned dirs in settings.
func (receiver *PlacesPage) buildLocalList() {
	// clear existing rows
	for {
		child := receiver.localListBox.FirstChild()
		if nil == child {
			break
		}
		receiver.localListBox.Remove(child)
	}

	// read pinned dirs from settings
	pinnedDirs := receiver.settings.PinnedDirs()
	receiver.localPlaces = nil

	for _, dir := range pinnedDirs {
		place := libplace.Place{
			Name: dir.Name,
			Path: dir.Path,
			Icon: libplace.IconForPath(dir.Path),
		}
		receiver.localPlaces = append(receiver.localPlaces, place)

		row := newPlaceRow(place)

		// long-press to unpin
		dirPath := dir.Path
		dirName := dir.Name
		longPress := gtk.NewGestureLongPress()
		longPress.ConnectPressed(func(x, y float64) {
			receiver.showUnpinBar(dirName, dirPath)
		})
		row.AddController(longPress)

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
}

// buildUnpinBar creates the unpin confirmation bar.
func (receiver *PlacesPage) buildUnpinBar() *gtk.Revealer {
	receiver.unpinLabel = gtk.NewLabel("")

	unpinBtn := gtk.NewButtonWithLabel("Unpin")
	unpinBtn.AddCSSClass("destructive-action")
	unpinBtn.ConnectClicked(func() {
		if nil != receiver.OnUnpin && "" != receiver.unpinPath {
			receiver.OnUnpin(receiver.unpinPath)
		}
		receiver.unpinRevealer.SetRevealChild(false)
	})

	cancelBtn := gtk.NewButtonWithLabel("Cancel")
	cancelBtn.ConnectClicked(func() {
		receiver.unpinRevealer.SetRevealChild(false)
	})

	box := gtk.NewBox(gtk.OrientationHorizontal, 6)
	box.SetMarginTop(6)
	box.SetMarginBottom(6)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)
	box.SetHAlign(gtk.AlignCenter)

	box.Append(receiver.unpinLabel)

	spacer := gtk.NewBox(gtk.OrientationHorizontal, 0)
	spacer.SetHExpand(true)
	box.Append(spacer)

	box.Append(cancelBtn)
	box.Append(unpinBtn)

	revealer := gtk.NewRevealer()
	revealer.SetChild(box)
	revealer.SetRevealChild(false)
	revealer.SetTransitionType(gtk.RevealerTransitionTypeSlideUp)

	return revealer
}

// showUnpinBar shows the unpin confirmation bar for a directory.
func (receiver *PlacesPage) showUnpinBar(name, path string) {
	receiver.unpinPath = path
	receiver.unpinLabel.SetLabel("Unpin \"" + name + "\"?")
	receiver.unpinRevealer.SetRevealChild(true)
}

// RebuildWebDAVList refreshes the WebDAV bookmarks section.
func (receiver *PlacesPage) RebuildWebDAVList() {
	receiver.buildWebDAVList()
}

// buildWebDAVList populates the WebDAV section with saved bookmarks.
func (receiver *PlacesPage) buildWebDAVList() {
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

// buildMainMenu creates the app menu with preferences toggles and about.
func buildMainMenu(settings *settingsrv.Settings, onAbout func()) *gtk.MenuButton {
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

	aboutRow := adw.NewActionRow()
	aboutRow.SetTitle("About Topotron")
	aboutRow.SetActivatable(true)

	aboutIcon := gtk.NewImageFromIconName("help-about-symbolic")
	aboutRow.AddPrefix(aboutIcon)

	menuList := gtk.NewListBox()
	menuList.SetSelectionMode(gtk.SelectionNone)
	menuList.AddCSSClass("boxed-list")
	menuList.SetMarginTop(6)
	menuList.SetMarginBottom(6)
	menuList.SetMarginStart(6)
	menuList.SetMarginEnd(6)
	menuList.Append(hiddenRow)
	menuList.Append(aboutRow)

	popover := gtk.NewPopover()
	popover.SetChild(menuList)

	menuList.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		if 1 == row.Index() && nil != onAbout {
			popover.Popdown()
			onAbout()
		}
	})

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
