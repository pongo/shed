package final

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"shed/internal/app"
	"shed/internal/core"
)

func TestFinalSummaryRendersPruningSectionWithSuccessAndFailures(t *testing.T) {
	view := formatFinalSummary(app.FinalSummaryRequest{
		Pruning: app.PruningFinalData{
			HadCandidates: true,
			Outcome:       app.PruningConfirmed,
			Summary: core.PruneSummary{
				PrunedSize:  3072,
				PrunedPaths: []string{filepath.Join("~", "Shed", "2024", "01"), filepath.Join("~", "Shed", "2024", "02")},
				FailedPaths: []string{filepath.Join("~", "Shed", "2024", "03")},
			},
		},
		Archiving: app.ArchivingFinalData{
			NothingToMove: true,
		},
	})

	for _, want := range []string{
		"Archive pruning",
		"3 KB moved to Recycle Bin",
		"Pruned Archive month: " + filepath.Join("~", "Shed", "2024", "01"),
		"Pruned Archive month: " + filepath.Join("~", "Shed", "2024", "02"),
		"Failed prune: " + filepath.Join("~", "Shed", "2024", "03"),
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected combined summary to contain %q, got:\n%s", want, view)
		}
	}
}

func TestFinalSummaryRendersPruningErrors(t *testing.T) {
	view := formatFinalSummary(app.FinalSummaryRequest{
		Pruning: app.PruningFinalData{
			HadCandidates: false,
			Outcome:       app.PruningSkipped,
			Err:           errors.New("scan denied"),
		},
		Archiving: app.ArchivingFinalData{
			NothingToMove: true,
		},
	})

	if !strings.Contains(view, "Pruning error: scan denied") {
		t.Fatalf("expected pruning error, got:\n%s", view)
	}
}

func TestFinalSummaryRendersArchivingMoveSummary(t *testing.T) {
	bucket := filepath.Join("C:", "Users", "pavel", "Shed", "2026", "05", "Downloads")
	view := formatFinalSummary(app.FinalSummaryRequest{
		Archiving: app.ArchivingFinalData{
			Summary: core.MoveSummary{
				MovedSize:     10,
				ArchiveBucket: bucket,
				FailedPaths:   []string{filepath.Join("C:", "Users", "pavel", "Downloads", "locked.txt")},
			},
			SkippedItems: []core.SkippedItem{{Path: filepath.Join("C:", "Users", "pavel", "Downloads", "unreadable.txt")}},
		},
	})

	for _, want := range []string{
		"Archiving",
		"10 B moved to " + bucket,
		"Failed move: " + filepath.Join("C:", "Users", "pavel", "Downloads", "locked.txt"),
		"Skipped item: " + filepath.Join("C:", "Users", "pavel", "Downloads", "unreadable.txt"),
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected archiving summary to contain %q, got:\n%s", want, view)
		}
	}
}

func TestFinalSummaryRendersNothingToMoveInArchivingSection(t *testing.T) {
	view := formatFinalSummary(app.FinalSummaryRequest{
		Pruning: app.PruningFinalData{
			HadCandidates: true,
			Outcome:       app.PruningConfirmed,
			Summary:       core.PruneSummary{PrunedSize: 1},
		},
		Archiving: app.ArchivingFinalData{
			NothingToMove: true,
		},
	})

	if !strings.Contains(view, "Archiving\nNothing to move") {
		t.Fatalf("expected Nothing to move in Archiving section, got:\n%s", view)
	}
}

func TestFinalSummarySeparatesSectionsWithBlankLine(t *testing.T) {
	view := formatFinalSummary(app.FinalSummaryRequest{
		Pruning: app.PruningFinalData{
			HadCandidates: true,
			Outcome:       app.PruningConfirmed,
			Summary:       core.PruneSummary{PrunedSize: 1},
		},
		Archiving: app.ArchivingFinalData{
			NothingToMove: true,
		},
	})

	if !strings.Contains(view, "Archive pruning\n1 B moved to Recycle Bin\n\nArchiving") {
		t.Fatalf("expected blank line between sections, got:\n%s", view)
	}
}

func TestFinalSummaryOmitsPruningWhenNoOpOrSkippedWithoutErrors(t *testing.T) {
	view := formatFinalSummary(app.FinalSummaryRequest{
		Pruning: app.PruningFinalData{
			HadCandidates: false,
			Outcome:       app.PruningSkipped,
		},
		Archiving: app.ArchivingFinalData{
			NothingToMove: true,
		},
	})

	if strings.Contains(view, "Archive pruning") {
		t.Fatalf("expected pruning section to be omitted, got:\n%s", view)
	}
	if !strings.Contains(view, "Archiving\nNothing to move") {
		t.Fatalf("expected archiving section to remain, got:\n%s", view)
	}
}
