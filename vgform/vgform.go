/*
Package vgform has Vugu components that wrap HTML form controls for convenience.

NOTES: It would be tempting to wrap every HTML form control that exists.
However, I think we're going to find as things move forward that there is a
high value in keeping things as much "just HTML" as possible.  For a button,
for instance, I can't think of anything a wrapper component can do that
a regular HTML button tag can't.  In this case, we don't provide a component
because it doesn't do anything.  The core idea is that keeping things simple and
not wrapping things that don't need wrapping is more important than keeping
things "consistent" by wrapping everything.  The aim here is to provide
components that are as close to the regular HTML tag as possible while
still providing useful functionality as it relates to integrating it with
the data in your Go program.  If a component doesn't do that, it should
not be included here.  That's the general idea anyway.  If it turns
out that this ends up wrapping 90% of the form controls anyway,
then maybe we just do 100% and the "consistency" argument wins.
Nonetheless, I still think this concern about not unnecessarily
abstracting things is important to Vugu component design.  Using forms
(and hopefully other components) should read as "it's basically HTML with
this additional functionality" rather than an entirely new language
to learn.  This will also come heavily into play in the design
of things like a library that works with Bootstrap CSS.
*/
package vgform

/*
Components:

Select
Input
	Text
	Password
	Email
    Checkbox - the only differnce here is the bool, might be better to just find a good way to bind a bool value
	Color
	Number
	Radio
	Range
	Search
	Tel
	Url

Textarea


https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input
Ones we're not doing because it would not add functionality:
Button
Output

Probably should add but needs more thought:
File
Date
Datetime-local
hidden
image
month
time
week

*/
