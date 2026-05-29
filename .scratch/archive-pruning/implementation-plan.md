# Archive Pruning Implementation Plan

## Decisions

- **Archive pruning** works on **Archive months** shaped as `~\Shed\<yyyy>\<MM>`, not on individual **Archive buckets**.
- An **Archive month** is eligible when its `<yyyy>\<MM>` is strictly older than the month produced by `now.AddDate(0, -6, 0)`.
- Invalid folders under `~\Shed` are ignored by pruning scan.
- If pruning scan cannot calculate a candidate size, pruning records an error and archiving still continues.
- `esc/n` in pruning skips pruning and continues to archiving.
- `q/ctrl-c` in pruning quits immediately without archiving and without final summary.
- `esc/n/q/ctrl-c` in archiving quit archiving. A final summary is still shown if pruning was attempted or produced an error.
- Successful confirmed pruning is included in the final summary, including the list of pruned **Archive months**.
- Pruning is best-effort after confirmation: failed `Archive month` paths are recorded and later candidates are still attempted.
- Empty year folder cleanup after successful month pruning is best-effort and is not included in the summary.
- `~\Shed` is not created by pruning scan when it does not exist.
- Final output is owned by a separate final Bubble Tea program.

## Architecture

1. Add `internal/core/prune.go`.
   - Define `ArchiveMonth`, `PruneCandidate`, `PruneScanResult`, and `PruneSummary`.
   - Parse year/month path components.
   - Filter candidates using the calendar-month retention rule.
   - Sort candidates oldest first by `<yyyy>\<MM>`.
   - Track successful pruned paths, failed prune paths, and actual pruned size.

2. Add `internal/fs/pruner.go`.
   - Read `ArchiveRoot` without creating it.
   - Ignore non-year folders, invalid months, files, and unrelated folders.
   - Use `RecursiveSize` to calculate candidate sizes.
   - Use `trash.Throw(path)` from `github.com/hymkor/trash-go` once per candidate for precise per-path results.
   - After successful month pruning, attempt to send an empty parent year folder to Recycle Bin as best-effort cleanup.

3. Add `internal/tui/pruning`.
   - `runner.go`: creates and runs the pruning Bubble Tea program.
   - `model.go`: owns confirmation and progress phases.
   - `confirmation.go`: renders size, confirmation prompt, and `Archive month` list.
   - Progress phase shows that Archive months are being moved to Recycle Bin.
   - No intermediate success message is printed after pruning.

4. Refactor `internal/tui/archiving`.
   - Keep archiving confirmation and moving behavior.
   - Return outcomes and errors only.
   - Remove runner-owned final summary output.
   - Remove `Cancelled` output.

5. Add `internal/tui/final`.
   - Render one combined final summary.
   - Use section headers `Archive pruning` and `Archiving`.
   - Separate sections with a blank line.
   - Omit the pruning section when pruning had no candidates or was skipped without errors.

6. Refactor `internal/app/app.go`.
   - Keep early checks in this order: platform, usage, selected folder resolve.
   - Run pruning before archiving scan.
   - Run archiving scan/move after pruning, even when pruning errors occurred.
   - Aggregate pruning and archiving outcomes into a run summary.
   - Run final summary when there is pruning output/error or archiving output/error to report.
   - Derive exit code from real operation errors and failed paths, not from user-declined archiving.

7. Update `cmd/shed/main.go`.
   - Wire `fs.Pruner`.
   - Wire `tui/pruning.Runner`.
   - Wire refactored `tui/archiving.Runner`.
   - Wire `tui/final.Runner`.

## Tests

1. Core tests.
   - Calendar-month cutoff behavior.
   - Invalid archive structure is ignored.
   - Archive months sort oldest first.
   - `PruneSummary` records successful and failed prune paths.

2. Filesystem adapter tests.
   - Missing `~\Shed` is no-op and does not create the Archive root.
   - Valid old months become candidates.
   - Invalid folders are ignored.
   - Size errors become pruning scan errors.
   - Injected trash function is called once per candidate.
   - Failed candidate does not stop later candidates.
   - Empty parent year cleanup is best-effort.

3. Pruning TUI tests.
   - `y/enter` confirms.
   - `esc/n` skips and continues.
   - `q/ctrl-c` hard quits.
   - Confirmation view shows total size and `Archive month` paths.
   - Progress view renders during pruning.

4. Archiving TUI tests.
   - Runner no longer prints `Cancelled`.
   - Runner no longer prints final summary.
   - Archiving quit outcome is returned for `esc/n/q/ctrl-c`.

5. Final TUI tests.
   - Successful pruning section lists pruned Archive months.
   - Failed prune paths are shown.
   - Archiving summary is separated by a blank line.
   - Pruning section is omitted for no-op/skipped pruning.

6. App orchestration tests.
   - Pruning no-op flows directly into archiving.
   - Pruning skip flows into archiving.
   - Pruning hard quit exits without archiving.
   - Pruning errors do not block archiving, but produce `ExitError`.
   - Archiving quit after attempted pruning still shows final pruning summary.
   - `Nothing to move` remains immediate output when pruning had no output/error.
   - `Nothing to move` appears in final summary when pruning has output/error.
   - Failed prune paths make exit code non-zero.

Commit message: `Add archive pruning phase with combined final summary`
