// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
)

func main() {

	b, err := ioutil.ReadFile(`renderer-js-script.js`)
	if err != nil {
		panic(err)
	}

	// minify the JS
	m := minify.New()
	m.AddFunc("application/javascript", js.Minify)
	mr := m.Reader("application/javascript", bytes.NewReader(b))
	b, err = ioutil.ReadAll(mr)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "package domrender\n\n// GENERATED FILE, DO NOT EDIT!  See renderer-js-script-maker.go\n\nconst jsHelperScript = %q\n", b)

	err = ioutil.WriteFile("renderer-js-script.go", buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

}
