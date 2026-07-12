package agentloop

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestIncidentFixturesReplay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		file             string
		violationCode    string
		requiredDecision string
		exitClass        string
	}{
		{name: "dead worker", file: "dead_worker.json", violationCode: "WORKER_LIVENESS_LOST", requiredDecision: "retry", exitClass: "retry_required"},
		{name: "false green", file: "false_green.json", violationCode: "VALIDATION_FALSE_GREEN", requiredDecision: "retry", exitClass: "retry_required"},
		{name: "fabricated authority", file: "fabricated_authority.json", violationCode: "AUTHORITY_FABRICATED", requiredDecision: "halt", exitClass: "halt_required"},
		{name: "halt worker survival", file: "halt_worker_survival.json", violationCode: "HALT_REVOCATION_MISSING", requiredDecision: "halt", exitClass: "halt_required"},
		{name: "mega turn", file: "mega_turn.json", violationCode: "TURN_CAP_EXCEEDED", requiredDecision: "halt", exitClass: "halt_required"},
		{name: "dual writer", file: "dual_writer.json", violationCode: "WORKTREE_DUAL_WRITER", requiredDecision: "halt", exitClass: "halt_required"},
		{name: "merge before ratification", file: "merge_before_ratification.json", violationCode: "MERGE_BEFORE_RATIFICATION", requiredDecision: "halt", exitClass: "halt_required"},
		{name: "merge stale attestation", file: "merge_stale_attestation.json", violationCode: "MERGE_ATTESTATION_STALE", requiredDecision: "halt", exitClass: "halt_required"},
		{name: "merge agent authority", file: "merge_agent_authority.json", violationCode: "MERGE_AUTHORITY_DENIED", requiredDecision: "halt", exitClass: "halt_required"},
		{name: "stale verify head", file: "stale_verify_head.json", violationCode: "VERIFY_HEAD_STALE", requiredDecision: "retry", exitClass: "retry_required"},
		{name: "dirty worktree", file: "dirty_worktree.json", violationCode: "WORKTREE_DIRTY", requiredDecision: "retry", exitClass: "retry_required"},
		{name: "interim human wait", file: "interim_human_wait.json", violationCode: "HUMAN_WAIT_PROJECTED_FINAL", requiredDecision: "wait", exitClass: "human_wait_required"},
		{name: "terminal projection mismatch", file: "terminal_projection_mismatch.json", violationCode: "TERMINAL_PROJECTION_MISMATCH", requiredDecision: "halt", exitClass: "halt_required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fixture, err := LoadFixture(filepath.Join("testdata", "incidents", tt.file))
			if err != nil {
				t.Fatalf("LoadFixture() error = %v", err)
			}
			if fixture.Expected.LegacyDecision == fixture.Expected.RequiredDecision {
				t.Fatalf("legacy decision %q does not distinguish fail-open from required result", fixture.Expected.LegacyDecision)
			}

			result, err := Replay(fixture)
			if err != nil {
				t.Fatalf("Replay() error = %v", err)
			}
			if result.ViolationCode != tt.violationCode {
				t.Errorf("violation code = %q, want %q", result.ViolationCode, tt.violationCode)
			}
			if result.RequiredDecision != tt.requiredDecision {
				t.Errorf("required decision = %q, want %q", result.RequiredDecision, tt.requiredDecision)
			}
			if result.RequiredExitClass != tt.exitClass {
				t.Errorf("exit class = %q, want %q", result.RequiredExitClass, tt.exitClass)
			}
			if !result.MatchedExpectation {
				t.Error("matched expectation = false, want true")
			}

			first, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("json.Marshal(first): %v", err)
			}
			second, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("json.Marshal(second): %v", err)
			}
			if !bytes.Equal(first, second) {
				t.Fatalf("replay JSON is not deterministic:\nfirst: %s\nsecond: %s", first, second)
			}
		})
	}
}

func TestLoadFixturesDeterministicOrder(t *testing.T) {
	t.Parallel()

	fixtures, err := LoadFixtures(filepath.Join("testdata", "incidents"))
	if err != nil {
		t.Fatalf("LoadFixtures() error = %v", err)
	}
	if len(fixtures) != 13 {
		t.Fatalf("fixture count = %d, want 13", len(fixtures))
	}

	incidentIDs := make([]string, 0, len(fixtures))
	for _, fixture := range fixtures {
		incidentIDs = append(incidentIDs, fixture.IncidentID)
	}
	if !slices.IsSorted(incidentIDs) {
		t.Fatalf("incident IDs are not deterministic/sorted: %v", incidentIDs)
	}

	results, err := ReplayAll(fixtures)
	if err != nil {
		t.Fatalf("ReplayAll() error = %v", err)
	}
	if len(results) != 13 {
		t.Fatalf("result count = %d, want 13", len(results))
	}
}

func TestLoadFixtureRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	base := mustLoadTestFixture(t, "dead_worker.json")
	tests := []struct {
		name string
		code string
		data func(t *testing.T) []byte
	}{
		{
			name: "unknown field",
			code: "FIXTURE_UNKNOWN_FIELD",
			data: func(t *testing.T) []byte {
				return bytes.Replace(mustJSON(t, base), []byte(`{"schema_version"`), []byte(`{"raw_command":"forbidden","schema_version"`), 1)
			},
		},
		{
			name: "trailing json",
			code: "FIXTURE_TRAILING_DATA",
			data: func(t *testing.T) []byte {
				return append(mustJSON(t, base), []byte(` {}`)...)
			},
		},
		{
			name: "incomplete binding",
			code: "FIXTURE_IDENTITY_INCOMPLETE",
			data: func(t *testing.T) []byte {
				fixture := base
				fixture.Binding.ControllerID = ""
				return mustJSON(t, fixture)
			},
		},
		{
			name: "non synthetic identity",
			code: "FIXTURE_IDENTITY_NOT_SYNTHETIC",
			data: func(t *testing.T) []byte {
				fixture := base
				fixture.Binding.RunID = "run-from-a-real-system"
				return mustJSON(t, fixture)
			},
		},
		{
			name: "non monotonic event",
			code: "FIXTURE_EVENT_NON_MONOTONIC",
			data: func(t *testing.T) []byte {
				fixture := base
				fixture.Events[1].Sequence = fixture.Events[0].Sequence
				return mustJSON(t, fixture)
			},
		},
		{
			name: "event binding mismatch",
			code: "FIXTURE_BINDING_MISMATCH",
			data: func(t *testing.T) []byte {
				fixture := base
				fixture.Events[0].Binding.TurnID = "synthetic:turn:different"
				return mustJSON(t, fixture)
			},
		},
		{
			name: "unknown event kind",
			code: "FIXTURE_EVENT_UNKNOWN",
			data: func(t *testing.T) []byte {
				fixture := base
				fixture.Events[0].Kind = "arbitrary_payload"
				return mustJSON(t, fixture)
			},
		},
		{
			name: "raw prompt",
			code: "FIXTURE_FORBIDDEN_CONTENT",
			data: func(t *testing.T) []byte {
				fixture := base
				fixture.Summary = "raw " + "prompt:" + " do an unbounded action"
				return mustJSON(t, fixture)
			},
		},
		{
			name: "raw session path",
			code: "FIXTURE_FORBIDDEN_CONTENT",
			data: func(t *testing.T) []byte {
				fixture := base
				fixture.Summary = strings.Join([]string{"/tmp", "sessions", "record.jsonl"}, "/")
				return mustJSON(t, fixture)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "fixture.json")
			if err := os.WriteFile(path, tt.data(t), 0o600); err != nil {
				t.Fatalf("os.WriteFile(): %v", err)
			}
			_, err := LoadFixture(path)
			assertValidationCode(t, err, tt.code)
		})
	}
}

func TestLoadFixtureRejectsJSONLBeforeOpening(t *testing.T) {
	t.Parallel()

	_, err := LoadFixture(filepath.Join(t.TempDir(), "missing.jsonl"))
	assertValidationCode(t, err, "FIXTURE_EXTENSION_DENIED")
}

func TestValidateFixtureRejectsSensitiveCanaryBuiltInMemory(t *testing.T) {
	t.Parallel()

	fixture := mustLoadTestFixture(t, "dead_worker.json")
	canaryValue := strings.Join([]string{"synthetic", "canary", "value"}, "-")
	fixture.Summary = strings.Join([]string{"SAMPLE", "TOKEN", canaryValue}, "=")

	err := ValidateFixture(fixture)
	assertValidationCode(t, err, "FIXTURE_SENSITIVE_DATA")
}

func TestReplayRejectsExpectationEcho(t *testing.T) {
	t.Parallel()

	fixture := mustLoadTestFixture(t, "dead_worker.json")
	fixture.Expected.ViolationCode = "WORKTREE_DIRTY"

	_, err := Replay(fixture)
	assertValidationCode(t, err, "FIXTURE_EXPECTATION_MISMATCH")
}

func mustLoadTestFixture(t *testing.T, name string) Fixture {
	t.Helper()
	fixture, err := LoadFixture(filepath.Join("testdata", "incidents", name))
	if err != nil {
		t.Fatalf("LoadFixture(%q): %v", name, err)
	}
	return fixture
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal(): %v", err)
	}
	return data
}

func assertValidationCode(t *testing.T, err error, expected string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want validation code %q", expected)
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error type = %T (%v), want *ValidationError", err, err)
	}
	if validationErr.Code != expected {
		t.Fatalf("validation code = %q, want %q (error: %v)", validationErr.Code, expected, err)
	}
}
