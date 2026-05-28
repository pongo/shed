package archiving

import (
	"path/filepath"
	"strings"
	"testing"

	"shed/internal/core"
)

func TestFinalSummaryRendersFailedAndSkippedItems(t *testing.T) {
	bucket := filepath.Join("C:", "Users", "pavel", "Shed")
	failedPath := filepath.Join("C:", "Users", "pavel", "locked.txt")
	skippedPath := filepath.Join("C:", "Users", "pavel", "unreadable.txt")

	view := finalSummaryView(core.MoveSummary{
		MovedSize:     10,
		ArchiveBucket: bucket,
		FailedPaths:   []string{failedPath},
	}, []core.SkippedItem{{Path: skippedPath}}).Content

	for _, want := range []string{
		"10 B moved to " + bucket,
		"Failed move: " + failedPath,
		"Skipped item: " + skippedPath,
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected final summary to contain %q, got %q", want, view)
		}
	}
}
