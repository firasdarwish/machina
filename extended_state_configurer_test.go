package machina_test

import (
	"github.com/firasdarwish/machina"
	"github.com/firasdarwish/machina/internal/testtools/assert2"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type Trigger int
type State int

const (
	Red Trigger = iota
	Orange
	Green
)

const (
	Stopped State = iota
	Forward
	Backward
	Running
	Handbrake
	Neutral
	FullStop
)

func TestSubStateOf_CyclicHierarchy(t *testing.T) {
	s := Stopped
	assert2.PanicsWithError(t, func(s string) bool {
		return strings.Contains(s, "cycle detected")
	}, func() {
		m := machina.New[State, Trigger](s, func(newState State) {
			s = newState
		})

		// Handbrake->Stopped->FullStop->Neutral->Backward->Running->Handbrake(cycle)->Forward

		m.Configure(Handbrake).SubstateOf(Stopped)
		m.Configure(Stopped).SubstateOf(FullStop)
		m.Configure(FullStop).SubstateOf(Neutral)
		m.Configure(Neutral).SubstateOf(Backward)
		m.Configure(Backward).SubstateOf(Running)
		m.Configure(Running).SubstateOf(Handbrake)
		m.Configure(Handbrake).SubstateOf(Forward)
	})
}

func TestSubStateOf_SuperstateAlreadyConfigured(t *testing.T) {
	s := Stopped
	assert.Panics(t, func() {
		m := machina.New[State, Trigger](s, func(newState State) {
			s = newState
		})

		m.Configure(Handbrake).SubstateOf(Stopped).SubstateOf(Running)
	})
}

func TestSubStateOf_SubstateSameAsSuperstate(t *testing.T) {
	s := Stopped
	assert.Panics(t, func() {
		m := machina.New[State, Trigger](s, func(newState State) {
			s = newState
		})

		m.Configure(Handbrake).SubstateOf(Handbrake)
	})
}
