package app

import (
	"context"
	"fmt"
)

func runFinalPhase(ctx context.Context, opts Options, pruning PruningFinalData, shedding SheddingFinalData) int {
	if shedding.scanFailed && !hasPruningReport(pruning) {
		_, _ = fmt.Fprintf(opts.Stderr, "Scan failed: %v\n", shedding.Err)
		return ExitError
	}
	if shedding.NothingToMove && !hasPruningReport(pruning) {
		_, _ = fmt.Fprintln(opts.Stdout, "Nothing to move")
		for _, skipped := range shedding.SkippedItems {
			_, _ = fmt.Fprintf(opts.Stdout, "Skipped item: %s\n", skipped.Path)
		}
		return ExitOK
	}
	if hasReport(pruning, shedding) {
		_ = opts.Final.RunFinal(ctx, FinalSummaryRequest{Pruning: pruning, Shedding: shedding})
	}

	if shouldExitError(pruning, shedding) {
		return ExitError
	}
	return ExitOK
}

func hasPruningReport(pruning PruningFinalData) bool {
	if pruning.Err != nil {
		return true
	}
	if pruning.Outcome == PruningConfirmed {
		return true
	}
	return len(pruning.Summary.FailedPaths) > 0 || len(pruning.Summary.PrunedPaths) > 0 || pruning.Summary.PrunedSize > 0
}

func hasReport(pruning PruningFinalData, shedding SheddingFinalData) bool {
	return hasPruningReport(pruning) || shedding.Show || shedding.Err != nil
}

func shouldExitError(pruning PruningFinalData, shedding SheddingFinalData) bool {
	if pruning.Err != nil || len(pruning.Summary.FailedPaths) > 0 {
		return true
	}
	if shedding.Err != nil || len(shedding.Summary.FailedPaths) > 0 {
		return true
	}
	return false
}
