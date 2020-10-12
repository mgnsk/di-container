## DI container for Go

A simpler compile-time dependency injection and a runtime singleton container. Inspired by [Wire](https://github.com/google/wire) and [Symfony Service Container](https://symfony.com/doc/current/service_container.html)

Installation: `go get github.com/mgnsk/di-container/cmd/initgen`

Also needs to have `goimports` installed `go get golang.org/x/tools/cmd/goimports`. [TODO]

### Usage
* Create an `initgen.go` file in the package you wish to generate. It must contain the registration of the provider functions for your types/interfaces.
* Run `initgen`.
* Use the generated `init.go` file.

It is also possible to use the container dynamically on runtime. In that case it acts like a singleton container. See the tests in `container`.

### Example
Example shown in `example` dir.

Documentation and design in progress
