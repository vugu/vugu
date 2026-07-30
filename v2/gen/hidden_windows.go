//go:build windows

package gen

import (
	"path/filepath"
	"syscall"
)

func isHidden(filename string) (bool, error) {
	// strip any path so we just look at the base filename (including any extension)
	base := filepath.Base(filename)
	pointer, err := syscall.UTF16PtrFromString(base)
	if err != nil {
		return false, err
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}
	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
