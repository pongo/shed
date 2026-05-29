# Add combined final summary

Status: ready-for-agent

## Parent

.scratch/archive-pruning/PRD.md

## What to build

Add the final summary program that renders one combined report for **Archive pruning** and archiving. The final output should use clear phase sections named `Archive pruning` and `Archiving`, separated by a blank line. It should report successful pruned Archive months, failed prune paths, pruning errors, archiving move summaries, failed move paths, skipped items, and `Nothing to move` when that belongs in the combined output.

The pruning section should be omitted when pruning had no candidates or was skipped without output or errors.

## Acceptance criteria

- [ ] Final output can render an `Archive pruning` section with total pruned size.
- [ ] Successful pruned Archive month paths are listed in the pruning section.
- [ ] Failed prune paths are listed in the pruning section.
- [ ] Pruning scan or run errors are listed in the pruning section.
- [ ] Final output can render an `Archiving` section with existing move summary information.
- [ ] `Nothing to move` can be rendered under `Archiving` when pruning already produced output or errors.
- [ ] The two phase sections are separated by a blank line when both are present.
- [ ] The pruning section is omitted for no-op or skipped pruning without errors.
- [ ] Final summary tests cover success, failure, omitted sections, and combined `Nothing to move`.

## Blocked by

- .scratch/archive-pruning/issues/01-model-archive-month-pruning-in-core.md
- .scratch/archive-pruning/issues/04-refactor-archiving-runner-to-stop-owning-final-output.md

Commit message: `Add combined pruning and archiving final summary`
