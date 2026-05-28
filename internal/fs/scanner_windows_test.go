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

func TestScannerFindsStaleRootFilesOnly(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	selected := t.TempDir()
	old := now.AddDate(0, 0, -core.RetentionDays)

	rootFile := filepath.Join(selected, "old.txt")
	writeFile(t, rootFile, "root")
	if err := os.Chtimes(rootFile, old, old); err != nil {
		t.Fatal(err)
	}

	nestedDir := filepath.Join(selected, "nested")
	if err := os.Mkdir(nestedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	nestedFile := filepath.Join(nestedDir, "old-nested.txt")
	writeFile(t, nestedFile, "nested")
	if err := os.Chtimes(nestedFile, old, old); err != nil {
		t.Fatal(err)
	}

	result, err := Scanner{Now: func() time.Time { return now }}.Scan(context.Background(), selected)
	if err != nil {
		t.Fatal(err)
	}

	names := fsStaleNames(result.StaleItems)
	want := []string{"old.txt"}
	if !fsEqualStrings(names, want) {
		t.Fatalf("expected stale names %v, got %v", want, names)
	}
	if result.MoveSize != 4 {
		t.Fatalf("expected root file move size 4, got %d", result.MoveSize)
	}
}

func TestScannerUsesFolderCreationTime(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	selected := t.TempDir()
	old := now.AddDate(0, 0, -core.RetentionDays)

	folder := filepath.Join(selected, "old-folder")
	if err := os.Mkdir(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(folder, "content.txt"), "content")
	setCreationTime(t, folder, old)

	result, err := Scanner{Now: func() time.Time { return now }}.Scan(context.Background(), selected)
	if err != nil {
		t.Fatal(err)
	}

	names := fsStaleNames(result.StaleItems)
	want := []string{"old-folder"}
	if !fsEqualStrings(names, want) {
		t.Fatalf("expected stale names %v, got %v", want, names)
	}
	if result.MoveSize != 7 {
		t.Fatalf("expected folder content size 7, got %d", result.MoveSize)
	}
}

func TestScannerAppliesHiddenAndDotRules(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	selected := t.TempDir()
	old := now.AddDate(0, 0, -core.RetentionDays)

	hidden := filepath.Join(selected, "hidden.txt")
	writeOldFile(t, hidden, "hidden", old)
	setHidden(t, hidden)

	dotFolder := filepath.Join(selected, ".git")
	if err := os.Mkdir(dotFolder, 0o700); err != nil {
		t.Fatal(err)
	}
	setCreationTime(t, dotFolder, old)

	dotFile := filepath.Join(selected, ".env")
	writeOldFile(t, dotFile, "env", old)

	result, err := Scanner{Now: func() time.Time { return now }}.Scan(context.Background(), selected)
	if err != nil {
		t.Fatal(err)
	}

	names := fsStaleNames(result.StaleItems)
	want := []string{".env"}
	if !fsEqualStrings(names, want) {
		t.Fatalf("expected stale names %v, got %v", want, names)
	}
}

func TestScannerTreatsSymlinkAsLeaf(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	selected := t.TempDir()
	old := now.AddDate(0, 0, -core.RetentionDays)

	target := filepath.Join(selected, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(target, "large.txt"), "large target")

	link := filepath.Join(selected, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	setLastWriteTimeNoFollow(t, link, old)

	result, err := Scanner{Now: func() time.Time { return now }}.Scan(context.Background(), selected)
	if err != nil {
		t.Fatal(err)
	}

	var found bool
	for _, item := range result.StaleItems {
		if item.DisplayName == "link" {
			found = true
			if item.Kind != core.SymlinkItem {
				t.Fatalf("expected symlink kind, got %v", item.Kind)
			}
			if item.MoveSize != 0 {
				t.Fatalf("expected symlink move size 0, got %d", item.MoveSize)
			}
		}
	}
	if !found {
		t.Fatalf("expected symlink stale item")
	}
}

func TestScannerRejectsArchiveSources(t *testing.T) {
	home := t.TempDir()
	archive := ArchiveRootFromHome(home)
	selected := filepath.Join(archive, "2026", "05", "Downloads")
	if err := os.MkdirAll(selected, 0o700); err != nil {
		t.Fatal(err)
	}

	_, err := Scanner{ArchiveRoot: archive}.Scan(context.Background(), selected)
	if err == nil {
		t.Fatalf("expected Archive source error")
	}
}

func TestRecursiveSizeExcludesSymlinkTargets(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "file.txt"), "file")

	target := filepath.Join(t.TempDir(), "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(target, "large.txt"), "large target")

	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}

	size, err := RecursiveSize(root)
	if err != nil {
		t.Fatal(err)
	}
	if size != 4 {
		t.Fatalf("expected size 4, got %d", size)
	}
}

func writeOldFile(t *testing.T, path, content string, modTime time.Time) {
	t.Helper()
	writeFile(t, path, content)
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func setHidden(t *testing.T, path string) {
	t.Helper()
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.SetFileAttributes(pointer, attributes|fileAttributeHidden); err != nil {
		t.Fatal(err)
	}
}

func setCreationTime(t *testing.T, path string, created time.Time) {
	t.Helper()
	handle := openForTimeUpdate(t, path, syscall.FILE_FLAG_BACKUP_SEMANTICS)

	filetime := syscall.NsecToFiletime(created.UnixNano())
	if err := syscall.SetFileTime(handle, &filetime, nil, nil); err != nil {
		t.Fatal(err)
	}
}

func setLastWriteTimeNoFollow(t *testing.T, path string, modified time.Time) {
	t.Helper()
	handle := openForTimeUpdate(t, path, syscall.FILE_FLAG_OPEN_REPARSE_POINT|syscall.FILE_FLAG_BACKUP_SEMANTICS)

	filetime := syscall.NsecToFiletime(modified.UnixNano())
	if err := syscall.SetFileTime(handle, nil, nil, &filetime); err != nil {
		t.Fatal(err)
	}
}

func openForTimeUpdate(t *testing.T, path string, flags uint32) syscall.Handle {
	t.Helper()
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(pointer, 0x0100, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE, nil, syscall.OPEN_EXISTING, flags, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := syscall.CloseHandle(handle); err != nil {
			t.Fatal(err)
		}
	})
	return handle
}

func fsStaleNames(items []core.StaleItem) []string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.DisplayName
	}
	return names
}

func fsEqualStrings(left, right []string) bool {
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
