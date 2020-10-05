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
		c.Register(newGreeter)
		c.Register(newMySentence)
		c.Register(constants.NewMyMultiplier)
		c.Register(newMyServiceProvider)
		c.Register(constants.NewMyInt)
		c.Register(newFactory)
	})
}
