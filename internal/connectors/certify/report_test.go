package certify_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/certify"
)

// sampleReport builds a populated report matching the wave0 subset of the
// certification design §A JSON shape (DATA-MODEL.md §6): kind, schema_version,
// connector, pm_version, started_at, completed_at, mode, passed,
// capabilities.{check,catalog,read,sync_modes,resume,json_contract,secret_redaction},
// stages[]. leaks/write_actions/flow/schedule/budget stay empty/absent in wave0.
func sampleReport() certify.Report {
	started := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	completed := started.Add(2500 * time.Millisecond)
	return certify.Report{
		Kind:          "ConnectorCertification",
		SchemaVersion: 1,
		Connector:     "sample",
		PMVersion:     "v0.0.0-test",
		StartedAt:     started,
		CompletedAt:   completed,
		Mode:          "live",
		Passed:        true,
		Capabilities: certify.Capabilities{
			Check:   certify.CapabilityResult{Result: "pass"},
			Catalog: certify.CapabilityResult{Result: "pass", Streams: 1},
			Read:    certify.CapabilityResult{Result: "pass", Stream: "customers", Records: 3},
			SyncModes: map[string]certify.SyncModeResult{
				"full_refresh_append": {Result: "pass", DataSource: "live"},
			},
			Resume:          certify.CapabilityResult{Result: "pass"},
			JSONContract:    certify.CapabilityResult{Result: "pass", StagesChecked: 12},
			SecretRedaction: certify.CapabilityResult{Result: "pass"},
		},
		Stages: []certify.StageResult{
			{
				Name:       "credentials_test",
				Tier:       2,
				Passed:     true,
				DurationMS: 812,
				CLI: certify.CLIStageInfo{
					ArgvRedacted: "pm credentials test cert-sample --json",
					ExitCode:     0,
					Kind:         "CredentialTest",
				},
			},
		},
	}
}

func TestReportMarshalRoundTrip(t *testing.T) {
	want := sampleReport()

	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got certify.Report
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got.Kind != want.Kind {
		t.Errorf("Kind = %q, want %q", got.Kind, want.Kind)
	}
	if got.SchemaVersion != want.SchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", got.SchemaVersion, want.SchemaVersion)
	}
	if got.Connector != want.Connector {
		t.Errorf("Connector = %q, want %q", got.Connector, want.Connector)
	}
	if !got.Passed {
		t.Errorf("Passed = false, want true")
	}
	if got.Mode != want.Mode {
		t.Errorf("Mode = %q, want %q", got.Mode, want.Mode)
	}
	if !got.StartedAt.Equal(want.StartedAt) {
		t.Errorf("StartedAt = %v, want %v", got.StartedAt, want.StartedAt)
	}
	if !got.CompletedAt.Equal(want.CompletedAt) {
		t.Errorf("CompletedAt = %v, want %v", got.CompletedAt, want.CompletedAt)
	}
	if got.Capabilities.Check.Result != "pass" {
		t.Errorf("Capabilities.Check.Result = %q, want pass", got.Capabilities.Check.Result)
	}
	if got.Capabilities.Catalog.Streams != 1 {
		t.Errorf("Capabilities.Catalog.Streams = %d, want 1", got.Capabilities.Catalog.Streams)
	}
	if got.Capabilities.Read.Records != 3 {
		t.Errorf("Capabilities.Read.Records = %d, want 3", got.Capabilities.Read.Records)
	}
	mode, ok := got.Capabilities.SyncModes["full_refresh_append"]
	if !ok {
		t.Fatalf("Capabilities.SyncModes missing full_refresh_append: %+v", got.Capabilities.SyncModes)
	}
	if mode.Result != "pass" || mode.DataSource != "live" {
		t.Errorf("SyncModes[full_refresh_append] = %+v, want {pass live}", mode)
	}
	if len(got.Stages) != 1 {
		t.Fatalf("len(Stages) = %d, want 1", len(got.Stages))
	}
	if got.Stages[0].CLI.ArgvRedacted != "pm credentials test cert-sample --json" {
		t.Errorf("Stages[0].CLI.ArgvRedacted = %q", got.Stages[0].CLI.ArgvRedacted)
	}
	if got.Stages[0].CLI.Kind != "CredentialTest" {
		t.Errorf("Stages[0].CLI.Kind = %q, want CredentialTest", got.Stages[0].CLI.Kind)
	}
}

func TestReportMarshalJSONShape(t *testing.T) {
	want := sampleReport()

	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatalf("Unmarshal(generic) error = %v", err)
	}

	for _, key := range []string{
		"kind", "schema_version", "connector", "pm_version",
		"started_at", "completed_at", "mode", "passed",
		"capabilities", "stages",
	} {
		if _, ok := generic[key]; !ok {
			t.Errorf("marshaled report missing top-level key %q: %s", key, string(raw))
		}
	}

	caps, ok := generic["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("capabilities is not an object: %s", string(raw))
	}
	for _, key := range []string{"check", "catalog", "read", "sync_modes", "resume", "json_contract", "secret_redaction"} {
		if _, ok := caps[key]; !ok {
			t.Errorf("capabilities missing key %q: %s", key, string(raw))
		}
	}

	// leaks/write_actions/flow/schedule/budget stay empty/absent in wave0
	// (design §A fields not yet populated); ensure we don't accidentally
	// serialize a placeholder null block for them under "capabilities".
	if _, ok := caps["write_actions"]; ok {
		t.Errorf("capabilities.write_actions should be absent in wave0, got: %s", string(raw))
	}
	if _, ok := caps["flow"]; ok {
		t.Errorf("capabilities.flow should be absent in wave0, got: %s", string(raw))
	}
	if _, ok := caps["schedule"]; ok {
		t.Errorf("capabilities.schedule should be absent in wave0, got: %s", string(raw))
	}
}

func TestReportSaveWritesConnectorFile(t *testing.T) {
	dir := t.TempDir()
	rep := sampleReport()

	if err := rep.Save(dir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path := filepath.Join(dir, "certifications", "sample.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}

	var saved certify.Report
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Unmarshal(saved) error = %v", err)
	}
	if saved.Connector != "sample" {
		t.Errorf("saved.Connector = %q, want sample", saved.Connector)
	}
	if !saved.Passed {
		t.Errorf("saved.Passed = false, want true")
	}
}

func TestReportSaveAppendsHistory(t *testing.T) {
	dir := t.TempDir()
	rep := sampleReport()

	if err := rep.Save(dir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	historyDir := filepath.Join(dir, "certifications", "history", "sample")
	entries, err := os.ReadDir(historyDir)
	if err != nil {
		t.Fatalf("ReadDir(%s) error = %v", historyDir, err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(history entries) = %d, want 1: %v", len(entries), entries)
	}

	wantName := rep.StartedAt.UTC().Format("20060102T150405Z") + ".json"
	if entries[0].Name() != wantName {
		t.Errorf("history file name = %q, want %q", entries[0].Name(), wantName)
	}

	data, err := os.ReadFile(filepath.Join(historyDir, entries[0].Name()))
	if err != nil {
		t.Fatalf("ReadFile(history entry) error = %v", err)
	}
	var saved certify.Report
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Unmarshal(history entry) error = %v", err)
	}
	if saved.Connector != "sample" {
		t.Errorf("history saved.Connector = %q, want sample", saved.Connector)
	}
}

func TestReportSaveSecondRunAppendsSecondHistoryEntry(t *testing.T) {
	dir := t.TempDir()
	first := sampleReport()
	if err := first.Save(dir); err != nil {
		t.Fatalf("Save(first) error = %v", err)
	}

	second := sampleReport()
	second.StartedAt = first.StartedAt.Add(time.Hour)
	second.CompletedAt = first.CompletedAt.Add(time.Hour)
	second.Passed = false
	if err := second.Save(dir); err != nil {
		t.Fatalf("Save(second) error = %v", err)
	}

	historyDir := filepath.Join(dir, "certifications", "history", "sample")
	entries, err := os.ReadDir(historyDir)
	if err != nil {
		t.Fatalf("ReadDir(%s) error = %v", historyDir, err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(history entries) = %d, want 2: %v", len(entries), entries)
	}

	// Latest connector-level file reflects the most recent run.
	latest, err := certify.LoadReport(filepath.Join(dir, "certifications", "sample.json"))
	if err != nil {
		t.Fatalf("LoadReport() error = %v", err)
	}
	if latest.Passed {
		t.Errorf("latest.Passed = true, want false (second run failed)")
	}
}

func TestLoadReportRoundTrip(t *testing.T) {
	dir := t.TempDir()
	rep := sampleReport()
	if err := rep.Save(dir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := certify.LoadReport(filepath.Join(dir, "certifications", "sample.json"))
	if err != nil {
		t.Fatalf("LoadReport() error = %v", err)
	}
	if loaded.Connector != rep.Connector {
		t.Errorf("loaded.Connector = %q, want %q", loaded.Connector, rep.Connector)
	}
	if loaded.SchemaVersion != rep.SchemaVersion {
		t.Errorf("loaded.SchemaVersion = %d, want %d", loaded.SchemaVersion, rep.SchemaVersion)
	}
	if !loaded.Passed {
		t.Errorf("loaded.Passed = false, want true")
	}
}

func TestLoadReportMissingFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := certify.LoadReport(filepath.Join(dir, "certifications", "missing.json")); err == nil {
		t.Fatalf("LoadReport() error = nil, want error for missing file")
	}
}

func TestReportSaveRequiresConnectorName(t *testing.T) {
	dir := t.TempDir()
	rep := sampleReport()
	rep.Connector = ""

	if err := rep.Save(dir); err == nil {
		t.Fatalf("Save() error = nil, want error when Connector is empty")
	}
}
