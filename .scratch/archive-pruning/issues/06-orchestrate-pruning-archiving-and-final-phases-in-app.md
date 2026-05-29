# Orchestrate pruning, archiving, and final phases in the app

Status: ready-for-agent

## Parent

.scratch/archive-pruning/PRD.md

## What to build

Refactor app orchestration so a run has explicit pruning, archiving, and final phases. Early validation still happens first, then pruning runs before archiving scan. Pruning errors do not block archiving. The app aggregates phase outcomes and decides whether to show the final summary and which exit code to return.

The app should keep ordinary no-op runs concise while still reporting any pruning work or pruning failures that occurred before archiving finished or was declined.

## Acceptance criteria

- [ ] App checks unsupported platform, usage, and selected folder resolution before pruning.
- [ ] Pruning runs before archiving scan.
- [ ] Pruning no-op flows silently into archiving.
- [ ] Pruning skip flows into archiving.
- [ ] Pruning hard quit exits without archiving and without final summary.
- [ ] Pruning errors do not block archiving.
- [ ] Archiving scan errors can appear in the combined final summary when pruning already produced output or errors.
- [ ] Archiving quit after attempted pruning still shows final pruning summary.
- [ ] When pruning has no output or error and archiving has no stale items, the existing simple `Nothing to move` output remains.
- [ ] When pruning has output or error and archiving has no stale items, `Nothing to move` appears under the `Archiving` section of the final summary.
- [ ] Exit code is non-zero for pruning errors, failed prune paths, archiving errors, and failed move paths.
- [ ] Exit code is zero for user-declined archiving when no operation failed.
- [ ] App orchestration tests cover phase order, continuation behavior, final summary conditions, and exit code policy.

## Blocked by

- .scratch/archive-pruning/issues/02-discover-and-prune-archive-months-through-filesystem-adapter.md
- .scratch/archive-pruning/issues/03-add-pruning-tui-phase.md
- .scratch/archive-pruning/issues/04-refactor-archiving-runner-to-stop-owning-final-output.md
- .scratch/archive-pruning/issues/05-add-combined-final-summary.md

Commit message: `Orchestrate archive pruning and archiving phases`
