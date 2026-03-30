package opsrv

import (
	"context"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/core/glib"

	"codeberg.org/reiver/go-erorr"

	"topotron/lib/backend"
	"topotron/lib/fileinfo"
)

// ProgressFunc is called on the GTK main thread to report progress.
type ProgressFunc func(current, total int, name string)

// DoneFunc is called on the GTK main thread when an operation completes.
type DoneFunc func(err error)

// DeleteFiles removes files in a background goroutine.
// Progress and completion are reported on the GTK main thread via [glib.IdleAdd].
func DeleteFiles(ctx context.Context, backend libbackend.FileBackend, entries []libfileinfo.FileInfo, onProgress ProgressFunc, onDone DoneFunc) {
	go func() {
		total := len(entries)
		for i, entry := range entries {
			select {
			case <-ctx.Done():
				glib.IdleAdd(func() { onDone(ctx.Err()) })
				return
			default:
			}

			name := entry.Name
			idx := i
			glib.IdleAdd(func() { onProgress(idx, total, name) })

			err := backend.Remove(ctx, entry.Path)
			if nil != err {
				finalErr := erorr.Wrap(err, "could not delete file")
				glib.IdleAdd(func() { onDone(finalErr) })
				return
			}
		}
		glib.IdleAdd(func() { onDone(nil) })
	}()
}

// CopyFiles copies files to a destination directory in a background goroutine.
// Progress and completion are reported on the GTK main thread via [glib.IdleAdd].
func CopyFiles(ctx context.Context, backend libbackend.FileBackend, entries []libfileinfo.FileInfo, destDir string, onProgress ProgressFunc, onDone DoneFunc) {
	go func() {
		total := len(entries)
		for i, entry := range entries {
			select {
			case <-ctx.Done():
				glib.IdleAdd(func() { onDone(ctx.Err()) })
				return
			default:
			}

			name := entry.Name
			idx := i
			glib.IdleAdd(func() { onProgress(idx, total, name) })

			dst := filepath.Join(destDir, entry.Name)
			err := backend.Copy(ctx, entry.Path, dst)
			if nil != err {
				finalErr := erorr.Wrap(err, "could not copy file")
				glib.IdleAdd(func() { onDone(finalErr) })
				return
			}
		}
		glib.IdleAdd(func() { onDone(nil) })
	}()
}

// MoveFiles moves files to a destination directory in a background goroutine.
// Progress and completion are reported on the GTK main thread via [glib.IdleAdd].
func MoveFiles(ctx context.Context, backend libbackend.FileBackend, entries []libfileinfo.FileInfo, destDir string, onProgress ProgressFunc, onDone DoneFunc) {
	go func() {
		total := len(entries)
		for i, entry := range entries {
			select {
			case <-ctx.Done():
				glib.IdleAdd(func() { onDone(ctx.Err()) })
				return
			default:
			}

			name := entry.Name
			idx := i
			glib.IdleAdd(func() { onProgress(idx, total, name) })

			dst := filepath.Join(destDir, entry.Name)
			err := backend.Rename(ctx, entry.Path, dst)
			if nil != err {
				finalErr := erorr.Wrap(err, "could not move file")
				glib.IdleAdd(func() { onDone(finalErr) })
				return
			}
		}
		glib.IdleAdd(func() { onDone(nil) })
	}()
}
