package gen

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleParseGoPkgRun(t *testing.T) {
	assert := assert.New(t)

	tmpDir, err := os.MkdirTemp("", "TestParseGoPkgRun")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Dir: %s\n", tmpDir)
	//	defer os.RemoveAll(tmpDir)

	assert.NoError(os.WriteFile(filepath.Join(tmpDir, "root.vugu"), []byte(`
<div id="root_comp">
	<h1>Hello!</h1>
</div>
`), 0644))

	assert.NoError(os.WriteFile(filepath.Join(tmpDir, "root.go"), []byte(`package main

type Root struct {
}
`), 0644))

	assert.NoError(Generate(tmpDir))

	b, err := os.ReadFile(filepath.Join(tmpDir, "root_gen_js_wasm.go"))
	assert.NoError(err)

	if !bytes.Contains(b, []byte(`func (c *Root) Build`)) {
		t.Errorf("failed to find Build method signature")
	}
	if !t.Failed() {
		os.RemoveAll(tmpDir)
	}
}

func TestRun(t *testing.T) {
	debug := true

	pwd, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	type tcase struct {
		name    string
		infiles map[string]string   // file structure to start with
		out     map[string][]string // regexps to match in output files
		bfiles  map[string]string   // additional files to write before building
	}

	tcList := []tcase{
		{
			name: "simple",
			infiles: map[string]string{
				"root.vugu": `<div>root here</div>`,
				"root.go":   "package main\ntype Root struct {\n}\n",
				"go.mod":    "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":   "//go:build js && wasm\n\npackage main\nfunc main(){}",
			},
			out: map[string][]string{
				"root_gen_js_wasm.go": {`func \(c \*Root\) Build`},
			},
		},
		{
			name: "recursive",
			infiles: map[string]string{
				"root.vugu":            `<div>root here</div>`,
				"root.go":              "package main\ntype Root struct {\n}\n",
				"go.mod":               "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":              "//go:build js && wasm\n\npackage main\nfunc main(){}",
				"subdir1/example.vugu": "<div>Example Here</div>",
				"subdir1/example.go":   "package main\ntype Example struct {\n}\n",
			},
			out: map[string][]string{
				"root_gen_js_wasm.go":            {`func \(c \*Root\) Build`, `root here`},
				"subdir1/example_gen_js_wasm.go": {"Example Here"},
			},
		},
	}

	for _, tc := range tcList {
		t.Run(tc.name, func(t *testing.T) {

			tmpDir, err := os.MkdirTemp("", "TestRun")
			if err != nil {
				t.Fatal(err)
			}

			if debug {
				t.Logf("Test %q using tmpDir: %s", tc.name, tmpDir)
			} else {
				t.Parallel()
			}

			tstWriteFiles(tmpDir, tc.infiles)

			err = Generate(tmpDir)
			if err != nil {
				t.Fatal(err)
			}

			for fname, patterns := range tc.out {
				b, err := os.ReadFile(filepath.Join(tmpDir, fname))
				if err != nil {
					t.Errorf("failed to read file %q after Run: %v", fname, err)
					break
				}
				for _, pattern := range patterns {
					re := regexp.MustCompile(pattern)
					if !re.Match(b) {
						t.Errorf("failed to match regexp on file %q: %s", fname, pattern)
						break
					}
				}
			}

			tstWriteFiles(tmpDir, tc.bfiles)

			cmd := exec.Command("go", "mod", "tidy")
			cmd.Dir = tmpDir
			b, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go mod tidy error: %s; OUTPUT:\n%s", err, b)
			}

			cmd = exec.Command("go", "build", "-o", "main.wasm", ".")
			cmd.Dir = tmpDir
			cmd.Env = os.Environ() // needed?
			cmd.Env = append(cmd.Env, "GOOS=js", "GOARCH=wasm")
			b, err = cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("build error: %s; OUTPUT:\n%s", err, b)
			}

			// only if everything is golden do we remove
			if !t.Failed() {
				os.RemoveAll(tmpDir)
			}

		})
	}

}

func TestMissingComponentGoFile(t *testing.T) {
	debug := true

	pwd, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	type tcase struct {
		name    string
		infiles map[string]string   // file structure to start with
		out     map[string][]string // regexps to match in output files
		bfiles  map[string]string   // additional files to write before building
	}

	tcList := []tcase{
		{
			name: "missing-go-file",
			infiles: map[string]string{
				"root.vugu": `<div>root here</div>`,
				"go.mod":    "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":   "//go:build js && wasm\n\npackage main\nfunc main(){}",
			},
		},
		{
			name: "incorrectly-named-go-file",
			infiles: map[string]string{
				"root.vugu": `<div>root here</div>`,
				"roots.go":  "package main\ntype Root struct{}\n",
				"go.mod":    "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":   "//go:build js && wasm\n\npackage main\nfunc main(){}",
			},
		},
	}

	for _, tc := range tcList {
		t.Run(tc.name, func(t *testing.T) {

			tmpDir, err := os.MkdirTemp("", "TestMissingComponentGoFile")
			if err != nil {
				t.Fatal(err)
			}

			if debug {
				t.Logf("Test %q using tmpDir: %s", tc.name, tmpDir)
			} else {
				t.Parallel()
			}

			tstWriteFiles(tmpDir, tc.infiles)

			err = Generate(tmpDir)
			if !errors.Is(err, ErrNoComponentGoFile) {
				t.Fatalf("Expected an ErrNoComponentGoFile but got: %v\n", err)
			}
			// only if everything is golden do we remove
			if !t.Failed() {
				os.RemoveAll(tmpDir)
			}
		})
	}

}

func TestParserErrors(t *testing.T) {
	debug := true

	pwd, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	type tcase struct {
		name    string
		infiles map[string]string   // file structure to start with
		out     map[string][]string // regexps to match in output files
		bfiles  map[string]string   // additional files to write before building
	}

	tcList := []tcase{
		{
			name: "simple",
			infiles: map[string]string{
				"root.vugu": `<div>root here</div>`,
				"root.go":   "package\ntype Root struct {\n}\n", // this is invalid because there is no package name
				"go.mod":    "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":   "//go:build js && wasm\n\npackage main\nfunc main(){}",
			},
		},
	}

	for _, tc := range tcList {
		t.Run(tc.name, func(t *testing.T) {

			tmpDir, err := os.MkdirTemp("", "TestParserErrors")
			if err != nil {
				t.Fatal(err)
			}

			if debug {
				t.Logf("Test %q using tmpDir: %s", tc.name, tmpDir)
			} else {
				t.Parallel()
			}

			tstWriteFiles(tmpDir, tc.infiles)

			err = Generate(tmpDir)

			if !errors.Is(err, ErrCouldNotDeterminePackage) {
				t.Fatalf("Expected an ErrCouldNotDeterminePackage but got: %v\n", err)
			}
			// only if everything is golden do we remove
			if !t.Failed() {
				os.RemoveAll(tmpDir)
			}
		})
	}

}

func tstWriteFiles(dir string, m map[string]string) {

	for name, contents := range m {
		p := filepath.Join(dir, name)
		err := os.MkdirAll(filepath.Dir(p), 0755)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(p, []byte(contents), 0644)
		if err != nil {
			panic(err)
		}
	}
}
