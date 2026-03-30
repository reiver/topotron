package libbackend

import (
	"context"
	"io"

	"topotron/lib/fileinfo"
)

// FileBackend provides a uniform interface for file operations across
// local and remote filesystems.
//
// Both [LocalBackend] and WebDAVBackend implement this interface,
// allowing the UI layer to operate without knowing which backend is active.
//
// All methods accept a [context.Context] for cancellation support.
type FileBackend interface {
	// List returns the contents of a directory.
	List(ctx context.Context, path string, showHidden bool) ([]libfileinfo.FileInfo, error)

	// Stat returns metadata for a single file or directory.
	Stat(ctx context.Context, path string) (libfileinfo.FileInfo, error)

	// Open returns a reader for the file contents.
	Open(ctx context.Context, path string) (io.ReadCloser, error)

	// Create writes content to a new or existing file.
	Create(ctx context.Context, path string, r io.Reader) error

	// Mkdir creates a directory.
	Mkdir(ctx context.Context, path string) error

	// Remove deletes a file or directory.
	Remove(ctx context.Context, path string) error

	// Rename moves/renames within the same backend.
	Rename(ctx context.Context, oldPath, newPath string) error

	// Copy copies within the same backend.
	Copy(ctx context.Context, srcPath, dstPath string) error
}
