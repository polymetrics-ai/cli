package rlm

import "context"

// ModelAnalyzer is a placeholder for the Phase 4 model backend.
//
// HUMAN GATE: Do NOT implement until Phase 4 is explicitly approved.
// The model backend makes outbound network calls to the Claude API,
// requires credential configuration, and changes the network/credential
// surface of pm. A human must approve Phase 4 before any HTTP client
// code, credential lookup, or response caching is added here.
type ModelAnalyzer struct{}

// Mode returns the backend identifier.
func (m *ModelAnalyzer) Mode() string { return "model" }

// Run always returns ErrNotImplemented.
// HUMAN GATE: model calls require Phase 4 approval before implementation.
func (m *ModelAnalyzer) Run(_ context.Context, _ RunRequest) (RunResult, error) {
	return RunResult{}, ErrNotImplemented
}
