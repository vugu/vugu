module github.com/vugu/vugu/wasm-test-suite/test-017-nesting

go 1.23

toolchain go1.23.5

replace github.com/vugu/vugu => ../..

require (
	github.com/chromedp/chromedp v0.11.2
	github.com/vugu/vjson v0.0.0-20200505061711-f9cbed27d3d9
	github.com/vugu/vugu v0.0.0-00010101000000-000000000000
)

require (
	github.com/chromedp/cdproto v0.0.0-20250120090109-d38428e4d9c8 // indirect
	github.com/chromedp/sysutil v1.1.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/vugu/xxhash v0.0.0-20191111030615-ed24d0179019 // indirect
	golang.org/x/sys v0.29.0 // indirect
)
