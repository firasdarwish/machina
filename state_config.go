package machina

type stateConfig[TState comparable, TTrigger comparable] struct {
	parentState *TState
	transitions map[TTrigger][]transition[TState, TTrigger]
	onEntries   []func(TransitionInfo[TState, TTrigger])
	onExits     []func(TransitionInfo[TState, TTrigger])
}

func (sc *stateConfig[TState, TTrigger]) addTransition(trigger TTrigger, t transition[TState, TTrigger]) {
	_, exists := sc.transitions[trigger]
	if !exists {
		sc.transitions[trigger] = []transition[TState, TTrigger]{}
	}

	exists = sc.transitionExists(trigger, t)
	if exists {
		panic(ErrTransitionDuplicated)
	}

	sc.transitions[trigger] = append(sc.transitions[trigger], t)
}

func (sc *stateConfig[TState, TTrigger]) addOnEntry(ti func(TransitionInfo[TState, TTrigger])) {
	sc.onEntries = append(sc.onEntries, ti)
}

func (sc *stateConfig[TState, TTrigger]) addOnExit(ti func(TransitionInfo[TState, TTrigger])) {
	sc.onExits = append(sc.onExits, ti)
}

func (sc *stateConfig[TState, TTrigger]) setParentState(parentState TState) {
	sc.parentState = &parentState
}

func (sc *stateConfig[TState, TTrigger]) transitionExists(trigger TTrigger, t transition[TState, TTrigger]) bool {
	for _, t2 := range sc.transitions[trigger] {
		if t2.toState == t.toState {
			return true
		}
	}

	return false
}
