package core

import (
	"fmt"
	"path/filepath"
	"time"
)

func HeaderTitle(selectedFolder string) string {
	clean := filepath.Clean(selectedFolder)
	if filepath.Dir(clean) == clean {
		return clean
	}
	return filepath.Base(clean)
}

func CompactArchiveBucket(moveDate time.Time, selectedFolder string) string {
	return filepath.Join("~", "Shed", fmt.Sprintf("%04d", moveDate.Year()), fmt.Sprintf("%02d", int(moveDate.Month())), HeaderTitle(selectedFolder))
}
