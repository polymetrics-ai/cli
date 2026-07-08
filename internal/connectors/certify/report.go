// Package certify implements the connector certification harness core:
// the CertificationReport schema (report.go), an in-process CLI driver
// (cliharness.go), and a minimal single-connector Runner skeleton
// (certify.go). Scope is limited to
// docs/architecture/connector-certification-design.md implementation-order
// steps 1-2; write/flow/schedule stages and CLI wiring land in later phases
// (SPEC.md §1.6).
package certify

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CapabilityResult is a single pass/fail/skip entry under Report.Capabilities,
// per the certification design §A "Report artifact" shape. Not every field
// applies to every capability; unused fields are omitted from JSON.
type CapabilityResult struct {
	Result        string `json:"result"`
	Streams       int    `json:"streams,omitempty"`
	Stream        string `json:"stream,omitempty"`
	Records       int    `json:"records,omitempty"`
	StagesChecked int    `json:"stages_checked,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

// SyncModeResult is one row of Capabilities.SyncModes.
type SyncModeResult struct {
	Result         string `json:"result"`
	DataSource     string `json:"data_source"` // "live" | "capture"
	CursorAdvanced bool   `json:"cursor_advanced,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

// ScheduleResult is Capabilities.Schedule (certification design §A report
// artifact "schedule": {"result", "backend", "residue"}).
type ScheduleResult struct {
	Result  string `json:"result"`
	Backend string `json:"backend,omitempty"`
	Residue bool   `json:"residue"`
	Reason  string `json:"reason,omitempty"`
}

// WriteActionResult is one entry of Capabilities.WriteActions (design §A
// report artifact example: "create_issue": {"result", "cleanup", "verify",
// "tag"}).
type WriteActionResult struct {
	Result  string `json:"result"`
	Cleanup string `json:"cleanup,omitempty"`
	Verify  string `json:"verify,omitempty"`
	Tag     string `json:"tag,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// Capabilities mirrors the design §A "capabilities" object. Flow/Schedule/
// WriteActions are pointers/nil-able maps so a report produced before those
// stages run (or that skips them) omits the keys entirely rather than
// serializing a zero-value placeholder block, matching
// TestReportMarshalJSONShape's absence assertion. Budget remains a
// later-phase field, deliberately absent from this struct (DATA-MODEL.md
// §6).
type Capabilities struct {
	Check           CapabilityResult             `json:"check"`
	Catalog         CapabilityResult             `json:"catalog"`
	Read            CapabilityResult             `json:"read"`
	SyncModes       map[string]SyncModeResult    `json:"sync_modes"`
	Resume          CapabilityResult             `json:"resume"`
	JSONContract    CapabilityResult             `json:"json_contract"`
	SecretRedaction CapabilityResult             `json:"secret_redaction"`
	DirectRead      *CapabilityResult            `json:"direct_read,omitempty"`
	Flow            *CapabilityResult            `json:"flow,omitempty"`
	Schedule        *ScheduleResult              `json:"schedule,omitempty"`
	WriteActions    map[string]WriteActionResult `json:"write_actions,omitempty"`
}

// CLIStageInfo records the redacted invocation and outcome of one in-process
// CLI call made during a stage (certification design §A "stages[].cli").
type CLIStageInfo struct {
	ArgvRedacted string `json:"argv_redacted"`
	ExitCode     int    `json:"exit_code"`
	Kind         string `json:"kind"`
}

// StageResult is one entry of Report.Stages.
type StageResult struct {
	Name       string       `json:"name"`
	Tier       int          `json:"tier"`
	Passed     bool         `json:"passed"`
	DurationMS int64        `json:"duration_ms"`
	Error      string       `json:"error,omitempty"`
	CLI        CLIStageInfo `json:"cli"`
}

// Leak is one entry of Report.Leaks: a write-protocol tag that was
// successfully created but never verified cleaned up (design §C "Create ok
// + cleanup/verify fails -> leaked_resource: top-level leaks[]").
type Leak struct {
	Tag       string `json:"tag"`
	Connector string `json:"connector"`
	Action    string `json:"action,omitempty"`
	Reason    string `json:"reason"`
}

// ExitCodeFor maps a completed Report to the certification design §A exit
// code convention: 0 pass, 2 certification failures, 3 leaked resources
// (dominates everything — checked first).
func ExitCodeFor(rep Report) int {
	if len(rep.Leaks) != 0 {
		return 3
	}
	if !rep.Passed {
		return 2
	}
	return 0
}

// Report is the CertificationReport artifact persisted at
// .polymetrics/certifications/<connector>.json (certification design §A).
// Budget remains empty/absent until a later certify phase (DATA-MODEL.md
// §6).
type Report struct {
	Kind          string    `json:"kind"`
	SchemaVersion int       `json:"schema_version"`
	Connector     string    `json:"connector"`
	PMVersion     string    `json:"pm_version"`
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   time.Time `json:"completed_at"`
	Mode          string    `json:"mode"`
	Passed        bool      `json:"passed"`

	Capabilities Capabilities  `json:"capabilities"`
	Stages       []StageResult `json:"stages"`
	Leaks        []Leak        `json:"leaks,omitempty"`
}

// certificationsDirName / historyDirName are the fixed on-disk layout under
// a project's .polymetrics directory (or, in tests, an arbitrary root dir):
// <root>/certifications/<connector>.json and
// <root>/certifications/history/<connector>/<timestamp>.json.
const (
	certificationsDirName = "certifications"
	historyDirName        = "history"
)

// historyTimestampFormat renders StartedAt as a filesystem-safe, sortable
// UTC timestamp: 20060102T150405Z.
const historyTimestampFormat = "20060102T150405Z"

// Save writes the report to <dir>/certifications/<connector>.json and
// appends a copy to <dir>/certifications/history/<connector>/<timestamp>.json,
// per certification design §A ("History appends to
// .polymetrics/certifications/history/<connector>/<timestamp>.json").
func (rep *Report) Save(dir string) error {
	if rep.Connector == "" {
		return errors.New("certify: report.Connector is required")
	}

	certDir := filepath.Join(dir, certificationsDirName)
	if err := os.MkdirAll(certDir, 0o755); err != nil {
		return fmt.Errorf("certify: create certifications dir: %w", err)
	}

	raw, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return fmt.Errorf("certify: marshal report: %w", err)
	}

	connectorPath := filepath.Join(certDir, rep.Connector+".json")
	if err := os.WriteFile(connectorPath, raw, 0o644); err != nil {
		return fmt.Errorf("certify: write %s: %w", connectorPath, err)
	}

	historyDir := filepath.Join(certDir, historyDirName, rep.Connector)
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		return fmt.Errorf("certify: create history dir: %w", err)
	}
	historyPath := filepath.Join(historyDir, rep.StartedAt.UTC().Format(historyTimestampFormat)+".json")
	if err := os.WriteFile(historyPath, raw, 0o644); err != nil {
		return fmt.Errorf("certify: write %s: %w", historyPath, err)
	}

	return nil
}

// LoadReport reads a Report previously written by Save from an exact file
// path (typically <dir>/certifications/<connector>.json or a history entry).
func LoadReport(path string) (Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Report{}, fmt.Errorf("certify: read %s: %w", path, err)
	}
	var rep Report
	if err := json.Unmarshal(raw, &rep); err != nil {
		return Report{}, fmt.Errorf("certify: unmarshal %s: %w", path, err)
	}
	return rep, nil
}
