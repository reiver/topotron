package libicon

import (
	"path/filepath"
	"strings"
)

// ForFile returns the GTK icon name for a file based on its name and
// whether it is a directory.
func ForFile(name string, isDir bool) string {
	if isDir {
		return "folder-symbolic"
	}

	ext := strings.ToLower(filepath.Ext(name))

	icon, found := extensionIcons[ext]
	if !found {
		return "text-x-generic-symbolic"
	}

	return icon
}

var extensionIcons = map[string]string{
	// text
	".txt":  "text-x-generic-symbolic",
	".md":   "text-x-generic-symbolic",
	".log":  "text-x-generic-symbolic",
	".csv":  "text-x-generic-symbolic",
	".json": "text-x-generic-symbolic",
	".xml":  "text-x-generic-symbolic",
	".yaml": "text-x-generic-symbolic",
	".yml":  "text-x-generic-symbolic",
	".toml": "text-x-generic-symbolic",
	".ini":  "text-x-generic-symbolic",
	".cfg":  "text-x-generic-symbolic",
	".conf": "text-x-generic-symbolic",

	// source code
	".go":   "text-x-generic-symbolic",
	".py":   "text-x-generic-symbolic",
	".js":   "text-x-generic-symbolic",
	".ts":   "text-x-generic-symbolic",
	".rs":   "text-x-generic-symbolic",
	".c":    "text-x-generic-symbolic",
	".h":    "text-x-generic-symbolic",
	".cpp":  "text-x-generic-symbolic",
	".java": "text-x-generic-symbolic",
	".rb":   "text-x-generic-symbolic",
	".sh":   "text-x-generic-symbolic",
	".html": "text-x-generic-symbolic",
	".css":  "text-x-generic-symbolic",

	// documents
	".pdf":  "x-office-document-symbolic",
	".doc":  "x-office-document-symbolic",
	".docx": "x-office-document-symbolic",
	".odt":  "x-office-document-symbolic",
	".rtf":  "x-office-document-symbolic",
	".xls":  "x-office-spreadsheet-symbolic",
	".xlsx": "x-office-spreadsheet-symbolic",
	".ods":  "x-office-spreadsheet-symbolic",
	".ppt":  "x-office-presentation-symbolic",
	".pptx": "x-office-presentation-symbolic",
	".odp":  "x-office-presentation-symbolic",

	// images
	".png":  "image-x-generic-symbolic",
	".jpg":  "image-x-generic-symbolic",
	".jpeg": "image-x-generic-symbolic",
	".gif":  "image-x-generic-symbolic",
	".bmp":  "image-x-generic-symbolic",
	".svg":  "image-x-generic-symbolic",
	".webp": "image-x-generic-symbolic",
	".ico":  "image-x-generic-symbolic",
	".tif":  "image-x-generic-symbolic",
	".tiff": "image-x-generic-symbolic",

	// audio
	".mp3":  "audio-x-generic-symbolic",
	".wav":  "audio-x-generic-symbolic",
	".ogg":  "audio-x-generic-symbolic",
	".flac": "audio-x-generic-symbolic",
	".aac":  "audio-x-generic-symbolic",
	".m4a":  "audio-x-generic-symbolic",
	".wma":  "audio-x-generic-symbolic",
	".opus": "audio-x-generic-symbolic",

	// video
	".mp4":  "video-x-generic-symbolic",
	".mkv":  "video-x-generic-symbolic",
	".avi":  "video-x-generic-symbolic",
	".mov":  "video-x-generic-symbolic",
	".wmv":  "video-x-generic-symbolic",
	".webm": "video-x-generic-symbolic",
	".flv":  "video-x-generic-symbolic",
	".m4v":  "video-x-generic-symbolic",

	// archives
	".zip":  "package-x-generic-symbolic",
	".tar":  "package-x-generic-symbolic",
	".gz":   "package-x-generic-symbolic",
	".bz2":  "package-x-generic-symbolic",
	".xz":   "package-x-generic-symbolic",
	".7z":   "package-x-generic-symbolic",
	".rar":  "package-x-generic-symbolic",
	".deb":  "package-x-generic-symbolic",
	".rpm":  "package-x-generic-symbolic",
}
