package gen

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMissingFixer(t *testing.T) {

	// NOTE: for more complex testing, see TestRun which is easier to add more general cases to.

	tmpDir, err := ioutil.TempDir("", "TestMissingFixer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	must := func(err error) {
		if err != nil {
			_, file, line, _ := runtime.Caller(1)
			t.Fatalf("from %s:%d: %v", file, line, err)
		}
	}
	vuguAbs, _ := filepath.Abs("..")

	must(ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module missingfixertest\n\nreplace github.com/vugu/vugu => "+vuguAbs+"\n"), 0644))
	must(ioutil.WriteFile(filepath.Join(tmpDir, "events.go"), []byte("package main\n\n//vugugen:event Something\n//vugugen:event SomeOtherThing\n//vugugen:event SomeOtherThing\n"), 0644))
	must(ioutil.WriteFile(filepath.Join(tmpDir, "root.vugu"), []byte("<div>root</div>"), 0644))
	must(ioutil.WriteFile(filepath.Join(tmpDir, "root_vgen.go"), []byte("package main\n\nimport \"github.com/vugu/vugu\"\n\nfunc (c *Root)Build(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {return nil}"), 0644))
	// a second component that does include it's own struct definition
	must(ioutil.WriteFile(filepath.Join(tmpDir, "comp1.vugu"), []byte("<div>comp1</div>"), 0644))
	must(ioutil.WriteFile(filepath.Join(tmpDir, "comp1_vgen.go"), []byte("package main\n\nimport \"github.com/vugu/vugu\"\n\ntype Comp1 struct{}\n\nfunc (c *Comp1)Build(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {return nil}"), 0644))
	// a file with an event where the event type is declared but not the handler interface or func
	must(ioutil.WriteFile(filepath.Join(tmpDir, "epart.go"), []byte("package main\n\n//vugugen:event Part\ntype PartEvent struct { A string }\n"), 0644))

	// 	// TEMP
	// 	must(ioutil.WriteFile(filepath.Join(tmpDir, "root_vgen.go"), []byte(`
	// package main

	// import "github.com/vugu/vugu"

	// type Root struct {}

	// func (c *Root)Build(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {return nil}

	// `), 0644))

	mf := newMissingFixer(tmpDir, "main", map[string]string{
		"root.vugu": "root_vgen.go",
		// "comp1.vugu": "comp1_vgen.go",
	})
	err = mf.run()
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(filepath.Join(tmpDir, "0_missing_vgen.go"))
	must(err)
	t.Logf("0_missing_vgen.go result:\n%s", b)
	s := string(b)
	checks := []string{
		"type Root struct",

		"!type Comp1 struct",

		"type SomethingEvent struct",
		"type SomethingHandler interface",
		"type SomethingFunc func",
		"func (f SomethingFunc) SomethingHandle(",
		"var _ SomethingHandler =",

		// "type PartEvent struct", // should exist only in epart.go

		// "type SomeOtherThingEvent interface",
		// "type SomeOtherThingHandler interface",
		// "type SomeOtherThingHandlerFunc func",
		// "func (f SomeOtherThingHandlerFunc) SomeOtherThingHandle(",
		// "var _ SomeOtherThingHandler =",
	}
	for _, check := range checks {
		if check[0] == '!' {
			if strings.Contains(s, check[1:]) {
				t.Errorf("found unexpected %q", check[1:])
			}
		} else if !strings.Contains(s, check) {
			t.Errorf("missing %q", check)
		}
	}

	if t.Failed() {
		return
	}
	// if the above worked, try compiling it
	must(ioutil.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n\nfunc main(){}\n"), 0644))

	cmd := exec.Command("go", "build", "-o", "a.out", ".")
	cmd.Dir = tmpDir
	b, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("build output: %s", b)
		t.Fatal(err)
	}

	// ensure the output is there
	_, err = os.Stat(filepath.Join(tmpDir, "a.out"))
	if err != nil {
		t.Fatal(err)
	}

}
