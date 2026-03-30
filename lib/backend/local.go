package libbackend

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/reiver/go-erorr"

	"topotron/lib/fileinfo"
	"topotron/lib/icon"
)

// LocalBackend implements [FileBackend] for the local filesystem.
type LocalBackend struct{}

var _ FileBackend = LocalBackend{}

// List returns the contents of a local directory.
func (receiver LocalBackend) List(ctx context.Context, path string, showHidden bool) ([]libfileinfo.FileInfo, error) {
	entries, err := os.ReadDir(path)
	if nil != err {
		return nil, erorr.Wrap(err, "could not read directory")
	}

	var result []libfileinfo.FileInfo

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		name := entry.Name()

		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		info, err := entry.Info()
		if nil != err {
			continue
		}

		fullPath := filepath.Join(path, name)

		result = append(result, libfileinfo.FileInfo{
			Name:    name,
			Path:    fullPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
			Icon:    libicon.ForFile(name, entry.IsDir()),
		})
	}

	return result, nil
}

// Stat returns metadata for a local file or directory.
func (receiver LocalBackend) Stat(ctx context.Context, path string) (libfileinfo.FileInfo, error) {
	info, err := os.Stat(path)
	if nil != err {
		return libfileinfo.FileInfo{}, erorr.Wrap(err, "could not stat file")
	}

	return libfileinfo.FileInfo{
		Name:    info.Name(),
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
		Icon:    libicon.ForFile(info.Name(), info.IsDir()),
	}, nil
}

// Open returns a reader for a local file.
func (receiver LocalBackend) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if nil != err {
		return nil, erorr.Wrap(err, "could not open file")
	}
	return f, nil
}

// Create writes content to a local file.
func (receiver LocalBackend) Create(ctx context.Context, path string, r io.Reader) error {
	f, err := os.Create(path)
	if nil != err {
		return erorr.Wrap(err, "could not create file")
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if nil != err {
		return erorr.Wrap(err, "could not write file")
	}

	return nil
}

// Mkdir creates a local directory and any necessary parents.
func (receiver LocalBackend) Mkdir(ctx context.Context, path string) error {
	err := os.MkdirAll(path, 0755)
	if nil != err {
		return erorr.Wrap(err, "could not create directory")
	}
	return nil
}

// Remove deletes a local file or directory.
func (receiver LocalBackend) Remove(ctx context.Context, path string) error {
	err := os.RemoveAll(path)
	if nil != err {
		return erorr.Wrap(err, "could not remove file")
	}
	return nil
}

// Rename moves or renames a local file or directory.
func (receiver LocalBackend) Rename(ctx context.Context, oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if nil != err {
		return erorr.Wrap(err, "could not rename file")
	}
	return nil
}

// Copy copies a local file or directory.
func (receiver LocalBackend) Copy(ctx context.Context, srcPath, dstPath string) error {
	srcInfo, err := os.Stat(srcPath)
	if nil != err {
		return erorr.Wrap(err, "could not stat source")
	}

	if srcInfo.IsDir() {
		return receiver.copyDir(ctx, srcPath, dstPath)
	}

	return receiver.copyFile(ctx, srcPath, dstPath)
}

func (receiver LocalBackend) copyFile(ctx context.Context, srcPath, dstPath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	src, err := os.Open(srcPath)
	if nil != err {
		return erorr.Wrap(err, "could not open source file")
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if nil != err {
		return erorr.Wrap(err, "could not create destination file")
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if nil != err {
		return erorr.Wrap(err, "could not copy file contents")
	}

	return nil
}

func (receiver LocalBackend) copyDir(ctx context.Context, srcPath, dstPath string) error {
	err := os.MkdirAll(dstPath, 0755)
	if nil != err {
		return erorr.Wrap(err, "could not create destination directory")
	}

	entries, err := os.ReadDir(srcPath)
	if nil != err {
		return erorr.Wrap(err, "could not read source directory")
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		src := filepath.Join(srcPath, entry.Name())
		dst := filepath.Join(dstPath, entry.Name())

		err := receiver.Copy(ctx, src, dst)
		if nil != err {
			return err
		}
	}

	return nil
}
