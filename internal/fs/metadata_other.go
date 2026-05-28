//go:build !windows

package fs

import (
	"fmt"
	"time"
)

type metadata struct {
	Created time.Time
	Hidden  bool
}

func readMetadata(string) (metadata, error) {
	return metadata{}, fmt.Errorf("Unsupported platform")
}
