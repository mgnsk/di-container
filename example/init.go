// DO NOT EDIT. This code is generated by initgen.
package example

import (
	"github.com/mgnsk/di-container/example/constants"
)

func InitMyInt() constants.MyInt {
	myint := constants.NewMyInt()
	return myint
}

func InitMyMultiplier() constants.MyMultiplier {
	mymultiplier := constants.NewMyMultiplier()
	return mymultiplier
}

func InitMySentence() mySentence {
	myint := InitMyInt()
	mymultiplier := InitMyMultiplier()
	mysentence := newMySentence(myint, mymultiplier)
	return mysentence
}

func InitGreeter() greeter {
	mysentence := InitMySentence()
	greeter, err := newMyGreeter(mysentence)
	if err != nil {
		panic(err)
	}
	return greeter
}

func InitMyService() *myService {
	greeter := InitGreeter()
	mymultiplier := InitMyMultiplier()
	myservice, err := myServiceProvider(greeter, mymultiplier)
	if err != nil {
		panic(err)
	}
	return myservice
}
