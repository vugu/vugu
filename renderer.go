package vugu

// Renderer takes a BuildOut ("virtual DOM") and renders it to it's final output.
type Renderer interface {
	Render(b *BuildOut) error
}
