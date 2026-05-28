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

	summary := core.MoveSummary{ArchiveBucket: bucket}
	for _, item := range scan.StaleItems {
		if err := ctx.Err(); err != nil {
			return summary, err
		}

		movedSize, err := moveRootItem(item, bucket)
		summary.MovedSize += movedSize
		if err != nil {
			summary.FailedPaths = append(summary.FailedPaths, item.Path)
			continue
		}
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

func moveRootItem(item core.StaleItem, bucket string) (int64, error) {
	if item.Kind == core.FolderItem {
		target := filepath.Join(bucket, item.DisplayName)
		targetInfo, err := os.Stat(target)
		if err == nil && targetInfo.IsDir() {
			return mergeFolder(item.Path, target)
		}
		if err != nil && !os.IsNotExist(err) {
			return 0, err
		}
	}

	target := filepath.Join(bucket, resolveTargetName(bucket, item.DisplayName))
	if err := os.Rename(item.Path, target); err != nil {
		return 0, err
	}
	return item.MoveSize, nil
}

func mergeFolder(source, target string) (int64, error) {
	entries, err := os.ReadDir(source)
	if err != nil {
		return 0, err
	}
	sortDirEntries(entries)

	var movedSize int64
	var failed []string
	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())

		info, err := os.Lstat(sourcePath)
		if err != nil {
			failed = append(failed, sourcePath)
			continue
		}

		if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			targetPath := filepath.Join(target, entry.Name())
			targetInfo, err := os.Stat(targetPath)
			if err == nil && targetInfo.IsDir() {
				size, mergeErr := mergeFolder(sourcePath, targetPath)
				movedSize += size
				if mergeErr != nil {
					failed = append(failed, sourcePath)
				}
				continue
			}
			if err != nil && !os.IsNotExist(err) {
				failed = append(failed, sourcePath)
				continue
			}
			targetPath = filepath.Join(target, resolveTargetName(target, entry.Name()))
			size, sizeErr := RecursiveSize(sourcePath)
			if sizeErr != nil {
				failed = append(failed, sourcePath)
				continue
			}
			if err := os.Rename(sourcePath, targetPath); err != nil {
				failed = append(failed, sourcePath)
				continue
			}
			movedSize += size
			continue
		}

		targetPath := filepath.Join(target, resolveTargetName(target, entry.Name()))
		size := int64(0)
		if info.Mode()&os.ModeSymlink == 0 {
			size = info.Size()
		}
		if err := os.Rename(sourcePath, targetPath); err != nil {
			failed = append(failed, sourcePath)
			continue
		}
		movedSize += size
	}

	_ = os.Remove(source)
	if len(failed) > 0 {
		return movedSize, fmt.Errorf("failed to move: %s", strings.Join(failed, ", "))
	}
	return movedSize, nil
}

func resolveTargetName(dir, name string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return name
	}
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return core.ResolveNumberedName(names, name)
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
