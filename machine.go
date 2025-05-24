package machina

import (
	"sync"
)

type StateMachine[TState comparable, TTrigger comparable] interface {
	// Configure Configures triggers, transitions, entry/exit callbacks on `state`
	Configure(state TState) extendedStateConfigurer[TState, TTrigger]

	// CurrentState gets the current internal state of the state machine object
	CurrentState() TState

	// Fire Invokes a `trigger` on the state machine, optionally with a list of custom parameters
	Fire(trigger TTrigger, params ...any) error

	// CanFire returns a boolean to indicate if the state machine currently accepts invocation using `trigger` with a list of parameters
	CanFire(trigger TTrigger, params ...any) bool

	// SetMaxDepth setts the maximum depth of superstate-substate hierarchy
	SetMaxDepth(int)

	// SetRollbackOnFailure when set to `true` and when a panic occurs on `OnTransitionCompleted` or `OnEntry` callbacks, it tries to reset both
	// the state machine`s internal state & the external state setter (if exists) back to its original state
	SetRollbackOnFailure(bool)

	// OnTransitionStarted registers a global callback that runs at the start of the transition
	// can use multiple times to register multiple callbacks
	OnTransitionStarted(func(Transition[TState, TTrigger]))

	// OnTransitionCompleted registers a global callback that runs at the end of the transition
	// can use multiple times to register multiple callbacks
	OnTransitionCompleted(func(Transition[TState, TTrigger]))

	// SetOnUnhandledTransition registers a global callback that runs when no triggers are configured to current state
	SetOnUnhandledTransition(f func(TState, TTrigger) error)
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
	onTransitionStartCallbacks     []func(Transition[TState, TTrigger])
	onTransitionCompletedCallbacks []func(Transition[TState, TTrigger])
}

// New Instantiates a new state machine object
func New[TState comparable, TTrigger comparable](initialState TState, stateSetter func(newState TState)) StateMachine[TState, TTrigger] {
	return &machine[TState, TTrigger]{
		lock:                &sync.RWMutex{},
		currentState:        initialState,
		statesConfig:        map[TState]*stateConfig[TState, TTrigger]{},
		externalStateSetter: &stateSetter,
		maxDepth:            DefaultMaxDepth,
		onUnhandledTransitionCallback: func(state TState, trigger TTrigger) error {
			return ErrInvalidTransition
		},
		onTransitionStartCallbacks:     make([]func(Transition[TState, TTrigger]), 0),
		onTransitionCompletedCallbacks: make([]func(Transition[TState, TTrigger]), 0),
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

func (m *machine[TState, TTrigger]) OnTransitionStarted(f func(Transition[TState, TTrigger])) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.onTransitionStartCallbacks = append(m.onTransitionStartCallbacks, f)
}

func (m *machine[TState, TTrigger]) OnTransitionCompleted(f func(Transition[TState, TTrigger])) {
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
