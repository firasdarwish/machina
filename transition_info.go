package machina

type TransitionInfo[TState comparable, TTrigger comparable] struct {
	FromState TState
	ToState   TState
	Trigger   TTrigger
}
