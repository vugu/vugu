package distutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// DefaultFileInclPattern is a sensible default set of "static" files.
// This might be updated from time to time to include new types of assets used on the web.
// Extensions which are used for server executables, server configuration files, or files with
// an empty extension will not be added here.
var DefaultFileInclPattern = regexp.MustCompile(`[.](css|js|html|map|jpg|jpeg|png|gif|svg|eot|ttf|otf|woff|woff2|wasm)$`)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustCopyDirFiltered is like CopyDirFiltered but panics on error.
func MustCopyDirFiltered(srcDir, dstDir string, fileInclPattern *regexp.Regexp) {
	must(CopyDirFiltered(srcDir, dstDir, fileInclPattern))
}

// CopyDirFiltered recursively copies from srcDir to dstDir that match fileInclPattern.
// fileInclPattern is only checked against the base name of the file, not its directory.
// If fileInclPattern is nil, DefaultFileInclPattern will be used.
// The dstDir is skipped if encountered when recursing into srcDir.
// Directories are only created in the output dir if there's a file there.
// dstDir must already exist.  For individual file copies, CopyFile() is used, which means
// files with the same name, modification time and size are assumed to be up to date and
// the function returns immediately.  Conversely when the copy succeeds the modification
// time is set to that of the source.
func CopyDirFiltered(srcDir, dstDir string, fileInclPattern *regexp.Regexp) error {

	if fileInclPattern == nil {
		fileInclPattern = DefaultFileInclPattern
	}

	srcDir, err := filepath.Abs(srcDir)
	if err != nil {
		return err
	}
	dstDir, err = filepath.Abs(dstDir)
	if err != nil {
		return err
	}

	var copydir func(src, dst string) error
	copydir = func(src, dst string) error {

		src, err := filepath.Abs(src)
		if err != nil {
			return err
		}
		if src == dstDir { // makes it so dstDir can be inside srcDir without causing problems
			// fmt.Printf("skipping destination dir: %s\n", dstDir)
			return nil
		}
		dst, err = filepath.Abs(dst)
		if err != nil {
			return err
		}

		srcf, err := os.Open(src)
		if err != nil {
			return err
		}
		defer srcf.Close()

		srcFIs, err := srcf.Readdir(-1)
		if err != nil {
			return err
		}

		for _, srcFI := range srcFIs {

			// for directories we recurse...
			if srcFI.IsDir() {
				nextSrc := filepath.Join(src, srcFI.Name())
				nextDst := filepath.Join(dst, srcFI.Name())
				err := copydir(nextSrc, nextDst)
				if err != nil {
					return err
				}
				continue
			}

			// for files...

			// skip if it doesn't match the pattern
			if !fileInclPattern.MatchString(srcFI.Name()) {
				continue
			}

			// make sure the destination directory exists
			err = os.MkdirAll(dst, 0755)
			if err != nil {
				return err
			}

			srcFile := filepath.Join(src, srcFI.Name())
			dstFile := filepath.Join(dst, srcFI.Name())

			// copy the file
			err = CopyFile(srcFile, dstFile)
			if err != nil {
				return err
			}

		}

		return nil
	}
	return copydir(srcDir, dstDir)

}

// MustCopyFile is like CopyFile but panics on error.
func MustCopyFile(src, dst string) {
	must(CopyFile(src, dst))
}

// CopyFile copies src to dest. Will not copy directories.  Symlinks will have their contents copied.
// Files with the same name, modification time and size are assumed to be up to date and
// the function returns immediately.  Conversely when the copy succeeds the modification
// time is set to that of the source.
func CopyFile(src, dst string) error {

	src, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	dst, err = filepath.Abs(dst)
	if err != nil {
		return err
	}

	srcFI, err := os.Stat(src)
	if err != nil {
		return err
	}

	dstFI, err := os.Stat(dst)
	if err == nil {

		if dstFI.IsDir() {
			return fmt.Errorf("destination (%q) is directory, cannot CopyFile", dst)
		}

		// destination file exists, let's see if it looks like it's the same
		dstModTime := dstFI.ModTime().Truncate(time.Second)
		srcModTime := srcFI.ModTime().Truncate(time.Second)
		if dstModTime == srcModTime && srcFI.Size() == dstFI.Size() {
			return nil // looks like our work is already done, just return
		}
	}

	// open the file, create if it doesn't exist, truncate if it does
	dstF, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, srcFI.Mode())
	if err != nil {
		return err
	}
	defer dstF.Close()

	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	_, err = io.Copy(dstF, srcF)
	if err != nil {
		return err
	}

	// update destionation file's mod timestamp to match the source file
	err = os.Chtimes(dst, time.Now(), srcFI.ModTime())
	if err != nil {
		return err
	}

	return nil
}
