package fs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"shed/internal/core"
)

func TestPrunerScanMissingArchiveRootIsNoOpAndDoesNotCreateIt(t *testing.T) {
	archiveRoot := filepath.Join(t.TempDir(), "Shed")
	pruner := Pruner{
		ArchiveRoot: archiveRoot,
		Now: func() time.Time {
			return time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		},
	}

	result, err := pruner.Scan(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(result.Candidates))
	}
	if _, err := os.Stat(archiveRoot); !os.IsNotExist(err) {
		t.Fatalf("expected Archive root to stay missing, stat err: %v", err)
	}
}

func TestPrunerScanFindsEligibleMonthsWithCalculatedSizes(t *testing.T) {
	now := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	archiveRoot := t.TempDir()

	createDir(t, filepath.Join(archiveRoot, "2025", "10"))
	createDir(t, filepath.Join(archiveRoot, "2025", "11"))
	createDir(t, filepath.Join(archiveRoot, "2026", "01"))

	sizes := map[string]int64{
		filepath.Join(archiveRoot, "2025", "10"): 100,
		filepath.Join(archiveRoot, "2025", "11"): 110,
		filepath.Join(archiveRoot, "2026", "01"): 10,
	}

	pruner := Pruner{
		ArchiveRoot: archiveRoot,
		Now:         func() time.Time { return now },
		Size: func(path string) (int64, error) {
			return sizes[path], nil
		},
	}

	result, err := pruner.Scan(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("expected one eligible candidate, got %d", len(result.Candidates))
	}
	if result.Candidates[0].Month.Path != filepath.Join("~", "Shed", "2025", "10") {
		t.Fatalf("unexpected candidate path: %s", result.Candidates[0].Month.Path)
	}
	if result.Candidates[0].Size != 100 {
		t.Fatalf("unexpected candidate size: %d", result.Candidates[0].Size)
	}
}

func TestPrunerScanIgnoresInvalidArchiveStructure(t *testing.T) {
	now := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	archiveRoot := t.TempDir()

	createDir(t, filepath.Join(archiveRoot, "tmp"))
	createDir(t, filepath.Join(archiveRoot, "2025", "ab"))
	createDir(t, filepath.Join(archiveRoot, "2025", "13"))
	createDir(t, filepath.Join(archiveRoot, "2025", "1"))
	createDir(t, filepath.Join(archiveRoot, "2025", "09", "nested"))
	writePrunerFile(t, filepath.Join(archiveRoot, "2025", "07.txt"), "not a month directory")

	pruner := Pruner{
		ArchiveRoot: archiveRoot,
		Now:         func() time.Time { return now },
		Size: func(path string) (int64, error) {
			if path == filepath.Join(archiveRoot, "2025", "09") {
				return 9, nil
			}
			return 0, nil
		},
	}

	result, err := pruner.Scan(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("expected one candidate, got %d", len(result.Candidates))
	}
	if result.Candidates[0].Month.Path != filepath.Join("~", "Shed", "2025", "09") {
		t.Fatalf("unexpected candidate path: %s", result.Candidates[0].Month.Path)
	}
}

func TestPrunerScanReportsSizeFailures(t *testing.T) {
	now := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	archiveRoot := t.TempDir()
	bad := filepath.Join(archiveRoot, "2025", "10")
	good := filepath.Join(archiveRoot, "2025", "09")

	createDir(t, bad)
	createDir(t, good)

	pruner := Pruner{
		ArchiveRoot: archiveRoot,
		Now:         func() time.Time { return now },
		Size: func(path string) (int64, error) {
			if path == bad {
				return 0, errors.New("boom")
			}
			return 9, nil
		},
	}

	result, err := pruner.Scan(context.Background())
	if err == nil {
		t.Fatalf("expected scan error")
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("expected one candidate from successful size calc, got %d", len(result.Candidates))
	}
}

func TestPrunerPruneCallsRecycleBinPerCandidateAndContinuesAfterFailures(t *testing.T) {
	archiveRoot := t.TempDir()

	scan := core.PruneScanResult{
		Candidates: []core.PruneCandidate{
			{Month: core.ArchiveMonth{Path: filepath.Join("~", "Shed", "2024", "01"), Year: 2024, Month: 1}, Size: 10},
			{Month: core.ArchiveMonth{Path: filepath.Join("~", "Shed", "2024", "02"), Year: 2024, Month: 2}, Size: 20},
		},
	}

	var calls []string
	pruner := Pruner{
		ArchiveRoot: archiveRoot,
		ReadDir: func(string) ([]os.DirEntry, error) {
			return []os.DirEntry{fakeDirEntry{name: "x", dir: true}}, nil
		},
		Throw: func(path string) error {
			calls = append(calls, path)
			if path == filepath.Join(archiveRoot, "2024", "01") {
				return errors.New("locked")
			}
			return nil
		},
	}

	summary, err := pruner.Prune(context.Background(), scan)
	if err != nil {
		t.Fatalf("expected no prune error, got %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 recycle calls, got %d", len(calls))
	}
	if !slices.Equal(summary.FailedPaths, []string{filepath.Join("~", "Shed", "2024", "01")}) {
		t.Fatalf("unexpected failed paths: %v", summary.FailedPaths)
	}
	if !slices.Equal(summary.PrunedPaths, []string{filepath.Join("~", "Shed", "2024", "02")}) {
		t.Fatalf("unexpected pruned paths: %v", summary.PrunedPaths)
	}
	if summary.PrunedSize != 20 {
		t.Fatalf("expected pruned size 20, got %d", summary.PrunedSize)
	}
}

func TestPrunerPruneCleansUpEmptyYearBestEffortWithoutSummaryItem(t *testing.T) {
	archiveRoot := t.TempDir()

	scan := core.PruneScanResult{
		Candidates: []core.PruneCandidate{
			{Month: core.ArchiveMonth{Path: filepath.Join("~", "Shed", "2024", "01"), Year: 2024, Month: 1}, Size: 10},
		},
	}

	monthPath := filepath.Join(archiveRoot, "2024", "01")
	yearPath := filepath.Join(archiveRoot, "2024")
	var calls []string

	pruner := Pruner{
		ArchiveRoot: archiveRoot,
		ReadDir: func(path string) ([]os.DirEntry, error) {
			if path == yearPath {
				return nil, nil
			}
			return []os.DirEntry{fakeDirEntry{name: "child", dir: true}}, nil
		},
		Throw: func(path string) error {
			calls = append(calls, path)
			if path == yearPath {
				return errors.New("ignored cleanup failure")
			}
			return nil
		},
	}

	summary, err := pruner.Prune(context.Background(), scan)
	if err != nil {
		t.Fatalf("expected no prune error, got %v", err)
	}
	if !slices.Equal(calls, []string{monthPath, yearPath}) {
		t.Fatalf("unexpected recycle calls: %v", calls)
	}
	if !slices.Equal(summary.PrunedPaths, []string{filepath.Join("~", "Shed", "2024", "01")}) {
		t.Fatalf("unexpected pruned paths: %v", summary.PrunedPaths)
	}
	if len(summary.FailedPaths) != 0 {
		t.Fatalf("cleanup failure must not appear in failed paths: %v", summary.FailedPaths)
	}
}

type fakeDirEntry struct {
	name string
	dir  bool
}

func (entry fakeDirEntry) Name() string               { return entry.name }
func (entry fakeDirEntry) IsDir() bool                { return entry.dir }
func (entry fakeDirEntry) Type() os.FileMode          { return 0 }
func (entry fakeDirEntry) Info() (os.FileInfo, error) { return nil, nil }

func createDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
}

func writePrunerFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
