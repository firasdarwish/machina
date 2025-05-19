package machina

import (
	"sync"
)

type IMachine[TState comparable, TTrigger comparable] interface {
	Configure(TState) extendedStateConfigurer[TState, TTrigger]
	CurrentState() TState
	Fire(TTrigger) error
	CanFire(TTrigger) bool
	SetMaxDepth(int)
	SetRollbackOnFailure(bool)
	OnTransitionStarted(func(TransitionInfo[TState, TTrigger]))
	OnTransitionCompleted(func(TransitionInfo[TState, TTrigger]))
}

const DefaultMaxDepth = 10

type machine[TState comparable, TTrigger comparable] struct {
	lock                           *sync.RWMutex
	currentState                   TState
	statesConfig                   map[TState]*stateConfig[TState, TTrigger]
	externalStateSetter            *func(TState)
	maxDepth                       int
	rollbackOnFailure              bool
	onUnhandledTransitionCallback  func(TState, TTrigger) error
	onTransitionStartCallbacks     []func(TransitionInfo[TState, TTrigger])
	onTransitionCompletedCallbacks []func(TransitionInfo[TState, TTrigger])
}

func New[TState comparable, TTrigger comparable](initialState TState, stateSetter func(newState TState)) IMachine[TState, TTrigger] {
	return &machine[TState, TTrigger]{
		lock:                &sync.RWMutex{},
		currentState:        initialState,
		statesConfig:        map[TState]*stateConfig[TState, TTrigger]{},
		externalStateSetter: &stateSetter,
		maxDepth:            DefaultMaxDepth,
		onUnhandledTransitionCallback: func(state TState, trigger TTrigger) error {
			return ErrInvalidTransition
		},
		onTransitionStartCallbacks:     make([]func(TransitionInfo[TState, TTrigger]), 0),
		onTransitionCompletedCallbacks: make([]func(TransitionInfo[TState, TTrigger]), 0),
	}
}

func (m *machine[TState, TTrigger]) CurrentState() TState {
	return m.currentState
}

func (m *machine[TState, TTrigger]) SetMaxDepth(maxDepth int) {
	if maxDepth < 0 {
		panic("max depth must be greater than or equal to 0")
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.maxDepth = maxDepth
}

func (m *machine[TState, TTrigger]) SetRollbackOnFailure(rollbackOnFailure bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.rollbackOnFailure = rollbackOnFailure
}

func (m *machine[TState, TTrigger]) SetOnUnhandledTransition(f func(TState, TTrigger) error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.onUnhandledTransitionCallback = f
}

func (m *machine[TState, TTrigger]) OnTransitionStarted(f func(TransitionInfo[TState, TTrigger])) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.onTransitionStartCallbacks = append(m.onTransitionStartCallbacks, f)
}

func (m *machine[TState, TTrigger]) OnTransitionCompleted(f func(TransitionInfo[TState, TTrigger])) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.onTransitionCompletedCallbacks = append(m.onTransitionCompletedCallbacks, f)
}

func (m *machine[TState, TTrigger]) unsafeSetState(state TState) {
	m.currentState = state
	if m.externalStateSetter != nil {
		(*m.externalStateSetter)(state)
	}
}
