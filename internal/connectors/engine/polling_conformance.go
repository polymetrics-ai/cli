package engine

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"polymetrics.ai/internal/synccontract"
)

const pollingWatermarkConformanceMechanism = "polling_watermark"

// pollingWatermarkConformanceV1Digest locks the immutable v1 fixture bytes.
// New cases require a new versioned corpus rather than editing this file.
const pollingWatermarkConformanceV1Digest = "3acf1e9bf13615c5355cc305a705cdcddec5d08ab80ece8024459860bb03e1a4"

//go:embed testdata/polling_watermark_conformance/v1.json
var embeddedPollingWatermarkConformanceV1 []byte

// PollingWatermarkConformancePosition is the complete ordered tuple sent to a
// polling source. A scalar cursor is intentionally not an alternative shape.
type PollingWatermarkConformancePosition struct {
	Watermark  string `json:"watermark"`
	TieBreaker string `json:"tie_breaker"`
}

// PollingWatermarkConformanceDescriptor is the descriptor-shaped fixture input
// used by the corpus. It describes only bounded, declaration-owned values.
type PollingWatermarkConformanceDescriptor struct {
	Mechanism         string `json:"mechanism"`
	ExecutorID        string `json:"executor_id"`
	SourceEngine      string `json:"source_engine"`
	SourceAccount     string `json:"source_account"`
	SourceScope       string `json:"source_scope"`
	SourceGeneration  string `json:"source_generation"`
	SchemaFingerprint string `json:"schema_fingerprint"`
	StableKeyset      bool   `json:"stable_keyset"`
	CursorPolicy      string `json:"cursor_policy"`
	BoundedOverlap    bool   `json:"bounded_overlap"`
	BoundedCommitLag  bool   `json:"bounded_commit_lag"`
	DeleteVisibility  string `json:"delete_visibility"`
}

// PollingWatermarkConformanceCheckpoint is a fixture-friendly projection of a
// #3810 checkpoint. Lanes materialize it as a CheckpointEnvelope and the
// runner verifies persisted envelopes rather than strings.
type PollingWatermarkConformanceCheckpoint struct {
	PollingWatermarkConformancePosition
	SourceGeneration  string `json:"source_generation"`
	SchemaFingerprint string `json:"schema_fingerprint"`
	Committed         bool   `json:"committed"`
}

// PollingWatermarkConformanceRecord is one deterministic source record.
// HardDelete marks an intentionally unobservable deletion; Tombstone marks a
// declared soft-delete record that must close history instead of being removed.
type PollingWatermarkConformanceRecord struct {
	StableIdentity string `json:"stable_identity"`
	Watermark      string `json:"watermark"`
	TieBreaker     string `json:"tie_breaker"`
	Tombstone      bool   `json:"tombstone,omitempty"`
	HardDelete     bool   `json:"hard_delete,omitempty"`
}

// PollingWatermarkConformanceCursorSample retains one raw source value so a
// lane proves NULL, precision, and coercion policy without float conversion.
type PollingWatermarkConformanceCursorSample struct {
	ID         string          `json:"id"`
	Policy     string          `json:"policy"`
	Field      string          `json:"field"`
	Encoding   string          `json:"encoding"`
	Value      json.RawMessage `json:"value"`
	Accepted   bool            `json:"accepted"`
	ExactValue string          `json:"exact_value,omitempty"`
}

// PollingWatermarkConformancePage is one physical source page.
type PollingWatermarkConformancePage struct {
	Records []PollingWatermarkConformanceRecord `json:"records"`
	More    bool                                `json:"more"`
}

// PollingWatermarkConformanceAcknowledgement controls deterministic downstream
// acknowledgement and checkpoint-persistence behavior for a fixture run.
type PollingWatermarkConformanceAcknowledgement struct {
	Durable bool `json:"durable"`
	Persist bool `json:"persist"`
}

// PollingWatermarkConformanceExpectation is the behavior and state contract
// the no-skip runner validates for every immutable fixture.
type PollingWatermarkConformanceExpectation struct {
	Outcome                      string                                `json:"outcome"`
	StableIdentities             []string                              `json:"stable_identities"`
	SourceRequests               []PollingWatermarkConformancePosition `json:"source_requests"`
	PersistedCheckpointCount     int                                   `json:"persisted_checkpoint_count"`
	PersistedCheckpointPositions []PollingWatermarkConformancePosition `json:"persisted_checkpoint_positions"`
	ReplayedFromPriorCommitted   bool                                  `json:"replayed_from_prior_committed,omitempty"`
	CheckpointUnchanged          bool                                  `json:"checkpoint_unchanged,omitempty"`
	RecoveryOutcome              string                                `json:"recovery_outcome,omitempty"`
	TombstoneIdentities          []string                              `json:"tombstone_identities,omitempty"`
	HistoryClosedIdentities      []string                              `json:"history_closed_identities,omitempty"`
	HardDeleteInvisible          bool                                  `json:"hard_delete_invisible,omitempty"`
	AdmissionRejected            bool                                  `json:"admission_rejected,omitempty"`
}

// PollingWatermarkConformanceFixture contains the complete deterministic input
// and expected state for one immutable polling conformance scenario.
type PollingWatermarkConformanceFixture struct {
	ID                string                                       `json:"id"`
	Scenario          string                                       `json:"scenario"`
	Descriptor        PollingWatermarkConformanceDescriptor        `json:"descriptor"`
	InitialCheckpoint *PollingWatermarkConformanceCheckpoint       `json:"initial_checkpoint,omitempty"`
	Pages             []PollingWatermarkConformancePage            `json:"pages"`
	CursorSamples     []PollingWatermarkConformanceCursorSample    `json:"cursor_samples,omitempty"`
	Acknowledgements  []PollingWatermarkConformanceAcknowledgement `json:"acknowledgements"`
	Expected          PollingWatermarkConformanceExpectation       `json:"expected"`
}

type pollingWatermarkConformanceCorpus struct {
	Version  uint                                 `json:"version"`
	Fixtures []PollingWatermarkConformanceFixture `json:"fixtures"`
}

var requiredPollingWatermarkConformanceScenarios = map[string]string{
	"equal-watermark-page-split-recovery":              "equal_watermark_page_split_recovery",
	"empty-page-does-not-advance":                      "empty_page_no_advance",
	"non-advancing-page-is-rejected":                   "non_advancing_page_rejected",
	"null-precision-coercion-is-rejected":              "null_precision_coercion_rejected",
	"unstable-or-non-unique-keyset-is-rejected":        "unstable_keyset_rejected",
	"bounded-overlap-and-commit-lag":                   "bounded_overlap_commit_lag_safe",
	"source-generation-mismatch-requires-rebootstrap":  "source_generation_mismatch",
	"schema-fingerprint-mismatch-requires-rebootstrap": "schema_fingerprint_mismatch",
	"acknowledged-before-checkpoint-replays":           "acknowledged_checkpoint_failure_replays",
	"tombstone-history-and-hard-delete-visibility":     "tombstone_history_hard_delete_invisibility",
	"missing-executor-or-evidence-is-rejected":         "missing_executor_or_evidence_rejected",
}

var (
	loadPollingWatermarkConformanceOnce sync.Once
	loadedPollingWatermarkConformance   pollingWatermarkConformanceCorpus
)

func mustPollingWatermarkConformanceCorpus() pollingWatermarkConformanceCorpus {
	loadPollingWatermarkConformanceOnce.Do(func() {
		digest := sha256.Sum256(embeddedPollingWatermarkConformanceV1)
		if got := hex.EncodeToString(digest[:]); got != pollingWatermarkConformanceV1Digest {
			panic(fmt.Sprintf("polling watermark conformance v1 digest = %s, want %s", got, pollingWatermarkConformanceV1Digest))
		}
		if err := json.Unmarshal(embeddedPollingWatermarkConformanceV1, &loadedPollingWatermarkConformance); err != nil {
			panic(fmt.Sprintf("parse embedded polling watermark conformance corpus: %v", err))
		}
		if err := validatePollingWatermarkConformanceCorpus(loadedPollingWatermarkConformance); err != nil {
			panic(fmt.Sprintf("embedded polling watermark conformance corpus: %v", err))
		}
	})
	return loadedPollingWatermarkConformance
}

// PollingWatermarkConformanceFixtures returns defensive copies of every v1
// fixture. The runner does not accept an alternate fixture list or skip input.
func PollingWatermarkConformanceFixtures() []PollingWatermarkConformanceFixture {
	corpus := mustPollingWatermarkConformanceCorpus()
	fixtures := make([]PollingWatermarkConformanceFixture, len(corpus.Fixtures))
	for i := range corpus.Fixtures {
		fixtures[i] = clonePollingWatermarkConformanceFixture(corpus.Fixtures[i])
	}
	sort.Slice(fixtures, func(i, j int) bool { return fixtures[i].ID < fixtures[j].ID })
	return fixtures
}

// PollingWatermarkConformanceEvidence identifies the exact immutable polling
// corpus that an executor/descriptor has passed. It is intentionally separate
// from synccontract's generic #3810 corpus and evidence digest.
type PollingWatermarkConformanceEvidence struct {
	FixtureVersion uint     `json:"fixture_version"`
	FixtureDigest  string   `json:"fixture_digest"`
	FixtureIDs     []string `json:"fixture_ids"`
}

// RequiredPollingWatermarkConformanceEvidence returns a defensive copy of the
// only evidence accepted by the v1 no-skip runner.
func RequiredPollingWatermarkConformanceEvidence() PollingWatermarkConformanceEvidence {
	corpus := mustPollingWatermarkConformanceCorpus()
	digest := sha256.Sum256(embeddedPollingWatermarkConformanceV1)
	ids := make([]string, 0, len(corpus.Fixtures))
	for _, fixture := range corpus.Fixtures {
		ids = append(ids, fixture.ID)
	}
	sort.Strings(ids)
	return PollingWatermarkConformanceEvidence{
		FixtureVersion: corpus.Version,
		FixtureDigest:  hex.EncodeToString(digest[:]),
		FixtureIDs:     ids,
	}
}

func (e PollingWatermarkConformanceEvidence) clone() PollingWatermarkConformanceEvidence {
	e.FixtureIDs = append([]string(nil), e.FixtureIDs...)
	return e
}

func (e PollingWatermarkConformanceEvidence) matchesRequired() bool {
	required := RequiredPollingWatermarkConformanceEvidence()
	if e.FixtureVersion != required.FixtureVersion || e.FixtureDigest != required.FixtureDigest || len(e.FixtureIDs) != len(required.FixtureIDs) {
		return false
	}
	got := append([]string(nil), e.FixtureIDs...)
	want := append([]string(nil), required.FixtureIDs...)
	sort.Strings(got)
	sort.Strings(want)
	return reflect.DeepEqual(got, want)
}

// PollingWatermarkConformanceRegistration is an opaque proof that a lane has
// an exact polling descriptor and exact v1 evidence. Its zero value is not a
// registered lane, so a caller cannot accidentally admit an executor by
// struct literal.
type PollingWatermarkConformanceRegistration struct {
	descriptor PollingWatermarkConformanceDescriptor
	evidence   PollingWatermarkConformanceEvidence
	registered bool
}

var (
	// ErrPollingWatermarkConformanceUnregistered rejects a lane that has no
	// exact polling descriptor registration.
	ErrPollingWatermarkConformanceUnregistered = errors.New("polling watermark conformance lane is not registered")
	// ErrPollingWatermarkConformanceEvidence rejects incomplete, stale, or
	// alternate fixture evidence before a lane gets any scenario.
	ErrPollingWatermarkConformanceEvidence = errors.New("polling watermark conformance evidence does not match the immutable corpus")
)

// NewPollingWatermarkConformanceRegistration constructs the only registration
// accepted by RunPollingWatermarkConformanceSuite.
func NewPollingWatermarkConformanceRegistration(descriptor PollingWatermarkConformanceDescriptor, evidence PollingWatermarkConformanceEvidence) (PollingWatermarkConformanceRegistration, error) {
	registration := PollingWatermarkConformanceRegistration{
		descriptor: descriptor,
		evidence:   evidence.clone(),
		registered: true,
	}
	if err := registration.validate(); err != nil {
		return PollingWatermarkConformanceRegistration{}, err
	}
	return registration, nil
}

// Descriptor returns the lane's immutable descriptor-shaped declaration.
func (r PollingWatermarkConformanceRegistration) Descriptor() PollingWatermarkConformanceDescriptor {
	return r.descriptor
}

// Evidence returns a defensive copy of the lane's immutable corpus evidence.
func (r PollingWatermarkConformanceRegistration) Evidence() PollingWatermarkConformanceEvidence {
	return r.evidence.clone()
}

func (r PollingWatermarkConformanceRegistration) validate() error {
	if !r.registered {
		return ErrPollingWatermarkConformanceUnregistered
	}
	if err := validatePollingWatermarkConformanceRegistrationDescriptor(r.descriptor); err != nil {
		return fmt.Errorf("%w: %v", ErrPollingWatermarkConformanceUnregistered, err)
	}
	if !r.evidence.matchesRequired() {
		return ErrPollingWatermarkConformanceEvidence
	}
	return nil
}

// PollingWatermarkConformanceLaneFactory makes one isolated registered lane
// per fixture. The runner owns fixture enumeration and exposes no skip/filter
// input, keeping core scenarios outside lane control.
type PollingWatermarkConformanceLaneFactory interface {
	NewPollingWatermarkConformanceLane(context.Context) (PollingWatermarkConformanceLane, error)
}

// PollingWatermarkConformanceLane is the narrow test seam future polling
// executors use to prove their behavior against the immutable corpus.
type PollingWatermarkConformanceLane interface {
	PollingWatermarkConformanceRegistration() PollingWatermarkConformanceRegistration
	RunPollingWatermarkConformance(context.Context, PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error)
}

// PollingWatermarkConformanceObservation records behavior and state observed
// from one fixture run. Persisted checkpoints use the real #3810 envelope.
type PollingWatermarkConformanceObservation struct {
	FixtureID                  string                                          `json:"fixture_id"`
	Outcome                    string                                          `json:"outcome"`
	StableIdentities           []string                                        `json:"stable_identities"`
	SourceRequests             []PollingWatermarkConformancePosition           `json:"source_requests"`
	CursorSampleResults        []PollingWatermarkConformanceCursorSampleResult `json:"cursor_sample_results,omitempty"`
	PersistedCheckpoints       []synccontract.CheckpointEnvelope               `json:"persisted_checkpoints"`
	ReplayedFromPriorCommitted bool                                            `json:"replayed_from_prior_committed,omitempty"`
	CheckpointUnchanged        bool                                            `json:"checkpoint_unchanged,omitempty"`
	RecoveryOutcome            synccontract.RecoveryOutcome                    `json:"recovery_outcome,omitempty"`
	RecoveryError              error                                           `json:"-"`
	TombstoneIdentities        []string                                        `json:"tombstone_identities,omitempty"`
	HistoryClosedIdentities    []string                                        `json:"history_closed_identities,omitempty"`
	HardDeleteInvisible        bool                                            `json:"hard_delete_invisible,omitempty"`
	AdmissionRejected          bool                                            `json:"admission_rejected,omitempty"`
}

// PollingWatermarkConformanceCursorSampleResult reports whether a lane kept a
// raw cursor sample exactly or refused it before coercion.
type PollingWatermarkConformanceCursorSampleResult struct {
	ID         string `json:"id"`
	Accepted   bool   `json:"accepted"`
	ExactValue string `json:"exact_value,omitempty"`
}

func (o PollingWatermarkConformanceObservation) clone() PollingWatermarkConformanceObservation {
	o.StableIdentities = append([]string(nil), o.StableIdentities...)
	o.SourceRequests = append([]PollingWatermarkConformancePosition(nil), o.SourceRequests...)
	o.CursorSampleResults = append([]PollingWatermarkConformanceCursorSampleResult(nil), o.CursorSampleResults...)
	if o.PersistedCheckpoints != nil {
		checkpoints := make([]synccontract.CheckpointEnvelope, len(o.PersistedCheckpoints))
		for i := range o.PersistedCheckpoints {
			checkpoints[i] = o.PersistedCheckpoints[i].Clone()
		}
		o.PersistedCheckpoints = checkpoints
	}
	o.TombstoneIdentities = append([]string(nil), o.TombstoneIdentities...)
	o.HistoryClosedIdentities = append([]string(nil), o.HistoryClosedIdentities...)
	return o
}

// PollingWatermarkConformanceSuiteReport records every executed immutable ID
// and its independently cloned observation.
type PollingWatermarkConformanceSuiteReport struct {
	ExecutedFixtureIDs []string                                 `json:"executed_fixture_ids"`
	Observations       []PollingWatermarkConformanceObservation `json:"observations"`
}

// Fixture returns a defensive observation for one executed fixture ID.
func (r PollingWatermarkConformanceSuiteReport) Fixture(id string) (PollingWatermarkConformanceObservation, bool) {
	for _, observation := range r.Observations {
		if observation.FixtureID == id {
			return observation.clone(), true
		}
	}
	return PollingWatermarkConformanceObservation{}, false
}

// RunPollingWatermarkConformanceSuite executes every immutable v1 fixture with
// a fresh registered lane. It deliberately accepts no fixture, filter, or skip
// argument: lanes cannot select, replace, or weaken core scenarios.
func RunPollingWatermarkConformanceSuite(ctx context.Context, factory PollingWatermarkConformanceLaneFactory) (PollingWatermarkConformanceSuiteReport, error) {
	if ctx == nil {
		return PollingWatermarkConformanceSuiteReport{}, errors.New("polling watermark conformance context is required")
	}
	if factory == nil {
		return PollingWatermarkConformanceSuiteReport{}, errors.New("polling watermark conformance lane factory is required")
	}

	fixtures := PollingWatermarkConformanceFixtures()
	report := PollingWatermarkConformanceSuiteReport{
		ExecutedFixtureIDs: make([]string, 0, len(fixtures)),
		Observations:       make([]PollingWatermarkConformanceObservation, 0, len(fixtures)),
	}
	for _, fixture := range fixtures {
		if err := ctx.Err(); err != nil {
			return PollingWatermarkConformanceSuiteReport{}, err
		}
		lane, err := factory.NewPollingWatermarkConformanceLane(ctx)
		if err != nil {
			return PollingWatermarkConformanceSuiteReport{}, fmt.Errorf("create lane for fixture %q: %w", fixture.ID, err)
		}
		if isNilPollingWatermarkConformanceLane(lane) {
			return PollingWatermarkConformanceSuiteReport{}, fmt.Errorf("fixture %q: %w", fixture.ID, ErrPollingWatermarkConformanceUnregistered)
		}
		if err := lane.PollingWatermarkConformanceRegistration().validate(); err != nil {
			return PollingWatermarkConformanceSuiteReport{}, fmt.Errorf("fixture %q: %w", fixture.ID, err)
		}
		observation, err := lane.RunPollingWatermarkConformance(ctx, clonePollingWatermarkConformanceFixture(fixture))
		if err != nil {
			return PollingWatermarkConformanceSuiteReport{}, fmt.Errorf("run fixture %q: %w", fixture.ID, err)
		}
		if err := validatePollingWatermarkConformanceObservation(fixture, observation); err != nil {
			return PollingWatermarkConformanceSuiteReport{}, fmt.Errorf("fixture %q: %w", fixture.ID, err)
		}
		report.ExecutedFixtureIDs = append(report.ExecutedFixtureIDs, fixture.ID)
		report.Observations = append(report.Observations, observation.clone())
	}
	if !reflect.DeepEqual(report.ExecutedFixtureIDs, RequiredPollingWatermarkConformanceEvidence().FixtureIDs) {
		return PollingWatermarkConformanceSuiteReport{}, errors.New("polling watermark conformance runner did not execute the complete immutable fixture set")
	}
	return report, nil
}

func isNilPollingWatermarkConformanceLane(lane PollingWatermarkConformanceLane) bool {
	if lane == nil {
		return true
	}
	value := reflect.ValueOf(lane)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func validatePollingWatermarkConformanceCorpus(corpus pollingWatermarkConformanceCorpus) error {
	if corpus.Version != 1 {
		return fmt.Errorf("unsupported corpus version %d", corpus.Version)
	}
	if len(corpus.Fixtures) != len(requiredPollingWatermarkConformanceScenarios) {
		return fmt.Errorf("fixture count = %d, want %d", len(corpus.Fixtures), len(requiredPollingWatermarkConformanceScenarios))
	}
	seen := make(map[string]struct{}, len(corpus.Fixtures))
	for _, fixture := range corpus.Fixtures {
		if wantScenario, ok := requiredPollingWatermarkConformanceScenarios[fixture.ID]; !ok {
			return fmt.Errorf("fixture %q is not a required v1 fixture", fixture.ID)
		} else if fixture.Scenario != wantScenario {
			return fmt.Errorf("fixture %q scenario = %q, want %q", fixture.ID, fixture.Scenario, wantScenario)
		}
		if _, duplicate := seen[fixture.ID]; duplicate {
			return fmt.Errorf("fixture ID %q is duplicated", fixture.ID)
		}
		seen[fixture.ID] = struct{}{}
		if err := validatePollingWatermarkConformanceFixture(fixture); err != nil {
			return fmt.Errorf("fixture %q: %w", fixture.ID, err)
		}
	}
	return nil
}

func validatePollingWatermarkConformanceFixture(fixture PollingWatermarkConformanceFixture) error {
	if strings.TrimSpace(fixture.ID) == "" || strings.TrimSpace(fixture.Scenario) == "" {
		return errors.New("ID and scenario are required")
	}
	if err := validatePollingWatermarkConformanceDescriptor(fixture.Descriptor); err != nil {
		return err
	}
	if strings.TrimSpace(fixture.Expected.Outcome) == "" {
		return errors.New("expected outcome is required")
	}
	if fixture.Expected.PersistedCheckpointCount < 0 {
		return errors.New("expected persisted checkpoint count cannot be negative")
	}
	if len(fixture.Expected.PersistedCheckpointPositions) != fixture.Expected.PersistedCheckpointCount {
		return fmt.Errorf("expected persisted checkpoint positions = %d, want %d", len(fixture.Expected.PersistedCheckpointPositions), fixture.Expected.PersistedCheckpointCount)
	}
	for _, record := range fixturePagesRecords(fixture.Pages) {
		if strings.TrimSpace(record.StableIdentity) == "" || strings.TrimSpace(record.Watermark) == "" || strings.TrimSpace(record.TieBreaker) == "" {
			return errors.New("source records require stable identity and composite position")
		}
		if record.Tombstone && record.HardDelete {
			return errors.New("source record cannot be both tombstone and hard delete")
		}
	}
	seenSamples := make(map[string]struct{}, len(fixture.CursorSamples))
	policies := make(map[string]struct{}, len(fixture.CursorSamples))
	for _, sample := range fixture.CursorSamples {
		if strings.TrimSpace(sample.ID) == "" || strings.TrimSpace(sample.Policy) == "" {
			return errors.New("cursor samples require an ID and policy")
		}
		if _, duplicate := seenSamples[sample.ID]; duplicate {
			return fmt.Errorf("cursor sample ID %q is duplicated", sample.ID)
		}
		seenSamples[sample.ID] = struct{}{}
		policies[sample.Policy] = struct{}{}
		if sample.Field != "watermark" && sample.Field != "tie_breaker" {
			return fmt.Errorf("cursor sample %q field = %q, want watermark or tie_breaker", sample.ID, sample.Field)
		}
		if sample.Encoding != "json" && sample.Encoding != "float64" {
			return fmt.Errorf("cursor sample %q encoding = %q, want json or float64", sample.ID, sample.Encoding)
		}
		if len(sample.Value) == 0 || !json.Valid(sample.Value) {
			return fmt.Errorf("cursor sample %q has invalid raw JSON", sample.ID)
		}
		if sample.Accepted && strings.TrimSpace(sample.ExactValue) == "" {
			return fmt.Errorf("accepted cursor sample %q requires an exact value", sample.ID)
		}
		if !sample.Accepted && sample.ExactValue != "" {
			return fmt.Errorf("rejected cursor sample %q cannot declare an exact value", sample.ID)
		}
	}
	if fixture.Scenario == "null_precision_coercion_rejected" {
		for _, policy := range []string{"null", "precision", "coercion"} {
			if _, found := policies[policy]; !found {
				return fmt.Errorf("cursor-policy fixture lacks %s coverage", policy)
			}
		}
	}
	return nil
}

func validatePollingWatermarkConformanceDescriptor(descriptor PollingWatermarkConformanceDescriptor) error {
	if descriptor.Mechanism != pollingWatermarkConformanceMechanism {
		return fmt.Errorf("mechanism = %q, want %q", descriptor.Mechanism, pollingWatermarkConformanceMechanism)
	}
	for name, value := range map[string]string{
		"executor_id":        descriptor.ExecutorID,
		"source_engine":      descriptor.SourceEngine,
		"source_account":     descriptor.SourceAccount,
		"source_scope":       descriptor.SourceScope,
		"source_generation":  descriptor.SourceGeneration,
		"schema_fingerprint": descriptor.SchemaFingerprint,
		"cursor_policy":      descriptor.CursorPolicy,
		"delete_visibility":  descriptor.DeleteVisibility,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}

func validatePollingWatermarkConformanceRegistrationDescriptor(descriptor PollingWatermarkConformanceDescriptor) error {
	if err := validatePollingWatermarkConformanceDescriptor(descriptor); err != nil {
		return err
	}
	if !descriptor.StableKeyset {
		return errors.New("stable keyset is required")
	}
	if descriptor.CursorPolicy != "lossless" {
		return fmt.Errorf("cursor policy = %q, want lossless", descriptor.CursorPolicy)
	}
	if !descriptor.BoundedOverlap {
		return errors.New("bounded overlap is required")
	}
	if !descriptor.BoundedCommitLag {
		return errors.New("bounded commit lag is required")
	}
	return nil
}

func pollingWatermarkConformanceSourceIdentity(descriptor PollingWatermarkConformanceDescriptor) synccontract.SourceIdentity {
	return synccontract.SourceIdentity{
		Engine:           descriptor.SourceEngine,
		AccountOrCluster: descriptor.SourceAccount,
		ObjectScope:      descriptor.SourceScope,
	}
}

func pollingWatermarkConformanceResumeExpectation(descriptor PollingWatermarkConformanceDescriptor) synccontract.ResumeExpectation {
	return synccontract.ResumeExpectation{
		Source:           pollingWatermarkConformanceSourceIdentity(descriptor),
		SourceGeneration: synccontract.OpaqueToken(descriptor.SourceGeneration),
	}
}

func validatePollingWatermarkConformanceObservation(fixture PollingWatermarkConformanceFixture, observation PollingWatermarkConformanceObservation) error {
	expected := fixture.Expected
	if observation.FixtureID != fixture.ID {
		return fmt.Errorf("observation fixture ID = %q, want %q", observation.FixtureID, fixture.ID)
	}
	if observation.Outcome != expected.Outcome {
		return fmt.Errorf("outcome = %q, want %q", observation.Outcome, expected.Outcome)
	}
	if !reflect.DeepEqual(observation.StableIdentities, expected.StableIdentities) {
		return fmt.Errorf("stable identities = %v, want %v", observation.StableIdentities, expected.StableIdentities)
	}
	if !reflect.DeepEqual(observation.SourceRequests, expected.SourceRequests) {
		return fmt.Errorf("source requests = %v, want %v", observation.SourceRequests, expected.SourceRequests)
	}
	if len(observation.CursorSampleResults) != len(fixture.CursorSamples) {
		return fmt.Errorf("cursor sample results = %d, want %d", len(observation.CursorSampleResults), len(fixture.CursorSamples))
	}
	for index, sample := range fixture.CursorSamples {
		result := observation.CursorSampleResults[index]
		if result.ID != sample.ID || result.Accepted != sample.Accepted || result.ExactValue != sample.ExactValue {
			return fmt.Errorf("cursor sample result %d = %+v, want id=%q accepted=%t exact_value=%q", index, result, sample.ID, sample.Accepted, sample.ExactValue)
		}
	}
	if len(observation.PersistedCheckpoints) != expected.PersistedCheckpointCount {
		return fmt.Errorf("persisted checkpoint count = %d, want %d", len(observation.PersistedCheckpoints), expected.PersistedCheckpointCount)
	}
	for index, checkpoint := range observation.PersistedCheckpoints {
		if err := validatePollingWatermarkConformancePersistedCheckpoint(fixture.Descriptor, checkpoint); err != nil {
			return fmt.Errorf("persisted checkpoint %d: %w", index, err)
		}
		if want := expected.PersistedCheckpointPositions[index]; !reflect.DeepEqual(checkpoint.Position, synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken(want.Watermark),
			TieBreaker: synccontract.OpaqueToken(want.TieBreaker),
		}) {
			return fmt.Errorf("persisted checkpoint %d position = %+v, want %+v", index, checkpoint.Position, want)
		}
	}
	if observation.ReplayedFromPriorCommitted != expected.ReplayedFromPriorCommitted {
		return fmt.Errorf("replayed from prior committed = %t, want %t", observation.ReplayedFromPriorCommitted, expected.ReplayedFromPriorCommitted)
	}
	if observation.CheckpointUnchanged != expected.CheckpointUnchanged {
		return fmt.Errorf("checkpoint unchanged = %t, want %t", observation.CheckpointUnchanged, expected.CheckpointUnchanged)
	}
	if string(observation.RecoveryOutcome) != expected.RecoveryOutcome {
		return fmt.Errorf("recovery outcome = %q, want %q", observation.RecoveryOutcome, expected.RecoveryOutcome)
	}
	if expected.RecoveryOutcome == "" {
		if observation.RecoveryError != nil {
			return fmt.Errorf("unexpected recovery error: %v", observation.RecoveryError)
		}
	} else {
		var recovery *synccontract.RebootstrapRequiredError
		if !errors.As(observation.RecoveryError, &recovery) {
			return fmt.Errorf("recovery error = %T %v, want typed rebootstrap outcome %q", observation.RecoveryError, observation.RecoveryError, expected.RecoveryOutcome)
		}
		if string(recovery.Outcome) != expected.RecoveryOutcome {
			return fmt.Errorf("typed recovery outcome = %q, want %q", recovery.Outcome, expected.RecoveryOutcome)
		}
	}
	if !reflect.DeepEqual(observation.TombstoneIdentities, expected.TombstoneIdentities) {
		return fmt.Errorf("tombstone identities = %v, want %v", observation.TombstoneIdentities, expected.TombstoneIdentities)
	}
	if !reflect.DeepEqual(observation.HistoryClosedIdentities, expected.HistoryClosedIdentities) {
		return fmt.Errorf("history closed identities = %v, want %v", observation.HistoryClosedIdentities, expected.HistoryClosedIdentities)
	}
	if observation.HardDeleteInvisible != expected.HardDeleteInvisible {
		return fmt.Errorf("hard delete invisible = %t, want %t", observation.HardDeleteInvisible, expected.HardDeleteInvisible)
	}
	if observation.AdmissionRejected != expected.AdmissionRejected {
		return fmt.Errorf("admission rejected = %t, want %t", observation.AdmissionRejected, expected.AdmissionRejected)
	}
	return nil
}

func validatePollingWatermarkConformancePersistedCheckpoint(descriptor PollingWatermarkConformanceDescriptor, checkpoint synccontract.CheckpointEnvelope) error {
	if err := checkpoint.Validate(); err != nil {
		return fmt.Errorf("is invalid: %w", err)
	}
	if checkpoint.CommittedAt == nil {
		return errors.New("lacks downstream-acknowledged committed_at")
	}
	if err := checkpoint.ValidateResume(pollingWatermarkConformanceResumeExpectation(descriptor)); err != nil {
		return fmt.Errorf("does not resume against the fixture descriptor: %w", err)
	}
	if checkpoint.SchemaVersion != descriptor.SchemaFingerprint {
		return fmt.Errorf("schema fingerprint = %q, want %q", checkpoint.SchemaVersion, descriptor.SchemaFingerprint)
	}
	if checkpoint.Mechanism != descriptor.Mechanism {
		return fmt.Errorf("mechanism = %q, want %q", checkpoint.Mechanism, descriptor.Mechanism)
	}
	return nil
}

func clonePollingWatermarkConformanceFixture(fixture PollingWatermarkConformanceFixture) PollingWatermarkConformanceFixture {
	clone := fixture
	if fixture.InitialCheckpoint != nil {
		checkpoint := *fixture.InitialCheckpoint
		clone.InitialCheckpoint = &checkpoint
	}
	if fixture.Pages != nil {
		clone.Pages = make([]PollingWatermarkConformancePage, len(fixture.Pages))
		for i, page := range fixture.Pages {
			clone.Pages[i] = page
			clone.Pages[i].Records = append([]PollingWatermarkConformanceRecord(nil), page.Records...)
		}
	}
	if fixture.CursorSamples != nil {
		clone.CursorSamples = make([]PollingWatermarkConformanceCursorSample, len(fixture.CursorSamples))
		for i, sample := range fixture.CursorSamples {
			clone.CursorSamples[i] = sample
			clone.CursorSamples[i].Value = append(json.RawMessage(nil), sample.Value...)
		}
	}
	clone.Acknowledgements = append([]PollingWatermarkConformanceAcknowledgement(nil), fixture.Acknowledgements...)
	clone.Expected.StableIdentities = append([]string(nil), fixture.Expected.StableIdentities...)
	clone.Expected.SourceRequests = append([]PollingWatermarkConformancePosition(nil), fixture.Expected.SourceRequests...)
	clone.Expected.PersistedCheckpointPositions = append([]PollingWatermarkConformancePosition(nil), fixture.Expected.PersistedCheckpointPositions...)
	clone.Expected.TombstoneIdentities = append([]string(nil), fixture.Expected.TombstoneIdentities...)
	clone.Expected.HistoryClosedIdentities = append([]string(nil), fixture.Expected.HistoryClosedIdentities...)
	return clone
}

func fixturePagesRecords(pages []PollingWatermarkConformancePage) []PollingWatermarkConformanceRecord {
	var records []PollingWatermarkConformanceRecord
	for _, page := range pages {
		records = append(records, page.Records...)
	}
	return records
}
