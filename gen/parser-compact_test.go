package gen

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	// "github.com/vugu/vugu/internal/htmlx"
	// "golang.org/x/net/html"
	"github.com/vugu/html"
)

func TestCompactNodeTree(t *testing.T) {

	assert := assert.New(t)

	var in = `<html>
<head></head>
<body>
	<div>
		<ul>
			<li vg-for="abc123" vg-html="something"></li>
		</ul>
		<p>
			This is some static text here, <strong>blah</strong> bleh <em>blee</em>.
		</p>
	</div>
</body>
</html>`

	n, err := html.Parse(bytes.NewReader([]byte(in)))
	if err != nil {
		t.Fatal(err)
	}

	assert.NoError(compactNodeTree(n))

	var buf bytes.Buffer
	assert.NoError(html.Render(&buf, n))

	log.Printf("OUT:\n%s", buf.String())

	assert.Contains(buf.String(), "<p vg-html=\"vugu.HTML(&#34;")

	// log.Printf("HERE: %#v", n.FirstChild.FirstChild.NextSibling.FirstChild.FirstChild)
	// log.Printf("HERE: %#v", n.FirstChild.FirstChild.NextSibling.NextSibling.FirstChild.FirstChild.NextSibling.NextSibling)
	// log.Printf("HERE: %#v", n.FirstChild.FirstChild.NextSibling.NextSibling) // body
	// log.Printf("HERE: %#v", n.FirstChild.FirstChild.NextSibling.NextSibling.FirstChild.NextSibling) // div
	// log.Printf("HERE: %#v", n.FirstChild.FirstChild.NextSibling.NextSibling.FirstChild.NextSibling.FirstChild.NextSibling) // ul
	// log.Printf("HERE: %#v", n.FirstChild.FirstChild.NextSibling.NextSibling.FirstChild.NextSibling.FirstChild.NextSibling.NextSibling.NextSibling)

	// p := n.FirstChild.FirstChild.NextSibling.NextSibling.FirstChild.NextSibling.FirstChild.NextSibling.NextSibling.NextSibling
	// log.Printf("p = %#v", p)
	// assert.NoError(html.Render(&buf, p))
	// log.Printf("OUT:\n%s", buf.String())

	// log.Printf("p.FirstChild = %#v", p.FirstChild)
	// log.Printf("p.FirstChild.NextSibling = %#v", p.FirstChild.NextSibling)
	// log.Printf("p.FirstChild.NextSibling.NextSibling = %#v", p.FirstChild.NextSibling.NextSibling)

	// assert.NoError(html.Render(&buf, p.FirstChild))
	// log.Printf("OUT:\n%s", buf.String())

}
