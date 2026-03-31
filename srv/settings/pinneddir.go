package settingsrv

// PinnedDir represents a directory pinned to the Places page.
type PinnedDir struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
