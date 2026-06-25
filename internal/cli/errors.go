package cli

import (
	"errors"
	"fmt"
	"io"

	"polymetrics/internal/safety"
)

const apiVersion = "polymetrics.io/v1"

type errorCategory string

const (
	categoryUsage      errorCategory = "usage"
	categoryAuth       errorCategory = "auth"
	categoryValidation errorCategory = "validation"
	categoryConnector  errorCategory = "connector"
	categoryRuntime    errorCategory = "runtime"
	categoryPolicy     errorCategory = "policy"
	categoryInternal   errorCategory = "internal"
)

type cliError struct {
	category errorCategory
	code     string
	message  string
	err      error
}

func (e *cliError) Error() string {
	if e.message != "" {
		return e.message
	}
	if e.err != nil {
		return e.err.Error()
	}
	return string(e.category)
}

func (e *cliError) Unwrap() error { return e.err }

func usageErrorf(format string, args ...any) error {
	return &cliError{category: categoryUsage, code: "usage_error", message: fmt.Sprintf(format, args...)}
}

func validationErrorf(format string, args ...any) error {
	return &cliError{category: categoryValidation, code: "validation_error", message: fmt.Sprintf(format, args...)}
}

func classifyError(err error) *cliError {
	if err == nil {
		return nil
	}
	var ce *cliError
	if errors.As(err, &ce) {
		return ce
	}
	if errors.Is(err, errUsage) {
		return &cliError{category: categoryUsage, code: "usage_error", message: errUsage.Error()}
	}
	return &cliError{category: categoryInternal, code: "internal_error", message: err.Error(), err: err}
}

func exitCodeFor(err *cliError) int {
	switch err.category {
	case categoryUsage:
		return 2
	case categoryValidation:
		return 3
	case categoryAuth:
		return 4
	case categoryConnector:
		return 5
	case categoryRuntime:
		return 6
	case categoryPolicy:
		return 7
	default:
		return 1
	}
}

func writeError(stdout, stderr io.Writer, err error, jsonOut bool) int {
	ce := classifyError(err)
	if ce == nil {
		return 0
	}
	message := safety.SanitizeTerminal(ce.Error())
	if jsonOut {
		_ = writeJSON(stdout, envelope{
			"api_version": apiVersion,
			"kind":        "Error",
			"error": envelope{
				"category": string(ce.category),
				"code":     ce.code,
				"message":  message,
			},
		})
	}
	fmt.Fprintf(stderr, "error: %s\n", message)
	return exitCodeFor(ce)
}
