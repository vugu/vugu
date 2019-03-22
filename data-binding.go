package vugu

// NOTE: I messed around with this idea for a little while and the conclusion I've come to
// is that data binding and observers and all this are complicated, not-obvious, and will be
// cumbersome to the developer. All we really need to know is "did this component's data change"
// and be able to answer that question relatively quickly.  Wiring up events for every little change
// may not be a good way to go: I'm going to try just making it easy to hash the full tree of data
// and see how that works out.

// type Observable interface {
// 	AddObserver(o Observer)
// 	RemoveObserver(o Observer)
// }

// type Observer interface {
// 	Changed(i interface{})
// }

// type ObserverFunc func(i interface{})

// func (f ObserverFunc) Changed(i interface{}) { f(i) }

// // type DirtyBit bool

// // func (b *DirtyBit) Set(v bool) { *b = DirtyBit(v) }

// type Bool struct {
// 	value     bool
// 	observers []Observer
// }

// func (v *Bool) Set(value bool) {
// 	v.value = value
// }

// func (v *Bool) Get() bool {
// 	return v.value
// }

// func (v *Bool) AddObserver(o Observer) {
// 	// do not add the same one twice
// 	for _, el := range v.observers {
// 		if el == o {
// 			return
// 		}
// 	}
// 	v.observers = append(v.observers, o)
// }

// func (v *Bool) RemoveObserver(o Observer) {
// 	// find and remove if present
// 	for i, el := range v.observers {
// 		if el == o {
// 			v.observers = append(v.observers[:i], v.observers[i+1:]...)
// 			return
// 		}
// 	}
// }
