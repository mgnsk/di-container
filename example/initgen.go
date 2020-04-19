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
		c.Register(new(greeter), newMyGreeter)
		c.Register(new(mySentence), newMySentence)
		c.Register(new(constants.MyMultiplier), constants.NewMyMultiplier)
		c.Register(new(myService), myServiceProvider)
		c.Register(new(constants.MyInt), constants.NewMyInt)
		c.Register(new(factory), newFactory)
	})
}
