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
* The following folders are being deprecated and should be used with caution: `simplehttp` is an earlier attempt at development tooling use `devutil` instead.  `tinygo-dev` will be removed, the working parts have moved into `devutil`.
* `wasm-test-suite` contains a test suite using Headless Chrome, see Running Tests below.

## Running Tests

In `wasm-test-suite`, `run-wasm-test-suite-docker.sh` can be used to launch a docker container with Headless Chrome and a small server.  You can then use `go test` in `wasm-test-suite` to execute the various end-to-end tests available.

This requires Docker to be installed properly and the test suite expects to be able to connect the running docker instance on ports 8846 for web serving/upload and port 9222 to control Headless Chrome.

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
