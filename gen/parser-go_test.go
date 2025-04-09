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
			name: "no iteration vars",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "c.Items"},
				},
			},
			// WARNING these tests are very brittle. The new line is required.
			expectedResult: `for key, value := range c.Items {
`,
		},
		{
			name: "no iteration vars with vg-key",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "c.Items"},
					{Key: "vg-key", Val: "1"},
				},
			},
			expectedResult: `for key, value := range c.Items {
`,
		},
		{
			name: "key and value vars",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "k, v := range c.Items"},
				},
			},
			expectedResult: `for k, v := range c.Items {
`,
		},
		{
			name: "only key var",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "k := range c.Items"},
				},
			},
			expectedResult: `for k := range c.Items {
`,
		},
		{
			name: "only key var with vg-key",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "k := range c.Items"},
					{Key: "vg-key", Val: "1"},
				},
			},
			expectedResult: `for k := range c.Items {
`,
		},
		{
			name: "only value var",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "_, v := range c.Items"},
				},
			},
			expectedResult: `for _, v := range c.Items {
`,
		},
		{
			name: "only value var with vg-key",
			node: &html.Node{
				Attr: []html.Attribute{
					{Key: "vg-for", Val: "_, v := range c.Items"},
					{Key: "vg-key", Val: "1"},
				},
			},
			expectedResult: `for _, v := range c.Items {
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
