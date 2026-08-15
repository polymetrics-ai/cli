package certify

import "time"

const defaultCertificationRateLimitAdmissionTimeout = 30 * time.Second

// Options configures a single-connector certification Runner (certification
// design §A command spec, single-connector subset only — batch/--all,
// --sweep flags are out of scope here; see credsfile.go for the batch-mode
// ConnectorCredsEntry.Write equivalent).
type Options struct {
	Connector string
	Stream    string            // default: first cursor stream, else first
	Limit     int               // default 50
	Modes     []string          // default: all 5 sync modes
	Config    map[string]string // connector config for credentials add
	SecretEnv map[string]string // field -> ENV name
	KeepWork  bool

	// Write enables the create-then-cleanup write protocol (stages 12-17,
	// design §C). When false, or when the connector has no available
	// WritePairing, the write stages record a documented skip rather than
	// attempting any live write (design §A "no credential -> uncertified,
	// never failed" applies analogously here: absence of a safe write path
	// must never fail the report).
	Write bool

	// Full enables the comprehensive sweep: every stream (not just the
	// first), every write pairing (not just the first), binary downloads,
	// and direct reads. The existing single-pairing write stages still run
	// first; Full adds a stageWriteSweepAllPairings stage that iterates the
	// remaining pairings. See
	// docs/plans/connector-complete-testing-and-mail-setup-plan.md.
	Full bool

	// ObserveHTTP retains bounded, exact HTTP exchanges in process memory for
	// an external-binary proof. It is deliberately not a user-facing capture
	// switch: the CLI enables it only inside a freshly built child process.
	ObserveHTTP bool

	// RateLimitAdmissionTimeout bounds a single provider-budget wait. A zero
	// value uses the certification default; it does not shorten normal stages.
	RateLimitAdmissionTimeout time.Duration
}

// Runner orchestrates certification stages for exactly one connector,
// serially, against an ephemeral project root (certification design §E
// package layout: certify.go "Runner + Options; per-connector
// orchestration"). Run (stages_source.go) executes source stages 0-11, the
// write protocol (stages_write.go, stages 12-17), and the glue stages
// (stages_glue.go, stages 18-19).
type Runner struct {
	opts Options

	// sabotage, stdoutLeakSabotage, cleanupVerifySabotage, and lastWorkdir
	// support self-tests only (see SabotageExpectedKind /
	// SabotageStdoutLeak / SabotageCleanupVerifyEntityStillPresent /
	// LastWorkdir in stages_source.go / stages_write.go) and are never set
	// by production callers.
	sabotage              *sabotage
	stdoutLeakSabotage    *stdoutLeakSabotage
	cleanupVerifySabotage bool
	lastWorkdir           string
	observedHTTP          []ObservedHTTPExchange
}

// NewRunner constructs a Runner for the given Options. Validation of
// Options (e.g. non-empty Connector) is deferred to Run so construction
// never fails.
func NewRunner(o Options) *Runner {
	return &Runner{opts: o}
}

// ObservedHTTPExchanges returns the last run's defensive observation snapshot.
// It contains raw process-memory values and must be passed directly to
// WriteExternalProof rather than logged or serialized by a caller.
func (r *Runner) ObservedHTTPExchanges() []ObservedHTTPExchange {
	if r == nil {
		return nil
	}
	out := make([]ObservedHTTPExchange, len(r.observedHTTP))
	for index, exchange := range r.observedHTTP {
		out[index] = cloneObservedHTTPExchange(exchange)
	}
	return out
}
