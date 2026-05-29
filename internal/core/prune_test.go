package core

import (
	"testing"
	"time"
)

func TestParseArchiveMonthAcceptsStrictArchiveMonthShape(t *testing.T) {
	month, ok := ParseArchiveMonth(`~\Shed\2025\11`)
	if !ok {
		t.Fatalf("expected valid archive month")
	}
	if month.Year != 2025 || month.Month != 11 {
		t.Fatalf("expected 2025/11, got %d/%d", month.Year, month.Month)
	}
	if month.Path != `~\Shed\2025\11` {
		t.Fatalf("expected original path, got %q", month.Path)
	}
}

func TestParseArchiveMonthRejectsInvalidStructure(t *testing.T) {
	paths := []string{
		`~\Shed\2025`,
		`~\Shed\25\11`,
		`~\Shed\yyyy\11`,
		`~\Shed\2025\00`,
		`~\Shed\2025\13`,
		`~\Other\2025\11`,
		`C:\Users\pavel\Shed\2025\11`,
	}

	for _, path := range paths {
		if _, ok := ParseArchiveMonth(path); ok {
			t.Fatalf("expected invalid archive month path %q", path)
		}
	}
}

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
