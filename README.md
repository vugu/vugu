# Vugu

This project is a prototype for a Vue-inspired framework for web UIs written in Go and targeting webassembly.

## TODO:

* Template compliation - to Go source DONE
* Static HTML output DONE
* Data hashing DONE
* CSS for templates DONE
* Component functionality DONE
* DOM syncing - can be naive at first and then get fancier with optimizations later
* Events in JS - @click and custom events from components
* Figure out production build (probably since dev is just `go run devserver.go` prod can be something like `go run build-prod.go`, each file being just a few lines where config tweaks can go - also figure out a good prefix for these two files so it's obvious they are build process and not part of wasm output)

* Component slots
* Template compilation - direcly to VDOM - can be called from within wasm without the use of compiler, expressions are all template syntax {{.Blah}}
* Type-safety wherever possible - one of the big strengths of the Go language is it's type system and compiler.  Where there is an idiomatic solution that uses it, prefer that over generic (type-unsafe) solutions.
* Use Go for what it's good at: concurrency should use Go routines, multiple web requests can go in sequence in a simple function, or in parallel using a WaitGroup, etc.
* Vuex-like store for shared state
* Dev/prod webserver tooling - need a way to just refresh the page and it everything recompiles during dev, and then separate production output.
* Router
* Single-file components (.vugu files)
* Hot module reloading (we could gob-encode everything, dump that sessionStorage or something, reload and restore)
* Running examples
* Server-side rendering - it's all Go code so this is probably pretty easy.
* See what it would take to make a basic VSCode plugin for .vugu files
* A vuetify-like framework with material design components.  As a alternative, something like bootstrap-vue.js.org could be a possibility as well.
* Look into async templates (async code might be outside the scope of current wasm capability, but certainly we can pull templates from the server)

## FIXES

* Support struct tags for data hashing, like `hash:"omit"` (don't hash this field) or `hash:"string"` (use fmt.Print)

## NOTES

* Data binding - simple observer pattern, prefer clarity and simplicity over magicness - Bailed on this, too complicated, opted for data hashing instead.
