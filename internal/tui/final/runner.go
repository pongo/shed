package final

import (
	"context"
	"fmt"
	"io"
	"strings"

	"shed/internal/app"
	"shed/internal/core"
)

type Runner struct {
	Output io.Writer
}

func (runner Runner) RunFinal(_ context.Context, request app.FinalSummaryRequest) error {
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	_, err := fmt.Fprint(output, formatFinalSummary(request))
	return err
}

func formatFinalSummary(request app.FinalSummaryRequest) string {
	sections := make([]string, 0, 2)

	if pruningSection, ok := renderPruningSection(request.Pruning); ok {
		sections = append(sections, pruningSection)
	}
	if archivingSection, ok := renderArchivingSection(request.Archiving); ok {
		sections = append(sections, archivingSection)
	}

	if len(sections) == 0 {
		return ""
	}
	return strings.Join(sections, "\n\n") + "\n"
}

func renderPruningSection(pruning app.PruningFinalData) (string, bool) {
	if !shouldRenderPruning(pruning) {
		return "", false
	}

	lines := []string{"Archive pruning"}
	lines = append(lines, fmt.Sprintf("%s moved to Recycle Bin", core.FormatSize(pruning.Summary.PrunedSize)))
	for _, path := range pruning.Summary.PrunedPaths {
		lines = append(lines, "Pruned: "+path)
	}
	for _, path := range pruning.Summary.FailedPaths {
		lines = append(lines, "Failed prune: "+path)
	}
	if pruning.Err != nil {
		lines = append(lines, "Pruning error: "+pruning.Err.Error())
	}

	return strings.Join(lines, "\n"), true
}

func shouldRenderPruning(pruning app.PruningFinalData) bool {
	if pruning.Err != nil {
		return true
	}
	if pruning.Summary.PrunedSize > 0 || len(pruning.Summary.PrunedPaths) > 0 || len(pruning.Summary.FailedPaths) > 0 {
		return true
	}
	if pruning.Outcome == app.PruningConfirmed {
		return true
	}
	if pruning.Outcome == app.PruningSkipped && !pruning.HadCandidates {
		return false
	}
	return false
}

func renderArchivingSection(archiving app.ArchivingFinalData) (string, bool) {
	if !archiving.Show && !archiving.NothingToMove && archiving.Err == nil && archiving.Summary.ArchiveBucket == "" && archiving.Summary.MovedSize == 0 && len(archiving.Summary.FailedPaths) == 0 && len(archiving.SkippedItems) == 0 {
		return "", false
	}

	lines := []string{"Archiving"}

	if archiving.NothingToMove {
		lines = append(lines, "Nothing to move")
	} else {
		lines = append(lines, fmt.Sprintf("%s moved to %s", core.FormatSize(archiving.Summary.MovedSize), archiving.Summary.ArchiveBucket))
	}

	for _, failed := range archiving.Summary.FailedPaths {
		lines = append(lines, "Failed move: "+failed)
	}
	for _, skipped := range archiving.SkippedItems {
		lines = append(lines, "Skipped item: "+skipped.Path)
	}
	if archiving.Err != nil {
		lines = append(lines, "Archiving error: "+archiving.Err.Error())
	}

	return strings.Join(lines, "\n"), true
}
