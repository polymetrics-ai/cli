package agentloop

const (
	safetySchemaVersion = "1.0"
	safetyClosedState   = "closed"
	safetyDisabledCode  = "AUTO_LOOP_DISABLED_PHASE_0"
	safetyDisabledClass = "safety_disabled"
)

// SafetyStatus reports the immutable Phase 0 run/resume posture.
type SafetyStatus struct {
	SchemaVersion string `json:"schema_version"`
	State         string `json:"state"`
	RunEnabled    bool   `json:"run_enabled"`
	ResumeEnabled bool   `json:"resume_enabled"`
	Code          string `json:"code"`
	ExitClass     string `json:"exit_class"`
}

// GuardResult binds the closed posture to one requested wrapper.
type GuardResult struct {
	SchemaVersion string `json:"schema_version"`
	State         string `json:"state"`
	RunEnabled    bool   `json:"run_enabled"`
	ResumeEnabled bool   `json:"resume_enabled"`
	Code          string `json:"code"`
	ExitClass     string `json:"exit_class"`
	Entrypoint    string `json:"entrypoint,omitempty"`
	ExitCode      int    `json:"-"`
}

// CurrentSafetyStatus is source-closed and reads no external configuration.
func CurrentSafetyStatus() SafetyStatus {
	return SafetyStatus{
		SchemaVersion: safetySchemaVersion,
		State:         safetyClosedState,
		RunEnabled:    false,
		ResumeEnabled: false,
		Code:          safetyDisabledCode,
		ExitClass:     safetyDisabledClass,
	}
}

// TrackedEntrypoints returns a fresh sorted inventory.
func TrackedEntrypoints() []string {
	return []string{
		"scripts/claude-auto-loop.sh",
		"scripts/pi-auto-loop.sh",
		"scripts/pi-shepherd-loop.sh",
	}
}

// GuardDriver denies tracked wrappers with EX_CONFIG and rejects unknown
// wrappers with EX_USAGE. There is deliberately no enable counterpart.
func GuardDriver(entrypoint string) GuardResult {
	for _, tracked := range TrackedEntrypoints() {
		if entrypoint == tracked {
			return GuardResult{
				SchemaVersion: safetySchemaVersion,
				State:         safetyClosedState,
				RunEnabled:    false,
				ResumeEnabled: false,
				Code:          safetyDisabledCode,
				ExitClass:     safetyDisabledClass,
				Entrypoint:    entrypoint,
				ExitCode:      78,
			}
		}
	}
	return GuardResult{
		SchemaVersion: safetySchemaVersion,
		State:         safetyClosedState,
		RunEnabled:    false,
		ResumeEnabled: false,
		Code:          "AUTO_LOOP_ENTRYPOINT_UNTRACKED",
		ExitClass:     "usage_error",
		ExitCode:      64,
	}
}
