package cli

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors/connsdk"
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

	// exitOverride, when non-nil, replaces the category-derived exit code in
	// exitCodeFor. Used exclusively by certify's typed exit-code contract
	// (certify.ExitError, wired via certifyErrorf in certify_cli.go): the
	// certification design's own exit codes (0 pass / 1 usage-internal / 2
	// certification failures / 3 leaked resources) are unrelated to every
	// other command's category-based mapping below, so this field is never
	// set by any other error constructor in this package.
	exitOverride *int

	// alreadyReported suppresses writeError's own stdout/stderr output. Set
	// only by certifyExitErrorf: `pm connectors certify` writes its
	// ConnectorCertification/BatchCertification report envelope (or
	// human-readable rendering) to stdout itself BEFORE returning this
	// error, so writeError must not also emit a second, conflicting "Error"
	// JSON envelope — cli.Run's one-envelope-per-invocation contract
	// (json_contract stage, THREAT-MODEL.md) would otherwise be violated.
	// writeError still returns the correct exit code either way.
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

// certifyExitErrorf builds the non-nil-error result `pm connectors certify`
// returns whenever it must exit non-zero, per certification design §A:
// "Exit codes: 0 pass · 1 usage/internal · 2 certification failures · 3
// leaked resources (dominates everything)". The caller is expected to have
// already written the report (JSON envelope or human-readable rendering) to
// stdout itself; this constructor's message is used only for the plain-text
// stderr summary line, never re-emitted as a second JSON envelope (see
// cliError.alreadyReported). A passing run (exit 0) instead writes its
// output and returns a nil error, exactly like every other successful
// command in this package — this constructor is only ever called with code
// 1, 2, or 3.
func certifyExitErrorf(code int, format string, args ...any) error {
	c := code
	return &cliError{
		category:        categoryInternal,
		code:            "certify_exit",
		message:         fmt.Sprintf(format, args...),
		exitOverride:    &c,
		alreadyReported: true,
	}
}

func classifyError(err error) *cliError {
	if err == nil {
		return nil
	}
	var ce *cliError
	if errors.As(err, &ce) {
		return ce
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
	if err.exitOverride != nil {
		return *err.exitOverride
	}
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
	message := safety.SanitizeTerminal(safety.RedactErrorText(ce.Error()))
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
	_, _ = fmt.Fprintf(stderr, "error: %s\n", message)
	return exitCodeFor(ce)
}
