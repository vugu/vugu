package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/vugu/vugu/vugufmt"
)

var (
	exitCode    = 0
	list        = flag.Bool("l", false, "list files whose formatting differs from vugufmt's")
	write       = flag.Bool("w", false, "write result to (source) file instead of stdout")
	simplifyAST = flag.Bool("s", false, "simplify code")
	imports     = flag.Bool("i", false, "run goimports instead of gofmt")
	doDiff      = flag.Bool("d", false, "display diffs instead of rewriting files")
)

func main() {
	vugufmtMain()
	os.Exit(exitCode)
}

func vugufmtMain() {
	// Handle input flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: vugufmt [flags] [path ...]\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// If no file paths given, we are reading from stdin.
	if flag.NArg() == 0 {
		if err := processFile("<standard input>", os.Stdin, os.Stdout); err != nil {
			report(err)
		}
		return
	}

	// Otherwise, we need to read a bunch of files
	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			report(err)
		case dir.IsDir():
			walkDir(path)
		default:
			if err := processFile(path, nil, os.Stdout); err != nil {
				report(err)
			}
		}
	}
}

func walkDir(path string) {
	filepath.Walk(path, visitFile)
}

func visitFile(path string, f os.FileInfo, err error) error {
	if err == nil && isVuguFile(f) {
		err = processFile(path, nil, os.Stdout)
	}

	// Don't complain if a file was deleted in the meantime (i.e.
	// the directory changed concurrently while running gofmt).
	if err != nil && !os.IsNotExist(err) {
		report(err)
	}
	return nil
}

func isVuguFile(f os.FileInfo) bool {
	// ignore non-Vugu files (except html)
	name := f.Name()
	return !f.IsDir() &&
		!strings.HasPrefix(name, ".") &&
		(strings.HasSuffix(name, ".vugu") || (strings.HasSuffix(name, ".html")))
}

func report(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(err.Error()))
	exitCode = 2
}

func processFile(filename string, in io.Reader, out io.Writer) error {
	var perm os.FileMode = 0644
	// open the file if needed
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return err
		}
		in = f
		perm = fi.Mode().Perm()
	}

	src, err := io.ReadAll(in)
	if err != nil {
		return err
	}

	var resBuff bytes.Buffer

	var formatter *vugufmt.Formatter
	if *imports {
		formatter = vugufmt.NewFormatter(vugufmt.UseGoImports)
	} else {
		formatter = vugufmt.NewFormatter(vugufmt.UseGoFmt(*simplifyAST))
	}

	if !*list && !*doDiff {
		if err := formatter.FormatHTML(filename, bytes.NewReader(src), &resBuff); err != nil {
			return err
		}
		res := resBuff.Bytes()

		if *write {
			// make a temporary backup before overwriting original
			bakname, err := backupFile(filename+".", src, perm)
			if err != nil {
				return err
			}
			err = os.WriteFile(filename, res, perm)
			if err != nil {
				os.Rename(bakname, filename)
				return err
			}
			err = os.Remove(bakname)
			if err != nil {
				return err
			}
		} else {
			// just write to stdout
			_, err = out.Write(res)
		}
	} else {
		different, err := formatter.Diff(filename, bytes.NewReader(src), &resBuff)
		if err != nil {
			return fmt.Errorf("computing diff: %s", err)
		}
		if *list {
			if different {
				fmt.Fprintln(out, filename)
			}
		} else if *doDiff {
			out.Write(resBuff.Bytes())
		}
	}

	return nil
}

const chmodSupported = runtime.GOOS != "windows"

// backupFile writes data to a new file named filename<number> with permissions perm,
// with <number randomly chosen such that the file name is unique. backupFile returns
// the chosen file name.
func backupFile(filename string, data []byte, perm os.FileMode) (string, error) {

	// create backup file
	f, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename))
	if err != nil {
		return "", err
	}

	bakname := f.Name()

	if chmodSupported {
		err = f.Chmod(perm)
		if err != nil {
			f.Close()
			os.Remove(bakname)
			return bakname, err
		}
	}

	// write data to backup file
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return bakname, err
}
