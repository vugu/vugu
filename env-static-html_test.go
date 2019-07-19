package vugu

// func TestStaticHTMLEnv(t *testing.T) {

// 	assert := assert.New(t)

// 	s, err := parserGoBuildAndRun(`

// <style>
// .outer { background: green; }
// </style>

// <div class="outer">
// 	<p vg-html="data.Example"></p>
// </div>

// <script type="application/x-go">
// import "os"

// type DemoComp struct {}

// func (c *DemoComp) NewData(props vugu.Props) (interface{}, error) {
// 	return &DemoCompData{Example:"Some Data!"}, nil
// }

// type DemoCompData struct {
// 	Example string
// }

// func main() {
// 	inst, err := vugu.New(&DemoComp{}, nil)
// 	if err != nil { panic(err) }
// 	env := vugu.NewStaticHTMLEnv(os.Stdout, inst, nil)
// 	err = env.Render()
// 	if err != nil { panic(err) }
// }
// </script>
// `, false)
// 	assert.NoError(err)
// 	// log.Printf("OUT: %s", s)
// 	assert.Equal(`<style>
// .outer { background: green; }
// </style><div class="outer">
// 	<p>Some Data!</p>
// </div>`, s)

// }
