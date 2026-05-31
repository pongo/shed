package core

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func HeaderTitle(selectedFolder string) string {
	clean := filepath.Clean(selectedFolder)
	if filepath.Dir(clean) == clean {
		return clean
	}
	return filepath.Base(clean)
}

func BucketSourcePath(invocationFolder, selectedFolder string) string {
	selectedSource := HeaderTitle(selectedFolder)
	if invocationFolder == "" {
		return selectedSource
	}

	invocation := filepath.Clean(invocationFolder)
	if filepath.Dir(invocation) == invocation {
		return selectedSource
	}

	relative, err := filepath.Rel(cleanForCompare(invocation), cleanForCompare(selectedFolder))
	if err != nil {
		return selectedSource
	}
	if relative == "." {
		return filepath.Base(invocation)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return selectedSource
	}

	sourceRelative, err := filepath.Rel(invocation, filepath.Clean(selectedFolder))
	if err != nil {
		sourceRelative = relative
	}
	return filepath.Join(filepath.Base(invocation), sourceRelative)
}

func CompactShedBucket(moveDate time.Time, invocationFolder, selectedFolder string) string {
	return filepath.Join("~", "Shed", fmt.Sprintf("%04d", moveDate.Year()), fmt.Sprintf("%02d", int(moveDate.Month())), BucketSourcePath(invocationFolder, selectedFolder))
}

func ShedBucket(shedRoot string, moveDate time.Time, invocationFolder, selectedFolder string) string {
	return filepath.Join(shedRoot, fmt.Sprintf("%04d", moveDate.Year()), fmt.Sprintf("%02d", int(moveDate.Month())), BucketSourcePath(invocationFolder, selectedFolder))
}
