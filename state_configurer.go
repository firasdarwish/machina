package machina

import "errors"

type StateConfigurer[TState comparable, TTrigger comparable] interface {
	// Permit allows the state machine to transition to `dest` state when invoked using `trigger`
	Permit(trigger TTrigger, dest TState) StateConfigurer[TState, TTrigger]

	// PermitIf allows the state machine to transition to `dest` state when invoked using `trigger` and when all guards are met
	PermitIf(trigger TTrigger, dest TState, guards ...guard[TState, TTrigger]) StateConfigurer[TState, TTrigger]
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

func (s *stateConfigurer[TState, TTrigger]) Permit(trigger TTrigger, state TState) StateConfigurer[TState, TTrigger] {
	s.m.lock.Lock()
	defer s.m.lock.Unlock()

	s.m.statesConfig[s.state].addTransition(trigger, transition[TState, TTrigger]{
		toState: state,
	})

	return s
}

func (s *stateConfigurer[TState, TTrigger]) PermitIf(trigger TTrigger, state TState, guards ...guard[TState, TTrigger]) StateConfigurer[TState, TTrigger] {
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
