package vugufmt

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/vugu/vugu/internal/htmlx"
	"github.com/vugu/vugu/internal/htmlx/atom"
)

// Formatter allows you to format vugu files.
type Formatter struct {
	// ScriptFormatters maps script blocks to formatting
	// functions.
	// For each type of script block,
	// we can run it through the supplied function.
	// If the function returns error, we should
	// not accept the output written to the writer.
	// You can add your own custom one for JS, for
	// example. If you want to use gofmt or goimports,
	// see how to apply options in NewFormatter.
	ScriptFormatters map[string]func([]byte) ([]byte, *FmtError)
	StyleFormatter   func([]byte) ([]byte, *FmtError)
}

// NewFormatter creates a new formatter.
// Pass in vugufmt.UseGoFmt to use gofmt.
// Pass in vugufmt.UseGoImports to use goimports.
func NewFormatter(opts ...func(*Formatter)) *Formatter {
	f := &Formatter{
		ScriptFormatters: make(map[string](func([]byte) ([]byte, *FmtError))),
	}

	// apply options
	for _, opt := range opts {
		opt(f)
	}

	return f
}

// FormatScript formats script text nodes.
func (f *Formatter) FormatScript(scriptType string, scriptContent []byte) ([]byte, *FmtError) {
	if f.ScriptFormatters == nil {
		return scriptContent, nil
	}
	fn, ok := f.ScriptFormatters[strings.ToLower(scriptType)]
	if !ok {
		return scriptContent, nil
	}
	return fn(scriptContent)
}

// FormatStyle formats script text nodes.
func (f *Formatter) FormatStyle(styleContent []byte) ([]byte, *FmtError) {
	if f.StyleFormatter == nil {
		return styleContent, nil
	}
	return f.StyleFormatter(styleContent)
}

// FormatHTML formats script and css nodes.
func (f *Formatter) FormatHTML(filename string, in io.Reader, out io.Writer) *FmtError {
	izer := htmlx.NewTokenizer(in)
	ts := tokenStack{}

	curTok := htmlx.Token{}
	for {
		curTokType := izer.Next()

		// quit on errors.
		if curTokType == htmlx.ErrorToken {
			if err := izer.Err(); err != nil {
				if err != io.EOF {
					return &FmtError{
						Msg:    err.Error(),
						Line:   curTok.Line,
						Column: curTok.Column,
					}
				}
				// it's ok if we hit the end,
				// provided the stack is empty
				if len(ts) == 0 {
					return nil
				}
				tagNames := make([]string, len(ts), len(ts))
				for i, t := range ts {
					tagNames[i] = t.Data
				}
				return &FmtError{
					Msg:    fmt.Sprintf("missing end tags (%s)", strings.Join(tagNames, ", ")),
					Line:   curTok.Line,
					Column: curTok.Column,
				}
			}
			return &FmtError{
				Msg:    "tokenization error",
				Line:   curTok.Line,
				Column: curTok.Column,
			}
		}

		curTok := izer.Token()

		raw := izer.RawData()
		// add or remove tokens from the stack
		switch curTokType {
		case htmlx.StartTagToken:
			ts.push(&curTok)
			out.Write(raw)
		case htmlx.EndTagToken:
			lastPushed := ts.pop()
			if lastPushed.DataAtom != curTok.DataAtom {
				return &FmtError{
					Msg:    fmt.Sprintf("mismatched ending tag (expected %s, found %s)", lastPushed.Data, curTok.Data),
					Line:   curTok.Line,
					Column: curTok.Column,
				}
			}
			out.Write(raw)
		case htmlx.TextToken:
			parent := ts.top()
			if parent == nil {
				out.Write(raw)
				//return fmt.Errorf("%s:%v:%v: orphaned text node",
				//	filename, curTok.Line, curTok.Column)
			} else if parent.DataAtom == atom.Script {
				// determine the type of the script
				scriptType := ""
				for _, st := range parent.Attr {
					if st.Key == "type" {
						scriptType = st.Val
					}
				}

				// hey we are in a script text node
				fmtr, err := f.FormatScript(scriptType, raw)
				// Exit out on error.
				if err != nil {
					err.Line += curTok.Line
					err.FileName = filename
					return err
				}
				out.Write(fmtr)

			} else if parent.DataAtom == atom.Style {
				// hey we are in a CSS text node
				fmtr, err := f.FormatStyle(raw)
				if err != nil {
					return &FmtError{
						Msg:    err.Error(),
						Line:   curTok.Line,
						Column: curTok.Column,
					}
				}
				out.Write(fmtr)
			} else {
				// we are in some other text node we don't care about.
				out.Write(raw)
			}
		default:
			out.Write(raw)
		}
	}
}

// tokenStack is a stack of nodes.
type tokenStack []*htmlx.Token

// pop pops the stack. It will panic if s is empty.
func (s *tokenStack) pop() *htmlx.Token {
	i := len(*s)
	n := (*s)[i-1]
	*s = (*s)[:i-1]
	return n
}

// push inserts a node
func (s *tokenStack) push(n *htmlx.Token) {
	i := len(*s)
	(*s) = append(*s, nil)
	(*s)[i] = n
}

// top returns the most recently pushed node, or nil if s is empty.
func (s *tokenStack) top() *htmlx.Token {
	if i := len(*s); i > 0 {
		return (*s)[i-1]
	}
	return nil
}

// index returns the index of the top-most occurrence of n in the stack, or -1
// if n is not present.
func (s *tokenStack) index(n *htmlx.Token) int {
	for i := len(*s) - 1; i >= 0; i-- {
		if (*s)[i] == n {
			return i
		}
	}
	return -1
}

// insert inserts a node at the given index.
func (s *tokenStack) insert(i int, n *htmlx.Token) {
	(*s) = append(*s, nil)
	copy((*s)[i+1:], (*s)[i:])
	(*s)[i] = n
}

// remove removes a node from the stack. It is a no-op if n is not present.
func (s *tokenStack) remove(n *htmlx.Token) {
	i := s.index(n)
	if i == -1 {
		return
	}
	copy((*s)[i:], (*s)[i+1:])
	j := len(*s) - 1
	(*s)[j] = nil
	*s = (*s)[:j]
}

// Diff will show differences between input and what
// Format() would do. It will return (true, nil) if there
// is a difference, (false, nil) if there is no difference,
// and (*, notnil) when the difference can't be determined.
// filename is optional, but helps with generating useful output.
func (f *Formatter) Diff(filename string, input io.Reader, output io.Writer) (bool, error) {
	if filename == "" {
		filename = "<not set>"
	}

	var resBuff bytes.Buffer
	src, err := ioutil.ReadAll(input)
	if err != nil {
		return false, err
	}
	if err := f.FormatHTML(filename, bytes.NewReader(src), &resBuff); err != nil {
		return false, err
	}
	res := resBuff.Bytes()

	// No difference!
	if bytes.Equal(src, res) {
		return false, nil
	}

	// There is a difference, so what is it?
	data, err := diff(src, res, filename)
	if err != nil {
		return true, fmt.Errorf("computing diff: %s", err)
	}
	output.Write([]byte(fmt.Sprintf("diff -u %s %s\n", filepath.ToSlash(filename+".orig"), filepath.ToSlash(filename))))
	output.Write(data)
	return true, nil
}
