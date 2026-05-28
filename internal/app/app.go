package app

import (
	"context"
	"fmt"
	"io"
	"runtime"

	"shed/internal/core"
)

const (
	ExitOK = iota
	ExitError
)

type Options struct {
	Args     []string
	GOOS     string
	Stdout   io.Writer
	Stderr   io.Writer
	Resolver SelectedFolderResolver
	Scanner  Scanner
}

type SelectedFolderResolver interface {
	Resolve(arg string) (string, error)
}

type Scanner interface {
	Scan(ctx context.Context, selectedFolder string) (core.ScanResult, error)
}

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
		fmt.Fprintln(opts.Stderr, "usage: shed [folder]")
		return ExitError
	}

	arg := ""
	if len(opts.Args) == 1 {
		arg = opts.Args[0]
	}

	selectedFolder, err := opts.Resolver.Resolve(arg)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "invalid selected folder: %v\n", err)
		return ExitError
	}

	result, err := opts.Scanner.Scan(ctx, selectedFolder)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "scan failed: %v\n", err)
		return ExitError
	}

	if len(result.StaleItems) == 0 {
		fmt.Fprintln(opts.Stdout, "Nothing to move")
		for _, skipped := range result.SkippedItems {
			fmt.Fprintf(opts.Stdout, "Skipped: %s\n", skipped.Path)
		}
		return ExitOK
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
	return opts
}
