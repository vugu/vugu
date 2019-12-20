package gen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vugu/html"
)

func TestVgAttr(t *testing.T) {
	tests := []struct {
		name string
		node *html.Node
		key  string

		expectedVal     string
		expectedOptions map[string]bool
	}{
		{
			name:        "no attr",
			node:        &html.Node{},
			key:         "vg-if",
			expectedVal: "",
			expectedOptions: map[string]bool{
				"opt1": false,
				"opt2": false,
			},
		},
		{
			name: "attr not found",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-html", Val: "html"},
				},
			},
			key:         "vg-if",
			expectedVal: "",
			expectedOptions: map[string]bool{
				"opt1": false,
				"opt2": false,
			},
		},
		{
			name: "simple attr",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-html", Val: "html"},
					{Key: "vg-if", Val: "  true "},
				},
			},
			key:         "vg-if",
			expectedVal: "true",
			expectedOptions: map[string]bool{
				"opt1": false,
				"opt2": false,
			},
		},
		{
			name: "attr with one option",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-html", Val: "html"},
					{Key: "vg-if.opt1", Val: "true"},
				},
			},
			key:         "vg-if",
			expectedVal: "true",
			expectedOptions: map[string]bool{
				"opt1": true,
				"opt2": false,
			},
		},
		{
			name: "attr with two options",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-html", Val: "html"},
					{Key: "vg-if.opt1.opt2", Val: "true"},
				},
			},
			key:         "vg-if",
			expectedVal: "true",
			expectedOptions: map[string]bool{
				"opt1": true,
				"opt2": true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			attr := vgAttr(tt.node, tt.key)

			assert.Equal(tt.expectedVal, attr.val)
			for k, v := range tt.expectedOptions {
				assert.Equal(v, attr.hasOption(k), "option %s %t want %t", k, attr.hasOption(k), v)
			}
		})
	}
}
