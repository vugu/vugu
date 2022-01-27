package gen

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleParseGoPkgRun(t *testing.T) {

	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "TestParseGoPkgRun")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 	assert.NoError(ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`
	// module main
	// `), 0644))

	assert.NoError(ioutil.WriteFile(filepath.Join(tmpDir, "root.vugu"), []byte(`
<div id="root_comp">
	<h1>Hello!</h1>
</div>
`), 0644))

	p := NewParserGoPkg(tmpDir, nil)

	assert.NoError(p.Run())

	b, err := ioutil.ReadFile(filepath.Join(tmpDir, "root_vgen.go"))
	assert.NoError(err)
	// t.Logf("OUT FILE root_vgen.go: %s", b)
	// log.Printf("OUT FILE root_vgen.go: %s", b)

	if !bytes.Contains(b, []byte(`func (c *Root) Build`)) {
		t.Errorf("failed to find Build method signature")
	}

	b, err = ioutil.ReadFile(filepath.Join(tmpDir, "0_missing_vgen.go"))
	assert.NoError(err)

	if !bytes.Contains(b, []byte(`type Root struct`)) {
		t.Errorf("failed to find Root struct definition")
	}

}

func TestRun(t *testing.T) {

	debug := false

	pwd, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	type tcase struct {
		name      string
		opts      ParserGoPkgOpts
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
			opts:      ParserGoPkgOpts{},
			recursive: false,
			infiles: map[string]string{
				"root.vugu": `<div>root here</div>`,
				"go.mod":    "module testcase\nreplace github.com/vugu/vugu => " + pwd + "\n",
				"main.go":   "package main\nfunc main(){}",
			},
			out: map[string][]string{
				"root_vgen.go":      {`func \(c \*Root\) Build`},
				"0_missing_vgen.go": {`type Root struct`},
			},
			build: "default",
		},
		{
			name:      "simple-wasm",
			opts:      ParserGoPkgOpts{},
			recursive: false,
			infiles: map[string]string{
				"root.vugu": `<div>root here</div>`,
				"go.mod":    "module testcase\nreplace github.com/vugu/vugu => " + pwd + "\n",
			},
			out: map[string][]string{
				"root_vgen.go":      {`func \(c \*Root\) Build`},
				"0_missing_vgen.go": {`type Root struct`},
			},
			build: "wasm",
		},
		{
			name:      "recursive",
			opts:      ParserGoPkgOpts{},
			recursive: true,
			infiles: map[string]string{
				"root.vugu":            `<div>root here</div>`,
				"go.mod":               "module testcase\nreplace github.com/vugu/vugu => " + pwd + "\n",
				"main.go":              "package main\nfunc main(){}",
				"subdir1/example.vugu": "<div>Example Here</div>",
			},
			out: map[string][]string{
				"root_vgen.go":            {`func \(c \*Root\) Build`, `root here`},
				"0_missing_vgen.go":       {`type Root struct`},
				"subdir1/example_vgen.go": {`Example Here`},
			},
			build: "default",
		},
		{
			name:      "recursive-single",
			opts:      ParserGoPkgOpts{MergeSingle: true},
			recursive: true,
			infiles: map[string]string{
				"root.vugu":            `<div>root here</div>`,
				"go.mod":               "module testcase\nreplace github.com/vugu/vugu => " + pwd + "\n",
				"main.go":              "package main\nfunc main(){}",
				"subdir1/example.vugu": "<div>Example Here</div>",
			},
			out: map[string][]string{
				"0_components_vgen.go":         {`func \(c \*Root\) Build`, `type Root struct`},
				"subdir1/0_components_vgen.go": {`Example Here`},
				"root.vugu":                    {`root here`}, // make sure vugu files didn't get nuked
				"subdir1/example.vugu":         {`Example Here`},
			},
			afterRun: func(dir string, t *testing.T) {
				noFile(filepath.Join(dir, "subdir1/example_vgen.go"), t)
			},
			build: "default",
		},
		{
			name:      "events",
			opts:      ParserGoPkgOpts{},
			recursive: false,
			infiles: map[string]string{
				"root.vugu": `<div>root here</div>`,
				"go.mod":    "module testcase\nreplace github.com/vugu/vugu => " + pwd + "\n",
				"main.go":   "package main\nfunc main(){}\n\n//vugugen:event Sample\n",
			},
			out: map[string][]string{
				"root_vgen.go":      {`func \(c \*Root\) Build`},
				"0_missing_vgen.go": {`type Root struct`, `SampleEvent`, `SampleHandler`, `SampleFunc`},
			},
			build: "default",
		},
	}

	for _, tc := range tcList {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			tmpDir, err := ioutil.TempDir("", "TestRun")
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
				err = RunRecursive(tmpDir, &tc.opts)
			} else {
				err = Run(tmpDir, &tc.opts)
			}
			if err != nil {
				t.Fatal(err)
			}

			for fname, patterns := range tc.out {
				b, err := ioutil.ReadFile(filepath.Join(tmpDir, fname))
				if err != nil {
					t.Errorf("failed to read file %q after Run: %v", fname, err)
					continue
				}
				for _, pattern := range patterns {
					re := regexp.MustCompile(pattern)
					if !re.Match(b) {
						t.Errorf("failed to match regexp on file %q: %s", fname, pattern)
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

			case "default":
				cmd := exec.Command("go", "build", "-o", "main.out", ".")
				cmd.Dir = tmpDir
				b, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("build error: %s; OUTPUT:\n%s", err, b)
				}

				cmd = exec.Command(filepath.Join(tmpDir, "main.out"))
				cmd.Dir = tmpDir
				b, err = cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("run error: %s; OUTPUT:\n%s", err, b)
				}

			case "none":

			default:
				t.Errorf("unknown build value %q", tc.build)
			}

			// only if everthing is golden do we remove
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
		os.MkdirAll(filepath.Dir(p), 0755)
		err := ioutil.WriteFile(p, []byte(contents), 0644)
		if err != nil {
			panic(err)
		}
	}

}
