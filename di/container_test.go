package di

import (
	"fmt"
	"strings"
	"testing"
)

type myint int

// does not have any dependencies.
func newMyInt() myint {
	return 21
}

type mymultiplier int

func newMyMultiplier() mymultiplier {
	return 2
}

type mysentence string

// depends on myint.
func newMySentence(number myint, mult mymultiplier) mysentence {
	return mysentence(fmt.Sprintf("hello world %d!", int(number)*int(mult)))
}

type mygreeter struct {
	sentence mysentence
}

// depends on mysentence.
func newMyGreeter(sentence mysentence) (*mygreeter, error) {
	return &mygreeter{sentence}, nil
}

func (s *mygreeter) greet() string {
	return string(s.sentence)
}

type greeter interface {
	greet() string
}

type myservice struct {
	greeter greeter
	f       factory
	mult    mymultiplier
}

type builder func(*myservice)

type factory func() string

func newFactory() factory {
	return func() string {
		return "test"
	}
}

// depends on a greeter interface.
func newMyService(g greeter, f factory) builder {
	return func(s *myservice) {
		s.greeter = g
		s.f = f
	}
}

// optional dependency.
func (build builder) withMultiplier(mult mymultiplier) builder {
	return func(s *myservice) {
		build(s)
		s.mult = mult
	}
}

func (build builder) build() (*myservice, error) {
	s := &myservice{}
	build(s)
	return s, nil
}

func (s *myservice) greetings() string {
	return fmt.Sprintf("sentence: %s, mult: %d, factory: %s", s.greeter.greet(), s.mult, s.f())
}

func TestMissingProvider(t *testing.T) {
	c := NewContainer()

	// Interface types must be declared with new().
	c.Register((*greeter)(nil), newMyGreeter)

	// All non-pointer and non-interface types must be declared as a zero value.
	c.Register((*mysentence)(nil), newMySentence)
	c.Register((*mymultiplier)(nil), newMyMultiplier)

	c.Register((*factory)(nil), newFactory)

	// For builders we create a provider wrapper which explicitly specifies
	// which dependencies (mandatory and optional) we are using.
	c.Register((**myservice)(nil), func(g greeter, f factory, mult mymultiplier) (*myservice, error) {
		return newMyService(g, f).withMultiplier(mult).build()
	})

	if err := c.Resolve(); err == nil {
		t.Fatal("expected resolve error")
	} else if !strings.Contains(err.Error(), "di.myint") {
		t.Fatal("expected missing di.myint provider")
	}
}

func TestDependencyLoop(t *testing.T) {
	c := NewContainer()

	intProvider := func(b []byte) int {
		return 0
	}

	stringProvider := func(i int) string {
		return fmt.Sprintf("%d", i)
	}

	bytesProvider := func(s string) []byte {
		return []byte(s)
	}

	c.Register((*int)(nil), intProvider)
	c.Register((*[]byte)(nil), bytesProvider)
	c.Register((*string)(nil), stringProvider)

	if err := c.Resolve(); err == nil {
		t.Fatal("expected dependency loop error")
	}
}

func TestContainer(t *testing.T) {
	c := NewContainer()

	c.Register((*greeter)(nil), newMyGreeter)
	c.Register((*mysentence)(nil), newMySentence)
	c.Register((*mymultiplier)(nil), newMyMultiplier)
	c.Register((**myservice)(nil), func(g greeter, f factory, mult mymultiplier) (*myservice, error) {
		return newMyService(g, f).withMultiplier(mult).build()
	})

	c.Register((*myint)(nil), newMyInt)
	c.Register((*factory)(nil), newFactory)

	if err := c.Resolve(); err != nil {
		t.Fatal(err)
	}

	err := c.Build()
	if err != nil {
		t.Fatal(err)
	}

	ig := c.Get((*greeter)(nil)).(greeter)
	if ig.greet() != "hello world 42!" {
		t.Fatal("invalid sentence")
	}

	g := c.Get((**myservice)(nil)).(*myservice)
	if g.greetings() != "sentence: hello world 42!, mult: 2, factory: test" {
		t.Fatal("invalid greeting")
	}
}
