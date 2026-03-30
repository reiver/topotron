package gtksrv

import (
	"context"
	"fmt"
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

// FileBrowserPage displays the contents of a directory and supports
// selection mode for file operations.
type FileBrowserPage struct {
	// gtk widgets
	page           *adw.NavigationPage
	listBox        *gtk.ListBox
	contentStack   *gtk.Stack
	header         *adw.HeaderBar
	bottomRevealer *gtk.Revealer
	selectionLabel *gtk.Label
	pasteBtn       *gtk.Button
	newFolderBtn   *gtk.Button

	// state
	backend     libbackend.FileBackend
	path        string
	entries     []libfileinfo.FileInfo
	isSelecting bool
	selected    map[int]bool
	checks      []*gtk.CheckButton

	// callbacks
	OnDirectoryActivated func(path string)
	OnCut                func(entries []libfileinfo.FileInfo)
	OnCopy               func(entries []libfileinfo.FileInfo)
	OnPaste              func(destPath string)
	HasClipboard         func() bool
	OnRefreshNeeded      func(path string)
}

// newFileBrowserPage creates a new [FileBrowserPage] for the given path.
func newFileBrowserPage(backend libbackend.FileBackend, path string) *FileBrowserPage {
	var receiver FileBrowserPage

	receiver.backend = backend
	receiver.path = path
	receiver.selected = make(map[int]bool)

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

	// long-press gesture for entering selection mode
	longPress := gtk.NewGestureLongPress()
	longPress.ConnectPressed(func(x, y float64) {
		row := receiver.listBox.GetRowAtY(int(y))
		if nil == row {
			return
		}
		if !receiver.isSelecting {
			receiver.enterSelectionMode(row.Index())
		}
	})
	receiver.listBox.AddController(longPress)

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
	loadingPage := adw.NewStatusPage()
	loadingPage.SetTitle("Loading")

	// content stack
	receiver.contentStack = gtk.NewStack()
	receiver.contentStack.AddNamed(loadingPage, "loading")
	receiver.contentStack.AddNamed(clamp, "content")
	receiver.contentStack.AddNamed(placeholder, "placeholder")
	receiver.contentStack.SetVisibleChildName("loading")

	// bottom action bar for selection mode
	receiver.bottomRevealer = receiver.buildBottomBar()

	// header bar
	receiver.header = adw.NewHeaderBar()

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

	actionBox.Append(selectAllBtn)
	actionBox.Append(cutBtn)
	actionBox.Append(copyBtn)
	actionBox.Append(renameBtn)
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
		entries, err := receiver.backend.List(context.Background(), receiver.path, false)

		glib.IdleAdd(func() {
			if nil != err {
				log := logsrv.Begin()
				defer log.End()
				log.Highlightf("could not load directory: %s", receiver.path)

				receiver.contentStack.SetVisibleChildName("placeholder")
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
		receiver.contentStack.SetVisibleChildName("placeholder")
		return
	}

	receiver.checks = nil
	receiver.clearListBox()

	for i, entry := range receiver.entries {
		row := receiver.buildRow(i, entry)
		receiver.listBox.Append(row)
	}

	receiver.contentStack.SetVisibleChildName("content")
}

// clearListBox removes all children from the list box.
func (receiver *FileBrowserPage) clearListBox() {
	for {
		child := receiver.listBox.FirstChild()
		if nil == child {
			break
		}
		receiver.listBox.Remove(child)
	}
}

// buildRow creates an [adw.ActionRow] for a file entry, with optional check button.
func (receiver *FileBrowserPage) buildRow(index int, entry libfileinfo.FileInfo) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle(entry.Name)
	row.SetActivatable(true)

	if receiver.isSelecting {
		check := gtk.NewCheckButton()
		check.SetActive(receiver.selected[index])
		check.ConnectToggled(func() {
			if check.Active() {
				receiver.selected[index] = true
			} else {
				delete(receiver.selected, index)
			}
			receiver.updateSelectionCount()
		})
		receiver.checks = append(receiver.checks, check)
		row.AddPrefix(check)
	}

	icon := gtk.NewImageFromIconName(entry.Icon)
	row.AddPrefix(icon)

	if entry.IsDir {
		if !receiver.isSelecting {
			arrow := gtk.NewImageFromIconName("go-next-symbolic")
			row.AddSuffix(arrow)
		}
	} else {
		sizeLabel := gtk.NewLabel(libformat.Size(entry.Size))
		sizeLabel.AddCSSClass("dim-label")
		row.AddSuffix(sizeLabel)
	}

	return row
}

// Refresh reloads the directory contents.
func (receiver *FileBrowserPage) Refresh() {
	receiver.contentStack.SetVisibleChildName("loading")
	receiver.selected = make(map[int]bool)
	receiver.load()
}

// enterSelectionMode switches to selection mode, selecting the given row.
func (receiver *FileBrowserPage) enterSelectionMode(initialIndex int) {
	receiver.isSelecting = true
	receiver.selected = make(map[int]bool)
	if initialIndex >= 0 && initialIndex < len(receiver.entries) {
		receiver.selected[initialIndex] = true
	}

	receiver.newFolderBtn.SetVisible(false)
	receiver.pasteBtn.SetVisible(false)
	receiver.bottomRevealer.SetRevealChild(true)
	receiver.populate()
	receiver.updateSelectionCount()
}

// exitSelectionMode switches back to normal navigation mode.
func (receiver *FileBrowserPage) exitSelectionMode() {
	receiver.isSelecting = false
	receiver.selected = make(map[int]bool)

	receiver.newFolderBtn.SetVisible(true)
	receiver.UpdatePasteButton()
	receiver.bottomRevealer.SetRevealChild(false)
	receiver.populate()
}

// selectAll selects all entries.
func (receiver *FileBrowserPage) selectAll() {
	for i := range receiver.entries {
		receiver.selected[i] = true
	}
	for _, check := range receiver.checks {
		check.SetActive(true)
	}
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
	for i, entry := range receiver.entries {
		if receiver.selected[i] {
			result = append(result, entry)
		}
	}
	return result
}

// onRowActivated handles a tap on a row.
func (receiver *FileBrowserPage) onRowActivated(row *gtk.ListBoxRow) {
	index := row.Index()
	if index < 0 || index >= len(receiver.entries) {
		return
	}

	if receiver.isSelecting {
		// toggle selection
		if receiver.selected[index] {
			delete(receiver.selected, index)
			if index < len(receiver.checks) {
				receiver.checks[index].SetActive(false)
			}
		} else {
			receiver.selected[index] = true
			if index < len(receiver.checks) {
				receiver.checks[index].SetActive(true)
			}
		}
		receiver.updateSelectionCount()
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

	// delete directly (confirmation dialog can be added in Phase 9)
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

	// TODO: show rename dialog (Phase 9 polish)
	_ = entries[0]
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
