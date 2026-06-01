package main

import (
	"fmt"
	"io"

	flag "github.com/spf13/pflag"

	"shed/internal/core"
)

type cliOptions struct {
	args             []string
	retentionAgeDays int
}

func parseCLI(args []string, output io.Writer) (cliOptions, error) {
	options := cliOptions{
		retentionAgeDays: core.DefaultRetentionAgeDays,
	}

	flags := newCLIFlagSet(output)
	addCLIFlags(flags, &options)

	if err := flags.Parse(args); err != nil {
		return cliOptions{}, err
	}

	options.args = flags.Args()
	if len(options.args) > 1 {
		return cliOptions{}, fmt.Errorf("too many arguments")
	}
	if options.retentionAgeDays < 0 {
		return cliOptions{}, fmt.Errorf("invalid age: must be greater than or equal to 0")
	}

	return options, nil
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
	flags.IntVar(&options.retentionAgeDays, "age", options.retentionAgeDays, "minimum item age in whole days (default 0)")
}

func printCLIHelp(output io.Writer) {
	options := cliOptions{
		retentionAgeDays: core.DefaultRetentionAgeDays,
	}
	flags := newCLIFlagSet(output)
	addCLIFlags(flags, &options)
	flags.Usage()
}
