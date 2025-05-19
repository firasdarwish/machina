package machina

type transition[TState comparable, TTrigger comparable] struct {
	toState TState
	guards  []guard[TState, TTrigger]
}
