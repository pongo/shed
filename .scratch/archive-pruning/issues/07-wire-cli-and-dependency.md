# Wire the CLI and dependency

Status: ready-for-agent

## Parent

.scratch/archive-pruning/PRD.md

## What to build

Wire the completed pruning, archiving, and final phase implementations into the CLI entrypoint and add the Recycle Bin dependency. A normal shed invocation should use the filesystem pruning adapter, pruning TUI runner, refactored archiving runner, and combined final runner.

This slice is the final integration pass and should verify that the complete feature builds and passes the test suite.

## Acceptance criteria

- [ ] The Recycle Bin dependency is added to module dependencies.
- [ ] The CLI constructs and passes the filesystem pruning adapter to the app.
- [ ] The CLI constructs and passes the pruning runner to the app.
- [ ] The CLI constructs and passes the final summary runner to the app.
- [ ] Existing unsupported-platform behavior remains intact.
- [ ] `go build ./cmd/shed` succeeds.
- [ ] `go test ./...` succeeds.

## Blocked by

- .scratch/archive-pruning/issues/06-orchestrate-pruning-archiving-and-final-phases-in-app.md

Commit message: `Wire archive pruning into the CLI`
