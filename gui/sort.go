package gui

import (
	"sort"
	"strings"

	"topotron/lib/fileinfo"
)

// SortOrder defines how file entries are sorted.
type SortOrder string

const (
	SortNameAsc  SortOrder = "name-asc"
	SortNameDesc SortOrder = "name-desc"
	SortDateAsc  SortOrder = "date-asc"
	SortDateDesc SortOrder = "date-desc"
	SortSizeAsc  SortOrder = "size-asc"
	SortSizeDesc SortOrder = "size-desc"
)

// SortLabel returns a human-legible label for a [SortOrder].
func SortLabel(order SortOrder) string {
	switch order {
	case SortNameAsc:
		return "Name (A\u2192Z)"
	case SortNameDesc:
		return "Name (Z\u2192A)"
	case SortDateAsc:
		return "Date (oldest)"
	case SortDateDesc:
		return "Date (newest)"
	case SortSizeAsc:
		return "Size (smallest)"
	case SortSizeDesc:
		return "Size (largest)"
	default:
		return "Name (A\u2192Z)"
	}
}

// AllSortOrders returns all available sort orders.
func AllSortOrders() []SortOrder {
	return []SortOrder{
		SortNameAsc,
		SortNameDesc,
		SortDateAsc,
		SortDateDesc,
		SortSizeAsc,
		SortSizeDesc,
	}
}

// SortEntries sorts a slice of [libfileinfo.FileInfo] by the given [SortOrder].
// Directories are always sorted before files.
func SortEntries(entries []libfileinfo.FileInfo, order SortOrder) {
	sort.SliceStable(entries, func(i, j int) bool {
		a := entries[i]
		b := entries[j]

		// directories first
		if a.IsDir && !b.IsDir {
			return true
		}
		if !a.IsDir && b.IsDir {
			return false
		}

		switch order {
		case SortNameAsc:
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		case SortNameDesc:
			return strings.ToLower(a.Name) > strings.ToLower(b.Name)
		case SortDateAsc:
			return a.ModTime.Before(b.ModTime)
		case SortDateDesc:
			return a.ModTime.After(b.ModTime)
		case SortSizeAsc:
			return a.Size < b.Size
		case SortSizeDesc:
			return a.Size > b.Size
		default:
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		}
	})
}

// FilterEntries returns entries whose name contains the search text (case-insensitive).
func FilterEntries(entries []libfileinfo.FileInfo, search string) []libfileinfo.FileInfo {
	if "" == search {
		return entries
	}

	lower := strings.ToLower(search)

	var result []libfileinfo.FileInfo
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name), lower) {
			result = append(result, entry)
		}
	}

	return result
}
