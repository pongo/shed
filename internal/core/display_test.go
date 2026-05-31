package core

import (
	"path/filepath"
	"testing"
	"time"
)

func TestHeaderTitleUsesSelectedFolderBaseName(t *testing.T) {
	selected := filepath.Join("C:", "Users", "pavel", "Downloads")

	if got := HeaderTitle(selected); got != "Downloads" {
		t.Fatalf("expected Downloads, got %q", got)
	}
}

func TestHeaderTitleUsesCleanPathForFilesystemRoot(t *testing.T) {
	root := filepath.Clean(`C:\`)

	if got := HeaderTitle(root); got != root {
		t.Fatalf("expected root title %q, got %q", root, got)
	}
}

func TestCompactShedBucket(t *testing.T) {
	invocation := filepath.Join("C:", "Users", "pavel", "Downloads")
	selected := filepath.Join("C:", "Users", "pavel", "Downloads")
	moveDate := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	want := filepath.Join("~", "Shed", "2026", "05", "Downloads")

	if got := CompactShedBucket(moveDate, invocation, selected); got != want {
		t.Fatalf("expected compact bucket %q, got %q", want, got)
	}
}

func TestShedBucketUsesShedRootDateAndSelectedFolderBaseName(t *testing.T) {
	shedRoot := filepath.Join("C:", "Users", "pavel", "Shed")
	invocation := filepath.Join("C:", "Users", "pavel", "Projects")
	selected := filepath.Join("D:", "Scratch", "Downloads")
	moveDate := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	want := filepath.Join(shedRoot, "2026", "05", "Downloads")

	if got := ShedBucket(shedRoot, moveDate, invocation, selected); got != want {
		t.Fatalf("expected bucket %q, got %q", want, got)
	}
}

func TestPlanShedBucketReturnsOneFrozenBucketFact(t *testing.T) {
	shedRoot := filepath.Join("C:", "Users", "pavel", "Shed")
	invocation := filepath.Join("C:", "Users", "pavel", "Projects", "shed")
	selected := filepath.Join(invocation, ".scratch")
	moveDate := time.Date(2026, 5, 31, 23, 59, 0, 0, time.UTC)

	planned := PlanShedBucket(shedRoot, moveDate, invocation, selected)

	if planned.MoveDate != moveDate {
		t.Fatalf("expected move date %v, got %v", moveDate, planned.MoveDate)
	}
	if planned.HeaderTitle != ".scratch" {
		t.Fatalf("expected header title .scratch, got %q", planned.HeaderTitle)
	}
	if planned.BucketSourcePath != filepath.Join("shed", ".scratch") {
		t.Fatalf("expected bucket source path %q, got %q", filepath.Join("shed", ".scratch"), planned.BucketSourcePath)
	}
	if planned.CompactShedBucket != filepath.Join("~", "Shed", "2026", "05", "shed", ".scratch") {
		t.Fatalf("expected compact Shed bucket for May, got %q", planned.CompactShedBucket)
	}
	if planned.ShedBucket != filepath.Join(shedRoot, "2026", "05", "shed", ".scratch") {
		t.Fatalf("expected actual Shed bucket for May, got %q", planned.ShedBucket)
	}
}

func TestBucketSourcePathUsesInvocationFolderForSelectedInvocation(t *testing.T) {
	invocation := filepath.Join("C:", "Users", "pavel", "Projects", "shed")

	if got := BucketSourcePath(invocation, invocation); got != "shed" {
		t.Fatalf("expected invocation folder name, got %q", got)
	}
}

func TestBucketSourcePathUsesInvocationRelativeSelectedDescendant(t *testing.T) {
	invocation := filepath.Join("C:", "Users", "pavel", "Projects", "shed")
	selected := filepath.Join(invocation, "docs", "plans")
	want := filepath.Join("shed", "docs", "plans")

	if got := BucketSourcePath(invocation, selected); got != want {
		t.Fatalf("expected invocation-relative source path %q, got %q", want, got)
	}
}

func TestBucketSourcePathNormalizesEquivalentSelectedPaths(t *testing.T) {
	invocation := filepath.Join("C:", "Users", "pavel", "Projects", "shed")
	first := filepath.Join(invocation, "docs", "plans")
	second := filepath.Join(invocation, "docs", "..", "docs", "plans")

	if got, want := BucketSourcePath(invocation, second), BucketSourcePath(invocation, first); got != want {
		t.Fatalf("expected normalized equivalent source path %q, got %q", want, got)
	}
}

func TestBucketSourcePathKeepsSelectedFolderNameOutsideInvocation(t *testing.T) {
	invocation := filepath.Join("C:", "Users", "pavel", "Projects", "shed")
	selected := filepath.Join("C:", "Users", "pavel", "Downloads")

	if got := BucketSourcePath(invocation, selected); got != "Downloads" {
		t.Fatalf("expected selected folder name, got %q", got)
	}
}

func TestBucketSourcePathKeepsRootBehaviorForFilesystemRootInvocation(t *testing.T) {
	invocation := filepath.Clean(`C:\`)
	selected := filepath.Join(invocation, "Users", "pavel")

	if got := BucketSourcePath(invocation, selected); got != "pavel" {
		t.Fatalf("expected selected folder name for root invocation, got %q", got)
	}
}

func TestHasNameConflictUsesCaseInsensitiveWindowsSemantics(t *testing.T) {
	existing := []string{"Report.PDF"}

	if !HasNameConflict(existing, "report.pdf") {
		t.Fatalf("expected case-insensitive conflict")
	}
}

func TestResolveNumberedNamePlacesSuffixBeforeFinalExtension(t *testing.T) {
	tests := map[string]string{
		"report.pdf":    "report (2).pdf",
		"bundle.tar.gz": "bundle.tar (1).gz",
		"README":        "README (1)",
	}
	existing := []string{"report.pdf", "report (1).pdf", "bundle.tar.gz", "README"}

	for name, want := range tests {
		if got := ResolveNumberedName(existing, name); got != want {
			t.Fatalf("ResolveNumberedName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestMoveSummaryStoresActualMovedSize(t *testing.T) {
	summary := MoveSummary{
		ShedBucket: filepath.Join("C:", "Users", "pavel", "Shed", "2026", "05", "Downloads"),
		MovedSize:  12,
	}

	if summary.MovedSize != 12 {
		t.Fatalf("expected actual moved size 12, got %d", summary.MovedSize)
	}
}
