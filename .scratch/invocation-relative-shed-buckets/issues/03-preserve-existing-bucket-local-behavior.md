# Preserve existing bucket-local behavior

Status: ready-for-agent

## Parent

.scratch/invocation-relative-shed-buckets/PRD.md

## What to build

Verify that nested **Shed bucket** paths do not change bucket-local behavior after the target bucket has been selected. **Name conflict**, **Merge**, **Numbered suffix**, and **Shed pruning** behavior should remain the same, with only path expectations updated where tests previously assumed basename-only buckets.

This slice is primarily regression coverage and expectation cleanup around existing behavior.

## Acceptance criteria

- [ ] Existing file and symlink **Name conflict** behavior still uses **Numbered suffix** inside the selected **Shed bucket**.
- [ ] Existing folder conflict behavior still uses **Merge** inside the selected **Shed bucket**.
- [ ] **Shed pruning** still operates on **Shed months** and is unaffected by nested bucket source paths.
- [ ] Existing final summary behavior continues to report Shedding and pruning results correctly.
- [ ] `go test ./...` passes.

## Blocked by

- .scratch/invocation-relative-shed-buckets/issues/02-shedding-uses-shared-bucket-path.md

Commit: `Preserve Shed conflict and pruning behavior`
