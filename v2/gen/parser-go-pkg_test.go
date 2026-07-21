package gen

import (
	"bytes"
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
	defer os.RemoveAll(tmpDir)

	assert.NoError(os.WriteFile(filepath.Join(tmpDir, "root.vugu"), []byte(`
<div id="root_comp">
	<h1>Hello!</h1>
</div>
`), 0644))

	assert.NoError(os.WriteFile(filepath.Join(tmpDir, "root.go"), []byte(`
package main

type Root struct {
}
`), 0644))

	p := NewParserGoPkg(tmpDir)

	assert.NoError(p.Run())

	b, err := os.ReadFile(filepath.Join(tmpDir, "root_gen_js_wasm.go"))
	assert.NoError(err)

	if !bytes.Contains(b, []byte(`func (c *Root) Build`)) {
		t.Errorf("failed to find Build method signature")
	}
}

func TestRun(t *testing.T) {
	debug := true

	pwd, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	type tcase struct {
		name      string
		recursive bool
		infiles   map[string]string              // file structure to start with
		out       map[string][]string            // regexps to match in output files
		afterRun  func(dir string, t *testing.T) // called after Run
		bfiles    map[string]string              // additional files to write before building
		build     string                         // "wasm", "default", "none"
	}

	tcList := []tcase{
		{
			name:      "simple",
			recursive: false,
			infiles: map[string]string{
				"root.vugu": `<div>root here</div>`,
				"root.go":   "package main\ntype Root struct {\n}\n",
				"go.mod":    "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":   "//go:build js && wasm\n\npackage main\nfunc main(){}",
			},
			out: map[string][]string{
				"root_gen_js_wasm.go": {`func \(c \*Root\) Build`},
			},
			build: "wasm",
		},
		{
			name:      "simple-wasm",
			recursive: false,
			infiles: map[string]string{
				"root.vugu": `<div>root here</div>`,
				"root.go":   "package main\ntype Root struct {\n}\n",
				"go.mod":    "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":   "//go:build js && wasm\n\npackage main\nfunc main(){}",
			},
			out: map[string][]string{
				"root_gen_js_wasm.go": {`func \(c \*Root\) Build`},
			},
			build: "wasm",
		},
		{
			name:      "recursive",
			recursive: true,
			infiles: map[string]string{
				"root.vugu":            `<div>root here</div>`,
				"root.go":              "package main\ntype Root struct {\n}\n",
				"go.mod":               "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":              "//go:build js && wasm\n\npackage main\nfunc main(){}",
				"subdir1/example.vugu": "<div>Example Here</div>",
			},
			out: map[string][]string{
				"root_gen_js_wasm.go":            {`func \(c \*Root\) Build`, `root here`},
				"subdir1/example_gen_js_wasm.go": {"Example Here"},
			},
			build: "wasm",
		},
		{
			name:      "recursive-single",
			recursive: true,
			infiles: map[string]string{
				"root.vugu":            `<div>root here</div>`,
				"root.go":              "package main\ntype Root struct {\n}\n",
				"go.mod":               "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":              "//go:build js && wasm\n\npackage main\nfunc main(){}",
				"subdir1/example.vugu": "<div>Example Here</div>",
				"subdir1/example.go":   "package main\ntype Example struct {\n}\n",
			},
			out: map[string][]string{
				"root_gen_js_wasm.go":            {`func \(c \*Root\) Build`},
				"subdir1/example_gen_js_wasm.go": {`func \(c \*Example\) Build`, "Example Here"},
				"root.vugu":                      {`root here`}, // make sure vugu files didn't get nuked
				"subdir1/example.vugu":           {`Example Here`},
			},
			build: "wasm",
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

			if tc.recursive {
				err = RunRecursive(tmpDir)
			} else {
				err = Run(tmpDir)
			}
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

			if tc.afterRun != nil {
				tc.afterRun(tmpDir, t)
			}

			tstWriteFiles(tmpDir, tc.bfiles)

			cmd := exec.Command("go", "mod", "tidy")
			cmd.Dir = tmpDir
			b, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go mod tidy error: %s; OUTPUT:\n%s", err, b)
			}

			switch tc.build {

			case "wasm":
				cmd := exec.Command("go", "build", "-o", "main.wasm", ".")
				cmd.Dir = tmpDir
				cmd.Env = os.Environ() // needed?
				cmd.Env = append(cmd.Env, "GOOS=js", "GOARCH=wasm")
				b, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("build error: %s; OUTPUT:\n%s", err, b)
				}

			case "none":

			default:
				t.Errorf("unknown build value %q", tc.build)
			}

			// only if everything is golden do we remove
			if !t.Failed() {
				os.RemoveAll(tmpDir)
			}

		})
	}

}

func noFile(p string, t *testing.T) {
	_, err := os.Stat(p)
	if err == nil {
		t.Errorf("file %q should not exist but does", p)
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
