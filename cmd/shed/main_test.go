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
			wantStderr: "Usage: shed [folder]\n",
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
