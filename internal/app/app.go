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
	Args      []string
	Stdout    io.Writer
	Stderr    io.Writer
	Resolver  SelectedFolderResolver
	Pruner    Pruner
	Pruning   PruningRunner
	Scanner   Scanner
	Mover     Mover
	Archiving ArchivingRunner
	Final     FinalRunner
	Now       func() time.Time
}

type SelectedFolderResolver interface {
	Resolve(arg string) (string, error)
}

type Scanner interface {
	Scan(ctx context.Context, selectedFolder string) (core.ScanResult, error)
}

type Mover interface {
	Move(ctx context.Context, selectedFolder string, scan core.ScanResult) (core.MoveSummary, error)
}
type Pruner interface {
	Scan(ctx context.Context) (core.PruneScanResult, error)
	Prune(ctx context.Context, scan core.PruneScanResult) (core.PruneSummary, error)
}

type MoveFunc func(ctx context.Context) (core.MoveSummary, error)
type PruneFunc func(ctx context.Context) (core.PruneSummary, error)

type ArchivingRunner interface {
	RunArchiving(ctx context.Context, request ArchivingRequest) (ArchivingResult, error)
}

type PruningRunner interface {
	RunPruning(ctx context.Context, request PruningRequest) (PruningResult, error)
}
type FinalRunner interface {
	RunFinal(ctx context.Context, request FinalSummaryRequest) error
}

type ArchivingRequest struct {
	Confirmation ConfirmationRequest
	Move         MoveFunc
	View         MoveViewData
}

type ArchivingResult struct {
	Outcome ArchivingOutcome
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
	Pruning   PruningFinalData
	Archiving ArchivingFinalData
}

type PruningFinalData struct {
	HadCandidates bool
	Outcome       PruningOutcome
	Summary       core.PruneSummary
	Err           error
}

type ArchivingFinalData struct {
	Show          bool
	NothingToMove bool
	Summary       core.MoveSummary
	SkippedItems  []core.SkippedItem
	Err           error
}

type MoveViewData struct {
	SkippedItems []core.SkippedItem
}

type ConfirmationRequest struct {
	SelectedFolder       string
	HeaderTitle          string
	CompactArchiveBucket string
	ScanResult           core.ScanResult
}

type ArchivingOutcome int

const (
	ArchivingCancelled ArchivingOutcome = iota
	ArchivingCompleted
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
		fmt.Fprintf(opts.Stderr, "Invalid selected folder: %v\n", err)
		return ExitError
	}

	pruneScan, pruneErr := opts.Pruner.Scan(ctx)
	pruning := PruningFinalData{
		HadCandidates: len(pruneScan.Candidates) > 0,
		Outcome:       PruningSkipped,
		Err:           pruneErr,
	}
	if pruning.HadCandidates {
		pruningResult, runErr := opts.Pruning.RunPruning(ctx, PruningRequest{
			Scan: pruneScan,
			Prune: func(ctx context.Context) (core.PruneSummary, error) {
				return opts.Pruner.Prune(ctx, pruneScan)
			},
		})
		pruning.Outcome = pruningResult.Outcome
		pruning.Summary = pruningResult.Summary
		if runErr != nil {
			pruning.Err = runErr
		}
		if pruning.Outcome == PruningQuit {
			return ExitOK
		}
	}
	result, archivingScanErr := opts.Scanner.Scan(ctx, selectedFolder)
	archiving := ArchivingFinalData{
		Show:         false,
		SkippedItems: result.SkippedItems,
		Err:          archivingScanErr,
	}
	if archivingScanErr == nil && len(result.StaleItems) == 0 {
		archiving.NothingToMove = true
		archiving.Show = hasPruningReport(pruning)
	}
	if archivingScanErr == nil && len(result.StaleItems) > 0 {
		archiving.Show = true
		confirmation := ConfirmationRequest{
			SelectedFolder:       selectedFolder,
			HeaderTitle:          core.HeaderTitle(selectedFolder),
			CompactArchiveBucket: core.CompactArchiveBucket(opts.Now(), selectedFolder),
			ScanResult:           result,
		}
		archivingResult, runErr := opts.Archiving.RunArchiving(ctx, ArchivingRequest{
			Confirmation: confirmation,
			Move: func(ctx context.Context) (core.MoveSummary, error) {
				return opts.Mover.Move(ctx, selectedFolder, result)
			},
			View: MoveViewData{
				SkippedItems: result.SkippedItems,
			},
		})
		archiving.Summary = archivingResult.Summary
		if runErr != nil {
			archiving.Err = runErr
		}
		if archivingResult.Outcome == ArchivingCancelled && !hasPruningReport(pruning) {
			return ExitOK
		}
		if archivingResult.Outcome == ArchivingCancelled {
			archiving.Show = false
		}
	}

	if archivingScanErr != nil && !hasPruningReport(pruning) {
		fmt.Fprintf(opts.Stderr, "Scan failed: %v\n", archivingScanErr)
		return ExitError
	}
	if archiving.NothingToMove && !hasPruningReport(pruning) {
		fmt.Fprintln(opts.Stdout, "Nothing to move")
		for _, skipped := range result.SkippedItems {
			fmt.Fprintf(opts.Stdout, "Skipped item: %s\n", skipped.Path)
		}
		return ExitOK
	}
	if hasReport(pruning, archiving) {
		_ = opts.Final.RunFinal(ctx, FinalSummaryRequest{Pruning: pruning, Archiving: archiving})
	}

	if shouldExitError(pruning, archiving) {
		return ExitError
	}
	return ExitOK
}

func hasPruningReport(pruning PruningFinalData) bool {
	if pruning.Err != nil {
		return true
	}
	if pruning.Outcome == PruningConfirmed {
		return true
	}
	return len(pruning.Summary.FailedPaths) > 0 || len(pruning.Summary.PrunedPaths) > 0 || pruning.Summary.PrunedSize > 0
}

func hasReport(pruning PruningFinalData, archiving ArchivingFinalData) bool {
	return hasPruningReport(pruning) || archiving.Show || archiving.Err != nil
}

func shouldExitError(pruning PruningFinalData, archiving ArchivingFinalData) bool {
	if pruning.Err != nil || len(pruning.Summary.FailedPaths) > 0 {
		return true
	}
	if archiving.Err != nil || len(archiving.Summary.FailedPaths) > 0 {
		return true
	}
	return false
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
	if opts.Archiving == nil {
		opts.Archiving = passthroughArchivingRunner{}
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

func (missingMover) Move(context.Context, string, core.ScanResult) (core.MoveSummary, error) {
	return core.MoveSummary{}, fmt.Errorf("mover is not configured")
}

type passthroughArchivingRunner struct{}

func (passthroughArchivingRunner) RunArchiving(ctx context.Context, request ArchivingRequest) (ArchivingResult, error) {
	summary, err := request.Move(ctx)
	return ArchivingResult{
		Outcome: ArchivingCompleted,
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
