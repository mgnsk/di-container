package example

import (
	"fmt"

	"github.com/mgnsk/di-container/example/constants"
)

type MySentence string

// depends on myint.
func NewMySentence(number constants.MyInt, mult constants.MyMultiplier) MySentence {
	return MySentence(fmt.Sprintf("hello world %d!", int(number)*int(mult)))
}

type MyGreeter struct {
	sentence MySentence
}

// depends on MySentence.
func NewMyGreeter(sentence MySentence) (*MyGreeter, error) {
	return &MyGreeter{sentence}, nil
}

func (s *MyGreeter) greet() string {
	return string(s.sentence)
}

type Greeter interface {
	greet() string
}

type MyService struct {
	greeter Greeter
	mult    constants.MyMultiplier
}

type Builder func(*MyService)

// depends on a Greeter interface.
func NewMyService(g Greeter) Builder {
	return func(s *MyService) {
		s.greeter = g
	}
}

// optional dependency.
func (build Builder) WithMultiplier(mult constants.MyMultiplier) Builder {
	return func(s *MyService) {
		build(s)
		s.mult = mult
	}
}

func (build Builder) Build() (*MyService, error) {
	s := &MyService{}
	build(s)
	return s, nil
}

func MyServiceProvider(g Greeter, mult constants.MyMultiplier) (*MyService, error) {
	return NewMyService(g).WithMultiplier(mult).Build()
}

func (s *MyService) Greetings() string {
	return fmt.Sprintf("sentence: %s, mult: %d", s.greeter.greet(), s.mult)
}

func (s *MyService) Close() error {
	return fmt.Errorf("MyService closed")
}
