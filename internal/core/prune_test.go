package core

import (
	"testing"
	"time"
)

func TestIsPruneEligibleUsesStrictSixMonthCalendarCutoff(t *testing.T) {
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)

	eligible := ArchiveMonth{Path: `~\Shed\2025\10`, Year: 2025, Month: 10}
	boundary := ArchiveMonth{Path: `~\Shed\2025\11`, Year: 2025, Month: 11}
	recent := ArchiveMonth{Path: `~\Shed\2026\01`, Year: 2026, Month: 1}

	if !IsPruneEligible(eligible, now) {
		t.Fatalf("expected %v to be eligible", eligible)
	}
	if IsPruneEligible(boundary, now) {
		t.Fatalf("expected %v to be ineligible at strict boundary", boundary)
	}
	if IsPruneEligible(recent, now) {
		t.Fatalf("expected %v to be ineligible", recent)
	}
}

func TestSelectPruneCandidatesFiltersAndSortsOldestFirst(t *testing.T) {
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)

	input := []PruneCandidate{
		{Month: ArchiveMonth{Path: `~\Shed\2025\12`, Year: 2025, Month: 12}, Size: 12},
		{Month: ArchiveMonth{Path: `~\Shed\2024\01`, Year: 2024, Month: 1}, Size: 1},
		{Month: ArchiveMonth{Path: `~\Shed\2025\10`, Year: 2025, Month: 10}, Size: 10},
		{Month: ArchiveMonth{Path: `~\Shed\2026\02`, Year: 2026, Month: 2}, Size: 2},
	}

	result := SelectPruneCandidates(input, now)
	if len(result.Candidates) != 2 {
		t.Fatalf("expected 2 prune candidates, got %d", len(result.Candidates))
	}

	gotPaths := []string{
		result.Candidates[0].Month.Path,
		result.Candidates[1].Month.Path,
	}
	wantPaths := []string{
		`~\Shed\2024\01`,
		`~\Shed\2025\10`,
	}
	if !equalStrings(gotPaths, wantPaths) {
		t.Fatalf("expected sorted paths %v, got %v", wantPaths, gotPaths)
	}
}

func TestPruneSummaryAccumulatesSuccessAndFailureWithoutRollback(t *testing.T) {
	summary := NewPruneSummary()

	summary.RecordPruned(`~\Shed\2024\01`, 50)
	summary.RecordFailed(`~\Shed\2024\02`)
	summary.RecordPruned(`~\Shed\2024\03`, 30)

	if summary.PrunedSize != 80 {
		t.Fatalf("expected pruned size 80, got %d", summary.PrunedSize)
	}
	if !equalStrings(summary.PrunedPaths, []string{`~\Shed\2024\01`, `~\Shed\2024\03`}) {
		t.Fatalf("unexpected pruned paths: %v", summary.PrunedPaths)
	}
	if !equalStrings(summary.FailedPaths, []string{`~\Shed\2024\02`}) {
		t.Fatalf("unexpected failed paths: %v", summary.FailedPaths)
	}
}
