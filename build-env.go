package vugu

// // Env specifies the common methods for environment implementations.
// // See JSEnv and StaticHtmlEnv for implementations.
// type Env interface {
// 	RegisterComponentType(tagName string, ct ComponentType)
// 	Render() error
// }

// func NewBuildEnv(root Builder) (*BuildEnv, error) {

func NewBuildEnv() (*BuildEnv, error) {
	return &BuildEnv{}, nil
}

// BuildEnv is the environment used when building virtual DOM.
type BuildEnv struct {

	// components in pool from prior build
	compPool map[string]Builder

	// components used so far in this build
	compUsed map[string]Builder

	// nodePositionHashMap map[*VGNode]uint64

	// root Builder // FIXME: does this even belong here?  BuildEnv keeps track of components created but why does it need a reference to the root component?

}

// RunBuild performs a bulid on a component, managing the lifecycles of nested components and related concerned.
func (e *BuildEnv) RunBuild(builder Builder) (*BuildOut, error) {
	var buildIn BuildIn
	buildIn.BuildEnv = e

	// the pool becomes the used, and used becomes empty
	e.compPool = e.compUsed
	e.compUsed = make(map[string]Builder, len(e.compPool))

	// e.nodePositionHashMap = make(map[*VGNode]uint64, len(e.nodePositionHashMap))

	return builder.Build(&buildIn)
}

// FIXME: IMPORTANT: If we can separate the hash computation from the equal comparision, then we can use
// the hash do map lookups but then have a stable equal comparision, this way components will never
// be incorrectly reused, but still get virtually all of the benefits of using the hash approach for
// rapid comparision (i.e. "is this probably the same"/"find me one that is probably the same" is fast
// to answer).

// NOTE: seems like we have two very distinct types of component comparisions:
// 1. Should we re-use this instance?  (Basically, is the input the same - should ignore things like computed properties and other internal state)
//    ^ This seems like a "shallow" comparision - pointer struct fields should be compared on the basis of do they point to the same thing.
// 2. Is this component changed since last render?  (This should examine whatever it needs to in order to determine if a re-render is needed)
//    ^ This seems like a "deep" comparision against a last known rendered state - you don't care about what the pointers are, you
//      follow it until you get a value, and you check if it's "changed".

// NOTE: This whole thing seems to be a question of optimization. We could just create
// a new component for each pass, but we want to reuse, so it's worth thinking the entire thought through,
// and ask what happens if we optmize each step.

//----

/*
	Points to optimize

	- Don't recreate components that have the same input, reuse them (actually, why?? - because if they
	  compute internal state we sholud preserve that where possible). If we don't do this properly, then the
	  other two optimizations likely won't work either. (THIS IS BASICALLY THE "SHALLOW COMPARE" APPAROACH - BOTH FOR HASHING
	  AND EQUAL COMPARISION - COMPARE THE POINTERS NOT FOLLOWING THEM ETC)

	- Don't re-create VGNode tree for component if output will be the same (algo for how to determine this tbd)
	  - Breaks into "our VGNode stuff is the same" and "ours plus children is the same".

	- Don't re-sync VGNodes to render pipeline if they are the same.
	  - two cases here: 1. Same exact DOM for a component returned , 2. newly generated the same DOM
*/

// NOTE: Should we be using a pool for VGNode and VGAttribute allocation?  We are going to be creating and
// destroying a whole lot of these. MAYBE, BUT BENCHMARK SHOWS ONLY ABOUT 15% IMPROVEMENT USING THE POOL, MIGHT
// BE DIFFERENT IN REAL LIFE BUT PROBABLY NOT WORTH DOING RIGHT OUT THE GATE.

/*

	Basic sequence:

	- Build is called on root component
	- Component checks self to see if DOM output will be same as last time and if we have cached BuildOut, return it if so.
	- No BuildOut cached, run through rest of Build.
	- For each component encountered, give BuildEnv the populated struct and ask it for the instance to use
	  (it will pull from cache or use the object it was sent).
	- Component is stored on VGNode.Component field.  BuildOut also should keep a slice of these for quick traversal.
	- BuildOut is returned from root component's Build.
	- The list of components in the BuildOut is traversed, Build called for each one,
	  and the result set on VGNode.ComponentOut.
	- This causes the above cycle to run again for each of these child components.  Runs until no more are left.
	- FIXME: need to see how we combine the CSS and JS and make this accessible to the renderer (although maybe
	  the renderer can follow the component trail in BuildOut, or something)

	- At this point we have a BuildOut with a tree of VGNodes, and each one either has content itself or
	  has another BuildOut in the VGNode.ComponentOut field.  Between the two caching mechanisms (component
	  checking itself to see if same output, and each component creation checked with BuildEnv for re-use),
	  the cached case for traversing even a large case should be fast.

	- During render: The BuildOut pointer (or maybe its Out field) is used as a cache key - same BuildOut ptr, we assume same
	  VGNodes, and renderer can safely skip to each child component and continue from there.
	- For each VGNode, we call String() and have a map of the prior output for this position, if it's the same,
	  we can skip all the sync stuff and just move to the next.  String() needs to be very carefully implemented
	  so it can be used for equality tests like this safely.  The idea is that if we get a different VGNode
	  but with the exact same content, we avoid the extra render instructions.

	------------------
	TODO: We need to verify that component events and slots as planned
	https://github.com/vugu/vugu/wiki/Component-Related-Features-Design
	still work with this idea above.  I THINK WE CAN JUST ASSIGN THE
	SLOT AND EVENT CALLBACKS EACH TIME, THAT SHOULD WORK JUST FINE, WE
	DON'T NEED TO COMPARE AND KEEP THE OLD SLOT FUNCS ETC, JUST OVERWRITE.
*/

/*

STRUCT TAGS:

type Widget struct {

	// component param
	Size int `vugu:"cparam"`
	FirstName *string `vugu:"cparam"`

	// computed property, used for display, but entirely dependent upon Size
	DisplaySize string
}

*/

/*

DIRTY CHECKING:

Basic idea:

type DirtyChecker interface{
	DirtyCheck(oldData []byte) (isDirty bool, newData []byte)
	// or maybe just interface{}
	DirtyCheck(oldData interface{}) (isDirty bool, newData interface{})
}

// "mod" is good!  doesn't sound weird, "modify" is pretty clearly on point, and "mod" is short.

type ModChecker interface{
	ModCheck(oldData interface{}) (isDirty bool, newData interface{})
}

type SomeComponent struct {
	FirstName string `vugu:"modcheck"`

	FirstNameFormatted string // computed field, not "modcheck"'ed
}

*/

func (e *BuildEnv) Component(vgparent *VGNode, comp Builder) Builder {

	return comp
}

// // BuildRoot creates a BuildIn struct and calls Build on the root component (Builder), returning it's output.
// func (e *BuildEnv) BuildRoot() (*BuildOut, error) {

// 	var buildIn BuildIn
// 	buildIn.BuildEnv = e

// 	// TODO: SlotMap?

// 	return e.root.Build(&buildIn)
// }

// func (e *BuildEnv) ComponentFor(n *VGNode) (Builder, error) {
// 	panic(fmt.Errorf("not yet implemented"))
// }

// func (e *BuildEnv) SetComponentFor(n *VGNode, c Builder) error {
// 	panic(fmt.Errorf("not yet implemented"))
// }
