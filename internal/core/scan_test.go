package core

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestScanRootItemsUsesRetentionBoundaryByKind(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	boundary := now.AddDate(0, 0, -RetentionDays)

	result := ScanRootItems([]RootItem{
		{Name: "old-file.txt", Path: `C:\Work\old-file.txt`, Kind: FileItem, Modified: boundary, Size: 10},
		{Name: "new-file.txt", Path: `C:\Work\new-file.txt`, Kind: FileItem, Modified: boundary.Add(time.Nanosecond), Size: 20},
		{Name: "old-folder", Path: `C:\Work\old-folder`, Kind: FolderItem, Created: boundary, Modified: now, Size: 30},
		{Name: "new-folder", Path: `C:\Work\new-folder`, Kind: FolderItem, Created: boundary.Add(time.Nanosecond), Size: 40},
	}, now)

	names := staleNames(result.StaleItems)
	want := []string{"old-folder", "old-file.txt"}
	if !equalStrings(names, want) {
		t.Fatalf("expected stale names %v, got %v", want, names)
	}
	if result.MoveSize != 40 {
		t.Fatalf("expected move size 40, got %d", result.MoveSize)
	}
}

func TestScanRootItemsUsesCustomRetentionAge(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	thirtyDaysOld := now.AddDate(0, 0, -30)

	result := ScanRootItemsWithRetentionAge([]RootItem{
		{Name: "custom-age.txt", Path: `C:\Work\custom-age.txt`, Kind: FileItem, Modified: thirtyDaysOld, Size: 10},
	}, now, 30)

	names := staleNames(result.StaleItems)
	want := []string{"custom-age.txt"}
	if !equalStrings(names, want) {
		t.Fatalf("expected stale names %v, got %v", want, names)
	}
}

func TestScanRootItemsTreatsAgeZeroAsAllEligibleItems(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	future := now.AddDate(0, 0, 1)

	result := ScanRootItemsWithRetentionAge([]RootItem{
		{Name: "future-file.txt", Path: `C:\Work\future-file.txt`, Kind: FileItem, Modified: future, Size: 10},
		{Name: "future-folder", Path: `C:\Work\future-folder`, Kind: FolderItem, Created: future, Size: 20},
		{Name: "hidden.txt", Path: `C:\Work\hidden.txt`, Kind: FileItem, Modified: future, Hidden: true, Size: 30},
	}, now, 0)

	names := staleNames(result.StaleItems)
	want := []string{"future-folder", "future-file.txt"}
	if !equalStrings(names, want) {
		t.Fatalf("expected stale names %v, got %v", want, names)
	}
}

func TestScanRootItemsAppliesEligibilityRules(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	old := now.AddDate(0, 0, -RetentionDays)

	result := ScanRootItems([]RootItem{
		{Name: "hidden.txt", Path: `C:\Work\hidden.txt`, Kind: FileItem, Modified: old, Hidden: true, Size: 10},
		{Name: ".git", Path: `C:\Work\.git`, Kind: FolderItem, Created: old, Size: 20},
		{Name: ".env", Path: `C:\Work\.env`, Kind: FileItem, Modified: old, Size: 30},
		{Name: "link", Path: `C:\Work\link`, Kind: SymlinkItem, Modified: old, Size: 0},
	}, now)

	names := staleNames(result.StaleItems)
	want := []string{".env", "link"}
	if !equalStrings(names, want) {
		t.Fatalf("expected stale names %v, got %v", want, names)
	}
	if result.MoveSize != 30 {
		t.Fatalf("expected symlink to contribute no move size, got %d", result.MoveSize)
	}
}

func TestScanRootItemsRecordsSkippedItems(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	old := now.AddDate(0, 0, -RetentionDays)
	readErr := errors.New("access denied")

	result := ScanRootItems([]RootItem{
		{Name: "bad", Path: `C:\Work\bad`, Kind: FolderItem, Created: old, SizeErr: readErr},
		{Name: "good", Path: `C:\Work\good`, Kind: FileItem, Modified: old, Size: 15},
	}, now)

	if len(result.SkippedItems) != 1 {
		t.Fatalf("expected one skipped item, got %d", len(result.SkippedItems))
	}
	if result.SkippedItems[0].Path != `C:\Work\bad` {
		t.Fatalf("expected skipped path, got %q", result.SkippedItems[0].Path)
	}
	if result.MoveSize != 15 {
		t.Fatalf("expected only readable stale item size, got %d", result.MoveSize)
	}
}

func TestScanRootItemsDoesNotSkipIneligibleItems(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	old := now.AddDate(0, 0, -RetentionDays)

	result := ScanRootItems([]RootItem{
		{Name: "hidden", Path: `C:\Work\hidden`, Kind: FolderItem, Created: old, Hidden: true, SizeErr: errors.New("access denied")},
		{Name: ".git", Path: `C:\Work\.git`, Kind: FolderItem, Created: old, SizeErr: errors.New("access denied")},
	}, now)

	if len(result.SkippedItems) != 0 {
		t.Fatalf("expected ineligible items to be excluded, got skipped %v", result.SkippedItems)
	}
}

func TestSortStaleItemsFoldersFirstThenCaseInsensitiveNames(t *testing.T) {
	items := []StaleItem{
		{DisplayName: "z.txt", Kind: FileItem},
		{DisplayName: "Beta", Kind: FolderItem},
		{DisplayName: "alpha.txt", Kind: FileItem},
		{DisplayName: "Alpha", Kind: FolderItem},
		{DisplayName: "ALPHA.txt", Kind: FileItem},
	}

	SortStaleItems(items)

	names := staleNames(items)
	want := []string{"Alpha", "Beta", "ALPHA.txt", "alpha.txt", "z.txt"}
	if !equalStrings(names, want) {
		t.Fatalf("expected sorted names %v, got %v", want, names)
	}
}

func TestIsShedSource(t *testing.T) {
	home := filepath.Join("C:", "Users", "pavel")
	shed := filepath.Join(home, "Shed")

	tests := []struct {
		name     string
		selected string
		want     bool
	}{
		{name: "shed root", selected: shed, want: true},
		{name: "inside shed", selected: filepath.Join(shed, "2026", "05", "Downloads"), want: true},
		{name: "sibling", selected: filepath.Join(home, "ShedNotes"), want: false},
		{name: "outside", selected: filepath.Join(home, "Downloads"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsShedSource(tt.selected, shed); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := map[int64]string{
		12:                     "12 B",
		2 * 1024:               "2 KB",
		3 * 1024 * 1024:        "3 MB",
		4 * 1024 * 1024 * 1024: "4 GB",
	}

	for size, want := range tests {
		if got := FormatSize(size); got != want {
			t.Fatalf("FormatSize(%d) = %q, want %q", size, got, want)
		}
	}
}

func staleNames(items []StaleItem) []string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.DisplayName
	}
	return names
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
