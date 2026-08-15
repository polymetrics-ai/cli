package engine

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

// PollingPreflightSourceExecutor is the runtime half of a definition-owned
// polling source declaration. It intentionally contains no read method: the
// source executor in the following delivery slice receives a ResolvedPollingWatermark
// only after this no-I/O admission has succeeded.
type PollingPreflightSourceExecutor interface {
	PollingSourceExecutorReference() connectors.TransportExecutorReference
	PollingSourceConformanceEvidence() PollingWatermarkConformanceEvidence
}

// PollingPreflightApplyExecutor is the runtime half of the closed target-apply
// declaration. It does not expose DML; #3859 owns that executor.
type PollingPreflightApplyExecutor interface {
	PollingApplyExecutorReference() connectors.TransportExecutorReference
	PollingApplyConformanceEvidence() PollingWatermarkConformanceEvidence
}

// PollingPreflightRegistry stores exact registered source and apply executors.
// It is deliberately independent of ChangefeedExecutor and commandrunner:
// polling-watermark admission cannot promote a CDC capability or REST command.
type PollingPreflightRegistry struct {
	mu      sync.RWMutex
	sources map[connectors.TransportExecutorReference]PollingPreflightSourceExecutor
	applies map[connectors.TransportExecutorReference]PollingPreflightApplyExecutor
}

// NewPollingPreflightRegistry returns an empty registry. Executors are
// registered explicitly so a definition cannot become executable on its own.
func NewPollingPreflightRegistry() *PollingPreflightRegistry {
	return &PollingPreflightRegistry{
		sources: make(map[connectors.TransportExecutorReference]PollingPreflightSourceExecutor),
		applies: make(map[connectors.TransportExecutorReference]PollingPreflightApplyExecutor),
	}
}

// RegisterSource adds a single concrete polling source executor.
func (r *PollingPreflightRegistry) RegisterSource(executor PollingPreflightSourceExecutor) error {
	if r == nil {
		return fmt.Errorf("polling preflight registry is required")
	}
	if isNilPollingPreflightExecutor(executor) {
		return fmt.Errorf("polling source executor is required")
	}
	reference := executor.PollingSourceExecutorReference()
	if err := validatePollingRuntimeReference(reference); err != nil {
		return fmt.Errorf("source polling executor: %w", err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sources == nil {
		r.sources = make(map[connectors.TransportExecutorReference]PollingPreflightSourceExecutor)
	}
	if _, exists := r.sources[reference]; exists {
		return fmt.Errorf("source polling executor %q is already registered", reference.ID)
	}
	r.sources[reference] = executor
	return nil
}

// RegisterApply adds a single concrete polling target-apply executor.
func (r *PollingPreflightRegistry) RegisterApply(executor PollingPreflightApplyExecutor) error {
	if r == nil {
		return fmt.Errorf("polling preflight registry is required")
	}
	if isNilPollingPreflightExecutor(executor) {
		return fmt.Errorf("target polling executor is required")
	}
	reference := executor.PollingApplyExecutorReference()
	if err := validatePollingRuntimeReference(reference); err != nil {
		return fmt.Errorf("target polling executor: %w", err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.applies == nil {
		r.applies = make(map[connectors.TransportExecutorReference]PollingPreflightApplyExecutor)
	}
	if _, exists := r.applies[reference]; exists {
		return fmt.Errorf("target polling executor %q is already registered", reference.ID)
	}
	r.applies[reference] = executor
	return nil
}

// ResolvedPollingWatermark is an immutable successful admission result. It
// has no source records and triggers no connection or I/O on its own.
type ResolvedPollingWatermark struct {
	Mode        synccontract.Mode
	Declaration *connectors.PollingWatermarkDescriptor
	Object      connectors.PollingCatalogObject
	Source      PollingPreflightSourceExecutor
	Apply       PollingPreflightApplyExecutor
}

// PollingPreflight validates a native polling declaration, resolves exact
// source/apply registrations, and checks the immutable #3856 corpus proof
// before source I/O is permitted. It does not call a source or target method.
func PollingPreflight(ctx context.Context, registry *PollingPreflightRegistry, declaration *connectors.PollingWatermarkDescriptor, object connectors.PollingCatalogObject, mode synccontract.Mode) (ResolvedPollingWatermark, error) {
	if ctx == nil {
		return ResolvedPollingWatermark{}, fmt.Errorf("polling preflight context is required")
	}
	if err := ctx.Err(); err != nil {
		return ResolvedPollingWatermark{}, err
	}
	if registry == nil {
		return ResolvedPollingWatermark{}, fmt.Errorf("polling preflight registry is required")
	}
	if declaration == nil {
		return ResolvedPollingWatermark{}, fmt.Errorf("polling watermark declaration is required")
	}
	if declaration.Status != connectors.PollingWatermarkStatusImplemented {
		return ResolvedPollingWatermark{}, fmt.Errorf("polling watermark declaration is not implemented")
	}
	if err := declaration.Validate(); err != nil {
		return ResolvedPollingWatermark{}, fmt.Errorf("polling watermark declaration: %w", err)
	}
	if err := mode.Validate(); err != nil {
		return ResolvedPollingWatermark{}, err
	}
	if !containsPollingPreflightMode(declaration.Source.Modes, mode) {
		return ResolvedPollingWatermark{}, fmt.Errorf("polling source does not support sync mode %q", mode)
	}
	if err := validatePollingCatalogObject(declaration.Source, object); err != nil {
		return ResolvedPollingWatermark{}, err
	}

	registry.mu.RLock()
	source, sourceRegistered := registry.sources[declaration.Source.Executor]
	apply, applyRegistered := registry.applies[declaration.Target.Executor]
	registry.mu.RUnlock()
	if !sourceRegistered || isNilPollingPreflightExecutor(source) {
		return ResolvedPollingWatermark{}, fmt.Errorf("source polling executor %q is not registered", declaration.Source.Executor.ID)
	}
	if source.PollingSourceExecutorReference() != declaration.Source.Executor {
		return ResolvedPollingWatermark{}, fmt.Errorf("registered source polling executor does not match the declaration")
	}
	if !source.PollingSourceConformanceEvidence().matchesRequired() {
		return ResolvedPollingWatermark{}, fmt.Errorf("source immutable polling conformance evidence is missing or stale")
	}
	if !applyRegistered || isNilPollingPreflightExecutor(apply) {
		return ResolvedPollingWatermark{}, fmt.Errorf("target polling executor %q is not registered", declaration.Target.Executor.ID)
	}
	if apply.PollingApplyExecutorReference() != declaration.Target.Executor {
		return ResolvedPollingWatermark{}, fmt.Errorf("registered target polling executor does not match the declaration")
	}
	if !apply.PollingApplyConformanceEvidence().matchesRequired() {
		return ResolvedPollingWatermark{}, fmt.Errorf("target immutable polling conformance evidence is missing or stale")
	}
	if err := validatePollingApplyMode(declaration.Target, mode); err != nil {
		return ResolvedPollingWatermark{}, err
	}
	if err := ctx.Err(); err != nil {
		return ResolvedPollingWatermark{}, err
	}

	return ResolvedPollingWatermark{
		Mode:        mode,
		Declaration: declaration.Clone(),
		Object:      clonePollingCatalogObject(object),
		Source:      source,
		Apply:       apply,
	}, nil
}

// PollingModeEligibility exposes inspection/catalog/connection eligibility as
// a projection of the real runtime preflight. It never makes a source read and
// it does not duplicate the preflight decision tree in a generator.
type PollingModeEligibility struct {
	Mode   synccontract.Mode
	Status string
	Reason string
}

// PollingModeEligibilityOf returns one deterministic row per definition-owned
// canonical mode. A row is executable only when PollingPreflight succeeds.
func PollingModeEligibilityOf(ctx context.Context, registry *PollingPreflightRegistry, declaration *connectors.PollingWatermarkDescriptor, object connectors.PollingCatalogObject) []PollingModeEligibility {
	if declaration == nil {
		return nil
	}
	modes := append([]synccontract.Mode(nil), declaration.Source.Modes...)
	sort.Slice(modes, func(i, j int) bool { return modes[i] < modes[j] })
	rows := make([]PollingModeEligibility, 0, len(modes))
	for _, mode := range modes {
		_, err := PollingPreflight(ctx, registry, declaration, object, mode)
		row := PollingModeEligibility{Mode: mode, Status: "blocked"}
		if err == nil {
			row.Status = "implemented"
		} else {
			row.Reason = err.Error()
		}
		rows = append(rows, row)
	}
	return rows
}

func validatePollingRuntimeReference(reference connectors.TransportExecutorReference) error {
	if err := reference.Validate(); err != nil {
		return err
	}
	if reference.Family != connectors.TransportExecutorFamilyNativeDatabase {
		return fmt.Errorf("polling executor must use native_database family")
	}
	return nil
}

func validatePollingCatalogObject(declaration connectors.PollingWatermarkSourceDescriptor, object connectors.PollingCatalogObject) error {
	if object.Kind != declaration.Object.Kind {
		return fmt.Errorf("discovered polling catalog object kind %q does not match declaration", object.Kind)
	}
	if len(object.NameParts) < 2 || len(object.NameParts) > 3 {
		return fmt.Errorf("discovered polling catalog object requires qualified name parts")
	}
	for _, part := range object.NameParts {
		if !validPollingCatalogName(part) {
			return fmt.Errorf("discovered polling catalog object contains an unsafe name part")
		}
	}
	columns := make(map[string]struct{}, len(object.Columns))
	for _, column := range object.Columns {
		if !validPollingCatalogName(column) {
			return fmt.Errorf("discovered polling catalog object contains an unsafe column")
		}
		columns[column] = struct{}{}
	}
	orderingFields := append(declaration.Ordering.Watermark.CatalogColumns(), declaration.Ordering.TieBreaker.CatalogColumns()...)
	for _, field := range orderingFields {
		if _, found := columns[field]; !found {
			return fmt.Errorf("discovered polling catalog object lacks ordering column %q", field)
		}
	}
	if declaration.DeleteVisibility == connectors.PollingDeleteVisibilityTombstone {
		if _, found := columns[declaration.SoftDeleteField]; !found {
			return fmt.Errorf("discovered polling catalog object lacks soft-delete column %q", declaration.SoftDeleteField)
		}
	}
	return nil
}

func validatePollingApplyMode(target connectors.PollingApplyDescriptor, mode synccontract.Mode) error {
	strategy, required := requiredPollingApplyStrategy(mode)
	if !required || !containsPollingPreflightStrategy(target.Strategies, strategy) {
		return fmt.Errorf("target polling apply does not support sync mode %q", mode)
	}
	if mode == synccontract.ModeFullOverwrite && target.Staging != connectors.PollingStagingReplaceSupported {
		return fmt.Errorf("target polling apply requires staging/replace for sync mode %q", mode)
	}
	if (mode == synccontract.ModeIncrementalUpsert || mode == synccontract.ModeIncrementalDedupe || mode == synccontract.ModeIncrementalDedupeHistory) && !target.ConditionalOrderFence {
		return fmt.Errorf("target polling apply requires conditional ordering fence for sync mode %q", mode)
	}
	if mode == synccontract.ModeIncrementalDedupeHistory && (target.Transaction != connectors.PollingTransactionRequired || !target.RetrySafeCloseAndInsert || target.ValidityWindow != connectors.PollingValidityWindowSupported) {
		return fmt.Errorf("history mode requires transaction and retry-safe close-and-insert")
	}
	return nil
}

func requiredPollingApplyStrategy(mode synccontract.Mode) (connectors.PollingApplyStrategy, bool) {
	switch mode {
	case synccontract.ModeFullAppend, synccontract.ModeIncrementalAppend:
		return connectors.PollingApplyStrategyAppend, true
	case synccontract.ModeFullOverwrite:
		return connectors.PollingApplyStrategyReplace, true
	case synccontract.ModeIncrementalUpsert:
		return connectors.PollingApplyStrategyMerge, true
	case synccontract.ModeIncrementalDedupe:
		return connectors.PollingApplyStrategyDedupe, true
	case synccontract.ModeIncrementalDedupeHistory:
		return connectors.PollingApplyStrategyDedupeHistory, true
	default:
		return "", false
	}
}

func containsPollingPreflightMode(modes []synccontract.Mode, expected synccontract.Mode) bool {
	for _, mode := range modes {
		if mode == expected {
			return true
		}
	}
	return false
}

func containsPollingPreflightStrategy(strategies []connectors.PollingApplyStrategy, expected connectors.PollingApplyStrategy) bool {
	for _, strategy := range strategies {
		if strategy == expected {
			return true
		}
	}
	return false
}

func clonePollingCatalogObject(object connectors.PollingCatalogObject) connectors.PollingCatalogObject {
	object.NameParts = append([]string(nil), object.NameParts...)
	object.Columns = append([]string(nil), object.Columns...)
	return object
}

func validPollingCatalogName(value string) bool {
	if value == "" || value != strings.TrimSpace(value) {
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

func isNilPollingPreflightExecutor(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
