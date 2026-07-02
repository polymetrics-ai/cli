// Package conformance implements conformance v2 (design §E.2): static
// structural/policy checks plus dynamic fixture-backed replay checks that
// exercise the REAL engine (internal/connectors/engine) against recorded
// fixture pages. It replaces the synthetic mode=fixture record shortcut of
// the legacy internal/connectors/native_conformance.go (untouched,
// out of scope, kept only as a reference for the report-shape/accumulator
// pattern this package mirrors).
package conformance

import (
	"encoding/json"
	"fmt"

	"polymetrics.ai/internal/connectors/engine"
)

// CheckResult is one named conformance check's outcome, mirroring
// native_conformance.go's NativeConformanceTest{Name, Passed, Error} shape
// plus Skipped: a check that could not run at all (e.g. every check after a
// Load failure, or a dynamic check whose bundle declares no fixtures for
// that stream/action) is reported Skipped rather than silently omitted, so
// the check list is always complete and machine-readable.
type CheckResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Skipped bool   `json:"skipped,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Report is the aggregate conformance result for one connector bundle:
// {Connector, Checks: [...], Passed}, per the dispatch brief's report-shape
// note (mirrors native_conformance.go's NativeConformanceReport, generalized
// to a flat check list since conformance v2 has no fixture-vs-live capability
// axis of its own — that distinction lives in certify, not here).
type Report struct {
	Connector string        `json:"connector"`
	Checks    []CheckResult `json:"checks"`
	Passed    bool          `json:"passed"`
}

// computePassed reports whether every non-skipped check in rep.Checks
// passed. A Report with zero checks (e.g. a bundle with no writes.json, so
// no write-fixture checks were even applicable) trivially passes.
func (rep Report) computePassed() bool {
	for _, c := range rep.Checks {
		if c.Skipped {
			continue
		}
		if !c.Passed {
			return false
		}
	}
	return true
}

// addCheck appends a CheckResult built from name+err (err == nil => Passed).
func addCheck(checks []CheckResult, name string, err error) []CheckResult {
	c := CheckResult{Name: name, Passed: err == nil}
	if err != nil {
		c.Error = err.Error()
	}
	return append(checks, c)
}

// addSkipped appends a Skipped CheckResult for name.
func addSkipped(checks []CheckResult, name string) []CheckResult {
	return append(checks, CheckResult{Name: name, Skipped: true})
}

// RunBundle runs every static and dynamic conformance check against an
// already-loaded Bundle and returns the aggregate Report.
func RunBundle(b engine.Bundle) Report {
	var checks []CheckResult
	checks = append(checks, runStaticChecks(b)...)
	checks = append(checks, runDynamicChecks(b)...)
	rep := Report{Connector: b.Name, Checks: checks}
	rep.Passed = rep.computePassed()
	return rep
}

// allCheckNames is every check name conformance v2 knows about, static
// followed by the dynamic checks that are always structurally applicable
// (check_fixture) — used by ReportFromLoadError to report a complete,
// all-Skipped check list when a bundle never even loads. Per-stream/
// per-action dynamic checks (read_fixture_nonempty, pagination_terminates,
// records_match_schema, cursor_advances, write_request_shape,
// delete_semantics) are inherently bundle-shaped (their count depends on
// how many streams/actions exist) and are therefore NOT enumerable before a
// bundle loads; ReportFromLoadError does not fabricate placeholders for
// them, matching the "a Load failure means we know nothing about the
// bundle's shape yet" reality.
var alwaysApplicableDynamicChecks = []string{"check_fixture"}

// ReportFromLoadError builds a Report for a bundle directory named
// connectorName whose engine.Load call itself failed. The specific failing
// static check is named by classifying the error message (mirroring
// cmd/connectorgen/validate.go's loadErrorFinding classification — kept as
// this package's own copy per PLAN.md's "no cross-package sharing between
// B-11's and B-13's corpora" note); every other check name is reported
// Skipped so callers always see the full, stable check-name set.
func ReportFromLoadError(connectorName string, err error) Report {
	failedCheck := classifyLoadError(err)

	var checks []CheckResult
	for _, name := range staticCheckNames {
		if name == failedCheck {
			checks = addCheck(checks, name, err)
			continue
		}
		checks = addSkipped(checks, name)
	}
	for _, name := range alwaysApplicableDynamicChecks {
		checks = addSkipped(checks, name)
	}

	rep := Report{Connector: connectorName, Checks: checks}
	rep.Passed = rep.computePassed()
	return rep
}

// jsonMarshal is a thin indirection over encoding/json.Marshal so tests can
// reference a package-local name (kept for symmetry with the rest of the
// package's small helper functions; also gives one place to switch to
// MarshalIndent if a future caller wants pretty output).
func jsonMarshal(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("conformance: marshal report: %w", err)
	}
	return raw, nil
}
