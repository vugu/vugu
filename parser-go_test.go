package vugu

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// parserGoBuildAndRunMulti will build and run multiple single-file components in the same package (requiring a main that prints to stdout) and return the captured output.
// pgmMap is the component struct name as the key and the program source as the value.
func parserGoBuildAndRunMulti(pgmMap map[string]string, debug bool) (string, error) {

	tmpDir, err := ioutil.TempDir("", "parserGoBuildAndRun")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	keys := make([]string, 0, len(pgmMap))
	for k := range pgmMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// log.Printf("keys = %#v", keys)

	for _, k := range keys {

		p := &ParserGo{
			PackageName:   "main",
			ComponentType: k,
			DataType:      k + "Data",
			OutDir:        tmpDir,
			OutFile:       k + ".go",
		}

		err = p.Parse(bytes.NewReader([]byte(pgmMap[k])))
		if err != nil {
			return "", fmt.Errorf("error parsing for %q: %v", k, err)
		}

		if debug {
			b, err := ioutil.ReadFile(filepath.Join(tmpDir, k+".go"))
			if err != nil {
				return "", err
			}
			log.Printf("OUT PROGRAM (%s.go):\n%s", k, b)
		}

	}

	wd, err := os.Getwd()
	// log.Printf("test working dir = %q", wd)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(wd) {
		panic(fmt.Errorf("wd is not absolute: %s", wd))
	}
	err = ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`
module main
replace github.com/vugu/vugu => `+wd+`
`), 0644)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("go", "build", "-o", "a.exe", ".")
	cmd.Dir = tmpDir
	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("BUILD OUTPUT: %s", b)
		return "", err
	}
	if debug {
		log.Printf("BUILD OUTPUT: %s", b)
	}

	cmd = exec.Command("./a.exe")
	cmd.Dir = tmpDir
	b, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("RUN OUTPUT: %s", b)
		return "", err
	}
	if debug {
		log.Printf("RUN OUTPUT: %s", b)
	}

	return string(b), nil
}

// parserGoBuildAndRun will build an run a single-file component (requiring a main that prints to stdout) and return the captured output
func parserGoBuildAndRun(pgm string, debug bool) (string, error) {

	tmpDir, err := ioutil.TempDir("", "parserGoBuildAndRun")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	p := &ParserGo{
		PackageName:   "main",
		ComponentType: "DemoComp",
		// TagName:       "demo-comp",
		DataType: "DemoCompData",
		OutDir:   tmpDir,
		OutFile:  "demo-component.go",
	}

	err = p.Parse(bytes.NewReader([]byte(pgm)))
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadFile(filepath.Join(tmpDir, "demo-component.go"))
	if err != nil {
		log.Printf("OUT PROGRAM:\n%s", b)
		return "", err
	}
	if debug {
		log.Printf("OUT PROGRAM:\n%s", b)
	}

	wd, err := os.Getwd()
	// log.Printf("test working dir = %q", wd)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(wd) {
		panic(fmt.Errorf("wd is not absolute: %s", wd))
	}
	err = ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`
module main
replace github.com/vugu/vugu => `+wd+`
`), 0644)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("go", "build", "-o", "a.exe", ".")
	cmd.Dir = tmpDir
	b, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("BUILD OUTPUT: %s", b)
		return "", err
	}
	if debug {
		log.Printf("BUILD OUTPUT: %s", b)
	}

	cmd = exec.Command("./a.exe")
	cmd.Dir = tmpDir
	b, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("RUN OUTPUT: %s", b)
		return "", err
	}
	if debug {
		log.Printf("RUN OUTPUT: %s", b)
	}

	return string(b), nil
}

func TestParserGo(t *testing.T) {

	assert := assert.New(t)

	out, err := parserGoBuildAndRun(`
<div id="whatever">
	<ul id="ul1" vg-if="data.ShowFirstUL">
		<li vg-range=".Test2" :testbind="data.TestBound">Blah1</li>
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

<script type="application/x-go">

func main() {
	_ = &vugu.VGNode{}
	_ = &DemoComp{}
	fmt.Println("OK")
}

type DemoComp struct {
}

type DemoCompData struct {
	ShowFirstUL bool
	SecondULItems []string
	TestBound bool
}

func (ct *DemoComp) NewData(props vugu.Props) (interface{}, error) {
	return &DemoCompData{
		ShowFirstUL: true,
		SecondULItems: []string{"a","b","c"},
	}, nil
}

</script>
`, false)
	assert.NoError(err)
	assert.Equal("OK", strings.TrimSpace(out))

}
