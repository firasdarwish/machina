package machina

type transitionInfo[TState comparable, TTrigger comparable] struct {
	fromState TState
	toState   TState
	trigger   TTrigger
	params    []any
}

// Source returns the source state
func (t *transitionInfo[TState, TTrigger]) Source() TState {
	return t.fromState
}

// Destination returns the (possible/confirmed) destination state
func (t *transitionInfo[TState, TTrigger]) Destination() TState {
	return t.toState
}

// Trigger returns the fired trigger
func (t *transitionInfo[TState, TTrigger]) Trigger() TTrigger {
	return t.trigger
}

// Params returns a list of custom params passed with fired trigger
func (t *transitionInfo[TState, TTrigger]) Params() []any {
	return t.params
}

// Transition represents a materialized transition info
type Transition[TState comparable, TTrigger comparable] interface {
	Source() TState
	Destination() TState
	Trigger() TTrigger
	Params() []any
}
