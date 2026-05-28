package app

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"shed/internal/core"
)

func TestRunSelectsCurrentWorkingDirectoryWithoutArgs(t *testing.T) {
	cwd := t.TempDir()
	scanner := &recordingScanner{}
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	code := Run(context.Background(), Options{
		Args:    nil,
		GOOS:    "windows",
		Stdout:  stdout,
		Stderr:  stderr,
		Scanner: scanner,
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				cwd: true,
			},
		},
	})

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d; stderr=%q", ExitOK, code, stderr.String())
	}
	if scanner.selectedFolder != cwd {
		t.Fatalf("expected selected folder %q, got %q", cwd, scanner.selectedFolder)
	}
	if stdout.String() != "Nothing to move\n" {
		t.Fatalf("expected Nothing to move output, got %q", stdout.String())
	}
}

func TestRunSelectsCurrentWorkingDirectoryForDotArg(t *testing.T) {
	cwd := t.TempDir()
	scanner := &recordingScanner{}

	code := Run(context.Background(), Options{
		Args:    []string{"."},
		GOOS:    "windows",
		Stdout:  new(bytes.Buffer),
		Stderr:  new(bytes.Buffer),
		Scanner: scanner,
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				cwd: true,
			},
		},
	})

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d", ExitOK, code)
	}
	if scanner.selectedFolder != cwd {
		t.Fatalf("expected selected folder %q, got %q", cwd, scanner.selectedFolder)
	}
}

func TestRunSelectsProvidedFolder(t *testing.T) {
	cwd := t.TempDir()
	selected := filepath.Join(cwd, "Downloads")
	scanner := &recordingScanner{}

	code := Run(context.Background(), Options{
		Args:    []string{selected},
		GOOS:    "windows",
		Stdout:  new(bytes.Buffer),
		Stderr:  new(bytes.Buffer),
		Scanner: scanner,
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				selected: true,
			},
		},
	})

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d", ExitOK, code)
	}
	if scanner.selectedFolder != selected {
		t.Fatalf("expected selected folder %q, got %q", selected, scanner.selectedFolder)
	}
}

func TestRunReportsUsageErrorForTooManyArgs(t *testing.T) {
	scanner := &recordingScanner{}
	stderr := new(bytes.Buffer)

	code := Run(context.Background(), Options{
		Args:     []string{"one", "two"},
		GOOS:     "windows",
		Stdout:   new(bytes.Buffer),
		Stderr:   stderr,
		Scanner:  scanner,
		Resolver: fakeResolver{},
	})

	if code == ExitOK {
		t.Fatalf("expected non-zero exit code")
	}
	if scanner.called {
		t.Fatalf("scanner should not be called for usage errors")
	}
	if !strings.Contains(stderr.String(), "Usage: shed [folder]") {
		t.Fatalf("expected usage error, got %q", stderr.String())
	}
}

func TestRunReportsInvalidSelectedFolder(t *testing.T) {
	cwd := t.TempDir()
	missing := filepath.Join(cwd, "missing")
	scanner := &recordingScanner{}
	stderr := new(bytes.Buffer)

	code := Run(context.Background(), Options{
		Args:    []string{missing},
		GOOS:    "windows",
		Stdout:  new(bytes.Buffer),
		Stderr:  stderr,
		Scanner: scanner,
		Resolver: fakeResolver{
			cwd: cwd,
			errs: map[string]error{
				missing: fs.ErrNotExist,
			},
		},
	})

	if code == ExitOK {
		t.Fatalf("expected non-zero exit code")
	}
	if scanner.called {
		t.Fatalf("scanner should not be called for invalid selected folders")
	}
	if !strings.Contains(stderr.String(), "Invalid selected folder") {
		t.Fatalf("expected invalid selected folder error, got %q", stderr.String())
	}
}

func TestRunReportsUnsupportedPlatformBeforeScanning(t *testing.T) {
	scanner := &recordingScanner{}
	stderr := new(bytes.Buffer)

	code := Run(context.Background(), Options{
		Args:     nil,
		GOOS:     "linux",
		Stdout:   new(bytes.Buffer),
		Stderr:   stderr,
		Scanner:  scanner,
		Resolver: fakeResolver{},
	})

	if code == ExitOK {
		t.Fatalf("expected non-zero exit code")
	}
	if scanner.called {
		t.Fatalf("scanner should not be called on unsupported platforms")
	}
	if !strings.Contains(stderr.String(), "Unsupported platform") {
		t.Fatalf("expected unsupported platform error, got %q", stderr.String())
	}
}

func TestRunReportsScannerFailure(t *testing.T) {
	cwd := t.TempDir()
	stderr := new(bytes.Buffer)

	code := Run(context.Background(), Options{
		GOOS:   "windows",
		Stdout: new(bytes.Buffer),
		Stderr: stderr,
		Scanner: &recordingScanner{
			err: errors.New("scan failed"),
		},
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				cwd: true,
			},
		},
	})

	if code == ExitOK {
		t.Fatalf("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "Scan failed") {
		t.Fatalf("expected scan failure, got %q", stderr.String())
	}
}

func TestRunPrintsSkippedPathsWhenNothingCanMove(t *testing.T) {
	cwd := t.TempDir()
	stdout := new(bytes.Buffer)

	code := Run(context.Background(), Options{
		GOOS:   "windows",
		Stdout: stdout,
		Stderr: new(bytes.Buffer),
		Scanner: &recordingScanner{
			result: core.ScanResult{
				SkippedItems: []core.SkippedItem{{Path: filepath.Join(cwd, "unreadable")}},
			},
		},
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				cwd: true,
			},
		},
	})

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d", ExitOK, code)
	}
	if !strings.Contains(stdout.String(), filepath.Join(cwd, "unreadable")) {
		t.Fatalf("expected skipped path in output, got %q", stdout.String())
	}
}

func TestRunPassesPopulatedScanResultToArchivingRunner(t *testing.T) {
	cwd := t.TempDir()
	moveDate := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	archiving := &recordingArchivingRunner{outcome: ArchivingCompleted}
	result := core.ScanResult{
		StaleItems: []core.StaleItem{{DisplayName: "old.txt", Kind: core.FileItem, MoveSize: 10}},
		MoveSize:   10,
	}

	code := Run(context.Background(), Options{
		GOOS:      "windows",
		Stdout:    new(bytes.Buffer),
		Stderr:    new(bytes.Buffer),
		Scanner:   &recordingScanner{result: result},
		Archiving: archiving,
		Mover:     &recordingMover{summary: core.MoveSummary{ArchiveBucket: filepath.Join(cwd, "Shed", "2026", "05", filepath.Base(cwd))}},
		Now:       func() time.Time { return moveDate },
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				cwd: true,
			},
		},
	})

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d", ExitOK, code)
	}
	if !archiving.called {
		t.Fatalf("expected archiving runner to be called")
	}
	if archiving.request.Confirmation.SelectedFolder != cwd {
		t.Fatalf("expected selected folder %q, got %q", cwd, archiving.request.Confirmation.SelectedFolder)
	}
	if archiving.request.Confirmation.HeaderTitle != filepath.Base(cwd) {
		t.Fatalf("expected header title %q, got %q", filepath.Base(cwd), archiving.request.Confirmation.HeaderTitle)
	}
	if !strings.Contains(archiving.request.Confirmation.CompactArchiveBucket, filepath.Join("~", "Shed", "2026", "05")) {
		t.Fatalf("expected compact archive bucket with move date, got %q", archiving.request.Confirmation.CompactArchiveBucket)
	}
}

func TestRunLeavesCancelledOutputToArchivingRunner(t *testing.T) {
	cwd := t.TempDir()
	stdout := new(bytes.Buffer)

	code := Run(context.Background(), Options{
		GOOS:      "windows",
		Stdout:    stdout,
		Stderr:    new(bytes.Buffer),
		Scanner:   &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}},
		Archiving: &recordingArchivingRunner{outcome: ArchivingCancelled},
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				cwd: true,
			},
		},
	})

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d", ExitOK, code)
	}
	if stdout.String() != "" {
		t.Fatalf("expected cancelled output to be owned by archiving runner, got %q", stdout.String())
	}
}

func TestRunMovesAfterConfirmationAndPassesViewData(t *testing.T) {
	cwd := t.TempDir()
	stdout := new(bytes.Buffer)
	archiving := &recordingArchivingRunner{outcome: ArchivingCompleted}
	skippedPath := filepath.Join(cwd, "unreadable")
	result := core.ScanResult{
		StaleItems:   []core.StaleItem{{DisplayName: "old.txt", Path: filepath.Join(cwd, "old.txt"), MoveSize: 10}},
		SkippedItems: []core.SkippedItem{{Path: skippedPath}},
		MoveSize:     10,
	}
	bucket := filepath.Join(cwd, "Shed", "2026", "05", filepath.Base(cwd))
	mover := &recordingMover{summary: core.MoveSummary{ArchiveBucket: bucket, MovedSize: 10}}

	code := Run(context.Background(), Options{
		GOOS:      "windows",
		Stdout:    stdout,
		Stderr:    new(bytes.Buffer),
		Scanner:   &recordingScanner{result: result},
		Mover:     mover,
		Archiving: archiving,
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				cwd: true,
			},
		},
	})

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d", ExitOK, code)
	}
	if !mover.called {
		t.Fatalf("expected mover to be called")
	}
	if !archiving.called {
		t.Fatalf("expected archiving runner to be called")
	}
	if stdout.String() != "" {
		t.Fatalf("expected final move summary to be owned by archiving runner, got stdout %q", stdout.String())
	}
	if len(archiving.request.View.SkippedItems) != 1 || archiving.request.View.SkippedItems[0].Path != skippedPath {
		t.Fatalf("expected skipped items to be passed to archiving runner, got %#v", archiving.request.View.SkippedItems)
	}
}

func TestRunReportsFailedMovesAndExitsNonZero(t *testing.T) {
	cwd := t.TempDir()
	failedPath := filepath.Join(cwd, "locked.txt")

	code := Run(context.Background(), Options{
		GOOS:    "windows",
		Stdout:  new(bytes.Buffer),
		Stderr:  new(bytes.Buffer),
		Scanner: &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "locked.txt", Path: failedPath}}}},
		Mover: &recordingMover{summary: core.MoveSummary{
			ArchiveBucket: filepath.Join(cwd, "Shed"),
			FailedPaths:   []string{failedPath},
		}},
		Archiving: &recordingArchivingRunner{outcome: ArchivingCompleted},
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				cwd: true,
			},
		},
	})

	if code == ExitOK {
		t.Fatalf("expected non-zero exit code")
	}
}

func TestRunReportsPreflightFailureBeforeSummary(t *testing.T) {
	cwd := t.TempDir()
	stderr := new(bytes.Buffer)

	code := Run(context.Background(), Options{
		GOOS:      "windows",
		Stdout:    new(bytes.Buffer),
		Stderr:    stderr,
		Scanner:   &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}},
		Mover:     &recordingMover{err: errors.New("selected folder unavailable")},
		Archiving: &recordingArchivingRunner{outcome: ArchivingCompleted},
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				cwd: true,
			},
		},
	})

	if code == ExitOK {
		t.Fatalf("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "Preflight failure") {
		t.Fatalf("expected preflight failure, got %q", stderr.String())
	}
}

func TestRunPassesSkippedPathsAfterConfirmedRun(t *testing.T) {
	cwd := t.TempDir()
	stdout := new(bytes.Buffer)
	archiving := &recordingArchivingRunner{outcome: ArchivingCompleted}
	skippedPath := filepath.Join(cwd, "unreadable")

	code := Run(context.Background(), Options{
		GOOS:   "windows",
		Stdout: stdout,
		Stderr: new(bytes.Buffer),
		Scanner: &recordingScanner{result: core.ScanResult{
			StaleItems:   []core.StaleItem{{DisplayName: "old.txt"}},
			SkippedItems: []core.SkippedItem{{Path: skippedPath}},
		}},
		Mover:     &recordingMover{summary: core.MoveSummary{ArchiveBucket: filepath.Join(cwd, "Shed")}},
		Archiving: archiving,
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				cwd: true,
			},
		},
	})

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d", ExitOK, code)
	}
	if stdout.String() != "" {
		t.Fatalf("expected final move summary to be owned by archiving runner, got stdout %q", stdout.String())
	}
	if len(archiving.request.View.SkippedItems) != 1 || archiving.request.View.SkippedItems[0].Path != skippedPath {
		t.Fatalf("expected skipped item to be passed to archiving runner, got %#v", archiving.request.View.SkippedItems)
	}
}

func TestRunReportsArchivingRunnerFailure(t *testing.T) {
	cwd := t.TempDir()
	stderr := new(bytes.Buffer)

	code := Run(context.Background(), Options{
		GOOS:      "windows",
		Stdout:    new(bytes.Buffer),
		Stderr:    stderr,
		Scanner:   &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}},
		Archiving: &recordingArchivingRunner{err: errors.New("terminal unavailable")},
		Resolver: fakeResolver{
			cwd: cwd,
			dirs: map[string]bool{
				cwd: true,
			},
		},
	})

	if code == ExitOK {
		t.Fatalf("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "terminal unavailable") {
		t.Fatalf("expected confirmation failure, got %q", stderr.String())
	}
}

type recordingScanner struct {
	called         bool
	selectedFolder string
	err            error
	result         core.ScanResult
}

func (s *recordingScanner) Scan(_ context.Context, selectedFolder string) (core.ScanResult, error) {
	s.called = true
	s.selectedFolder = selectedFolder
	if s.err != nil {
		return core.ScanResult{}, s.err
	}
	return s.result, nil
}

type recordingMover struct {
	called         bool
	selectedFolder string
	scan           core.ScanResult
	summary        core.MoveSummary
	err            error
}

func (m *recordingMover) Move(_ context.Context, selectedFolder string, scan core.ScanResult) (core.MoveSummary, error) {
	m.called = true
	m.selectedFolder = selectedFolder
	m.scan = scan
	if m.err != nil {
		return core.MoveSummary{}, m.err
	}
	return m.summary, nil
}

type recordingArchivingRunner struct {
	called  bool
	request ArchivingRequest
	outcome ArchivingOutcome
	err     error
}

func (runner *recordingArchivingRunner) RunArchiving(ctx context.Context, request ArchivingRequest) (ArchivingResult, error) {
	runner.called = true
	runner.request = request
	if runner.err != nil {
		return ArchivingResult{}, runner.err
	}
	if runner.outcome == ArchivingCancelled {
		return ArchivingResult{Outcome: ArchivingCancelled}, nil
	}
	summary, err := request.Move(ctx)
	return ArchivingResult{
		Outcome: ArchivingCompleted,
		Summary: summary,
	}, err
}

type fakeResolver struct {
	cwd  string
	dirs map[string]bool
	errs map[string]error
}

func (r fakeResolver) Resolve(arg string) (string, error) {
	selected := arg
	if arg == "" || arg == "." {
		selected = r.cwd
	}
	if err, ok := r.errs[selected]; ok {
		return "", err
	}
	if !r.dirs[selected] {
		return "", os.ErrNotExist
	}
	return selected, nil
}
