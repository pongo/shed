Status: ready-for-agent

# Render Confirmation TUI

## Parent

.scratch/shed/PRD.md

## What to build

Build the Bubble Tea confirmation TUI for a populated scan result. The TUI should render the Header title, planned Move size, compact Archive bucket path, skipped count, a display-name-only list, and a help footer. It should return typed confirm or cancel outcomes to the app without performing filesystem mutation.

The TUI should be structured as a state model that can later accept an Archive cleanup screen before confirmation, but this slice should only implement the confirmation path.

## Acceptance criteria

- [ ] Bubble Tea v2, Bubbles v2, and Lip Gloss v2 are used.
- [ ] Alternate screen is not used.
- [ ] The Header title shows the resolved Selected folder name, not the raw argument.
- [ ] Filesystem roots show a clean full path as Header title.
- [ ] The header uses the agreed magenta background, white foreground, and horizontal padding style.
- [ ] The summary shows planned Move size.
- [ ] The summary shows the compact Archive bucket path as `~\Shed\<yyyy>\<MM>\<folder>`.
- [ ] The summary includes skipped item count when skipped items exist.
- [ ] The list is a single paginated list of Display names only.
- [ ] The list has no full paths, sizes, descriptions, or group headings.
- [ ] List ordering is folders alphabetically first, then files and symlinks alphabetically, using case-insensitive comparison with stable tie-breaker.
- [ ] The help footer uses `bubbles/v2/help`.
- [ ] `y` and `enter` return a confirm outcome.
- [ ] `n`, `q`, `esc`, and `ctrl+c` return a Cancelled run outcome.
- [ ] Cancelled runs print `Cancelled` and exit zero through the app flow.
- [ ] Window resize updates list height while reserving space for header, summary, separators, and footer.
- [ ] TUI tests cover key handling, rendered text, display-name-only list content, skipped count, and responsive sizing without a real terminal.

## Blocked by

- .scratch/shed/issues/01-bootstrap-cli-with-empty-scan-path.md
- .scratch/shed/issues/02-scan-root-items-into-stale-candidates.md

Commit message: Add confirmation TUI for stale items
