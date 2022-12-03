package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"

	"github.com/vugu/vugu/devutil"
	"github.com/vugu/vugu/distutil"
	"github.com/vugu/vugu/gen"
	"github.com/vugu/vugu/simplehttp"
)

func queryNode(ref string, assert func(n *cdp.Node)) chromedp.QueryAction {
	return chromedp.QueryAfter(ref, func(ctx context.Context, nodes ...*cdp.Node) error {
		if len(nodes) == 0 {
			return fmt.Errorf("no %s element found", ref)
		}
		assert(nodes[0])
		return nil
	})
}

func queryAttributes(ref string, assert func(attributes map[string]string)) chromedp.QueryAction {
	return chromedp.QueryAfter(ref, func(ctx context.Context, nodes ...*cdp.Node) error {
		attributes := make(map[string]string)
		if err := chromedp.Attributes(ref, &attributes).Do(ctx); err != nil {
			return err
		}
		assert(attributes)
		return nil
	})
}

// WaitInnerTextTrimEq will wait for the innerText of the specified element to match a specific string after whitespace trimming.
func WaitInnerTextTrimEq(sel, innerText string) chromedp.QueryAction {

	return chromedp.Query(sel, func(s *chromedp.Selector) {

		chromedp.WaitFunc(func(ctx context.Context, cur *cdp.Frame, ids ...cdp.NodeID) ([]*cdp.Node, error) {

			nodes := make([]*cdp.Node, len(ids))
			cur.RLock()
			for i, id := range ids {
				nodes[i] = cur.Nodes[id]
				if nodes[i] == nil {
					cur.RUnlock()
					// not yet ready
					return nil, nil
				}
			}
			cur.RUnlock()

			var ret string
			err := chromedp.EvaluateAsDevTools("document.querySelector('"+sel+"').innerText", &ret).Do(ctx)
			if err != nil {
				return nodes, err
			}
			if strings.TrimSpace(ret) != innerText {
				// log.Printf("found text: %s", ret)
				return nodes, errors.New("unexpected value: " + ret)
			}

			// log.Printf("NodeValue: %#v", nodes[0])

			// return nil, errors.New("not ready yet")
			return nodes, nil
		})(s)

	})

}

// returns absdir
func mustUseDir(reldir string) (newdir, olddir string) {

	odir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	olddir = odir

	dir, err := filepath.Abs(reldir)
	if err != nil {
		panic(err)
	}

	must(os.Chdir(dir))

	newdir = dir

	return
}

func mustGen(absdir string) {

	os.Remove(filepath.Join(absdir, "main_wasm.go")) // ensure it gets re-generated
	pp := gen.NewParserGoPkg(absdir, nil)
	err := pp.Run()
	if err != nil {
		panic(err)
	}

}

func mustTGGen(absdir string) {

	os.Remove(filepath.Join(absdir, "main_wasm.go")) // ensure it gets re-generated
	pp := gen.NewParserGoPkg(absdir, &gen.ParserGoPkgOpts{TinyGo: true})
	err := pp.Run()
	if err != nil {
		panic(err)
	}

}

func mustGenBuildAndLoad(absdir string) string {
	mustGen(absdir)
	return mustBuildAndLoad(absdir)
}

// returns path suffix
func mustBuildAndLoad(absdir string) string {

	fmt.Print(distutil.MustEnvExec([]string{"GOOS=js", "GOARCH=wasm"}, "go", "mod", "tidy"))
	fmt.Print(distutil.MustEnvExec([]string{"GOOS=js", "GOARCH=wasm"}, "go", "build", "-o", filepath.Join(absdir, "main.wasm"), "."))

	mustWriteSupportFiles(absdir, true)

	uploadPath := mustUploadDir(absdir, "http://localhost:8846/upload")
	// log.Printf("uploadPath = %q", uploadPath)

	return uploadPath
}

// // like mustBuildAndLoad but with tinygo
// func mustBuildAndLoadTinygo(absdir string) string {

// 	fmt.Print(distutil.MustEnvExec([]string{"GOOS=js", "GOARCH=wasm"}, "go", "build", "-o", filepath.Join(absdir, "main.wasm"), "."))

// 	mustWriteSupportFiles(absdir)

// 	uploadPath := mustUploadDir(absdir, "http://localhost:8846/upload")
// 	// log.Printf("uploadPath = %q", uploadPath)

// 	return uploadPath
// }

func mustChromeCtx() (context.Context, context.CancelFunc) {

	debugURL := func() string {
		resp, err := http.Get("http://localhost:9222/json/version")
		if err != nil {
			panic(err)
		}

		var result map[string]interface{}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			panic(err)
		}
		return result["webSocketDebuggerUrl"].(string)
	}()

	// t.Log(debugURL)

	allocCtx, _ := chromedp.NewRemoteAllocator(context.Background(), debugURL)
	// defer cancel()

	ctx, _ := chromedp.NewContext(allocCtx) // , chromedp.WithLogf(log.Printf))
	// defer cancel()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	// defer cancel()

	return ctx, cancel
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustCleanDir(dir string) {
	must(os.Chdir(dir))
	b, err := ioutil.ReadFile(".gitignore")
	if err != nil {
		panic(err)
	}
	ss := strings.Split(string(b), "\n")
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" || s == "." || s == ".." || strings.HasPrefix(s, "#") {
			continue
		}
		// log.Printf("removing: %s", s)
		os.Remove(s)
	}

}

// mustWriteSupportFiles will write index.html and wasm_exec.js to a directory
func mustWriteSupportFiles(dir string, doWasmExec bool) {
	if doWasmExec {
		distutil.MustCopyFile(distutil.MustWasmExecJsPath(), filepath.Join(dir, "wasm_exec.js"))
	}
	// distutil.MustCopyFile(distutil.(), filepath.Join(dir, "wasm_exec.js"))
	// log.Println(simplehttp.DefaultPageTemplateSource)

	// "/wasm_exec.js"

	var buf bytes.Buffer

	req, _ := http.NewRequest("GET", "/index.html", nil)
	outf, err := os.OpenFile(filepath.Join(dir, "index.html"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	distutil.Must(err)
	defer outf.Close()
	template.Must(template.New("_page_").Parse(simplehttp.DefaultPageTemplateSource)).Execute(&buf, map[string]interface{}{"Request": req})
	// HACK: fix wasm_exec.js path, unti we can come up with a better way to do this
	outf.Write(
		bytes.Replace(
			bytes.Replace(buf.Bytes(), []byte(`"/wasm_exec.js"`), []byte(`"wasm_exec.js"`), 1),
			[]byte("/main.wasm"), []byte("main.wasm"), 1,
		),
	)

}

// mustUploadDir tar+gz's the given directory and posts that file to the specified endpoint,
// returning the path of where to access the files
func mustUploadDir(dir, endpoint string) string {

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	absDir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}

	// var hdr tar.Header
	err = filepath.Walk(dir, filepath.WalkFunc(func(fpath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		absPath, err := filepath.Abs(fpath)
		if err != nil {
			panic(err)
		}

		relPath := path.Clean("/" + strings.TrimPrefix(absPath, absDir))

		// log.Printf("path = %q, fi.Name = %q", path, fi.Name())
		hdr, err := tar.FileInfoHeader(fi, "")
		hdr.Name = relPath
		// hdr = tar.Header{
		// 	Name: fi.Name(),
		// 	Mode: 0644,
		// 	Size: fi.Size(),
		// }
		err = tw.WriteHeader(hdr)
		if err != nil {
			return err
		}

		// no body to write for directories
		if fi.IsDir() {
			return nil
		}

		inf, err := os.Open(fpath)
		if err != nil {
			return err
		}
		defer inf.Close()
		_, err = io.Copy(tw, inf)

		if err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		panic(err)
	}

	tw.Close()
	gw.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("archive", filepath.Base(dir)+".tar.gz")
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(part, &buf)
	if err != nil {
		panic(err)
	}
	err = writer.Close()
	if err != nil {
		panic(err)
	}

	// for key, val := range params {
	// 	_ = writer.WriteField(key, val)
	// }
	// err = writer.Close()
	// if err != nil {
	// 	return nil, err
	// }

	req, err := http.NewRequest("POST", endpoint, &body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	var retData struct {
		Path string `json:"path"`
	}
	err = json.NewDecoder(res.Body).Decode(&retData)
	if err != nil {
		panic(err)
	}
	return retData.Path
}

// mustTGTempGopathSetup makes a temp dir and recursively copies from testPjtDir into
// filepath.Join(tmpDir, outRelPath) and returns the temp dir, which can be used
// as the GOPATH for a tinygo build
func mustTGTempGopathSetup(testPjtDir, outRelPath string) string {
	// buildGopath := mustTGTempGopathSetup(dir, "src/main")

	tmpParent, err := filepath.Abs(filepath.Join(testPjtDir, "../tmp"))
	if err != nil {
		panic(err)
	}

	name, err := ioutil.TempDir(tmpParent, "tggopath")
	if err != nil {
		panic(err)
	}

	log.Printf("testPjtdir=%s, name=%s", testPjtDir, name)

	// copy vugu package files

	// HACK: for now we use specific files names in order to avoid recursive stuff getting out of control - we should figure out something better
	srcDir := filepath.Join(testPjtDir, "../..")
	dstDir := filepath.Join(name, "src/github.com/vugu/vugu")
	must(os.MkdirAll(dstDir, 0755))
	fis, err := ioutil.ReadDir(srcDir)
	must(err)
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		distutil.MustCopyFile(filepath.Join(srcDir, fi.Name()), filepath.Join(dstDir, fi.Name()))
	}

	// for _, n := range []string{
	// 	"build-env.go",
	// 	"change-counter.go",
	// 	"change-counter_test.go",
	// 	"comp-key.go",
	// 	"comp-key_test.go",
	// 	"component.go",
	// 	"doc.go",
	// 	"events-component.go",
	// 	"events-dom.go",
	// 	"mod-check-common.go",
	// 	"mod-check-default.go",
	// 	"mod-check-tinygo.go",
	// 	"mod-check_test.go",
	// 	"vgnode.go",
	// 	"vgnode_test.go",
	// } {
	// 	distutil.MustCopyFile(filepath.Join(srcDir, n), filepath.Join(dstDir, n))
	// }

	allPattern := regexp.MustCompile(`.*`)
	for _, n := range []string{"domrender", "internal", "js"} {
		distutil.MustCopyDirFiltered(
			filepath.Join(srcDir, n),
			filepath.Join(name, "src/github.com/vugu/vugu", n),
			allPattern)
	}

	// now finally copy the actual test program
	distutil.MustCopyDirFiltered(
		testPjtDir,
		filepath.Join(name, "src/tgtestpgm"),
		allPattern)

	// distutil.MustCopyDirFiltered(srcDir, filepath.Join(name, "src/github.com/vugu/vugu"),
	// 	regexp.MustCompile(`^.*\.go$`),
	// 	// regexp.MustCompile(`^((.*\.go)|internal|domrender|js)$`),
	// )

	// log.Printf("TODO: copy vugu source into place")

	return name
}

// mustTGGoGet runs `go get` on the packages you give it with GO111MODULE=off and GOPATH set to the path you give.
// This can be used to
func mustTGGoGet(buildGopath string, pkgNames ...string) {
	// mustTGGoGet(buildGopath, "github.com/vugu/xxhash", "github.com/vugu/vjson")

	// oldDir, err := os.Getwd()
	// if err != nil {
	// 	panic(err)
	// }
	// defer os.Chdir(oldDir)

	// os.Chdir(buildGopath)

	var args []string
	args = append(args, "get")
	args = append(args, pkgNames...)
	fmt.Print(distutil.MustEnvExec([]string{"GO111MODULE=off", "GOPATH=" + buildGopath}, "go", args...))
}

// mustTGBuildAndLoad does a build and load - absdir is the original program path (the "test-NNN-desc" folder),
// and buildGopath is the temp dir where everything was copied in order to make non-module version that tinygo can compile
func mustTGBuildAndLoad(absdir, buildGopath string) string {
	// pathSuffix := mustTGBuildAndLoad(dir, buildGopath)

	// fmt.Print(distutil.MustEnvExec([]string{"GOOS=js", "GOARCH=wasm"}, "go", "build", "-o", filepath.Join(absdir, "main.wasm"), "."))

	// mustWriteSupportFiles(absdir)

	// FROM tinygo/tinygo-dev:latest

	args := []string{
		"run",
		"--rm", // remove after run
		// "-it",                            // connect console
		"-v", buildGopath + "/src:/go/src", // map src from buildGopath
		"-v", absdir + ":/out", // map original dir as /out so it can just write the .wasm file
		"-e", "GOPATH=/go", // set GOPATH so it picks up buildGopath/src
		"vugu/tinygo-dev:latest",                                                  // use latest dev (for now)
		"tinygo", "build", "-o", "/out/main.wasm", "-target", "wasm", "tgtestpgm", // tinygo command line
	}

	log.Printf("Executing: docker %v", args)

	fmt.Print(distutil.MustExec("docker", args...))

	fmt.Println("TODO: tinygo support files")

	// docker run --rm -it -v `pwd`/tinygo-dev:/go/src/testpgm -e "GOPATH=/go" tinygotest \
	// tinygo build -o /go/src/testpgm/testpgm.wasm -target wasm testpgm

	// # copy wasm_exec.js out
	// if ! [ -f tinygo-dev/wasm_exec.js ]; then
	// echo "Copying wasm_exec.js"
	// docker run --rm -it -v `pwd`/tinygo-dev:/go/src/testpgm tinygotest /bin/bash -c "cp /usr/local/tinygo/targets/wasm_exec.js /go/src/testpgm/"
	// fi

	uploadPath := mustUploadDir(absdir, "http://localhost:8846/upload")
	// log.Printf("uploadPath = %q", uploadPath)

	return uploadPath
}

func mustTGGenBuildAndLoad(absdir string, useDocker bool) string {

	mustTGGen(absdir)

	wc := devutil.MustNewTinygoCompiler().SetDir(absdir)
	defer wc.Close()

	if !useDocker {
		wc = wc.NoDocker()
	}

	outfile, err := wc.Execute()
	if err != nil {
		panic(err)
	}
	defer os.Remove(outfile)

	must(distutil.CopyFile(outfile, filepath.Join(absdir, "main.wasm")))

	wasmExecJSR, err := wc.WasmExecJS()
	must(err)
	wasmExecJSB, err := ioutil.ReadAll(wasmExecJSR)
	must(err)
	wasmExecJSPath := filepath.Join(absdir, "wasm_exec.js")
	must(ioutil.WriteFile(wasmExecJSPath, wasmExecJSB, 0644))

	mustWriteSupportFiles(absdir, false)

	uploadPath := mustUploadDir(absdir, "http://localhost:8846/upload")

	return uploadPath
}
