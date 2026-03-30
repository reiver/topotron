package libbackend

import (
	"context"
	"io"
	"path"
	"time"

	"codeberg.org/reiver/go-erorr"
	"github.com/studio-b12/gowebdav"

	"topotron/lib/fileinfo"
	"topotron/lib/icon"
)

// WebDAVBackend implements [FileBackend] for a remote WebDAV server.
type WebDAVBackend struct {
	// connection
	client *gowebdav.Client

	// display
	name string
}

var _ FileBackend = &WebDAVBackend{}

// NewWebDAVBackend creates a new [WebDAVBackend] connected to the given URL.
func NewWebDAVBackend(url, user, pass, name string) *WebDAVBackend {
	client := gowebdav.NewClient(url, user, pass)
	client.SetTimeout(30 * time.Second)

	return &WebDAVBackend{
		client: client,
		name:   name,
	}
}

// Connect tests the connection to the WebDAV server.
func (receiver *WebDAVBackend) Connect() error {
	err := receiver.client.Connect()
	if nil != err {
		return erorr.Wrap(err, "could not connect to WebDAV server")
	}
	return nil
}

// Name returns the display name for this WebDAV server.
func (receiver *WebDAVBackend) Name() string {
	return receiver.name
}

// List returns the contents of a remote directory.
func (receiver *WebDAVBackend) List(ctx context.Context, remotePath string, showHidden bool) ([]libfileinfo.FileInfo, error) {
	files, err := receiver.client.ReadDir(remotePath)
	if nil != err {
		return nil, erorr.Wrap(err, "could not list remote directory")
	}

	var result []libfileinfo.FileInfo

	for _, f := range files {
		name := f.Name()

		if !showHidden && len(name) > 0 && name[0] == '.' {
			continue
		}

		fullPath := path.Join(remotePath, name)

		result = append(result, libfileinfo.FileInfo{
			Name:    name,
			Path:    fullPath,
			Size:    f.Size(),
			ModTime: f.ModTime(),
			IsDir:   f.IsDir(),
			Icon:    libicon.ForFile(name, f.IsDir()),
		})
	}

	return result, nil
}

// Stat returns metadata for a remote file or directory.
func (receiver *WebDAVBackend) Stat(ctx context.Context, remotePath string) (libfileinfo.FileInfo, error) {
	f, err := receiver.client.Stat(remotePath)
	if nil != err {
		return libfileinfo.FileInfo{}, erorr.Wrap(err, "could not stat remote file")
	}

	return libfileinfo.FileInfo{
		Name:    f.Name(),
		Path:    remotePath,
		Size:    f.Size(),
		ModTime: f.ModTime(),
		IsDir:   f.IsDir(),
		Icon:    libicon.ForFile(f.Name(), f.IsDir()),
	}, nil
}

// Open returns a reader for a remote file.
func (receiver *WebDAVBackend) Open(ctx context.Context, remotePath string) (io.ReadCloser, error) {
	reader, err := receiver.client.ReadStream(remotePath)
	if nil != err {
		return nil, erorr.Wrap(err, "could not open remote file")
	}
	return reader, nil
}

// Create writes content to a remote file.
func (receiver *WebDAVBackend) Create(ctx context.Context, remotePath string, r io.Reader) error {
	err := receiver.client.WriteStream(remotePath, r, 0644)
	if nil != err {
		return erorr.Wrap(err, "could not create remote file")
	}
	return nil
}

// Mkdir creates a remote directory and any necessary parents.
func (receiver *WebDAVBackend) Mkdir(ctx context.Context, remotePath string) error {
	err := receiver.client.MkdirAll(remotePath, 0755)
	if nil != err {
		return erorr.Wrap(err, "could not create remote directory")
	}
	return nil
}

// Remove deletes a remote file or directory.
func (receiver *WebDAVBackend) Remove(ctx context.Context, remotePath string) error {
	err := receiver.client.Remove(remotePath)
	if nil != err {
		return erorr.Wrap(err, "could not remove remote file")
	}
	return nil
}

// Rename moves or renames a remote file or directory.
func (receiver *WebDAVBackend) Rename(ctx context.Context, oldPath, newPath string) error {
	err := receiver.client.Rename(oldPath, newPath, false)
	if nil != err {
		return erorr.Wrap(err, "could not rename remote file")
	}
	return nil
}

// Copy copies a remote file or directory.
func (receiver *WebDAVBackend) Copy(ctx context.Context, srcPath, dstPath string) error {
	err := receiver.client.Copy(srcPath, dstPath, false)
	if nil != err {
		return erorr.Wrap(err, "could not copy remote file")
	}
	return nil
}
