package devutil

import (
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultIndex is the default index.html content for a development Vugu app.
var DefaultIndex = StaticContent(`<!doctype html>
<html>
<head>
<title>Vugu App</title>
<meta charset="utf-8"/>
<!-- TODO: should we include bootstrap or something on the default page? 
<link rel="stylesheet" href="" />
-->
<script src="https://cdn.jsdelivr.net/npm/text-encoding@0.7.0/lib/encoding.min.js"></script> <!-- MS Edge polyfill -->
<script src="/wasm_exec.js"></script>
</head>
<body>
<div id="vugu_mount_point">
<img style="position: absolute; top: 50%; left: 50%;" src="https://cdnjs.cloudflare.com/ajax/libs/galleriffic/2.0.1/css/loader.gif">
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
	WebAssembly.instantiateStreaming(fetch("/main.wasm"), go.importObject).then((result) => {
		go.run(result.instance);
	});
} else {
	document.getElementById("vugu_mount_point").innerHTML = 'This application requires WebAssembly support.  Please upgrade your browser.';
}
</script>
<!--

TODO: in the index page's loader code we should include some logic to dump
the contents of a non-200 response into a div so we can see it, or
something - this way we can keep our structure but get pretty error 
messages from the wasm compiler on-screen;

TODO: script include for vgrun?  hm, think through if there will be an issue
with this accidentally ending up live and if we need some "if localhost" logic.

-->
</body>
</html>
`)

var startupTime = time.Now()

// StaticContent implements http.Handler and serves the HTML content in this string.
type StaticContent string

// ServeHTTP implements http.Handler
func (sc StaticContent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, r.URL.Path, startupTime, strings.NewReader(string(sc)))
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
