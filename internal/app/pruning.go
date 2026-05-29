package app

import (
	"context"

	"shed/internal/core"
)

func runPruningPhase(ctx context.Context, opts Options) (PruningFinalData, bool) {
	pruneScan, pruneErr := opts.Pruner.Scan(ctx)
	pruning := PruningFinalData{
		HadCandidates: len(pruneScan.Candidates) > 0,
		Outcome:       PruningSkipped,
		Err:           pruneErr,
	}
	if !pruning.HadCandidates {
		return pruning, true
	}

	pruningResult, runErr := opts.Pruning.RunPruning(ctx, PruningRequest{
		Scan: pruneScan,
		Prune: func(ctx context.Context) (core.PruneSummary, error) {
			return opts.Pruner.Prune(ctx, pruneScan)
		},
	})
	pruning.Outcome = pruningResult.Outcome
	pruning.Summary = pruningResult.Summary
	if runErr != nil {
		pruning.Err = runErr
	}

	return pruning, pruning.Outcome != PruningQuit
}
