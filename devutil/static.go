package devutil

import (
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultIndex is the default index.html content for a development Vugu app.
// The exact text `<title>Vugu App</title>`, `<!-- styles -->` and `<!-- scripts -->`
// are meant to be replaced as needed if you quickly need to hack in CSS or
// JS references for a development Vugu application.  If you need more control
// than that, just copy it into your application.
var DefaultIndex = StaticContent(`<!doctype html>
<html>
<head>
<title>Vugu App</title>
<meta charset="utf-8"/>
<!-- styles -->
</head>
<body>
<div id="vugu_mount_point">
<img style="position: absolute; top: 50%; left: 50%;" src="https://cdnjs.cloudflare.com/ajax/libs/galleriffic/2.0.1/css/loader.gif">
</div>
<script src="https://cdn.jsdelivr.net/npm/text-encoding@0.7.0/lib/encoding.min.js"></script> <!-- MS Edge polyfill -->
<script src="/wasm_exec.js"></script>
<!-- scripts -->
<script>
var wasmSupported = (typeof WebAssembly === "object");
if (wasmSupported) {
	if (!WebAssembly.instantiateStreaming) { // polyfill
		WebAssembly.instantiateStreaming = async (resp, importObject) => {
			const source = await (await resp).arrayBuffer();
			return await WebAssembly.instantiate(source, importObject);
		};
	}
	var mainWasmReq = fetch("/main.wasm").then(function(res) {
		if (res.ok) {
			const go = new Go();
			WebAssembly.instantiateStreaming(res, go.importObject).then((result) => {
				go.run(result.instance);
			});		
		} else {
			res.text().then(function(txt) {
				var el = document.getElementById("vugu_mount_point");
				el.style = 'font-family: monospace; background: black; color: red; padding: 10px';
				el.innerText = txt;
			})
		}
	})
} else {
	document.getElementById("vugu_mount_point").innerHTML = 'This application requires WebAssembly support.  Please upgrade your browser.';
}
</script>
</body>
</html>
`)

// DefaultAutoReloadIndex is like DefaultIndex but also includes a script tag to load
// auto-reload.js from the default URL.
var DefaultAutoReloadIndex = DefaultIndex.Replace(
	"<!-- scripts -->",
	"<script src=\"http://localhost:8324/auto-reload.js\"></script>\n<!-- scripts -->")

var startupTime = time.Now()

// StaticContent implements http.Handler and serves the HTML content in this string.
type StaticContent string

// ServeHTTP implements http.Handler
func (sc StaticContent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, r.URL.Path, startupTime, strings.NewReader(string(sc)))
}

// Replace performs a single string replacement on this StaticContent and returns the new value.
func (sc StaticContent) Replace(old, new string) StaticContent {
	return StaticContent(strings.Replace(string(sc), old, new, 1))
}

// StaticFilePath implements http.Handler and serves the file at this path.
type StaticFilePath string

// ServeHTTP implements http.Handler
func (sfp StaticFilePath) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(string(sfp))
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	http.ServeContent(w, r, r.URL.Path, st.ModTime(), f)
}
