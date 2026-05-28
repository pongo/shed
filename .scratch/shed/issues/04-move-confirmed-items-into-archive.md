Status: ready-for-agent

# Move Confirmed Items Into Archive

## Parent

.scratch/shed/PRD.md

## What to build

Complete the confirmed cleanup path: after the user confirms in the TUI, show a spinner, run preflight checks, create or reuse the Archive bucket, move items with `os.Rename`, resolve Name conflicts, report Failed move paths, print skipped paths after the run when present, and emit the final Move summary.

This slice should make `shed` useful end-to-end for a real same-volume cleanup run.

## Acceptance criteria

- [ ] The Archive root resolves to the current user's home directory plus `Shed`.
- [ ] Archive buckets use `~\Shed\<yyyy>\<MM>\<source-folder-name>` with a two-digit month.
- [ ] Archive buckets are keyed by Selected folder base name only.
- [ ] Existing Archive buckets are reused.
- [ ] Preflight checks run before moving any items.
- [ ] A missing Selected folder before move is a Preflight failure and prevents all moves.
- [ ] Archive or Archive bucket paths occupied by files are Preflight failures and prevent all moves.
- [ ] Moving uses `os.Rename`.
- [ ] Cross-volume rename failures are reported as Failed moves; no copy fallback is implemented.
- [ ] Moves are non-transactional and successful moves are not rolled back when later items fail.
- [ ] Individual item staleness is not rechecked before moving.
- [ ] File and symlink Name conflicts use Numbered suffixes before the final extension.
- [ ] Folder Name conflicts use recursive Merge.
- [ ] Conflicts inside merged folders are resolved recursively with the same rules.
- [ ] The moving state displays a spinner.
- [ ] Final Move summary reports actual successfully moved size and the resolved absolute Archive bucket path.
- [ ] Final Move summary does not list successful item paths.
- [ ] Failed move paths are printed in the final output.
- [ ] Skipped item paths are printed after the run when skipped items exist.
- [ ] Runs with Failed moves exit non-zero.
- [ ] Successful move runs exit zero.
- [ ] Core tests cover Archive bucket construction, case-insensitive conflict detection, Numbered suffixes, Merge planning, preflight outcomes, and actual moved-size summary behavior.
- [ ] Windows adapter tests cover same-volume rename, file conflict rename, folder conflict Merge, and preflight failures with temporary directories.
- [ ] App tests cover confirmed success, partial failure, final summaries, skipped path reporting, and exit codes.

## Blocked by

- .scratch/shed/issues/02-scan-root-items-into-stale-candidates.md
- .scratch/shed/issues/03-render-confirmation-tui.md

Commit message: Move confirmed stale items into archive
