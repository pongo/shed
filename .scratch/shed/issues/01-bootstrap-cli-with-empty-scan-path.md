Status: ready-for-agent

# Bootstrap CLI With Empty Scan Path

## Parent

.scratch/shed/PRD.md

## What to build

Create the initial `shed` CLI skeleton and the smallest complete app path: parse arguments, reject unsupported platforms and invalid usage, resolve the Selected folder, and handle a scan result with no Stale items by printing `Nothing to move` without opening the TUI.

This slice establishes the module, command entrypoint, app orchestration shape, exit-code convention, and first tests. It should not implement real stale scanning or moving yet; the app can use a stubbed scanner that returns no Stale items.

## Acceptance criteria

- [ ] A Go module exists for the project.
- [ ] Running `shed` selects the current working directory.
- [ ] Running `shed .` selects the current working directory.
- [ ] Running `shed <path>` selects the provided folder.
- [ ] More than one positional argument reports a usage error and exits non-zero.
- [ ] A missing path or non-folder path reports a clear error and exits non-zero.
- [ ] On non-Windows platforms, shed reports Unsupported platform before scanning and exits non-zero.
- [ ] When there are no Stale items, shed prints `Nothing to move` and exits zero.
- [ ] The empty scan path does not initialize or run the TUI.
- [ ] Tests cover argument handling, selected-folder resolution, unsupported platform behavior, invalid selected folders, and `Nothing to move`.

## Blocked by

None - can start immediately

Commit message: Bootstrap shed CLI with empty scan flow
