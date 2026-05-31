package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"shed/internal/core"
)

type Scanner struct {
	ShedRoot         string
	RetentionAgeDays *int
	Now              func() time.Time
}

func NewScanner(shedRoot string) Scanner {
	return Scanner{
		ShedRoot: shedRoot,
		Now:      time.Now,
	}
}

func (scanner Scanner) Scan(ctx context.Context, selectedFolder string) (core.ScanResult, error) {
	if scanner.Now == nil {
		scanner.Now = time.Now
	}
	if scanner.ShedRoot != "" && core.IsShedSource(selectedFolder, scanner.ShedRoot) {
		return core.ScanResult{}, fmt.Errorf("selected folder is a Shed source")
	}

	entries, err := os.ReadDir(selectedFolder)
	if err != nil {
		return core.ScanResult{}, err
	}

	items := make([]core.RootItem, 0, len(entries))
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return core.ScanResult{}, err
		}

		path := filepath.Join(selectedFolder, entry.Name())
		item, err := scanner.rootItem(path, entry.Name())
		if err != nil {
			items = append(items, core.RootItem{
				Name:    entry.Name(),
				Path:    path,
				Kind:    core.FileItem,
				SizeErr: err,
			})
			continue
		}
		items = append(items, item)
	}

	retentionAgeDays := core.DefaultRetentionAgeDays
	if scanner.RetentionAgeDays != nil {
		retentionAgeDays = *scanner.RetentionAgeDays
	}

	return core.ScanRootItemsWithRetentionAge(items, scanner.Now(), retentionAgeDays), nil
}

func (scanner Scanner) rootItem(path, name string) (core.RootItem, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return core.RootItem{}, err
	}

	kind := core.FileItem
	if info.Mode()&os.ModeSymlink != 0 {
		kind = core.SymlinkItem
	} else if info.IsDir() {
		kind = core.FolderItem
	}

	metadata, err := readMetadata(path)
	if err != nil {
		return core.RootItem{}, err
	}

	item := core.RootItem{
		Name:     name,
		Path:     path,
		Kind:     kind,
		Modified: info.ModTime(),
		Created:  metadata.Created,
		Hidden:   metadata.Hidden,
	}
	if !core.Eligible(item) {
		return item, nil
	}

	switch kind {
	case core.FolderItem:
		item.Size, item.SizeErr = RecursiveSize(path)
	case core.SymlinkItem:
		item.Size = 0
	default:
		item.Size = info.Size()
	}

	return item, nil
}

func ShedRootFromHome(homeDir string) string {
	return filepath.Join(homeDir, "Shed")
}

func UserShedRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return ShedRootFromHome(home), nil
}
