package core

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func HeaderTitle(selectedFolder string) string {
	clean := filepath.Clean(selectedFolder)
	if filepath.Dir(clean) == clean {
		return clean
	}
	return filepath.Base(clean)
}

func CompactArchiveBucket(moveDate time.Time, selectedFolder string) string {
	return filepath.Join("~", "Shed", fmt.Sprintf("%04d", moveDate.Year()), fmt.Sprintf("%02d", int(moveDate.Month())), HeaderTitle(selectedFolder))
}

func ArchiveBucket(archiveRoot string, moveDate time.Time, selectedFolder string) string {
	return filepath.Join(archiveRoot, fmt.Sprintf("%04d", moveDate.Year()), fmt.Sprintf("%02d", int(moveDate.Month())), HeaderTitle(selectedFolder))
}

func WindowsNameKey(name string) string {
	return strings.ToLower(name)
}

func HasNameConflict(existing []string, name string) bool {
	key := WindowsNameKey(name)
	for _, candidate := range existing {
		if WindowsNameKey(candidate) == key {
			return true
		}
	}
	return false
}

func NumberedSuffixName(name string, number int) string {
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	if ext == "" {
		return fmt.Sprintf("%s (%d)", name, number)
	}
	return fmt.Sprintf("%s (%d)%s", stem, number, ext)
}

func ResolveNumberedName(existing []string, name string) string {
	if !HasNameConflict(existing, name) {
		return name
	}
	for number := 1; ; number++ {
		candidate := NumberedSuffixName(name, number)
		if !HasNameConflict(existing, candidate) {
			return candidate
		}
	}
}

type MoveSummary struct {
	ArchiveBucket string
	MovedSize     int64
	FailedPaths   []string
}
