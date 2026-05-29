# Shedding uses the same bucket path end-to-end

Status: ready-for-agent

## Parent

.scratch/invocation-relative-shed-buckets/PRD.md

## What to build

Pass invocation-folder context through the application run lifecycle so **Shedding** uses the same **Shed bucket** rule in both the confirmation UI and the real move destination. The compact path shown before confirmation must match the bucket path later reported in the **Move summary**, modulo the compact `~\Shed` prefix.

Keep the confirmation **Header title** focused on the **Selected folder** display name, not the bucket source path.

## Acceptance criteria

- [ ] The confirmation UI shows the invocation-relative compact **Shed bucket** for a Selected folder inside the invocation folder.
- [ ] The real **Shed bucket** used by the mover matches the invocation-relative bucket shown in compact form before confirmation.
- [ ] The **Move summary** reports the actual invocation-relative **Shed bucket** used by Shedding.
- [ ] The confirmation **Header title** remains the selected folder's header title.
- [ ] Existing outside-invocation folder behavior remains unchanged end-to-end.

## Blocked by

- .scratch/invocation-relative-shed-buckets/issues/01-invocation-relative-bucket-source-path.md

Commit: `Use invocation-relative Shed buckets during shedding`
