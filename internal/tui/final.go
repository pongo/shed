package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"shed/internal/core"
)

func finalSummaryView(summary core.MoveSummary, skippedItems []core.SkippedItem) tea.View {
	return tea.NewView(formatFinalSummary(summary, skippedItems))
}

func formatFinalSummary(summary core.MoveSummary, skippedItems []core.SkippedItem) string {
	var output strings.Builder
	fmt.Fprintf(&output, "%s moved to %s\n", core.FormatSize(summary.MovedSize), summary.ArchiveBucket)
	for _, failed := range summary.FailedPaths {
		fmt.Fprintf(&output, "Failed move: %s\n", failed)
	}
	for _, skipped := range skippedItems {
		fmt.Fprintf(&output, "Skipped item: %s\n", skipped.Path)
	}
	return output.String()
}
