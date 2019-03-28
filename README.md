# Vugu

This project is an experimental Vue-inspired framework for web UIs written in Go and targeting webassembly.

## TODO:

* Template compliation - to Go source DONE
* Static HTML output DONE
* Data hashing DONE
* CSS for templates DONE
* Component functionality DONE IN STATIC HTML
* DOM syncing - can be naive at first and then get fancier with optimizations later NEED TESTING
* Events in JS - @click and custom events from components NEEDS TESTING
* Render loop - when does program exit and how often do we re-render DONE
* Components in JSEnv
* Figure out production build (probably since dev is just `go run devserver.go` prod can be something like `go run build-prod.go`, each file being just a few lines where config tweaks can go - also figure out a good prefix for these two files so it's obvious they are build process and not part of wasm output)

* Component events (separate from DOMEvents above)
* Component slots
* goimports to help with missing imports in .vugu files
* Pretty compile errors, so when things fail during dev you get the output right in the browser.
* Optimize DOM syncing.  Several places for immediate improvement: case where DOM is built from scratch (or replaced like from SSR), static nodes can in certain cases be collapsed to one big innerHTML (this applies to the vdom->dom sync code as well as to the code generator), and vdom->dom sync can have the number of js calls reduces (possibly by shipping over larger chunks of data like all of the attrs at once and doing the rest in JS with a helper lib, or maybe we can just be smarter about which methods we use)
* vugufmt - can we make the HTML all nice and pretty, do some basic indentation on the CSS, and then run the script tag through gofmt?  We could this during 
code conversion, could provide gofmt-like functionality overall and help a lot.
* Template compilation - direcly to VDOM - can be called from within wasm without the use of compiler, expressions are all template syntax {{.Blah}}
* Explore the idea of having the VDOM include html, head and body tags.  This topic of how to change the title and meta tags is important for SEO and yet rather silly, they are just tags like anything else.  It raises a some questions when we talk about replacing the entire page, even if we do it conditionally or with certain rules - for example if the root component sends back an empty head tag, does that mean it should remove all of the script and css includes?  Need to make some use cases and see what functionality would be useful and how it fits into what we have.  But it should not be so damned hard to change the title tag or include meta tags, etc.
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

* Type-safety wherever possible - one of the big strengths of the Go language is it's type system and compiler.  Where there is an idiomatic solution that uses it, prefer that over generic (type-unsafe) solutions.
