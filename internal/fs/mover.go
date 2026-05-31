package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"shed/internal/core"
)

type Mover struct {
	ShedRoot string
}

func NewMover(shedRoot string) Mover {
	return Mover{
		ShedRoot: shedRoot,
	}
}

func (mover Mover) Move(ctx context.Context, selectedFolder string, planned core.PlannedShedBucket, scan core.ScanResult) (core.MoveSummary, error) {
	bucket := planned.ShedBucket
	if err := preflight(selectedFolder, mover.ShedRoot, bucket); err != nil {
		return core.MoveSummary{ShedBucket: bucket}, err
	}

	return core.MoveIntoPlannedShedBucket(ctx, bucket, scan.StaleItems, moveAdapter{})
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

type moveAdapter struct{}

func (moveAdapter) ListShedEntries(dir string) ([]core.ShedMoveEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
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
	return moveEntries, nil
}

func (moveAdapter) MoveSize(path string, kind core.ItemKind) (int64, error) {
	if kind == core.SymlinkItem {
		return 0, nil
	}
	if kind == core.FolderItem {
		return RecursiveSize(path)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (moveAdapter) Rename(source, target string) error {
	return os.Rename(source, target)
}

func (moveAdapter) RemoveEmptyFolder(path string) error {
	return os.Remove(path)
}

func (moveAdapter) JoinPath(base, name string) string {
	return filepath.Join(base, name)
}
