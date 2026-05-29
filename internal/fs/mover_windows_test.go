//go:build windows

package fs

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"shed/internal/core"
)

func TestMoverRenamesFileIntoArchiveBucket(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	root := t.TempDir()
	selected := filepath.Join(root, "Downloads")
	archive := filepath.Join(root, "Shed")
	if err := os.Mkdir(selected, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(selected, "old.txt")
	writeFile(t, source, "old")

	summary, err := Mover{ArchiveRoot: archive, Now: func() time.Time { return now }}.Move(context.Background(), selected, core.ScanResult{
		StaleItems: []core.StaleItem{{DisplayName: "old.txt", Path: source, Kind: core.FileItem, MoveSize: 3}},
	})
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(archive, "2026", "05", "Downloads", "old.txt")
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected moved file at %q: %v", target, err)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("expected source file to be moved, stat err=%v", err)
	}
	if summary.MovedSize != 3 {
		t.Fatalf("expected moved size 3, got %d", summary.MovedSize)
	}
}

func TestMoverRenamesFileConflictWithNumberedSuffix(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	root := t.TempDir()
	selected := filepath.Join(root, "Downloads")
	archive := filepath.Join(root, "Shed")
	bucket := filepath.Join(archive, "2026", "05", "Downloads")
	if err := os.MkdirAll(bucket, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(selected, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(bucket, "report.pdf"), "existing")
	source := filepath.Join(selected, "report.pdf")
	writeFile(t, source, "new")

	_, err := Mover{ArchiveRoot: archive, Now: func() time.Time { return now }}.Move(context.Background(), selected, core.ScanResult{
		StaleItems: []core.StaleItem{{DisplayName: "report.pdf", Path: source, Kind: core.FileItem, MoveSize: 3}},
	})
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(bucket, "report (1).pdf")
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected conflict target at %q: %v", target, err)
	}
}

func TestMoverRenamesSymlinkConflictWithNumberedSuffix(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	root := t.TempDir()
	selected := filepath.Join(root, "Downloads")
	archive := filepath.Join(root, "Shed")
	bucket := filepath.Join(archive, "2026", "05", "Downloads")
	if err := os.MkdirAll(bucket, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(selected, 0o700); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(root, "target.txt")
	writeFile(t, target, "target")
	source := filepath.Join(selected, "shortcut")
	if err := os.Symlink(target, source); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	writeFile(t, filepath.Join(bucket, "shortcut"), "existing")

	_, err := Mover{ArchiveRoot: archive, Now: func() time.Time { return now }}.Move(context.Background(), selected, core.ScanResult{
		StaleItems: []core.StaleItem{{DisplayName: "shortcut", Path: source, Kind: core.SymlinkItem}},
	})
	if err != nil {
		t.Fatal(err)
	}

	targetPath := filepath.Join(bucket, "shortcut (1)")
	info, err := os.Lstat(targetPath)
	if err != nil {
		t.Fatalf("expected conflict symlink target at %q: %v", targetPath, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected moved item to remain a symlink")
	}
}

func TestMoverMergesFolderConflictsRecursively(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	root := t.TempDir()
	selected := filepath.Join(root, "Downloads")
	archive := filepath.Join(root, "Shed")
	bucket := filepath.Join(archive, "2026", "05", "Downloads")
	if err := os.MkdirAll(filepath.Join(bucket, "project", "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(selected, "project", "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(bucket, "project", "nested", "keep.txt"), "keep")
	writeFile(t, filepath.Join(selected, "project", "nested", "keep.txt"), "new")
	writeFile(t, filepath.Join(selected, "project", "nested", "extra.txt"), "extra")

	summary, err := Mover{ArchiveRoot: archive, Now: func() time.Time { return now }}.Move(context.Background(), selected, core.ScanResult{
		StaleItems: []core.StaleItem{{DisplayName: "project", Path: filepath.Join(selected, "project"), Kind: core.FolderItem, MoveSize: 8}},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"keep.txt", "keep (1).txt", "extra.txt"} {
		target := filepath.Join(bucket, "project", "nested", name)
		if _, err := os.Stat(target); err != nil {
			t.Fatalf("expected merged file at %q: %v", target, err)
		}
	}
	if summary.MovedSize != 8 {
		t.Fatalf("expected moved size 8, got %d", summary.MovedSize)
	}
}

func TestMoverReportsSpecificNestedFailedMoveInPartialMerge(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	root := t.TempDir()
	selected := filepath.Join(root, "Downloads")
	archive := filepath.Join(root, "Shed")
	bucket := filepath.Join(archive, "2026", "05", "Downloads")
	if err := os.MkdirAll(filepath.Join(bucket, "project", "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(selected, "project", "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	locked := filepath.Join(selected, "project", "nested", "locked.txt")
	writeFile(t, locked, "locked")
	lockFileNoDelete(t, locked)

	summary, err := Mover{ArchiveRoot: archive, Now: func() time.Time { return now }}.Move(context.Background(), selected, core.ScanResult{
		StaleItems: []core.StaleItem{{DisplayName: "project", Path: filepath.Join(selected, "project"), Kind: core.FolderItem, MoveSize: 6}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(summary.FailedPaths) != 1 {
		t.Fatalf("expected one failed path, got %v", summary.FailedPaths)
	}
	if summary.FailedPaths[0] != locked {
		t.Fatalf("expected nested failed path %q, got %q", locked, summary.FailedPaths[0])
	}
}

func TestMoverPreflightFailsWhenArchiveBucketIsFile(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	root := t.TempDir()
	selected := filepath.Join(root, "Downloads")
	archive := filepath.Join(root, "Shed")
	if err := os.MkdirAll(filepath.Join(archive, "2026", "05"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(selected, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(archive, "2026", "05", "Downloads"), "not a directory")
	source := filepath.Join(selected, "old.txt")
	writeFile(t, source, "old")

	_, err := Mover{ArchiveRoot: archive, Now: func() time.Time { return now }}.Move(context.Background(), selected, core.ScanResult{
		StaleItems: []core.StaleItem{{DisplayName: "old.txt", Path: source, Kind: core.FileItem, MoveSize: 3}},
	})
	if err == nil {
		t.Fatalf("expected preflight error")
	}
	if _, statErr := os.Stat(source); statErr != nil {
		t.Fatalf("expected source to remain after preflight failure: %v", statErr)
	}
}

func TestMoverPreflightFailsWhenSelectedFolderIsMissing(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	root := t.TempDir()
	selected := filepath.Join(root, "Downloads")
	source := filepath.Join(selected, "old.txt")

	_, err := Mover{ArchiveRoot: filepath.Join(root, "Shed"), Now: func() time.Time { return now }}.Move(context.Background(), selected, core.ScanResult{
		StaleItems: []core.StaleItem{{DisplayName: "old.txt", Path: source, Kind: core.FileItem, MoveSize: 3}},
	})
	if err == nil {
		t.Fatalf("expected preflight error")
	}
}

func lockFileNoDelete(t *testing.T, path string) {
	t.Helper()
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(pointer, syscall.GENERIC_READ, syscall.FILE_SHARE_READ, nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := syscall.CloseHandle(handle); err != nil {
			t.Fatal(err)
		}
	})
}
