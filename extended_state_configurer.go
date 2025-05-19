package machina

type extendedStateConfigurer[TState comparable, TTrigger comparable] interface {
	iStateConfigurer[TState, TTrigger]
	SubstateOf(TState) extendedStateConfigurer[TState, TTrigger]
	OnEntry(func(TransitionInfo[TState, TTrigger])) extendedStateConfigurer[TState, TTrigger]
	OnExit(func(TransitionInfo[TState, TTrigger])) extendedStateConfigurer[TState, TTrigger]
}

func (s *stateConfigurer[TState, TTrigger]) SubstateOf(parentState TState) extendedStateConfigurer[TState, TTrigger] {
	s.m.lock.Lock()
	defer s.m.lock.Unlock()

	s.m.statesConfig[s.state].setParentState(parentState)
	return s
}

func (s *stateConfigurer[TState, TTrigger]) OnEntry(f func(TransitionInfo[TState, TTrigger])) extendedStateConfigurer[TState, TTrigger] {
	s.m.lock.Lock()
	defer s.m.lock.Unlock()

	s.m.statesConfig[s.state].addOnEntry(f)
	return s
}

func (s *stateConfigurer[TState, TTrigger]) OnExit(f func(TransitionInfo[TState, TTrigger])) extendedStateConfigurer[TState, TTrigger] {
	s.m.lock.Lock()
	defer s.m.lock.Unlock()

	s.m.statesConfig[s.state].addOnExit(f)
	return s
}
