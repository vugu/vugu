package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func PkgName(t *testing.T) string {
	t.Helper()
	// The magefile mounts the test's parent directory as the directory into the nginx container
	// So we need to know our package name, which is the last component of the directory path.
	// We then need to append the package name onto the end of the URL passed to chromedp
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Base(cwd)

}
