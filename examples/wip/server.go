// +build ignore

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/vugu/vugu"
)

func main() {

	wd, _ := os.Getwd()
	l := ":8855"
	log.Printf("Starting HTTP Server at %q", l)
	h := vugu.NewDevHTTPHandler(wd, http.Dir(wd))
	h.ParserGoPkgOpts.SkipGoMod = true
	log.Fatal(http.ListenAndServe(l, h))

	// 	listen := flag.String("listen", ":8855", "Host:port to listen on for HTTP requests")
	// 	buildPath := flag.String("build-path", ".", "Path to pass to go build indicating the package folder")
	// 	flag.Parse()

	// 	b, err := exec.Command("go", "env", "GOROOT").CombinedOutput()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	wasmExecJSPath := filepath.Join(strings.TrimSpace(string(b)), "misc/wasm/wasm_exec.js")

	// 	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	// 		if r.URL.Path == "/" {
	// 			fmt.Fprintf(w, `<!doctype html>
	// <html>
	// <head>
	// <meta charset="utf-8">
	// <script src="wasm_exec.js"></script>
	// </head>
	// <body>
	// <script>
	// var blah = function(e) {
	// 	console.log(e);
	// }
	// </script>
	// <div id="main">
	// Loading...
	// <button id="testbtn1" onClick="blah(event)">Test1</button>
	// <button id="testbtn2" onClick="vugucb(event)">Test2</button>
	// </div>
	// <script>
	// // FIXME: need to handle unloading properly and making sure the app exits and doesn't keep eating memory
	// // FIXME: check for wasm support and show an error message if not
	// const go = new Go();
	// WebAssembly.instantiateStreaming(fetch("/main.wasm"), go.importObject).then((result) => {
	// 	go.run(result.instance);
	// });
	// </script>
	// </body>
	// </html>
	// `)
	// 			return
	// 		}

	// 		if r.URL.Path == "/wasm_exec.js" {
	// 			w.Header().Set("Content-Type", "application/javascript")
	// 			http.ServeFile(w, r, wasmExecJSPath)
	// 			return
	// 		}

	// 		if r.URL.Path == "/main.wasm" {
	// 			// TODO: conditionally recompile, for now we do it every time; and use ServeFile get nice caching
	// 			// TODO: implement gzipping at least for this - initial test shows 5x size reduction
	// 			f, err := ioutil.TempFile("", "main_wasm_")
	// 			if err != nil {
	// 				panic(err)
	// 			}
	// 			fpath := f.Name()
	// 			f.Close()

	// 			cmd := exec.Command("go", "generate", *buildPath)
	// 			cmd.Env = append(cmd.Env, os.Environ()...)
	// 			b, err := cmd.CombinedOutput()
	// 			if err != nil {
	// 				log.Printf("Error from generate: %v; Output:\n%s", err, b)
	// 				http.Error(w, "Generation error!", 500)
	// 				return
	// 			}

	// 			// GOOS=js GOARCH=wasm go build -o wip.wasm .
	// 			cmd = exec.Command("go", "build", "-o", fpath, *buildPath)
	// 			cmd.Env = append(cmd.Env, os.Environ()...)
	// 			cmd.Env = append(cmd.Env, "GOOS=js", "GOARCH=wasm")
	// 			b, err = cmd.CombinedOutput()
	// 			if err != nil {
	// 				log.Printf("Error from compile: %v; Output:\n%s", err, b)
	// 				http.Error(w, "Compilation error!", 500)
	// 				return
	// 			}
	// 			w.Header().Set("Content-Type", "application/wasm")
	// 			http.ServeFile(w, r, fpath)
	// 			return
	// 		}

	// 		http.NotFound(w, r)
	// 	})

	// 	log.Printf("Starting server on %q", *listen)
	// 	s := &http.Server{
	// 		Addr:    *listen,
	// 		Handler: h,
	// 	}
	// 	log.Fatal(s.ListenAndServe())

}
