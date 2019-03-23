package vugu

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
	"strings"
)

// Env specifies the common methods for environment implementations.
// See JSEnv and StaticHtmlEnv for implementations.
type Env interface {
	RegisterComponentType(tagName string, ct ComponentType)
	Render() error
}

// refid returns a unique ID string for an interface value.  The same input will
// return the same string each time, and it should be pretty quick.
func refid(i interface{}) string {
	ptrs := reflect.ValueOf(&i).Elem().InterfaceData()
	return fmt.Sprintf("%x_%x", ptrs[0], ptrs[1])
}

// tmplRun compiles and runs a template (TODO: with caching) and returns the output as a trimmed string
func tmplRun(
	tmplText string, // original template text to run
	vardefs map[string]interface{}, // gets converted to functional equiv of {{$key := $value}} for each one
	data interface{}, // the data ("." and "$") to execute with
	itemData interface{}, // gets converted to functional equiv of {{with $data}}...tmplText...{{end}}
) (s string, refs map[string]interface{}, reterr error) {

	txt := tmplText
	if itemData != nil {
		txt = `{{with itemdata}}` + txt + `{{end}}`
	}
	for k, v := range vardefs {
		if refs == nil {
			refs = make(map[string]interface{})
		}
		id := refid(v)
		refs[id] = v
		txt = `{{` + k + ` := (forref "` + id + `")}}` + txt
	}

	t, err := template.New("").Parse(tmplText)
	if err != nil {
		return "", nil, fmt.Errorf("template parse error for %q: %v", tmplText, err)
	}

	t.Funcs(template.FuncMap{
		"itemdata": func() interface{} {
			return itemData
		},
		"mkref": func(i interface{}) string {
			if refs == nil {
				refs = make(map[string]interface{})
			}
			id := refid(i)
			refs[id] = i
			return id
		},
		"forref": func(s string) interface{} {
			return refs[s]
		},
	})

	var buf bytes.Buffer
	reterr = t.Execute(&buf, data)

	s = strings.TrimSpace(buf.String())

	return
}
