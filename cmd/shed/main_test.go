package main

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"shed/internal/app"
)

func TestRunValidatesCLIBeforeShedRoot(t *testing.T) {
	cases := []struct {
		name                string
		args                []string
		unsupportedPlatform bool
		wantStderr          string
	}{
		{
			name:                "unsupported platform",
			unsupportedPlatform: true,
			wantStderr:          "Unsupported platform\n",
		},
		{
			name:       "too many arguments",
			args:       []string{"a", "b"},
			wantStderr: "too many arguments\n" + cliHelp,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stderr := new(bytes.Buffer)
			shedRootCalled := false

			code := run(
				context.Background(),
				tc.args,
				strings.NewReader(""),
				new(bytes.Buffer),
				stderr,
				func() bool { return tc.unsupportedPlatform },
				func() (string, error) {
					shedRootCalled = true
					return "", errors.New("shed root should not be requested")
				},
			)

			if code != app.ExitError {
				t.Fatalf("expected exit error, got %d", code)
			}
			if stderr.String() != tc.wantStderr {
				t.Fatalf("expected stderr %q, got %q", tc.wantStderr, stderr.String())
			}
			if shedRootCalled {
				t.Fatalf("expected no shed root lookup before CLI validation")
			}
		})
	}
}

func TestRunValidatesAgeBeforeShedRoot(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		wantStderr string
	}{
		{
			name:       "missing age",
			args:       []string{"--age"},
			wantStderr: "flag needs an argument: --age\n" + cliHelp,
		},
		{
			name:       "non numeric age",
			args:       []string{"--age", "abc"},
			wantStderr: "invalid argument \"abc\" for \"--age\" flag: strconv.ParseInt: parsing \"abc\": invalid syntax\n" + cliHelp,
		},
		{
			name:       "negative age",
			args:       []string{"--age", "-1"},
			wantStderr: "invalid age: must be greater than or equal to 0\n" + cliHelp,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stderr := new(bytes.Buffer)
			shedRootCalled := false

			code := run(
				context.Background(),
				tc.args,
				strings.NewReader(""),
				new(bytes.Buffer),
				stderr,
				func() bool { return false },
				func() (string, error) {
					shedRootCalled = true
					return "", errors.New("shed root should not be requested")
				},
			)

			if code != app.ExitError {
				t.Fatalf("expected exit error, got %d", code)
			}
			if stderr.String() != tc.wantStderr {
				t.Fatalf("expected stderr %q, got %q", tc.wantStderr, stderr.String())
			}
			if shedRootCalled {
				t.Fatalf("expected no shed root lookup before CLI validation")
			}
		})
	}
}

func TestParseCLI(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantArg []string
		wantAge int
	}{
		{
			name:    "default age",
			args:    []string{"Downloads"},
			wantArg: []string{"Downloads"},
			wantAge: 60,
		},
		{
			name:    "custom age before folder",
			args:    []string{"--age", "30", "Downloads"},
			wantArg: []string{"Downloads"},
			wantAge: 30,
		},
		{
			name:    "custom age after folder",
			args:    []string{"Downloads", "--age", "0"},
			wantArg: []string{"Downloads"},
			wantAge: 0,
		},
		{
			name:    "custom age before folder with cmd trailing slash quote",
			args:    []string{"--age", "0", `C:\Users\pavel\Downloads\Telegram Desktop2"`},
			wantArg: []string{`C:\Users\pavel\Downloads\Telegram Desktop2`},
			wantAge: 0,
		},
		{
			name:    "custom age after folder with cmd trailing slash quote",
			args:    []string{`C:\Users\pavel\Downloads\Telegram Desktop2"`, "--age", "0"},
			wantArg: []string{`C:\Users\pavel\Downloads\Telegram Desktop2`},
			wantAge: 0,
		},
		{
			name:    "custom age fused after folder by cmd trailing slash quote",
			args:    []string{`C:\Users\pavel\Downloads\Telegram Desktop2" --age 0`},
			wantArg: []string{`C:\Users\pavel\Downloads\Telegram Desktop2`},
			wantAge: 0,
		},
		{
			name:    "folder with explicit current directory suffix",
			args:    []string{`C:\Users\pavel\Downloads\Telegram Desktop2\.`},
			wantArg: []string{`C:\Users\pavel\Downloads\Telegram Desktop2`},
			wantAge: 60,
		},
		{
			name:    "folder with quote before trailing slash",
			args:    []string{`C:\Users\pavel\Downloads\Telegram Desktop2"\`},
			wantArg: []string{`C:\Users\pavel\Downloads\Telegram Desktop2`},
			wantAge: 60,
		},
		{
			name:    "folder with escaped quote from cmd trailing slash",
			args:    []string{`C:\Users\pavel\Downloads\Telegram Desktop2\"`},
			wantArg: []string{`C:\Users\pavel\Downloads\Telegram Desktop2`},
			wantAge: 60,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseCLI(tc.args, new(bytes.Buffer))
			if err != nil {
				t.Fatal(err)
			}
			if !equalStrings(got.args, tc.wantArg) {
				t.Fatalf("expected args %v, got %v", tc.wantArg, got.args)
			}
			if got.retentionAgeDays != tc.wantAge {
				t.Fatalf("expected age %d, got %d", tc.wantAge, got.retentionAgeDays)
			}
		})
	}
}

func TestRunPrintsHelp(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	shedRootCalled := false

	code := run(
		context.Background(),
		[]string{"--help"},
		strings.NewReader(""),
		stdout,
		stderr,
		func() bool { return false },
		func() (string, error) {
			shedRootCalled = true
			return "", errors.New("shed root should not be requested")
		},
	)

	if code != app.ExitOK {
		t.Fatalf("expected exit ok, got %d", code)
	}
	if stdout.String() != cliHelp {
		t.Fatalf("expected stdout %q, got %q", cliHelp, stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	if shedRootCalled {
		t.Fatalf("expected no shed root lookup for help")
	}
}

func TestRunReportsShedRootError(t *testing.T) {
	stderr := new(bytes.Buffer)

	code := run(
		context.Background(),
		nil,
		strings.NewReader(""),
		new(bytes.Buffer),
		stderr,
		func() bool { return false },
		func() (string, error) { return "", errors.New("home missing") },
	)

	if code != app.ExitError {
		t.Fatalf("expected exit error, got %d", code)
	}
	if stderr.String() != "Shed path unavailable: home missing\n" {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

const cliHelp = `Usage: shed [flags] [folder]

Arguments:
      folder   folder to scan (defaults to current working directory)

Flags:
      --age int   minimum item age in whole days (default 60)
`

func TestRepairFusedWindowsArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			// No fusion — regular args passed through unchanged
			name: "no fused args",
			args: []string{`C:\Users\me\Downloads\file`, "--age", "0"},
			want: []string{`C:\Users\me\Downloads\file`, "--age", "0"},
		},
		{
			// Classic case from the docstring
			name: "single fused arg with one flag",
			args: []string{`C:\Users\me\Downloads\Telegram Desktop2" --age 0`},
			want: []string{`C:\Users\me\Downloads\Telegram Desktop2`, "--age", "0"},
		},
		{
			// Multiple flags fused into one argument
			name: "single fused arg with multiple flags",
			args: []string{`C:\My Folder" --verbose --output out.txt`},
			want: []string{`C:\My Folder`, "--verbose", "--output", "out.txt"},
		},
		{
			// Mix: first arg is fused, second is normal
			name: "fused arg followed by normal arg",
			args: []string{`C:\Folder With Spaces" --flag value`, "--other", "arg"},
			want: []string{`C:\Folder With Spaces`, "--flag", "value", "--other", "arg"},
		},
		{
			// Empty input
			name: "empty args",
			args: []string{},
			want: []string{},
		},
		{
			// Nil input
			name: "nil args",
			args: nil,
			want: []string{},
		},
		{
			// Arg that contains a quote but not the fuse pattern `" --`
			name: "quote without following flag",
			args: []string{`Some"Thing`},
			want: []string{`Some"Thing`},
		},
		{
			// Path without spaces, still fused
			name: "path without spaces fused",
			args: []string{`C:\NoSpaces\" --dry-run`},
			want: []string{`C:\NoSpaces\`, "--dry-run"},
		},
		{
			// Multiple fused args in input slice
			name: "multiple fused args in slice",
			args: []string{
				`C:\First Folder" --foo bar`,
				`C:\Second Folder" --baz`,
			},
			want: []string{
				`C:\First Folder`, "--foo", "bar",
				`C:\Second Folder`, "--baz",
			},
		},
		{
			name: "dashes inside folder name",
			args: []string{`C:\Users\me\Downloads\Telegram --age 0" --age 0`},
			want: []string{`C:\Users\me\Downloads\Telegram --age 0`, "--age", "0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := repairFusedWindowsArgs(tc.args)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("repairFusedWindowsArgs(%q)\n  got:  %q\n  want: %q", tc.args, got, tc.want)
			}
		})
	}
}
