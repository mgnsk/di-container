// package initgen generates initializers for provider functions registered in the current working dir package.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/moznion/gowrtr/generator"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getCurrentPkg() string {
	pkgImport, err := exec.Command("go", "list", "-f", "{{.ImportPath}}", ".").Output()
	check(err)

	return string(bytes.TrimSpace(pkgImport))
}

func main() {
	cwd, err := os.Getwd()
	check(err)

	source := filepath.Join(cwd, "initgen.go")
	target := filepath.Join(cwd, "init.go")
	tmpDir := filepath.Join(cwd, "initgen")

	_, err = os.Stat(source)
	check(err)

	check(os.RemoveAll(tmpDir))
	check(os.RemoveAll(target))
	check(os.Mkdir(tmpDir, 0o755))
	defer os.RemoveAll(tmpDir)

	fmt.Printf("initgen: generating %s\n", target)

	pkg := getCurrentPkg()
	g := generator.
		NewRoot(
			generator.NewPackage("main"),
			generator.NewNewline(),
			generator.NewImport(pkg),
			generator.NewNewline(),
		).
		AddStatements(
			generator.NewFunc(
				nil,
				generator.NewFuncSignature("main"),
			).AddStatements(generator.NewRawStatement(path.Base(pkg) + ".Generate()")),
		)

	generated, err := g.Generate(0)
	check(err)

	mainFile := filepath.Join(tmpDir, "main.go")
	check(ioutil.WriteFile(mainFile, []byte(generated), 0o644))

	// Run the container generator.
	res, err := exec.Command("go", "run", mainFile).CombinedOutput()
	fmt.Printf(string(res))
	check(err)
}
