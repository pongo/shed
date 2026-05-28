Status: ready-for-agent

# Scan Root Items Into Stale Candidates

## Parent

.scratch/shed/PRD.md

## What to build

Implement the scan path that turns the Selected folder into Stale items, Skipped items, Display names, and planned Move size. The scan should inspect only Root items, apply the retention boundary and eligibility rules, protect Archive sources, and isolate Windows filesystem behavior behind the filesystem adapter.

This slice should make the app capable of producing a complete scan result without yet requiring a confirmation TUI or move execution.

## Acceptance criteria

- [ ] Root files become Stale items when last modification time is `<= now - 60 days`.
- [ ] Root folders become Stale items when creation time is `<= now - 60 days`.
- [ ] Folder contents do not affect folder staleness.
- [ ] Nested items are not listed as Root items.
- [ ] Symlink items are treated as leaf items and their targets are never followed.
- [ ] Symlink targets do not contribute to Move size.
- [ ] Windows hidden root items are excluded from candidates.
- [ ] Dot-prefixed folders are excluded from candidates.
- [ ] Dot-prefixed files are not excluded unless they have the Windows hidden attribute.
- [ ] The Archive and folders inside the Archive are rejected as Archive sources.
- [ ] Items that cannot be read safely become Skipped items reported by path.
- [ ] Move size includes stale files and recursive contents of stale folders, excluding symlink targets.
- [ ] Core tests cover retention, eligibility, Archive source protection, Move size, Move order inputs, and skipped behavior with in-memory models.
- [ ] Windows adapter tests cover root-only enumeration, hidden attributes, creation time extraction, symlink leaf detection, and recursive size calculation with temporary directories.

## Blocked by

- .scratch/shed/issues/01-bootstrap-cli-with-empty-scan-path.md

Commit message: Add stale root item scanning
