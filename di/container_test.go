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
	mult    mymultiplier
}

type builder func(*myservice)

// depends on a greeter interface.
func newMyService(g greeter) builder {
	return func(s *myservice) {
		s.greeter = g
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
	return fmt.Sprintf("sentence: %s, mult: %d", s.greeter.greet(), s.mult)
}

func (s *myservice) Close() error {
	return fmt.Errorf("myservice closed")
}

func TestMissingProvider(t *testing.T) {
	c := NewContainer()

	// Interface types must be declared with new().
	c.Register(new(greeter), newMyGreeter)

	// All non-pointer and non-interface types must be declared as a zero value.
	c.Register(mysentence(""), newMySentence)
	c.Register(mymultiplier(0), newMyMultiplier)

	// For builders we create a provider wrapper which explicitly specifies
	// which dependencies (mandatory and optional) we are using.
	c.Register(&myservice{}, func(g greeter, mult mymultiplier) (*myservice, error) {
		return newMyService(g).withMultiplier(mult).build()
	})

	if err := c.Resolve(); err == nil {
		t.Fatal("expected resolve error")
	} else if !strings.Contains(err.Error(), "di.myint") {
		t.Fatal("expected missing di.myint provider")
	}
}

func TestContainer(t *testing.T) {
	c := NewContainer()

	c.Register(new(greeter), newMyGreeter)
	c.Register(mysentence(""), newMySentence)
	c.Register(mymultiplier(0), newMyMultiplier)
	c.Register(&myservice{}, func(g greeter, mult mymultiplier) (*myservice, error) {
		return newMyService(g).withMultiplier(mult).build()
	})

	c.Register(myint(0), newMyInt)

	if err := c.Resolve(); err != nil {
		t.Fatal(err)
	}

	err := c.Build()
	if err != nil {
		t.Fatal(err)
	}

	ig := c.Get(new(greeter)).(greeter)
	if ig.greet() != "hello world 42!" {
		t.Fatal("invalid sentence")
	}

	g := c.Get(&myservice{}).(*myservice)
	if g.greetings() != "sentence: hello world 42!, mult: 2" {
		t.Fatal("invalid greeting")
	}

	g2 := c.Get(&myservice{}).(*myservice)
	if g != g2 {
		t.Fatal("expected singleton pointer")
	}

	if errs := c.Close(); (<-errs).Error() != "myservice closed" {
		t.Fatal("expected myservice closed error")
	}
}
