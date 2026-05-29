# Discover and prune Archive months through the filesystem adapter

Status: ready-for-agent

## Parent

.scratch/archive-pruning/PRD.md

## What to build

Add the filesystem adapter for **Archive pruning**. It should scan the user's **Archive** for eligible **Archive months**, calculate their sizes, and after confirmation send each candidate to the Recycle Bin one at a time so the **Prune summary** can report per-path success and failure.

Missing Archive root is a no-op and must not create `~\Shed`. Invalid Archive structure is ignored. If candidate size calculation fails, the scan reports a pruning error and archiving can still continue at the app level. After a successful Archive month prune, empty parent year folder cleanup is best-effort and not reported as a separate summary item.

## Acceptance criteria

- [ ] Missing Archive root returns a no-op pruning scan result and does not create the Archive root.
- [ ] Valid old Archive months are discovered and returned with their calculated sizes.
- [ ] Non-year folders, invalid month folders, files, and unrelated Archive contents are ignored.
- [ ] Candidate size calculation failures become pruning scan errors.
- [ ] The Recycle Bin library is called once per Archive month candidate.
- [ ] Failed Recycle Bin operations record failed prune paths and do not stop later candidates.
- [ ] Successful Recycle Bin operations record pruned Archive month paths and actual pruned size.
- [ ] Empty parent year folders are sent to the Recycle Bin after their last successful month prune when possible.
- [ ] Empty year folder cleanup failures do not fail the run and do not appear in the Prune summary.
- [ ] Adapter tests use injected dependencies where needed to verify Recycle Bin calls and failure behavior.

## Blocked by

- .scratch/archive-pruning/issues/01-model-archive-month-pruning-in-core.md

Commit message: `Add filesystem adapter for archive pruning`
