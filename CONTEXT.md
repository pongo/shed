# shed

This context describes the language of shed, a Windows-only console utility for moving stale root items from a selected folder into the Shed.

## Language

**shed**:
A console utility that finds stale root items in a selected folder and moves them into the **Shed**.
_Avoid_: Shed, cleaner

**Shed**:
The root folder `~\Shed` where **shed** stores moved items, with `~` resolved to the current user's home directory. The Shed contains one or more **Shed buckets**.
_Avoid_: Archive, Shed Archive, destination

**Shed source**:
The **Shed** itself or any folder inside the Shed when used as the selected folder. shed never moves items from a Shed source.
_Avoid_: Archive source, Self archive

**Shed bucket**:
A folder inside the **Shed** for items moved from a selected folder during a calendar month, shaped as `~\Shed\<yyyy>\<MM>\<bucket-source-path>`. When the Selected folder is the invocation folder or one of its descendants, based on their normalized absolute paths, and the invocation folder is not a filesystem root, the bucket source path starts with the invocation folder name and continues with the Selected folder's relative path from that folder. Otherwise, the bucket source path is the Selected folder name. shed does not guarantee that the bucket source path matches the selected folder's on-disk letter casing.
_Avoid_: Archive bucket, archive destination

**Planned Shed bucket**:
The **Shed bucket** chosen for a shed run before confirmation. The confirmation view and the later **Move summary** must refer to the same Planned Shed bucket, even if the calendar month changes between confirmation and **Shedding**.
_Avoid_: Tentative destination, recalculated bucket

**Shed month**:
A calendar-month folder inside the **Shed**, shaped as `~\Shed\<yyyy>\<MM>`, containing zero or more **Shed buckets**. Its age is determined only from its `<yyyy>\<MM>` path components.
_Avoid_: Archive month, monthly bucket

**Shed pruning**:
The cleanup step that sends **Shed months** older than six months to the Recycle Bin. If a year folder becomes empty after pruning, shed also sends that year folder to the Recycle Bin.
_Avoid_: Archive pruning, Archive cleanup, pruning

**Shedding**:
The phase that moves confirmed **Selected stale items** from the **Selected folder** into a **Shed bucket**.
_Avoid_: Archiving

**Shed item**:
A file, folder, or symlink that has been moved into a **Shed bucket**.
_Avoid_: Archived item, moved file, cleaned item

**Failed move**:
An attempted move that did not complete after the user confirmed **Shedding**. For a partial **Merge**, the Failed move path is the specific nested item path that failed, not the top-level **Stale item** folder. Failed moves do not roll back successful moves.
_Avoid_: Transaction failure

**Move size**:
The total byte size that stale items will remove from the selected folder. It includes stale root files and recursive contents of stale root folders, but excludes symlink targets.
_Avoid_: Archive growth, disk growth

**Move summary**:
The final report after **Shedding** completes, containing the actual size successfully moved, the **Shed bucket** path, and any **Failed move** paths. A successful folder rename may add that folder's full **Move size** at once; a partial **Merge** adds only the sizes of nested items that moved successfully.
_Avoid_: Move log

**Prune summary**:
The final report after confirmed **Shed pruning**, containing the actual size sent to the Recycle Bin, the pruned **Shed month** paths, and any failed prune paths.
_Avoid_: Prune log, cleanup summary

**Move order**:
The deterministic order for displaying stale items in the TUI: folders alphabetically first, then files and symlinks alphabetically. shed may move stale items in any order after confirmation.
_Avoid_: Scan order

**Display name**:
The root item name shown in the TUI list, without path or metadata.
_Avoid_: Label with metadata

**Name conflict**:
A case where a **Root item** being moved has the same name as an existing item in the target **Shed bucket**, compared case-insensitively using Windows name semantics.
_Avoid_: Collision

**Nothing to move**:
The outcome when scanning finds no **Stale items** to offer for **Shedding**.
_Avoid_: Empty TUI

**Merge**:
The conflict resolution for folders with the same name: shed moves the source folder's contents into the existing target folder, applying the same conflict rules recursively.
_Avoid_: Combine, overwrite

**Numbered suffix**:
The conflict resolution for files and symlinks with the same name, inserted before the final extension as `name (N).ext`.
_Avoid_: Rename suffix, duplicate suffix

**Preflight failure**:
A run-level failure detected after confirmation while preparing the **Selected folder**, **Shed**, or **Shed bucket**, before the first attempted move. Errors after the first attempted move are reported as **Failed move** paths.
_Avoid_: Setup error

**Unsupported platform**:
Any operating system other than Windows. shed reports Unsupported platform instead of trying to scan or move files.
_Avoid_: Cross-platform mode

**Selected folder**:
The folder shed scans in a run. It is the current working directory when the user runs `shed` or `shed .`, or the single folder path passed as the command argument.
_Avoid_: Target folder, input folder

**Root item**:
A direct child of the selected folder. Items nested inside root folders are not Root items.
_Avoid_: Entry, child

**Stale item**:
A **Root item** that is eligible to be moved into the **Shed** because it is at least the configured retention age old at the retention boundary. The default retention age is 0 days.
_Avoid_: Old item, expired item

**Selected stale item**:
A **Stale item** that the user leaves selected for **Shedding** in the confirmation view. Stale items that are not selected remain in the **Selected folder** and are not **Skipped items** or **Failed moves**.
_Avoid_: Chosen file, included item

**Retention age**:
The minimum age in whole days that an eligible **Root item** must reach before shed treats it as a **Stale item**. A retention age of 0 means every eligible Root item is stale, regardless of timestamp.
_Avoid_: Expiration, TTL, cleanup age

**Skipped item**:
A **Root item** that shed does not offer for **Shedding** because it could not read required metadata or calculate **Move size** safely.
_Avoid_: Failed item

**Symlink item**:
A **Root item** that is a symbolic link. shed treats Symlink items as leaf items, not as their targets.
_Avoid_: Link target

**Hidden item**:
A **Root item** with the Windows hidden file attribute. A dot-prefixed folder is also treated as hidden by shed, but a dot-prefixed file is not.
_Avoid_: Dotfile

**Header title**:
The TUI title for the Selected folder. It is the base name of the canonical absolute Selected folder path, or the full clean path when the Selected folder is a filesystem root.
_Avoid_: Raw argument
