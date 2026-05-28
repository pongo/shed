# Implementation Plan

## Goal

Build `shed`, a Windows-only Go CLI that scans a selected folder for stale root items, shows a Bubble Tea confirmation TUI, and moves confirmed items into the Archive at `~\Shed\<yyyy>\<MM>\<source-folder-name>`.

## Product Rules

- `shed` and `shed .` scan the current working directory; `shed <path>` scans the single provided folder.
- More than one positional argument is a usage error.
- Non-folder selected paths, missing selected paths, and unsupported platforms exit non-zero with a clear error.
- The Archive is `filepath.Join(os.UserHomeDir(), "Shed")`.
- The Archive bucket uses the current move date with a two-digit month: `~\Shed\2026\05\Downloads`.
- Archive buckets are keyed only by selected folder base name, not by full path.
- The Archive itself and any folder inside it are never valid sources for cleanup.
- A root file is stale when its last modification time is `<= now - 60 days`.
- A root folder is stale when its creation time is `<= now - 60 days`; contents do not affect folder staleness.
- Symlinks are leaf items: never follow targets, never merge symlink targets, count symlink move size as `0 B`.
- Windows hidden attribute excludes a root item from candidates.
- Dot-prefixed folders are excluded; dot-prefixed files are not excluded unless they have the Windows hidden attribute.
- Root items that cannot be read safely during scanning become skipped items and are reported by path.
- If scanning finds no stale items, print `Nothing to move`, print skipped paths if any, and exit without opening the TUI.

## Move Rules

- Display and move order is deterministic: folders alphabetically first, then files and symlinks alphabetically.
- Name comparison uses Windows-style case-insensitive semantics with original name as a stable tie-breaker.
- Existing Archive buckets are reused.
- File and symlink name conflicts are resolved with a numbered suffix before the final extension:
  - `report.pdf` -> `report (1).pdf`
  - `archive.tar.gz` -> `archive.tar (1).gz`
  - `README` -> `README (1)`
- Folder conflicts use recursive merge.
- Conflicts inside merged folders use the same rules recursively.
- Moving uses `os.Rename`.
- Cross-volume moves are not supported; a failed `os.Rename` is a failed move.
- Moves are non-transactional: successful moves are not rolled back when later items fail.
- Do not recheck individual item staleness before moving; rely on the scan result.
- Before moving anything, perform preflight checks:
  - selected folder still exists as a folder;
  - Archive path can be created or already exists as a folder;
  - Archive bucket path can be created or already exists as a folder.
- A preflight failure prevents the whole move from starting.
- Final move summary reports actual successfully moved size, the resolved absolute Archive bucket path, and failed move paths. It does not list successful item paths.
- A run with any failed moves exits non-zero.

## TUI Rules

- Use Bubble Tea v2, Bubbles v2, and Lip Gloss v2.
- Do not use alternate screen.
- Header title is the base name of the canonical absolute selected folder path; for filesystem roots, show the clean full path.
- Header style:

```go
headerStyle := lipgloss.NewStyle().
	Background(lipgloss.Color("170")).Padding(0, 1).
	Foreground(lipgloss.Color("255"))
```

- Summary shows planned Move size and compact Archive bucket path using `~\Shed\yyyy\MM\<folder>`.
- Summary mentions skipped item count when skipped items exist.
- Confirmation keys:
  - `y` / `enter`: start moving;
  - `n` / `q` / `esc` / `ctrl+c`: cancel.
- Cancelled runs move nothing, print `Cancelled`, and exit zero.
- Use `bubbles/v2/help` for the footer.
- Use `bubbles/v2/list` for stale item display.
- The list is a single list with display names only: no paths, sizes, descriptions, or group headings.
- List height is responsive to `WindowSizeMsg`; reserve space for header, summary, separators, and help footer.
- Use a spinner while moving.
- Structure the TUI as a screen/state model so a future Archive cleanup screen can be inserted before confirmation without rewriting everything.
- If skipped paths exist and the TUI opened, print skipped paths after the run ends.

## Size Formatting

Use the provided integer unit formatter:

```go
func FormatSize(size int64) string {
	const unit = int64(1024)

	switch {
	case size < unit:
		return fmt.Sprintf("%d B", size)
	case size < unit*unit:
		return fmt.Sprintf("%d KB", size/unit)
	case size < unit*unit*unit:
		return fmt.Sprintf("%d MB", size/(unit*unit))
	default:
		return fmt.Sprintf("%d GB", size/(unit*unit*unit))
	}
}
```

Planned Move size means the bytes that will be removed from the selected folder: stale files plus recursive contents of stale folders, excluding symlink targets.

Final moved size means bytes successfully moved.

## Package Structure

- `cmd/shed`: CLI entrypoint, argument parsing, platform check, process exit codes.
- `internal/app`: orchestration of scan, empty result handling, TUI run, move execution, summaries.
- `internal/core`: functional core for domain models and decisions.
- `internal/fs`: Windows filesystem adapter for metadata, enumeration, size calculation, and `os.Rename` moves.
- `internal/tui`: Bubble Tea models, views, key handling, list/help/spinner integration.

## Core Responsibilities

`internal/core` should own:

- selected folder normalization decisions that do not require real filesystem calls;
- Archive bucket path construction from home dir, move date, and selected folder name;
- retention boundary evaluation;
- root item classification model;
- hidden/dot-prefixed folder eligibility rules;
- Move size calculations over abstract trees;
- Move order sorting;
- case-insensitive name conflict detection;
- numbered suffix generation;
- merge planning;
- result summaries and exit outcome classification;
- `FormatSize`.

Keep this package free of Bubble Tea and real filesystem dependencies.

## Filesystem Adapter Responsibilities

`internal/fs` should own:

- resolving actual selected folder paths;
- reading Windows metadata, including creation time and hidden attribute;
- detecting symlinks without following targets;
- enumerating root items only;
- calculating recursive folder sizes without following symlink targets;
- creating Archive and Archive bucket directories;
- moving with `os.Rename`;
- recursive merge execution for folder conflicts;
- reporting failed move paths.

Keep Windows-specific calls isolated here.

## TUI Responsibilities

`internal/tui` should own:

- confirmation screen rendering;
- help footer key map;
- list model construction from core display items;
- responsive sizing;
- cancellation and confirmation update logic;
- moving screen spinner;
- returning a typed result to `internal/app`.

Avoid placing filesystem mutation or scan decisions in TUI code.

## Test Strategy

Core tests should be table-driven and use in-memory models:

- retention boundary inclusive at exactly 60 days;
- file last-write staleness;
- folder creation-time staleness;
- symlink treated as file/leaf;
- Windows hidden attribute exclusion;
- dot-prefixed folder exclusion;
- dot-prefixed file inclusion;
- Archive source exclusion;
- Archive bucket path with two-digit month;
- bucket identity by folder name only;
- Move size excluding symlink targets;
- skipped item behavior;
- Move order folders first, then files/symlinks;
- case-insensitive sorting and tie-breaker;
- case-insensitive conflict detection;
- numbered suffix generation;
- folder merge planning;
- preflight outcome classification;
- final summary uses actual moved size.

Filesystem adapter integration tests may use `t.TempDir()` and should be Windows-only:

- hidden attribute detection;
- creation time extraction;
- symlink detection as leaf;
- root-only enumeration;
- recursive size calculation excluding symlink targets;
- `os.Rename` move success on same volume;
- file conflict rename;
- folder conflict merge;
- Archive bucket path occupied by file causes preflight failure;
- selected folder missing before move causes preflight failure.

TUI tests should avoid a real terminal:

- `y` and `enter` confirm;
- `n`, `q`, `esc`, and `ctrl+c` cancel;
- view contains header title, planned size, compact Archive bucket path, and help;
- list contains display names only;
- skipped count appears in summary when present;
- window resize updates list height.

CLI/app tests:

- no args selects current working directory;
- `.` resolves to current working directory display name;
- one path arg selects that folder;
- too many args is usage error;
- `Nothing to move` path exits zero and does not open TUI;
- cancelled run prints `Cancelled` and exits zero;
- failed moves print paths and exit non-zero;
- unsupported platform reports unsupported platform before scanning.

## Implementation Steps

1. Create `go.mod` with `module shed`.
2. Add Charm v2 dependencies.
3. Define core domain types and `FormatSize`.
4. Implement Archive path and bucket path construction.
5. Implement retention, eligibility, Move order, and conflict helpers in `internal/core`.
6. Implement Windows filesystem metadata/enumeration in `internal/fs`.
7. Implement recursive size calculation excluding symlink targets.
8. Implement move execution using `os.Rename`, numbered suffixes, and recursive folder merge.
9. Implement app orchestration and outcome types.
10. Implement Bubble Tea confirmation, moving, and result flow in `internal/tui`.
11. Wire `cmd/shed` argument parsing, platform check, app run, and exit codes.
12. Add core unit tests.
13. Add Windows adapter integration tests with `t.TempDir()`.
14. Add TUI update/view tests.
15. Run `go test ./...`.
16. Run `go build ./cmd/shed`.

Commit message: Add implementation plan for shed cleanup CLI
