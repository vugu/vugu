module github.com/vugu/vugu/wasm-test-suite/test-005-issue-80

replace github.com/vugu/vugu => ../..

go 1.22.3

require (
	github.com/chromedp/chromedp v0.10.0
	github.com/vugu/vjson v0.0.0-20200505061711-f9cbed27d3d9
	github.com/vugu/vugu v0.0.0-00010101000000-000000000000
)

require (
	github.com/chromedp/cdproto v0.0.0-20240810084448-b931b754e476 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/vugu/xxhash v0.0.0-20191111030615-ed24d0179019 // indirect
	golang.org/x/sys v0.24.0 // indirect
)
