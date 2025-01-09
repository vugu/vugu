# `html-form` Conditional example

A very simple example that demonstrates how write a form without the `vgform` package


## Building the example

*This build instructions are based on using `mage` as the built tool. If you require a different built tool please use the generic method of building in the [Readme](https://github.com/vugu/vugu/blob/master/examples/Readme.md) at the root of the `examples` package.*

Building the example is easy. 

You will need [`docker`](docker.com) and [`mage`](https://magefile.org/) preinstalled before you begin.

First you must clone the `vugu` repository:

`git clone https://github.com/vugu/vugu`

Now `cd` into the cloned repository directory

`cd vugu`

The `vugu` project uses `mage` as its preferred build tool, so building the example is simply

`mage SingleExample github.com/vugu/vugu/examples/html-form`

If that works then in the shell you will see:

```
Local nginx container started.
Connect to http://localhost:8889/<example-test-directory-name>
e,g. http://localhost:8889/fetch-and-display
To stop the local nginx container please run:
	mage StopLocalNginxForExamples
```

The `mage SingleExample github.com/vugu/vugu/examples/html-form` command will build the example in question, as well as all of the `vugu` tools. This ensures that the example is always built with the latest version of the `vugu` tools.

## Running the example

To run the example, we follow the instructions in the shell. 

Open a browser at [http://localhost:8889/html-form/](http://localhost:8889/html-form/)

And the example will load and run.

## Changing the example

Changing the example is also easy.

The general principle is that you should update the `root.go` or `root.vugu` files as needed and then rerun `mage SingleExample github.com/vugu/vugu/examples/html-form`, then refresh the browser so that it loads new `wasm` example binary. 

So for example update the `Root.ListLength()` method from:

```
// Return the current list length
func (c *Root) ListLength() int {
	return len(c.list)
}
```

to

```
// Return the current list length
func (c *Root) ListLength() int {
    fmt.Printf("Current list length: %d\n", len(c.list))
	return len(c.list)
}
```

Then run

```
mage SingleExample github.com/vugu/vugu/examples/html-form
```

Again browse to [http://localhost:8889/html-form/](http://localhost:8889/html-form/) and refresh the browser.

You should now see the list length being displayed in the JavaScript console when the list is rendered.
