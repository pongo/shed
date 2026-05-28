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
A folder inside the **Archive** for items moved from a selected folder name during a calendar month, shaped as `~\Shed\<yyyy>\<MM>\<source-folder-name>` with a two-digit month. The same Archive bucket can receive items from multiple shed runs and from different selected folders that have the same name.
_Avoid_: Archive destination

**Cancelled run**:
A run where the user declined archiving from the TUI. shed moves nothing and reports `Cancelled`.
_Avoid_: Closed run

**Archived item**:
A file, folder, or symlink that has been moved into an **Archive bucket**.
_Avoid_: Moved file, cleaned item

**Failed move**:
An attempted `os.Rename` move that did not complete after the user confirmed archiving. A Failed move is reported with the path of the item that could not be moved; it does not roll back successful moves, and the run exits non-zero. shed does not recheck item staleness before moving and relies on the scan result.
_Avoid_: Transaction failure

**Move size**:
The total byte size that stale items will remove from the selected folder. It includes stale root files and recursive contents of stale root folders, but excludes symlink targets.
_Avoid_: Archive growth, disk growth

**Move summary**:
The final report after archiving completes. It shows the actual size successfully moved, the Archive bucket path, and any **Failed move** paths, but not successful item paths.
_Avoid_: Move log

**Move order**:
The deterministic order for displaying and moving stale items: folders alphabetically first, then files and symlinks alphabetically. Names are compared case-insensitively, with the original name as a stable tie-breaker.
_Avoid_: Scan order

**Display name**:
The root item name shown in the TUI list. The list uses Display names only, without full paths, sizes, descriptions, or group headings.
_Avoid_: Label with metadata

**Name conflict**:
A case where a **Root item** being moved has the same name as an existing item in the target **Archive bucket**, compared case-insensitively using Windows name semantics.
_Avoid_: Collision

**Nothing to move**:
The outcome when scanning finds no **Stale items** to offer for archiving. shed reports this as a plain terminal message and exits without opening the TUI; any **Skipped items** are still reported by path.
_Avoid_: Empty TUI

**Merge**:
The conflict resolution for folders with the same name: shed moves the source folder's contents into the existing target folder, applying the same conflict rules recursively.
_Avoid_: Combine, overwrite

**Numbered suffix**:
The conflict resolution for files and symlinks with the same name, inserted before the final extension as `name (N).ext`.
_Avoid_: Rename suffix, duplicate suffix

**Preflight failure**:
A run-level failure detected after confirmation but before moving any items, such as the Selected folder no longer being a folder or the Archive bucket path being occupied by a file. A Preflight failure prevents the whole move from starting.
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
A **Root item** that is eligible to be moved into the **Archive** because it is at least 60 days old at the retention boundary. A root file or symlink is stale by last modification time; a root folder is stale by creation time, without inspecting its contents.
_Avoid_: Old item, expired item

**Skipped item**:
A **Root item** that shed does not offer for archiving because it could not read required metadata or calculate **Move size** safely. If the TUI opens, skipped items are summarized there and reported by path after the run ends.
_Avoid_: Failed item

**Symlink item**:
A **Root item** that is a symbolic link. shed treats a Symlink item as a leaf item: it never follows the link, never inspects the target, and never applies **Merge** to it.
_Avoid_: Link target

**Hidden item**:
A **Root item** with the Windows hidden file attribute. A dot-prefixed folder is also treated as hidden by shed, but a dot-prefixed file is not.
_Avoid_: Dotfile

**Header title**:
The TUI title for the Selected folder. It is the base name of the canonical absolute Selected folder path, or the full clean path when the Selected folder is a filesystem root.
_Avoid_: Raw argument

## Example Dialogue

Developer: "Where should shed move stale items from `Downloads`?"

Domain expert: "Into the Archive, under the Archive bucket for the current year, month, and `Downloads`."

Developer: "What is the Selected folder when the user runs `shed` without arguments?"

Domain expert: "The current working directory."

Developer: "Should the TUI header show `.` when the user runs `shed .`?"

Domain expert: "No. It shows the Header title of the resolved Selected folder."

Developer: "After the move, what do we call those files and folders?"

Domain expert: "They are Archived items."

Developer: "If a folder was created long ago but contains files modified yesterday, is it stale?"

Domain expert: "Yes. A folder's staleness is based on the folder creation time only."

Developer: "Is an item exactly 60 days old stale?"

Domain expert: "Yes. The retention boundary is inclusive."

Developer: "Can shed reuse the same Archive bucket for another run later in the same month?"

Domain expert: "Yes. A later run adds to the existing Archive bucket."

Developer: "Do `C:\Users\pavel\Downloads` and `D:\Temp\Downloads` use different Archive buckets?"

Domain expert: "No. In the same month, both use the `Downloads` Archive bucket."

Developer: "Does the summary show how much the Archive will grow?"

Domain expert: "No. It shows the Move size: how much data will be removed from the selected folder."

Developer: "What does shed show when there are no stale items?"

Domain expert: "It prints Nothing to move and exits without opening the TUI. If there were Skipped items, their paths are also printed."

Developer: "What happens when the user presses `n` in the TUI?"

Domain expert: "The run becomes a Cancelled run: shed moves nothing and reports `Cancelled`."

Developer: "What if shed cannot read a root folder well enough to calculate its Move size?"

Domain expert: "That folder becomes a Skipped item and is not offered for archiving."

Developer: "If one item fails to move after confirmation, does shed undo earlier moves?"

Domain expert: "No. That item becomes a Failed move and successful moves stay archived."

Developer: "Does shed recheck each item's staleness after the user confirms?"

Domain expert: "No. Moving relies on the earlier scan result."

Developer: "Does the final Move summary list every successfully archived item?"

Domain expert: "No. It lists the total moved size, Archive bucket path, and Failed move paths only."

Developer: "If some moves fail, does the Move summary still use the planned Move size?"

Domain expert: "No. It reports only the size that was successfully moved."

Developer: "What if the Archive bucket path already exists as a file?"

Domain expert: "That is a Preflight failure, so shed starts no moves."

Developer: "Does shed move items in the filesystem scan order?"

Domain expert: "No. shed uses Move order: folders alphabetically first, then files and symlinks alphabetically."

Developer: "Does the TUI list show item paths or sizes?"

Domain expert: "No. It shows Display names only."

Developer: "Does `Report.pdf` conflict with `report.pdf` in the Archive bucket?"

Domain expert: "Yes. shed compares names case-insensitively like Windows."

Developer: "What happens if a stale file has the same name as a file already in the Archive bucket?"

Domain expert: "shed keeps both by applying a Numbered suffix to the incoming file."

Developer: "What happens if a stale folder has the same name as a folder already in the Archive bucket?"

Domain expert: "shed uses Merge and resolves conflicts inside the folder recursively."

Developer: "If a symlink points to a folder, does shed scan or merge that target folder?"

Domain expert: "No. A Symlink item is moved as the link itself."

Developer: "Does shed ignore `.env` just because the file name starts with a dot?"

Domain expert: "No. A dot-prefixed file is not hidden unless it has the Windows hidden attribute."

Developer: "Does shed ignore `.git` because the folder name starts with a dot?"

Domain expert: "Yes. A dot-prefixed folder is treated as a Hidden item."

Developer: "Can shed clean `~\Shed` or a folder inside it?"

Domain expert: "No. The Archive and its contents are never sources for shed."

Developer: "What happens if shed runs outside Windows?"

Domain expert: "shed reports Unsupported platform and does not scan."
