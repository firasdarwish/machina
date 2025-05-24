package main

import (
	"github.com/firasdarwish/machina"
	"log"
)

type State string
type Trigger string

const (
	On  State = "on"
	Off State = "off"
)

const (
	Activate   Trigger = "activate"
	Deactivate Trigger = "deactivate"
)

var currentState State = Off

func main() {
	m := machina.New[State, Trigger](currentState, func(newState State) {
		currentState = newState
	})

	m.Configure(On).PermitIf(Deactivate, Off, func(transition machina.Transition[State, Trigger]) error {
		return nil
	})
	m.Configure(Off).OnEntry(func(t machina.Transition[State, Trigger]) {
		log.Println(t.Params())
	}).Permit(Activate, On)

	log.Println(currentState)
	err := m.Fire(Activate)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(currentState)
	err = m.Fire(Deactivate)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(currentState)
}
