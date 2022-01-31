package staticrender

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/vugu/vugu/gen"
)

func TestRendererStaticTable(t *testing.T) {

	debug := false

	vuguDir, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	type tcase struct {
		name          string
		opts          gen.ParserGoPkgOpts
		recursive     bool
		infiles       map[string]string              // file structure to start with
		outReMatch    []string                       // regexps that must match against output
		outReNotMatch []string                       // regexps that must not match against output
		afterRun      func(dir string, t *testing.T) // called after Run
		bfiles        map[string]string              // additional files to write before building
	}

	tcList := []tcase{
		{
			name:      "simple",
			opts:      gen.ParserGoPkgOpts{},
			recursive: false,
			infiles: map[string]string{
				"root.vugu": `<div>root here</div>`,
			},
			outReMatch:    []string{`root here`},
			outReNotMatch: []string{`should not match`},
		},
		{
			name:      "full-html",
			opts:      gen.ParserGoPkgOpts{},
			recursive: false,
			infiles: map[string]string{
				"root.vugu": `<html><title vg-if='true'>test title</title><body><div>root here</div></body></html><script src="/a.js"></script>`,
			},
			outReMatch: []string{
				`root here`,
				`<title>test title</title>`, // if statement should have fired
				`</div><script src="/a.js"></script></body>`, // js should have be written directly inside the body tag
			},
			outReNotMatch: []string{`should not match`},
		},
		{
			name:      "comp",
			opts:      gen.ParserGoPkgOpts{},
			recursive: false,
			infiles: map[string]string{
				"root.vugu": `<html>
<head>
<title>testing!</title>
<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css"/>
<script>
console.log("Some script here");
</script>
</head>
<body>
<div>
	This is a test!
	Component here:
	<main:Comp1/>
</div>
</body>
</html>`,
				"comp1.vugu": `<span>
comp1 in the house
<div vg-content='vugu.HTML("<p>Some <strong>nested</strong> craziness</p>")'></div>
</span>`,
			},
			outReMatch: []string{
				`<div><p>Some <strong>nested</strong> craziness</p></div>`,
				`bootstrap.min.css`,
				`Some script here`,
			},
			outReNotMatch: []string{`should not match`},
		},
		{
			name:      "vg-template",
			opts:      gen.ParserGoPkgOpts{},
			recursive: false,
			infiles: map[string]string{
				"root.vugu": `<div><span>example1</span><vg-template vg-if='true'>text here</vg-template></div>`,
			},
			outReMatch: []string{
				`<span>example1</span>text here`,
			},
			outReNotMatch: []string{`vg-template`},
		},
	}

	for _, tc := range tcList {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			tmpDir, err := ioutil.TempDir("", "TestRendererStaticTable")
			if err != nil {
				t.Fatal(err)
			}

			if debug {
				t.Logf("Test %q using tmpDir: %s", tc.name, tmpDir)
			} else {
				t.Parallel()
			}

			// write a sensible go.mod and main.go, individual tests can override if they really want
			startf := make(map[string]string, 2)
			startf["go.mod"] = "module testcase\nreplace github.com/vugu/vugu => " + vuguDir + "\n"
			startf["main.go"] = `// +build !wasm

package main

import (
	"os"

	"github.com/vugu/vugu"
	"github.com/vugu/vugu/staticrender"
)

func main() {

	rootBuilder := &Root{}

	buildEnv, err := vugu.NewBuildEnv()
	if err != nil { panic(err) }

	renderer := staticrender.New(os.Stdout)

	buildResults := buildEnv.RunBuild(rootBuilder)

	err = renderer.Render(buildResults)
	if err != nil { panic(err) }
	
}
`
			tstWriteFiles(tmpDir, startf)

			tstWriteFiles(tmpDir, tc.infiles)

			tc.opts.SkipGoMod = true
			tc.opts.SkipMainGo = true
			if tc.recursive {
				err = gen.RunRecursive(tmpDir, &tc.opts)
			} else {
				err = gen.Run(tmpDir, &tc.opts)
			}
			if err != nil {
				t.Fatal(err)
			}

			if tc.afterRun != nil {
				tc.afterRun(tmpDir, t)
			}

			tstWriteFiles(tmpDir, tc.bfiles)

			// build executable for this platform
			cmd := exec.Command("go", "mod", "tidy")
			cmd.Dir = tmpDir
			b, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("build error: %s; OUTPUT:\n%s", err, b)
			}
			cmd = exec.Command("go", "build", "-o", "main.out", ".")
			cmd.Dir = tmpDir
			b, err = cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("build error: %s; OUTPUT:\n%s", err, b)
			}

			// now execute the command and capture the output
			cmd = exec.Command(filepath.Join(tmpDir, "main.out"))
			cmd.Dir = tmpDir
			b, err = cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("run error: %s; OUTPUT:\n%s", err, b)
			}

			// verify the output
			for _, reTxt := range tc.outReMatch {
				re := regexp.MustCompile(reTxt)
				if !re.Match(b) {
					t.Errorf("Failed to match regexp %q on output", reTxt)
				}
			}
			for _, reTxt := range tc.outReNotMatch {
				re := regexp.MustCompile(reTxt)
				if re.Match(b) {
					t.Errorf("Unexpected match for regexp %q on output", reTxt)
				}
			}

			// only if everthing is golden do we remove
			if !t.Failed() {
				os.RemoveAll(tmpDir)
			} else {
				// and if not then dump the output that was produced
				t.Logf("FULL OUTPUT:\n%s", b)
			}

		})
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

// NOTE: this was moved into the table test above
// func TestRendererStatic(t *testing.T) {

// 	cachekiller := 0
// 	_ = cachekiller

// 	// make a temp dir

// 	tmpDir, err := ioutil.TempDir("", "TestRendererStatic")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	log.Printf("tmpDir: %s", tmpDir)
// 	// defer os.RemoveAll(tmpDir)

// 	wd, err := os.Getwd()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	vuguwd, err := filepath.Abs(filepath.Join(wd, ".."))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// put a go.mod here that points back to the local copy of vugu
// 	err = ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(fmt.Sprintf(`module test-render-static
// replace github.com/vugu/vugu => %s
// require github.com/vugu/vugu v0.0.0-00010101000000-000000000000
// `, vuguwd)), 0644)

// 	// output some components

// 	err = ioutil.WriteFile(filepath.Join(tmpDir, "root.vugu"), []byte(`<html>
// <head>
// <title>testing!</title>
// <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css"/>
// <script>
// console.log("Some script here");
// </script>
// </head>
// <body>
// <div>
// 	This is a test!
// 	Component here:
// 	<main:Comp1/>
// </div>
// </body>
// </html>`), 0644)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	err = ioutil.WriteFile(filepath.Join(tmpDir, "comp1.vugu"), []byte(`<span>
// comp1 in the house
// <div vg-content='vugu.HTML("<p>Some <strong>nested</strong> craziness</p>")'></div>
// </span>`), 0644)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// run the vugu codegen

// 	p := gen.NewParserGoPkg(tmpDir, nil)
// 	err = p.Run()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// put our static output generation code here

// 	err = ioutil.WriteFile(filepath.Join(tmpDir, "staticout.go"), []byte(`// +build !wasm

// package main

// import (
// 	"log"
// 	//"fmt"
// 	"flag"
// 	"os"

// 	"github.com/vugu/vugu"
// 	"github.com/vugu/vugu/staticrender"
// )

// func main() {

// 	//mountPoint := flag.String("mount-point", "#vugu_mount_point", "The query selector for the mount point for the root component, if it is not a full HTML component")
// 	flag.Parse()

// 	//fmt.Printf("Entering main(), -mount-point=%q\n", *mountPoint)
// 	//defer fmt.Printf("Exiting main()\n")

// 	rootBuilder := &Root{}

// 	buildEnv, err := vugu.NewBuildEnv()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	renderer := staticrender.New(os.Stdout)

// 	buildResults := buildEnv.RunBuild(rootBuilder)

// 	err = renderer.Render(buildResults)
// 	if err != nil {
// 		panic(err)
// 	}

// }
// 	`), 0644)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// build it
// 	cmd := exec.Command("go", "build", "-v", "-o", "staticout")
// 	cmd.Dir = tmpDir
// 	b, err := cmd.CombinedOutput()
// 	log.Printf("go build produced:\n%s", b)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// run it and see what it output

// 	cmd = exec.Command("./staticout")
// 	cmd.Dir = tmpDir
// 	b, err = cmd.CombinedOutput()
// 	log.Printf("staticout produced:\n%s", b)

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if !strings.Contains(string(b), "<div><p>Some <strong>nested</strong> craziness</p></div>") {
// 		t.Errorf("falied to find target string in output")
// 	}
// 	if !strings.Contains(string(b), "bootstrap.min.css") {
// 		t.Errorf("falied to find target string in output")
// 	}
// 	if !strings.Contains(string(b), "Some script here") {
// 		t.Errorf("falied to find target string in output")
// 	}

// }
