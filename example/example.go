package example

import (
	"fmt"

	"github.com/mgnsk/di-container/example/constants"
)

type mySentence string

// depends on myint.
func newMySentence(number constants.MyInt, mult constants.MyMultiplier) mySentence {
	return mySentence(fmt.Sprintf("hello world %d!", int(number)*int(mult)))
}

type mygreeter struct {
	sentence mySentence
}

// depends on mySentence.
func newMyGreeter(sentence mySentence) (mygreeter, error) {
	return mygreeter{sentence}, nil
}

func (s mygreeter) greet() string {
	return string(s.sentence)
}

type greeter interface {
	greet() string
}

type MyService struct {
	greeter greeter
	f       factory
	mult    constants.MyMultiplier
}

type builder func(*MyService)

type factory func() string

func newFactory() factory {
	return func() string {
		return "test"
	}
}

// depends on a greeter interface.
func newMyService(g greeter, f factory) builder {
	return func(s *MyService) {
		s.greeter = g
		s.f = f
	}
}

// optional dependency.
func (build builder) withMultiplier(mult constants.MyMultiplier) builder {
	return func(s *MyService) {
		build(s)
		s.mult = mult
	}
}

func (build builder) build() (MyService, error) {
	s := &MyService{}
	build(s)
	return *s, nil
}

func myServiceProvider(g greeter, f factory, mult constants.MyMultiplier) (MyService, error) {
	return newMyService(g, f).withMultiplier(mult).build()
}

func (s MyService) Greetings() string {
	return fmt.Sprintf("sentence: %s, mult: %d", s.greeter.greet(), s.mult)
}

func (s MyService) Close() error {
	return fmt.Errorf("myService closed")
}
