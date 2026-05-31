package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	flag "github.com/spf13/pflag"

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

	cli, err := parseCLI(args, stdout)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return app.ExitOK
		}
		_, _ = fmt.Fprintln(stderr, err)
		printCLIHelp(stderr)
		return app.ExitError
	}

	shedRoot, err := userShedRoot()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Shed path unavailable: %v\n", err)
		return app.ExitError
	}

	invocationFolder, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Invocation folder unavailable: %v\n", err)
		return app.ExitError
	}

	return app.Run(ctx, app.Options{
		Args:             cli.args,
		InvocationFolder: invocationFolder,
		ShedRoot:         shedRoot,
		Stdout:           stdout,
		Stderr:           stderr,
		Resolver:         shedfs.SelectedFolderResolver{},
		Pruner:           shedfs.NewPruner(shedRoot),
		Pruning: pruning.Runner{
			Input:  stdin,
			Output: stdout,
		},
		Scanner: shedfs.Scanner{
			ShedRoot:         shedRoot,
			RetentionAgeDays: &cli.retentionAgeDays,
		},
		Mover: shedfs.NewMover(shedRoot),
		Shedding: shedding.Runner{
			Input:  stdin,
			Output: stdout,
		},
		Final: final.Runner{
			Output: stdout,
		},
	})
}
