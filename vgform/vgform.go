/*
Package vgform has Vugu components that wrap HTML form controls for convenience.

NOTES: It would be tempting to wrap every HTML form control that exists.
However, I think we're going to find as things move forward that there is a
high value in keeping things as much "just HTML" as possible.  For a button,
for instance, I can't think of anything a wrapper component can do that
a regular HTML button tag can't.  In this case, we don't provide a component
because it doesn't do anything.  The core idea is that keeping things simple and
not wrapping things that don't need wrapping is more important than keeping
things "consistent" by wrapping everything.
*/
package vgform
