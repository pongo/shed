package main

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParseCLIRepairsWindowsFusedArgs(t *testing.T) {
	got, err := parseCLI([]string{`C:\Users\pavel\Downloads\Telegram Desktop2" --age 0`}, new(bytes.Buffer))
	if err != nil {
		t.Fatal(err)
	}
	if !equalStrings(got.args, []string{`C:\Users\pavel\Downloads\Telegram Desktop2`}) {
		t.Fatalf("expected args %v, got %v", []string{`C:\Users\pavel\Downloads\Telegram Desktop2`}, got.args)
	}
	if got.retentionAgeDays != 0 {
		t.Fatalf("expected age %d, got %d", 0, got.retentionAgeDays)
	}
}

func TestRepairPlatformArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			// Regular args pass through unchanged.
			name: "no fused args",
			args: []string{`C:\Users\me\Downloads\file`, "--age", "0"},
			want: []string{`C:\Users\me\Downloads\file`, "--age", "0"},
		},
		{
			// Classic case from the docstring.
			name: "single fused arg with one flag",
			args: []string{`C:\Users\me\Downloads\Telegram Desktop2" --age 0`},
			want: []string{`C:\Users\me\Downloads\Telegram Desktop2`, "--age", "0"},
		},
		{
			// Multiple flags fused into one argument.
			name: "single fused arg with multiple flags",
			args: []string{`C:\My Folder" --verbose --output out.txt`},
			want: []string{`C:\My Folder`, "--verbose", "--output", "out.txt"},
		},
		{
			// First arg is fused, second is normal.
			name: "fused arg followed by normal arg",
			args: []string{`C:\Folder With Spaces" --flag value`, "--other", "arg"},
			want: []string{`C:\Folder With Spaces`, "--flag", "value", "--other", "arg"},
		},
		{
			name: "empty args",
			args: []string{},
			want: []string{},
		},
		{
			name: "nil args",
			args: nil,
			want: []string{},
		},
		{
			// Args with quotes but without the fuse pattern pass through unchanged.
			name: "quote without following flag",
			args: []string{`Some"Thing`},
			want: []string{`Some"Thing`},
		},
		{
			// Paths without spaces can still be fused.
			name: "path without spaces fused",
			args: []string{`C:\NoSpaces\" --dry-run`},
			want: []string{`C:\NoSpaces\`, "--dry-run"},
		},
		{
			// Each fused arg is repaired independently.
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
			got := repairPlatformArgs(tc.args)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("repairPlatformArgs(%q)\n  got:  %q\n  want: %q", tc.args, got, tc.want)
			}
		})
	}
}
