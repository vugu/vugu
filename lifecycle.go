package vugu

// // UnlockRenderer is something that releases a lock and requests a re-render.
// type UnlockRenderer interface {
// 	UnlockRender()
// }

// InitCtx is the context passed to an Init callback.
type InitCtx interface {
	EventEnv() EventEnv

	// TODO: decide if we want to do something like this for convenience
	// Lock() UnlockRenderer
}

type initCtx struct {
	eventEnv EventEnv
}

// EventEnv implements InitCtx
func (c *initCtx) EventEnv() EventEnv {
	return c.eventEnv
}

type initer0 interface {
	Init()
}
type initer1 interface {
	Init(ctx InitCtx)
}

func invokeInit(c interface{}, eventEnv EventEnv) {
	if i, ok := c.(initer0); ok {
		i.Init()
	} else if i, ok := c.(initer1); ok {
		i.Init(&initCtx{eventEnv: eventEnv})
	}
}

// ComputeCtx is the context passed to a Compute callback.
type ComputeCtx interface {
	EventEnv() EventEnv
}

type computeCtx struct {
	eventEnv EventEnv
}

// EventEnv implements ComputeCtx
func (c *computeCtx) EventEnv() EventEnv {
	return c.eventEnv
}

type computer0 interface {
	Compute()
}
type computer1 interface {
	Compute(ctx ComputeCtx)
}

func invokeCompute(c interface{}, eventEnv EventEnv) {
	if i, ok := c.(computer0); ok {
		i.Compute()
	} else if i, ok := c.(computer1); ok {
		i.Compute(&computeCtx{eventEnv: eventEnv})
	}
}

// DestroyCtx is the context passed to a Destroy callback.
type DestroyCtx interface {
	EventEnv() EventEnv
}

type destroyCtx struct {
	eventEnv EventEnv
}

// EventEnv implements DestroyCtx
func (c *destroyCtx) EventEnv() EventEnv {
	return c.eventEnv
}

type destroyer0 interface {
	Destroy()
}
type destroyer1 interface {
	Destroy(ctx DestroyCtx)
}

func invokeDestroy(c interface{}, eventEnv EventEnv) {
	if i, ok := c.(destroyer0); ok {
		i.Destroy()
	} else if i, ok := c.(destroyer1); ok {
		i.Destroy(&destroyCtx{eventEnv: eventEnv})
	}
}

// RenderedCtx is the context passed to the Rendered callback.
type RenderedCtx interface {
	EventEnv() EventEnv // in case you need to request re-render
	First() bool        // true the first time this component is rendered
}
