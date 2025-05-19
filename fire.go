package machina

func (m *machine[TState, TTrigger]) Fire(trigger TTrigger) error {
	return m.fire(trigger, false)
}

func (m *machine[TState, TTrigger]) CanFire(trigger TTrigger) bool {
	return m.fire(trigger, true) == nil
}

func (m *machine[TState, TTrigger]) fire(trigger TTrigger, dryRun bool) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	originalState := m.currentState

	if m.rollbackOnFailure {
		defer func() {
			if originalState != m.currentState {
				if r := recover(); r != nil {
					m.unsafeSetState(originalState)
					panic(r)
				}
			}
		}()
	}

	sc, exists := m.statesConfig[m.currentState]

	if !exists {
		return m.onUnhandledTransitionCallback(m.currentState, trigger)
	}

	// try fire substate
	var loopSc *stateConfig[TState, TTrigger]
	loopSc = sc
	depth := 0

	for {
		if depth > m.maxDepth {
			return ErrMaxDepthReached
		}

		transitionFound, err := m.tryFire(trigger, loopSc, dryRun)
		if err == nil {
			return nil
		}

		// error was due to a guard
		if transitionFound {
			return err
		}

		if loopSc.parentState == nil {
			return m.onUnhandledTransitionCallback(m.currentState, trigger)
		}

		parentState := *loopSc.parentState

		parentLoopSc, parentLoopScExists := m.statesConfig[parentState]
		if !parentLoopScExists {
			return m.onUnhandledTransitionCallback(m.currentState, trigger)
		}

		loopSc = parentLoopSc
		depth++
	}
}

func (m *machine[TState, TTrigger]) tryFire(trigger TTrigger, sc *stateConfig[TState, TTrigger], dryRun bool) (bool, error) {
	var transitionInfo TransitionInfo[TState, TTrigger]

	possibleTransitions, possibleTransitionExists := sc.transitions[trigger]
	if !possibleTransitionExists {
		return false, m.onUnhandledTransitionCallback(m.currentState, trigger)
	}

	var guardErrors []error

	for _, possibleTransition := range possibleTransitions {
		transitionInfo = TransitionInfo[TState, TTrigger]{
			FromState: m.currentState,
			ToState:   possibleTransition.toState,
			Trigger:   trigger,
		}

		guardError := m.evalGuards(transitionInfo, possibleTransition.guards)
		if guardError != nil {
			guardErrors = append(guardErrors, guardError)
			continue
		}

		// destination found
		dest := possibleTransition.toState

		if !dryRun {
			for _, callback := range m.onTransitionStartCallbacks {
				callback(transitionInfo)
			}

			for _, onExit := range sc.onExits {
				onExit(transitionInfo)
			}

			m.unsafeSetState(dest)

			destSc, destScExists := m.statesConfig[dest]
			if destScExists {
				for _, onEntry := range destSc.onEntries {
					onEntry(transitionInfo)
				}
			}

			for _, callback := range m.onTransitionCompletedCallbacks {
				callback(transitionInfo)
			}
		}

		return true, nil
	}

	return false, m.onUnhandledTransitionCallback(m.currentState, trigger)
}

func (m *machine[TState, TTrigger]) evalGuards(transitionInfo TransitionInfo[TState, TTrigger], guards []guard[TState, TTrigger]) error {
	for _, guard := range guards {
		result := guard(transitionInfo)
		if result != nil {
			return result
		}
	}

	return nil
}
