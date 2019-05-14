package vugu

import (
	// "io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	// "regexp"
	"testing"
	// "github.com/vugu/vugu/distutil"
)

// var defaultGoMod = `
// module example.org/someone/testapp
// `

// // default development server that launches all the wasm stuff
// var defaultDevServerGo = `// +build ignore

// package main

// import (
// 	"log"
// 	"net/http"
// 	"os"

// 	"github.com/vugu/vugu/simplehttp"
// )

// func main() {
// 	wd, _ := os.Getwd()
// 	l := "127.0.0.1:19944"
// 	log.Printf("Starting HTTP Server at %q", l)
// 	h := simplehttp.New(wd, true)
// 	// include a CSS file
// 	// simplehttp.DefaultStaticData["CSSFiles"] = []string{ "/my/file.css" }
// 	log.Fatal(http.ListenAndServe(l, h))
// }
// `

// // default testdrive.go which uses chromedp to drive the UI and verify output - exit status determines test pass/fail;
// // by default it finds a #run div and clicks it and waits to see if a #success div shows up
// var defaultTestDrive = `// +build ignore

// package main

// import (
// 	"context"
// 	"log"
// 	"time"

// 	"github.com/chromedp/chromedp"
// )

// func main() {

// 	// create chrome instance
// 	ctx, cancel := chromedp.NewContext(
// 		context.Background(),
// 		chromedp.WithLogf(log.Printf),
// 	)
// 	defer cancel()

// 	// create a timeout
// 	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
// 	defer cancel()

// 	// navigate to a page, wait for an element, click
// 	err := chromedp.Run(ctx,
// 		chromedp.Navigate("http://127.0.0.1:19944/"),
// 		// wait for a #run element to appear
// 		chromedp.WaitVisible("#run"),
// 		// click it
// 		chromedp.Click("#run"),
// 		// wait until a #success div shows up
// 		// chromedp.WaitVisible("#success"),
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	//log.Printf("GOT HERE")
// 	//time.Sleep(10 * time.Second)

// }
// `

// TestFull performs tests on various full example programs to ensure they build and run correctly.
func TestFull(t *testing.T) {

	mustRun(t, "full-test-data/test1")

}

func mustRun(t *testing.T, p string) {

	// dir, err := ioutil.TempDir("", "test-full")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// distutil.MustCopyDirFiltered(p, dir, regexp.MustCompile(`.*`))
	// log.Printf("Using temp dir: %s", dir)
	// defer os.RemoveAll(dir)

	// devServerGoPath := filepath.Join(dir, "devserver.go")
	// _, err = os.Stat(devServerGoPath)
	// if err != nil { // if error then write out default
	// 	err = ioutil.WriteFile(devServerGoPath, []byte(defaultDevServerGo), 0644)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// }

	// goModPath := filepath.Join(dir, "go.mod")
	// _, err = os.Stat(goModPath)
	// if err != nil {
	// 	err = ioutil.WriteFile(goModPath, []byte(defaultGoMod), 0644)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// }

	// testDriveGoPath := filepath.Join(dir, "testdrive.go")
	// _, err = os.Stat(testDriveGoPath)
	// if err != nil {
	// 	err = ioutil.WriteFile(testDriveGoPath, []byte(defaultTestDrive), 0644)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// }

	dir, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "build", "-o", "devserver", "devserver.go")
	cmd.Dir = dir
	b, err := cmd.CombinedOutput()
	log.Printf("GO BUILD OUTPUT: %s", b)
	if err != nil {
		t.Fatal(err)
	}

	// var outBuf bytes.Buffer
	devServerCmd := exec.Command("./devserver")
	devServerCmd.Dir = dir
	devServerCmd.Stdout = os.Stdout
	devServerCmd.Stderr = os.Stderr

	err = devServerCmd.Start()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := devServerCmd.Process.Kill()
		if err != nil {
			log.Printf("devserver kill err: %v", err)
		}
		os.Remove(filepath.Join(dir, "devserver")) // remove executable
	}()

	cmd = exec.Command("go", "run", "testdrive.go")
	cmd.Dir = dir
	b, err = cmd.CombinedOutput()
	log.Printf("TESTDRIVE OUTPUT: %s", b)
	if err != nil {
		t.Fatal(err)
	}

}
