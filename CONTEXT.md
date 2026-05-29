# shed

This context describes the language of shed, a Windows-only console utility for moving stale root items from a selected folder into an archive.

## Language

**shed**:
A console utility that finds stale root items in a selected folder and moves them into the **Archive**.
_Avoid_: Shed, cleaner

**Archive**:
The root folder `~\Shed` where **shed** stores moved items, with `~` resolved to the current user's home directory. The Archive contains one or more **Archive buckets**.
_Avoid_: Shed Archive, destination

**Archive source**:
The **Archive** itself or any folder inside the Archive when used as the selected folder. shed never moves items from an Archive source.
_Avoid_: Self archive

**Archive bucket**:
A folder inside the **Archive** for items moved from a selected folder name during a calendar month, shaped as `~\Shed\<yyyy>\<MM>\<source-folder-name>`.
_Avoid_: Archive destination

**Archive pruning**:
The cleanup step that sends **Archive buckets** older than six months to the Recycle Bin.
_Avoid_: Archive cleanup, pruning

**Cancelled run**:
A run where the user declined archiving from the TUI.
_Avoid_: Closed run

**Archived item**:
A file, folder, or symlink that has been moved into an **Archive bucket**.
_Avoid_: Moved file, cleaned item

**Failed move**:
An attempted move that did not complete after the user confirmed archiving. For a partial **Merge**, the Failed move path is the specific nested item path that failed, not the top-level **Stale item** folder. Failed moves do not roll back successful moves.
_Avoid_: Transaction failure

**Move size**:
The total byte size that stale items will remove from the selected folder. It includes stale root files and recursive contents of stale root folders, but excludes symlink targets.
_Avoid_: Archive growth, disk growth

**Move summary**:
The final report after archiving completes, containing the actual size successfully moved, the **Archive bucket** path, and any **Failed move** paths. A successful folder rename may add that folder's full **Move size** at once; a partial **Merge** adds only the sizes of nested items that moved successfully.
_Avoid_: Move log

**Move order**:
The deterministic order for displaying stale items in the TUI: folders alphabetically first, then files and symlinks alphabetically. shed may move stale items in any order after confirmation.
_Avoid_: Scan order

**Display name**:
The root item name shown in the TUI list, without path or metadata.
_Avoid_: Label with metadata

**Name conflict**:
A case where a **Root item** being moved has the same name as an existing item in the target **Archive bucket**, compared case-insensitively using Windows name semantics.
_Avoid_: Collision

**Nothing to move**:
The outcome when scanning finds no **Stale items** to offer for archiving.
_Avoid_: Empty TUI

**Merge**:
The conflict resolution for folders with the same name: shed moves the source folder's contents into the existing target folder, applying the same conflict rules recursively.
_Avoid_: Combine, overwrite

**Numbered suffix**:
The conflict resolution for files and symlinks with the same name, inserted before the final extension as `name (N).ext`.
_Avoid_: Rename suffix, duplicate suffix

**Preflight failure**:
A run-level failure detected after confirmation while preparing the **Selected folder**, **Archive**, or **Archive bucket**, before the first attempted move. Errors after the first attempted move are reported as **Failed move** paths.
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
A **Root item** that is eligible to be moved into the **Archive** because it is at least 60 days old at the retention boundary.
_Avoid_: Old item, expired item

**Skipped item**:
A **Root item** that shed does not offer for archiving because it could not read required metadata or calculate **Move size** safely.
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
