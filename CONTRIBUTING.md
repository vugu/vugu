# Contributing

Please first discuss via GitHub issue the change you wish to make.
And please follow the code of conduct outlined below during all interactions.

## Developing on a Fork

To work on a copy of Vugu, the easiest way is to use the `replace` module directive as described in https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive

This way your forked or local copy of Vugu can correspond to the import path `github.com/vugu/vugu`.  Example:

```bash
git checkout https://github.com/vugu/vugu # or, your forked copy
```

And now you can reference it from another project:

some-other-project/go.mod:
```
module some-other-project

// point vugu imports at your local copy instead of public download
replace github.com/vugu/vugu => ../vugu
```

## Vugu Project Layout

Some important subdirectories in the Vugu project are:
* The top level (corresponding to `github.com/vugu/vugu` import path) has common things shared across the other packages.
* `cmd` contains command-line executables, `vugugen` being the most commonly used.  Note that these command-line tools are generally small and just a thin wrapper around functionality from another package.
* `devutil` has utilities specifically created to facility rapid development such as wrappers for each Go+Wasm compilation and an HTTP muxer more suitable for Vugu development than the default one from net/http.
* `domrender` has the client-side code which synchronizes Vugu's virtual DOM tree with the browser's DOM.
* `gen` has code generaiton logic used to convert .vugu files into .go source code.  Most of the implementation of `vugugen` is available here.
* `js` has a wrapper around `syscall/js` which delegates to the default implementation in WebAssembly but in other environments allows compilation and functionality gracefully degrades.  Useful for writing code that needs to compile both in and outside Wasm environments without having to maintain two separate implementations via build tags just to include one reference to e.g. js.Value.
* `wasm-test-suite` contains the wasm tests. These require the use of a dockerized nginx and dockerised headless chrome run. See Legacy Running Tests below
* `legacy-wasm-test-suite` contains a test suite using Headless Chrome, see Legacy Running Tests below. 
* `magefiles` contains the Mage build script called `magefile.go`. This defines the top level targets. The other Go files contain the lower level functionality that the Magefile depend on. The `go.mod` and `go.sum` in this directory are related to `mage` and NOT to `vugu`.
* `testing` contains a collection of packages used by the `wasm-test-suite` that simply the test cases.
* The following folders are being deprecated and should be used with caution: 
  * `simplehttp` is an earlier attempt at development tooling use `devutil` instead.  `tinygo-dev` will be removed, the working parts have moved into `devutil`.


## Building

`vugu` now uses [mage](https://magefile.org/) to manage the build of the `vugu` tools - `vugugen`, `vugufmt` and `vgfrom`.
[Mage](https://magefile.org/) is also used to manage the testing process.


`vugugen` is the tool that combines your `*.vugu` files with your component `*.go` files before generating the final Go source files that
can then be compiled into a `wasm` binary by the Go Compiler.

If you don't have the `mage` tool installed the simplest way to install it is:

```
git clone https://github.com/magefile/mage
cd mage
go run bootstrap.go
```

You must run `mage` from the module root of `vugu`, this is the directory where the top level `go.mod` exists.

You will also need the [`goimports`](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) tool installed. It is very likely that you have this installed already, normally as part of an editor plugin. If not then the it can be installed with:

```
go install golang.org/x/tools/cmd/goimports@latest
```

In order to run the tests you will also need `docker` installed. If you don't have `docker` installed then follow the [docker install instructions](https://docs.docker.com/engine/install/#licensing).


**You must have Go v1.22.3 as a minimum to build `vugu`. We require the for loop changes that were introduced in Go v1.22, and v1.22.3 was the lastest at the time writing.**

The `Magefile` is self documenting. To see all of the targets, execute:

```
cd /path/to/vugu
mage -l
```

In general you want to use the `all` target to build, lint and test everything. Its used like so:

```
cd /path/to/vugu
mage all
```

or

```
cd /path/to/vugu
mage -v all
```

At the moment the `mage` build is still quite noisy, it will generate a significant amount output to the console even without the verbose `-v` option. This will be gradually reduce in future releases.

### Building using Mage

If you only need to build the `vugugen` and `vugufmt` tools then building these is as simple as:

```
cd /path/to/vugu
mage build
```

This will build the `vugu` root module, and build and install the `vugugen` and `vugufmt` tools.

It won't run any of the test suites.

## Running Unit Tests

If you just want to run the unit tests for `vugu` itself then that is simply:

```
cd /path/to/vugu
mage test
```

This won't however run amy of of the `wasm` tests.

## Cleaning the autogenerated files

Running any of the `all`, `build`, `test`, `testWasm` or `testlegacyWasm` targets will generate a significant number of autogenerated files.

This should be cleaned before committing to the repository. This can be achieved with:

```
cd /path/to/vugu
mage cleanAutoGeneratedFiles
```

These files are only cleaned when a new build is started. They are left after building in case the build failed and the files need to be examined to determine the reason for the failure.

## Running `wasm` tests

The `wasm` tests are more complex to execute. A simple `go test` won't work.

The `wasm` tests work like this. `vugu` will generate via the `vugugen` tool, that should be called via the `go generate` tooling a series of `*.go` files for each of the `*.vugu` files in the test. However the only way to validate that compiling the resulting Go files to `wasm` does what we expect is to serve the resulting `wasm` file via a web browser.

This however isn't enough, we also need to execute the `wasm` that is being served. For this we need a headless browser.

To validate the result of the execution of the `wasm` in a headless browser we need a standard Go test. This Go test uses the [`chronedp`](https://github.com/chromedp/chromedp) package to connect to the headless browser and examine the resulting HTML after the `wasm` file has executed in the headless browser.

To achieve this requires using `docker` and specifically it relies on `docker`'s networking capabilities.

The approach works like this:

There are two docker containers. One container runs `nginx` and serves the files in the `/wasm-test-suite` directory. The second container contains a headless chrome image, via `chromedp`.

Both of these containers are connected to a private `docker` network called `vugu-net`. The network provices communication as well as service loopkup/DNS resolution.

The test itself is a standard Go test. It first connects to the headless chrome instance running on `localhost:9222`, and then asks the headless chrome instance to connect to the `nginx` container.

Once the `nginx` instance (in the first docker container) has servered the `wasm` to the headless chrome browser (in the second docker container) the Go test can query the resulting state of the DOM via `chromedp` in the headless chrome (in the second container). It is this DOM querying that determines if the test passes of fails.

The good news is that `mage` hides all of this complexity.

Running the whole of the `wasm-test-suite` is as simple as:

```
cd /path/to/vugu
mage testWasm
```

if you only want to run a single wasm test, useful if only one test is failing or you are developing a new test, that can be achieved with:

```
cd /path/to/vugu
mage testSingleWasmTest <test-module-name>
```

For example to run the test `test-002-click` which is located in the `wasm-test-suite/test-002-click` directory its as simple as:

```
cd /path/to/vugu
mage testSingleWasmTest github.com/vugu/vugu/wasm-test-suite/test-002-click
```

*Note: At present all of the `wasm` tests are built with the standard Go compiler. The `mage` based build does not yet support building the test cases with the `tinygo` compiler suite. Using the `tinygo` compiler suite to built these tests will be added again at a future date.* 

## Running a `wasm` test manually

In order to debug a failing test it can often be helpful to see what is happening in a browser. The `mage` based build supports this quite naturally. Just execute:

```
cd /path/to/vugu
mage startLocalNginx
```

and browse to a URL of the following format:

```
http://localhost:8888/<test-directory-name>
```

For example to examine the test `test-001-simple` the URL would be:

```
http://localhost:8888/test-001-simple
```

Using the `testSingleWasmTest` will rebuild the `wasm` for you, although it will also execute the presumably failing test.

*We may add the ability to build, but not execute a single `wasm` test at a future date.*

The `startLocalNginx` target is safe, in that it can be called multiple times without calling the corresponding `stopLocalNginx` target. Any running `nginx` container will always be stopped before any new one is started.


## Creating a new `wasm` test

If you need to create a new `wasm` test in the `wasm-test-suite` the process is fairly straight forward. The critical point is to base it on a working test case and to ideally follow the directory naming convention.

For example

```
cd /path/to/vugu
cp -r ./wasm-test-suite/test-001-simple/ ./wasm-test-suite/test-NNN-objective

```

Where `NNN` is a 3 digit number

The `cp` will copy everything in the directory including the critical local `.gitignore` file to the new tests directory. Please make sire the `.gitignore` is present to ensure that vugu generated files are not submitted to the repository.

You then need to edit the `./wasm-test-suite/test-NNN-objective/go.mod` to change the module name. ***This step is critical.***

The module name must be changed to match the test, so in this case the module name would be changed to `github.com/vugu/vugu/wasm-test-suite/test-NNN-objective`

You can then edit the `root.vugu`, `root.go` as needed to support the test case.

The Go test function itself should also be renamed to reflect the test, so in this case the test function in `main_test.go` should be renamed to `TestNNNObjective`

The test file `main_test.go` can then be edited as required to reflect the test new case. However the hostname used to connect to the `nginx` container ***must not*** be changed from `vugu-nginx`. The `nginx` container will already have been started with this hostname name by `mage` before this test is compiled.

The files `main_wasm.go` and `wasm_exec.js` should not be edited. Likewise the file `index.html.tmpl` should **not** be edited or renamed - the build process depends on the name remaining `index.html.tmpl`.

Any local `index.html` file will be overwritten by the build process, so it can be removed as you wish. The local `.gitignore` file also ensures it is never committed to the repository.

The new test can then be run with the rest of the wasm test suite like this:

```
cd /path/to/vugu
mage testWasm
```

Or individually like this:

```
cd /path/to/vugu
mage testSingleWasmTest github.com/vugu/vugu/wasm-test-suite/test-xxx-objective
```

## Running the `legacy-wasm-test-suite`

The `legacy-wasm-test-suite` uses an entirely different approach to building and testing.

First an external script `run-wasm-test-suite-docker.sh` launches a docker container with headless chrome and a small server

Secondly a standard Go test function i.e. a function named `TestXxx`, runs `vugugen`, builds the `wasm` binary, builds the test, copys the resulting `wasm` binary into the directory that is being served by the web server. 

The standard Go test then queries the DOM to determine of the actions taken by the `wasm` binary are correct.

These tests are built twice. Once using the standard Go compiler and once with the `tinygo` compiler suite, The later is considerably slower. The legacy test suite can therefor take as much as 35 minutes to execute. There is no way to only build for one compiler or the other.

*These tests have been compaltely replaced by teh more flexible approach outlined above. We expect these tests to be removed entirely at some point.*

If you have to run these tests then that can be achieved using `mage` with:

```
cd /path/to/vugu
mage testlegacyWasm
```

This target is **not** run by default. It must be run explicitly or via the `allWithlegacyWasm` target.

The legacy test suite still supports running an individual test, however this does not use `mage`, rather it uses a shell script (not portable to Windows) and the `go test` command. 

To execute a individual legacy wasm test in `legacy-wasm-test-suite`, `run-wasm-test-suite-docker.sh` can be used to launch a docker container with Headless Chrome and a small server.  You can then use `go test` in `wasm-test-suite` to execute the various end-to-end tests available.

This requires Docker to be installed properly and the test suite expects to be able to connect the running docker instance on ports 8846 for web serving/upload and port 9222 to control headless chrome.

## Vugu Documentation

Documention for Vugu, other than GoDoc (https://pkg.go.dev/github.com/vugu/vugu), lives on https://vugu.org.  The source for it is at https://github.com/vugu/vugu-site and you can submit PRs to that repository to propose changes.  For small improvements and errata it is fine to just submit a PR.  For more significant changes or if a discussion is required, please create a GitHub issue first

## Submitting Pull Requests

Please make sure that any PRs:
* Only include necessary modifications.  Avoid unnecessary changes such as reformatting code, renaming existing variables and the like unless previously discussed via issue.
* Do not include files not intended for version control - .DS_Store, compiled binaries, debug output, etc.
* Try hard to not break publicly exposed APIs.  The more likely a change is to cause breakage for other issues, the more vital prior discussion is.  Vugu is not yet 1.0 but compatibility is taken seriously nonetheless.
* Introducing new publicly exposed APIs, as mentioned above, require discussion first.  Depending on the feature, it may be okay to introduce some things as a prototype, but in general before something gets merged into master, both you and the Vugu maintainers should be happy with the design.

## Code of Conduct

### Our Pledge

In the interest of fostering an open and welcoming environment, we as
contributors and maintainers pledge to making participation in our project and
our community a harassment-free experience for everyone, regardless of age, body
size, disability, ethnicity, gender identity and expression, level of experience,
nationality, personal appearance, race, religion, or sexual identity and
orientation.

### Our Standards

Examples of behavior that contributes to creating a positive environment
include:

* Using welcoming and inclusive language
* Being respectful of differing viewpoints and experiences
* Gracefully accepting constructive criticism
* Focusing on what is best for the community
* Showing empathy towards other community members

Examples of unacceptable behavior by participants include:

* The use of sexualized language or imagery and unwelcome sexual attention or
advances
* Trolling, insulting/derogatory comments, and personal or political attacks
* Public or private harassment
* Publishing others' private information, such as a physical or electronic
  address, without explicit permission
* Other conduct which could reasonably be considered inappropriate in a
  professional setting

### Our Responsibilities

Project maintainers are responsible for clarifying the standards of acceptable
behavior and are expected to take appropriate and fair corrective action in
response to any instances of unacceptable behavior.

Project maintainers have the right and responsibility to remove, edit, or
reject comments, commits, code, wiki edits, issues, and other contributions
that are not aligned to this Code of Conduct, or to ban temporarily or
permanently any contributor for other behaviors that they deem inappropriate,
threatening, offensive, or harmful.

### Scope

This Code of Conduct applies both within project spaces and in public spaces
when an individual is representing the project or its community. Examples of
representing a project or community include using an official project e-mail
address, posting via an official social media account, or acting as an appointed
representative at an online or offline event. Representation of a project may be
further defined and clarified by project maintainers.

### Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be
reported by contacting the project team at admin at vugu dot org. All
complaints will be reviewed and investigated and will result in a response that
is deemed necessary and appropriate to the circumstances. The project team is
obligated to maintain confidentiality with regard to the reporter of an incident.
Further details of specific enforcement policies may be posted separately.

Project maintainers who do not follow or enforce the Code of Conduct in good
faith may face temporary or permanent repercussions as determined by other
members of the project's leadership.

### Attribution

This Code of Conduct is adapted from the [Contributor Covenant][homepage], version 1.4,
available at [http://contributor-covenant.org/version/1/4][version]

[homepage]: http://contributor-covenant.org
[version]: http://contributor-covenant.org/version/1/4/
