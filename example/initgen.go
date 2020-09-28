//go:generate initgen

package example

import (
	"github.com/mgnsk/di-container/di"
	"github.com/mgnsk/di-container/example/constants"
	"github.com/mgnsk/di-container/initgen"
)

// Generate registers a container for code generation.
func Generate() {
	initgen.Generate(func(c *di.Container) {
		// pointer to an interface
		c.Register((*greeter)(nil), newMyGreeter)
		c.Register((*mySentence)(nil), newMySentence)
		c.Register((*constants.MyMultiplier)(nil), constants.NewMyMultiplier)
		c.Register((**MyService)(nil), myServiceProvider)
		c.Register((*constants.MyInt)(nil), constants.NewMyInt)
		c.Register((*factory)(nil), newFactory)
	})
}
