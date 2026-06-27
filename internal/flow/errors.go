package flow

import "errors"

var (
	ErrManifestInvalid  = errors.New("flow: manifest invalid")
	ErrCyclicDependency = errors.New("flow: cyclic dependency detected")
	ErrLeaseHeld        = errors.New("flow: another run is already in progress")
	ErrStepFailed       = errors.New("flow: step failed")
	ErrUnknownStepKind  = errors.New("flow: unknown step kind")

	// Action-step errors.
	ErrApprovalRequired = errors.New("flow: action step requires approval token")
	ErrSchemaDrift      = errors.New("flow: schema drift detected — step paused")
	ErrTokenExpired     = errors.New("flow: approval token has expired")
	ErrTokenInvalid     = errors.New("flow: approval token is invalid")
)
