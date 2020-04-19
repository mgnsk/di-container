package di

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/mgnsk/di-container/initgen"
)

// Generate code for type initializers in the context of the resolved container.
func Generate(register func(*Container)) (dummy struct{}) {
	c := NewContainer()
	register(c)
	if err := c.Resolve(); err != nil {
		panic(err)
	}

	curPkg := path.Base(initgen.GetCurrentPackage())
	trim := func(s string) string {
		s = strings.TrimPrefix(s, "*")
		s = strings.TrimPrefix(s, curPkg+".")
		return s
	}

	// Collect type dependencies.
	deps := make(map[string][]string)
	var order []string
	c.Range(func(item *Item) {
		tp := trim(item.typ.String())
		order = append(order, tp)
		var itemDeps []string
		for _, n := range item.n.Edges {
			itemDeps = append(itemDeps, trim(n.Value.(*Item).typ.String()))
		}
		deps[tp] = itemDeps
	})

	spew.Dump(order)
	spew.Dump(deps)

	// Parse functions from source code.
	labels := parseLabels()

	spew.Dump(labels)

	// take positions of texts to cut???
	//	spew.Dump(containers)

	return
}

func parseLabels() map[string]string {
	filename, err := filepath.Abs(filepath.Join(".", "initgen.go"))
	if err != nil {
		panic(err)
	}

	fset, node := initgen.ParseFile(filename)

	containers := initgen.ParseContainers(node)
	if len(containers) != 1 {
		panic("Container missing or invalid number of")
	}

	registers := initgen.ParseRegister(containers[0])
	_ = registers

	lmap := initgen.ParseLabels(fset, registers)

	return lmap
}
