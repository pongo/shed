# shed

shed is a Windows-only console utility that moves stale root items from a selected folder into `~\Shed`.

It is meant for quick, reversible cleanup of busy folders such as Downloads. Before moving anything, shed shows an interactive confirmation screen with the items and total size.

## Usage

```powershell
shed [--age days=0] [folder]
```

Examples:

```powershell
shed
shed C:\Users\name\Downloads
shed --age 0 docs\plans
shed --age 30 .
shed . --age 30
shed --help
```

Arguments:

- `folder`: folder to scan. Defaults to the current working directory.
- `--age`: minimum item age in whole days. Defaults to `0`, which treats every eligible root item as stale. Positive values use the file modified date and folder creation date.

Flags may be placed before or after the folder.

## What It Moves

shed scans only direct children of the selected folder. Nested files are moved only when their root folder is moved.

It skips hidden items and items whose metadata or size cannot be read safely.

Moved items are placed under:

```text
~\Shed\<yyyy>\<MM>\<bucket-source-path>
```

If a file or symlink name already exists in the target Shed bucket, shed adds a numbered suffix. If a folder name already exists, shed merges the folder contents using the same conflict rules recursively.

## Shed Pruning

At startup, shed checks for Shed months older than six months. If any are found, it asks whether to send those months to the Recycle Bin before continuing with shedding.

## Build And Test

```powershell
go build ./cmd/shed
go test ./...
```
