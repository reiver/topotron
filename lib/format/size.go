package libformat

import "fmt"

// Size formats a byte count as a human-legible string.
//
// Examples:
//
//	Size(0)          → "0 B"
//	Size(1023)       → "1023 B"
//	Size(1024)       → "1.0 KB"
//	Size(1048576)    → "1.0 MB"
//	Size(1073741824) → "1.0 GB"
func Size(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
		tb = gb * 1024
	)

	switch {
	case bytes >= tb:
		return fmt.Sprintf("%.1f TB", float64(bytes)/float64(tb))
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
