## DI container for Go

Compile-time dependency injection and a runtime singleton container for go. Inspired by [Wire](https://github.com/google/wire) and [Symfony Service Container](https://symfony.com/doc/current/service_container.html)

Installation: `go get github.com/mgnsk/di-container/cmd/initgen`

Also needs to have `goimports` installed `go get golang.org/x/tools/cmd/goimports`. [TODO]

### Example
* `$ cd example`
* `$ go generate`
* Run the example app using the initializers: `$ go run cmd/main.go`

It is also possible to use the container dynamically on runtime. In that case it acts like a singleton container.
