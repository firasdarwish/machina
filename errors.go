package machina

import "errors"

var ErrInvalidTransition = errors.New("unhandled transition")
var ErrEmptyGuards = errors.New("you must have at least one guard")
var ErrStateAlreadyConfigured = errors.New("state is already configured")
var ErrMaxDepthReached = errors.New("max depth reached")
var ErrTransitionDuplicated = errors.New("transition already exists with same Source-Trigger-Destination")
var ErrCyclicSubSuperState = errors.New("cyclic sub-state; the substate cannot be a superstate of itself")
var ErrSuperstateAlreadyConfigured = errors.New("superstate already configured")
