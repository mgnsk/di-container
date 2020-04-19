package initgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/mgnsk/di-container/di"
	"github.com/moznion/gowrtr/generator"
	"golang.org/x/tools/go/ast/astutil"
)

// GetCurrentPkg returns the go pkg in the working dir.
func GetCurrentPkg() string {
	pkgImport, err := exec.Command("go", "list", "-f", "{{.ImportPath}}", ".").Output()
	check(err)

	return string(bytes.TrimSpace(pkgImport))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type dep struct {
	localType       string
	rawType         string
	isPointer       bool
	deps            []string
	providerReturns int
	newFunc         string
}

// Generate code for type initializers in the context of the resolved container.
func Generate(register func(*di.Container)) (dummy struct{}) {
	c := di.NewContainer()
	register(c)
	check(c.Resolve())

	curPkg := path.Base(GetCurrentPkg())

	trimCurrentPkgPrefix := func(s string) string {
		s = strings.TrimPrefix(s, "*")
		s = strings.TrimPrefix(s, curPkg+".")
		return s
	}

	trimPkgPrefix := func(s string) string {
		safeName := strings.Split(s, ".")
		name := safeName[len(safeName)-1]
		return name
	}

	// Collect type dependencies.
	typeDeps := make(map[string]*dep)
	var order []*dep
	c.Range(func(item *di.Item) {
		d := &dep{
			localType:       trimCurrentPkgPrefix(item.Typ.String()),
			rawType:         trimPkgPrefix(item.Typ.String()),
			isPointer:       item.IsPointer,
			deps:            make([]string, len(item.Node.Edges)),
			providerReturns: item.Provider.Type().NumOut(),
		}
		for i, n := range item.Node.Edges {
			depItem := n.Value.(*di.Item)
			d.deps[i] = trimCurrentPkgPrefix(depItem.Typ.String())
		}
		order = append(order, d)
		typeDeps[d.localType] = d
	})

	// Parse functions from source code.
	for typeName, fName := range parseDefaultNewFuncs() {
		for _, d := range order {
			if d.rawType == typeName {
				d.newFunc = fName
			}
		}
	}

	g := generator.NewRoot(
		generator.NewComment(" DO NOT EDIT. This code is generated by initgen."),
		generator.NewPackage(curPkg),
		//	generator.NewImport(pkg),
		// TODO it seems the auto imports work.
		generator.NewNewline(),
	)

	for _, d := range order {
		sig := generator.NewFuncSignature("init" + d.rawType)

		var returnType string
		if d.isPointer {
			returnType = "*" + d.localType
		} else {
			returnType = d.localType
		}

		sig = sig.AddReturnTypes(returnType)
		initFunc := generator.NewFunc(nil, sig)

		// Collect arguments for type provider function.
		var providerArgs []string
		for _, depName := range typeDeps[d.localType].deps {
			varName := strings.ToLower(trimPkgPrefix(depName))
			providerArgs = append(providerArgs, varName)
			initFunc = initFunc.AddStatements(
				generator.NewRawStatement(fmt.Sprintf(
					"%s := init%s()",
					varName,
					trimPkgPrefix(depName),
				)),
			)
		}

		varName := strings.ToLower(d.rawType)

		argString := strings.Join(providerArgs, ", ")

		if d.providerReturns == 2 {
			// Returns a value + error.
			initFunc = initFunc.AddStatements(
				generator.NewRawStatement(fmt.Sprintf(
					"%s, err := %s(%s)",
					varName,
					d.newFunc,
					argString,
				)),
				generator.NewRawStatement(`if err != nil { panic(err) }`),
			)
		} else {
			// Returns only value.
			initFunc = initFunc.AddStatements(
				generator.NewRawStatement(fmt.Sprintf(
					"%s := %s(%s)",
					varName,
					d.newFunc,
					argString,
				)),
			)
		}

		initFunc = initFunc.AddStatements(
			generator.NewRawStatement("return " + varName),
		)

		g = g.AddStatements(initFunc, generator.NewNewline())
	}

	g = g.Gofmt("-s").Goimports()

	generated, err := g.Generate(0)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(filepath.Join(".", "init.go"), []byte(generated), os.ModePerm); err != nil {
		panic(err)
	}

	return
}

func parseDefaultNewFuncs() map[string]string {
	filename, err := filepath.Abs(filepath.Join(".", "initgen.go"))
	if err != nil {
		panic(err)
	}

	fset, node := parseFile(filename)

	containers := parsecontainers(node)
	if len(containers) != 1 {
		panic("container missing or invalid number of")
	}

	registers := parseRegister(containers[0])
	_ = registers

	lmap := parseNewFuncs(fset, registers)

	return lmap
}

type container struct {
	f *ast.CallExpr
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

func parsecontainers(f *ast.File) []container {
	var containers []container
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		switch t := c.Node().(type) {
		case *ast.CallExpr:
			if isFunction(t, "initgen", "Generate") {
				containers = append(containers, container{
					f: t,
				})
				return false
			}
		}
		return true
	}, nil)
	return containers
}

type registerCall struct {
	f *ast.CallExpr
}

func parseRegister(container container) []registerCall {
	var calls []registerCall
	astutil.Apply(container.f, func(c *astutil.Cursor) bool {
		switch t := c.Node().(type) {
		case *ast.CallExpr:
			if isFunction(t, "c", "Register") {
				calls = append(calls, registerCall{
					f: t,
				})
				return false
			}
		}
		return true
	}, nil)
	return calls
}

func parseNewFuncs(fset *token.FileSet, registers []registerCall) map[string]string {
	funcs := make(map[string]string)

	for _, r := range registers {
		if len(r.f.Args) != 2 {
			panic("Invalid number of Register args")
		}

		// Get the innermost ident name.
		typeIdent := getInnermostIdent(r.f.Args[0])

		safeType := strings.Split(typeIdent, ".")
		typeIdent = safeType[len(safeType)-1]

		// Provider func ident.
		providerIdent := getInnermostIdent(r.f.Args[1])

		funcs[typeIdent] = providerIdent
	}

	return funcs
}

func parseFile(filename string) (*token.FileSet, *ast.File) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	check(err)

	return fset, node
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
