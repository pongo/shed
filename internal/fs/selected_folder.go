package fs

import (
	"os"
	"path/filepath"
)

type SelectedFolderResolver struct{}

func (SelectedFolderResolver) Resolve(arg string) (string, error) {
	if arg == "" || arg == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		arg = cwd
	}

	absolute, err := filepath.Abs(arg)
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

type NotDirectoryError struct {
	Path string
}

func (err *NotDirectoryError) Error() string {
	return err.Path + " is not a folder"
}
