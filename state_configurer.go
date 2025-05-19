package machina

import "errors"

type iStateConfigurer[TState comparable, TTrigger comparable] interface {
	Permit(TTrigger, TState) iStateConfigurer[TState, TTrigger]
	PermitIf(TTrigger, TState, ...guard[TState, TTrigger]) iStateConfigurer[TState, TTrigger]
}

type stateConfigurer[TState comparable, TTrigger comparable] struct {
	state TState
	m     *machine[TState, TTrigger]
}

func (m *machine[TState, TTrigger]) Configure(state TState) extendedStateConfigurer[TState, TTrigger] {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, exists := m.statesConfig[state]
	if exists {
		panic(ErrStateAlreadyConfigured)
	}

	m.statesConfig[state] = &stateConfig[TState, TTrigger]{
		transitions: make(map[TTrigger][]transition[TState, TTrigger]),
	}

	sc := &stateConfigurer[TState, TTrigger]{
		m:     m,
		state: state,
	}

	return sc
}

func (s *stateConfigurer[TState, TTrigger]) Permit(trigger TTrigger, state TState) iStateConfigurer[TState, TTrigger] {
	s.m.lock.Lock()
	defer s.m.lock.Unlock()

	s.m.statesConfig[s.state].addTransition(trigger, transition[TState, TTrigger]{
		toState: state,
	})

	return s
}

func (s *stateConfigurer[TState, TTrigger]) PermitIf(trigger TTrigger, state TState, guards ...guard[TState, TTrigger]) iStateConfigurer[TState, TTrigger] {
	s.m.lock.Lock()
	defer s.m.lock.Unlock()

	if guards == nil || len(guards) == 0 {
		panic(errors.Join(errors.New("PermitIf"), ErrEmptyGuards))
	}

	s.m.statesConfig[s.state].addTransition(trigger, transition[TState, TTrigger]{
		toState: state,
		guards:  guards,
	})

	return s
}
