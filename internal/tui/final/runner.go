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
	if sheddingSection, ok := renderSheddingSection(request.Shedding); ok {
		sections = append(sections, sheddingSection)
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

	lines := []string{"Shed pruning"}
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

func renderSheddingSection(shedding app.SheddingFinalData) (string, bool) {
	if !shedding.Show && !shedding.NothingToMove && shedding.Err == nil && shedding.Summary.ShedBucket == "" && shedding.Summary.MovedSize == 0 && len(shedding.Summary.FailedPaths) == 0 && len(shedding.SkippedItems) == 0 {
		return "", false
	}

	lines := []string{"Shedding"}

	if shedding.NothingToMove {
		lines = append(lines, "Nothing to move")
	} else {
		lines = append(lines, fmt.Sprintf("%s moved to %s", core.FormatSize(shedding.Summary.MovedSize), shedding.Summary.ShedBucket))
	}

	for _, failed := range shedding.Summary.FailedPaths {
		lines = append(lines, "Failed move: "+failed)
	}
	for _, skipped := range shedding.SkippedItems {
		lines = append(lines, "Skipped item: "+skipped.Path)
	}
	if shedding.Err != nil {
		lines = append(lines, "Shedding error: "+shedding.Err.Error())
	}

	return strings.Join(lines, "\n"), true
}
