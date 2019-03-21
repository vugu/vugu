package vugu

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserGo(t *testing.T) {

	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "TestParserGo")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	p := &ParserGo{
		PackageName:   "main",
		ComponentType: "DemoComp",
		TagName:       "demo-comp",
		DataType:      "*DemoCompData",
		OutDir:        tmpDir,
		OutFile:       "demo-component.go",
	}

	err = p.Parse(bytes.NewReader([]byte(`
<div id="whatever">
	<ul id="ul1" vg-if="data.ShowFirstUL">
		<li vg-range=".Test2" @click="something" :testbind="bound">Blah1</li>
		<li>Blah2</li>
	</ul>
	<ul id="ul2">
		<li class="li3" vg-for="_, item := range data.SecondULItems" vg-html="item"></li>
	</ul>
	<ul id="ul3">
		<!-- shorthand version -->
		<li class="li4" vg-for="data.SecondULItems" vg-html="value"></li>
	</ul>
</div>
`)))
	assert.NoError(err)

	b, err := ioutil.ReadFile(filepath.Join(tmpDir, "demo-component.go"))
	assert.NoError(err)

	t.Logf("OUT PROGRAM:\n%s", b)

	// test program
	assert.NoError(ioutil.WriteFile(filepath.Join(tmpDir, "demo-main.go"), []byte(`
package main 

import "github.com/vugu/vugu"

func main() {
	_ = &vugu.VGNode{}
	_ = &DemoComp{}
}

type DemoCompData struct {
	ShowFirstUL bool
	SecondULItems []string
}

`), 0644))

	// go.mod file that maps vugu to the source tree we're testing
	wd, err := os.Getwd()
	log.Printf("test working dir = %q", wd)
	assert.NoError(err)
	assert.True(filepath.IsAbs(wd))
	assert.NoError(ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`
module main
replace github.com/vugu/vugu => `+wd+`
`), 0644))
	// require (
	// golang.org/x/net v0.0.0-20190320064053-1272bf9dcd53
	// )

	cmd := exec.Command("go", "build", "-o", "a.exe", ".")
	cmd.Dir = tmpDir
	b, err = cmd.CombinedOutput()
	assert.NoError(err)
	log.Printf("BUILD OUTPUT: %s", b)

}

// func TestRandomParserStuff(t *testing.T) {

// 	assert := assert.New(t)

// 	var r = bytes.NewReader([]byte(`

// <div id="test1">
// 	Blah
// </div>

// <style>
// .my-funk {
// 	background: brown;
// }
// </style>

// <script>
// console.log("This is my funk!")
// </script>

// <script type="application/x-go">
// func test1() {
// 	log.Printf("This is my func!")
// }
// <script>

// `))

// 	nodeList, err := html.ParseFragment(r, cruftBody)
// 	assert.NoError(err)

// 	// should be only one node with type Element and that's what we want
// 	// var el *html.Node
// 	for _, n := range nodeList {

// 		if n.Type == html.ElementNode {
// 			// log.Printf("Node; %#v", n)
// 			if n.Data == "style" || n.Data == "script" {
// 				log.Printf("style first child: %#v", n.FirstChild)
// 				log.Printf("style first child next: %#v", n.FirstChild.NextSibling)
// 			}
// 		}

// 		// if n.Type == html.ElementNode {
// 		// 	if el != nil {
// 		// 		return fmt.Errorf("found more than one element at root of component template")
// 		// 	}
// 		// 	el = n
// 		// }

// 	}
// 	// if el == nil {
// 	// 	return fmt.Errorf("unable to find an element at root of component template")
// 	// }

// }
