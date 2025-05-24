package machina

func (m *machine[TState, TTrigger]) Fire(trigger TTrigger, params ...any) error {
	return m.fire(trigger, false, params)
}

func (m *machine[TState, TTrigger]) CanFire(trigger TTrigger, params ...any) bool {
	return m.fire(trigger, true, params) == nil
}

func (m *machine[TState, TTrigger]) fire(trigger TTrigger, dryRun bool, params []any) error {
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

		transitionFound, err := m.tryFire(trigger, loopSc, dryRun, params)
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

func (m *machine[TState, TTrigger]) tryFire(trigger TTrigger, sc *stateConfig[TState, TTrigger], dryRun bool, params []any) (bool, error) {
	var tInfo *transitionInfo[TState, TTrigger]

	possibleTransitions, possibleTransitionExists := sc.transitions[trigger]
	if !possibleTransitionExists {
		return false, m.onUnhandledTransitionCallback(m.currentState, trigger)
	}

	var guardErrors []error

	for _, possibleTransition := range possibleTransitions {

		tInfo = &transitionInfo[TState, TTrigger]{
			fromState: m.currentState,
			toState:   possibleTransition.toState,
			trigger:   trigger,
			params:    params,
		}

		guardError := m.evalGuards(tInfo, possibleTransition.guards)
		if guardError != nil {
			guardErrors = append(guardErrors, guardError)
			continue
		}

		// destination found
		dest := possibleTransition.toState

		if !dryRun {
			for _, callback := range m.onTransitionStartCallbacks {
				callback(tInfo)
			}

			for _, onExit := range sc.onExits {
				onExit(tInfo)
			}

			m.unsafeSetState(dest)

			destSc, destScExists := m.statesConfig[dest]
			if destScExists {
				for _, onEntry := range destSc.onEntries {
					onEntry(tInfo)
				}
			}

			for _, callback := range m.onTransitionCompletedCallbacks {
				callback(tInfo)
			}
		}

		return true, nil
	}

	return false, m.onUnhandledTransitionCallback(m.currentState, trigger)
}

func (m *machine[TState, TTrigger]) evalGuards(transitionInfo Transition[TState, TTrigger], guards []guard[TState, TTrigger]) error {
	for _, g := range guards {
		result := g(transitionInfo)
		if result != nil {
			return result
		}
	}

	return nil
}
