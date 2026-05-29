# PRD: Archive Pruning

Status: ready-for-agent

## Problem Statement

shed stores archived items under the user's **Archive** by calendar month. Over time, old **Archive months** accumulate and there is no built-in way to remove them from the Archive without manually navigating `~\Shed` and deciding which month folders are safe to delete.

Users need shed to detect **Archive months** older than six months, ask for confirmation before moving them to the Recycle Bin, and then continue with the normal archiving workflow. The final output must clearly summarize both **Archive pruning** and archiving when both phases matter in the same run.

## Solution

Add an **Archive pruning** phase before the existing archiving phase. On startup, shed scans the Archive for eligible **Archive months**. If none are found, shed proceeds silently into archiving. If eligible months are found, shed shows a confirmation screen with the total size and the list of `~\Shed\<yyyy>\<MM>` folders that will be moved to the Recycle Bin.

When the user confirms pruning, shed sends each eligible **Archive month** to the Recycle Bin, records a **Prune summary**, and proceeds to archiving without an intermediate success message. If the user skips pruning, shed proceeds to archiving. If the user quits from pruning, shed exits immediately.

At the end of the run, shed renders a combined final summary when there is pruning output, pruning error, archiving output, or archiving error to report. The final summary uses separate sections for `Archive pruning` and `Archiving` so the two phases do not read as one stream of text.

## User Stories

1. As a shed user, I want old Archive months to be detected automatically on startup, so that I do not have to manually inspect `~\Shed`.
2. As a shed user, I want Archive pruning to work on `~\Shed\<yyyy>\<MM>` folders, so that an entire calendar month of archived items is handled together.
3. As a shed user, I want Archive month age to be determined from the `<yyyy>\<MM>` path, so that pruning behavior is predictable and independent of filesystem timestamps.
4. As a shed user, I want only Archive months older than six calendar months to be offered for pruning, so that recent archived data remains available.
5. As a shed user, I want invalid folders under the Archive to be ignored, so that unrelated files or folders do not create noisy warnings.
6. As a shed user, I want pruning to do nothing when `~\Shed` does not exist, so that a normal first run does not create empty Archive folders.
7. As a shed user, I want a pruning confirmation screen before anything is sent to the Recycle Bin, so that I stay in control of Archive removal.
8. As a shed user, I want the pruning confirmation screen to show the total size, so that I understand the amount of data being removed from the Archive.
9. As a shed user, I want the pruning confirmation screen to list each Archive month path, so that I know exactly which folders will be moved to the Recycle Bin.
10. As a shed user, I want `y` and `enter` to confirm pruning, so that the confirmation flow matches the existing archiving interaction.
11. As a shed user, I want `esc` and `n` during pruning to skip pruning and continue to archiving, so that I can defer Archive pruning without stopping the run.
12. As a shed user, I want `q` and `ctrl-c` during pruning to quit immediately, so that I can stop before any archiving work begins.
13. As a shed user, I want confirmed pruning to show progress while Recycle Bin operations are running, so that the terminal does not look stuck on large Archive months.
14. As a shed user, I want shed to continue trying later Archive months when one pruning attempt fails, so that one locked month does not block cleanup of other old months.
15. As a shed user, I want failed prune paths in the final summary, so that I know which Archive months still require attention.
16. As a shed user, I want successful pruned Archive months listed in the final summary, so that I can verify what happened after confirmation.
17. As a shed user, I want empty year folders to be cleaned up after their last pruned Archive month is removed, so that `~\Shed` does not accumulate empty year directories.
18. As a shed user, I do not want empty year folder cleanup shown as a separate pruned item, so that the summary only lists Archive months I confirmed.
19. As a shed user, I want pruning scan errors to be reported in the final summary, so that Archive access problems are visible.
20. As a shed user, I want pruning scan errors not to block archiving, so that an Archive maintenance problem does not prevent moving stale items from the selected folder.
21. As a shed user, I want archiving to run after pruning succeeds, so that a normal shed run can both prune old Archive months and archive stale root items.
22. As a shed user, I want archiving to run after pruning is skipped, so that declining Archive pruning does not cancel the normal archiving workflow.
23. As a shed user, I want archiving to run even when the selected folder later turns out to be an Archive source, so that Archive pruning can still happen for Archive maintenance runs.
24. As a shed user, I want `esc`, `n`, `q`, and `ctrl-c` during archiving confirmation to quit archiving without printing `Cancelled`, so that the output reflects only completed or failed operations.
25. As a shed user, I want a final pruning summary even if I quit archiving afterward, so that already confirmed pruning is not hidden.
26. As a shed user, I want no final summary when pruning had no candidates and I quit archiving before moving anything, so that a declined no-op run stays quiet.
27. As a shed user, I want `Nothing to move` to remain the simple output when there was no pruning output and no stale items, so that ordinary empty runs stay concise.
28. As a shed user, I want `Nothing to move` inside the combined final summary when pruning already produced output, so that both phase outcomes are visible together.
29. As a shed user, I want final output separated into `Archive pruning` and `Archiving` sections, so that I can read the results of each phase independently.
30. As a shed user, I want a non-zero exit code when pruning or archiving has real errors or failed paths, so that scripts can detect operational failures.
31. As a shed user, I want user-declined archiving to exit successfully when no operation failed, so that intentional quit behavior is not treated as an error.

## Implementation Decisions

- Add a core pruning module that models Archive month parsing, retention filtering, deterministic sorting, and Prune summary accumulation without touching the real filesystem.
- Add a filesystem pruning adapter that owns Archive root traversal, size calculation, Recycle Bin integration, and empty year folder cleanup.
- Keep the `trash-go` dependency in the filesystem adapter. The core and TUI layers should not depend on the Recycle Bin library.
- Call the Recycle Bin library once per Archive month candidate. This supports precise per-path success and failure reporting.
- Treat missing Archive root as pruning no-op. Pruning scan must not create `~\Shed`.
- Treat invalid Archive structure as irrelevant. Non-year folders, invalid month names, files, and unrelated folders are ignored by pruning scan.
- Treat candidate size calculation failure as a pruning error. Archiving still continues, and the final summary reports the pruning error.
- Apply calendar-month retention. An Archive month is eligible only when its `<yyyy>\<MM>` is strictly earlier than the month produced by subtracting six months from the current month.
- Sort Archive month candidates oldest first. Confirmation and final summaries use the same order.
- Add a pruning TUI runner with confirmation and progress phases.
- Preserve different key semantics by phase. Pruning `esc/n` skips and continues, pruning `q/ctrl-c` quits immediately, and archiving `esc/n/q/ctrl-c` quits archiving.
- Refactor archiving so its runner returns outcomes and errors but does not own final summary output.
- Add a final TUI runner that renders the combined run summary.
- Render final output with `Archive pruning` and `Archiving` sections separated by a blank line.
- Omit the pruning section when pruning had no candidates or was skipped without output or errors.
- Include successful pruned Archive month paths in Prune summary output.
- Include failed prune paths in Prune summary output.
- Make pruning best-effort after confirmation. Failed candidates do not stop later candidates.
- Make empty year folder cleanup best-effort and exclude it from summary and exit-code decisions.
- Keep early app checks in this order: unsupported platform, usage, selected folder resolution.
- Run pruning before archiving scan.
- Continue into archiving after pruning errors, then derive the final exit code from all phase errors and failed paths.
- Record the architecture decision in the ADR for the combined final summary: phase runners return outcomes and the app aggregates results for a final Bubble Tea program.

## Testing Decisions

- Tests should assert externally visible behavior and stable contracts: domain decisions, runner outcomes, rendered text, exit codes, and adapter side effects through injected dependencies.
- Core pruning tests should cover calendar-month cutoff behavior, invalid Archive structure filtering, oldest-first sorting, and Prune summary success/failure accumulation.
- Filesystem pruning adapter tests should cover missing Archive root no-op behavior, candidate discovery, invalid folder ignoring, size errors, one Recycle Bin call per candidate, best-effort continuation after failures, and best-effort empty year cleanup.
- Pruning TUI tests should cover confirmation keys, skip keys, hard quit keys, confirmation view content, and progress view content.
- Archiving TUI tests should cover the changed ownership contract: no `Cancelled` output, no runner-owned final summary, and quit outcome for archiving confirmation keys.
- Final TUI tests should cover section headings, blank-line separation, successful pruned Archive month list, failed prune paths, archiving summary, omitted pruning section for no-op/skipped pruning, and `Nothing to move` in combined output.
- App orchestration tests should cover phase ordering, pruning no-op flow, pruning skip flow, pruning hard quit flow, pruning errors continuing into archiving, final summary after archiving quit when pruning was attempted, simple `Nothing to move`, combined `Nothing to move`, failed prune exit code, and failed archiving exit code.
- Existing archiving scan and move tests are useful prior art for functional-core testing and adapter testing; new pruning tests should follow the same shape.

## Out of Scope

- Configurable pruning retention.
- Pruning individual Archive buckets.
- Permanent deletion without the Recycle Bin.
- Showing invalid Archive folders as warnings.
- Reporting empty year folder cleanup as a user-visible operation.
- Changing stale item retention for normal archiving.
- Changing Archive bucket conflict resolution.
- Cross-platform support beyond existing Windows-only app behavior.
- Shell-conventional `130` exit code for `ctrl-c`.
- Creating implementation issues; this PRD is the feature specification only.

## Further Notes

- The domain glossary now includes **Archive month** and **Prune summary**, and **Archive pruning** now refers to Archive months rather than Archive buckets.
- The old **Cancelled run** term was removed because archiving confirmation quit no longer prints or models `Cancelled` as a domain outcome.
- The implementation plan for this feature is stored alongside this PRD.
- The agreed commit message for the eventual implementation is: `Add archive pruning phase with combined final summary`
