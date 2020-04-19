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
		c.Register(mySentence(""), newMySentence)
		c.Register(constants.MyMultiplier(0), constants.NewMyMultiplier)
		c.Register(&myService{}, myServiceProvider)
		c.Register(constants.MyInt(0), constants.NewMyInt)
	})
}
