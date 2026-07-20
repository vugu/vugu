//go:build !windows

package gen

func isHidden(filename string) (bool, error) {
	return filename[0] == '.', nil
}
