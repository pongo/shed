package main

import (
	"bytes"
	"context"
	"errors"
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
