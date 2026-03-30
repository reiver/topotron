package settingsrv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"topotron/srv/log"
)

// Settings provides persistent user preferences.
//
// Preferences are stored in a JSON file at ~/.config/topotron/settings.json.
// When GSettings is available (schema installed), it can be used instead;
// for now this implementation uses a plain JSON file for development simplicity.
type Settings struct {
	mu   sync.Mutex
	data settingsData
	path string

	// change listeners
	listeners []func()
}

type settingsData struct {
	ShowHidden bool       `json:"show-hidden"`
	SortOrder  string     `json:"sort-order"`
	Bookmarks  []Bookmark `json:"bookmarks"`
}

// New creates a [Settings] instance, loading saved preferences from disk.
func New() *Settings {
	configDir, err := os.UserConfigDir()
	if nil != err {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}

	dir := filepath.Join(configDir, "topotron")
	path := filepath.Join(dir, "settings.json")

	s := &Settings{
		path: path,
		data: settingsData{
			ShowHidden: false,
			SortOrder:  "name-asc",
		},
	}

	s.load()

	return s
}

// ShowHidden returns whether hidden files should be displayed.
func (receiver *Settings) ShowHidden() bool {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	return receiver.data.ShowHidden
}

// SetShowHidden sets whether hidden files should be displayed.
func (receiver *Settings) SetShowHidden(value bool) {
	receiver.mu.Lock()
	receiver.data.ShowHidden = value
	receiver.mu.Unlock()

	receiver.save()
	receiver.notifyListeners()
}

// SortOrder returns the current sort order string.
func (receiver *Settings) SortOrder() string {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	return receiver.data.SortOrder
}

// SetSortOrder sets the current sort order string.
func (receiver *Settings) SetSortOrder(value string) {
	receiver.mu.Lock()
	receiver.data.SortOrder = value
	receiver.mu.Unlock()

	receiver.save()
	receiver.notifyListeners()
}

// Bookmarks returns the saved WebDAV server bookmarks.
func (receiver *Settings) Bookmarks() []Bookmark {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	result := make([]Bookmark, len(receiver.data.Bookmarks))
	copy(result, receiver.data.Bookmarks)
	return result
}

// AddBookmark adds a WebDAV server bookmark and saves.
func (receiver *Settings) AddBookmark(bookmark Bookmark) {
	receiver.mu.Lock()
	receiver.data.Bookmarks = append(receiver.data.Bookmarks, bookmark)
	receiver.mu.Unlock()

	receiver.save()
	receiver.notifyListeners()
}

// RemoveBookmark removes a WebDAV server bookmark by index and saves.
func (receiver *Settings) RemoveBookmark(index int) {
	receiver.mu.Lock()
	if index >= 0 && index < len(receiver.data.Bookmarks) {
		receiver.data.Bookmarks = append(receiver.data.Bookmarks[:index], receiver.data.Bookmarks[index+1:]...)
	}
	receiver.mu.Unlock()

	receiver.save()
	receiver.notifyListeners()
}

// OnChanged registers a listener that is called when any setting changes.
func (receiver *Settings) OnChanged(fn func()) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	receiver.listeners = append(receiver.listeners, fn)
}

func (receiver *Settings) notifyListeners() {
	receiver.mu.Lock()
	listeners := make([]func(), len(receiver.listeners))
	copy(listeners, receiver.listeners)
	receiver.mu.Unlock()

	for _, fn := range listeners {
		fn()
	}
}

func (receiver *Settings) load() {
	data, err := os.ReadFile(receiver.path)
	if nil != err {
		return
	}

	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	err = json.Unmarshal(data, &receiver.data)
	if nil != err {
		log := logsrv.Begin()
		defer log.End()
		log.Highlightf("could not parse settings: %v", err)
	}
}

func (receiver *Settings) save() {
	receiver.mu.Lock()
	data, err := json.MarshalIndent(receiver.data, "", "  ")
	receiver.mu.Unlock()

	if nil != err {
		return
	}

	dir := filepath.Dir(receiver.path)
	err = os.MkdirAll(dir, 0755)
	if nil != err {
		return
	}

	err = os.WriteFile(receiver.path, data, 0644)
	if nil != err {
		log := logsrv.Begin()
		defer log.End()
		log.Highlightf("could not save settings: %v", err)
	}
}
