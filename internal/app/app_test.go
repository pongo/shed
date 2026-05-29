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

func TestRunPruningNoOpFlowsToShedding(t *testing.T) {
	cwd := t.TempDir()
	pruner := &recordingPruner{}
	shedding := &recordingSheddingRunner{outcome: SheddingCompleted}
	scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}}

	code := Run(context.Background(), baseOptions(cwd, pruner, scanner, shedding, &recordingFinalRunner{}))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if !pruner.scanCalled || !scanner.called || !shedding.called {
		t.Fatalf("expected pruning scan then shedding path to run")
	}
}

func TestRunPruningSkipFlowsToShedding(t *testing.T) {
	cwd := t.TempDir()
	pruner := &recordingPruner{scan: core.PruneScanResult{Candidates: []core.PruneCandidate{{Month: core.ShedMonth{Path: "m"}}}}}
	pruning := &recordingPruningRunner{result: PruningResult{Outcome: PruningSkipped}}
	scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}}
	shedding := &recordingSheddingRunner{outcome: SheddingCompleted}

	code := Run(context.Background(), baseOptions(cwd, pruner, scanner, shedding, &recordingFinalRunner{}, withPruning(pruning)))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if !shedding.called {
		t.Fatalf("expected shedding after pruning skip")
	}
}

func TestRunPruningQuitStopsWithoutSheddingAndFinal(t *testing.T) {
	cwd := t.TempDir()
	pruner := &recordingPruner{scan: core.PruneScanResult{Candidates: []core.PruneCandidate{{Month: core.ShedMonth{Path: "m"}}}}}
	pruning := &recordingPruningRunner{result: PruningResult{Outcome: PruningQuit}}
	scanner := &recordingScanner{}
	shedding := &recordingSheddingRunner{}
	final := &recordingFinalRunner{}

	code := Run(context.Background(), baseOptions(cwd, pruner, scanner, shedding, final, withPruning(pruning)))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if scanner.called || shedding.called || final.called {
		t.Fatalf("expected hard quit to skip shedding and final summary")
	}
}

func TestRunPruningErrorDoesNotBlockSheddingAndExitsError(t *testing.T) {
	cwd := t.TempDir()
	pruner := &recordingPruner{scanErr: errors.New("scan denied")}
	scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}}
	shedding := &recordingSheddingRunner{outcome: SheddingCompleted}
	final := &recordingFinalRunner{}

	code := Run(context.Background(), baseOptions(cwd, pruner, scanner, shedding, final))
	if code != ExitError {
		t.Fatalf("expected ExitError, got %d", code)
	}
	if !shedding.called {
		t.Fatalf("expected shedding to continue after pruning error")
	}
	if !final.called || final.request.Pruning.Err == nil {
		t.Fatalf("expected final summary with pruning error")
	}
}

func TestRunSheddingScanErrorWithoutPruningPrintsSimpleError(t *testing.T) {
	cwd := t.TempDir()
	stderr := new(bytes.Buffer)
	final := &recordingFinalRunner{}
	opts := baseOptions(cwd, &recordingPruner{}, &recordingScanner{err: errors.New("scan denied")}, &recordingSheddingRunner{}, final, withStderr(stderr))

	code := Run(context.Background(), opts)
	if code != ExitError {
		t.Fatalf("expected ExitError, got %d", code)
	}
	if final.called {
		t.Fatalf("expected no final summary for plain scan error")
	}
	if !strings.Contains(stderr.String(), "Scan failed: scan denied") {
		t.Fatalf("expected simple scan error output, got %q", stderr.String())
	}
}

func TestRunSheddingMoveErrorUsesFinalSummary(t *testing.T) {
	cwd := t.TempDir()
	stderr := new(bytes.Buffer)
	scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}}
	shedding := &recordingSheddingRunner{outcome: SheddingCompleted, err: errors.New("move failed")}
	final := &recordingFinalRunner{}
	opts := baseOptions(cwd, &recordingPruner{}, scanner, shedding, final, withStderr(stderr))

	code := Run(context.Background(), opts)
	if code != ExitError {
		t.Fatalf("expected ExitError, got %d", code)
	}
	if !final.called || final.request.Shedding.Err == nil {
		t.Fatalf("expected final summary with shedding move error")
	}
	if strings.Contains(stderr.String(), "Scan failed") {
		t.Fatalf("expected no simple scan error output for move error, got %q", stderr.String())
	}
}

func TestRunSheddingQuitAfterPruningShowsPruningSummary(t *testing.T) {
	cwd := t.TempDir()
	pruner := &recordingPruner{scan: core.PruneScanResult{Candidates: []core.PruneCandidate{{Month: core.ShedMonth{Path: "m"}}}}}
	pruning := &recordingPruningRunner{result: PruningResult{Outcome: PruningConfirmed, Summary: core.PruneSummary{PrunedSize: 1, PrunedPaths: []string{"m"}}}}
	scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old.txt"}}}}
	shedding := &recordingSheddingRunner{outcome: SheddingCancelled}
	final := &recordingFinalRunner{}

	code := Run(context.Background(), baseOptions(cwd, pruner, scanner, shedding, final, withPruning(pruning)))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if !final.called {
		t.Fatalf("expected final summary after shedding quit")
	}
	if !final.request.Pruning.HadCandidates {
		t.Fatalf("expected pruning summary data")
	}
}

func TestRunKeepsSimpleNothingToMoveWithoutPruningOutput(t *testing.T) {
	cwd := t.TempDir()
	stdout := new(bytes.Buffer)
	final := &recordingFinalRunner{}

	code := Run(context.Background(), baseOptions(cwd, &recordingPruner{}, &recordingScanner{}, &recordingSheddingRunner{}, final, withStdout(stdout)))
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
	pruner := &recordingPruner{scan: core.PruneScanResult{Candidates: []core.PruneCandidate{{Month: core.ShedMonth{Path: "m"}}}}}
	pruning := &recordingPruningRunner{result: PruningResult{Outcome: PruningConfirmed, Summary: core.PruneSummary{PrunedSize: 1}}}
	final := &recordingFinalRunner{}
	stdout := new(bytes.Buffer)

	code := Run(context.Background(), baseOptions(cwd, pruner, &recordingScanner{}, &recordingSheddingRunner{}, final, withPruning(pruning), withStdout(stdout)))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if final.request.Shedding.NothingToMove != true || !final.request.Shedding.Show {
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
				pruner := &recordingPruner{scan: core.PruneScanResult{Candidates: []core.PruneCandidate{{Month: core.ShedMonth{Path: "m"}}}}}
				pruning := &recordingPruningRunner{result: PruningResult{Outcome: PruningConfirmed, Summary: core.PruneSummary{FailedPaths: []string{"m"}}}}
				return baseOptions(cwd, pruner, &recordingScanner{}, &recordingSheddingRunner{}, &recordingFinalRunner{}, withPruning(pruning))
			},
			want: ExitError,
		},
		{
			name: "shedding error",
			opts: func() Options {
				scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old"}}}}
				shedding := &recordingSheddingRunner{outcome: SheddingCompleted, err: errors.New("move failed")}
				return baseOptions(cwd, &recordingPruner{}, scanner, shedding, &recordingFinalRunner{})
			},
			want: ExitError,
		},
		{
			name: "failed move paths",
			opts: func() Options {
				scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old"}}}}
				shedding := &recordingSheddingRunner{outcome: SheddingCompleted, result: core.MoveSummary{FailedPaths: []string{"x"}}}
				return baseOptions(cwd, &recordingPruner{}, scanner, shedding, &recordingFinalRunner{})
			},
			want: ExitError,
		},
		{
			name: "declined shedding without failures",
			opts: func() Options {
				scanner := &recordingScanner{result: core.ScanResult{StaleItems: []core.StaleItem{{DisplayName: "old"}}}}
				shedding := &recordingSheddingRunner{outcome: SheddingCancelled}
				return baseOptions(cwd, &recordingPruner{}, scanner, shedding, &recordingFinalRunner{})
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

type recordingSheddingRunner struct {
	called  bool
	outcome SheddingOutcome
	result  core.MoveSummary
	err     error
}

func (r *recordingSheddingRunner) RunShedding(context.Context, SheddingRequest) (SheddingResult, error) {
	r.called = true
	return SheddingResult{Outcome: r.outcome, Summary: r.result}, r.err
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

func withStderr(stderr *bytes.Buffer) optionMutator {
	return func(opts *Options) { opts.Stderr = stderr }
}

func baseOptions(cwd string, pruner Pruner, scanner Scanner, shedding SheddingRunner, final FinalRunner, mutators ...optionMutator) Options {
	opts := Options{
		Stdout:   new(bytes.Buffer),
		Stderr:   new(bytes.Buffer),
		Resolver: fakeResolver{cwd: cwd},
		Pruner:   pruner,
		Scanner:  scanner,
		Mover:    &recordingMover{},
		Shedding: shedding,
		Final:    final,
	}
	for _, mutate := range mutators {
		mutate(&opts)
	}
	return opts
}

type recordingMover struct{}

func (recordingMover) Move(_ context.Context, selectedFolder string, _ core.ScanResult) (core.MoveSummary, error) {
	return core.MoveSummary{ShedBucket: filepath.Join(selectedFolder, "Shed")}, nil
}

func TestNothingToMoveStillIncludesSkippedInSimpleOutput(t *testing.T) {
	cwd := t.TempDir()
	stdout := new(bytes.Buffer)
	skipped := filepath.Join(cwd, "unreadable")
	code := Run(context.Background(), baseOptions(cwd, &recordingPruner{}, &recordingScanner{result: core.ScanResult{SkippedItems: []core.SkippedItem{{Path: skipped}}}}, &recordingSheddingRunner{}, &recordingFinalRunner{}, withStdout(stdout)))
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Nothing to move") || !strings.Contains(stdout.String(), skipped) {
		t.Fatalf("expected skipped path in simple output, got %q", stdout.String())
	}
}
