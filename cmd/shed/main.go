package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"

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

type cliOptions struct {
	args             []string
	retentionAgeDays int
}

func parseCLI(args []string, output io.Writer) (cliOptions, error) {
	options := cliOptions{
		retentionAgeDays: core.DefaultRetentionAgeDays,
	}

	args = repairFusedWindowsArgs(args)

	flags := newCLIFlagSet(output)
	addCLIFlags(flags, &options)

	if err := flags.Parse(args); err != nil {
		return cliOptions{}, err
	}

	options.args = normalizeFolderArgs(flags.Args())
	if len(options.args) > 1 {
		return cliOptions{}, fmt.Errorf("too many arguments")
	}
	if options.retentionAgeDays < 0 {
		return cliOptions{}, fmt.Errorf("invalid age: must be greater than or equal to 0")
	}

	return options, nil
}

// repairFusedWindowsArgs repairs cmd.exe argv after a quoted folder ending in a
// backslash swallows the closing quote and fuses the following flags into the
// same argument.
//
// Example:
//
//	shed "C:\Users\me\Downloads\Telegram Desktop2\" --age 0
//
// can arrive as:
//
//	[]string{`C:\Users\me\Downloads\Telegram Desktop2" --age 0`}
//
// pflag cannot parse --age until that fused argument is split back into a
// folder argument and flag arguments.
func repairFusedWindowsArgs(args []string) []string {
	repaired := make([]string, 0, len(args))
	for _, arg := range args {
		folder, rest, ok := strings.Cut(arg, `" --`)
		if !ok {
			repaired = append(repaired, arg)
			continue
		}

		repaired = append(repaired, folder)
		repaired = append(repaired, strings.Fields("--"+rest)...)
	}
	return repaired
}

func newCLIFlagSet(output io.Writer) *flag.FlagSet {
	flags := flag.NewFlagSet("shed", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.SortFlags = false
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Usage: shed [flags] [folder]")
		_, _ = fmt.Fprintln(flags.Output())
		_, _ = fmt.Fprintln(flags.Output(), "Arguments:")
		_, _ = fmt.Fprintln(flags.Output(), "      folder   folder to scan (defaults to current working directory)")
		_, _ = fmt.Fprintln(flags.Output())
		_, _ = fmt.Fprintln(flags.Output(), "Flags:")
		flags.PrintDefaults()
	}
	return flags
}

func addCLIFlags(flags *flag.FlagSet, options *cliOptions) {
	flags.IntVar(&options.retentionAgeDays, "age", options.retentionAgeDays, "minimum item age in whole days")
}

// normalizeFolderArgs removes quotes left behind by cmd.exe quoted path parsing
// and cleans the selected folder path.
//
// Examples:
//
//	shed --age 0 "C:\Users\me\Downloads\Telegram Desktop2\"
//	shed "C:\Users\me\Downloads\Telegram Desktop2\" --age 0
//
// can leave a literal quote in the folder argument, such as:
//
//	C:\Users\me\Downloads\Telegram Desktop2"
//	C:\Users\me\Downloads\Telegram Desktop2\"
func normalizeFolderArgs(args []string) []string {
	normalized := make([]string, len(args))
	for i, arg := range args {
		normalized[i] = filepath.Clean(strings.ReplaceAll(arg, `"`, ""))
	}
	return normalized
}

func printCLIHelp(output io.Writer) {
	options := cliOptions{
		retentionAgeDays: core.DefaultRetentionAgeDays,
	}
	flags := newCLIFlagSet(output)
	addCLIFlags(flags, &options)
	flags.Usage()
}
