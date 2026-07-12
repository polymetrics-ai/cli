// Package agentloop provides dependency-free safety and sanitized incident
// replay primitives for autonomous delivery loops.
package agentloop

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	FixtureSchemaVersion = "1.0"
	maxFixtureBytes      = 1 << 20
	maxFixtureCount      = 64
	maxFixtureEvents     = 64
	maxFixtureString     = 1024
)

// Binding identifies a replay event without referring to a real run.
type Binding struct {
	RunID        string `json:"run_id"`
	GenerationID string `json:"generation_id"`
	ControllerID string `json:"controller_id"`
	TurnID       string `json:"turn_id"`
	AttemptID    string `json:"attempt_id"`
	StageID      string `json:"stage_id"`
	TicketID     string `json:"ticket_id"`
	EvidenceID   string `json:"evidence_id"`
	HeadSHA      string `json:"head_sha"`
}

// Fact records one neutral transition for a typed resource and owner.
type Fact struct {
	Kind       string `json:"kind"`
	ResourceID string `json:"resource_id"`
	OwnerID    string `json:"owner_id"`
	Before     string `json:"before"`
	After      string `json:"after"`
}

// Event binds one ordered fact to a complete synthetic identity.
type Event struct {
	Sequence uint64  `json:"sequence"`
	Action   string  `json:"action"`
	ActorID  string  `json:"actor_id"`
	Binding  Binding `json:"binding"`
	Fact     Fact    `json:"fact"`
}

// Expected separates historical observation from required policy. Pointer
// booleans make omitted correctness fields distinguishable from false.
type Expected struct {
	ViolationCode           string `json:"violation_code"`
	ObservedDecision        string `json:"observed_decision"`
	ObservedOutcome         string `json:"observed_outcome"`
	ObservedDecisionCorrect *bool  `json:"observed_decision_correct"`
	ObservedOutcomeCorrect  *bool  `json:"observed_outcome_correct"`
	RequiredDecision        string `json:"required_decision"`
	RequiredOutcome         string `json:"required_outcome"`
	RequiredExitClass       string `json:"required_exit_class"`
}

// Fixture is the strict synthetic incident input contract.
type Fixture struct {
	SchemaVersion string   `json:"schema_version"`
	IncidentID    string   `json:"incident_id"`
	Summary       string   `json:"summary"`
	Binding       Binding  `json:"binding"`
	Events        []Event  `json:"events"`
	Expected      Expected `json:"expected"`
}

// ValidationError exposes a stable code and only a sanitized internal detail.
type ValidationError struct {
	Code string
	Err  error
}

func (e *ValidationError) Error() string {
	if e.Err == nil {
		return e.Code
	}
	return fmt.Sprintf("%s: %v", e.Code, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func validationError(code, detail string) error {
	return &ValidationError{Code: code, Err: errors.New(detail)}
}

// LoadFixture checks extension, type, and size before opening content, then
// strictly decodes and validates one fixture.
func LoadFixture(path string) (Fixture, error) {
	if filepath.Ext(path) != ".json" {
		return Fixture{}, validationError("FIXTURE_EXTENSION_DENIED", "fixture extension is denied")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return Fixture{}, validationError("FIXTURE_IO", "fixture cannot be inspected")
	}
	if !info.Mode().IsRegular() {
		return Fixture{}, validationError("FIXTURE_FILE_TYPE_DENIED", "fixture is not a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return Fixture{}, validationError("FIXTURE_IO", "fixture cannot be opened")
	}
	defer file.Close()
	return loadOpenedFixture(file, info)
}

func loadOpenedFixture(file *os.File, pathInfo os.FileInfo) (Fixture, error) {
	openedInfo, err := file.Stat()
	if err != nil {
		return Fixture{}, validationError("FIXTURE_IO", "opened fixture cannot be inspected")
	}
	if !openedInfo.Mode().IsRegular() || !os.SameFile(pathInfo, openedInfo) {
		return Fixture{}, validationError("FIXTURE_FILE_TYPE_DENIED", "opened fixture identity is denied")
	}
	if openedInfo.Size() > maxFixtureBytes {
		return Fixture{}, validationError("FIXTURE_TOO_LARGE", "fixture exceeds the byte limit")
	}
	return decodeFixture(file)
}

func decodeFixture(reader io.Reader) (Fixture, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxFixtureBytes+1))
	if err != nil {
		return Fixture{}, validationError("FIXTURE_IO", "fixture content cannot be read")
	}
	if len(data) > maxFixtureBytes {
		return Fixture{}, validationError("FIXTURE_TOO_LARGE", "fixture exceeds the byte limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var fixture Fixture
	if err := decoder.Decode(&fixture); err != nil {
		if strings.Contains(err.Error(), "json: unknown field") {
			return Fixture{}, validationError("FIXTURE_UNKNOWN_FIELD", "fixture contains an unknown field")
		}
		return Fixture{}, validationError("FIXTURE_JSON_INVALID", "fixture json is invalid")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Fixture{}, validationError("FIXTURE_TRAILING_DATA", "fixture contains trailing data")
	}
	if err := ValidateFixture(fixture); err != nil {
		return Fixture{}, err
	}
	return fixture, nil
}

// LoadFixtures loads a bounded regular directory in deterministic filename
// order and rejects duplicate incident identities.
func LoadFixtures(dir string) ([]Fixture, error) {
	pathInfo, err := os.Lstat(dir)
	if err != nil {
		return nil, validationError("FIXTURE_IO", "fixture directory cannot be inspected")
	}
	if !pathInfo.IsDir() || pathInfo.Mode()&os.ModeSymlink != 0 {
		return nil, validationError("FIXTURE_FILE_TYPE_DENIED", "fixture directory type is denied")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, validationError("FIXTURE_IO", "fixture directory cannot be opened")
	}
	defer root.Close()
	rootInfo, err := root.Lstat(".")
	if err != nil || !os.SameFile(pathInfo, rootInfo) {
		return nil, validationError("FIXTURE_FILE_TYPE_DENIED", "opened fixture directory identity is denied")
	}
	directory, err := root.Open(".")
	if err != nil {
		return nil, validationError("FIXTURE_IO", "fixture directory cannot be opened")
	}
	defer directory.Close()
	entries, err := directory.ReadDir(-1)
	if err != nil {
		return nil, validationError("FIXTURE_IO", "fixture directory cannot be read")
	}
	if len(entries) == 0 || len(entries) > maxFixtureCount {
		return nil, validationError("FIXTURE_COUNT_INVALID", "fixture count is outside the allowed range")
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		entryInfo, err := root.Lstat(entry.Name())
		if err != nil {
			return nil, validationError("FIXTURE_IO", "fixture entry cannot be inspected")
		}
		if entryInfo.Mode()&os.ModeSymlink != 0 {
			return nil, validationError("FIXTURE_FILE_TYPE_DENIED", "fixture directory contains a denied entry type")
		}
		if !entryInfo.Mode().IsRegular() {
			return nil, validationError("FIXTURE_FILE_TYPE_DENIED", "fixture directory contains a denied entry type")
		}
		if filepath.Ext(entry.Name()) != ".json" {
			return nil, validationError("FIXTURE_EXTENSION_DENIED", "fixture directory contains a denied extension")
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	fixtures := make([]Fixture, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		pathEntryInfo, err := root.Lstat(name)
		if err != nil || pathEntryInfo.Mode()&os.ModeSymlink != 0 {
			return nil, validationError("FIXTURE_FILE_TYPE_DENIED", "fixture entry identity changed")
		}
		file, err := root.Open(name)
		if err != nil {
			return nil, validationError("FIXTURE_IO", "fixture entry cannot be opened")
		}
		fixture, loadErr := loadOpenedFixture(file, pathEntryInfo)
		closeErr := file.Close()
		if loadErr != nil {
			return nil, loadErr
		}
		if closeErr != nil {
			return nil, validationError("FIXTURE_IO", "fixture entry cannot be closed")
		}
		if _, exists := seen[fixture.IncidentID]; exists {
			return nil, validationError("FIXTURE_INCIDENT_DUPLICATE", "fixture incident id is duplicated")
		}
		seen[fixture.IncidentID] = struct{}{}
		fixtures = append(fixtures, fixture)
	}
	return fixtures, nil
}

// ValidateFixture enforces synthetic identity, typed fact, ordering, bound,
// observation, and sensitive-content invariants.
func ValidateFixture(fixture Fixture) error {
	if fixture.SchemaVersion != FixtureSchemaVersion {
		return validationError("FIXTURE_SCHEMA_UNSUPPORTED", "fixture schema version is unsupported")
	}
	if err := validateSyntheticID(fixture.IncidentID); err != nil {
		return err
	}
	if _, ok := canonicalViolationForIncident(fixture.IncidentID); !ok {
		return validationError("FIXTURE_INCIDENT_UNKNOWN", "fixture incident id is outside the replay corpus")
	}
	if fixture.Summary == "" {
		return validationError("FIXTURE_SUMMARY_MISSING", "fixture summary is required")
	}
	if err := validateBinding(fixture.Binding); err != nil {
		return err
	}
	if len(fixture.Events) == 0 || len(fixture.Events) > maxFixtureEvents {
		return validationError("FIXTURE_EVENT_COUNT_INVALID", "fixture event count is outside the allowed range")
	}
	for i, event := range fixture.Events {
		if event.Sequence != uint64(i+1) {
			return validationError("FIXTURE_EVENT_NON_MONOTONIC", "fixture event sequence is not contiguous")
		}
		if event.Action != "transition" {
			return validationError("FIXTURE_ACTION_UNKNOWN", "fixture event action is unknown")
		}
		if err := validateSyntheticID(event.ActorID); err != nil {
			return err
		}
		if err := validateBinding(event.Binding); err != nil {
			return err
		}
		if !sameRunBinding(fixture.Binding, event.Binding) {
			return validationError("FIXTURE_BINDING_MISMATCH", "fixture event belongs to a different run binding")
		}
		if err := validateFact(event.Fact); err != nil {
			return err
		}
	}
	if err := validateExpected(fixture.Expected); err != nil {
		return err
	}
	if err := validateFixtureStrings(fixture); err != nil {
		return err
	}
	return nil
}

func canonicalViolationForIncident(incidentID string) (string, bool) {
	switch incidentID {
	case "synthetic:incident:dead_worker":
		return "WORKER_COMPLETION_UNPROVEN", true
	case "synthetic:incident:false_green":
		return "VALIDATION_FALSE_GREEN", true
	case "synthetic:incident:fabricated_authority":
		return "AUTHORITY_FABRICATED", true
	case "synthetic:incident:halt_worker_survival":
		return "HALT_REVOCATION_MISSING", true
	case "synthetic:incident:mega_turn":
		return "TURN_SUPERVISION_EXCEEDED", true
	case "synthetic:incident:dual_writer":
		return "WORKTREE_DUAL_WRITER", true
	case "synthetic:incident:merge_before_ratification":
		return "MERGE_BEFORE_RATIFICATION", true
	case "synthetic:incident:merge_stale_attestation":
		return "MERGE_ATTESTATION_STALE", true
	case "synthetic:incident:merge_agent_authority":
		return "MERGE_AUTHORITY_DENIED", true
	case "synthetic:incident:stale_verify_head":
		return "VERIFY_HEAD_STALE", true
	case "synthetic:incident:dirty_worktree":
		return "WORKTREE_DIRTY", true
	case "synthetic:incident:interim_human_wait":
		return "HUMAN_WAIT_PROJECTED_FINAL", true
	case "synthetic:incident:terminal_projection_mismatch":
		return "TERMINAL_PROJECTION_MISMATCH", true
	default:
		return "", false
	}
}

func validateBinding(binding Binding) error {
	values := []string{
		binding.RunID,
		binding.GenerationID,
		binding.ControllerID,
		binding.TurnID,
		binding.AttemptID,
		binding.StageID,
		binding.TicketID,
		binding.EvidenceID,
		binding.HeadSHA,
	}
	for _, value := range values {
		if value == "" {
			return validationError("FIXTURE_IDENTITY_INCOMPLETE", "fixture identity binding is incomplete")
		}
		if err := validateSyntheticID(value); err != nil {
			return err
		}
	}
	return nil
}

func sameRunBinding(root, event Binding) bool {
	return root.RunID == event.RunID
}

func validateSyntheticID(value string) error {
	if value == "" {
		return validationError("FIXTURE_IDENTITY_INCOMPLETE", "fixture identity is missing")
	}
	if !strings.HasPrefix(value, "synthetic:") {
		return validationError("FIXTURE_IDENTITY_NOT_SYNTHETIC", "fixture identity is outside the synthetic namespace")
	}
	if len(value) > maxFixtureString {
		return validationError("FIXTURE_STRING_TOO_LARGE", "fixture identity exceeds the string limit")
	}
	parts := strings.Split(value, ":")
	if len(parts) < 2 || parts[0] != "synthetic" {
		return validationError("FIXTURE_IDENTITY_INVALID", "fixture identity grammar is invalid")
	}
	for _, part := range parts[1:] {
		if part == "" {
			return validationError("FIXTURE_IDENTITY_INVALID", "fixture identity grammar is invalid")
		}
		for i := 0; i < len(part); i++ {
			if !isSyntheticIDByte(part[i]) {
				return validationError("FIXTURE_IDENTITY_INVALID", "fixture identity grammar is invalid")
			}
		}
	}
	return nil
}

func isSyntheticIDByte(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= '0' && value <= '9' || value == '_' || value == '-' || value == '.'
}

func validateFact(fact Fact) error {
	if !isKnownFactKind(fact.Kind) {
		return validationError("FIXTURE_FACT_UNKNOWN", "fixture fact kind is unknown")
	}
	if err := validateSyntheticID(fact.ResourceID); err != nil {
		return err
	}
	if err := validateSyntheticID(fact.OwnerID); err != nil {
		return err
	}
	if err := validateFactValue(fact.Before); err != nil {
		return err
	}
	if err := validateFactValue(fact.After); err != nil {
		return err
	}
	return nil
}

func validateFactValue(value string) error {
	if strings.HasPrefix(value, "synthetic:") {
		return validateSyntheticID(value)
	}
	if !isKnownFactValue(value) {
		return validationError("FIXTURE_FACT_VALUE_UNKNOWN", "fixture fact value is unknown")
	}
	return nil
}

func validateExpected(expected Expected) error {
	if expected.ViolationCode == "" || expected.ObservedDecision == "" || expected.ObservedOutcome == "" ||
		expected.RequiredDecision == "" || expected.RequiredOutcome == "" || expected.RequiredExitClass == "" ||
		expected.ObservedDecisionCorrect == nil || expected.ObservedOutcomeCorrect == nil {
		return validationError("FIXTURE_REQUIRED_FIELD_MISSING", "fixture expectation is incomplete")
	}
	if !isKnownObservedDecision(expected.ObservedDecision) ||
		!isKnownObservedOutcome(expected.ObservedOutcome) ||
		!isKnownRequiredDecision(expected.RequiredDecision) ||
		!isKnownRequiredOutcome(expected.RequiredOutcome) ||
		!isKnownRequiredExitClass(expected.RequiredExitClass) {
		return validationError("FIXTURE_EXPECTATION_VALUE_UNKNOWN", "fixture expectation contains an unknown value")
	}
	decisionCorrect := expected.ObservedDecision == expected.RequiredDecision
	outcomeCorrect := expected.ObservedOutcome == expected.RequiredOutcome
	if *expected.ObservedDecisionCorrect != decisionCorrect || *expected.ObservedOutcomeCorrect != outcomeCorrect {
		return validationError("FIXTURE_EXPECTATION_INVALID", "fixture observation correctness is inconsistent")
	}
	return nil
}

func isKnownObservedDecision(value string) bool {
	switch value {
	case "done", "halt", "merged", "proceed", "retry", "wait":
		return true
	default:
		return false
	}
}

func isKnownObservedOutcome(value string) bool {
	switch value {
	case "agent_merge_completed", "correction_preserved_once", "detached_progress",
		"driver_exited_children_alive", "driver_exited_without_latch", "halt_persisted", "merged_before_ratification",
		"overlapping_mutation", "proceeded_then_repo_gate_failed", "projected_done",
		"divergent_terminal_state_accepted", "stale_proceed_after_merge", "unproven_completion_accepted":
		return true
	default:
		return false
	}
}

func isKnownRequiredDecision(value string) bool {
	switch value {
	case "halt", "retry", "wait":
		return true
	default:
		return false
	}
}

func isKnownRequiredOutcome(value string) bool {
	switch value {
	case "correction_preserved_once", "halt_persisted", "halt_persisted_children_revoked", "human_wait_preserved":
		return true
	default:
		return false
	}
}

func isKnownRequiredExitClass(value string) bool {
	switch value {
	case "halt_required", "human_wait_required", "retry_required":
		return true
	default:
		return false
	}
}

func isKnownFactKind(kind string) bool {
	switch kind {
	case "attestation_binding", "halt_latch", "head_state", "human_gate", "human_wait",
		"merge_authority", "merge_state", "ratification_state", "repo_gate",
		"required_artifact", "resume_authorization", "stage_decision", "stage_state", "terminal_projection",
		"terminal_state", "turn_budget", "turn_supervision", "validation_binding",
		"validator_decision", "verification_binding", "worker_dispatch", "worker_execution",
		"worker_handoff", "worker_liveness", "worktree_state", "writer_lease":
		return true
	default:
		return false
	}
}

func isKnownFactValue(value string) bool {
	switch value {
	case "absent", "active", "agent_requested", "blocked", "clean", "cleared", "completed", "denied",
		"detached", "dirty", "done", "exceeded", "failed", "halt", "halted", "human_gate",
		"human_ready", "human_required", "human_wait", "inactive", "merged", "missing", "mutated", "none",
		"not_requested", "not_started", "open", "pending", "present", "proceed", "recorded", "running",
		"requested", "retry", "supervised", "unowned", "within":
		return true
	default:
		return false
	}
}
