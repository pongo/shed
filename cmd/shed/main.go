package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"shed/internal/app"
	"shed/internal/core"
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

	cli, err := parseCLI(args)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		printUsage(stderr)
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

type cliOptions struct {
	args             []string
	retentionAgeDays int
}

// parseCLI accepts --age before or after the folder, unlike Go's flag package.
func parseCLI(args []string) (cliOptions, error) {
	options := cliOptions{
		retentionAgeDays: core.DefaultRetentionAgeDays,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg != "--age" {
			options.args = append(options.args, normalizeFolderCLIArg(arg))
			continue
		}

		if i+1 >= len(args) {
			return cliOptions{}, fmt.Errorf("invalid age: missing value")
		}
		i++

		age, err := strconv.Atoi(args[i])
		if err != nil {
			return cliOptions{}, fmt.Errorf("invalid age: must be a whole number")
		}
		if age < 0 {
			return cliOptions{}, fmt.Errorf("invalid age: must be greater than or equal to 0")
		}
		options.retentionAgeDays = age
	}

	if len(options.args) > 1 {
		return cliOptions{}, fmt.Errorf("too many arguments")
	}

	return options, nil
}

func normalizeFolderCLIArg(arg string) string {
	return strings.ReplaceAll(arg, `"`, "")
}

func printUsage(stderr io.Writer) {
	_, _ = fmt.Fprintf(stderr, "Usage: shed [--age days=%d] [folder]\n", core.DefaultRetentionAgeDays)
}
