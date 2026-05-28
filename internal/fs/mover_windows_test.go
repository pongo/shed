//go:build windows

package fs

import (
	"context"
	"os"
	"path/filepath"
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
