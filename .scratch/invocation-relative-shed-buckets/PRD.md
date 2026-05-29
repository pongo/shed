# PRD: Invocation-relative Shed buckets

Status: ready-for-agent

## Problem Statement

When a user runs **shed** from a project folder and selects a folder nested inside that project, **shed** currently places **Shed items** into a **Shed bucket** named only after the selected folder. This loses the invocation-folder context.

For example, running `shed .scratch` from a folder named `shed` currently produces a bucket shaped like `~\Shed\<yyyy>\<MM>\.scratch`. The user expects the bucket to preserve the project context and produce `~\Shed\<yyyy>\<MM>\shed\.scratch`.

This matters because common nested folder names such as `.scratch`, `docs`, `plans`, or `tmp` can appear under many projects. Buckets named only by the selected folder make the **Shed** harder to scan and increase the chance that unrelated project material is merged into the same bucket.

## Solution

When the **Selected folder** is the invocation folder or one of its descendants, **shed** should derive the **Shed bucket** source path from the invocation folder name plus the selected folder's relative path from that invocation folder.

When the **Selected folder** is outside the invocation folder, **shed** should keep the existing behavior and use the selected folder name as the bucket source path.

The same rule should apply to both the real **Shed bucket** used during **Shedding** and the compact bucket path shown in the confirmation UI. The **Header title** should remain unchanged because it describes the **Selected folder**, not the **Shed bucket**.

## User Stories

1. As a shed user, I want `shed .` from a project folder to move stale root items into a bucket named after that project folder, so that the Shed preserves where the items came from.
2. As a shed user, I want `shed .scratch` from a project folder to move stale root items into a bucket under the project folder name, so that project-local scratch material does not collide with scratch folders from other projects.
3. As a shed user, I want `shed docs\plans` from a project folder to preserve the nested relative path in the Shed bucket, so that the bucket remains understandable later.
4. As a shed user, I want normalized equivalent paths like `docs\..\docs\plans` to produce the same Shed bucket as `docs\plans`, so that incidental path spelling does not change the destination.
5. As a shed user, I want absolute paths that point inside the invocation folder to use the invocation-relative bucket path, so that behavior depends on the selected folder location rather than the CLI argument text.
6. As a shed user, I want selected folders outside the invocation folder to keep the existing basename bucket behavior, so that unrelated absolute selections do not unexpectedly include my current folder name.
7. As a shed user, I want the confirmation UI to show the same bucket path that Shedding will actually use, so that I can trust the confirmation prompt.
8. As a shed user, I want the Header title to remain the selected folder's display name, so that the scan target remains easy to read.
9. As a shed user, I want filesystem-root invocations to keep the existing root behavior, so that shed does not create odd bucket source paths from drive roots.
10. As a shed user, I want Windows path normalization to decide whether a selected folder is inside the invocation folder, so that `.` and equivalent absolute paths behave consistently.
11. As a shed user, I accept that shed does not guarantee on-disk letter casing in the bucket source path, so that the feature can stay simple and predictable within the current path model.
12. As a shed user, I want existing conflict behavior inside a Shed bucket to remain unchanged, so that folder merges and numbered suffixes still work as before.
13. As a shed user, I want **Shed pruning** behavior to remain unchanged, so that monthly cleanup still operates on Shed months rather than bucket-source-path details.
14. As a shed user, I want final summaries to report the actual Shed bucket used, so that post-run output matches the move destination.
15. As a shed maintainer, I want the bucket-source-path rule isolated in the functional core, so that path behavior can be tested without moving real files.
16. As a shed maintainer, I want app-level wiring to pass invocation context explicitly, so that UI and mover paths cannot drift apart.

## Implementation Decisions

- Add invocation-folder context to the application run lifecycle and pass it to the places that calculate **Shed bucket** paths.
- Introduce a dedicated functional-core helper that derives the bucket source path from an invocation folder and a **Selected folder**.
- Use normalized absolute paths to determine whether the **Selected folder** is the invocation folder or its descendant.
- Keep existing selected-folder-name behavior when the **Selected folder** is outside the invocation folder.
- Keep existing filesystem-root behavior when the invocation folder is a filesystem root.
- Use the same bucket-source-path helper for the real **Shed bucket** and the compact confirmation path.
- Keep **Header title** behavior unchanged.
- Keep **Name conflict**, **Merge**, and **Numbered suffix** behavior unchanged once the target **Shed bucket** has been selected.
- Keep **Shed pruning** unchanged because pruning operates on **Shed months**, not on the meaning of bucket source paths.
- Avoid a new ADR for this change because it is a local behavior clarification that follows the existing functional-core/imperative-shell architecture.

## Testing Decisions

- Good tests should assert observable behavior: the computed **Shed bucket**, compact confirmation bucket path, move summary bucket path, and unchanged header title. They should not assert incidental helper internals.
- Core tests should cover invocation-relative selected folders, normalized equivalent paths, outside-invocation selected folders, and filesystem-root invocation behavior.
- Application tests should cover that confirmation data uses the invocation-relative compact bucket path while keeping the Header title unchanged.
- Filesystem mover tests should cover that real moved items land in the invocation-relative **Shed bucket**.
- Existing conflict-resolution tests should continue to cover behavior once a bucket has been selected; this feature should not duplicate those tests.
- Prior art exists in the current core path tests for **Shed bucket** and compact bucket formatting, app phase tests for confirmation data, and mover tests for real filesystem moves.

## Out of Scope

- Changing **Header title** display rules.
- Reworking **Shed pruning**.
- Changing conflict resolution for files, symlinks, or folders.
- Guaranteeing on-disk letter casing for bucket source paths.
- Supporting non-Windows platforms.
- Adding multi-argument CLI support.
- Changing retention age or stale-item selection rules.

## Further Notes

The domain glossary has been updated so **Shed bucket** now describes a `bucket-source-path` rather than only a selected folder name. The implementation should preserve the current architecture: path rules belong in the functional core, while filesystem movement remains in the imperative shell.

Commit: `Derive Shed buckets from invocation-relative selected paths`
