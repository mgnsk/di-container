// package initgen generates a container generator for the package in current working dir.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/mgnsk/di-container/di"
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
	pkg := di.GetCurrentPkg()

	// First we generate the temporary generator wrapper which is run later.
	g := generator.NewRoot(
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

	tmp := filepath.Join(filepath.Dir(filename), "tmp")
	os.RemoveAll(tmp)
	check(os.Mkdir(tmp, os.ModePerm))
	defer os.RemoveAll(tmp)

	tmpMain := filepath.Join(".", "tmp", "main.go")
	check(ioutil.WriteFile(tmpMain, []byte(generated), os.ModePerm))

	// Run the container generator.
	res, err := exec.Command("go", "run", tmpMain).CombinedOutput()
	fmt.Println(string(res))
	check(err)
}
