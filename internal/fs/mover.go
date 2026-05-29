package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"shed/internal/core"
)

type Mover struct {
	ArchiveRoot string
	Now         func() time.Time
}

func NewMover(archiveRoot string) Mover {
	return Mover{
		ArchiveRoot: archiveRoot,
		Now:         time.Now,
	}
}

func (mover Mover) Move(ctx context.Context, selectedFolder string, scan core.ScanResult) (core.MoveSummary, error) {
	if mover.Now == nil {
		mover.Now = time.Now
	}

	bucket := core.ArchiveBucket(mover.ArchiveRoot, mover.Now(), selectedFolder)
	if err := preflight(selectedFolder, mover.ArchiveRoot, bucket); err != nil {
		return core.MoveSummary{ArchiveBucket: bucket}, err
	}

	summary := core.NewMoveSummary(bucket)
	for _, item := range scan.StaleItems {
		if err := ctx.Err(); err != nil {
			return summary, err
		}

		moveRootItem(item, bucket, &summary)
	}

	return summary, nil
}

func preflight(selectedFolder, archiveRoot, bucket string) error {
	if err := requireDirectory(selectedFolder, "selected folder"); err != nil {
		return err
	}
	if err := ensureDirectory(archiveRoot, "Archive"); err != nil {
		return err
	}
	if err := ensureDirectory(bucket, "Archive bucket"); err != nil {
		return err
	}
	return nil
}

func requireDirectory(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s unavailable: %w", label, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a folder: %s", label, path)
	}
	return nil
}

func ensureDirectory(path, label string) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s is not a folder: %s", label, path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("%s unavailable: %w", label, err)
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return fmt.Errorf("%s unavailable: %w", label, err)
	}
	return nil
}

func moveRootItem(item core.StaleItem, bucket string, summary *core.MoveSummary) {
	decision := core.DecideArchiveMove(core.ArchiveMoveCandidate{
		Name: item.DisplayName,
		Kind: item.Kind,
	}, archiveEntries(bucket))
	target := filepath.Join(bucket, decision.TargetName)
	if decision.Action == core.MergeIntoExistingFolder {
		mergeFolder(item.Path, target, summary)
		return
	}

	if err := os.Rename(item.Path, target); err != nil {
		summary.RecordFailed(item.Path)
		return
	}
	summary.RecordMoved(item.MoveSize)
}

func mergeFolder(source, target string, summary *core.MoveSummary) {
	entries, err := os.ReadDir(source)
	if err != nil {
		summary.RecordFailed(source)
		return
	}
	sortDirEntries(entries)

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())

		info, err := os.Lstat(sourcePath)
		if err != nil {
			summary.RecordFailed(sourcePath)
			continue
		}

		kind := core.FileItem
		if info.Mode()&os.ModeSymlink != 0 {
			kind = core.SymlinkItem
		} else if info.IsDir() {
			kind = core.FolderItem
		}
		decision := core.DecideArchiveMove(core.ArchiveMoveCandidate{
			Name: entry.Name(),
			Kind: kind,
		}, archiveEntries(target))
		targetPath := filepath.Join(target, decision.TargetName)

		if decision.Action == core.MergeIntoExistingFolder {
			mergeFolder(sourcePath, targetPath, summary)
			continue
		}

		if kind == core.FolderItem {
			size, sizeErr := RecursiveSize(sourcePath)
			if sizeErr != nil {
				summary.RecordFailed(sourcePath)
				continue
			}
			if err := os.Rename(sourcePath, targetPath); err != nil {
				summary.RecordFailed(sourcePath)
				continue
			}
			summary.RecordMoved(size)
			continue
		}

		size := int64(0)
		if kind != core.SymlinkItem {
			size = info.Size()
		}
		if err := os.Rename(sourcePath, targetPath); err != nil {
			summary.RecordFailed(sourcePath)
			continue
		}
		summary.RecordMoved(size)
	}

	_ = os.Remove(source)
}

func archiveEntries(dir string) []core.ArchiveMoveEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	moveEntries := make([]core.ArchiveMoveEntry, 0, len(entries))
	for _, entry := range entries {
		kind := core.FileItem
		if entry.Type()&os.ModeSymlink != 0 {
			kind = core.SymlinkItem
		} else if entry.IsDir() {
			kind = core.FolderItem
		}
		moveEntries = append(moveEntries, core.ArchiveMoveEntry{
			Name: entry.Name(),
			Kind: kind,
		})
	}
	return moveEntries
}

func sortDirEntries(entries []os.DirEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		left := strings.ToLower(entries[i].Name())
		right := strings.ToLower(entries[j].Name())
		if left == right {
			return entries[i].Name() < entries[j].Name()
		}
		return left < right
	})
}
