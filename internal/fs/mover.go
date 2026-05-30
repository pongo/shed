package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"shed/internal/core"
)

type Mover struct {
	ShedRoot string
	Now      func() time.Time
}

func NewMover(shedRoot string) Mover {
	return Mover{
		ShedRoot: shedRoot,
		Now:      time.Now,
	}
}

func (mover Mover) Move(ctx context.Context, selectedFolder string, scan core.ScanResult) (core.MoveSummary, error) {
	if mover.Now == nil {
		mover.Now = time.Now
	}

	bucket := core.ShedBucket(mover.ShedRoot, mover.Now(), selectedFolder)
	if err := preflight(selectedFolder, mover.ShedRoot, bucket); err != nil {
		return core.MoveSummary{ShedBucket: bucket}, err
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

func preflight(selectedFolder, shedRoot, bucket string) error {
	if err := requireDirectory(selectedFolder, "selected folder"); err != nil {
		return err
	}
	if err := ensureDirectory(shedRoot, "Shed"); err != nil {
		return err
	}
	if err := ensureDirectory(bucket, "Shed bucket"); err != nil {
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
	decision := core.DecideShedMove(core.ShedMoveCandidate{
		Name: item.DisplayName,
		Kind: item.Kind,
	}, shedEntries(bucket))
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
		decision := core.DecideShedMove(core.ShedMoveCandidate{
			Name: entry.Name(),
			Kind: kind,
		}, shedEntries(target))
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

func shedEntries(dir string) []core.ShedMoveEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	moveEntries := make([]core.ShedMoveEntry, 0, len(entries))
	for _, entry := range entries {
		kind := core.FileItem
		if entry.Type()&os.ModeSymlink != 0 {
			kind = core.SymlinkItem
		} else if entry.IsDir() {
			kind = core.FolderItem
		}
		moveEntries = append(moveEntries, core.ShedMoveEntry{
			Name: entry.Name(),
			Kind: kind,
		})
	}
	return moveEntries
}
