package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	js "syscall/js"
	"time"

	"golang.org/x/net/html"
)

func main() {

	log.Printf("Starting...")

	vugucbName := fmt.Sprintf("vugucb_%d", rand.NewSource(time.Now().UnixNano()).Int63())
	vugucbName = "vugucb" // tmp
	log.Printf("vugucbName = %q", vugucbName)

	js.Global().Set(vugucbName, js.NewCallback(vugucbFunc))

	tmplSrc := `
<div id="testcomp">
	<ul>
		{{range .Items}}
		<li @click="handleClick" data-id="{{.ID}}">{{.Name}}</li>
		{{end}}
	</ul>
</div>
`
	// startTime := time.Now()
	t, err := template.New("").Parse(tmplSrc)
	if err != nil {
		panic(err)
	}
	// log.Printf("parse time: %v", time.Since(startTime))

	data := map[string]interface{}{
		"Items": []map[string]interface{}{
			map[string]interface{}{"Name": "test1", "ID": "t1"},
			map[string]interface{}{"Name": "test2", "ID": "t2"},
			map[string]interface{}{"Name": "test3", "ID": "t3"},
		},
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	log.Printf("Result: %s", buf.String())

	ts := time.Now()
	node, err := html.Parse(&buf)
	if err != nil {
		panic(err)
	}
	log.Printf("parse time: %v", time.Since(ts))
	log.Printf("node: %#v", node)
	log.Printf("node.FirstChild: %#v", node.FirstChild)
	log.Printf("node.FirstChild.FirstChild: %#v", node.FirstChild.FirstChild)

	// log.Printf("wip")
	// fmt.Printf("wip1abs\n")

	// FIXME: out of memory funk, need to make sure we have a clean exit where everything gets released
	<-make(chan struct{}, 0) // sleep indefinitely
}

func vugucbFunc(args []js.Value) {
	log.Printf("args[0].String() = %#v", args[0].String())
	args[0].Call("preventDefault")
	// log.Printf("args = %#v", args)
}
