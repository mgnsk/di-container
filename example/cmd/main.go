package main

import (
	"fmt"

	"github.com/mgnsk/di-container/example"
)

func main() {
	s := example.InitMyService()
	fmt.Println(s.Greetings())
}
