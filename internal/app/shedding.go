package app

import (
	"context"

	"shed/internal/core"
)

func runSheddingPhase(ctx context.Context, opts Options, invocationFolder, selectedFolder string, pruning PruningFinalData) SheddingFinalData {
	result, scanErr := opts.Scanner.Scan(ctx, selectedFolder)
	shedding := SheddingFinalData{
		SkippedItems: result.SkippedItems,
		Err:          scanErr,
		scanFailed:   scanErr != nil,
	}
	if scanErr != nil {
		return shedding
	}
	if len(result.StaleItems) == 0 {
		shedding.NothingToMove = true
		shedding.Show = hasPruningReport(pruning)
		return shedding
	}

	shedding.Show = true
	plannedBucket := core.PlanShedBucket(opts.ShedRoot, opts.Now(), invocationFolder, selectedFolder)
	confirmation := ConfirmationRequest{
		SelectedFolder:    selectedFolder,
		HeaderTitle:       plannedBucket.HeaderTitle,
		CompactShedBucket: plannedBucket.CompactShedBucket,
		ScanResult:        result,
	}
	sheddingResult, runErr := opts.Shedding.RunShedding(ctx, SheddingRequest{
		Confirmation: confirmation,
		Move: func(ctx context.Context) (core.MoveSummary, error) {
			return opts.Mover.Move(ctx, selectedFolder, plannedBucket, result)
		},
		View: MoveViewData{
			SkippedItems: result.SkippedItems,
		},
	})
	shedding.Summary = sheddingResult.Summary
	if runErr != nil {
		shedding.Err = runErr
	}
	if sheddingResult.Outcome == SheddingCancelled {
		shedding.Show = false
	}

	return shedding
}
