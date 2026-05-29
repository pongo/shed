package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"shed/internal/app"
	shedfs "shed/internal/fs"
	"shed/internal/tui/final"
	"shed/internal/tui/pruning"
	"shed/internal/tui/shedding"
)

func main() {
	os.Exit(run(
		context.Background(),
		os.Args[1:],
		os.Stdin,
		os.Stdout,
		os.Stderr,
		shedfs.UnsupportedPlatform,
		shedfs.UserShedRoot,
	))
}

func run(
	ctx context.Context,
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	unsupportedPlatform func() bool,
	userShedRoot func() (string, error),
) int {
	if unsupportedPlatform() {
		_, _ = fmt.Fprintln(stderr, "Unsupported platform")
		return app.ExitError
	}

	if len(args) > 1 {
		_, _ = fmt.Fprintln(stderr, "Usage: shed [folder]")
		return app.ExitError
	}

	shedRoot, err := userShedRoot()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Shed path unavailable: %v\n", err)
		return app.ExitError
	}

	return app.Run(ctx, app.Options{
		Args:     args,
		Stdout:   stdout,
		Stderr:   stderr,
		Resolver: shedfs.SelectedFolderResolver{},
		Pruner:   shedfs.NewPruner(shedRoot),
		Pruning: pruning.Runner{
			Input:  stdin,
			Output: stdout,
		},
		Scanner: shedfs.NewScanner(shedRoot),
		Mover:   shedfs.NewMover(shedRoot),
		Shedding: shedding.Runner{
			Input:  stdin,
			Output: stdout,
		},
		Final: final.Runner{
			Output: stdout,
		},
	})
}
