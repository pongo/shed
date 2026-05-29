package app

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"shed/internal/core"
)

func TestRunResolvesSelectedFolderBeforePruning(t *testing.T) {
	pruner := &recordingPruner{}
	stderr := new(bytes.Buffer)

	code := Run(context.Background(), Options{
		Stderr: stderr,
		Pruner: pruner,
		Resolver: fakeResolver{
			err: errors.New("bad folder"),
		},
	})
	if code != ExitError {
		t.Fatalf("expected ExitError, got %d", code)
	}
	if pruner.scanCalled {
		t.Fatalf("expected no pruning before selected folder resolve")
	}
	if !strings.Contains(stderr.String(), "Invalid selected folder: bad folder") {
		t.Fatalf("expected selected folder error, got %q", stderr.String())
	}
}

func TestRunPruningNoOpFlowsToArchiving(t *testing.T) {
	cwd := t.TempDir()
	pruner := &recordingPruner{}
	archiving := &recordingArchivingRunner{outcome: ArchivingCompleted}
	scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}}

	code := Run(context.Background(), baseOptions(cwd, pruner, scanner, archiving, &recordingFinalRunner{}))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if !pruner.scanCalled || !scanner.called || !archiving.called {
		t.Fatalf("expected pruning scan then archiving path to run")
	}
}

func TestRunPruningSkipFlowsToArchiving(t *testing.T) {
	cwd := t.TempDir()
	pruner := &recordingPruner{scan: core.PruneScanResult{Candidates: []core.PruneCandidate{{Month: core.ArchiveMonth{Path: "m"}}}}}
	pruning := &recordingPruningRunner{result: PruningResult{Outcome: PruningSkipped}}
	scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}}
	archiving := &recordingArchivingRunner{outcome: ArchivingCompleted}

	code := Run(context.Background(), baseOptions(cwd, pruner, scanner, archiving, &recordingFinalRunner{}, withPruning(pruning)))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if !archiving.called {
		t.Fatalf("expected archiving after pruning skip")
	}
}

func TestRunPruningQuitStopsWithoutArchivingAndFinal(t *testing.T) {
	cwd := t.TempDir()
	pruner := &recordingPruner{scan: core.PruneScanResult{Candidates: []core.PruneCandidate{{Month: core.ArchiveMonth{Path: "m"}}}}}
	pruning := &recordingPruningRunner{result: PruningResult{Outcome: PruningQuit}}
	scanner := &recordingScanner{}
	archiving := &recordingArchivingRunner{}
	final := &recordingFinalRunner{}

	code := Run(context.Background(), baseOptions(cwd, pruner, scanner, archiving, final, withPruning(pruning)))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if scanner.called || archiving.called || final.called {
		t.Fatalf("expected hard quit to skip archiving and final summary")
	}
}

func TestRunPruningErrorDoesNotBlockArchivingAndExitsError(t *testing.T) {
	cwd := t.TempDir()
	pruner := &recordingPruner{scanErr: errors.New("scan denied")}
	scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}}
	archiving := &recordingArchivingRunner{outcome: ArchivingCompleted}
	final := &recordingFinalRunner{}

	code := Run(context.Background(), baseOptions(cwd, pruner, scanner, archiving, final))
	if code != ExitError {
		t.Fatalf("expected ExitError, got %d", code)
	}
	if !archiving.called {
		t.Fatalf("expected archiving to continue after pruning error")
	}
	if !final.called || final.request.Pruning.Err == nil {
		t.Fatalf("expected final summary with pruning error")
	}
}

func TestRunArchivingQuitAfterPruningShowsPruningSummary(t *testing.T) {
	cwd := t.TempDir()
	pruner := &recordingPruner{scan: core.PruneScanResult{Candidates: []core.PruneCandidate{{Month: core.ArchiveMonth{Path: "m"}}}}}
	pruning := &recordingPruningRunner{result: PruningResult{Outcome: PruningConfirmed, Summary: core.PruneSummary{PrunedSize: 1, PrunedPaths: []string{"m"}}}}
	scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}}
	archiving := &recordingArchivingRunner{outcome: ArchivingCancelled}
	final := &recordingFinalRunner{}

	code := Run(context.Background(), baseOptions(cwd, pruner, scanner, archiving, final, withPruning(pruning)))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if !final.called {
		t.Fatalf("expected final summary after archiving quit")
	}
	if !final.request.Pruning.HadCandidates {
		t.Fatalf("expected pruning summary data")
	}
}

func TestRunKeepsSimpleNothingToMoveWithoutPruningOutput(t *testing.T) {
	cwd := t.TempDir()
	stdout := new(bytes.Buffer)
	final := &recordingFinalRunner{}

	code := Run(context.Background(), baseOptions(cwd, &recordingPruner{}, &recordingScanner{}, &recordingArchivingRunner{}, final, withStdout(stdout)))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if stdout.String() != "Nothing to move\n" {
		t.Fatalf("expected simple Nothing to move output, got %q", stdout.String())
	}
	if final.called {
		t.Fatalf("expected no final summary for plain no-op")
	}
}

func TestRunShowsNothingToMoveInFinalWhenPruningHasOutput(t *testing.T) {
	cwd := t.TempDir()
	pruner := &recordingPruner{scan: core.PruneScanResult{Candidates: []core.PruneCandidate{{Month: core.ArchiveMonth{Path: "m"}}}}}
	pruning := &recordingPruningRunner{result: PruningResult{Outcome: PruningConfirmed, Summary: core.PruneSummary{PrunedSize: 1}}}
	final := &recordingFinalRunner{}
	stdout := new(bytes.Buffer)

	code := Run(context.Background(), baseOptions(cwd, pruner, &recordingScanner{}, &recordingArchivingRunner{}, final, withPruning(pruning), withStdout(stdout)))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if final.request.Archiving.NothingToMove != true || !final.request.Archiving.Show {
		t.Fatalf("expected NothingToMove to be reported in final summary")
	}
	if stdout.String() != "" {
		t.Fatalf("expected no direct no-op output when final summary is shown")
	}
}

func TestRunExitCodePolicy(t *testing.T) {
	cwd := t.TempDir()

	cases := []struct {
		name string
		opts func() Options
		want int
	}{
		{
			name: "failed prune paths",
			opts: func() Options {
				pruner := &recordingPruner{scan: core.PruneScanResult{Candidates: []core.PruneCandidate{{Month: core.ArchiveMonth{Path: "m"}}}}}
				pruning := &recordingPruningRunner{result: PruningResult{Outcome: PruningConfirmed, Summary: core.PruneSummary{FailedPaths: []string{"m"}}}}
				return baseOptions(cwd, pruner, &recordingScanner{}, &recordingArchivingRunner{}, &recordingFinalRunner{}, withPruning(pruning))
			},
			want: ExitError,
		},
		{
			name: "archiving error",
			opts: func() Options {
				scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old"}}}}
				archiving := &recordingArchivingRunner{outcome: ArchivingCompleted, err: errors.New("move failed")}
				return baseOptions(cwd, &recordingPruner{}, scanner, archiving, &recordingFinalRunner{})
			},
			want: ExitError,
		},
		{
			name: "failed move paths",
			opts: func() Options {
				scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old"}}}}
				archiving := &recordingArchivingRunner{outcome: ArchivingCompleted, result: core.MoveSummary{FailedPaths: []string{"x"}}}
				return baseOptions(cwd, &recordingPruner{}, scanner, archiving, &recordingFinalRunner{})
			},
			want: ExitError,
		},
		{
			name: "declined archiving without failures",
			opts: func() Options {
				scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old"}}}}
				archiving := &recordingArchivingRunner{outcome: ArchivingCancelled}
				return baseOptions(cwd, &recordingPruner{}, scanner, archiving, &recordingFinalRunner{})
			},
			want: ExitOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Run(context.Background(), tc.opts()); got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

type recordingPruner struct {
	scanCalled bool
	scan       core.PruneScanResult
	scanErr    error
}

func (p *recordingPruner) Scan(context.Context) (core.PruneScanResult, error) {
	p.scanCalled = true
	return p.scan, p.scanErr
}

func (p *recordingPruner) Prune(context.Context, core.PruneScanResult) (core.PruneSummary, error) {
	return core.PruneSummary{}, nil
}

type recordingPruningRunner struct {
	result PruningResult
	err    error
}

func (r *recordingPruningRunner) RunPruning(context.Context, PruningRequest) (PruningResult, error) {
	return r.result, r.err
}

type recordingScanner struct {
	called bool
	result core.ScanResult
	err    error
}

func (s *recordingScanner) Scan(context.Context, string) (core.ScanResult, error) {
	s.called = true
	return s.result, s.err
}

type recordingArchivingRunner struct {
	called  bool
	outcome ArchivingOutcome
	result  core.MoveSummary
	err     error
}

func (r *recordingArchivingRunner) RunArchiving(context.Context, ArchivingRequest) (ArchivingResult, error) {
	r.called = true
	return ArchivingResult{Outcome: r.outcome, Summary: r.result}, r.err
}

type recordingFinalRunner struct {
	called  bool
	request FinalSummaryRequest
}

func (r *recordingFinalRunner) RunFinal(_ context.Context, request FinalSummaryRequest) error {
	r.called = true
	r.request = request
	return nil
}

type fakeResolver struct {
	cwd string
	err error
}

func (r fakeResolver) Resolve(arg string) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	if arg == "" || arg == "." {
		return r.cwd, nil
	}
	return arg, nil
}

type optionMutator func(*Options)

func withPruning(pruning PruningRunner) optionMutator {
	return func(opts *Options) { opts.Pruning = pruning }
}

func withStdout(stdout *bytes.Buffer) optionMutator {
	return func(opts *Options) { opts.Stdout = stdout }
}

func baseOptions(cwd string, pruner Pruner, scanner Scanner, archiving ArchivingRunner, final FinalRunner, mutators ...optionMutator) Options {
	opts := Options{
		Stdout:    new(bytes.Buffer),
		Stderr:    new(bytes.Buffer),
		Resolver:  fakeResolver{cwd: cwd},
		Pruner:    pruner,
		Scanner:   scanner,
		Mover:     &recordingMover{},
		Archiving: archiving,
		Final:     final,
	}
	for _, mutate := range mutators {
		mutate(&opts)
	}
	return opts
}

type recordingMover struct{}

func (recordingMover) Move(_ context.Context, selectedFolder string, _ core.ScanResult) (core.MoveSummary, error) {
	return core.MoveSummary{ArchiveBucket: filepath.Join(selectedFolder, "Shed")}, nil
}

func TestNothingToMoveStillIncludesSkippedInSimpleOutput(t *testing.T) {
	cwd := t.TempDir()
	stdout := new(bytes.Buffer)
	skipped := filepath.Join(cwd, "unreadable")
	code := Run(context.Background(), baseOptions(cwd, &recordingPruner{}, &recordingScanner{result: core.ScanResult{SkippedItems: []core.SkippedItem{{Path: skipped}}}}, &recordingArchivingRunner{}, &recordingFinalRunner{}, withStdout(stdout)))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Nothing to move") || !strings.Contains(stdout.String(), skipped) {
		t.Fatalf("expected skipped path in simple output, got %q", stdout.String())
	}
}
