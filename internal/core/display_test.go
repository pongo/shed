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

func TestArchiveBucketUsesArchiveRootDateAndSelectedFolderBaseName(t *testing.T) {
	archiveRoot := filepath.Join("C:", "Users", "pavel", "Shed")
	selected := filepath.Join("D:", "Scratch", "Downloads")
	moveDate := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	want := filepath.Join(archiveRoot, "2026", "05", "Downloads")

	if got := ArchiveBucket(archiveRoot, moveDate, selected); got != want {
		t.Fatalf("expected bucket %q, got %q", want, got)
	}
}

func TestHasNameConflictUsesCaseInsensitiveWindowsSemantics(t *testing.T) {
	existing := []string{"Report.PDF"}

	if !HasNameConflict(existing, "report.pdf") {
		t.Fatalf("expected case-insensitive conflict")
	}
}

func TestResolveNumberedNamePlacesSuffixBeforeFinalExtension(t *testing.T) {
	tests := map[string]string{
		"report.pdf":     "report (2).pdf",
		"archive.tar.gz": "archive.tar (1).gz",
		"README":         "README (1)",
	}
	existing := []string{"report.pdf", "report (1).pdf", "archive.tar.gz", "README"}

	for name, want := range tests {
		if got := ResolveNumberedName(existing, name); got != want {
			t.Fatalf("ResolveNumberedName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestMoveSummaryStoresActualMovedSize(t *testing.T) {
	summary := MoveSummary{
		ArchiveBucket: filepath.Join("C:", "Users", "pavel", "Shed", "2026", "05", "Downloads"),
		MovedSize:     12,
	}

	if summary.MovedSize != 12 {
		t.Fatalf("expected actual moved size 12, got %d", summary.MovedSize)
	}
}
