package libfileinfo

import "time"

// FileInfo holds metadata for a file or directory, regardless of whether
// it comes from a local filesystem or a WebDAV server.
type FileInfo struct {
	// identity
	Name string
	Path string

	// metadata
	Size    int64
	ModTime time.Time
	IsDir   bool

	// display
	Icon string // GTK icon name (e.g., "folder-symbolic", "text-x-generic-symbolic")
}
