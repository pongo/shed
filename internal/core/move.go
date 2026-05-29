package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

type ArchiveMoveAction int

const (
	MoveIntoArchive ArchiveMoveAction = iota
	MergeIntoExistingFolder
)

type ArchiveMoveCandidate struct {
	Name string
	Kind ItemKind
}

type ArchiveMoveEntry struct {
	Name string
	Kind ItemKind
}

type ArchiveMoveDecision struct {
	Action     ArchiveMoveAction
	TargetName string
}

type MoveSummary struct {
	ArchiveBucket string
	MovedSize     int64
	FailedPaths   []string
}

func DecideArchiveMove(candidate ArchiveMoveCandidate, existing []ArchiveMoveEntry) ArchiveMoveDecision {
	if candidate.Kind == FolderItem {
		for _, entry := range existing {
			if entry.Kind == FolderItem && WindowsNameKey(entry.Name) == WindowsNameKey(candidate.Name) {
				return ArchiveMoveDecision{
					Action:     MergeIntoExistingFolder,
					TargetName: entry.Name,
				}
			}
		}
	}

	return ArchiveMoveDecision{
		Action:     MoveIntoArchive,
		TargetName: ResolveNumberedName(entryNames(existing), candidate.Name),
	}
}

func NewMoveSummary(archiveBucket string) MoveSummary {
	return MoveSummary{ArchiveBucket: archiveBucket}
}

func (summary *MoveSummary) RecordMoved(size int64) {
	summary.MovedSize += size
}

func (summary *MoveSummary) RecordFailed(path string) {
	summary.FailedPaths = append(summary.FailedPaths, path)
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

func entryNames(entries []ArchiveMoveEntry) []string {
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name
	}
	return names
}
