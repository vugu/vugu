// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

func main() {

	b, err := ioutil.ReadFile(`renderer-js-script.js`)
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
