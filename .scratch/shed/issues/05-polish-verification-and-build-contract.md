Status: ready-for-agent

# Polish Verification And Build Contract

## Parent

.scratch/shed/PRD.md

## What to build

Finish the feature by tightening UX text, checking glossary consistency, filling remaining test gaps, and verifying the documented build and test contract. This slice should leave `shed` ready for normal local use and future issue work.

## Acceptance criteria

- [ ] User-facing terms match the project glossary where applicable.
- [ ] TUI text is concise and visually consistent with the agreed behavior.
- [ ] Terminal output is consistent for `Nothing to move`, `Cancelled`, successful Move summary, Preflight failure, Unsupported platform, usage errors, and Failed moves.
- [ ] The test suite covers the PRD's core, filesystem adapter, TUI, CLI, and app behavior decisions.
- [ ] Tests avoid real filesystem usage except in focused Windows adapter integration tests.
- [ ] No config or logging behavior has been introduced.
- [ ] No Archive cleanup screen has been implemented.
- [ ] No cross-volume copy fallback has been implemented.
- [ ] `go test ./...` passes.
- [ ] `go build ./cmd/shed` succeeds.
- [ ] The implementation remains aligned with the functional core, imperative shell ADR.

## Blocked by

- .scratch/shed/issues/04-move-confirmed-items-into-archive.md

Commit message: Polish shed verification and build contract
