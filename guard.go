package machina

type guard[TState comparable, TTrigger comparable] func(info TransitionInfo[TState, TTrigger]) error
