# Add the pruning TUI phase

Status: ready-for-agent

## Parent

.scratch/archive-pruning/PRD.md

## What to build

Add the interactive **Archive pruning** phase. When pruning candidates exist, shed should show a confirmation screen with the total size that will be moved to the Recycle Bin and the list of eligible **Archive month** paths. After confirmation, shed should show a progress state while pruning runs, then return control to the app without printing an intermediate success message.

The pruning TUI owns pruning-specific key semantics: `y/enter` confirms, `esc/n` skips pruning and continues to archiving, and `q/ctrl-c` quits immediately without archiving or final summary.

## Acceptance criteria

- [ ] The confirmation view shows the total candidate size and asks the user to press `y/enter` to confirm.
- [ ] The confirmation view lists eligible Archive month paths in oldest-first order.
- [ ] `y` and `enter` produce a confirmed pruning outcome.
- [ ] `esc` and `n` produce a skipped pruning outcome.
- [ ] `q` and `ctrl-c` produce a hard quit pruning outcome.
- [ ] Confirmed pruning runs through a progress phase instead of leaving the confirmation view frozen.
- [ ] The runner returns pruning outcome, Prune summary, and error information without printing final summary output.
- [ ] TUI tests cover confirmation keys, skip keys, hard quit keys, confirmation content, and progress content.

## Blocked by

- .scratch/archive-pruning/issues/01-model-archive-month-pruning-in-core.md

Commit message: `Add archive pruning TUI phase`
