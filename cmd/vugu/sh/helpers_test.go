package sh_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/magefile/mage/sh"
)

// compareFiles checks that two files are identical for testing purposes. That means they have the same length,
// the same contents, and the same permissions. It does NOT mean they have the same timestamp, as that is expected
// to change in normal Mage sh.Copy operation.
func compareFiles(file1, file2 string) error {
	s1, err := os.Stat(file1)
	if err != nil {
		return fmt.Errorf("can't stat %s: %w", file1, err)
	}
	s2, err := os.Stat(file2)
	if err != nil {
		return fmt.Errorf("can't stat %s: %w", file2, err)
	}
	if s1.Size() != s2.Size() {
		return fmt.Errorf("files %s and %s have different sizes: %d vs %d", file1, file2, s1.Size(), s2.Size())
	}
	if s1.Mode() != s2.Mode() {
		return fmt.Errorf("files %s and %s have different permissions: %#4o vs %#4o", file1, file2, s1.Mode(), s2.Mode())
	}
	f1bytes, err := os.ReadFile(file1)
	if err != nil {
		return fmt.Errorf("can't read %s: %w", file1, err)
	}
	f2bytes, err := os.ReadFile(file2)
	if err != nil {
		return fmt.Errorf("can't read %s: %w", file2, err)
	}
	if !bytes.Equal(f1bytes, f2bytes) {
		return fmt.Errorf("files %s and %s have different contents", file1, file2)
	}
	return nil
}

func TestHelpers(t *testing.T) {
	mytmpdir := t.TempDir()
	defer func() {
		derr := os.RemoveAll(mytmpdir)
		if derr != nil {
			fmt.Printf("error cleaning up after TestHelpers: %v", derr)
		}
	}()
	srcname := filepath.Join(mytmpdir, "test1.txt")
	err := os.WriteFile(srcname, []byte("All work and no play makes Jack a dull boy."), 0o600)
	if err != nil {
		t.Fatalf("can't create test file %s: %v", srcname, err)
	}
	destname := filepath.Join(mytmpdir, "test2.txt")

	t.Run("sh/copy", func(t *testing.T) {
		cerr := sh.Copy(destname, srcname)
		if cerr != nil {
			t.Errorf("test file copy from %s to %s failed: %v", srcname, destname, cerr)
		}
		cerr = compareFiles(srcname, destname)
		if cerr != nil {
			t.Errorf("test file copy verification failed: %v", cerr)
		}
	})

	// While we've got a temporary directory, test how forgiving sh.Rm is
	t.Run("sh/rm/ne", func(t *testing.T) {
		nef := filepath.Join(mytmpdir, "file_not_exist.txt")
		rerr := sh.Rm(nef)
		if rerr != nil {
			t.Errorf("sh.Rm complained when removing nonexistent file %s: %v", nef, rerr)
		}
	})

	t.Run("sh/copy/ne", func(t *testing.T) {
		nef := filepath.Join(mytmpdir, "file_not_exist.txt")
		nedf := filepath.Join(mytmpdir, "file_not_exist2.txt")
		cerr := sh.Copy(nedf, nef)
		if cerr == nil {
			t.Errorf("sh.Copy succeeded copying nonexistent file %s", nef)
		}
	})

	// We test sh.Rm by clearing up our own test files and directories
	t.Run("sh/rm", func(t *testing.T) {
		rerr := sh.Rm(destname)
		if rerr != nil {
			t.Errorf("failed to remove file %s: %v", destname, rerr)
		}
		rerr = sh.Rm(srcname)
		if rerr != nil {
			t.Errorf("failed to remove file %s: %v", srcname, rerr)
		}
		rerr = sh.Rm(mytmpdir)
		if rerr != nil {
			t.Errorf("failed to remove dir %s: %v", mytmpdir, rerr)
		}
		_, rerr = os.Stat(mytmpdir)
		if rerr == nil {
			t.Errorf("removed dir %s but it's still there?", mytmpdir)
		}
	})

	t.Run("sh/rm/nedir", func(t *testing.T) {
		rerr := sh.Rm(mytmpdir)
		if rerr != nil {
			t.Errorf("sh.Rm complained removing nonexistent dir %s", mytmpdir)
		}
	})
}

func TestCopyNonExistentSource(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "dst.txt")
	err := sh.Copy(dst, filepath.Join(dir, "nonexistent.txt"))
	if err == nil {
		t.Fatal("expected error copying from non-existent source")
	}
}

func TestCopyReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	if err := os.WriteFile(src, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Try to copy to a path inside a non-existent subdirectory
	dst := filepath.Join(dir, "nodir", "dst.txt")
	err := sh.Copy(dst, src)
	if err == nil {
		t.Fatal("expected error copying to non-existent directory")
	}
}

func TestCopyPreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	if err := os.WriteFile(src, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "dst.txt")
	if err := sh.Copy(dst, src); err != nil {
		t.Fatal(err)
	}
	srcInfo, _ := os.Stat(src)
	dstInfo, _ := os.Stat(dst)
	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("permissions differ: src=%v dst=%v", srcInfo.Mode(), dstInfo.Mode())
	}
}

func TestCopyOverwrite(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("new content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := sh.Copy(dst, src); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new content" {
		t.Fatalf("expected 'new content', got %q", string(data))
	}
}
