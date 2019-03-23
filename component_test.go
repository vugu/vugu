package vugu

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComponents(t *testing.T) {

	assert := assert.New(t)

	outs, err := parserGoBuildAndRunMulti(map[string]string{
		"RootComp": `
<div id="root">
	<top-comp :headline="data.TheTitle"></top-comp>
</div>

<script type="application/x-go">
import "os"

func main() {
	env := vugu.NewStaticHTMLEnv(os.Stdout, nil)
	env.RegisterComponentType("comp-one", &CompOne{})
	r, err := vugu.New(&RootComp{}, nil)
	if err != nil { panic(err) }
	err = env.Render(r)
	if err != nil { panic(err) }
}

type RootCompData struct {
	TheTitle string
}
func (c *RootComp) NewData(props vugu.Props) (interface{}, error) {
	return &RootCompData {
		TheTitle: "title goes here",
	}, nil
}
</script>
`,
		"CompOne": `
<div class="comp-one">
	<h1 vg-html="data.Headline"></h1>
</div>

<script type="application/x-go">
type CompOneData struct {
	Headline string
}
func (c *CompOne) NewData(props vugu.Props) (interface{}, error) {
	ret := &CompOneData {}
	ret.Headline, _ = props["headline"].(string)
	return ret, nil
}
</script>
`,
	}, false)
	assert.NoError(err)
	log.Printf("outs: %s", outs)
	assert.Equal(`<div id="root">
<div class="comp-one">
	<h1>title goes here</h1>
</div>
</div>`, outs)

	// var buf bytes.Buffer
	// env := NewStaticHTMLEnv(&buf, nil)
	// assert.NotNil(env)

	// env.RegisterComponentType("comp-one")

	// env.

	// assert.NoError(env.Render())

}
