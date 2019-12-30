package gen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vugu/html"
)

func TestVgForExpr(t *testing.T) {
	tests := []struct {
		name string
		node *html.Node

		expectedRet   vgForAttr
		expectedError string
	}{
		{
			name:        "no attr",
			node:        &html.Node{},
			expectedRet: vgForAttr{},
		},
		{
			name: "attr not found",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-html", Val: "html"},
				},
			},
			expectedRet: vgForAttr{},
		},
		{
			name: "simple attr",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-html", Val: "html"},
					{Key: "vg-for", Val: "  value "},
				},
			},
			expectedRet: vgForAttr{
				expr: "value",
			},
		},
		{
			name: "attr with unknown option",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-html", Val: "html"},
					{Key: "vg-for.unknown", Val: "value"},
				},
			},
			expectedError: "option \"unknown\" unknown",
		},
		{
			name: "attr with noshadow option",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-html", Val: "html"},
					{Key: "vg-for.noshadow", Val: "value"},
				},
			},
			expectedRet: vgForAttr{
				expr:     "value",
				noshadow: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			attr, err := vgForExpr(tt.node)

			assert.Equal(tt.expectedRet, attr)
			if tt.expectedError == "" {
				assert.NoError(err)
			} else {
				assert.EqualError(err, tt.expectedError)
			}
		})
	}
}
