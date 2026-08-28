package gen

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

const rootDotVugu = `<div class="test-div" id="testdiv">

    <ul>
        <vg-type vg-pkg="main" vg-struct="DemoLine" vg-for="i := 0; i < c.ItemCount; i++" vg-key="i" vg-field="Num=i"></vg-type>
    </ul>

    <button id="addbtn" @click="c.OnAdd()">Add</button>

</div>

<style>
#test_div_id {
    background: #ddd;
}
</style>
`

const rootDotGo = `package main

type Root struct {
	ItemCount int ` + "`" + `vugu:"data"` + "`" + `
}

func (c *Root) BeforeBuild() {
	if c.ItemCount == 0 {
		c.ItemCount = 3
	}
}

func (c *Root) OnAdd() {
	c.ItemCount++
}
`

const demoLineDotGo = `package main

type DemoLine struct {
	Num int ` + "`" + `vugu:"data"` + "`" + `
}
`

const demoLineDotVugu = `<div id="demoline"><li class="demo-line"><strong vg-html="c.Num"></strong> a line is here</li></div>

<style>
.demo-line {
    color: green;
}
</style>
`

func TestComponentGen(t *testing.T) {
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
			name: "component",
			infiles: map[string]string{
				"root.vugu":      rootDotVugu,
				"root.go":        rootDotGo,
				"demo-line.vugu": demoLineDotVugu,
				"demo-line.go":   demoLineDotGo,
				"go.mod":         "module testcase\nreplace github.com/vugu/vugu/v2 => " + pwd + "\n",
				"main.go":        "//go:build js && wasm\n\npackage main\nfunc main(){}",
			},
			out: map[string][]string{
				"root_gen_js_wasm.go":      {`func \(c \*Root\) Build`},
				"demo-line_gen_js_wasm.go": {`func \(c \*DemoLine\) Build`},
			},
		},
	}

	for _, tc := range tcList {
		t.Run(tc.name, func(t *testing.T) {

			// tmpDir, err := os.MkdirTemp("", "TestRun")
			// if err != nil {
			// 	t.Fatal(err)
			// }
			tmpDir := "/tmp/component-test"
			err := os.RemoveAll(tmpDir)
			if err != nil {
				t.Fatal(err)
			}
			err = os.Mkdir(tmpDir, 0755)
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
			// if !t.Failed() {
			// 	os.RemoveAll(tmpDir)
			// }

		})
	}

}
