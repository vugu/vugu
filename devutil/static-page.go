package devutil

type StaticPage string

// TODO: StaticPage.ServeHTTP method

var DefaultIndex = StaticPage(`...

TODO: in the index page's loader code we should include some logic to dump
the contents of a non-200 response into a div so we can see it, or
something - this way we can keep our structure but get pretty error 
messages from the wasm compiler on-screen;

TODO: script include for vgrun?  hm, think through if there will be an issue
with this accidentally ending up live and if we need some "if localhost" logic.

`)
