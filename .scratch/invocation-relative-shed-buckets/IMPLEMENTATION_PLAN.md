# Invocation-relative Shed buckets

## Context

When `shed` is run from a project folder and the selected folder is that folder or one of its descendants, the Shed bucket should preserve that local context.

Example with `cwd = C:\Users\pavel\Documents\Projects\go\shed`:

```text
shed .scratch
=> C:\Users\pavel\Shed\2026\05\shed\.scratch

shed docs\plans
=> C:\Users\pavel\Shed\2026\05\shed\docs\plans
```

For selected folders outside the invocation folder, keep the current basename behavior.

## Decisions

- Derive the bucket source path from normalized absolute paths, not from the raw CLI argument.
- Apply the new bucket source path to both the real Shed bucket and the compact UI path.
- Keep `Header title` unchanged; it still describes the selected folder, not the bucket path.
- Do not apply invocation-relative bucket paths when the invocation folder is a filesystem root.
- Do not guarantee that the bucket source path matches the selected folder's on-disk letter casing.

## Implementation Plan

1. Add invocation-folder context to the app boundary where bucket paths are calculated.
2. Introduce a dedicated core helper for deriving the bucket source path from `invocationFolder` and `selectedFolder`.
3. Update `ShedBucket` and `CompactShedBucket` to use the bucket source path helper.
4. Keep `HeaderTitle` behavior unchanged.
5. Update mover and confirmation wiring so real and compact bucket paths use the same source-path rule.
6. Add core tests for invocation-relative, normalized-relative, outside-cwd, and filesystem-root cases.
7. Update affected app/fs/TUI expectations that currently assume the bucket segment is always the selected folder basename.
8. Run `go test ./...`.

Commit: `Derive Shed buckets from invocation-relative selected paths`
