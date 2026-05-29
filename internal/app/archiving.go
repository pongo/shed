package app

import (
	"context"

	"shed/internal/core"
)

func runArchivingPhase(ctx context.Context, opts Options, selectedFolder string, pruning PruningFinalData) ArchivingFinalData {
	result, scanErr := opts.Scanner.Scan(ctx, selectedFolder)
	archiving := ArchivingFinalData{
		SkippedItems: result.SkippedItems,
		Err:          scanErr,
		scanFailed:   scanErr != nil,
	}
	if scanErr != nil {
		return archiving
	}
	if len(result.StaleItems) == 0 {
		archiving.NothingToMove = true
		archiving.Show = hasPruningReport(pruning)
		return archiving
	}

	archiving.Show = true
	confirmation := ConfirmationRequest{
		SelectedFolder:       selectedFolder,
		HeaderTitle:          core.HeaderTitle(selectedFolder),
		CompactArchiveBucket: core.CompactArchiveBucket(opts.Now(), selectedFolder),
		ScanResult:           result,
	}
	archivingResult, runErr := opts.Archiving.RunArchiving(ctx, ArchivingRequest{
		Confirmation: confirmation,
		Move: func(ctx context.Context) (core.MoveSummary, error) {
			return opts.Mover.Move(ctx, selectedFolder, result)
		},
		View: MoveViewData{
			SkippedItems: result.SkippedItems,
		},
	})
	archiving.Summary = archivingResult.Summary
	if runErr != nil {
		archiving.Err = runErr
	}
	if archivingResult.Outcome == ArchivingCancelled {
		archiving.Show = false
	}

	return archiving
}
