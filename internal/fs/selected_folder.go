package fs

import (
	"os"
	"path/filepath"
	"strings"
)

type SelectedFolderResolver struct{}

func (SelectedFolderResolver) Resolve(arg string) (string, error) {
	selected := normalizeSelectedFolderArg(arg)
	if selected == "" || selected == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		selected = cwd
	}

	absolute, err := filepath.Abs(selected)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absolute)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", &NotDirectoryError{Path: absolute}
	}

	return absolute, nil
}

func normalizeSelectedFolderArg(arg string) string {
	return strings.Trim(arg, `"`)
}

type NotDirectoryError struct {
	Path string
}

func (err *NotDirectoryError) Error() string {
	return err.Path + " is not a folder"
}
