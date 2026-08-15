package connectors

import (
	"encoding/json"
	"fmt"
	"strings"

	"polymetrics.ai/internal/synccontract"
)

// PollingWatermarkStatus is the closed lifecycle for the native polling
// declaration. It deliberately does not share the changefeed status: a
// polling scan is not change capture.
type PollingWatermarkStatus string

const (
	PollingWatermarkStatusImplemented PollingWatermarkStatus = "implemented"
	PollingWatermarkStatusPlanned     PollingWatermarkStatus = "planned"
	PollingWatermarkStatusUnsupported PollingWatermarkStatus = "unsupported"
)

// PollingCatalogObjectKind is the catalog object family selected by a native
// polling descriptor. The object name itself is supplied by catalog discovery
// at preflight, never by a caller as SQL text.
type PollingCatalogObjectKind string

const (
	PollingCatalogObjectRelation PollingCatalogObjectKind = "relation"
)

// PollingCatalogObjectSelector selects one discovered catalog object family.
type PollingCatalogObjectSelector struct {
	Kind PollingCatalogObjectKind `json:"kind"`
}

// PollingCatalogObject is the discovered object passed into runtime preflight.
// NameParts and Columns are data from a native cataloger, not query fragments.
type PollingCatalogObject struct {
	Kind      PollingCatalogObjectKind `json:"kind"`
	NameParts []string                 `json:"name_parts"`
	Columns   []string                 `json:"columns"`
}

// PollingReadProtocolKind is the only read protocol available to a mutable
// polling source. Offset traversal is intentionally absent.
type PollingReadProtocolKind string

const (
	PollingReadProtocolKeyset PollingReadProtocolKind = "keyset"
)

// PollingKeysetPredicateDialect identifies a renderer-owned predicate shape.
// It never carries SQL supplied by a definition or caller.
type PollingKeysetPredicateDialect string

const (
	PollingKeysetPredicateLexicographic PollingKeysetPredicateDialect = "lexicographic_tuple"
)

// PollingReadProtocol bounds one native source read protocol.
type PollingReadProtocol struct {
	Kind            PollingReadProtocolKind       `json:"kind"`
	MaxPageSize     int                           `json:"max_page_size"`
	MaxPages        int                           `json:"max_pages"`
	MaxRequests     int                           `json:"max_requests"`
	StableTraversal bool                          `json:"stable_traversal"`
	Predicate       PollingKeysetPredicateDialect `json:"predicate"`
}

// PollingSnapshotBarrierKind is the explicit source snapshot/barrier policy.
type PollingSnapshotBarrierKind string

const (
	PollingSnapshotBarrierTransaction PollingSnapshotBarrierKind = "transaction_snapshot"
	PollingSnapshotBarrierNone        PollingSnapshotBarrierKind = "none"
)

// PollingSnapshotBarrier declares the source isolation/barrier mechanism.
type PollingSnapshotBarrier struct {
	Kind PollingSnapshotBarrierKind `json:"kind"`
}

// PollingCursorCodec is a closed, lossless cursor encoding family.
type PollingCursorCodec string

const (
	PollingCursorCodecRFC3339Nano PollingCursorCodec = "rfc3339_nano"
	PollingCursorCodecDecimal     PollingCursorCodec = "decimal"
	// PollingCursorCodecFloat64 is deliberately representable only so the
	// runtime can refuse a lossy declaration with a specific diagnostic.
	PollingCursorCodecFloat64 PollingCursorCodec = "float64"
)

// PollingCursorType identifies the physical source type without coercing it.
type PollingCursorType string

const (
	PollingCursorTypeTimestamp PollingCursorType = "timestamp"
	PollingCursorTypeInteger   PollingCursorType = "integer"
)

// PollingCursor records the type, codec, precision, and null policy consumed
// by the registered source executor.
type PollingCursor struct {
	Codec      PollingCursorCodec `json:"codec"`
	Type       PollingCursorType  `json:"type"`
	Precision  string             `json:"precision"`
	AllowsNull bool               `json:"allows_null,omitempty"`
}

// PollingOrderingField names one discovered catalog column or, for a unique
// tie breaker, the complete ordered column tuple. Exactly one of CatalogField
// and CatalogFields is populated.
type PollingOrderingField struct {
	CatalogField  string   `json:"catalog_field,omitempty"`
	CatalogFields []string `json:"catalog_fields,omitempty"`
	Ascending     bool     `json:"ascending"`
	Unique        bool     `json:"unique,omitempty"`
}

// CatalogColumns returns an independent copy of the declared order columns.
func (f PollingOrderingField) CatalogColumns() []string {
	if len(f.CatalogFields) > 0 {
		return append([]string(nil), f.CatalogFields...)
	}
	if f.CatalogField == "" {
		return nil
	}
	return []string{f.CatalogField}
}

// PollingOrderingTuple is the required watermark plus unique tie-breaker
// order. It makes tied watermarks resumable without a scalar fallback.
type PollingOrderingTuple struct {
	Watermark  PollingOrderingField `json:"watermark"`
	TieBreaker PollingOrderingField `json:"tie_breaker"`
}

// PollingMutationPolicy makes source mutation and commit-order assumptions
// explicit instead of inferring them from a cursor type.
type PollingMutationPolicy struct {
	Mutable        bool `json:"mutable"`
	CommitOrdered  bool `json:"commit_ordered"`
	BoundedOverlap bool `json:"bounded_overlap"`
}

// PollingSourceIdentity identifies the source parts incorporated in #3810
// checkpoints. It never contains a credential or raw connection string.
type PollingSourceIdentity struct {
	Engine       string `json:"engine"`
	AccountScope string `json:"account_scope"`
	ObjectScope  string `json:"object_scope"`
}

// PollingSchemaCompatibility defines the only accepted resume schema rule.
type PollingSchemaCompatibility string

const (
	PollingSchemaCompatibilityExactFingerprint PollingSchemaCompatibility = "exact_fingerprint"
)

// PollingDeleteVisibility makes polling's hard-delete limitation explicit.
type PollingDeleteVisibility string

const (
	PollingDeleteVisibilityHardDeleteInvisible PollingDeleteVisibility = "hard_delete_invisible"
	PollingDeleteVisibilityTombstone           PollingDeleteVisibility = "tombstone"
)

// PollingWatermarkSourceDescriptor declares only the bounded source facts
// needed for a registered native polling executor.
type PollingWatermarkSourceDescriptor struct {
	Executor                 TransportExecutorReference   `json:"executor"`
	Object                   PollingCatalogObjectSelector `json:"object"`
	Read                     PollingReadProtocol          `json:"read"`
	Snapshot                 PollingSnapshotBarrier       `json:"snapshot"`
	Cursor                   PollingCursor                `json:"cursor"`
	Ordering                 PollingOrderingTuple         `json:"ordering"`
	Mutation                 PollingMutationPolicy        `json:"mutation"`
	Identity                 PollingSourceIdentity        `json:"identity"`
	Schema                   PollingSchemaCompatibility   `json:"schema_compatibility"`
	DeleteVisibility         PollingDeleteVisibility      `json:"delete_visibility"`
	SoftDeleteField          string                       `json:"soft_delete_field,omitempty"`
	SoftDeleteAdvancesCursor bool                         `json:"soft_delete_advances_cursor,omitempty"`
	Modes                    []synccontract.Mode          `json:"modes"`
}

// PollingStagingCapability records whether a bounded target can stage and
// replace data without selecting a caller-authored table or statement.
type PollingStagingCapability string

const (
	PollingStagingReplaceSupported   PollingStagingCapability = "staging_replace_supported"
	PollingStagingReplaceUnsupported PollingStagingCapability = "staging_replace_unsupported"
)

// PollingTransactionPolicy is the target's closed atomicity policy.
type PollingTransactionPolicy string

const (
	PollingTransactionRequired PollingTransactionPolicy = "required"
	PollingTransactionNone     PollingTransactionPolicy = "none"
)

// PollingPartialResultPolicy states the destination policy when a batch does
// not completely apply.
type PollingPartialResultPolicy string

const (
	PollingPartialResultRollback PollingPartialResultPolicy = "rollback"
	PollingPartialResultUnknown  PollingPartialResultPolicy = "unknown"
)

// PollingValidityWindowCapability is the destination history support claim.
type PollingValidityWindowCapability string

const (
	PollingValidityWindowSupported   PollingValidityWindowCapability = "supported"
	PollingValidityWindowUnsupported PollingValidityWindowCapability = "unsupported"
)

// PollingApplyStrategy is a closed native target operation family.
type PollingApplyStrategy string

const (
	PollingApplyStrategyAppend        PollingApplyStrategy = "append"
	PollingApplyStrategyReplace       PollingApplyStrategy = "replace"
	PollingApplyStrategyMerge         PollingApplyStrategy = "merge"
	PollingApplyStrategyDedupe        PollingApplyStrategy = "dedupe"
	PollingApplyStrategyDedupeHistory PollingApplyStrategy = "dedupe_history"
)

// PollingApplyDescriptor declares a bounded native target contract without
// carrying target DML, a table name, or a generic write escape hatch.
type PollingApplyDescriptor struct {
	Executor                TransportExecutorReference      `json:"executor"`
	MaxBatchRecords         int                             `json:"max_batch_records"`
	MaxBatchBytes           int                             `json:"max_batch_bytes"`
	Staging                 PollingStagingCapability        `json:"staging"`
	StableKeyMapping        []string                        `json:"stable_key_mapping"`
	ConditionalOrderFence   bool                            `json:"conditional_order_fence"`
	Transaction             PollingTransactionPolicy        `json:"transaction"`
	PartialResult           PollingPartialResultPolicy      `json:"partial_result"`
	RetrySafeCloseAndInsert bool                            `json:"retry_safe_close_and_insert"`
	ValidityWindow          PollingValidityWindowCapability `json:"validity_window"`
	Strategies              []PollingApplyStrategy          `json:"strategies"`
}

// PollingWatermarkDescriptor is the complete definition-owned declaration
// consumed by engine.PollingPreflight. It is distinct from ChangefeedDescriptor
// because polling cannot truthfully claim change-capture behavior.
type PollingWatermarkDescriptor struct {
	Status PollingWatermarkStatus           `json:"status"`
	Reason string                           `json:"reason,omitempty"`
	Source PollingWatermarkSourceDescriptor `json:"source"`
	Target PollingApplyDescriptor           `json:"target"`
}

// MarshalJSON omits the intentionally absent runtime contract on planned and
// unsupported declarations. Empty nested source and target values look like a
// usable binding in inspection/catalog output even though those statuses can
// never pass preflight.
func (d PollingWatermarkDescriptor) MarshalJSON() ([]byte, error) {
	if d.Status != PollingWatermarkStatusImplemented {
		return json.Marshal(struct {
			Status PollingWatermarkStatus `json:"status"`
			Reason string                 `json:"reason,omitempty"`
		}{
			Status: d.Status,
			Reason: d.Reason,
		})
	}
	type encoded PollingWatermarkDescriptor
	return json.Marshal(encoded(d))
}

// LegacyPollingWatermarkMode adapts one of #3810's five retained public mode
// spellings to its closed canonical mode. It deliberately reads the shared
// compatibility table rather than reproducing strings here, and declarations
// themselves must still contain only canonical synccontract.Mode values.
func LegacyPollingWatermarkMode(raw string) (synccontract.Mode, bool) {
	name := strings.ToLower(strings.TrimSpace(raw))
	for _, mode := range synccontract.PublicModes() {
		if name == mode.Name {
			return mode.ContractMode, true
		}
	}
	return "", false
}

// Clone returns an independent descriptor so inspection and runtime preflight
// never mutate an authored declaration.
func (d PollingWatermarkDescriptor) Clone() *PollingWatermarkDescriptor {
	clone := d
	clone.Source.Modes = append([]synccontract.Mode(nil), d.Source.Modes...)
	clone.Source.Ordering.Watermark.CatalogFields = append([]string(nil), d.Source.Ordering.Watermark.CatalogFields...)
	clone.Source.Ordering.TieBreaker.CatalogFields = append([]string(nil), d.Source.Ordering.TieBreaker.CatalogFields...)
	clone.Target.StableKeyMapping = append([]string(nil), d.Target.StableKeyMapping...)
	clone.Target.Strategies = append([]PollingApplyStrategy(nil), d.Target.Strategies...)
	return &clone
}

// Validate ensures a descriptor has only closed, bounded declaration values.
// Runtime registration and immutable corpus proof are intentionally reserved
// for engine.PollingPreflight rather than duplicated here.
func (d PollingWatermarkDescriptor) Validate() error {
	switch d.Status {
	case PollingWatermarkStatusImplemented:
		if strings.TrimSpace(d.Reason) != "" {
			return fmt.Errorf("implemented polling watermark cannot declare a reason")
		}
		if err := d.Source.validate(); err != nil {
			return err
		}
		return d.Target.validate(d.Source.Modes)
	case PollingWatermarkStatusPlanned, PollingWatermarkStatusUnsupported:
		if strings.TrimSpace(d.Reason) == "" {
			return fmt.Errorf("non-implemented polling watermark requires a reason")
		}
		return nil
	default:
		return fmt.Errorf("unsupported polling watermark status %q", d.Status)
	}
}

func (d PollingWatermarkSourceDescriptor) validate() error {
	if err := validatePollingExecutor(d.Executor); err != nil {
		return fmt.Errorf("source polling executor: %w", err)
	}
	if d.Object.Kind != PollingCatalogObjectRelation {
		return fmt.Errorf("unsupported polling catalog object kind %q", d.Object.Kind)
	}
	if d.Read.Kind != PollingReadProtocolKeyset {
		return fmt.Errorf("polling read protocol must be keyset")
	}
	if d.Read.MaxPageSize <= 0 || d.Read.MaxPages <= 0 || d.Read.MaxRequests <= 0 || d.Read.MaxPageSize > 100000 || d.Read.MaxPages > 10000 || d.Read.MaxRequests > 100000 {
		return fmt.Errorf("polling read requires bounded positive page and request limits")
	}
	if !d.Read.StableTraversal {
		return fmt.Errorf("page checkpoints require stable keyset traversal")
	}
	if d.Read.Predicate != PollingKeysetPredicateLexicographic {
		return fmt.Errorf("polling keyset predicate must use the closed lexicographic tuple dialect")
	}
	if d.Snapshot.Kind != PollingSnapshotBarrierTransaction && d.Snapshot.Kind != PollingSnapshotBarrierNone {
		return fmt.Errorf("unsupported polling snapshot barrier %q", d.Snapshot.Kind)
	}
	if err := d.Cursor.validate(); err != nil {
		return err
	}
	if err := d.Ordering.validate(); err != nil {
		return err
	}
	if d.Mutation.Mutable && (!d.Mutation.CommitOrdered || !d.Mutation.BoundedOverlap) {
		return fmt.Errorf("mutable source requires bounded overlap and commit ordering")
	}
	if !validPollingIdentity(d.Identity.Engine) || !validPollingIdentity(d.Identity.AccountScope) || !validPollingIdentity(d.Identity.ObjectScope) {
		return fmt.Errorf("polling source identity requires concrete engine, account_scope, and object_scope")
	}
	if d.Schema != PollingSchemaCompatibilityExactFingerprint {
		return fmt.Errorf("polling schema compatibility must be exact_fingerprint")
	}
	if err := d.validateDeletes(); err != nil {
		return err
	}
	if err := validatePollingModes(d.Modes); err != nil {
		return err
	}
	return nil
}

func (c PollingCursor) validate() error {
	if c.AllowsNull {
		return fmt.Errorf("polling watermark cursor cannot allow null")
	}
	if c.Codec == PollingCursorCodecFloat64 {
		return fmt.Errorf("polling watermark cursor codec must preserve values losslessly")
	}
	switch c.Type {
	case PollingCursorTypeTimestamp:
		if c.Codec != PollingCursorCodecRFC3339Nano || c.Precision != "nanosecond" {
			return fmt.Errorf("timestamp polling cursor requires rfc3339_nano nanosecond precision")
		}
	case PollingCursorTypeInteger:
		if c.Codec != PollingCursorCodecDecimal || c.Precision != "exact" {
			return fmt.Errorf("integer polling cursor requires exact decimal encoding")
		}
	default:
		return fmt.Errorf("unsupported polling cursor type %q", c.Type)
	}
	return nil
}

func (o PollingOrderingTuple) validate() error {
	watermark, err := validatePollingOrderingColumns(o.Watermark)
	if err != nil || len(watermark) != 1 {
		return fmt.Errorf("polling ordering fields must name discovered catalog columns")
	}
	tieBreaker, err := validatePollingOrderingColumns(o.TieBreaker)
	if err != nil || len(tieBreaker) == 0 {
		return fmt.Errorf("polling ordering fields must name discovered catalog columns")
	}
	for _, field := range tieBreaker {
		if field == watermark[0] {
			return fmt.Errorf("polling ordering watermark and tie_breaker must differ")
		}
	}
	if !o.Watermark.Ascending || !o.TieBreaker.Ascending {
		return fmt.Errorf("polling ordering tuple must be ascending")
	}
	if !o.TieBreaker.Unique {
		return fmt.Errorf("polling ordering tie_breaker must be unique")
	}
	return nil
}

func validatePollingOrderingColumns(field PollingOrderingField) ([]string, error) {
	if (field.CatalogField == "") == (len(field.CatalogFields) == 0) {
		return nil, fmt.Errorf("polling ordering field requires exactly one scalar or tuple declaration")
	}
	columns := field.CatalogColumns()
	seen := make(map[string]struct{}, len(columns))
	for _, column := range columns {
		if !validPollingName(column) {
			return nil, fmt.Errorf("polling ordering field names an invalid catalog column")
		}
		if _, duplicate := seen[column]; duplicate {
			return nil, fmt.Errorf("polling ordering field repeats catalog column %q", column)
		}
		seen[column] = struct{}{}
	}
	return columns, nil
}

func (d PollingWatermarkSourceDescriptor) validateDeletes() error {
	switch d.DeleteVisibility {
	case PollingDeleteVisibilityHardDeleteInvisible:
		if d.SoftDeleteField != "" || d.SoftDeleteAdvancesCursor {
			return fmt.Errorf("hard-delete-invisible polling cannot declare a soft delete mapping")
		}
	case PollingDeleteVisibilityTombstone:
		if !validPollingName(d.SoftDeleteField) || !d.SoftDeleteAdvancesCursor {
			return fmt.Errorf("polling watermark cannot advertise tombstones without a cursor-advancing soft delete")
		}
	default:
		return fmt.Errorf("unsupported polling delete visibility %q", d.DeleteVisibility)
	}
	return nil
}

func (d PollingApplyDescriptor) validate(sourceModes []synccontract.Mode) error {
	if err := validatePollingExecutor(d.Executor); err != nil {
		return fmt.Errorf("target polling executor: %w", err)
	}
	if d.MaxBatchRecords <= 0 || d.MaxBatchRecords > 100000 {
		return fmt.Errorf("target polling apply requires a bounded positive batch size")
	}
	if d.MaxBatchBytes <= 0 || d.MaxBatchBytes > 1<<30 {
		return fmt.Errorf("target polling apply requires a bounded positive byte limit")
	}
	if d.Staging != PollingStagingReplaceSupported && d.Staging != PollingStagingReplaceUnsupported {
		return fmt.Errorf("unsupported polling staging capability %q", d.Staging)
	}
	if len(d.StableKeyMapping) == 0 {
		return fmt.Errorf("target polling apply requires stable key mapping")
	}
	for _, key := range d.StableKeyMapping {
		if !validPollingName(key) {
			return fmt.Errorf("target polling stable key mapping is invalid")
		}
	}
	if d.Transaction != PollingTransactionRequired && d.Transaction != PollingTransactionNone {
		return fmt.Errorf("unsupported polling transaction policy %q", d.Transaction)
	}
	if d.PartialResult != PollingPartialResultRollback && d.PartialResult != PollingPartialResultUnknown {
		return fmt.Errorf("unsupported polling partial-result policy %q", d.PartialResult)
	}
	if d.ValidityWindow != PollingValidityWindowSupported && d.ValidityWindow != PollingValidityWindowUnsupported {
		return fmt.Errorf("unsupported polling validity-window capability %q", d.ValidityWindow)
	}
	if len(d.Strategies) == 0 {
		return fmt.Errorf("target polling apply requires at least one closed strategy")
	}
	seen := make(map[PollingApplyStrategy]struct{}, len(d.Strategies))
	for _, strategy := range d.Strategies {
		if !validPollingApplyStrategy(strategy) {
			return fmt.Errorf("unsupported polling apply strategy %q", strategy)
		}
		if _, duplicate := seen[strategy]; duplicate {
			return fmt.Errorf("target polling apply declares duplicate strategy %q", strategy)
		}
		seen[strategy] = struct{}{}
	}
	if containsPollingMode(sourceModes, synccontract.ModeIncrementalDedupeHistory) && (d.Transaction != PollingTransactionRequired || !d.RetrySafeCloseAndInsert || d.ValidityWindow != PollingValidityWindowSupported || !containsPollingApplyStrategy(d.Strategies, PollingApplyStrategyDedupeHistory)) {
		return fmt.Errorf("history mode requires transaction and retry-safe close-and-insert")
	}
	return nil
}

func validatePollingExecutor(reference TransportExecutorReference) error {
	if err := reference.Validate(); err != nil {
		return err
	}
	if reference.Family != TransportExecutorFamilyNativeDatabase {
		return fmt.Errorf("polling executor must use native_database family")
	}
	return nil
}

func validatePollingModes(modes []synccontract.Mode) error {
	if len(modes) == 0 {
		return fmt.Errorf("polling watermark requires at least one canonical sync mode")
	}
	seen := make(map[synccontract.Mode]struct{}, len(modes))
	for _, mode := range modes {
		if err := mode.Validate(); err != nil {
			return err
		}
		if mode == synccontract.ModeChangeCapture {
			return fmt.Errorf("polling watermark cannot declare change_capture mode")
		}
		if _, duplicate := seen[mode]; duplicate {
			return fmt.Errorf("polling watermark declares duplicate sync mode %q", mode)
		}
		seen[mode] = struct{}{}
	}
	return nil
}

func validPollingApplyStrategy(strategy PollingApplyStrategy) bool {
	switch strategy {
	case PollingApplyStrategyAppend, PollingApplyStrategyReplace, PollingApplyStrategyMerge, PollingApplyStrategyDedupe, PollingApplyStrategyDedupeHistory:
		return true
	default:
		return false
	}
}

func containsPollingApplyStrategy(strategies []PollingApplyStrategy, expected PollingApplyStrategy) bool {
	for _, strategy := range strategies {
		if strategy == expected {
			return true
		}
	}
	return false
}

func containsPollingMode(modes []synccontract.Mode, expected synccontract.Mode) bool {
	for _, mode := range modes {
		if mode == expected {
			return true
		}
	}
	return false
}

func validPollingName(value string) bool {
	if value == "" || strings.TrimSpace(value) != value {
		return false
	}
	for index, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || (index > 0 && r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

func validPollingIdentity(value string) bool {
	if value == "" || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			continue
		}
		return false
	}
	return true
}
