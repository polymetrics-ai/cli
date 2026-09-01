package cli

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/credential"
	"polymetrics.ai/internal/flow"
	"polymetrics.ai/internal/safety"
	"polymetrics.ai/internal/schedule"
)

const apiVersion = "polymetrics.ai/v1"

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

	// alreadyReported suppresses writeError's own stdout/stderr output after a
	// command has emitted a terminal result envelope.
	alreadyReported bool
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

// alreadyReportedExecutionError preserves the original categorized exit while
// preventing writeError from replacing a terminal run envelope with a second
// Error envelope. Callers use it only after stdout has accepted the persisted
// non-success run.
func alreadyReportedExecutionError(err error) error {
	classified := classifyError(err)
	clone := *classified
	clone.alreadyReported = true
	return &clone
}

func classifyError(err error) *cliError {
	if err == nil {
		return nil
	}
	var ce *cliError
	if errors.As(err, &ce) {
		return ce
	}
	var missingRequiredFlag *commandrunner.MissingRequiredFlagError
	if errors.As(err, &missingRequiredFlag) {
		return &cliError{category: categoryUsage, code: "usage_error", message: missingRequiredFlag.Error(), err: err}
	}
	var emptySecret *credential.EmptySecretError
	if errors.As(err, &emptySecret) {
		return &cliError{category: categoryValidation, code: "validation_error", message: emptySecret.Error(), err: err}
	}
	var invalidSecret *credential.InvalidSecretValueError
	if errors.As(err, &invalidSecret) {
		return &cliError{category: categoryValidation, code: "validation_error", message: invalidSecret.Error(), err: err}
	}
	var replay *app.AuthorizationTokenReplayError
	if errors.As(err, &replay) {
		// Replaying an approval token is invalid caller input, not an internal
		// fault. Preserve the typed rejection's redacted message while exposing
		// the standard validation category/code for scripts and monitoring.
		return &cliError{category: categoryValidation, code: "validation_error", message: err.Error(), err: err}
	}
	var flowReference *schedule.FlowReferenceError
	if errors.As(err, &flowReference) {
		return &cliError{category: categoryValidation, code: "schedule_flow_reference_refused", message: err.Error(), err: err}
	}
	var jobReference *flow.JobReferenceError
	if errors.As(err, &jobReference) {
		return &cliError{category: categoryValidation, code: "flow_job_reference_refused", message: err.Error(), err: err}
	}
	var preparedReplay *app.PreparedExecutionReplayError
	if errors.As(err, &preparedReplay) {
		return &cliError{category: categoryPolicy, code: "prepared_execution_replay", message: err.Error(), err: err}
	}
	var preparedRefused *app.PreparedExecutionRefusedError
	if errors.As(err, &preparedRefused) {
		return &cliError{category: categoryPolicy, code: "prepared_execution_refused", message: err.Error(), err: err}
	}
	var rateBudgetRefusal *connsdk.RateBudgetRefusalError
	if errors.As(err, &rateBudgetRefusal) {
		return &cliError{category: categoryPolicy, code: string(rateBudgetRefusal.Code), message: err.Error(), err: err}
	}
	var credentialRejected *connsdk.CredentialRejectedError
	if errors.As(err, &credentialRejected) {
		return &cliError{category: categoryAuth, code: "credential_error", message: credentialRejected.Error(), err: err}
	}
	var providerHTTPError *connsdk.HTTPError
	if errors.As(err, &providerHTTPError) && providerHTTPError.Status == http.StatusUnauthorized {
		// A provider-verified 401 means the caller's credential is missing,
		// revoked, or otherwise unacceptable. Deliberately replace the raw
		// provider error here: its URL and body can contain credential material.
		return &cliError{category: categoryAuth, code: "credential_error", message: "provider rejected the credential", err: err}
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
	if ce.alreadyReported {
		return exitCodeFor(ce)
	}
	public := publicErrorEnvelope(ce)
	message, _ := public["message"].(string)
	if jsonOut {
		_ = writeJSON(stdout, envelope{
			"api_version": apiVersion,
			"kind":        "Error",
			"error":       public,
		})
	}
	_, _ = fmt.Fprintf(stderr, "error: %s\n", message)
	return exitCodeFor(ce)
}

func publicErrorEnvelope(err error) envelope {
	ce := classifyError(err)
	if ce == nil {
		return nil
	}
	return envelope{
		"category": string(ce.category),
		"code":     ce.code,
		"message":  safety.SanitizeTerminal(safety.RedactErrorText(ce.Error())),
	}
}
