//go:build windows

package fs

import (
	"syscall"
	"time"
	"unsafe"
)

const fileAttributeHidden = 0x2

type metadata struct {
	Created time.Time
	Hidden  bool
}

func readMetadata(path string) (metadata, error) {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return metadata{}, err
	}

	var data syscall.Win32FileAttributeData
	if err := syscall.GetFileAttributesEx(pointer, syscall.GetFileExInfoStandard, (*byte)(unsafe.Pointer(&data))); err != nil {
		return metadata{}, err
	}

	return metadata{
		Created: time.Unix(0, data.CreationTime.Nanoseconds()),
		Hidden:  data.FileAttributes&fileAttributeHidden != 0,
	}, nil
}
