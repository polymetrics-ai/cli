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
		name                    string
		file                    string
		violationCode           string
		requiredDecision        string
		requiredOutcome         string
		exitClass               string
		observedDecisionCorrect bool
		observedOutcomeCorrect  bool
	}{
		{name: "dead worker", file: "dead_worker.json", violationCode: "WORKER_COMPLETION_UNPROVEN", requiredDecision: "retry", requiredOutcome: "correction_preserved_once", exitClass: "retry_required", observedDecisionCorrect: true, observedOutcomeCorrect: true},
		{name: "false green", file: "false_green.json", violationCode: "VALIDATION_FALSE_GREEN", requiredDecision: "retry", requiredOutcome: "correction_preserved_once", exitClass: "retry_required"},
		{name: "fabricated authority", file: "fabricated_authority.json", violationCode: "AUTHORITY_FABRICATED", requiredDecision: "halt", requiredOutcome: "halt_persisted", exitClass: "halt_required", observedDecisionCorrect: true, observedOutcomeCorrect: true},
		{name: "halt worker survival", file: "halt_worker_survival.json", violationCode: "HALT_REVOCATION_MISSING", requiredDecision: "halt", requiredOutcome: "halt_persisted_children_revoked", exitClass: "halt_required", observedDecisionCorrect: true},
		{name: "mega turn", file: "mega_turn.json", violationCode: "TURN_SUPERVISION_EXCEEDED", requiredDecision: "halt", requiredOutcome: "halt_persisted", exitClass: "halt_required"},
		{name: "dual writer", file: "dual_writer.json", violationCode: "WORKTREE_DUAL_WRITER", requiredDecision: "halt", requiredOutcome: "halt_persisted", exitClass: "halt_required"},
		{name: "merge before ratification", file: "merge_before_ratification.json", violationCode: "MERGE_BEFORE_RATIFICATION", requiredDecision: "halt", requiredOutcome: "halt_persisted", exitClass: "halt_required"},
		{name: "merge stale attestation", file: "merge_stale_attestation.json", violationCode: "MERGE_ATTESTATION_STALE", requiredDecision: "halt", requiredOutcome: "halt_persisted", exitClass: "halt_required"},
		{name: "merge agent authority", file: "merge_agent_authority.json", violationCode: "MERGE_AUTHORITY_DENIED", requiredDecision: "halt", requiredOutcome: "halt_persisted", exitClass: "halt_required"},
		{name: "stale verify head", file: "stale_verify_head.json", violationCode: "VERIFY_HEAD_STALE", requiredDecision: "retry", requiredOutcome: "correction_preserved_once", exitClass: "retry_required", observedDecisionCorrect: true, observedOutcomeCorrect: true},
		{name: "dirty worktree", file: "dirty_worktree.json", violationCode: "WORKTREE_DIRTY", requiredDecision: "retry", requiredOutcome: "correction_preserved_once", exitClass: "retry_required", observedDecisionCorrect: true, observedOutcomeCorrect: true},
		{name: "interim human wait", file: "interim_human_wait.json", violationCode: "HUMAN_WAIT_PROJECTED_FINAL", requiredDecision: "wait", requiredOutcome: "human_wait_preserved", exitClass: "human_wait_required", observedDecisionCorrect: true},
		{name: "terminal projection mismatch", file: "terminal_projection_mismatch.json", violationCode: "TERMINAL_PROJECTION_MISMATCH", requiredDecision: "halt", requiredOutcome: "halt_persisted", exitClass: "halt_required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fixture, err := LoadFixture(filepath.Join("testdata", "incidents", tt.file))
			if err != nil {
				t.Fatalf("LoadFixture() error = %v", err)
			}
			if fixture.Expected.ObservedDecisionCorrect == nil || *fixture.Expected.ObservedDecisionCorrect != tt.observedDecisionCorrect {
				t.Errorf("observed decision correct = %v, want %t", fixture.Expected.ObservedDecisionCorrect, tt.observedDecisionCorrect)
			}
			if fixture.Expected.ObservedOutcomeCorrect == nil || *fixture.Expected.ObservedOutcomeCorrect != tt.observedOutcomeCorrect {
				t.Errorf("observed outcome correct = %v, want %t", fixture.Expected.ObservedOutcomeCorrect, tt.observedOutcomeCorrect)
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
			if result.RequiredOutcome != tt.requiredOutcome {
				t.Errorf("required outcome = %q, want %q", result.RequiredOutcome, tt.requiredOutcome)
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

func TestCorrectNegativeVerdictsRemainLoadBearing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		file             string
		requiredDecision string
	}{
		{name: "hard gate halt", file: "fabricated_authority.json", requiredDecision: "halt"},
		{name: "stale head retry", file: "stale_verify_head.json", requiredDecision: "retry"},
		{name: "dirty state retry", file: "dirty_worktree.json", requiredDecision: "retry"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixture := mustLoadTestFixture(t, tt.file)
			if fixture.Expected.ObservedDecision != tt.requiredDecision {
				t.Fatalf("observed decision = %q, want %q", fixture.Expected.ObservedDecision, tt.requiredDecision)
			}
			if fixture.Expected.ObservedDecisionCorrect == nil || !*fixture.Expected.ObservedDecisionCorrect {
				t.Fatal("observed decision was incorrectly labeled wrong")
			}
		})
	}

	haltTurns := make(map[string]struct{})
	for _, fixtureName := range []string{"fabricated_authority.json", "halt_worker_survival.json"} {
		fixture := mustLoadTestFixture(t, fixtureName)
		for _, event := range fixture.Events {
			if (event.Fact.Kind == "validator_decision" && event.Fact.After == "halt") ||
				(event.Fact.Kind == "halt_latch" && event.Fact.After == "halted") {
				haltTurns[event.Binding.TurnID] = struct{}{}
			}
		}
	}
	if len(haltTurns) != 3 {
		t.Fatalf("distinct preserved halt turns = %d, want 3 (%v)", len(haltTurns), haltTurns)
	}
}

func TestDeadWorkerFixtureSeparatesPhantomAndMissingHandoff(t *testing.T) {
	t.Parallel()

	result, err := Replay(mustLoadTestFixture(t, "dead_worker.json"))
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if !slices.Contains(result.ReasonCodes, "PHANTOM_DISPATCH") {
		t.Fatalf("reason codes = %v, want PHANTOM_DISPATCH", result.ReasonCodes)
	}
	if !slices.Contains(result.ReasonCodes, "HANDOFF_MISSING") {
		t.Fatalf("reason codes = %v, want HANDOFF_MISSING", result.ReasonCodes)
	}
}

func TestFalseGreenUsesMissingArtifactAndLaterGateFailureFacts(t *testing.T) {
	t.Parallel()

	fixture := mustLoadTestFixture(t, "false_green.json")
	if !hasFact(fixture, "required_artifact", "missing") {
		t.Fatal("false-green fixture lacks typed missing required artifact fact")
	}
	if !hasFact(fixture, "repo_gate", "failed") {
		t.Fatal("false-green fixture lacks typed later repo-gate failure fact")
	}
}

func TestRepeatedHaltIncludesUnauthorizedResumeAndRepeatedOutcome(t *testing.T) {
	t.Parallel()

	fixture := mustLoadTestFixture(t, "halt_worker_survival.json")
	haltCount := 0
	unauthorizedResume := false
	for _, event := range fixture.Events {
		if event.Fact.Kind == "halt_latch" && event.Fact.After == "halted" {
			haltCount++
		}
		if event.Fact.Kind == "resume_authorization" && event.Fact.After == "requested" && event.ActorID != event.Fact.OwnerID {
			unauthorizedResume = true
		}
	}
	if haltCount != 2 {
		t.Fatalf("halt facts = %d, want 2", haltCount)
	}
	if !unauthorizedResume {
		t.Fatal("fixture lacks unauthorized resume fact")
	}
}

func TestReplayDerivesViolationsFromFacts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		file   string
		mutate func(t *testing.T, fixture *Fixture)
	}{
		{
			name: "missing artifact becomes present",
			file: "false_green.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				factByKind(t, fixture, "required_artifact").After = "present"
			},
		},
		{
			name: "verified head does not move",
			file: "stale_verify_head.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				fact := factByKind(t, fixture, "head_state")
				fact.After = fact.Before
			},
		},
		{
			name: "attested head does not move",
			file: "merge_stale_attestation.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				fact := factByKind(t, fixture, "head_state")
				fact.After = fact.Before
			},
		},
		{
			name: "writer owner is unchanged",
			file: "dual_writer.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				first := -1
				for i := range fixture.Events {
					if fixture.Events[i].Fact.Kind != "writer_lease" {
						continue
					}
					if first == -1 {
						first = i
						continue
					}
					fixture.Events[i].Fact.OwnerID = fixture.Events[first].Fact.OwnerID
					fixture.Events[i].ActorID = fixture.Events[first].ActorID
					return
				}
				t.Fatal("dual writer fixture lacks two lease facts")
			},
		},
		{
			name: "ratification occurs before merge",
			file: "merge_before_ratification.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				fixture.Events[0], fixture.Events[1] = fixture.Events[1], fixture.Events[0]
				resequence(fixture.Events)
			},
		},
		{
			name: "human owns gate transition",
			file: "fabricated_authority.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				for i := range fixture.Events {
					if fixture.Events[i].Fact.Kind == "human_gate" {
						fixture.Events[i].ActorID = fixture.Events[i].Fact.OwnerID
						return
					}
				}
				t.Fatal("authority fixture lacks human_gate fact")
			},
		},
		{
			name: "human owns merge transition",
			file: "merge_agent_authority.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				for i := range fixture.Events {
					if fixture.Events[i].Fact.Kind == "merge_state" {
						fixture.Events[i].ActorID = fixture.Events[i].Fact.OwnerID
						return
					}
				}
				t.Fatal("merge authority fixture lacks merge_state fact")
			},
		},
		{
			name: "turn remains within budget",
			file: "mega_turn.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				factByKind(t, fixture, "turn_budget").After = "within"
			},
		},
		{
			name: "worktree remains clean",
			file: "dirty_worktree.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				factByKind(t, fixture, "worktree_state").After = "clean"
			},
		},
		{
			name: "human wait is projected as wait",
			file: "interim_human_wait.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				factByKind(t, fixture, "terminal_projection").After = "human_wait"
			},
		},
		{
			name: "terminal projection matches durable state",
			file: "terminal_projection_mismatch.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				factByKind(t, fixture, "terminal_projection").After = "human_gate"
			},
		},
		{
			name: "repeated halt is removed",
			file: "halt_worker_survival.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				fixture.Events = fixture.Events[:len(fixture.Events)-1]
			},
		},
		{
			name: "worker execution and handoff are proven",
			file: "dead_worker.json",
			mutate: func(t *testing.T, fixture *Fixture) {
				for i := range fixture.Events {
					switch fixture.Events[i].Fact.Kind {
					case "worker_execution":
						fixture.Events[i].Fact.After = "completed"
					case "worker_handoff":
						fixture.Events[i].Fact.After = "recorded"
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixture := cloneFixture(t, mustLoadTestFixture(t, tt.file))
			tt.mutate(t, &fixture)
			_, err := Replay(fixture)
			assertOneOfValidationCodes(t, err, "FIXTURE_INCIDENT_UNCLASSIFIED", "FIXTURE_EXPECTATION_MISMATCH")
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
		{name: "unknown field", code: "FIXTURE_UNKNOWN_FIELD", data: func(t *testing.T) []byte {
			return bytes.Replace(mustJSON(t, base), []byte(`{"schema_version"`), []byte(`{"raw_command":"forbidden","schema_version"`), 1)
		}},
		{name: "trailing json", code: "FIXTURE_TRAILING_DATA", data: func(t *testing.T) []byte {
			return append(mustJSON(t, base), []byte(` {}`)...)
		}},
		{name: "incomplete binding", code: "FIXTURE_IDENTITY_INCOMPLETE", data: func(t *testing.T) []byte {
			fixture := cloneFixture(t, base)
			fixture.Binding.ControllerID = ""
			return mustJSON(t, fixture)
		}},
		{name: "non synthetic identity", code: "FIXTURE_IDENTITY_NOT_SYNTHETIC", data: func(t *testing.T) []byte {
			fixture := cloneFixture(t, base)
			fixture.Binding.RunID = "run-from-a-real-system"
			return mustJSON(t, fixture)
		}},
		{name: "non monotonic event", code: "FIXTURE_EVENT_NON_MONOTONIC", data: func(t *testing.T) []byte {
			fixture := cloneFixture(t, base)
			fixture.Events[1].Sequence = fixture.Events[0].Sequence
			return mustJSON(t, fixture)
		}},
		{name: "event run binding mismatch", code: "FIXTURE_BINDING_MISMATCH", data: func(t *testing.T) []byte {
			fixture := cloneFixture(t, base)
			fixture.Events[0].Binding.RunID = "synthetic:run:different"
			return mustJSON(t, fixture)
		}},
		{name: "unknown fact kind", code: "FIXTURE_FACT_UNKNOWN", data: func(t *testing.T) []byte {
			fixture := cloneFixture(t, base)
			fixture.Events[0].Fact.Kind = "pre_labeled_conclusion"
			return mustJSON(t, fixture)
		}},
		{name: "raw prompt", code: "FIXTURE_FORBIDDEN_CONTENT", data: func(t *testing.T) []byte {
			fixture := cloneFixture(t, base)
			fixture.Summary = "raw " + "prompt:" + " unbounded content"
			return mustJSON(t, fixture)
		}},
		{name: "raw session path", code: "FIXTURE_FORBIDDEN_CONTENT", data: func(t *testing.T) []byte {
			fixture := cloneFixture(t, base)
			fixture.Summary = strings.Join([]string{"/tmp", "sessions", "record.jsonl"}, "/")
			return mustJSON(t, fixture)
		}},
		{name: "oversized string", code: "FIXTURE_STRING_TOO_LARGE", data: func(t *testing.T) []byte {
			fixture := cloneFixture(t, base)
			fixture.Summary = strings.Repeat("x", 1025)
			return mustJSON(t, fixture)
		}},
		{name: "oversized event count", code: "FIXTURE_EVENT_COUNT_INVALID", data: func(t *testing.T) []byte {
			fixture := cloneFixture(t, base)
			prototype := fixture.Events[0]
			fixture.Events = make([]Event, 65)
			for i := range fixture.Events {
				fixture.Events[i] = prototype
				fixture.Events[i].Sequence = uint64(i + 1)
			}
			return mustJSON(t, fixture)
		}},
		{name: "missing observed decision correctness", code: "FIXTURE_REQUIRED_FIELD_MISSING", data: func(t *testing.T) []byte {
			return bytes.Replace(mustJSON(t, base), []byte(`"observed_decision_correct":true,`), nil, 1)
		}},
		{name: "missing observed outcome correctness", code: "FIXTURE_REQUIRED_FIELD_MISSING", data: func(t *testing.T) []byte {
			return bytes.Replace(mustJSON(t, base), []byte(`"observed_outcome_correct":true,`), nil, 1)
		}},
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

func TestLoadFixtureRejectsOversizedFileBeforeDecode(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "oversized.json")
	if err := os.WriteFile(path, bytes.Repeat([]byte{'x'}, (1<<20)+1), 0o600); err != nil {
		t.Fatalf("os.WriteFile(): %v", err)
	}
	_, err := LoadFixture(path)
	assertValidationCode(t, err, "FIXTURE_TOO_LARGE")
}

func TestFixtureDirectoryBoundsAndIdentity(t *testing.T) {
	t.Parallel()

	base := mustJSON(t, mustLoadTestFixture(t, "dead_worker.json"))

	t.Run("too many fixtures", func(t *testing.T) {
		dir := t.TempDir()
		for i := 0; i < 65; i++ {
			name := filepath.Join(dir, strings.Repeat("x", i/10+1)+string(rune('a'+i%10))+".json")
			if err := os.WriteFile(name, base, 0o600); err != nil {
				t.Fatalf("os.WriteFile(): %v", err)
			}
		}
		_, err := LoadFixtures(dir)
		assertValidationCode(t, err, "FIXTURE_COUNT_INVALID")
	})

	t.Run("duplicate incident ids", func(t *testing.T) {
		dir := t.TempDir()
		for _, name := range []string{"a.json", "b.json"} {
			if err := os.WriteFile(filepath.Join(dir, name), base, 0o600); err != nil {
				t.Fatalf("os.WriteFile(): %v", err)
			}
		}
		_, err := LoadFixtures(dir)
		assertValidationCode(t, err, "FIXTURE_INCIDENT_DUPLICATE")
	})
}

func TestLoadFixtureRejectsSymlinks(t *testing.T) {
	t.Parallel()

	target := filepath.Join("testdata", "incidents", "dead_worker.json")
	link := filepath.Join(t.TempDir(), "linked.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("os.Symlink(): %v", err)
	}
	_, err := LoadFixture(link)
	assertValidationCode(t, err, "FIXTURE_FILE_TYPE_DENIED")

	dir := t.TempDir()
	if err := os.Symlink(target, filepath.Join(dir, "linked.json")); err != nil {
		t.Fatalf("os.Symlink(directory entry): %v", err)
	}
	_, err = LoadFixtures(dir)
	assertValidationCode(t, err, "FIXTURE_FILE_TYPE_DENIED")
}

func TestLoadFixtureRejectsJSONLBeforeOpening(t *testing.T) {
	t.Parallel()
	_, err := LoadFixture(filepath.Join(t.TempDir(), "missing.jsonl"))
	assertValidationCode(t, err, "FIXTURE_EXTENSION_DENIED")
}

func TestSensitiveRejectionDoesNotEchoCanaryOrPath(t *testing.T) {
	t.Parallel()

	fixture := mustLoadTestFixture(t, "dead_worker.json")
	canaryValue := strings.Join([]string{"synthetic", "canary", "value"}, "-")
	fixture.Summary = strings.Join([]string{"SAMPLE", "TOKEN", canaryValue}, "=")
	err := ValidateFixture(fixture)
	assertValidationCode(t, err, "FIXTURE_SENSITIVE_DATA")
	if strings.Contains(err.Error(), canaryValue) {
		t.Fatal("validation error echoed the in-memory canary")
	}

	pathCanary := strings.Join([]string{"synthetic", "private", "path"}, "-")
	_, err = LoadFixture(filepath.Join(t.TempDir(), pathCanary+".json"))
	assertValidationCode(t, err, "FIXTURE_IO")
	if strings.Contains(err.Error(), pathCanary) {
		t.Fatal("fixture I/O error echoed the caller path")
	}
}

func TestReplayRejectsExpectationEcho(t *testing.T) {
	t.Parallel()
	fixture := mustLoadTestFixture(t, "dead_worker.json")
	fixture.Expected.ViolationCode = "WORKTREE_DIRTY"
	_, err := Replay(fixture)
	assertValidationCode(t, err, "FIXTURE_EXPECTATION_MISMATCH")
}

func hasFact(fixture Fixture, kind, after string) bool {
	for _, event := range fixture.Events {
		if event.Fact.Kind == kind && event.Fact.After == after {
			return true
		}
	}
	return false
}

func factByKind(t *testing.T, fixture *Fixture, kind string) *Fact {
	t.Helper()
	for i := range fixture.Events {
		if fixture.Events[i].Fact.Kind == kind {
			return &fixture.Events[i].Fact
		}
	}
	t.Fatalf("fixture lacks fact kind %q", kind)
	return nil
}

func resequence(events []Event) {
	for i := range events {
		events[i].Sequence = uint64(i + 1)
	}
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

func cloneFixture(t *testing.T, fixture Fixture) Fixture {
	t.Helper()
	data := mustJSON(t, fixture)
	var cloned Fixture
	if err := json.Unmarshal(data, &cloned); err != nil {
		t.Fatalf("json.Unmarshal(clone): %v", err)
	}
	return cloned
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

func assertOneOfValidationCodes(t *testing.T, err error, expected ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want one of %v", expected)
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error type = %T (%v), want *ValidationError", err, err)
	}
	if !slices.Contains(expected, validationErr.Code) {
		t.Fatalf("validation code = %q, want one of %v", validationErr.Code, expected)
	}
}
