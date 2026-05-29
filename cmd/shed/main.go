package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"shed/internal/app"
	shedfs "shed/internal/fs"
	"shed/internal/tui/archiving"
	"shed/internal/tui/final"
	"shed/internal/tui/pruning"
)

func main() {
	os.Exit(run(
		context.Background(),
		os.Args[1:],
		os.Stdin,
		os.Stdout,
		os.Stderr,
		shedfs.UnsupportedPlatform,
		shedfs.UserArchiveRoot,
	))
}

func run(
	ctx context.Context,
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	unsupportedPlatform func() bool,
	userArchiveRoot func() (string, error),
) int {
	if unsupportedPlatform() {
		_, _ = fmt.Fprintln(stderr, "Unsupported platform")
		return app.ExitError
	}

	if len(args) > 1 {
		_, _ = fmt.Fprintln(stderr, "Usage: shed [folder]")
		return app.ExitError
	}

	archiveRoot, err := userArchiveRoot()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Archive path unavailable: %v\n", err)
		return app.ExitError
	}

	return app.Run(ctx, app.Options{
		Args:     args,
		Stdout:   stdout,
		Stderr:   stderr,
		Resolver: shedfs.SelectedFolderResolver{},
		Pruner:   shedfs.NewPruner(archiveRoot),
		Pruning: pruning.Runner{
			Input:  stdin,
			Output: stdout,
		},
		Scanner: shedfs.NewScanner(archiveRoot),
		Mover:   shedfs.NewMover(archiveRoot),
		Archiving: archiving.Runner{
			Input:  stdin,
			Output: stdout,
		},
		Final: final.Runner{
			Output: stdout,
		},
	})
}
