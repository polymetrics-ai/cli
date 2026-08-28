package connectors

import (
	"fmt"
	"strings"
	"unicode"

	"polymetrics.ai/internal/synccontract"
)

// TransportRole identifies the side of the closed, warehouse-mediated sync
// contract an executor performs. It is deliberately not a provider direction
// alias: the warehouse remains between the two roles.
type TransportRole string

const (
	TransportRoleSource      TransportRole = "source"
	TransportRoleDestination TransportRole = "destination"
)

// TransportExecutorFamily is the closed implementation family an integration
// type may use. It prevents a connector descriptor from becoming a generic
// HTTP, SQL, shell, or arbitrary-action dispatch surface.
type TransportExecutorFamily string

const (
	TransportExecutorFamilyDeclarativeAPI TransportExecutorFamily = "declarative_api"
	TransportExecutorFamilyNativeAPI      TransportExecutorFamily = "native_api"
	TransportExecutorFamilyNativeDatabase TransportExecutorFamily = "native_database"
	TransportExecutorFamilyFile           TransportExecutorFamily = "file"
	TransportExecutorFamilyQueue          TransportExecutorFamily = "queue"
)

// TransportExecutorReference names one concrete, registered executor. It
// carries neither a provider URL nor a query, command, or action payload.
type TransportExecutorReference struct {
	Family TransportExecutorFamily `json:"family"`
	ID     string                  `json:"id"`
}

// DeliveryGuarantees records the provider/native facts a descriptor must
// make explicit. The orchestrator does not invent a guarantee from a protocol
// or HTTP verb.
type DeliveryGuarantees struct {
	Idempotency DeliveryIdempotency `json:"idempotency"`
	Ordering    DeliveryOrdering    `json:"ordering"`
	Deletes     DeliveryDeletes     `json:"deletes"`
}

type DeliveryIdempotency string

const (
	DeliveryIdempotencyKeyed       DeliveryIdempotency = "keyed"
	DeliveryIdempotencyAtLeastOnce DeliveryIdempotency = "at_least_once"
	DeliveryIdempotencyNone        DeliveryIdempotency = "none"
)

type DeliveryOrdering string

const (
	DeliveryOrderingSource DeliveryOrdering = "source_ordered"
	// DeliveryOrderingWindowCoalesced means the executor exhausts one complete
	// provider-defined token window, coalesces it by stable key without relying
	// on event order, and only then emits current state. It is not source order
	// and therefore cannot represent ordered history or change capture.
	DeliveryOrderingWindowCoalesced DeliveryOrdering = "window_coalesced"
	DeliveryOrderingUnordered       DeliveryOrdering = "unordered"
)

type DeliveryDeletes string

const (
	DeliveryDeletesTombstone   DeliveryDeletes = "tombstone"
	DeliveryDeletesUnavailable DeliveryDeletes = "not_available"
)

// ConformanceEvidenceReference identifies an externally recorded verification
// result. It is intentionally only a reference: a descriptor and an executor
// cannot self-admit by returning this value. synctransport asks its separately
// supplied verifier to establish whether this reference is accepted.
type ConformanceEvidenceReference struct {
	Suite string `json:"suite"`
	RunID string `json:"run_id"`
}

// TransportAcknowledgement states the durability policy a destination
// declares. `none` is structurally valid so inspection can report an honest
// declaration, but runtime preflight refuses to execute it.
type TransportAcknowledgement string

const (
	TransportAcknowledgementDurableWarehouse TransportAcknowledgement = "durable_warehouse"
	TransportAcknowledgementNone             TransportAcknowledgement = "none"
)

// ApplyStrategy is a closed physical application strategy selected by a
// canonical mode. It is not a free-form connector write action.
type ApplyStrategy string

const (
	ApplyStrategyAppend        ApplyStrategy = "append"
	ApplyStrategyReplace       ApplyStrategy = "replace"
	ApplyStrategyMerge         ApplyStrategy = "merge"
	ApplyStrategyDedupe        ApplyStrategy = "dedupe"
	ApplyStrategyDedupeHistory ApplyStrategy = "dedupe_history"
	ApplyStrategyChangeApply   ApplyStrategy = "change_apply"
)

// DestinationApplyStrategy binds one #3810 mode to a closed strategy and a
// descriptor-listed destination action. Future provider/database adapters own
// implementation of that named action; this package never invokes arbitrary
// action strings.
type DestinationApplyStrategy struct {
	Mode            synccontract.Mode `json:"mode"`
	Strategy        ApplyStrategy     `json:"strategy"`
	Action          string            `json:"action"`
	TombstoneAction string            `json:"tombstone_action,omitempty"`
	// ReadBack belongs to this exact physical apply action. A descriptor-wide
	// policy cannot truthfully validate distinct action schemas after a write.
	ReadBack          *DestinationReadBackPolicy          `json:"read_back,omitempty"`
	TombstoneReadBack *DestinationTombstoneReadBackPolicy `json:"tombstone_read_back,omitempty"`
}

// SourceRecordMappingKind is the closed source-record behavior a destination
// may select for one admitted source executor. It is deliberately a small
// declaration vocabulary, not a generic expression or record-mapping engine.
type SourceRecordMappingKind string

const (
	SourceRecordMappingKindConfigMatch SourceRecordMappingKind = "config_match"
	SourceRecordMappingKindInputFields SourceRecordMappingKind = "input_fields"
)

// SourceRecordInputBinding maps one destination action input to a concrete
// source record field. The destination action still owns its output field and
// shape; this binding owns only the upstream field it admits.
type SourceRecordInputBinding struct {
	Input string `json:"input"`
	Field string `json:"field"`
}

// TombstoneRecordMappingImage is the durable tombstone image from which a
// declaration-owned delete action may take its inputs. It is not a record
// expression language: the only values admitted are the source contract's
// key or before image.
type TombstoneRecordMappingImage string

const (
	TombstoneRecordMappingImageKey    TombstoneRecordMappingImage = "key"
	TombstoneRecordMappingImageBefore TombstoneRecordMappingImage = "before"
)

// TombstoneRecordMapping maps one delete action's inputs from the selected
// tombstone image. It remains separate from SourceRecordMapping so a delete
// cannot accidentally be treated as a create/update source record.
type TombstoneRecordMapping struct {
	Image  TombstoneRecordMappingImage `json:"image"`
	Inputs []SourceRecordInputBinding  `json:"inputs"`
}

func (m TombstoneRecordMapping) Clone() TombstoneRecordMapping {
	clone := m
	clone.Inputs = append([]SourceRecordInputBinding(nil), m.Inputs...)
	return clone
}

func (m TombstoneRecordMapping) Validate() error {
	switch m.Image {
	case TombstoneRecordMappingImageKey, TombstoneRecordMappingImageBefore:
	default:
		return fmt.Errorf("unsupported tombstone record mapping image %q", m.Image)
	}
	if len(m.Inputs) == 0 {
		return fmt.Errorf("tombstone record mapping requires at least one input field")
	}
	seenInputs := make(map[string]struct{}, len(m.Inputs))
	seenFields := make(map[string]struct{}, len(m.Inputs))
	for _, input := range m.Inputs {
		if input.Input == "" || input.Field == "" {
			return fmt.Errorf("tombstone record mapping requires non-empty input and field names")
		}
		if _, duplicate := seenInputs[input.Input]; duplicate {
			return fmt.Errorf("tombstone record mapping duplicates input %q", input.Input)
		}
		if _, duplicate := seenFields[input.Field]; duplicate {
			return fmt.Errorf("tombstone record mapping duplicates field %q", input.Field)
		}
		seenInputs[input.Input] = struct{}{}
		seenFields[input.Field] = struct{}{}
	}
	return nil
}

// DestinationBatchDisposition is the closed delivery unit declared beside an
// action-owned source binding. It is deliberately not a provider bulk-write
// instruction: the named writes.json action remains the only request shape.
type DestinationBatchDisposition string

const (
	// DestinationBatchPerRecord preserves the engine's deterministic one
	// declaration-owned request per record behavior while MaxRecords bounds the
	// reopened warehouse workset that may be applied under one acknowledgement.
	DestinationBatchPerRecord DestinationBatchDisposition = "per_record"
)

// DestinationBatch bounds one action-owned destination apply unit. A caller
// may request a smaller source batch but cannot enlarge this declaration.
type DestinationBatch struct {
	Disposition DestinationBatchDisposition `json:"disposition"`
	MaxRecords  int                         `json:"max_records"`
}

func (b DestinationBatch) Validate() error {
	if b.Disposition != DestinationBatchPerRecord {
		return fmt.Errorf("unsupported destination batch disposition %q", b.Disposition)
	}
	if b.MaxRecords < 1 || b.MaxRecords > 1000 {
		return fmt.Errorf("destination batch max_records must be between 1 and 1000")
	}
	return nil
}

// SourceRecordMapping binds an admitted source record to one of the closed
// mapping forms. config_match compares one source record field with a source
// endpoint configuration value. input_fields supplies the declared action
// inputs from named source record fields.
type SourceRecordMapping struct {
	Kind        SourceRecordMappingKind    `json:"kind"`
	ConfigKey   string                     `json:"config_key,omitempty"`
	RecordField string                     `json:"record_field,omitempty"`
	Inputs      []SourceRecordInputBinding `json:"inputs,omitempty"`
}

// DestinationSourceBinding declares one exact source executor and stream
// allowlist that a destination transport accepts. The source record contract is
// definition-owned, so adding another connector cannot require shared
// provider-named routing code.
type DestinationSourceBinding struct {
	// Action is the exact eligible destination action this source mapping may
	// materialize. It is optional only for legacy specialized adapters; the
	// reusable declarative typed destination requires it.
	Action          string                     `json:"action,omitempty"`
	Executor        TransportExecutorReference `json:"executor"`
	EligibleStreams []string                   `json:"eligible_streams"`
	RecordMapping   SourceRecordMapping        `json:"record_mapping,omitempty"`
	// TombstoneMapping maps the exact action's delete inputs from a durable
	// tombstone key or before image. It is mutually exclusive with
	// RecordMapping: one action is either an ordinary record action or an
	// explicitly named tombstone delete action.
	TombstoneMapping *TombstoneRecordMapping `json:"tombstone_mapping,omitempty"`
	// Batch is required by the reusable declarative typed destination. Keeping
	// it adjacent to Action makes the bound unit part of the sealed mapping,
	// rather than a caller-selected provider request control.
	Batch *DestinationBatch `json:"batch,omitempty"`
}

func (m SourceRecordMapping) Clone() SourceRecordMapping {
	clone := m
	clone.Inputs = append([]SourceRecordInputBinding(nil), m.Inputs...)
	return clone
}

func (b DestinationSourceBinding) Clone() DestinationSourceBinding {
	clone := b
	clone.EligibleStreams = append([]string(nil), b.EligibleStreams...)
	clone.RecordMapping = b.RecordMapping.Clone()
	if b.Batch != nil {
		batch := *b.Batch
		clone.Batch = &batch
	}
	if b.TombstoneMapping != nil {
		mapping := b.TombstoneMapping.Clone()
		clone.TombstoneMapping = &mapping
	}
	return clone
}

func (m SourceRecordMapping) Validate() error {
	switch m.Kind {
	case SourceRecordMappingKindConfigMatch:
		if !isConcreteTransportIdentifier(m.ConfigKey) || !isConcreteTransportIdentifier(m.RecordField) {
			return fmt.Errorf("config_match source record mapping requires concrete config_key and record_field")
		}
		if len(m.Inputs) != 0 {
			return fmt.Errorf("config_match source record mapping does not accept input fields")
		}
	case SourceRecordMappingKindInputFields:
		if m.ConfigKey != "" || m.RecordField != "" {
			return fmt.Errorf("input_fields source record mapping does not accept config_match fields")
		}
		if len(m.Inputs) == 0 {
			return fmt.Errorf("input_fields source record mapping requires at least one input field")
		}
		seenInputs := make(map[string]struct{}, len(m.Inputs))
		seenFields := make(map[string]struct{}, len(m.Inputs))
		for _, input := range m.Inputs {
			if input.Input == "" || input.Field == "" {
				return fmt.Errorf("input_fields source record mapping requires non-empty input and field names")
			}
			if _, duplicate := seenInputs[input.Input]; duplicate {
				return fmt.Errorf("input_fields source record mapping duplicates input %q", input.Input)
			}
			if _, duplicate := seenFields[input.Field]; duplicate {
				return fmt.Errorf("input_fields source record mapping duplicates field %q", input.Field)
			}
			seenInputs[input.Input] = struct{}{}
			seenFields[input.Field] = struct{}{}
		}
	default:
		return fmt.Errorf("unsupported source record mapping kind %q", m.Kind)
	}
	return nil
}

func (b DestinationSourceBinding) Validate() error {
	if b.Action != "" && !isConcreteTransportIdentifier(b.Action) {
		return fmt.Errorf("destination source binding action must be a concrete identifier")
	}
	if err := b.Executor.Validate(); err != nil {
		return err
	}
	if err := validateSourceTransportStreams(b.EligibleStreams); err != nil {
		return err
	}
	if b.Batch != nil {
		if err := b.Batch.Validate(); err != nil {
			return err
		}
	}
	hasRecordMapping := b.RecordMapping.Kind != ""
	hasTombstoneMapping := b.TombstoneMapping != nil
	if hasRecordMapping == hasTombstoneMapping {
		return fmt.Errorf("destination source binding requires exactly one record_mapping or tombstone_mapping")
	}
	if hasRecordMapping {
		return b.RecordMapping.Validate()
	}
	return b.TombstoneMapping.Validate()
}

// DestinationReadBackField maps one field in provider state to the
// corresponding field in the destination record that was written.
type DestinationReadBackField struct {
	ProviderField string `json:"provider_field"`
	ExpectedField string `json:"expected_field"`
}

// DestinationReceiptLocator is a deliberately small provider-response to
// provider-read mapping. It allows a declared write response's one top-level
// scalar field to fill one declared query parameter; it is not a JSONPath,
// URL, method, header, or generic read capability.
type DestinationReceiptLocator struct {
	ResponseIndex  int    `json:"response_index"`
	BodyField      string `json:"body_field"`
	QueryParameter string `json:"query_parameter"`
	MaxValueBytes  int    `json:"max_value_bytes"`
	MaxPages       int    `json:"max_pages"`
}

func (l DestinationReceiptLocator) Validate() error {
	if l.ResponseIndex < 0 || l.ResponseIndex > 1023 {
		return fmt.Errorf("destination receipt locator response_index must be between 0 and 1023")
	}
	if !isConcreteTransportIdentifier(l.BodyField) || !isConcreteTransportIdentifier(l.QueryParameter) {
		return fmt.Errorf("destination receipt locator requires concrete body_field and query_parameter")
	}
	if l.MaxValueBytes < 1 || l.MaxValueBytes > 4096 {
		return fmt.Errorf("destination receipt locator max_value_bytes must be between 1 and 4096")
	}
	if l.MaxPages < 1 || l.MaxPages > 10 {
		return fmt.Errorf("destination receipt locator max_pages must be between 1 and 10")
	}
	return nil
}

// DestinationReadBackPolicy declares the bounded provider read that proves a
// typed destination write before its source checkpoint may advance. Operation
// is interpreted only by the connector that owns the declaration.
type DestinationReadBackPolicy struct {
	Operation              string                       `json:"operation"`
	Identity               []DestinationReadBackField   `json:"identity"`
	Expected               []DestinationReadBackField   `json:"expected"`
	MaxRecords             int                          `json:"max_records"`
	MaxAttempts            int                          `json:"max_attempts"`
	TimeoutMilliseconds    int                          `json:"timeout_milliseconds"`
	RetryDelayMilliseconds int                          `json:"retry_delay_milliseconds,omitempty"`
	ReceiptLocator         DestinationReceiptLocator    `json:"receipt_locator"`
	Conformance            ConformanceEvidenceReference `json:"conformance"`
}

// DestinationTombstoneReadBackPolicy declares the bounded provider read that
// proves a delete action removed its declared destination identity. It is
// separate from ordinary read-back because the successful state is absence,
// not a create/update payload that happens to omit fields.
type DestinationTombstoneReadBackPolicy struct {
	Operation              string                       `json:"operation"`
	Identity               []DestinationReadBackField   `json:"identity"`
	MaxRecords             int                          `json:"max_records"`
	MaxAttempts            int                          `json:"max_attempts"`
	TimeoutMilliseconds    int                          `json:"timeout_milliseconds"`
	RetryDelayMilliseconds int                          `json:"retry_delay_milliseconds,omitempty"`
	ReceiptLocator         DestinationReceiptLocator    `json:"receipt_locator"`
	Conformance            ConformanceEvidenceReference `json:"conformance"`
}

func (p DestinationTombstoneReadBackPolicy) Validate() error {
	if !isConcreteTransportIdentifier(p.Operation) {
		return fmt.Errorf("destination tombstone read-back requires a concrete operation")
	}
	if p.MaxRecords < 1 || p.MaxRecords > 10000 {
		return fmt.Errorf("destination tombstone read-back max_records must be between 1 and 10000")
	}
	if p.MaxAttempts < 1 || p.MaxAttempts > 10 {
		return fmt.Errorf("destination tombstone read-back max_attempts must be between 1 and 10")
	}
	if p.TimeoutMilliseconds < 1 || p.TimeoutMilliseconds > 60000 {
		return fmt.Errorf("destination tombstone read-back timeout_milliseconds must be between 1 and 60000")
	}
	if p.RetryDelayMilliseconds < 0 || p.RetryDelayMilliseconds > 10000 {
		return fmt.Errorf("destination tombstone read-back retry_delay_milliseconds must be between 0 and 10000")
	}
	if err := validateDestinationReadBackFields("tombstone identity", p.Identity); err != nil {
		return err
	}
	if err := p.ReceiptLocator.Validate(); err != nil {
		return err
	}
	return p.Conformance.Validate()
}

func (p DestinationReadBackPolicy) Validate() error {
	if !isConcreteTransportIdentifier(p.Operation) {
		return fmt.Errorf("destination read-back requires a concrete operation")
	}
	if p.MaxRecords < 1 || p.MaxRecords > 10000 {
		return fmt.Errorf("destination read-back max_records must be between 1 and 10000")
	}
	if p.MaxAttempts < 1 || p.MaxAttempts > 10 {
		return fmt.Errorf("destination read-back max_attempts must be between 1 and 10")
	}
	if p.TimeoutMilliseconds < 1 || p.TimeoutMilliseconds > 60000 {
		return fmt.Errorf("destination read-back timeout_milliseconds must be between 1 and 60000")
	}
	if p.RetryDelayMilliseconds < 0 || p.RetryDelayMilliseconds > 10000 {
		return fmt.Errorf("destination read-back retry_delay_milliseconds must be between 0 and 10000")
	}
	if err := validateDestinationReadBackFields("identity", p.Identity); err != nil {
		return err
	}
	if err := validateDestinationReadBackFields("expected", p.Expected); err != nil {
		return err
	}
	if err := p.ReceiptLocator.Validate(); err != nil {
		return err
	}
	return p.Conformance.Validate()
}

func validateDestinationReadBackFields(kind string, fields []DestinationReadBackField) error {
	if len(fields) == 0 {
		return fmt.Errorf("destination read-back %s fields are required", kind)
	}
	provider := make(map[string]struct{}, len(fields))
	expected := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		if strings.TrimSpace(field.ProviderField) == "" || len(field.ProviderField) > 256 || strings.TrimSpace(field.ExpectedField) == "" || len(field.ExpectedField) > 256 {
			return fmt.Errorf("destination read-back %s fields require concrete provider and expected names", kind)
		}
		if _, duplicate := provider[field.ProviderField]; duplicate {
			return fmt.Errorf("destination read-back %s duplicates provider field %q", kind, field.ProviderField)
		}
		if _, duplicate := expected[field.ExpectedField]; duplicate {
			return fmt.Errorf("destination read-back %s duplicates expected field %q", kind, field.ExpectedField)
		}
		provider[field.ProviderField] = struct{}{}
		expected[field.ExpectedField] = struct{}{}
	}
	return nil
}

// SourceTransportDescriptor is the source side of a connector's transport
// declaration. EligibleStreams are the only streams this role may receive.
type SourceTransportDescriptor struct {
	Executor        TransportExecutorReference `json:"executor"`
	EligibleStreams []string                   `json:"eligible_streams"`
	Modes           []synccontract.Mode        `json:"modes"`
	// OrderedPipeline declares that this exact endpoint can safely have its
	// next bounded extraction overlap a prior ordered destination apply. It
	// does not declare source partitioning or unordered concurrent reads.
	OrderedPipeline bool                         `json:"ordered_pipeline,omitempty"`
	Delivery        DeliveryGuarantees           `json:"delivery"`
	Conformance     ConformanceEvidenceReference `json:"conformance"`
}

// DestinationTransportDescriptor is the destination side of a connector's
// transport declaration. Its strategies are resolved from the requested mode
// before a source can begin reading.
type DestinationTransportDescriptor struct {
	Executor        TransportExecutorReference `json:"executor"`
	EligibleActions []string                   `json:"eligible_actions"`
	Modes           []synccontract.Mode        `json:"modes"`
	// OrderedPipeline declares that this exact endpoint can safely consume one
	// ordered bounded pipeline. It is not a declaration of multi-COPY support.
	OrderedPipeline bool `json:"ordered_pipeline,omitempty"`
	// CopyWorkerMaximum is the target-declared connection-pool ceiling that a
	// future immutable full-overwrite COPY lane may consume. It is a bounded
	// connection policy, never a generic per-run worker dial.
	CopyWorkerMaximum int                                 `json:"copy_worker_maximum,omitempty"`
	Delivery          DeliveryGuarantees                  `json:"delivery"`
	Conformance       ConformanceEvidenceReference        `json:"conformance"`
	Acknowledgement   TransportAcknowledgement            `json:"acknowledgement"`
	ApplyStrategies   []DestinationApplyStrategy          `json:"apply_strategies"`
	SourceBindings    []DestinationSourceBinding          `json:"source_bindings,omitempty"`
	ReadBack          *DestinationReadBackPolicy          `json:"read_back,omitempty"`
	TombstoneReadBack *DestinationTombstoneReadBackPolicy `json:"tombstone_read_back,omitempty"`
}

// SyncTransportDescriptor declares one or both roles a connector can perform.
// A role has no executable meaning until a separate registry matches its exact
// executor reference and an external conformance verifier admits it.
type SyncTransportDescriptor struct {
	Source      *SourceTransportDescriptor      `json:"source_transport,omitempty"`
	Destination *DestinationTransportDescriptor `json:"destination_transport,omitempty"`
}

func (r TransportExecutorReference) Validate() error {
	switch r.Family {
	case TransportExecutorFamilyDeclarativeAPI, TransportExecutorFamilyNativeAPI,
		TransportExecutorFamilyNativeDatabase, TransportExecutorFamilyFile,
		TransportExecutorFamilyQueue:
	default:
		return fmt.Errorf("unsupported transport executor family %q", r.Family)
	}
	if !isConcreteTransportIdentifier(r.ID) {
		return fmt.Errorf("transport executor requires a concrete executor ID")
	}
	return nil
}

// ValidateTransportExecutorFamily applies the closed integration-type to
// executor-family mapping at runtime preflight. `read`/`write` capability bits
// do not substitute for this admission rule.
func ValidateTransportExecutorFamily(integrationType string, executor TransportExecutorReference) error {
	if err := executor.Validate(); err != nil {
		return err
	}
	allowed := false
	switch strings.TrimSpace(strings.ToLower(integrationType)) {
	case "api":
		allowed = executor.Family == TransportExecutorFamilyDeclarativeAPI || executor.Family == TransportExecutorFamilyNativeAPI
	case "database":
		allowed = executor.Family == TransportExecutorFamilyNativeDatabase
	case "file":
		allowed = executor.Family == TransportExecutorFamilyFile
	case "queue":
		allowed = executor.Family == TransportExecutorFamilyQueue
	default:
		return fmt.Errorf("integration type %q has no closed transport executor family", integrationType)
	}
	if !allowed {
		return fmt.Errorf("integration type %q is incompatible with transport executor family %q", integrationType, executor.Family)
	}
	return nil
}

func (d DeliveryGuarantees) Validate() error {
	switch d.Idempotency {
	case DeliveryIdempotencyKeyed, DeliveryIdempotencyAtLeastOnce, DeliveryIdempotencyNone:
	default:
		return fmt.Errorf("unsupported transport idempotency guarantee %q", d.Idempotency)
	}
	switch d.Ordering {
	case DeliveryOrderingSource, DeliveryOrderingWindowCoalesced, DeliveryOrderingUnordered:
	default:
		return fmt.Errorf("unsupported transport ordering guarantee %q", d.Ordering)
	}
	switch d.Deletes {
	case DeliveryDeletesTombstone, DeliveryDeletesUnavailable:
	default:
		return fmt.Errorf("unsupported transport delete guarantee %q", d.Deletes)
	}
	return nil
}

func (r ConformanceEvidenceReference) Validate() error {
	if !isConcreteTransportIdentifier(r.Suite) || !isConcreteTransportIdentifier(r.RunID) {
		return fmt.Errorf("transport conformance reference requires concrete suite and run IDs")
	}
	return nil
}

func (s ApplyStrategy) Validate() error {
	switch s {
	case ApplyStrategyAppend, ApplyStrategyReplace, ApplyStrategyMerge, ApplyStrategyDedupe,
		ApplyStrategyDedupeHistory, ApplyStrategyChangeApply:
		return nil
	default:
		return fmt.Errorf("unsupported destination apply strategy %q", s)
	}
}

func (d SourceTransportDescriptor) Validate() error {
	if err := d.Executor.Validate(); err != nil {
		return err
	}
	if err := validateSourceTransportStreams(d.EligibleStreams); err != nil {
		return err
	}
	if err := validateTransportModes(d.Modes); err != nil {
		return err
	}
	if err := d.Delivery.Validate(); err != nil {
		return err
	}
	return d.Conformance.Validate()
}

func validateSourceTransportStreams(streams []string) error {
	if len(streams) == 1 && streams[0] == "*" {
		return nil
	}
	for _, stream := range streams {
		if stream == "*" {
			return fmt.Errorf("source eligible stream wildcard must be the only entry")
		}
	}
	return validateTransportNames("source eligible stream", streams)
}

func (d DestinationTransportDescriptor) Validate() error {
	if d.CopyWorkerMaximum < 0 || d.CopyWorkerMaximum > 8 {
		return fmt.Errorf("destination transport copy worker maximum must be zero or between 1 and 8")
	}
	if err := d.Executor.Validate(); err != nil {
		return err
	}
	if err := validateTransportNames("destination eligible action", d.EligibleActions); err != nil {
		return err
	}
	if err := validateDestinationTransportModes(d.Modes); err != nil {
		return err
	}
	if err := d.Delivery.Validate(); err != nil {
		return err
	}
	if err := d.Conformance.Validate(); err != nil {
		return err
	}
	if err := validateDestinationSourceBindings(d.SourceBindings); err != nil {
		return err
	}
	if d.ReadBack != nil {
		if err := d.ReadBack.Validate(); err != nil {
			return err
		}
	}
	if d.TombstoneReadBack != nil {
		if err := d.TombstoneReadBack.Validate(); err != nil {
			return err
		}
	}
	switch d.Acknowledgement {
	case TransportAcknowledgementDurableWarehouse, TransportAcknowledgementNone:
	default:
		return fmt.Errorf("unsupported destination acknowledgement policy %q", d.Acknowledgement)
	}

	strategies := make(map[synccontract.Mode]struct{}, len(d.ApplyStrategies))
	strategyActions := make(map[synccontract.Mode]map[string]struct{}, len(d.ApplyStrategies))
	declaredActions := make(map[string]struct{}, len(d.ApplyStrategies)*2)
	for _, strategy := range d.ApplyStrategies {
		if err := strategy.Mode.Validate(); err != nil {
			return err
		}
		if !containsTransportMode(d.Modes, strategy.Mode) {
			return fmt.Errorf("destination apply strategy mode %q is not a declared destination mode", strategy.Mode)
		}
		if err := strategy.Strategy.Validate(); err != nil {
			return err
		}
		if strategy.ReadBack != nil {
			if err := strategy.ReadBack.Validate(); err != nil {
				return fmt.Errorf("destination action %q read-back: %w", strategy.Action, err)
			}
		}
		if strategy.TombstoneReadBack != nil {
			if err := strategy.TombstoneReadBack.Validate(); err != nil {
				return fmt.Errorf("destination tombstone action %q read-back: %w", strategy.TombstoneAction, err)
			}
		}
		if strategy.Mode == synccontract.ModeChangeCapture && strategy.Strategy != ApplyStrategyChangeApply {
			return fmt.Errorf("destination change_capture mode requires change_apply strategy, got %q", strategy.Strategy)
		}
		if strategy.Mode != synccontract.ModeChangeCapture && strategy.Strategy == ApplyStrategyChangeApply {
			return fmt.Errorf("destination change_apply strategy is only valid for change_capture mode")
		}
		if !containsTransportName(d.EligibleActions, strategy.Action) {
			return fmt.Errorf("destination apply strategy action %q is not an eligible action", strategy.Action)
		}
		if strategy.TombstoneAction != "" && !containsTransportName(d.EligibleActions, strategy.TombstoneAction) {
			return fmt.Errorf("destination tombstone action %q is not an eligible action", strategy.TombstoneAction)
		}
		if strategy.TombstoneAction == strategy.Action && strategy.TombstoneAction != "" {
			return fmt.Errorf("destination tombstone action %q cannot equal its ordinary apply action", strategy.Action)
		}
		if strategyActions[strategy.Mode] == nil {
			strategyActions[strategy.Mode] = make(map[string]struct{})
		}
		if _, exists := strategyActions[strategy.Mode][strategy.Action]; exists {
			return fmt.Errorf("destination transport declares duplicate apply strategy action %q for sync mode %q", strategy.Action, strategy.Mode)
		}
		strategyActions[strategy.Mode][strategy.Action] = struct{}{}
		declaredActions[strategy.Action] = struct{}{}
		if strategy.TombstoneAction != "" {
			declaredActions[strategy.TombstoneAction] = struct{}{}
		}
		strategies[strategy.Mode] = struct{}{}
	}
	for _, mode := range d.Modes {
		if _, exists := strategies[mode]; !exists {
			return fmt.Errorf("destination transport is missing declared apply strategy for sync mode %q", mode)
		}
	}
	for _, action := range d.EligibleActions {
		if _, found := declaredActions[action]; !found {
			return fmt.Errorf("destination eligible action %q has no declared apply strategy", action)
		}
	}
	for action := range declaredActions {
		if !containsTransportName(d.EligibleActions, action) {
			return fmt.Errorf("destination apply strategy action %q is not an eligible action", action)
		}
	}
	for _, binding := range d.SourceBindings {
		if binding.Action == "" {
			continue
		}
		if _, found := declaredActions[binding.Action]; !found {
			return fmt.Errorf("destination source binding action %q has no reachable apply strategy", binding.Action)
		}
	}
	return nil
}

func validateDestinationTransportModes(modes []synccontract.Mode) error {
	if err := validateTransportModes(modes); err != nil {
		return err
	}
	if containsTransportMode(modes, synccontract.ModeChangeCapture) {
		return fmt.Errorf("destination transport cannot declare change_capture mode; change capture is source-only into the connection warehouse")
	}
	return nil
}

func validateDestinationSourceBindings(bindings []DestinationSourceBinding) error {
	for index, binding := range bindings {
		if err := binding.Validate(); err != nil {
			return fmt.Errorf("destination source binding: %w", err)
		}
		for previousIndex := 0; previousIndex < index; previousIndex++ {
			previous := bindings[previousIndex]
			if previous.Executor != binding.Executor || previous.Action != binding.Action || !sourceBindingStreamsOverlap(previous.EligibleStreams, binding.EligibleStreams) {
				continue
			}
			if binding.Action == "" {
				return fmt.Errorf("destination source binding duplicates legacy executor %q stream selection", binding.Executor.ID)
			}
			return fmt.Errorf("destination source binding duplicates action %q for executor %q stream selection", binding.Action, binding.Executor.ID)
		}
	}
	return nil
}

func sourceBindingStreamsOverlap(left, right []string) bool {
	for _, leftStream := range left {
		for _, rightStream := range right {
			if leftStream == "*" || rightStream == "*" || leftStream == rightStream {
				return true
			}
		}
	}
	return false
}

// SourceBindingFor returns the definition-owned binding for an exact source
// executor and source stream. Destinations without bindings retain their
// existing transport semantics; destinations that declare bindings have a
// positive source admission allowlist.
func (d DestinationTransportDescriptor) SourceBindingFor(source TransportExecutorReference, stream string) (DestinationSourceBinding, bool) {
	for _, binding := range d.SourceBindings {
		if binding.Action != "" || binding.Executor != source || !containsSourceTransportStream(binding.EligibleStreams, stream) {
			continue
		}
		return binding.Clone(), true
	}
	return DestinationSourceBinding{}, false
}

// SourceBindingForAction resolves an exact persisted destination action. If a
// source/stream has action-owned bindings, no legacy mapping can serve as a
// fallback; that prevents one action's field mapping from being reused by a
// neighboring action. Specialized legacy adapters may still use a binding
// without Action through SourceBindingFor.
func (d DestinationTransportDescriptor) SourceBindingForAction(source TransportExecutorReference, stream, action string) (DestinationSourceBinding, bool) {
	var legacy DestinationSourceBinding
	legacyFound := false
	actionOwned := false
	for _, binding := range d.SourceBindings {
		if binding.Executor != source || !containsSourceTransportStream(binding.EligibleStreams, stream) {
			continue
		}
		if binding.Action == "" {
			legacy, legacyFound = binding, true
			continue
		}
		actionOwned = true
		if binding.Action == action {
			return binding.Clone(), true
		}
	}
	if actionOwned {
		return DestinationSourceBinding{}, false
	}
	if legacyFound {
		return legacy.Clone(), true
	}
	return DestinationSourceBinding{}, false
}

func containsSourceTransportStream(streams []string, want string) bool {
	for _, stream := range streams {
		if stream == want || stream == "*" {
			return true
		}
	}
	return false
}

// ApplyStrategyFor resolves the descriptor-owned strategy for mode. It never
// returns a default, which prevents a legacy `upsert` fallback from appearing
// in a closed transport path.
func (d DestinationTransportDescriptor) ApplyStrategyFor(mode synccontract.Mode) (DestinationApplyStrategy, error) {
	return d.ApplyStrategyForAction(mode, "")
}

// ApplyStrategyForAction resolves the exact definition-owned action selected
// by a persisted connection stream. Empty selection is accepted only when the
// descriptor has exactly one strategy for the mode, preserving existing
// single-action destinations while refusing an ambiguous multi-action route.
func (d DestinationTransportDescriptor) ApplyStrategyForAction(mode synccontract.Mode, action string) (DestinationApplyStrategy, error) {
	if err := d.Validate(); err != nil {
		return DestinationApplyStrategy{}, err
	}
	var candidate DestinationApplyStrategy
	count := 0
	for _, strategy := range d.ApplyStrategies {
		if strategy.Mode != mode {
			continue
		}
		if action != "" && strategy.Action == action {
			return strategy, nil
		}
		candidate = strategy
		count++
	}
	if action != "" {
		return DestinationApplyStrategy{}, fmt.Errorf("destination transport does not declare action %q for sync mode %q", action, mode)
	}
	if count == 1 {
		return candidate, nil
	}
	if count > 1 {
		return DestinationApplyStrategy{}, fmt.Errorf("destination transport requires a persisted action selection for sync mode %q", mode)
	}
	return DestinationApplyStrategy{}, fmt.Errorf("destination transport does not support sync mode %q", mode)
}

func (d SyncTransportDescriptor) Validate() error {
	if d.Source == nil && d.Destination == nil {
		return fmt.Errorf("sync transport descriptor must declare a source or destination transport")
	}
	if d.Source != nil {
		if err := d.Source.Validate(); err != nil {
			return fmt.Errorf("source transport: %w", err)
		}
	}
	if d.Destination != nil {
		if err := d.Destination.Validate(); err != nil {
			return fmt.Errorf("destination transport: %w", err)
		}
	}
	return nil
}

// Clone returns a fully independent descriptor projection. Inspection and
// runtime preflight therefore cannot mutate a connector's authored contract.
func (d SyncTransportDescriptor) Clone() *SyncTransportDescriptor {
	clone := d
	if d.Source != nil {
		source := *d.Source
		source.EligibleStreams = append([]string(nil), d.Source.EligibleStreams...)
		source.Modes = append([]synccontract.Mode(nil), d.Source.Modes...)
		clone.Source = &source
	}
	if d.Destination != nil {
		destination := *d.Destination
		destination.EligibleActions = append([]string(nil), d.Destination.EligibleActions...)
		destination.Modes = append([]synccontract.Mode(nil), d.Destination.Modes...)
		destination.ApplyStrategies = make([]DestinationApplyStrategy, len(d.Destination.ApplyStrategies))
		for index, strategy := range d.Destination.ApplyStrategies {
			destination.ApplyStrategies[index] = strategy
			if strategy.ReadBack != nil {
				readBack := *strategy.ReadBack
				readBack.Identity = append([]DestinationReadBackField(nil), strategy.ReadBack.Identity...)
				readBack.Expected = append([]DestinationReadBackField(nil), strategy.ReadBack.Expected...)
				destination.ApplyStrategies[index].ReadBack = &readBack
			}
			if strategy.TombstoneReadBack != nil {
				readBack := *strategy.TombstoneReadBack
				readBack.Identity = append([]DestinationReadBackField(nil), strategy.TombstoneReadBack.Identity...)
				destination.ApplyStrategies[index].TombstoneReadBack = &readBack
			}
		}
		destination.SourceBindings = make([]DestinationSourceBinding, len(d.Destination.SourceBindings))
		for index, binding := range d.Destination.SourceBindings {
			destination.SourceBindings[index] = binding.Clone()
		}
		if d.Destination.ReadBack != nil {
			readBack := *d.Destination.ReadBack
			readBack.Identity = append([]DestinationReadBackField(nil), d.Destination.ReadBack.Identity...)
			readBack.Expected = append([]DestinationReadBackField(nil), d.Destination.ReadBack.Expected...)
			destination.ReadBack = &readBack
		}
		if d.Destination.TombstoneReadBack != nil {
			readBack := *d.Destination.TombstoneReadBack
			readBack.Identity = append([]DestinationReadBackField(nil), d.Destination.TombstoneReadBack.Identity...)
			destination.TombstoneReadBack = &readBack
		}
		clone.Destination = &destination
	}
	return &clone
}

// SyncTransportDescriptorProvider supplies an additive descriptor for a
// connector that has not yet adopted DefinitionProvider.
type SyncTransportDescriptorProvider interface {
	SyncTransportDescriptor() *SyncTransportDescriptor
}

// SyncTransportDescriptorOf returns a defensive descriptor projection. A
// missing descriptor is intentionally not inferred from Read/Write methods.
func SyncTransportDescriptorOf(c Connector) (*SyncTransportDescriptor, bool) {
	if c == nil {
		return nil, false
	}
	if def, ok := DefinitionOf(c); ok && def.SyncTransport != nil {
		return def.SyncTransport.Clone(), true
	}
	if provider, ok := c.(SyncTransportDescriptorProvider); ok {
		descriptor := provider.SyncTransportDescriptor()
		if descriptor != nil {
			return descriptor.Clone(), true
		}
	}
	return nil, false
}

func SourceTransportDescriptorOf(c Connector) (*SourceTransportDescriptor, bool) {
	descriptor, ok := SyncTransportDescriptorOf(c)
	if !ok || descriptor.Source == nil {
		return nil, false
	}
	source := *descriptor.Source
	source.EligibleStreams = append([]string(nil), descriptor.Source.EligibleStreams...)
	source.Modes = append([]synccontract.Mode(nil), descriptor.Source.Modes...)
	return &source, true
}

func DestinationTransportDescriptorOf(c Connector) (*DestinationTransportDescriptor, bool) {
	descriptor, ok := SyncTransportDescriptorOf(c)
	if !ok || descriptor.Destination == nil {
		return nil, false
	}
	destination := *descriptor.Destination
	destination.EligibleActions = append([]string(nil), descriptor.Destination.EligibleActions...)
	destination.Modes = append([]synccontract.Mode(nil), descriptor.Destination.Modes...)
	destination.ApplyStrategies = append([]DestinationApplyStrategy(nil), descriptor.Destination.ApplyStrategies...)
	return &destination, true
}

// SyncTransportEligibility is the metadata-only inspection projection. A
// declared role remains short of executable until runtime preflight and the
// external verifier admit it.
type SyncTransportEligibility struct {
	Source      TransportRoleEligibility `json:"source"`
	Destination TransportRoleEligibility `json:"destination"`
}

type TransportRoleEligibility struct {
	Status          string                      `json:"status"`
	Executor        *TransportExecutorReference `json:"executor,omitempty"`
	Streams         []string                    `json:"streams,omitempty"`
	Actions         []string                    `json:"actions,omitempty"`
	Modes           []synccontract.Mode         `json:"modes,omitempty"`
	ApplyStrategies []DestinationApplyStrategy  `json:"apply_strategies,omitempty"`
	Acknowledgement TransportAcknowledgement    `json:"acknowledgement,omitempty"`
}

func SyncTransportEligibilityOf(c Connector) SyncTransportEligibility {
	eligibility := SyncTransportEligibility{
		Source:      TransportRoleEligibility{Status: "unsupported"},
		Destination: TransportRoleEligibility{Status: "unsupported"},
	}
	descriptor, ok := SyncTransportDescriptorOf(c)
	if !ok {
		return eligibility
	}
	if descriptor.Source != nil && descriptor.Source.Validate() == nil {
		executor := descriptor.Source.Executor
		eligibility.Source = TransportRoleEligibility{
			Status:   "declared",
			Executor: &executor,
			Streams:  append([]string(nil), descriptor.Source.EligibleStreams...),
			Modes:    append([]synccontract.Mode(nil), descriptor.Source.Modes...),
		}
	}
	if descriptor.Destination != nil && descriptor.Destination.Validate() == nil {
		executor := descriptor.Destination.Executor
		eligibility.Destination = TransportRoleEligibility{
			Status:          "declared",
			Executor:        &executor,
			Actions:         append([]string(nil), descriptor.Destination.EligibleActions...),
			Modes:           append([]synccontract.Mode(nil), descriptor.Destination.Modes...),
			ApplyStrategies: append([]DestinationApplyStrategy(nil), descriptor.Destination.ApplyStrategies...),
			Acknowledgement: descriptor.Destination.Acknowledgement,
		}
	}
	return eligibility
}

func validateTransportModes(modes []synccontract.Mode) error {
	if len(modes) == 0 {
		return fmt.Errorf("transport descriptor requires at least one sync mode")
	}
	seen := make(map[synccontract.Mode]struct{}, len(modes))
	for _, mode := range modes {
		if err := mode.Validate(); err != nil {
			return err
		}
		if _, exists := seen[mode]; exists {
			return fmt.Errorf("transport descriptor declares duplicate sync mode %q", mode)
		}
		seen[mode] = struct{}{}
	}
	return nil
}

func validateTransportNames(label string, names []string) error {
	if len(names) == 0 {
		return fmt.Errorf("%s requires at least one name", label)
	}
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if !isConcreteTransportIdentifier(name) {
			return fmt.Errorf("%s %q must be a concrete identifier", label, name)
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("%s %q is duplicated", label, name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func containsTransportName(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
}

func containsTransportMode(modes []synccontract.Mode, want synccontract.Mode) bool {
	for _, mode := range modes {
		if mode == want {
			return true
		}
	}
	return false
}

func isConcreteTransportIdentifier(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	normalized := strings.ReplaceAll(value, "-", "_")
	for _, forbidden := range []string{"generic", "sql", "http", "shell", "generic_sql", "generic_http", "generic_shell"} {
		if normalized == forbidden {
			return false
		}
	}
	for _, character := range value {
		if unicode.IsLower(character) || unicode.IsDigit(character) || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}
