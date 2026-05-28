package app

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"time"

	"shed/internal/core"
)

const (
	ExitOK = iota
	ExitError
)

type Options struct {
	Args      []string
	GOOS      string
	Stdout    io.Writer
	Stderr    io.Writer
	Resolver  SelectedFolderResolver
	Scanner   Scanner
	Mover     Mover
	Archiving ArchivingRunner
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

type MoveFunc func(ctx context.Context) (core.MoveSummary, error)

type ArchivingRunner interface {
	RunArchiving(ctx context.Context, request ArchivingRequest) (ArchivingResult, error)
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

	if opts.GOOS != "windows" {
		fmt.Fprintln(opts.Stderr, "Unsupported platform")
		return ExitError
	}

	if len(opts.Args) > 1 {
		fmt.Fprintln(opts.Stderr, "Usage: shed [folder]")
		return ExitError
	}

	arg := ""
	if len(opts.Args) == 1 {
		arg = opts.Args[0]
	}

	selectedFolder, err := opts.Resolver.Resolve(arg)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "Invalid selected folder: %v\n", err)
		return ExitError
	}

	result, err := opts.Scanner.Scan(ctx, selectedFolder)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "Scan failed: %v\n", err)
		return ExitError
	}

	if len(result.StaleItems) == 0 {
		fmt.Fprintln(opts.Stdout, "Nothing to move")
		for _, skipped := range result.SkippedItems {
			fmt.Fprintf(opts.Stdout, "Skipped item: %s\n", skipped.Path)
		}
		return ExitOK
	}

	confirmation := ConfirmationRequest{
		SelectedFolder:       selectedFolder,
		HeaderTitle:          core.HeaderTitle(selectedFolder),
		CompactArchiveBucket: core.CompactArchiveBucket(opts.Now(), selectedFolder),
		ScanResult:           result,
	}
	archiving, err := opts.Archiving.RunArchiving(ctx, ArchivingRequest{
		Confirmation: confirmation,
		Move: func(ctx context.Context) (core.MoveSummary, error) {
			return opts.Mover.Move(ctx, selectedFolder, result)
		},
		View: MoveViewData{
			SkippedItems: result.SkippedItems,
		},
	})
	if err != nil {
		if archiving.Outcome == ArchivingCompleted {
			fmt.Fprintf(opts.Stderr, "Preflight failure: %v\n", err)
			return ExitError
		}
		fmt.Fprintf(opts.Stderr, "Archiving failed: %v\n", err)
		return ExitError
	}
	if archiving.Outcome == ArchivingCancelled {
		return ExitOK
	}

	if len(archiving.Summary.FailedPaths) > 0 {
		return ExitError
	}
	return ExitOK
}

func withDefaults(opts Options) Options {
	if opts.GOOS == "" {
		opts.GOOS = runtime.GOOS
	}
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
	if opts.Archiving == nil {
		opts.Archiving = passthroughArchivingRunner{}
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
