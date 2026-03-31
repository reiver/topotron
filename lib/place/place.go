package libplace

import (
	"os"
	"path/filepath"
)

// Place represents a browsable location in the file manager,
// such as a local directory or a WebDAV server.
type Place struct {
	// identity
	Name string
	Path string

	// display
	Icon string
}

// DefaultPlaces returns the standard set of local filesystem places.
func DefaultPlaces() []Place {
	home := homeDir()

	places := []Place{
		{
			Name: "Home",
			Path: home,
			Icon: "user-home-symbolic",
		},
		{
			Name: "Documents",
			Path: filepath.Join(home, "Documents"),
			Icon: "folder-documents-symbolic",
		},
		{
			Name: "Downloads",
			Path: filepath.Join(home, "Downloads"),
			Icon: "folder-download-symbolic",
		},
		{
			Name: "Music",
			Path: filepath.Join(home, "Music"),
			Icon: "folder-music-symbolic",
		},
		{
			Name: "Pictures",
			Path: filepath.Join(home, "Pictures"),
			Icon: "folder-pictures-symbolic",
		},
		{
			Name: "Videos",
			Path: filepath.Join(home, "Videos"),
			Icon: "folder-videos-symbolic",
		},
	}

	// filter out directories that do not exist
	var result []Place
	for _, place := range places {
		info, err := os.Stat(place.Path)
		if nil != err {
			continue
		}
		if !info.IsDir() {
			continue
		}
		result = append(result, place)
	}

	// root is always available
	result = append(result, Place{
		Name: "Root",
		Path: "/",
		Icon: "drive-harddisk-symbolic",
	})

	return result
}

// IconForPath returns the appropriate icon for a directory path.
// Known XDG paths get special icons; others get folder-symbolic.
func IconForPath(path string) string {
	home := homeDir()

	switch path {
	case home:
		return "user-home-symbolic"
	case filepath.Join(home, "Documents"):
		return "folder-documents-symbolic"
	case filepath.Join(home, "Downloads"):
		return "folder-download-symbolic"
	case filepath.Join(home, "Music"):
		return "folder-music-symbolic"
	case filepath.Join(home, "Pictures"):
		return "folder-pictures-symbolic"
	case filepath.Join(home, "Videos"):
		return "folder-videos-symbolic"
	case "/":
		return "drive-harddisk-symbolic"
	default:
		return "folder-symbolic"
	}
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if nil != err {
		return "/home"
	}
	return home
}
