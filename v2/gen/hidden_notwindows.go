//go:build !windows

package gen

import "path/filepath"

func isHidden(filename string) (bool, error) {
	// strip any path so we just look at the base filename (including any extension)
	base := filepath.Base(filename)
	return base[0] == '.', nil
}
