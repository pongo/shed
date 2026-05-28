Status: ready-for-agent

# PRD: shed cleanup CLI

## Problem Statement

Windows folders such as Downloads, temporary work directories, and project scratch folders accumulate old root files and folders. Manually deciding what to move is repetitive, and deleting is too destructive for a low-friction cleanup workflow.

The user wants `shed`: a small, beautiful Windows-only console utility that scans a Selected folder, identifies Stale items, previews the cleanup in a TUI, and moves confirmed items into an Archive instead of deleting them. The tool should require no config, no logs, and no cross-platform behavior.

## Solution

`shed` scans the Selected folder, finds eligible root items, and moves them into the Archive at `~\Shed\<yyyy>\<MM>\<source-folder-name>`. Before moving, it shows a Bubble Tea confirmation UI with the Header title, planned Move size, compact Archive bucket path, and a paginated list of Display names.

The move operation is conservative and transparent. Hidden items, Archive sources, dot-prefixed folders, unreadable items, and unsupported platforms are handled explicitly. Existing Archive buckets are reused. File and symlink conflicts receive Numbered suffixes; folder conflicts are merged recursively. Moves use `os.Rename`, so cross-volume moves are not supported in the first version.

The implementation follows functional core, imperative shell. Domain decisions live in a pure, highly tested core. Windows filesystem access and TUI interaction stay in thin adapters.

## User Stories

1. As a Windows user, I want to run `shed` in the current folder, so that I can clean that folder without typing a path.
2. As a Windows user, I want `shed .` to behave like `shed`, so that common CLI habits work.
3. As a Windows user, I want to run `shed <path>`, so that I can clean a folder without changing directories.
4. As a Windows user, I want invalid paths to fail clearly, so that I know the command did not scan anything.
5. As a Windows user, I want too many command arguments to fail as usage errors, so that accidental invocations are not ambiguous.
6. As a Windows user, I want shed to reject non-Windows platforms, so that unsupported filesystem semantics do not produce surprising behavior.
7. As a Windows user, I want old root files to be detected by last modification time, so that files I have not changed recently are proposed for archiving.
8. As a Windows user, I want old root folders to be detected by creation time, so that folder eligibility is simple and predictable.
9. As a Windows user, I want folders to be judged without inspecting their contents for staleness, so that scan behavior is easy to understand.
10. As a Windows user, I want exactly 60-day-old items to count as stale, so that the retention boundary is predictable.
11. As a Windows user, I want nested items to stay out of the top-level candidate list, so that the confirmation screen remains focused on root items.
12. As a Windows user, I want Windows hidden items to be ignored, so that intentionally hidden data is not offered for archiving.
13. As a Windows user, I want dot-prefixed folders such as `.git` to be ignored, so that important technical folders are not moved accidentally.
14. As a Windows user, I want dot-prefixed files such as `.env` to be treated like normal files, so that Windows hidden semantics still control file hiding.
15. As a Windows user, I want symlinks to be treated as leaf items, so that shed never follows a link into unrelated data.
16. As a Windows user, I want symlink targets excluded from Move size, so that summary sizes do not include data outside the Selected folder.
17. As a Windows user, I want the Archive itself to be protected from cleanup, so that shed cannot recursively archive its own storage.
18. As a Windows user, I want folders inside the Archive to be protected from cleanup, so that archived data is not reprocessed.
19. As a Windows user, I want Archive buckets grouped by year, two-digit month, and Selected folder name, so that archived data is easy to browse.
20. As a Windows user, I want Archive buckets to be keyed by folder name only, so that the Archive path stays readable.
21. As a Windows user, I want existing Archive buckets to be reused, so that repeated monthly cleanup accumulates in one readable location.
22. As a Windows user, I want the confirmation header to show the real folder name rather than `.`, so that I can verify what I am cleaning.
23. As a Windows user, I want filesystem roots to show as clear full paths in the header, so that root-folder runs are not visually blank.
24. As a Windows user, I want the confirmation UI to look polished, so that the tool feels trustworthy before moving files.
25. As a Windows user, I want the confirmation summary to show the planned Move size, so that I understand the scale of the cleanup.
26. As a Windows user, I want the confirmation summary to show a compact Archive bucket path, so that I know where items will go.
27. As a Windows user, I want skipped item counts to be visible, so that I know the scan was not fully complete.
28. As a Windows user, I want stale items listed as names only, so that the list remains easy to scan.
29. As a Windows user, I want folders listed before files, so that larger structural cleanup candidates are visible first.
30. As a Windows user, I want alphabetical ordering to be deterministic, so that the list is stable between runs.
31. As a Windows user, I want Windows-style case-insensitive ordering and conflict checks, so that shed matches filesystem expectations.
32. As a Windows user, I want a paginated list, so that large folders remain navigable.
33. As a Windows user, I want help keys in the footer, so that confirmation and cancellation are discoverable.
34. As a Windows user, I want `y` and `enter` to confirm, so that accepting the move is fast.
35. As a Windows user, I want `n`, `q`, `esc`, and `ctrl+c` to cancel, so that leaving the tool is natural.
36. As a Windows user, I want cancelled runs to print `Cancelled`, so that the terminal history records that no move happened.
37. As a Windows user, I want no full TUI when there is nothing to move, so that empty cleanups finish quickly.
38. As a Windows user, I want skipped paths printed when nothing can be moved, so that I can inspect access problems.
39. As a Windows user, I want a spinner while files are being moved, so that I can see work is in progress.
40. As a Windows user, I want file name conflicts to keep both files, so that archived data is not overwritten.
41. As a Windows user, I want conflict suffixes to appear before the final extension, so that renamed files stay readable.
42. As a Windows user, I want folder conflicts to merge recursively, so that repeated cleanups consolidate related data.
43. As a Windows user, I want symlink conflicts to be renamed like files, so that symlink targets are never merged.
44. As a Windows user, I want failed moves not to roll back successful moves, so that cleanup can still make progress.
45. As a Windows user, I want failed move paths printed, so that I know exactly which items still need attention.
46. As a Windows user, I want the final summary to show actual moved size, so that partial failures are represented honestly.
47. As a Windows user, I want the final summary to show the resolved absolute Archive bucket path, so that I can open the result directly.
48. As a Windows user, I do not want successful paths listed in the final summary, so that normal runs stay concise.
49. As a Windows user, I want runs with failed moves to exit non-zero, so that scripts can detect incomplete cleanup.
50. As a future maintainer, I want the TUI shaped as a state model, so that an Archive cleanup screen can be added before confirmation later.
51. As a future maintainer, I want domain behavior tested without real files where possible, so that tests are fast and precise.
52. As a future maintainer, I want Windows filesystem behavior isolated, so that platform-specific edge cases have focused tests.

## Implementation Decisions

- Use Go with a Windows-only product boundary.
- Use Bubble Tea v2, Bubbles v2, and Lip Gloss v2 for the TUI.
- Do not use alternate screen.
- Use `os.UserHomeDir()` to resolve the Archive root.
- Use `os.Rename` for moves.
- Do not implement cross-volume copy fallback in the first version.
- Do not implement configuration files or logging.
- Use functional core, imperative shell as recorded in the architecture decision.
- Build a CLI entry module for argument parsing, platform checks, app invocation, and exit codes.
- Build an app orchestration module for scan flow, empty-result behavior, TUI invocation, move execution, terminal summaries, and outcome classification.
- Build a core module as a deep module for domain rules, sorting, Archive bucket construction, conflict naming, merge planning, size formatting, and summary decisions.
- Build a filesystem module as the Windows adapter for metadata, enumeration, hidden attributes, creation time, symlink detection, size calculation, directory creation, and `os.Rename` move execution.
- Build a TUI module for Bubble Tea models, views, list/help/spinner integration, responsive sizing, and key handling.
- Keep Bubble Tea dependencies out of the core module.
- Keep real filesystem dependencies out of the core module.
- Keep filesystem mutation out of the TUI module.
- Treat the scan result as authoritative for individual items after confirmation; do not recheck item staleness before moving.
- Perform run-level preflight checks before moving any item.
- Reuse existing Archive buckets.
- Resolve file and symlink Name conflicts with Numbered suffixes.
- Resolve folder Name conflicts with recursive Merge.
- Use actual moved size, not planned Move size, in the final Move summary.
- Return zero for successful move, Nothing to move, and Cancelled run.
- Return non-zero for usage errors, invalid selected folder, Unsupported platform, Preflight failure, scan-level fatal errors, and runs with Failed moves.

## Testing Decisions

- Good tests should verify observable behavior and domain outcomes, not private implementation details.
- Core tests should be table-driven and use in-memory models.
- Core tests should cover retention boundaries, file and folder staleness, symlink leaf behavior, hidden and dot-prefix eligibility, Archive source exclusion, Archive bucket construction, Move size, skipped item behavior, Move order, conflict detection, Numbered suffixes, Merge planning, preflight outcomes, final summaries, and size formatting.
- Filesystem adapter tests may use temporary directories because Windows metadata, hidden attributes, creation time, symlink behavior, and rename behavior must be verified against the real filesystem.
- Filesystem adapter tests should be Windows-only.
- Filesystem adapter tests should cover root-only enumeration, hidden attribute detection, creation time extraction, symlink detection as a leaf, recursive size calculation excluding symlink targets, same-volume rename success, file conflict rename, folder conflict merge, Archive bucket path occupied by a file, and missing Selected folder preflight.
- TUI tests should avoid a real terminal and focus on update results and rendered view text.
- TUI tests should cover confirmation keys, cancellation keys, header rendering, planned size rendering, compact Archive bucket path rendering, help footer rendering, display-name-only list content, skipped count rendering, and responsive list height.
- CLI/app tests should cover argument handling, current directory selection, dot path resolution, usage errors, Nothing to move behavior, cancellation output, failed move output, exit codes, and unsupported platform behavior.
- The existing architecture decision is prior art for test boundaries: core behavior belongs in in-memory tests, while the thin Windows adapter gets focused integration coverage.

## Out of Scope

- Deleting files.
- Cleaning or compacting the Archive.
- A pre-confirmation Archive cleanup screen.
- Config files.
- Logs.
- Cross-platform support.
- Cross-volume copy fallback.
- Progress bars for scanning.
- Scan cancellation beyond normal process interruption.
- Full recursive stale-item discovery.
- Rechecking individual item staleness after confirmation.
- Listing successful item paths in the final summary.
- Alternate screen mode.
- Watch mode, scheduling, or background automation.

## Further Notes

- The detailed sequencing plan lives in the implementation plan for this feature.
- Domain vocabulary is defined in the project glossary and should be used consistently in implementation, tests, and future issues.
- The first implementation should leave room in the TUI state model for a future Archive cleanup screen, but should not implement that screen.
