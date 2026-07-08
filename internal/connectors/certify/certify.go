package certify

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
}

// NewRunner constructs a Runner for the given Options. Validation of
// Options (e.g. non-empty Connector) is deferred to Run so construction
// never fails.
func NewRunner(o Options) *Runner {
	return &Runner{opts: o}
}
