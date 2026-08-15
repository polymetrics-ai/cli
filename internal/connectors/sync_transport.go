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
	DeliveryOrderingSource    DeliveryOrdering = "source_ordered"
	DeliveryOrderingUnordered DeliveryOrdering = "unordered"
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
	Mode     synccontract.Mode `json:"mode"`
	Strategy ApplyStrategy     `json:"strategy"`
	Action   string            `json:"action"`
}

// SourceTransportDescriptor is the source side of a connector's transport
// declaration. EligibleStreams are the only streams this role may receive.
type SourceTransportDescriptor struct {
	Executor        TransportExecutorReference   `json:"executor"`
	EligibleStreams []string                     `json:"eligible_streams"`
	Modes           []synccontract.Mode          `json:"modes"`
	Delivery        DeliveryGuarantees           `json:"delivery"`
	Conformance     ConformanceEvidenceReference `json:"conformance"`
}

// DestinationTransportDescriptor is the destination side of a connector's
// transport declaration. Its strategies are resolved from the requested mode
// before a source can begin reading.
type DestinationTransportDescriptor struct {
	Executor        TransportExecutorReference   `json:"executor"`
	EligibleActions []string                     `json:"eligible_actions"`
	Modes           []synccontract.Mode          `json:"modes"`
	Delivery        DeliveryGuarantees           `json:"delivery"`
	Conformance     ConformanceEvidenceReference `json:"conformance"`
	Acknowledgement TransportAcknowledgement     `json:"acknowledgement"`
	ApplyStrategies []DestinationApplyStrategy   `json:"apply_strategies"`
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
	case DeliveryOrderingSource, DeliveryOrderingUnordered:
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
	if err := validateTransportNames("source eligible stream", d.EligibleStreams); err != nil {
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

func (d DestinationTransportDescriptor) Validate() error {
	if err := d.Executor.Validate(); err != nil {
		return err
	}
	if err := validateTransportNames("destination eligible action", d.EligibleActions); err != nil {
		return err
	}
	if err := validateTransportModes(d.Modes); err != nil {
		return err
	}
	for _, mode := range d.Modes {
		if mode == synccontract.ModeChangeCapture {
			return fmt.Errorf("destination transport cannot declare change_capture mode")
		}
	}
	if err := d.Delivery.Validate(); err != nil {
		return err
	}
	if err := d.Conformance.Validate(); err != nil {
		return err
	}
	switch d.Acknowledgement {
	case TransportAcknowledgementDurableWarehouse, TransportAcknowledgementNone:
	default:
		return fmt.Errorf("unsupported destination acknowledgement policy %q", d.Acknowledgement)
	}

	strategies := make(map[synccontract.Mode]struct{}, len(d.ApplyStrategies))
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
		if !containsTransportName(d.EligibleActions, strategy.Action) {
			return fmt.Errorf("destination apply strategy action %q is not an eligible action", strategy.Action)
		}
		if _, exists := strategies[strategy.Mode]; exists {
			return fmt.Errorf("destination transport declares duplicate apply strategy for sync mode %q", strategy.Mode)
		}
		strategies[strategy.Mode] = struct{}{}
	}
	for _, mode := range d.Modes {
		if _, exists := strategies[mode]; !exists {
			return fmt.Errorf("destination transport is missing declared apply strategy for sync mode %q", mode)
		}
	}
	return nil
}

// ApplyStrategyFor resolves the descriptor-owned strategy for mode. It never
// returns a default, which prevents a legacy `upsert` fallback from appearing
// in a closed transport path.
func (d DestinationTransportDescriptor) ApplyStrategyFor(mode synccontract.Mode) (DestinationApplyStrategy, error) {
	if err := d.Validate(); err != nil {
		return DestinationApplyStrategy{}, err
	}
	for _, strategy := range d.ApplyStrategies {
		if strategy.Mode == mode {
			return strategy, nil
		}
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
		destination.ApplyStrategies = append([]DestinationApplyStrategy(nil), d.Destination.ApplyStrategies...)
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
