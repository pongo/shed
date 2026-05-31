package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"shed/internal/core"
)

const (
	ExitOK = iota
	ExitError
)

type Options struct {
	Args             []string
	InvocationFolder string
	ShedRoot         string
	Stdout           io.Writer
	Stderr           io.Writer
	Resolver         SelectedFolderResolver
	Pruner           Pruner
	Pruning          PruningRunner
	Scanner          Scanner
	Mover            Mover
	Shedding         SheddingRunner
	Final            FinalRunner
	Now              func() time.Time
}

type SelectedFolderResolver interface {
	Resolve(arg string) (string, error)
}

type Scanner interface {
	Scan(ctx context.Context, selectedFolder string) (core.ScanResult, error)
}

type Mover interface {
	Move(ctx context.Context, selectedFolder string, planned core.PlannedShedBucket, scan core.ScanResult) (core.MoveSummary, error)
}
type Pruner interface {
	Scan(ctx context.Context) (core.PruneScanResult, error)
	Prune(ctx context.Context, scan core.PruneScanResult) (core.PruneSummary, error)
}

type MoveFunc func(ctx context.Context) (core.MoveSummary, error)
type PruneFunc func(ctx context.Context) (core.PruneSummary, error)

type SheddingRunner interface {
	RunShedding(ctx context.Context, request SheddingRequest) (SheddingResult, error)
}

type PruningRunner interface {
	RunPruning(ctx context.Context, request PruningRequest) (PruningResult, error)
}
type FinalRunner interface {
	RunFinal(ctx context.Context, request FinalSummaryRequest) error
}

type SheddingRequest struct {
	Confirmation ConfirmationRequest
	Move         MoveFunc
	View         MoveViewData
}

type SheddingResult struct {
	Outcome SheddingOutcome
	Summary core.MoveSummary
}

type PruningRequest struct {
	Scan  core.PruneScanResult
	Prune PruneFunc
}

type PruningResult struct {
	Outcome PruningOutcome
	Summary core.PruneSummary
}

type FinalSummaryRequest struct {
	Pruning  PruningFinalData
	Shedding SheddingFinalData
}

type PruningFinalData struct {
	HadCandidates bool
	Outcome       PruningOutcome
	Summary       core.PruneSummary
	Err           error
}

type SheddingFinalData struct {
	Show          bool
	NothingToMove bool
	Summary       core.MoveSummary
	SkippedItems  []core.SkippedItem
	Err           error
	scanFailed    bool
}

type MoveViewData struct {
	SkippedItems []core.SkippedItem
}

type ConfirmationRequest struct {
	SelectedFolder    string
	HeaderTitle       string
	CompactShedBucket string
	ScanResult        core.ScanResult
}

type SheddingOutcome int

const (
	SheddingCancelled SheddingOutcome = iota
	SheddingCompleted
)

type PruningOutcome int

const (
	PruningSkipped PruningOutcome = iota
	PruningConfirmed
	PruningQuit
)

type EmptyScanner struct{}

func (EmptyScanner) Scan(context.Context, string) (core.ScanResult, error) {
	return core.ScanResult{}, nil
}

type missingResolver struct{}

func (missingResolver) Resolve(string) (string, error) {
	return "", fmt.Errorf("selected folder resolver is not configured")
}

func Run(ctx context.Context, opts Options) int {
	opts = withDefaults(opts)

	arg := ""
	if len(opts.Args) == 1 {
		arg = opts.Args[0]
	}

	selectedFolder, err := opts.Resolver.Resolve(arg)
	if err != nil {
		_, _ = fmt.Fprintf(opts.Stderr, "Invalid selected folder: %v\n", err)
		return ExitError
	}

	pruning, keepRunning := runPruningPhase(ctx, opts)
	if !keepRunning {
		return ExitOK
	}

	invocationFolder := opts.InvocationFolder
	if invocationFolder == "" {
		invocationFolder = selectedFolder
	}

	shedding := runSheddingPhase(ctx, opts, invocationFolder, selectedFolder, pruning)

	return runFinalPhase(ctx, opts, pruning, shedding)
}

func withDefaults(opts Options) Options {
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	if opts.Scanner == nil {
		opts.Scanner = EmptyScanner{}
	}
	if opts.Resolver == nil {
		opts.Resolver = missingResolver{}
	}
	if opts.Mover == nil {
		opts.Mover = missingMover{}
	}
	if opts.Pruner == nil {
		opts.Pruner = emptyPruner{}
	}
	if opts.Pruning == nil {
		opts.Pruning = passthroughPruningRunner{}
	}
	if opts.Shedding == nil {
		opts.Shedding = passthroughSheddingRunner{}
	}
	if opts.Final == nil {
		opts.Final = noopFinalRunner{}
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	return opts
}

type missingMover struct{}

func (missingMover) Move(context.Context, string, core.PlannedShedBucket, core.ScanResult) (core.MoveSummary, error) {
	return core.MoveSummary{}, fmt.Errorf("mover is not configured")
}

type passthroughSheddingRunner struct{}

func (passthroughSheddingRunner) RunShedding(ctx context.Context, request SheddingRequest) (SheddingResult, error) {
	summary, err := request.Move(ctx)
	return SheddingResult{
		Outcome: SheddingCompleted,
		Summary: summary,
	}, err
}

type passthroughPruningRunner struct{}

func (passthroughPruningRunner) RunPruning(ctx context.Context, request PruningRequest) (PruningResult, error) {
	summary, err := request.Prune(ctx)
	return PruningResult{
		Outcome: PruningConfirmed,
		Summary: summary,
	}, err
}

type emptyPruner struct{}

func (emptyPruner) Scan(context.Context) (core.PruneScanResult, error) {
	return core.PruneScanResult{}, nil
}

func (emptyPruner) Prune(context.Context, core.PruneScanResult) (core.PruneSummary, error) {
	return core.PruneSummary{}, nil
}

type noopFinalRunner struct{}

func (noopFinalRunner) RunFinal(context.Context, FinalSummaryRequest) error {
	return nil
}
