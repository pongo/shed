# Invocation-relative bucket source path

Status: ready-for-agent

## Parent

.scratch/invocation-relative-shed-buckets/PRD.md

## What to build

Build the core rule for deriving a **Shed bucket** source path from an invocation folder and a **Selected folder**. When the Selected folder is the invocation folder or one of its descendants, and the invocation folder is not a filesystem root, the bucket source path should start with the invocation folder name and continue with the Selected folder's relative path from that folder. When the Selected folder is outside the invocation folder, keep the existing selected-folder-name behavior.

This slice should make both real and compact **Shed bucket** path formatting capable of using the new source-path rule, while keeping **Header title** unchanged.

## Acceptance criteria

- [ ] `shed .` from a non-root invocation folder produces a bucket source path named after that invocation folder.
- [ ] A Selected folder nested under the invocation folder produces a bucket source path containing the invocation folder name plus the selected folder's relative path.
- [ ] Normalized equivalent selected paths produce the same bucket source path.
- [ ] A Selected folder outside the invocation folder keeps the existing selected-folder-name bucket behavior.
- [ ] A filesystem-root invocation folder keeps existing root behavior.
- [ ] **Header title** behavior remains unchanged.

## Blocked by

None - can start immediately

Commit: `Add invocation-relative Shed bucket source paths`
