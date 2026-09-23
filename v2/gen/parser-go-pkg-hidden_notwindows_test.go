//go:build!windows

package gen

import (
	"errors"
	"maps" // NOTE use of the maps package requires go 1.23 (The go.mod has been upgraded to go 1.25.0)
	"os"
	"path/filepath"
	"testing"
)

func TestParserHidden(t *testing.T) {
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
			name: "hidden_file",
			infiles: map[string]string{
				".root.vugu": `<div>root here</div>`, // make this a hidden file
				"root.go":    "package main\ntype Root struct {\n}\n",
				"go.mod":     "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":    "//go:build js && wasm\n\npackage main\nfunc main(){}",
			},
			out: map[string][]string{
				".root_gen_js_wasm.go": nil,
				"root_gen_js_wasm.go":  nil,
			},
		},
		{
			name: "hidden_file_in_subdir",
			infiles: map[string]string{
				"root.vugu":            `<div>root here</div>`,             // make this a hidden file
				"root.go":              "package\ntype Root struct {\n}\n", // this is invalid because there is no package name
				"go.mod":               "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":              "//go:build js && wasm\n\npackage main\nfunc main(){}",
				"subdir/.example.vugu": `<div>example here</div>`, // make this a hidden file
				"subdir/example.go":    "package main\ntype Example struct {\n}\n",
			},
			out: map[string][]string{
				"subdir/.example_gen_js_wasm.go": nil,
				"subdir/example_gen_js_wasm.go":  nil,
			},
		},
		{
			name: "hidden_directory",
			infiles: map[string]string{
				"root.vugu":            `<div>root here</div>`,             // make this a hidden file
				"root.go":              "package\ntype Root struct {\n}\n", // this is invalid because there is no package name
				"go.mod":               "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":              "//go:build js && wasm\n\npackage main\nfunc main(){}",
				".subdir/example.vugu": `<div>example here</div>`, // file in a hidden directory
				".subdir/example.go":   "package main\ntype Example struct {\n}\n",
			},
			out: map[string][]string{
				".subdir/example_gen_js_wasm.go": nil,
			},
		},
	}

	for _, tc := range tcList {
		t.Run(tc.name, func(t *testing.T) {

			tmpDir, err := os.MkdirTemp("", "TestParserHidden")
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

			for key := range maps.Keys(tc.out) {
				_, err = os.Stat(filepath.Join(tmpDir, key))
				if !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("Did not expect to generate: %q", filepath.Join(tmpDir, key))
				}
			}
			// only if everything is golden do we remove
			if !t.Failed() {
				os.RemoveAll(tmpDir)
			}
		})
	}

}
