//go:generate initgen

package example

import (
	"github.com/mgnsk/di-container/di"
	"github.com/mgnsk/di-container/example/constants"
)

// Generate registers a container for code generation.
func Generate() {
	di.Generate(func(c *di.Container) {
		c.Register(new(Greeter), NewMyGreeter)
		c.Register(MySentence(""), NewMySentence)
		c.Register(constants.MyMultiplier(0), constants.NewMyMultiplier)
		c.Register(&MyService{}, MyServiceProvider)
		c.Register(constants.MyInt(0), constants.NewMyInt)
	})
}
