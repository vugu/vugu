package vugufmt

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/erinpentecost/byteline"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
	// StyleFormatter handles CSS blocks.
	StyleFormatter func([]byte) ([]byte, *FmtError)
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

// breaks returns the number of newlines if all input
// text is whitespace. Otherwise returns 0.
func breaks(input string) int {
	numBreaks := 0
	for _, s := range input {
		if !unicode.IsSpace(s) {
			return 0
		}
		if s == '\n' {
			numBreaks++
		}
	}
	return numBreaks
}

// FormatHTML formats script and css nodes.
func (f *Formatter) FormatHTML(filename string, in io.Reader, out io.Writer) *FmtError {
	lineTracker := byteline.NewReader(in)
	izer := html.NewTokenizer(lineTracker)
	ts := tokenStack{}

	previousLineBreak := false

	prevLine := 0
	prevCol := 0

loop:
	for {
		curTokType := izer.Next()

		totalOffset, _ := lineTracker.GetCurrentOffset()
		tokenOffset := totalOffset - len(izer.Buffered())
		curLine, curCol, _ := lineTracker.GetLineAndColumn(tokenOffset)

		// quit on errors.
		if curTokType == html.ErrorToken {
			if err := izer.Err(); err != nil {
				if err != io.EOF {
					return &FmtError{
						Msg:    err.Error(),
						Line:   prevLine,
						Column: prevCol,
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
					Line:   prevLine,
					Column: prevCol,
				}
			}
			return &FmtError{
				Msg:    "tokenization error",
				Line:   prevLine,
				Column: prevCol,
			}
		}

		curTok := izer.Token()

		// do indentation if we broke the line before this token.
		if previousLineBreak {
			indentLevel := len(ts)
			if curTokType == html.EndTagToken && indentLevel > 0 {
				indentLevel--
			}
			for i := 0; i < indentLevel; i++ {
				out.Write([]byte{'\t'})
			}
		}
		previousLineBreak = false

		raw := izer.Raw()
		raws := string(raw)
		// add or remove tokens from the stack
		switch curTokType {
		case html.StartTagToken:
			ts.push(&curTok)
			out.Write(raw)
		case html.EndTagToken:
			lastPushed := ts.pop()
			if lastPushed.DataAtom != curTok.DataAtom {
				return &FmtError{
					Msg:    fmt.Sprintf("mismatched ending tag (expected %s, found %s)", lastPushed.Data, curTok.Data),
					Line:   prevLine,
					Column: prevCol,
				}
			}
			out.Write(raw)
		case html.TextToken:
			parent := ts.top()

			if breakCount := breaks(raws); breakCount > 0 {
				// This is a break between tags.
				for i := 0; i < breakCount; i++ {
					out.Write([]byte{'\n'})
				}
				previousLineBreak = true
				continue loop
			}

			if parent == nil {
				out.Write(raw)
				// orphaned text node
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
					err.Line = prevLine + err.Line // messy!
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
						Line:   prevLine,
						Column: prevCol,
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

		prevLine = curLine
		prevCol = curCol
	}
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
