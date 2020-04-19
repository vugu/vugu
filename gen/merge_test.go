package gen

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestMerge(t *testing.T) {

	debug := true

	type tcase struct {
		name    string
		infiles map[string]string   // file structure to start with
		out     map[string][]string // regexps to match in output files
		outNot  map[string][]string // regexps to NOT match in output files
	}

	tcList := []tcase{
		{
			name: "simple",
			infiles: map[string]string{
				"file1.go": "package main\nfunc main(){}",
				"file2.go": "package main\nvar a string",
			},
			out: map[string][]string{
				"out.go": {`func main`, `var a string`},
			},
		},
		{
			name: "comments",
			infiles: map[string]string{
				"file1.go": "package main\n// main comment here\nfunc main(){}",
				"file2.go": "package main\nvar a string // a comment here\n",
			},
			out: map[string][]string{
				"out.go": {`func main`, `// main comment here`, `var a string`, `// a comment here`},
			},
		},
		{
			name: "import-dedup",
			infiles: map[string]string{
				"file1.go": "package main\nimport \"fmt\"\n// main comment here\nfunc main(){}",
				"file2.go": "package main\nimport \"fmt\"\nvar a string // a comment here\n",
			},
			out: map[string][]string{
				"out.go": {`import "fmt"`},
			},
			outNot: map[string][]string{
				"out.go": {`(?ms)import "fmt".*import "fmt"`},
			},
		},
		{
			name: "import-dedup-2",
			infiles: map[string]string{
				"file1.go": "package main\nimport \"fmt\"\n// main comment here\nfunc main(){}",
				"file2.go": "package main\nimport \"fmt\"\nimport \"log\"\nvar a string // a comment here\n",
			},
			out: map[string][]string{
				"out.go": {`import "fmt"`, `import "log"`},
			},
			outNot: map[string][]string{
				"out.go": {`(?ms)\}.*import "log"`},
			},
		},
	}

	for _, tc := range tcList {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			tmpDir, err := ioutil.TempDir("", "TestMerge")
			if err != nil {
				t.Fatal(err)
			}

			if debug {
				t.Logf("Test %q using tmpDir: %s", tc.name, tmpDir)
			} else {
				defer os.RemoveAll(tmpDir)
				t.Parallel()
			}

			tstWriteFiles(tmpDir, tc.infiles)
			var in []string
			for k := range tc.infiles {
				// in = append(in, filepath.Join(tmpDir, k))
				in = append(in, k)
			}

			err = mergeGoFiles(tmpDir, "out.go", in...)
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

			for fname, patterns := range tc.outNot {
				b, err := ioutil.ReadFile(filepath.Join(tmpDir, fname))
				if err != nil {
					t.Errorf("failed to read file %q after Run: %v", fname, err)
					continue
				}
				for _, pattern := range patterns {
					re := regexp.MustCompile(pattern)
					if re.Match(b) {
						t.Errorf("incorrectly matched regexp on file %q: %s", fname, pattern)
					}
				}
			}

			if debug {
				outb, _ := ioutil.ReadFile(filepath.Join(tmpDir, "out.go"))
				t.Logf("OUTPUT:\n%s", outb)
			}

		})
	}

}
