# Build Instructions for the Examples

There are two ways to build and run the examples. 

The simplest uses [`docker`](docker.com) and [`mage`](https://magefile.org/) and uses the same build process that is used to build `vugu` as a whole.

The alternative, if you need to use an alternative build tool, is to replicate the steps that the `mage` tool takes.

## Before you begin

First you must clone the `vugu` repository:

`git clone https://github.com/vugu/vugu`

Now `cd` into the cloned repository directory

`cd vugu`

The examples themselves are in the `examples` directory.

This is only way to get the examples. In particular they are not `go get`'able. 

Once you have the examples you can either build them in place or copy an individual example elsewhere as required.


## The `mage` way

The `vugu` project uses `mage` as its preferred build tool to take care of the complexity of building and running the examples for you. You will need [`docker`](docker.com) and [`mage`](https://magefile.org/) preinstalled to use this approach

From the `vugu` directory where the repositotry has been cloned, the command to run all of the examples the command is:

`mage Examples`

If you only want a specific example then the command is:

`mage SingleExample <name of the example module to run>`

for example to run the `vg-if` example the command is

`mage SingleExample github.com/vugu/vugu/examples/vg-if`

Regardless of whether you run all the examples or just one you will then see this in message in the shell

```
Local nginx container started.
Connect to http://localhost:8889/<example-test-directory-name>
e,g. http://localhost:8889/fetch-and-display
To stop the local nginx container please run:
    mage StopLocalNginxForExamples
```

So for example to see the `vg-if` example you would browse to:

[http://localhost:8889/vg-if/](http://localhost:8889/vg-if/)

If you change an example then you need to rebuild it - using either `mage Examples` or `mage SingleExample <name of the example module to run>` and then browse to [http://localhost:8889/<example-test-directory-name>](http://localhost:8889/<example-test-directory-name>/) and refresh the browser to reload the Web Assembly file.

The `mage SingleExample github.com/vugu/vugu/examples/vg-if` command will build the example in question, as well as all of the `vugu` tools. This ensures that the example is always built with the latest version of the `vugu` tools.


## Building and running the examples without `mage`

*Note: This is a summary of the steps that the `mage` tool takes. In the case of any doubt the steps documented in `magefile.go` which is the top level `magefile` that `mage` executes are definitive,*

The steps to build an example are:

```
# install vugu itself as the next step depends on the vugu package
go install github.com/vugu/vugu

# install the vugugen tool
go install github.com/vugu/vugu/cmd/vugugen

# cd into the example you want to build from the directory where `vugu` has been cloned
cd examples/vg-if # for example the `vg-if` example

# now run vugugen to generate the *_gen.go fiels from the *.vugu files for the example
vugugen

# Now you have all of the *.go files you can build the web assembly. 
# This results in the web assembly file main.wasm being generated in the current directory.
# Note: we pass the module name and not the list of source files to the go build command
GOOS=js GOARCH=wasm go build -o ./main.wasm github.com/vugu/vugu/examples/vg-if
```

Now you have a web assembly binary you need to serve the web assembly file from a web server and the associated `index.html` and any other HTML and stylesheets.

The `vugu` project uses `nginx` running in a container to do this.
For the `vg-if` example the command executed by mage is:

```
# vugu-examples is the container name
# we mount the entire examples directory into the container so this command will serve any example in the examples directory
# assuming that the example has been built using the previous steps
# 8889 is the exposed port to connect to
docker run --name vugu-examples --mount type=bind,source=./examples,target=/usr/share/nginx/html,readonly -p 8889:80 -d nginx
``` 

You should now be able to browser to 

[http://localhost:8889/vg-if/](http://localhost:8889/vg-if/)

To run the `vg-if` example.

If you need to supply a different `nginx` config to the `nginx` `docker` container that can be dine with an additional mount option. [See the `nginx` docs](https://docs.nginx.com/nginx/admin-guide/installing-nginx/installing-nginx-docker/#maintaining-content-and-configuration-files-on-the-docker-host).


If you use an alternative web to `nginx` the overall proces to serve the `.wasm` web assembly file and the `index.html` will similar.
