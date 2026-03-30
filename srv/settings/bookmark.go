package settingsrv

// Bookmark represents a saved WebDAV server connection.
type Bookmark struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}
