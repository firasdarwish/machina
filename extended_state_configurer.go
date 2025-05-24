package machina

import (
	"errors"
	"fmt"
)

type extendedStateConfigurer[TState comparable, TTrigger comparable] interface {
	StateConfigurer[TState, TTrigger]
	SubstateOf(TState) extendedStateConfigurer[TState, TTrigger]
	OnEntry(func(TransitionInfo[TState, TTrigger])) extendedStateConfigurer[TState, TTrigger]
	OnExit(func(TransitionInfo[TState, TTrigger])) extendedStateConfigurer[TState, TTrigger]
}

func (s *stateConfigurer[TState, TTrigger]) SubstateOf(parentState TState) extendedStateConfigurer[TState, TTrigger] {
	if parentState == s.state {
		panic(ErrCyclicSubSuperState)
	}

	if s.m.statesConfig[s.state].parentState != nil {
		panic(ErrSuperstateAlreadyConfigured)
	}

	s.m.lock.Lock()
	defer s.m.lock.Unlock()

	var seenStates = []TState{s.state}

	loopState := parentState
	depth := 0
	for {
		if depth > s.m.maxDepth {
			panic(ErrMaxDepthReached)
		}
		depth += 1

		if stateExistsInList(seenStates, loopState) {
			seenStates = append(seenStates, loopState)
			panic(errors.New(fmt.Sprintf("cycle detected in state `%v` configuration; `%v`", s.state, seenStates)))
		}

		seenStates = append(seenStates, loopState)

		pState, parentStateExists := s.m.statesConfig[loopState]
		if !parentStateExists || pState.parentState == nil {
			break
		}

		loopState = *pState.parentState

	}

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

func stateExistsInList[TState comparable](states []TState, state TState) bool {
	for _, s := range states {
		if state == s {
			return true
		}
	}

	return false
}
