package core

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestMoveIntoPlannedShedBucketUsesNumberedSuffixForFileConflict(t *testing.T) {
	bucket := filepath.Join("Shed", "2026", "05", "Downloads")
	source := filepath.Join("Downloads", "report.pdf")
	adapter := newRecordingMoveAdapter()
	adapter.entries[bucket] = []ShedMoveEntry{{Name: "report.pdf", Kind: FileItem}}

	summary, err := MoveIntoPlannedShedBucket(context.Background(), bucket, []StaleItem{
		{DisplayName: "report.pdf", Path: source, Kind: FileItem, MoveSize: 3},
	}, adapter)
	if err != nil {
		t.Fatal(err)
	}

	if got := adapter.renames[source]; got != filepath.Join(bucket, "report (1).pdf") {
		t.Fatalf("expected numbered target, got %q", got)
	}
	if summary.MovedSize != 3 {
		t.Fatalf("expected moved size 3, got %d", summary.MovedSize)
	}
}

func TestMoveIntoPlannedShedBucketMergesFolderConflictRecursively(t *testing.T) {
	bucket := filepath.Join("Shed", "2026", "05", "Downloads")
	source := filepath.Join("Downloads", "project")
	sourceNested := filepath.Join(source, "nested")
	targetProject := filepath.Join(bucket, "project")
	targetNested := filepath.Join(targetProject, "nested")
	adapter := newRecordingMoveAdapter()
	adapter.entries[bucket] = []ShedMoveEntry{{Name: "project", Kind: FolderItem}}
	adapter.entries[source] = []ShedMoveEntry{{Name: "nested", Kind: FolderItem}}
	adapter.entries[targetProject] = []ShedMoveEntry{{Name: "nested", Kind: FolderItem}}
	adapter.entries[sourceNested] = []ShedMoveEntry{
		{Name: "keep.txt", Kind: FileItem},
		{Name: "extra.txt", Kind: FileItem},
	}
	adapter.entries[targetNested] = []ShedMoveEntry{{Name: "keep.txt", Kind: FileItem}}
	adapter.sizes[filepath.Join(sourceNested, "keep.txt")] = 3
	adapter.sizes[filepath.Join(sourceNested, "extra.txt")] = 5

	summary, err := MoveIntoPlannedShedBucket(context.Background(), bucket, []StaleItem{
		{DisplayName: "project", Path: source, Kind: FolderItem, MoveSize: 8},
	}, adapter)
	if err != nil {
		t.Fatal(err)
	}

	if got := adapter.renames[filepath.Join(sourceNested, "keep.txt")]; got != filepath.Join(targetNested, "keep (1).txt") {
		t.Fatalf("expected nested conflict target, got %q", got)
	}
	if got := adapter.renames[filepath.Join(sourceNested, "extra.txt")]; got != filepath.Join(targetNested, "extra.txt") {
		t.Fatalf("expected nested extra target, got %q", got)
	}
	if summary.MovedSize != 8 {
		t.Fatalf("expected moved size 8, got %d", summary.MovedSize)
	}
	if !adapter.removed[sourceNested] || !adapter.removed[source] {
		t.Fatalf("expected merged folders to be removed when empty")
	}
}

func TestMoveIntoPlannedShedBucketRecordsSpecificNestedFailedMove(t *testing.T) {
	bucket := filepath.Join("Shed", "2026", "05", "Downloads")
	source := filepath.Join("Downloads", "project")
	locked := filepath.Join(source, "locked.txt")
	adapter := newRecordingMoveAdapter()
	adapter.entries[bucket] = []ShedMoveEntry{{Name: "project", Kind: FolderItem}}
	adapter.entries[source] = []ShedMoveEntry{{Name: "locked.txt", Kind: FileItem}}
	adapter.sizes[locked] = 6
	adapter.renameErrs[locked] = errors.New("locked")

	summary, err := MoveIntoPlannedShedBucket(context.Background(), bucket, []StaleItem{
		{DisplayName: "project", Path: source, Kind: FolderItem, MoveSize: 6},
	}, adapter)
	if err != nil {
		t.Fatal(err)
	}

	if summary.MovedSize != 0 {
		t.Fatalf("expected no moved size, got %d", summary.MovedSize)
	}
	if len(summary.FailedPaths) != 1 || summary.FailedPaths[0] != locked {
		t.Fatalf("expected nested failed path %q, got %v", locked, summary.FailedPaths)
	}
}

func TestMoveIntoPlannedShedBucketRecordsSizeFailureWithoutMovedBytes(t *testing.T) {
	bucket := filepath.Join("Shed", "2026", "05", "Downloads")
	source := filepath.Join("Downloads", "project")
	nested := filepath.Join(source, "nested")
	adapter := newRecordingMoveAdapter()
	adapter.entries[bucket] = []ShedMoveEntry{{Name: "project", Kind: FolderItem}}
	adapter.entries[source] = []ShedMoveEntry{{Name: "nested", Kind: FolderItem}}
	adapter.sizeErrs[nested] = errors.New("size denied")

	summary, err := MoveIntoPlannedShedBucket(context.Background(), bucket, []StaleItem{
		{DisplayName: "project", Path: source, Kind: FolderItem, MoveSize: 6},
	}, adapter)
	if err != nil {
		t.Fatal(err)
	}

	if summary.MovedSize != 0 {
		t.Fatalf("expected no moved size, got %d", summary.MovedSize)
	}
	if len(summary.FailedPaths) != 1 || summary.FailedPaths[0] != nested {
		t.Fatalf("expected size failure path %q, got %v", nested, summary.FailedPaths)
	}
	if _, renamed := adapter.renames[nested]; renamed {
		t.Fatalf("expected no rename after size failure")
	}
}

func TestMoveSummaryRecordsMovedSizeAndSpecificFailedPaths(t *testing.T) {
	summary := NewMoveSummary(`C:\Users\pavel\Shed\2026\05\Downloads`)

	summary.RecordMoved(12)
	summary.RecordFailed(`C:\Users\pavel\Downloads\project\nested\file.txt`)

	if summary.MovedSize != 12 {
		t.Fatalf("expected moved size 12, got %d", summary.MovedSize)
	}
	if len(summary.FailedPaths) != 1 {
		t.Fatalf("expected one failed path, got %d", len(summary.FailedPaths))
	}
	if summary.FailedPaths[0] != `C:\Users\pavel\Downloads\project\nested\file.txt` {
		t.Fatalf("expected nested failed path, got %q", summary.FailedPaths[0])
	}
}

type recordingMoveAdapter struct {
	entries    map[string][]ShedMoveEntry
	entryErrs  map[string]error
	sizes      map[string]int64
	sizeErrs   map[string]error
	renameErrs map[string]error
	renames    map[string]string
	removed    map[string]bool
}

func newRecordingMoveAdapter() *recordingMoveAdapter {
	return &recordingMoveAdapter{
		entries:    map[string][]ShedMoveEntry{},
		entryErrs:  map[string]error{},
		sizes:      map[string]int64{},
		sizeErrs:   map[string]error{},
		renameErrs: map[string]error{},
		renames:    map[string]string{},
		removed:    map[string]bool{},
	}
}

func (adapter *recordingMoveAdapter) ListShedEntries(path string) ([]ShedMoveEntry, error) {
	if err := adapter.entryErrs[path]; err != nil {
		return nil, err
	}
	return adapter.entries[path], nil
}

func (adapter *recordingMoveAdapter) MoveSize(path string, kind ItemKind) (int64, error) {
	if err := adapter.sizeErrs[path]; err != nil {
		return 0, err
	}
	if kind == SymlinkItem {
		return 0, nil
	}
	return adapter.sizes[path], nil
}

func (adapter *recordingMoveAdapter) Rename(source, target string) error {
	if err := adapter.renameErrs[source]; err != nil {
		return err
	}
	adapter.renames[source] = target
	return nil
}

func (adapter *recordingMoveAdapter) RemoveEmptyFolder(path string) error {
	adapter.removed[path] = true
	return nil
}

func (adapter *recordingMoveAdapter) JoinPath(base, name string) string {
	return filepath.Join(base, name)
}
