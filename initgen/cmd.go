package initgen

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"

	"golang.org/x/tools/go/ast/astutil"
)

type Container struct {
	GenFunc *ast.CallExpr
}

func isFunction(t *ast.CallExpr, pkg, fn string) bool {
	if fun, ok := t.Fun.(*ast.SelectorExpr); ok {
		if x, ok := fun.X.(*ast.Ident); ok &&
			x.Name == pkg &&
			fun.Sel.Name == fn {
			return true
		}
	}
	return false
}

func ParseContainers(f *ast.File) []Container {
	var containers []Container
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		switch t := c.Node().(type) {
		case *ast.CallExpr:
			if isFunction(t, "di", "Generate") {
				containers = append(containers, Container{
					GenFunc: t,
				})
				return false
			}
		}
		return true
	}, nil)
	return containers
}

type RegisterCall struct {
	RegisterFunc *ast.CallExpr
}

func ParseRegister(container Container) []RegisterCall {
	var calls []RegisterCall
	astutil.Apply(container.GenFunc, func(c *astutil.Cursor) bool {
		switch t := c.Node().(type) {
		case *ast.CallExpr:
			if isFunction(t, "c", "Register") {
				calls = append(calls, RegisterCall{
					RegisterFunc: t,
				})
				return false
			}
		}
		return true
	}, nil)
	return calls
}

func ParseLabels(fset *token.FileSet, registers []RegisterCall) map[string]string {
	labels := make(map[string]string)

	for _, r := range registers {
		if len(r.RegisterFunc.Args) != 2 {
			panic("Invalid number of Register args")
		}

		// Get the innermost ident name.
		typeIdent := getInnermostIdent(r.RegisterFunc.Args[0])

		// Provider func ident.
		providerIdent := getInnermostIdent(r.RegisterFunc.Args[1])

		labels[typeIdent] = providerIdent
	}

	return labels
}

func ParseFile(filename string) (*token.FileSet, *ast.File) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	check(err)

	return fset, node
}

func GetCurrentPackage() string {
	pkgImport, err := exec.Command("go", "list", "-f", "{{.ImportPath}}", ".").Output()
	check(err)

	return string(bytes.TrimSpace(pkgImport))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func getInnermostIdent(node ast.Node) string {
	var ident string
	astutil.Apply(node, func(c *astutil.Cursor) bool {
		switch t := c.Node().(type) {
		case *ast.Ident:
			ident = t.Name
			if se, ok := c.Parent().(*ast.SelectorExpr); ok {
				if x, ok := se.X.(*ast.Ident); ok {
					ident = x.Name + "." + t.Name
				}
			}
		}
		return true
	}, nil)
	return ident
}
