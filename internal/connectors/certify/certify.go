package certify

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
// orchestration"). Run (stages_source.go) executes source stages 0-11;
// write/flow/schedule stages are out of scope for wave0 (SPEC.md §1.6).
type Runner struct {
	opts Options

	// sabotage, stdoutLeakSabotage, and lastWorkdir support self-tests only
	// (see SabotageExpectedKind / SabotageStdoutLeak / LastWorkdir in
	// stages_source.go) and are never set by production callers.
	sabotage           *sabotage
	stdoutLeakSabotage *stdoutLeakSabotage
	lastWorkdir        string
}

// NewRunner constructs a Runner for the given Options. Validation of
// Options (e.g. non-empty Connector) is deferred to Run so construction
// never fails.
func NewRunner(o Options) *Runner {
	return &Runner{opts: o}
}
