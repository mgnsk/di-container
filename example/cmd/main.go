package main

import (
	"fmt"
	"log"

	"github.com/mgnsk/di-container/example"
)

func main() {
	s, err := example.InitMyService()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s.Greetings())
}
