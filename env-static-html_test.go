package vugu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticHTMLEnv(t *testing.T) {

	assert := assert.New(t)

	s, err := parserGoBuildAndRun(`

<div class="outer">
	<p vg-html="data.Example"></p>
</div>

<script type="application/x-go">
import "os"

// type DemoComp struct {}

func (c *DemoComp) InitData() (interface{}, error) {
	return &DemoCompData{Example:"Some Data!"}, nil
}

type DemoCompData struct {
	Example string
}

func main() {
	env := vugu.NewStaticHTMLEnv(os.Stdout)
	inst, err := vugu.New(&DemoComp{})
	if err != nil { panic(err) }
	err = env.Render(inst)
	if err != nil { panic(err) }
}
</script>
`, false)
	assert.NoError(err)
	// log.Printf("OUT: %s", s)
	assert.Equal(`<div class="outer">
	<p>Some Data!</p>
</div>`, s)

}
