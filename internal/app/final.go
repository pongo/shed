package app

import (
	"context"
	"fmt"
)

func runFinalPhase(ctx context.Context, opts Options, pruning PruningFinalData, archiving ArchivingFinalData) int {
	if archiving.scanFailed && !hasPruningReport(pruning) {
		_, _ = fmt.Fprintf(opts.Stderr, "Scan failed: %v\n", archiving.Err)
		return ExitError
	}
	if archiving.NothingToMove && !hasPruningReport(pruning) {
		_, _ = fmt.Fprintln(opts.Stdout, "Nothing to move")
		for _, skipped := range archiving.SkippedItems {
			_, _ = fmt.Fprintf(opts.Stdout, "Skipped item: %s\n", skipped.Path)
		}
		return ExitOK
	}
	if hasReport(pruning, archiving) {
		_ = opts.Final.RunFinal(ctx, FinalSummaryRequest{Pruning: pruning, Archiving: archiving})
	}

	if shouldExitError(pruning, archiving) {
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

func hasReport(pruning PruningFinalData, archiving ArchivingFinalData) bool {
	return hasPruningReport(pruning) || archiving.Show || archiving.Err != nil
}

func shouldExitError(pruning PruningFinalData, archiving ArchivingFinalData) bool {
	if pruning.Err != nil || len(pruning.Summary.FailedPaths) > 0 {
		return true
	}
	if archiving.Err != nil || len(archiving.Summary.FailedPaths) > 0 {
		return true
	}
	return false
}
