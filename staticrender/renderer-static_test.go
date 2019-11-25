package staticrender

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vugu/vugu/gen"
)

func TestRendererStatic(t *testing.T) {

	cachekiller := 0
	_ = cachekiller

	// make a temp dir

	tmpDir, err := ioutil.TempDir("", "TestRendererStatic")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("tmpDir: %s", tmpDir)
	// defer os.RemoveAll(tmpDir)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	vuguwd, err := filepath.Abs(filepath.Join(wd, ".."))
	if err != nil {
		t.Fatal(err)
	}

	// put a go.mod here that points back to the local copy of vugu
	err = ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(fmt.Sprintf(`module test-render-static
replace github.com/vugu/vugu => %s
require github.com/vugu/vugu v0.0.0-00010101000000-000000000000
`, vuguwd)), 0644)

	// output some components

	err = ioutil.WriteFile(filepath.Join(tmpDir, "root.vugu"), []byte(`<html>
<head>
<title>testing!</title>
<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css"/>
</head>
<body>
<div>
	This is a test!
	Component here:
	<main:Comp1/>
</div>
</body>
</html>`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(tmpDir, "comp1.vugu"), []byte(`<span>
comp1 in the house
<div vg-html='"<p>Some <strong>nested</strong> craziness</p>"'></div>
</span>`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// run the vugu codegen

	p := gen.NewParserGoPkg(tmpDir, nil)
	err = p.Run()
	if err != nil {
		t.Fatal(err)
	}

	// put our static output generation code here

	err = ioutil.WriteFile(filepath.Join(tmpDir, "staticout.go"), []byte(`// +build !wasm

package main

import (
	"log"
	//"fmt"
	"flag"
	"os"

	"github.com/vugu/vugu"
	"github.com/vugu/vugu/staticrender"
)

func main() {

	//mountPoint := flag.String("mount-point", "#vugu_mount_point", "The query selector for the mount point for the root component, if it is not a full HTML component")
	flag.Parse()

	//fmt.Printf("Entering main(), -mount-point=%q\n", *mountPoint)
	//defer fmt.Printf("Exiting main()\n")

	rootBuilder := &Root{}

	buildEnv, err := vugu.NewBuildEnv()
	if err != nil {
		log.Fatal(err)
	}

	renderer := staticrender.NewStaticRenderer(os.Stdout)

	buildResults := buildEnv.RunBuild(rootBuilder)

	err = renderer.Render(buildResults)
	if err != nil {
		panic(err)
	}
	
}
	`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// build it
	cmd := exec.Command("go", "build", "-v", "-o", "staticout")
	cmd.Dir = tmpDir
	b, err := cmd.CombinedOutput()
	log.Printf("go build produced:\n%s", b)
	if err != nil {
		t.Fatal(err)
	}

	// run it and see what it output

	cmd = exec.Command("./staticout")
	cmd.Dir = tmpDir
	b, err = cmd.CombinedOutput()
	log.Printf("staticout produced:\n%s", b)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(b), "<div><p>Some <strong>nested</strong> craziness</p></div>") {
		t.Errorf("falied to find target string in output")
	}

}

// func TestRendererStatic(t *testing.T) {

// 	in := "\xEF\xBB\xBF" + `   <html> <body> <div iD="whatever">
// 	</Div> </body></html>`

// 	offset := 0

// 	z := html.NewTokenizer(bytes.NewReader([]byte(in)))
// 	for {
// 		tt := z.Next()
// 		if tt == html.ErrorToken {
// 			log.Fatal(z.Err())
// 		}

// 		tokenLen := len(z.Raw())

// 		log.Printf("raw: %q (offset=%d, inpart=%q)", z.Raw(), offset, in[offset:offset+tokenLen])
// 		offset += tokenLen
// 	}

// }
