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

	m.Configure(On).Permit(Deactivate, Off)
	m.Configure(Off).OnEntry(func(t machina.TransitionInfo[State, Trigger]) {
		log.Println(t.Params)
	}).Permit(Activate, On)

	log.Println(currentState)
	m.Fire(Activate)
	log.Println(currentState)
	m.Fire(Deactivate)
	log.Println(currentState)
}
