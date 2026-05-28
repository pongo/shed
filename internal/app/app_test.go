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
	if !strings.Contains(stderr.String(), "usage: shed [folder]") {
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
	if !strings.Contains(stderr.String(), "invalid selected folder") {
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
	if !strings.Contains(stderr.String(), "scan failed") {
		t.Fatalf("expected scan failure, got %q", stderr.String())
	}
}

type recordingScanner struct {
	called         bool
	selectedFolder string
	err            error
}

func (s *recordingScanner) Scan(_ context.Context, selectedFolder string) (ScanResult, error) {
	s.called = true
	s.selectedFolder = selectedFolder
	if s.err != nil {
		return ScanResult{}, s.err
	}
	return ScanResult{}, nil
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
