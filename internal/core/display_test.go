package core

import (
	"path/filepath"
	"testing"
	"time"
)

func TestHeaderTitleUsesSelectedFolderBaseName(t *testing.T) {
	selected := filepath.Join("C:", "Users", "pavel", "Downloads")

	if got := HeaderTitle(selected); got != "Downloads" {
		t.Fatalf("expected Downloads, got %q", got)
	}
}

func TestHeaderTitleUsesCleanPathForFilesystemRoot(t *testing.T) {
	root := filepath.Clean(`C:\`)

	if got := HeaderTitle(root); got != root {
		t.Fatalf("expected root title %q, got %q", root, got)
	}
}

func TestCompactArchiveBucket(t *testing.T) {
	selected := filepath.Join("C:", "Users", "pavel", "Downloads")
	moveDate := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	want := filepath.Join("~", "Shed", "2026", "05", "Downloads")

	if got := CompactArchiveBucket(moveDate, selected); got != want {
		t.Fatalf("expected compact bucket %q, got %q", want, got)
	}
}
