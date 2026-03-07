# Fantastic coffee (decaffeinated)

This repository contains the basic structure for [Web and Software Architecture](http://gamificationlab.uniroma1.it/en/wasa/) homework project.
It has been described in class.

WASAText

WASAText is a backend web service written in Go that exposes a RESTful API for managing text-based resources.
The project is built starting from a provided template and has been adapted, configured, and extended to meet the project requirements.

This repository contains the server-side implementation, API definitions, and supporting services.

📌 Project Goals

The main goals of this project are:

To design and implement a structured REST API in Go

To correctly manage Go modules and project structure

To connect and interact with a database or external resources

To document APIs using an OpenAPI (YAML) specification

To demonstrate correct usage of a provided project template

## Project structure

* `cmd/` contains all executables; Go programs here should only do "executable-stuff", like reading options from the CLI/env, etc.
	* `cmd/healthcheck` is an example of a daemon for checking the health of servers daemons; useful when the hypervisor is not providing HTTP readiness/liveness probes (e.g., Docker engine)
	* `cmd/webapi` contains an example of a web API server daemon
* `demo/` contains a demo config file
* `doc/` contains the documentation (usually, for APIs, this means an OpenAPI file)
* `service/` has all packages for implementing project-specific functionalities
	* `service/api` contains an example of an API server
	* `service/globaltime` contains a wrapper package for `time.Time` (useful in unit testing)
* `vendor/` is managed by Go, and contains a copy of all dependencies
* `webui/` is an example of a web frontend in Vue.js; it includes:
	* Bootstrap JavaScript framework
	* a customized version of "Bootstrap dashboard" template
	* feather icons as SVG
	* Go code for release embedding

Other project files include:
* `open-node.sh` starts a new (temporary) container using `node:20` image for safe and secure web frontend development (you don't want to use `node` in your system, do you?).

## Go vendoring

This project uses [Go Vendoring](https://go.dev/ref/mod#vendoring). You must use `go mod vendor` after changing some dependency (`go get` or `go mod tidy`) and add all files under `vendor/` directory in your commit.

For more information about vendoring:

* https://go.dev/ref/mod#vendoring
* https://www.ardanlabs.com/blog/2020/04/modules-06-vendoring.html

## Node/YARN vendoring

This repository uses `yarn` and a vendoring technique that exploits the ["Offline mirror"](https://yarnpkg.com/features/caching). As for the Go vendoring, the dependencies are inside the repository.

You should commit the files inside the `.yarn` directory.


## License

See [LICENSE](LICENSE).


## HOW TO RUN
This Code is the docker version of assignment 1+2+3 so it can be run independently
The other commit of HW1+HW2+HW3+FrotendRunned (commit id - 3606be8) does not have docker version of code. So this commit cannot be runned with docker commands. It should run independently to check assignment 1+2+3 without docker.
Assignment 1+2+3 is not a docker version of assignment.
Assignment 4 includes assignment 1+2+3 with docker version.
TO check assignment 1+2+3, we do not expect to run on docker.
To check assignment 4 we expect the docker version of assignment 1+2+3.