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

// NOTE: a bunch of this borrowed from: https://github.com/otiai10/copy/blob/master/copy.go

// const (
// 	// tmpPermissionForDirectory makes the destination directory writable,
// 	// so that stuff can be copied recursively even if any original directory is NOT writable.
// 	// See https://github.com/otiai10/copy/pull/9 for more information.
// 	tmpPermissionForDirectory = os.FileMode(0755)
// )

// // CopyFile copies src to dest. Handles symlinks but not directories.
// // Files with the same name, modification time and size are assumed to be up to date and
// // the function returns immediately.  Conversely when the copy succeeds the modification
// // time is set to that of the source.
// func CopyFile(src, dest string) error {
// 	info, err := os.Lstat(src)
// 	if err != nil {
// 		return err
// 	}
// 	return copy(src, dest, info)
// }

// // copy dispatches copy-funcs according to the mode.
// // Because this "copy" could be called recursively,
// // "info" MUST be given here, NOT nil.
// func copy(src, dest string, info os.FileInfo) error {
// 	if info.Mode()&os.ModeSymlink != 0 {
// 		return lcopy(src, dest, info)
// 	}
// 	if info.IsDir() {
// 		return dcopy(src, dest, info)
// 	}
// 	return fcopy(src, dest, info)
// }

// func copy(src, dest string, info os.FileInfo) error {
// 	if info.Mode()&os.ModeSymlink != 0 {
// 		return fmt.Errorf("cannot copy symlink, only file")
// 		// return lcopy(src, dest, info)
// 	}
// 	if info.IsDir() {
// 		return fmt.Errorf("cannot copy directory, only file")
// 	}
// 	return fcopy(src, dest, info)
// }

// MustCopyFile is like CopyFile but panics on error.
func MustCopyFile(src, dst string) {
	must(CopyFile(src, dst))
}

// CopyFile copies src to dest. Will not copy directories.  Symlinks will have thier contents copied.
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

// // fcopy is for just a file,
// // with considering existence of parent directory
// // and file permission.
// func fcopy(src, dest string, info os.FileInfo) error {

// 	// if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
// 	// 	return err
// 	// }

// 	f, err := os.Create(dest)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
// 		return err
// 	}

// 	s, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer s.Close()

// 	_, err = io.Copy(f, s)
// 	return err
// }

// // dcopy is for a directory,
// // with scanning contents inside the directory
// // and pass everything to "copy" recursively.
// func dcopy(srcdir, destdir string, info os.FileInfo) error {

// 	originalMode := info.Mode()

// 	// Make dest dir with 0755 so that everything writable.
// 	if err := os.MkdirAll(destdir, tmpPermissionForDirectory); err != nil {
// 		return err
// 	}
// 	// Recover dir mode with original one.
// 	defer os.Chmod(destdir, originalMode)

// 	contents, err := ioutil.ReadDir(srcdir)
// 	if err != nil {
// 		return err
// 	}

// 	for _, content := range contents {
// 		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())
// 		if err := copy(cs, cd, content); err != nil {
// 			// If any error, exit immediately
// 			return err
// 		}
// 	}

// 	return nil
// }

// // lcopy is for a symlink,
// // with just creating a new symlink by replicating src symlink.
// func lcopy(src, dest string, info os.FileInfo) error {
// 	src, err := os.Readlink(src)
// 	if err != nil {
// 		return err
// 	}
// 	return os.Symlink(src, dest)
// }
