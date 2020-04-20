package vgform

import (
	"sort"
	"strings"
)

// KeyLister provides a list keys as a string slice.
// Keys are used in the `value` attribute of HTML option tags (with a select).
type KeyLister interface {
	KeyList() []string
}

// KeyListerFunc implements KeyLister as a function.
type KeyListerFunc func() []string

// KeyList implements the KeyLister interface.
func (f KeyListerFunc) KeyList() []string { return f() }

// TextMapper provides mapping from a key to the corresponding text.
// Text is used inside the contents of an HTML option tag (with a select).
// Text values are always HTML escaped.
type TextMapper interface {
	TextMap(key string) string
}

// TextMapperFunc implements TextMapper as a function.
type TextMapperFunc func(key string) string

// TextMap implements the TextMapper interface.
func (f TextMapperFunc) TextMap(key string) string { return f(key) }

// SimpleTitle implements TextMapper by replacing '-' and '_' with a space and calling strings.Title.
var SimpleTitle = TextMapperFunc(func(key string) string {
	return strings.Title(strings.NewReplacer("-", " ", "_", " ").Replace(key))
})

// Options is an interface with KeyList and TextMap.
// It is used to express the options for a select element.
// It intentionally does not support option groups or other
// advanced behavior as that can be accomplished using slots (TO BE IMPLEMENTED).
// Options is provided to make it easy for the common case of
// adapting a slice or map to be used as select options.
type Options interface {
	KeyLister
	TextMapper
}

// MapOptions implements the Options interface on a map[string]string.
// The keys will be returned in alphanumeric sequence (using sort.Strings),
// or you can call SortFunc to assign a custom sort function.
type MapOptions map[string]string

// KeyList implements KeyLister by returning the map keys sorted with sort.Strings().
func (m MapOptions) KeyList() []string {
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	sort.Strings(s)
	return s
}

// TextMap implements TextMapper by returning `m[key]`.
func (m MapOptions) TextMap(key string) string { return m[key] }

// SortFunc returns an Options instance that uses this map for
// keys and text and sorts according to the order specified by this
// function.
func (m MapOptions) SortFunc(sf func(i, j int) bool) Options {
	return customOptions{
		KeyLister: KeyListerFunc(func() []string {
			// build the key list directly, calling m.KeyList would call sort.Strings unnecessarily
			s := make([]string, 0, len(m))
			for k := range m {
				s = append(s, k)
			}
			sort.Slice(s, sf)
			return s
		}),
		TextMapper: m,
	}
}

// SliceOptions implements the Options interface on a []string.
// The slice specifies the sequence and these exact string keys are
// also used as the text. You can also call Title() to use the
// SimpleTitle mapper or use TextFunc to assign a custom text mapper.
type SliceOptions []string

// Title is shorthand for s.TextFunc(SimpleTitle).
func (s SliceOptions) Title() Options {
	return s.TextFunc(SimpleTitle)
}

// KeyList implements KeyLister with a type conversion ([]string(s)).
func (s SliceOptions) KeyList() []string { return []string(s) }

// TextMap implements TextMapper by returning the key as the text.
func (s SliceOptions) TextMap(key string) string { return key }

// TextFunc returns an Options instance that uses this slice
// as the key list and the specified function for text mapping.
func (s SliceOptions) TextFunc(tmf TextMapperFunc) Options {
	return customOptions{
		KeyLister:  s,
		TextMapper: tmf,
	}
}

type customOptions struct {
	KeyLister
	TextMapper
}
