package flow

import "errors"

var (
	ErrManifestInvalid  = errors.New("flow: manifest invalid")
	ErrCyclicDependency = errors.New("flow: cyclic dependency detected")
	ErrLeaseHeld        = errors.New("flow: another run is already in progress")
	ErrStepFailed       = errors.New("flow: step failed")
	ErrUnknownStepKind  = errors.New("flow: unknown step kind")
)
