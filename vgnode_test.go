package vugu

import (
	"sync"
	"testing"
	// "github.com/vugu/html"
	// "github.com/vugu/html/atom"
	// html "golang.org/x/net/html"
	// atom "golang.org/x/net/html/atom"
)

//go:noinline
func allocVGNode() *VGNode {
	var ret VGNode
	return &ret
}

func BenchmarkAlloc(b *testing.B) {

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		v := allocVGNode()
		_ = v
	}

}

var vgnodePool = sync.Pool{New: func() interface{} { return &VGNode{} }}

func BenchmarkPool(b *testing.B) {

	objlist := make([]*VGNode, 0, 10)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		v := vgnodePool.Get().(*VGNode)
		*v = VGNode{} // zero it out

		// stack up 10 at a time and then release them to the pool
		objlist = append(objlist, v)
		if len(objlist) >= 10 {
			for _, o := range objlist {
				vgnodePool.Put(o)
			}
			objlist = objlist[:0]
		}
	}

}

// func TestFuncPtr(t *testing.T) {

// 	s1 := "s1"
// 	s2 := "s2"

// 	fA := fp(&s1)
// 	fB := fp(&s1)
// 	fC := fp(&s2)

// 	log.Printf("fA()=%v, fB()=%v, fC()=%v", fA(), fB(), fC())

// 	// log.Printf("fA=%v, fB=%v, fC=%v", unsafe.Pointer(fA), unsafe.Pointer(fB), unsafe.Pointer(fC))
// 	// log.Printf("fA=%v, fB=%v, fC=%v", fA, fB, fC)
// 	log.Printf("fA=%v, fB=%v, fC=%v", reflect.ValueOf(fA).Pointer(), reflect.ValueOf(fB).Pointer(), reflect.ValueOf(fC).Pointer())

// 	log.Printf("fA==fB=%v, fA==fC=%v", fA == fB, fA == fC)

// }

// func fp(strp *string) func() string {
// 	// capture one variable
// 	return func() string {
// 		return *strp
// 	}
// }

// func TestParseTemplate(t *testing.T) {

// 	assert := assert.New(t)

// 	in := `
// <div id="whatever">
// 	<ul>
// 		<li vg-if=".Test1" vg-range=".Test2" @click="something" :testbind="bound">Blah!</li>
// 	</ul>
// </div>
// `

// 	n, err := ParseTemplate(bytes.NewReader([]byte(in)))
// 	assert.NoError(err)
// 	assert.NotNil(n)

// 	found := false
// 	assert.NoError(n.Walk(func(v *VGNode) error {
// 		if v.Type == ElementNode && v.Data == "li" {
// 			found = true
// 			assert.Equal(".Test1", v.VGIf.Val)
// 			assert.Equal(".Test2", v.VGRange.Val)
// 			assert.Equal("@click", v.EventAttr[0].Key)
// 			assert.Equal(":testbind", v.BindAttr[0].Key)
// 			assert.Equal("bound", v.BindAttr[0].Val)
// 		}
// 		return nil
// 	}))
// 	assert.True(found)

// }

// func TestTmp(t *testing.T) {

// 	inRaw := []byte(`<li>hello</li>

// <script type="application/x-go">

// type DemoLine struct {
// 	Num int
// }

// </script>
// `)

// 	nlist, err := html.ParseFragment(bytes.NewReader(inRaw), &html.Node{
// 		Type:     html.ElementNode,
// 		DataAtom: atom.Body,
// 		Data:     "body",
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	var buf bytes.Buffer
// 	for i := range nlist {
// 		log.Printf("%#v\n", nlist[i])
// 		err = html.Render(&buf, nlist[i])
// 		if err != nil {
// 			panic(err)
// 		}
// 	}

// 	log.Printf("OUT: %s", buf.String())

// }
