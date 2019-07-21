# Vugu

[![Go Report Card](https://goreportcard.com/badge/github.com/vugu/vugu)](https://goreportcard.com/report/github.com/vugu/vugu)
[![Travis CI](https://travis-ci.org/vugu/vugu.svg?branch=master)](https://travis-ci.org/vugu/vugu)
[![GoDoc](https://godoc.org/github.com/vugu/vugu?status.svg)](https://godoc.org/github.com/vugu/vugu)
[![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)](https://github.com/emersion/stability-badges#experimental)

Vugu is an experimental library for web UIs written in Go and targeting webassembly.  Guide and docs at https://www.vugu.org.
Godoc at https://godoc.org/github.com/vugu/vugu.

If you've ever wanted to write a UI not in JS but pure Go... and run it in your browser, right now... That (experimental;) future is here!

Introducing Vugu (pronounced /ˈvuː.ɡuː/), a VueJS-inspired library in Go targeting wasm.

No node. No JS. No npm. No node_modules folder competing with your music library for disk space.

## Updates ♨

* 2019-07-20 Some design info on how "data binding" (hashing actually) will work in Vugu: https://github.com/vugu/vugu/wiki/Data-Hashing-vs-Binding
* 2019-07-16 Vugu has a logo! https://www.instagram.com/p/Bz3zmtYAYcM/  Good things are in the works, the plan is to get a bunch of much-awaited updates pushed to master before the end of the month.
* 2019-05-19 Refactor still in progress - this is the cleaned-up architecture concept: https://github.com/vugu/vugu/wiki/Architecture-Overview
* 2019-04-07 The Vugu Playground is up at: https://play.vugu.org/
* 2019-04-05 Thanks to @erinpentecost, **vugufmt is now available** and provides gofmt-like functionality on your .vugu files. ("go get github.com/vugu/vugu/cmd/vugufmt && go install github.com/vugu/vugu/cmd/vugufmt")
* 2019-04-05 The component playground should be available soon; followed by some internal work to properly handle nested components in a type-safe way; then probably a router...

<img src="https://cdnjs.cloudflare.com/ajax/libs/ionicons/4.5.6/collection/build/ionicons/svg/logo-slack.svg" width="17" height="17"> Join the conversation: [Gophers on Slack](https://invite.slack.golangbridge.org/), channel #vugu

## Highlights

* Runs in-browser using WebAssembly
* Single-file components
* Vue-like markup syntax
* Write idiomatic Go code
* Rapid prototyping
* ~3 minute setup
* Standard Go build tools

## Start

Get started: http://www.vugu.org/doc/start

Still a work in progress, but a lot of things are already functional. Some work really well.

## Abbreviated Roadmap

- [x] Single-file components (looks similar to .vue); .vugu -> .go code generation.
- [x] Includes CSS in components.
- [x] Basic flow control with vg-if, vg-for and output with vg-html.
- [x] Dynamic attributes with `<tag :prop='expr'>`.
- [x] Nested components with dynamic properties `<my-custom-component :prop='expr'>`.
- [x] Efficiently syncs to browser DOM.
- [x] Static HTML output (great for tests).
- [x] DOM Events, click, etc.
- [x] Basic data hashing to avoid unnecessary computation where possible.
- [x] Basic dev and prod server tooling, easy to get started
- [ ] URL Router
- [ ] Server-side rendering (HTML generation works, needs URL Router to make it usable)
- [ ] Performance optimizations
- [ ] Go-only component events
- And much more...

## Notes

It's built **more like a library than a framework**.  While Vugu does do code generation for your .vugu component
files, (and will even output a default main_wasm.go for a new project and build your program automatically upon page refresh), 
fundamentally you are still in control.  Overall program flow, application wiring and initialization, the render loop
that keeps the page in sync with your components - you have control over all of that.
Frameworks call your code.  Vugu is a library, your code calls it (even if Vugu generates a bit of that for you in
the beginning to make things easier). One of the primary goals for Vugu, when it comes to developers first encountering it, 
was to make it very fast and easy to get started, but without imposing unnecessary limitations on how a project is structured.
Go build tooling (and now the module system) is awesome.  The idea is to leverage that to the furthest extent possible,
rather than reprogramming the wheel.

So you won't find a vugu command line tool that runs a development server, instead
you'll find in the docs an appropriate snippet of code you can paste in a file and `go run` yourself.  For the code
generation while there is an http.Handler that can do this upon page refresh, you also can (and should!) run `vugugen`
via `go generate`. There are many small decisions in Vugu which follow this philosophy: wherever reasonably possible,
just use the existing mechanism that already exists.  And keep doing that until there's proof that something
else is really needed.  So far it's been working well.  And it allows Vugu to focus on the specific things it 
brings to the table.
