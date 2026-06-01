package core

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	DefaultRetentionAgeDays = 0
)

type ItemKind int

const (
	FileItem ItemKind = iota
	FolderItem
	SymlinkItem
)

type RootItem struct {
	Name     string
	Path     string
	Kind     ItemKind
	Modified time.Time
	Created  time.Time
	Hidden   bool
	Size     int64
	SizeErr  error
}

type StaleItem struct {
	DisplayName string
	Path        string
	Kind        ItemKind
	MoveSize    int64
}

type SkippedItem struct {
	Path string
	Err  error
}

type ScanResult struct {
	StaleItems   []StaleItem
	SkippedItems []SkippedItem
	MoveSize     int64
}

func ScanRootItems(items []RootItem, now time.Time) ScanResult {
	return ScanRootItemsWithRetentionAge(items, now, DefaultRetentionAgeDays)
}

func ScanRootItemsWithRetentionAge(items []RootItem, now time.Time, retentionAgeDays int) ScanResult {
	boundary := now.AddDate(0, 0, -retentionAgeDays)
	result := ScanResult{}

	for _, item := range items {
		if !Eligible(item) {
			continue
		}
		if item.SizeErr != nil {
			result.SkippedItems = append(result.SkippedItems, SkippedItem{Path: item.Path, Err: item.SizeErr})
			continue
		}
		if retentionAgeDays > 0 && !Stale(item, boundary) {
			continue
		}

		stale := StaleItem{
			DisplayName: item.Name,
			Path:        item.Path,
			Kind:        item.Kind,
			MoveSize:    item.Size,
		}
		result.StaleItems = append(result.StaleItems, stale)
		result.MoveSize += stale.MoveSize
	}

	SortStaleItems(result.StaleItems)
	return result
}

func Eligible(item RootItem) bool {
	if item.Hidden {
		return false
	}
	return item.Kind != FolderItem || !strings.HasPrefix(item.Name, ".")
}

func Stale(item RootItem, boundary time.Time) bool {
	switch item.Kind {
	case FolderItem:
		return !item.Created.After(boundary)
	default:
		return !item.Modified.After(boundary)
	}
}

func SortStaleItems(items []StaleItem) {
	sort.SliceStable(items, func(i, j int) bool {
		left, right := items[i], items[j]
		if left.Kind == FolderItem && right.Kind != FolderItem {
			return true
		}
		if left.Kind != FolderItem && right.Kind == FolderItem {
			return false
		}

		leftName := strings.ToLower(left.DisplayName)
		rightName := strings.ToLower(right.DisplayName)
		if leftName == rightName {
			return left.DisplayName < right.DisplayName
		}
		return leftName < rightName
	})
}

func IsShedSource(selectedFolder, shedRoot string) bool {
	selected := cleanForCompare(selectedFolder)
	shed := cleanForCompare(shedRoot)
	if selected == shed {
		return true
	}

	relative, err := filepath.Rel(shed, selected)
	if err != nil {
		return false
	}
	return relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func FormatSize(size int64) string {
	const unit = int64(1024)

	switch {
	case size < unit:
		return fmt.Sprintf("%d B", size)
	case size < unit*unit:
		return fmt.Sprintf("%d KB", size/unit)
	case size < unit*unit*unit:
		return fmt.Sprintf("%d MB", size/(unit*unit))
	default:
		return fmt.Sprintf("%d GB", size/(unit*unit*unit))
	}
}

func cleanForCompare(path string) string {
	return strings.ToLower(filepath.Clean(path))
}
