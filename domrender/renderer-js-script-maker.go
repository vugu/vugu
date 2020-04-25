// +build ignore

package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
)

func main() {

	debug := flag.Bool("debug", false, "Keep debug lines in output")
	flag.Parse()

	// *debug = true

	f, err := os.Open("renderer-js-script.js")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	opcodeMap := make(map[string]bool, 42)

	var buf bytes.Buffer

	br := bufio.NewReader(f)
	for {
		bline, err := br.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			if len(bline) == 0 {
				break
			}
			continue
		} else if err != nil {
			panic(err)
		}

		// keep a map of the opcodes
		if bytes.HasPrefix(bytes.TrimSpace(bline), []byte("const opcode")) {
			fields := strings.Fields(string(bline))
			opcodeMap[fields[1]] = true
		}

		// only include debug lines if in debug mode
		if bytes.HasPrefix(bytes.TrimSpace(bline), []byte("/*DEBUG*/")) {
			if *debug {
				buf.Write(bline)
			}
			continue
		}

		// map of opcodes as text
		if bytes.Compare(bytes.TrimSpace(bline), []byte("/*DEBUG OPCODE STRINGS*/")) == 0 {
			if *debug {
				fmt.Fprintf(&buf, "let textOpcodes = [];\n")
				for k := range opcodeMap {
					fmt.Fprintf(&buf, "textOpcodes[%s] = %q; ", k, k)
				}
				fmt.Fprintf(&buf, "\n")
			}
			continue
		}

		// anything else just goes as-is
		buf.Write(bline)

	}

	b := buf.Bytes()

	// minify the JS
	m := minify.New()
	m.AddFunc("text/javascript", js.Minify)
	mr := m.Reader("text/javascript", bytes.NewReader(b))
	b, err = ioutil.ReadAll(mr)
	if err != nil {
		panic(err)
	}

	buf.Reset()

	fmt.Fprintf(&buf, "package domrender\n\n// GENERATED FILE, DO NOT EDIT!  See renderer-js-script-maker.go\n\nconst jsHelperScript = %q\n", b)

	err = ioutil.WriteFile("renderer-js-script.go", buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

}
