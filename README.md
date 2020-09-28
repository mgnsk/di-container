## DI container for Go

Another Go DI container with optional init code generator.

Installation: `go get github.com/mgnsk/di-container/cmd/initgen`

Also needs to have `goimports` installed `go get golang.org/x/tools/cmd/goimports`.

### Example
Given an `initgen.go` it generates an `init.go`.

It is possible to use the container on runtime. See the tests in `container`.

It reports errors when a type is missing a provider function.
```
go generate -x ./... (if you use the go:generate comment) or run initgen manually in the package dir you wish to generate.
initgen
panic: Missing provider for type 'constants.MyInt'
...
```
Types must be registered as pointers to their types. This enables registering interface providers which return any value implementing the interface.

Example shown in `example` dir. In general: install `initgen` binary, create an `initgen.go` file in the package and run `go generate ./...` and 
`go run example/cmd/main.go`

Documentation and design in progress
