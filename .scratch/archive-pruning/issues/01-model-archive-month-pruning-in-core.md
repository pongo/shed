# Model Archive month pruning in the core

Status: ready-for-agent

## Parent

.scratch/archive-pruning/PRD.md

## What to build

Add the core behavior for **Archive pruning** without touching the real filesystem. The core should understand **Archive months**, determine whether they are older than the six-month calendar cutoff, ignore invalid Archive structure, sort candidates oldest first, and accumulate a **Prune summary** that records successful pruned paths, failed prune paths, and actual pruned size.

This slice should make the pruning domain rules testable in isolation and should not depend on the Recycle Bin library or Bubble Tea.

## Acceptance criteria

- [ ] Archive month parsing accepts only calendar month folders shaped as `~\Shed\<yyyy>\<MM>` with months `01` through `12`.
- [ ] Invalid Archive folders are ignored rather than represented as errors.
- [ ] Archive month age is determined only from the `<yyyy>\<MM>` path components.
- [ ] Archive months are eligible only when they are strictly older than the month produced by subtracting six months from the current month.
- [ ] Eligible Archive months sort oldest first by year and month.
- [ ] Prune summary records successful pruned Archive month paths and adds their actual pruned size.
- [ ] Prune summary records failed prune paths without rolling back previous successes.
- [ ] Core tests cover cutoff boundaries, invalid structure, sorting, and summary accumulation.

## Blocked by

None - can start immediately

Commit message: `Model archive month pruning rules in core`
