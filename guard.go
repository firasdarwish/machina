package machina

type guard[TState comparable, TTrigger comparable] func(transition Transition[TState, TTrigger]) error
