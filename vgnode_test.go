package vugu

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
