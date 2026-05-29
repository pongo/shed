package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

type ShedMoveAction int

const (
	MoveIntoShed ShedMoveAction = iota
	MergeIntoExistingFolder
)

type ShedMoveCandidate struct {
	Name string
	Kind ItemKind
}

type ShedMoveEntry struct {
	Name string
	Kind ItemKind
}

type ShedMoveDecision struct {
	Action     ShedMoveAction
	TargetName string
}

type MoveSummary struct {
	ShedBucket  string
	MovedSize   int64
	FailedPaths []string
}

func DecideShedMove(candidate ShedMoveCandidate, existing []ShedMoveEntry) ShedMoveDecision {
	if candidate.Kind == FolderItem {
		for _, entry := range existing {
			if entry.Kind == FolderItem && WindowsNameKey(entry.Name) == WindowsNameKey(candidate.Name) {
				return ShedMoveDecision{
					Action:     MergeIntoExistingFolder,
					TargetName: entry.Name,
				}
			}
		}
	}

	return ShedMoveDecision{
		Action:     MoveIntoShed,
		TargetName: ResolveNumberedName(entryNames(existing), candidate.Name),
	}
}

func NewMoveSummary(shedBucket string) MoveSummary {
	return MoveSummary{ShedBucket: shedBucket}
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

func entryNames(entries []ShedMoveEntry) []string {
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name
	}
	return names
}
