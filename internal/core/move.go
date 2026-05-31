package core

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type shedMoveAction int

const (
	moveIntoShed shedMoveAction = iota
	mergeIntoExistingFolder
)

type shedMoveCandidate struct {
	Name string
	Kind ItemKind
}

type ShedMoveEntry struct {
	Name string
	Kind ItemKind
}

type shedMoveDecision struct {
	Action     shedMoveAction
	TargetName string
}

type ShedMoveAdapter interface {
	ListShedEntries(path string) ([]ShedMoveEntry, error)
	MoveSize(path string, kind ItemKind) (int64, error)
	Rename(source, target string) error
	RemoveEmptyFolder(path string) error
	JoinPath(base, name string) string
}

type MoveSummary struct {
	ShedBucket  string
	MovedSize   int64
	FailedPaths []string
}

func MoveIntoPlannedShedBucket(ctx context.Context, shedBucket string, staleItems []StaleItem, adapter ShedMoveAdapter) (MoveSummary, error) {
	summary := NewMoveSummary(shedBucket)
	for _, item := range staleItems {
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		moveRootItem(item, shedBucket, adapter, &summary)
	}
	return summary, nil
}

func moveRootItem(item StaleItem, bucket string, adapter ShedMoveAdapter, summary *MoveSummary) {
	decision := decideShedMove(shedMoveCandidate{
		Name: item.DisplayName,
		Kind: item.Kind,
	}, existingEntries(bucket, adapter))
	target := adapter.JoinPath(bucket, decision.TargetName)
	if decision.Action == mergeIntoExistingFolder {
		mergeFolder(item.Path, target, adapter, summary)
		return
	}

	if err := adapter.Rename(item.Path, target); err != nil {
		summary.RecordFailed(item.Path)
		return
	}
	summary.RecordMoved(item.MoveSize)
}

func mergeFolder(source, target string, adapter ShedMoveAdapter, summary *MoveSummary) {
	entries, err := adapter.ListShedEntries(source)
	if err != nil {
		summary.RecordFailed(source)
		return
	}

	for _, entry := range entries {
		sourcePath := adapter.JoinPath(source, entry.Name)
		decision := decideShedMove(shedMoveCandidate{
			Name: entry.Name,
			Kind: entry.Kind,
		}, existingEntries(target, adapter))
		targetPath := adapter.JoinPath(target, decision.TargetName)

		if decision.Action == mergeIntoExistingFolder {
			mergeFolder(sourcePath, targetPath, adapter, summary)
			continue
		}

		size, sizeErr := adapter.MoveSize(sourcePath, entry.Kind)
		if sizeErr != nil {
			summary.RecordFailed(sourcePath)
			continue
		}
		if err := adapter.Rename(sourcePath, targetPath); err != nil {
			summary.RecordFailed(sourcePath)
			continue
		}
		summary.RecordMoved(size)
	}

	_ = adapter.RemoveEmptyFolder(source)
}

func existingEntries(path string, adapter ShedMoveAdapter) []ShedMoveEntry {
	entries, err := adapter.ListShedEntries(path)
	if err != nil {
		return nil
	}
	return entries
}

func decideShedMove(candidate shedMoveCandidate, existing []ShedMoveEntry) shedMoveDecision {
	if candidate.Kind == FolderItem {
		for _, entry := range existing {
			if entry.Kind == FolderItem && WindowsNameKey(entry.Name) == WindowsNameKey(candidate.Name) {
				return shedMoveDecision{
					Action:     mergeIntoExistingFolder,
					TargetName: entry.Name,
				}
			}
		}
	}

	return shedMoveDecision{
		Action:     moveIntoShed,
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
