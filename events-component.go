package vugu

// TODO: we definitely need component events, not sure how yet...

// var ComponentEventStop = fmt.Errorf("stop") // special err to mean stop propagating but don't cause the render loop to error

// // ComponentEventHandler is implemented by component types that need to handle events from nested components.
// // Events at the component level are different from DOM events because they occur only in Go code and do not
// // call out to the browser environment.
// type ComponentEventHandler interface {
// 	OnComponentEvent(name string, compData interface{}, eventDetail interface{}) error
// }
