package connectors

import "context"

// ReadInputValidationRequest contains only the selected stream and non-secret
// input values. Validation cannot resolve credentials or construct a runtime.
type ReadInputValidationRequest struct {
	Stream string
	Config map[string]string
	Query  map[string]string
}

// ReadInputValidator is the optional pure pre-secret check for generated input
// contracts. An absent selected contract retains legacy behavior.
type ReadInputValidator interface {
	ValidateReadInputs(context.Context, ReadInputValidationRequest) error
}
