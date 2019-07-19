package vugu

import (
	"fmt"
	"io"
)

type StaticHTMLRenderer struct {
	Out io.Writer
}

func (r *StaticHTMLRenderer) Render(b *BuildOut) error {
	panic(fmt.Errorf("not yet implemented"))
}
