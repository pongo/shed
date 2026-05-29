# Refactor archiving runner to stop owning final output

Status: ready-for-agent

## Parent

.scratch/archive-pruning/PRD.md

## What to build

Refactor the existing archiving runner so it no longer prints final summary output or `Cancelled`. The runner should still own archiving confirmation and moving progress, but final reporting belongs to the combined final summary. Archiving confirmation quit keys should return a quit outcome without treating user-declined archiving as an operational failure.

This slice preserves existing archiving behavior where possible while changing the runner contract to fit the two-phase app orchestration.

## Acceptance criteria

- [ ] Archiving runner returns outcomes and errors without printing successful move summaries.
- [ ] Archiving runner no longer prints `Cancelled`.
- [ ] `esc`, `n`, `q`, and `ctrl-c` during archiving confirmation return the same archiving quit outcome.
- [ ] Archiving move errors are returned to the app instead of being solely printed by the runner.
- [ ] Existing confirmation and moving progress behavior remains intact.
- [ ] Tests cover the absence of runner-owned final output and the archiving quit outcome.

## Blocked by

None - can start immediately

Commit message: `Refactor archiving runner final output ownership`
