// +build ignore

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/vugu/vugu/simplehttp"
)

func init() {

	simplehttp.DefaultPageTemplateSource = `<!doctype html>
<html>
<head>
{{if .Title}}
<title>{{.Title}}</title>
{{else}}
<title>Vugu Dev - {{.Request.URL.Path}}</title>
{{end}}
<meta charset="utf-8"/>
{{if .MetaTags}}{{range $k, $v := .MetaTags}}
<meta name="{{$k}}" content="{{$v}}"/>
{{end}}{{end}}
{{if .CSSFiles}}{{range $f := .CSSFiles}}
<link rel="stylesheet" href="{{$f}}" />
{{end}}{{end}}
<script src="https://cdn.jsdelivr.net/npm/text-encoding@0.7.0/lib/encoding.min.js"></script> <!-- MS Edge polyfill -->
<script src="/wasm_exec.js"></script>
</head>
<body>
<div id="root_mp">
{{if .ServerRenderedOutput}}{{.ServerRenderedOutput}}{{else}}
<img style="position: absolute; top: 50%; left: 50%;" src="https://cdnjs.cloudflare.com/ajax/libs/galleriffic/2.0.1/css/loader.gif">
{{end}}
</div>
<script>
var wasmSupported = (typeof WebAssembly === "object");
if (wasmSupported) {
	if (!WebAssembly.instantiateStreaming) { // polyfill
		WebAssembly.instantiateStreaming = async (resp, importObject) => {
			const source = await (await resp).arrayBuffer();
			return await WebAssembly.instantiate(source, importObject);
		};
	}
	const go = new Go();
	go.argv.push("-mount-point=#root_mp");
	WebAssembly.instantiateStreaming(fetch("/main.wasm"), go.importObject).then((result) => {
		go.run(result.instance);
	});
} else {
	document.getElementById("root_mount_parent").innerHTML = 'This application requires WebAssembly support.  Please upgrade your browser.';
}
</script>
</body>
</html>
`
}

func main() {
	wd, _ := os.Getwd()
	l := "127.0.0.1:19944"
	log.Printf("Starting HTTP Server at %q", l)
	h := simplehttp.New(wd, true)
	log.Fatal(http.ListenAndServe(l, h))
}
