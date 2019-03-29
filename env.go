package vugu

// Env specifies the common methods for environment implementations.
// See JSEnv and StaticHtmlEnv for implementations.
type Env interface {
	RegisterComponentType(tagName string, ct ComponentType)
	Render() error
}
