// package initgen generates a container generator for the package in current working dir.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/mgnsk/di-container/initgen"
	"github.com/moznion/gowrtr/generator"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	filename, err := filepath.Abs(filepath.Join(".", "initgen.go"))
	check(err)

	_, err = os.Stat(filename)
	check(err)

	// Package which is being generated.
	pkg := initgen.GetCurrentPkg()

	// First we generate the temporary generator wrapper which is run later.
	g := generator.NewRoot(
		generator.NewComment(" DO NOT EDIT. This code is generated by initgen."),
		generator.NewPackage("main"),
		generator.NewImport(pkg),
		generator.NewNewline(),
	)

	// Add a main function to run the Generate function provided by pkg.
	g = g.AddStatements(
		generator.NewFunc(
			nil,
			generator.NewFuncSignature("main"),
		).AddStatements(generator.NewRawStatement(path.Base(pkg) + ".Generate()")),
	).
		Gofmt("-s").
		Goimports()

	generated, err := g.Generate(0)
	check(err)

	dir := filepath.Join(filepath.Dir(filename), "tmp")
	os.RemoveAll(dir)
	check(os.Mkdir(dir, os.ModePerm))
	defer os.RemoveAll(dir)

	tmpFile := filepath.Join(".", "tmp", "initgen_build.go")
	check(ioutil.WriteFile(tmpFile, []byte(generated), os.ModePerm))

	// Run the container generator.
	res, err := exec.Command("go", "run", tmpFile).CombinedOutput()
	fmt.Println(string(res))
	check(err)
}
