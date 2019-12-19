package gen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vugu/html"
)

func TestEmitForExpr(t *testing.T) {
	tests := []struct {
		name           string
		node           *html.Node
		expectedError  string
		expectedResult string
	}{
		{
			name:          "no vg-for attributes",
			node:          &html.Node{},
			expectedError: "no for expression, code should not be calling emitForExpr when no vg-for is present",
		},
		{
			name: "no iteration variables",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "c.Items"},
				},
			},
			expectedResult: `for key, value := range c.Items {
var vgiterkey interface{} = key
_ = vgiterkey
key := key
_ = key
value := value
_ = value
`,
		},
		{
			name: "no iteration variables with vg-key",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "c.Items"},
					{Key: "vg-key", Val: "1"},
				},
			},
			expectedResult: `for key, value := range c.Items {
var vgiterkey interface{} = 1
_ = vgiterkey
key := key
_ = key
value := value
_ = value
`,
		},
		{
			name: "key and value variables",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "k, v := range c.Items"},
				},
			},
			expectedResult: `for k, v := range c.Items {
var vgiterkey interface{} = k
_ = vgiterkey
k := k
_ = k
v := v
_ = v
`,
		},
		{
			name: "only value variable",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "_, v := range c.Items"},
				},
			},
			expectedResult: `for vgiterkeyt , v := range c.Items {
var vgiterkey interface{} = vgiterkeyt
_ = vgiterkey
v := v
_ = v
`,
		},
		{
			name: "only value variable with vg-key",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "_, v := range c.Items"},
					{Key: "vg-key", Val: "1"},
				},
			},
			expectedResult: `for _, v := range c.Items {
var vgiterkey interface{} = 1
_ = vgiterkey
v := v
_ = v
`,
		},
		{
			name: "iteration with for clause",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "i:= 0; i < 5; i++"},
				},
			},
			expectedResult: `for i:= 0; i < 5; i++ {
var vgiterkey interface{} = i
_ = vgiterkey
i := i
_ = i
`,
		},
		{
			name: "iteration with for clause with vg-key",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "i:= 0; i < 5; i++"},
					{Key: "vg-key", Val: "1"},
				},
			},
			expectedResult: `for i:= 0; i < 5; i++ {
var vgiterkey interface{} = 1
_ = vgiterkey
i := i
_ = i
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			pg := &ParserGo{}
			state := &parseGoState{}

			err := pg.emitForExpr(state, tt.node)

			if tt.expectedError != "" {
				require.EqualError(err, tt.expectedError)
				return
			}
			require.NoError(err)
			assert.Exactly(tt.expectedResult, state.buildBuf.String())
		})
	}
}
