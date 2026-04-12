package gui

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/lib/backend"
	"topotron/lib/fileinfo"
	"topotron/lib/format"
	"topotron/srv/log"
)

// fileRowWidgets holds per-recycled-row widget references, stored in a map
// to avoid fragile child-tree traversal.
type fileRowWidgets struct {
	box       *gtk.Box
	check     *gtk.CheckButton
	icon      *gtk.Image
	nameLabel *gtk.Label
	sizeLabel *gtk.Label
	arrow     *gtk.Image

	position int // updated in Bind, read by gesture closure
}

// FileBrowserPage displays the contents of a directory and supports
// selection mode for file operations, search filtering, and sorting.
type FileBrowserPage struct {
	// gtk widgets
	page           *adw.NavigationPage
	listView       *gtk.ListView
	model          *gtk.StringList
	contentStack   *gtk.Stack
	header         *adw.HeaderBar
	bottomRevealer *gtk.Revealer
	selectionLabel *gtk.Label
	pasteBtn       *gtk.Button
	newFolderBtn   *gtk.Button
	searchBtn      *gtk.ToggleButton
	searchBar      *gtk.SearchBar
	searchEntry    *gtk.SearchEntry
	sortBtn        *gtk.MenuButton

	// state
	backend          libbackend.FileBackend
	path             string
	entries          []libfileinfo.FileInfo
	displayEntries   []libfileinfo.FileInfo
	isSelecting      bool
	selected         map[int]bool
	rowWidgets       map[uintptr]*fileRowWidgets
	longPressHandled bool
	sortOrder        SortOrder
	showHidden       bool
	searchText       string
	searchTimer      *time.Timer

	// callbacks
	OnDirectoryActivated func(path string)
	OnCut                func(entries []libfileinfo.FileInfo)
	OnCopy               func(entries []libfileinfo.FileInfo)
	OnPaste              func(destPath string)
	HasClipboard         func() bool
	OnRefreshNeeded      func(path string)
	OnSortChanged        func(order SortOrder)
	OnProperties         func(entry libfileinfo.FileInfo)
	OnRename             func(entry libfileinfo.FileInfo)
	OnPin                func(entry libfileinfo.FileInfo)
}

// newFileBrowserPage creates a new [FileBrowserPage] for the given path.
func newFileBrowserPage(backend libbackend.FileBackend, path string, sortOrder SortOrder, showHidden bool) *FileBrowserPage {
	var receiver FileBrowserPage

	receiver.backend = backend
	receiver.path = path
	receiver.selected = make(map[int]bool)
	receiver.sortOrder = sortOrder
	receiver.showHidden = showHidden

	// row widget map
	receiver.rowWidgets = make(map[uintptr]*fileRowWidgets)

	// string list model (one dummy entry per file)
	receiver.model = gtk.NewStringList(nil)

	// factory
	factory := gtk.NewSignalListItemFactory()
	factory.ConnectSetup(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)
		receiver.onSetup(listItem)
	})
	factory.ConnectBind(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)
		receiver.onBind(listItem)
	})
	factory.ConnectUnbind(func(object *glib.Object) {
		// nothing needed — no signals to disconnect; widget state is fully reset in the next Bind
	})
	factory.ConnectTeardown(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)
		delete(receiver.rowWidgets, listItem.Native())
	})

	// selection model (no selection — we manage it ourselves)
	selModel := gtk.NewNoSelection(receiver.model)

	// list view
	receiver.listView = gtk.NewListView(selModel, &factory.ListItemFactory)
	receiver.listView.AddCSSClass("boxed-list")
	receiver.listView.SetShowSeparators(true)
	receiver.listView.SetSingleClickActivate(true)
	receiver.listView.SetMarginTop(12)
	receiver.listView.SetMarginBottom(12)
	receiver.listView.SetMarginStart(12)
	receiver.listView.SetMarginEnd(12)

	receiver.listView.ConnectActivate(func(position uint) {
		receiver.onItemActivated(int(position))
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	scrolled.SetChild(receiver.listView)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(600)
	clamp.SetChild(scrolled)

	// placeholder for empty directories
	placeholder := adw.NewStatusPage()
	placeholder.SetIconName("folder-open-symbolic")
	placeholder.SetTitle("Empty Folder")

	// loading spinner
	loadingPage := adw.NewStatusPage()
	loadingPage.SetTitle("Loading")

	// content stack
	receiver.contentStack = gtk.NewStack()
	receiver.contentStack.AddNamed(loadingPage, "loading")
	receiver.contentStack.AddNamed(clamp, "content")
	receiver.contentStack.AddNamed(placeholder, "placeholder")
	receiver.contentStack.SetVisibleChildName("loading")

	// search bar
	receiver.searchEntry = gtk.NewSearchEntry()
	receiver.searchEntry.SetHExpand(true)
	receiver.searchEntry.ConnectSearchChanged(func() {
		receiver.onSearchChanged()
	})
	receiver.searchEntry.ConnectStopSearch(func() {
		receiver.searchBtn.SetActive(false)
	})

	receiver.searchBar = gtk.NewSearchBar()
	receiver.searchBar.SetChild(receiver.searchEntry)
	receiver.searchBar.ConnectEntry(receiver.searchEntry)

	// bottom action bar for selection mode
	receiver.bottomRevealer = receiver.buildBottomBar()

	// header bar
	receiver.header = adw.NewHeaderBar()

	// search toggle button
	receiver.searchBtn = gtk.NewToggleButton()
	receiver.searchBtn.SetIconName("system-search-symbolic")
	receiver.searchBtn.SetTooltipText("Search")
	receiver.searchBtn.ConnectToggled(func() {
		active := receiver.searchBtn.Active()
		receiver.searchBar.SetSearchMode(active)
		if !active {
			receiver.searchEntry.SetText("")
			receiver.searchText = ""
			receiver.applyFilterAndSort()
		}
	})
	receiver.header.PackEnd(receiver.searchBtn)

	// sort menu button
	receiver.sortBtn = receiver.buildSortMenu()
	receiver.header.PackEnd(receiver.sortBtn)

	// new folder button (visible in normal mode)
	receiver.newFolderBtn = gtk.NewButtonFromIconName("folder-new-symbolic")
	receiver.newFolderBtn.SetTooltipText("New Folder")
	receiver.newFolderBtn.ConnectClicked(func() {
		receiver.onNewFolder()
	})
	receiver.header.PackEnd(receiver.newFolderBtn)

	// paste button (visible when clipboard has content)
	receiver.pasteBtn = gtk.NewButtonFromIconName("edit-paste-symbolic")
	receiver.pasteBtn.SetTooltipText("Paste")
	receiver.pasteBtn.SetVisible(false)
	receiver.pasteBtn.ConnectClicked(func() {
		if nil != receiver.OnPaste {
			receiver.OnPaste(receiver.path)
		}
	})
	receiver.header.PackEnd(receiver.pasteBtn)

	// toolbar view
	toolbar := adw.NewToolbarView()
	toolbar.AddTopBar(receiver.header)
	toolbar.AddTopBar(receiver.searchBar)
	toolbar.SetContent(receiver.contentStack)
	toolbar.AddBottomBar(receiver.bottomRevealer)

	dirName := filepath.Base(path)
	if path == "/" {
		dirName = "Root"
	}

	receiver.page = adw.NewNavigationPage(toolbar, dirName)

	receiver.load()

	return &receiver
}

// onSetup creates the row widget tree for a recycled list item.
func (receiver *FileBrowserPage) onSetup(listItem *gtk.ListItem) {
	widgets := &fileRowWidgets{}

	// checkbox — non-interactive, purely visual
	widgets.check = gtk.NewCheckButton()
	widgets.check.SetCanTarget(false)
	widgets.check.SetVisible(false)

	// file/folder icon
	widgets.icon = gtk.NewImageFromIconName("text-x-generic-symbolic")
	widgets.icon.SetPixelSize(16)

	// file name label
	widgets.nameLabel = gtk.NewLabel("")
	widgets.nameLabel.SetHExpand(true)
	widgets.nameLabel.SetXAlign(0)
	widgets.nameLabel.SetEllipsize(pango.EllipsizeEnd)

	// file size label (hidden for directories)
	widgets.sizeLabel = gtk.NewLabel("")
	widgets.sizeLabel.AddCSSClass("dim-label")

	// directory arrow (hidden for files)
	widgets.arrow = gtk.NewImageFromIconName("go-next-symbolic")

	// row box
	widgets.box = gtk.NewBox(gtk.OrientationHorizontal, 6)
	widgets.box.SetSizeRequest(-1, 48)
	widgets.box.SetMarginStart(12)
	widgets.box.SetMarginEnd(12)
	widgets.box.SetMarginTop(6)
	widgets.box.SetMarginBottom(6)
	widgets.box.Append(widgets.check)
	widgets.box.Append(widgets.icon)
	widgets.box.Append(widgets.nameLabel)
	widgets.box.Append(widgets.sizeLabel)
	widgets.box.Append(widgets.arrow)

	// long-press gesture — enters selection mode
	widgets.position = -1 // no valid position until first Bind
	longPress := gtk.NewGestureLongPress()
	longPress.ConnectPressed(func(x, y float64) {
		if receiver.isSelecting {
			return // already in selection mode — don't re-enter (would clear selections)
		}
		receiver.longPressHandled = true
		receiver.enterSelectionMode(widgets.position)
	})
	widgets.box.AddController(longPress)

	listItem.SetChild(widgets.box)
	receiver.rowWidgets[listItem.Native()] = widgets
}

// onBind populates a recycled row with data from displayEntries.
func (receiver *FileBrowserPage) onBind(listItem *gtk.ListItem) {
	widgets := receiver.rowWidgets[listItem.Native()]
	pos := int(listItem.Position())
	widgets.position = pos

	entry := receiver.displayEntries[pos]

	// icon and name
	widgets.icon.SetFromIconName(entry.Icon)
	widgets.nameLabel.SetLabel(entry.Name)

	// size label vs arrow (mutually exclusive)
	if entry.IsDir {
		widgets.sizeLabel.SetVisible(false)
		widgets.arrow.SetVisible(!receiver.isSelecting)
	} else {
		widgets.sizeLabel.SetLabel(libformat.Size(entry.Size))
		widgets.sizeLabel.SetVisible(true)
		widgets.arrow.SetVisible(false)
	}

	// checkbox — no signal blocking needed (checkbox is non-interactive)
	widgets.check.SetVisible(receiver.isSelecting)
	widgets.check.SetActive(receiver.selected[pos])
}

// buildSortMenu creates the sort order menu button.
func (receiver *FileBrowserPage) buildSortMenu() *gtk.MenuButton {
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionNone)
	listBox.AddCSSClass("boxed-list")

	for _, order := range AllSortOrders() {
		row := adw.NewActionRow()
		row.SetTitle(SortLabel(order))
		row.SetActivatable(true)

		if order == receiver.sortOrder {
			check := gtk.NewImageFromIconName("object-select-symbolic")
			row.AddSuffix(check)
		}

		listBox.Append(row)
	}

	popover := gtk.NewPopover()
	popover.SetChild(listBox)

	listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		index := row.Index()
		orders := AllSortOrders()
		if index >= 0 && index < len(orders) {
			receiver.sortOrder = orders[index]
			receiver.applyFilterAndSort()
			popover.Popdown()
			if nil != receiver.OnSortChanged {
				receiver.OnSortChanged(receiver.sortOrder)
			}

			// rebuild menu to update check mark
			receiver.rebuildSortMenu()
		}
	})

	btn := gtk.NewMenuButton()
	btn.SetIconName("view-sort-ascending-symbolic")
	btn.SetTooltipText("Sort")
	btn.SetPopover(popover)

	return btn
}

// rebuildSortMenu rebuilds the sort menu to reflect the current sort order.
func (receiver *FileBrowserPage) rebuildSortMenu() {
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionNone)
	listBox.AddCSSClass("boxed-list")

	for _, order := range AllSortOrders() {
		row := adw.NewActionRow()
		row.SetTitle(SortLabel(order))
		row.SetActivatable(true)

		if order == receiver.sortOrder {
			check := gtk.NewImageFromIconName("object-select-symbolic")
			row.AddSuffix(check)
		}

		listBox.Append(row)
	}

	popover := receiver.sortBtn.Popover()
	popover.SetChild(listBox)

	listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		index := row.Index()
		orders := AllSortOrders()
		if index >= 0 && index < len(orders) {
			receiver.sortOrder = orders[index]
			receiver.applyFilterAndSort()
			popover.Popdown()
			if nil != receiver.OnSortChanged {
				receiver.OnSortChanged(receiver.sortOrder)
			}
			receiver.rebuildSortMenu()
		}
	})
}

// buildBottomBar creates the selection mode action bar.
func (receiver *FileBrowserPage) buildBottomBar() *gtk.Revealer {
	receiver.selectionLabel = gtk.NewLabel("0 selected")

	cancelBtn := gtk.NewButtonWithLabel("Cancel")
	cancelBtn.ConnectClicked(func() {
		receiver.exitSelectionMode()
	})

	selectAllBtn := gtk.NewButtonFromIconName("edit-select-all-symbolic")
	selectAllBtn.SetTooltipText("Select All")
	selectAllBtn.ConnectClicked(func() {
		receiver.selectAll()
	})

	cutBtn := gtk.NewButtonFromIconName("edit-cut-symbolic")
	cutBtn.SetTooltipText("Cut")
	cutBtn.ConnectClicked(func() {
		receiver.onCut()
	})

	copyBtn := gtk.NewButtonFromIconName("edit-copy-symbolic")
	copyBtn.SetTooltipText("Copy")
	copyBtn.ConnectClicked(func() {
		receiver.onCopy()
	})

	deleteBtn := gtk.NewButtonFromIconName("user-trash-symbolic")
	deleteBtn.SetTooltipText("Delete")
	deleteBtn.AddCSSClass("destructive-action")
	deleteBtn.ConnectClicked(func() {
		receiver.onDelete()
	})

	renameBtn := gtk.NewButtonFromIconName("document-edit-symbolic")
	renameBtn.SetTooltipText("Rename")
	renameBtn.ConnectClicked(func() {
		receiver.onRename()
	})

	// layout: label | spacer | actions | cancel
	actionBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	actionBox.SetMarginTop(6)
	actionBox.SetMarginBottom(6)
	actionBox.SetMarginStart(12)
	actionBox.SetMarginEnd(12)
	actionBox.SetHAlign(gtk.AlignCenter)

	actionBox.Append(receiver.selectionLabel)

	spacer := gtk.NewBox(gtk.OrientationHorizontal, 0)
	spacer.SetHExpand(true)
	actionBox.Append(spacer)

	propertiesBtn := gtk.NewButtonFromIconName("dialog-information-symbolic")
	propertiesBtn.SetTooltipText("Properties")
	propertiesBtn.ConnectClicked(func() {
		receiver.onProperties()
	})

	pinBtn := gtk.NewButtonFromIconName("view-pin-symbolic")
	pinBtn.SetTooltipText("Pin to Places")
	pinBtn.ConnectClicked(func() {
		receiver.onPin()
	})

	actionBox.Append(selectAllBtn)
	actionBox.Append(cutBtn)
	actionBox.Append(copyBtn)
	actionBox.Append(renameBtn)
	actionBox.Append(pinBtn)
	actionBox.Append(propertiesBtn)
	actionBox.Append(deleteBtn)

	spacer2 := gtk.NewBox(gtk.OrientationHorizontal, 0)
	spacer2.SetHExpand(true)
	actionBox.Append(spacer2)

	actionBox.Append(cancelBtn)

	revealer := gtk.NewRevealer()
	revealer.SetChild(actionBox)
	revealer.SetRevealChild(false)
	revealer.SetTransitionType(gtk.RevealerTransitionTypeSlideUp)

	return revealer
}

// UpdatePasteButton shows or hides the paste button based on clipboard state.
func (receiver *FileBrowserPage) UpdatePasteButton() {
	if nil != receiver.HasClipboard {
		receiver.pasteBtn.SetVisible(receiver.HasClipboard())
	}
}

// load reads the directory contents in a goroutine and populates the list.
func (receiver *FileBrowserPage) load() {
	go func() {
		entries, err := receiver.backend.List(context.Background(), receiver.path, receiver.showHidden)

		glib.IdleAdd(func() {
			if nil != err {
				log := logsrv.Begin()
				defer log.End()
				log.Highlightf("could not load directory: %s", receiver.path)

				receiver.contentStack.SetVisibleChildName("placeholder")
				return
			}

			receiver.entries = entries
			receiver.applyFilterAndSort()
		})
	}()
}

// applyFilterAndSort filters and sorts the entries, then rebuilds the model.
func (receiver *FileBrowserPage) applyFilterAndSort() {
	filtered := FilterEntries(receiver.entries, receiver.searchText)

	display := make([]libfileinfo.FileInfo, len(filtered))
	copy(display, filtered)
	SortEntries(display, receiver.sortOrder)

	receiver.displayEntries = display
	receiver.selected = make(map[int]bool) // clear selection — indices are no longer valid
	receiver.rebuildModel()
}

// rebuildModel replaces the StringList contents to match displayEntries.
func (receiver *FileBrowserPage) rebuildModel() {
	if 0 == len(receiver.displayEntries) {
		receiver.model.Splice(0, receiver.model.NItems(), nil)
		receiver.contentStack.SetVisibleChildName("placeholder")
		return
	}

	receiver.model.Splice(0, receiver.model.NItems(), dummyStrings(len(receiver.displayEntries)))
	receiver.contentStack.SetVisibleChildName("content")
}

// refreshVisibleRows updates checkbox/arrow visibility on all currently
// recycled rows without touching the model. Used when only presentation
// changes (entering/exiting selection mode).
func (receiver *FileBrowserPage) refreshVisibleRows() {
	for _, widgets := range receiver.rowWidgets {
		pos := widgets.position
		if pos < 0 || pos >= len(receiver.displayEntries) {
			continue
		}
		entry := receiver.displayEntries[pos]

		// checkbox visibility and state (no signal blocking needed — checkbox is non-interactive)
		widgets.check.SetVisible(receiver.isSelecting)
		widgets.check.SetActive(receiver.selected[pos])

		// arrow visibility: shown for dirs when NOT selecting
		if entry.IsDir {
			widgets.arrow.SetVisible(!receiver.isSelecting)
		}
	}
}

// updateCheckboxAt updates only the tapped row's checkbox during selection toggle.
func (receiver *FileBrowserPage) updateCheckboxAt(index int) {
	for _, widgets := range receiver.rowWidgets {
		if widgets.position == index {
			widgets.check.SetActive(receiver.selected[index])
			return
		}
	}
}

// dummyStrings returns a slice of n empty strings for use as StringList entries.
func dummyStrings(n int) []string {
	result := make([]string, n)
	return result
}

// Refresh reloads the directory contents.
func (receiver *FileBrowserPage) Refresh() {
	if receiver.isSelecting {
		receiver.exitSelectionMode()
	}
	receiver.contentStack.SetVisibleChildName("loading")
	receiver.load()
}

// onSearchChanged handles search text changes with a 500ms debounce.
func (receiver *FileBrowserPage) onSearchChanged() {
	if nil != receiver.searchTimer {
		receiver.searchTimer.Stop()
	}

	receiver.searchTimer = time.AfterFunc(500*time.Millisecond, func() {
		glib.IdleAdd(func() {
			receiver.searchText = receiver.searchEntry.Text()
			receiver.applyFilterAndSort()
		})
	})
}

// enterSelectionMode switches to selection mode, selecting the given row.
func (receiver *FileBrowserPage) enterSelectionMode(initialIndex int) {
	receiver.isSelecting = true
	receiver.selected = make(map[int]bool)
	if initialIndex >= 0 && initialIndex < len(receiver.displayEntries) {
		receiver.selected[initialIndex] = true
	}

	receiver.newFolderBtn.SetVisible(false)
	receiver.pasteBtn.SetVisible(false)
	receiver.searchBtn.SetVisible(false)
	receiver.sortBtn.SetVisible(false)
	receiver.bottomRevealer.SetRevealChild(true)

	receiver.refreshVisibleRows()
	receiver.updateSelectionCount()
}

// exitSelectionMode switches back to normal navigation mode.
func (receiver *FileBrowserPage) exitSelectionMode() {
	receiver.isSelecting = false
	receiver.selected = make(map[int]bool)

	receiver.newFolderBtn.SetVisible(true)
	receiver.searchBtn.SetVisible(true)
	receiver.sortBtn.SetVisible(true)
	receiver.UpdatePasteButton()
	receiver.bottomRevealer.SetRevealChild(false)

	receiver.refreshVisibleRows()
}

// selectAll selects all display entries.
func (receiver *FileBrowserPage) selectAll() {
	if !receiver.isSelecting {
		return
	}
	for i := range receiver.displayEntries {
		receiver.selected[i] = true
	}
	receiver.refreshVisibleRows()
	receiver.updateSelectionCount()
}

// updateSelectionCount updates the selection label text.
func (receiver *FileBrowserPage) updateSelectionCount() {
	count := len(receiver.selected)
	if 1 == count {
		receiver.selectionLabel.SetLabel("1 selected")
	} else {
		receiver.selectionLabel.SetLabel(fmt.Sprintf("%d selected", count))
	}
}

// selectedEntries returns the currently selected [libfileinfo.FileInfo] entries.
func (receiver *FileBrowserPage) selectedEntries() []libfileinfo.FileInfo {
	var result []libfileinfo.FileInfo
	for i, entry := range receiver.displayEntries {
		if receiver.selected[i] {
			result = append(result, entry)
		}
	}
	return result
}

// onItemActivated handles a tap on a list item.
func (receiver *FileBrowserPage) onItemActivated(index int) {
	// guard against long-press + activate double-fire
	if receiver.longPressHandled {
		receiver.longPressHandled = false
		return
	}

	if index < 0 || index >= len(receiver.displayEntries) {
		return
	}

	if receiver.isSelecting {
		if receiver.selected[index] {
			delete(receiver.selected, index)
		} else {
			receiver.selected[index] = true
		}
		receiver.updateCheckboxAt(index)
		receiver.updateSelectionCount()
		return
	}

	entry := receiver.displayEntries[index]

	if entry.IsDir {
		if nil != receiver.OnDirectoryActivated {
			receiver.OnDirectoryActivated(entry.Path)
		}
		return
	}

	receiver.openFile(entry.Path)
}

// onCut stores selected entries for cutting.
func (receiver *FileBrowserPage) onCut() {
	entries := receiver.selectedEntries()
	if 0 == len(entries) {
		return
	}
	if nil != receiver.OnCut {
		receiver.OnCut(entries)
	}
	receiver.exitSelectionMode()
}

// onCopy stores selected entries for copying.
func (receiver *FileBrowserPage) onCopy() {
	entries := receiver.selectedEntries()
	if 0 == len(entries) {
		return
	}
	if nil != receiver.OnCopy {
		receiver.OnCopy(entries)
	}
	receiver.exitSelectionMode()
}

// onDelete confirms and deletes selected entries.
func (receiver *FileBrowserPage) onDelete() {
	entries := receiver.selectedEntries()
	if 0 == len(entries) {
		return
	}

	log := logsrv.Begin()
	defer log.End()

	ctx := context.Background()
	for _, entry := range entries {
		err := receiver.backend.Remove(ctx, entry.Path)
		if nil != err {
			log.Highlightf("could not delete: %s", entry.Path)
		}
	}

	receiver.exitSelectionMode()
	receiver.Refresh()
}

// onRename renames the single selected entry.
func (receiver *FileBrowserPage) onRename() {
	entries := receiver.selectedEntries()
	if 1 != len(entries) {
		return
	}

	if nil != receiver.OnRename {
		receiver.OnRename(entries[0])
	}
	receiver.exitSelectionMode()
}

// onProperties shows properties for the single selected entry.
func (receiver *FileBrowserPage) onProperties() {
	entries := receiver.selectedEntries()
	if 1 != len(entries) {
		return
	}

	if nil != receiver.OnProperties {
		receiver.OnProperties(entries[0])
	}
}

// onPin pins the single selected directory to the Places page.
func (receiver *FileBrowserPage) onPin() {
	entries := receiver.selectedEntries()
	if 1 != len(entries) {
		return
	}
	if !entries[0].IsDir {
		return
	}
	if nil != receiver.OnPin {
		receiver.OnPin(entries[0])
	}
	receiver.exitSelectionMode()
}

// onNewFolder creates a new folder in the current directory.
func (receiver *FileBrowserPage) onNewFolder() {
	log := logsrv.Begin()
	defer log.End()

	name := "New Folder"
	path := filepath.Join(receiver.path, name)

	// find a unique name
	for i := 2; ; i++ {
		err := receiver.backend.Mkdir(context.Background(), path)
		if nil == err {
			break
		}

		name = fmt.Sprintf("New Folder (%d)", i)
		path = filepath.Join(receiver.path, name)

		if i > 100 {
			log.Highlightf("could not create new folder in: %s", receiver.path)
			return
		}
	}

	receiver.Refresh()
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
