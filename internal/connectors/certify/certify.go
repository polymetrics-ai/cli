package certify

import (
	"context"
	"errors"
)

// Options configures a single-connector certification Runner (certification
// design §A command spec, single-connector subset only — batch/--all,
// --sweep, write/flow/schedule flags are out of scope for wave0).
type Options struct {
	Connector string
	Stream    string            // default: first cursor stream, else first
	Limit     int               // default 50
	Modes     []string          // default: all 5 sync modes
	Config    map[string]string // connector config for credentials add
	SecretEnv map[string]string // field -> ENV name
	KeepWork  bool
}

// Runner orchestrates certification stages for exactly one connector,
// serially, against an ephemeral project root (certification design §E
// package layout: certify.go "Runner + Options; per-connector
// orchestration"). Wave0 ships the skeleton only: stage execution
// (stages_source.go, 0-11) is implemented in a later task (T/B-14).
type Runner struct {
	opts Options
}

// NewRunner constructs a Runner for the given Options. Validation of
// Options (e.g. non-empty Connector) is deferred to Run so construction
// never fails.
func NewRunner(o Options) *Runner {
	return &Runner{opts: o}
}

// ErrNotImplemented is returned by Run in wave0: source stage execution
// (stages_source.go) lands in T/B-14; this package only proves the report
// schema and CLI harness end-to-end (SPEC.md §1.6).
var ErrNotImplemented = errors.New("certify: stage execution not implemented in this phase (see T/B-14)")

// Run will orchestrate source stages 0-11 once stages_source.go lands
// (T/B-14). In wave0's certify-core slice it validates Options and returns
// ErrNotImplemented so callers get a clear, typed signal rather than a
// panic or a silently empty report.
func (r *Runner) Run(ctx context.Context) (Report, error) {
	if r.opts.Connector == "" {
		return Report{}, errors.New("certify: Options.Connector is required")
	}
	if ctx == nil {
		return Report{}, errors.New("certify: nil context")
	}
	return Report{}, ErrNotImplemented
}
