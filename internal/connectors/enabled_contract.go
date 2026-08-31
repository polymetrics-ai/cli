package connectors

import (
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"
)

// EnabledConnectorContract is the closed declaration that makes an enabled
// connector's provider lanes visible independently of certification. It is a
// definition-owned capability inventory, never a transport selector: runtime
// preflight and the selected executor still decide whether a request may run.
//
// A contract is intentionally additive. Connectors without this artifact keep
// their existing definition semantics while a cohort is being migrated.
type EnabledConnectorContract struct {
	SchemaVersion           int                                     `json:"schema_version"`
	Connector               string                                  `json:"connector"`
	SourceLock              EnabledContractSourceLock               `json:"source_lock"`
	SupplementalSourceLocks []EnabledContractSupplementalSourceLock `json:"supplemental_source_locks,omitempty"`
	Lanes                   []EnabledConnectorLane                  `json:"lanes"`
}

// EnabledContractSourceLock identifies the immutable source lock the contract
// reconciles. Hash and byte values are retained evidence, not an admission
// switch; source-import remains responsible for verifying retained bytes.
type EnabledContractSourceLock struct {
	Path              string `json:"path"`
	SHA256            string `json:"sha256"`
	Bytes             int64  `json:"bytes"`
	CanonicalEvidence bool   `json:"canonical_evidence"`
}

// EnabledContractSupplementalSourceLock is an additional immutable source
// inventory for one connector. It preserves provider documentation that is
// outside a locked primary OpenAPI artifact without replacing or rehashing
// that artifact. Every operation in a supplement must be claimed by one exact
// non-partition lane source selector.
type EnabledContractSupplementalSourceLock struct {
	Path       string `json:"path"`
	Operations int    `json:"operations"`
}

// EnabledConnectorLane is one fixed, discoverable provider lane. States are
// deliberately distinct from CLI command availability: an implemented lane
// can still contain source identities whose exact request shape remains a
// named foundation gap in Source.Projected.
type EnabledConnectorLane struct {
	Name      string                         `json:"name"`
	State     string                         `json:"state"`
	Reason    string                         `json:"reason"`
	Citations []EnabledContractCitation      `json:"citations"`
	Artifacts []string                       `json:"artifacts"`
	Source    EnabledContractSourceCoverage  `json:"source"`
	Transport *EnabledContractTransport      `json:"transport,omitempty"`
	Warehouse []EnabledContractWarehouseFlow `json:"warehouse_flows,omitempty"`
}

type EnabledContractCitation struct {
	URL      string `json:"url"`
	Location string `json:"location"`
}

// EnabledContractSourceCoverage is a bounded selector over the retained lock.
// Partition=true means this selector owns one mutually-exclusive member of
// the source denominator. Non-partition selectors are overlays (such as an
// ETL stream or sync transport) and must name their exact source IDs.
type EnabledContractSourceCoverage struct {
	SourceLock   string   `json:"source_lock,omitempty"`
	Partition    bool     `json:"partition"`
	Methods      []string `json:"methods,omitempty"`
	OperationIDs []string `json:"operation_ids,omitempty"`
	Coverage     string   `json:"coverage"`
	Expected     int      `json:"expected"`
	Implemented  int      `json:"implemented"`
	// MappedUnproven is source-semantic coverage that has been classified and
	// retained but has no field-complete executable declaration yet. It is not
	// an absent mapping and must not be relabeled as a runtime foundation gap.
	MappedUnproven     int `json:"mapped_unproven"`
	UnmappedMapping    int `json:"unmapped_mapping"`
	DeferredFoundation int `json:"deferred_foundation"`
	Unsupported        int `json:"unsupported_with_provider_evidence"`
}

const (
	EnabledCoverageComplete      = "complete"
	EnabledCoveragePartial       = "partial"
	EnabledCoverageNotApplicable = "not_applicable"
)

// EnabledContractWarehouseFlow makes the real local mediation explicit. It
// cannot represent an API-to-API shortcut: only provider_to_duckdb and
// duckdb_to_provider are accepted directions.
type EnabledContractWarehouseFlow struct {
	Direction string `json:"direction"`
	Runtime   string `json:"runtime"`
	Proof     string `json:"proof"`
}

// EnabledContractTransport keeps source evidence distinct from conservative
// runtime delivery policy. A missing provider order, cursor, or delete
// contract is retained as not_declared; it is never borrowed from an executor
// used by another connector.
type EnabledContractTransport struct {
	Modes           []string                                 `json:"modes"`
	RuntimeDelivery DeliveryGuarantees                       `json:"runtime_delivery"`
	Streams         []EnabledContractTransportStreamEvidence `json:"streams"`
	Destination     *EnabledContractTransportDestination     `json:"destination,omitempty"`
}

type EnabledContractTransportStreamEvidence struct {
	Stream          string `json:"stream"`
	SourceOperation string `json:"source_operation"`
	CursorEvidence  string `json:"cursor_evidence"`
	DeleteEvidence  string `json:"delete_evidence"`
	OrderEvidence   string `json:"order_evidence"`
}

// EnabledContractTransportDestination binds a local DuckDB source to the
// closed reverse plan/preview/approval lifecycle; it is not an API-to-API
// execution permission.
type EnabledContractTransportDestination struct {
	Kind             string `json:"kind"`
	ActionsArtifact  string `json:"actions_artifact"`
	ExpectedActions  int    `json:"expected_actions"`
	PlanExecutor     string `json:"plan_executor"`
	PreviewExecutor  string `json:"preview_executor"`
	ApprovalExecutor string `json:"approval_executor"`
}

// EnabledContractSourceOperation is the minimal immutable identity needed for
// denominator reconciliation. The engine never reads source locks itself;
// connectorgen supplies these values from its source-lock bridge.
type EnabledContractSourceOperation struct {
	ID     string
	Method string
}

const (
	EnabledLaneImplemented = "implemented"
	// EnabledLaneMappedUnproven keeps a semantic source-lane classification
	// visible without implying a runnable command or a missing runtime engine.
	EnabledLaneMappedUnproven = "mapped_unproven"
	EnabledLaneDeferred       = "deferred_foundation"
	EnabledLaneUnsupported    = "unsupported_with_provider_evidence"
)

var enabledContractLaneNames = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

// Validate rejects an incomplete enabled-lane artifact before it can become a
// misleading inspection or documentation claim. Provider-operation membership
// is checked separately by ReconcileSourceOperations, because the runtime does
// not own source locks.
func (c EnabledConnectorContract) Validate() error {
	if c.SchemaVersion != 1 {
		return fmt.Errorf("enabled connector contract has unsupported schema version %d", c.SchemaVersion)
	}
	if !enabledContractIdentifier(c.Connector) {
		return fmt.Errorf("enabled connector contract connector is invalid")
	}
	if !safeEnabledContractArtifact(c.SourceLock.Path) || len(c.SourceLock.SHA256) != 64 || c.SourceLock.Bytes <= 0 {
		return fmt.Errorf("enabled connector contract source_lock must retain a path, sha256, and positive byte count")
	}
	seenLocks := map[string]bool{c.SourceLock.Path: true}
	for _, sourceLock := range c.SupplementalSourceLocks {
		if !safeEnabledContractArtifact(sourceLock.Path) || sourceLock.Operations < 1 || seenLocks[sourceLock.Path] {
			return fmt.Errorf("enabled connector contract supplemental source lock is invalid or duplicated")
		}
		seenLocks[sourceLock.Path] = true
	}
	if len(c.Lanes) != len(enabledContractLaneNames) {
		return fmt.Errorf("enabled connector contract must declare exactly %d lanes", len(enabledContractLaneNames))
	}
	want := make(map[string]bool, len(enabledContractLaneNames))
	for _, name := range enabledContractLaneNames {
		want[name] = true
	}
	seen := make(map[string]bool, len(c.Lanes))
	for _, lane := range c.Lanes {
		if !want[lane.Name] || seen[lane.Name] {
			return fmt.Errorf("enabled connector contract lane %q is unknown or duplicated", lane.Name)
		}
		seen[lane.Name] = true
		if err := lane.Validate(); err != nil {
			return fmt.Errorf("enabled connector contract lane %q: %w", lane.Name, err)
		}
	}
	for _, name := range enabledContractLaneNames {
		if !seen[name] {
			return fmt.Errorf("enabled connector contract omits lane %q", name)
		}
	}
	return nil
}

func (l EnabledConnectorLane) Validate() error {
	if l.State != EnabledLaneImplemented && l.State != EnabledLaneMappedUnproven && l.State != EnabledLaneDeferred && l.State != EnabledLaneUnsupported {
		return fmt.Errorf("state %q is invalid", l.State)
	}
	if strings.TrimSpace(l.Reason) == "" {
		return fmt.Errorf("reason is required")
	}
	if len(l.Citations) == 0 || len(l.Artifacts) == 0 {
		return fmt.Errorf("citations and artifacts are required")
	}
	for _, citation := range l.Citations {
		if strings.TrimSpace(citation.URL) == "" || strings.TrimSpace(citation.Location) == "" {
			return fmt.Errorf("citation must retain url and location")
		}
	}
	for _, artifact := range l.Artifacts {
		if !safeEnabledContractArtifact(artifact) {
			return fmt.Errorf("artifact %q is not a safe connector-relative path", artifact)
		}
	}
	if err := l.Source.Validate(); err != nil {
		return err
	}
	for _, flow := range l.Warehouse {
		if (flow.Direction != "provider_to_duckdb" && flow.Direction != "duckdb_to_provider") || strings.TrimSpace(flow.Runtime) == "" || strings.TrimSpace(flow.Proof) == "" {
			return fmt.Errorf("warehouse flow must name a supported direction, runtime, and proof")
		}
	}
	if l.Transport != nil {
		if err := l.Transport.Validate(); err != nil {
			return err
		}
	}
	if (l.Name == "etl" || l.Name == "reverse_etl") && l.State == EnabledLaneImplemented && len(l.Warehouse) == 0 {
		return fmt.Errorf("implemented %s lane requires a warehouse-mediated runtime reference", l.Name)
	}
	if l.Name == "sync_transport" && l.State == EnabledLaneImplemented && l.Transport == nil {
		return fmt.Errorf("implemented sync_transport lane requires exact transport evidence")
	}
	if l.Name == "reverse_etl" && l.State == EnabledLaneImplemented && (l.Transport == nil || l.Transport.Destination == nil) {
		return fmt.Errorf("implemented reverse_etl lane requires a closed DuckDB-to-provider destination binding")
	}
	return nil
}

func (t EnabledContractTransport) Validate() error {
	if len(t.Modes) == 0 || len(t.Streams) == 0 {
		return fmt.Errorf("transport evidence requires modes and stream evidence")
	}
	if err := t.RuntimeDelivery.Validate(); err != nil {
		return fmt.Errorf("transport runtime delivery: %w", err)
	}
	seenModes := map[string]bool{}
	for _, mode := range t.Modes {
		if !enabledTransportMode(mode) || seenModes[mode] {
			return fmt.Errorf("transport mode %q is invalid or duplicated", mode)
		}
		seenModes[mode] = true
	}
	seenStreams := map[string]bool{}
	seenOperations := map[string]bool{}
	for _, stream := range t.Streams {
		if !enabledContractIdentifier(stream.Stream) || seenStreams[stream.Stream] {
			return fmt.Errorf("transport stream evidence has an invalid or duplicate stream")
		}
		seenStreams[stream.Stream] = true
		if stream.SourceOperation != "" {
			if !enabledContractSourceOperationIdentifier(stream.SourceOperation) || seenOperations[stream.SourceOperation] {
				return fmt.Errorf("transport stream evidence has an invalid or duplicate source operation")
			}
			seenOperations[stream.SourceOperation] = true
		}
		if !enabledContractTransportEvidence(stream.CursorEvidence) || !enabledContractTransportEvidence(stream.DeleteEvidence) || !enabledContractTransportEvidence(stream.OrderEvidence) {
			return fmt.Errorf("transport stream %q must preserve the closed evidence vocabulary", stream.Stream)
		}
	}
	if t.Destination == nil {
		return nil
	}
	d := t.Destination
	if d.Kind != "reverse_plan_preview_approval" || !safeEnabledContractArtifact(d.ActionsArtifact) || d.ExpectedActions < 1 || strings.TrimSpace(d.PlanExecutor) == "" || strings.TrimSpace(d.PreviewExecutor) == "" || strings.TrimSpace(d.ApprovalExecutor) == "" {
		return fmt.Errorf("transport destination must bind a closed reverse plan, preview, approval, and actions artifact")
	}
	return nil
}

func enabledContractTransportEvidence(value string) bool {
	return value == "not_declared" || value == "source_cited"
}

func enabledTransportMode(mode string) bool {
	switch mode {
	case "full_overwrite", "full_append", "incremental_append", "incremental_upsert", "incremental_dedupe", "incremental_dedupe_history":
		return true
	default:
		return false
	}
}

func (s EnabledContractSourceCoverage) Validate() error {
	if s.Expected < 0 || s.Implemented < 0 || s.MappedUnproven < 0 || s.UnmappedMapping < 0 || s.DeferredFoundation < 0 || s.Unsupported < 0 {
		return fmt.Errorf("source coverage counts must be non-negative")
	}
	if s.Expected != s.Implemented+s.MappedUnproven+s.UnmappedMapping+s.DeferredFoundation+s.Unsupported {
		return fmt.Errorf("source coverage expected=%d does not equal implemented+mapped_unproven+unmapped+deferred+unsupported=%d", s.Expected, s.Implemented+s.MappedUnproven+s.UnmappedMapping+s.DeferredFoundation+s.Unsupported)
	}
	wantCoverage := EnabledCoveragePartial
	if s.Expected == 0 {
		wantCoverage = EnabledCoverageNotApplicable
	} else if s.Implemented == s.Expected {
		wantCoverage = EnabledCoverageComplete
	}
	if s.Coverage != wantCoverage {
		return fmt.Errorf("source coverage %q is invalid for implemented=%d expected=%d; want %q", s.Coverage, s.Implemented, s.Expected, wantCoverage)
	}
	seenMethods := map[string]bool{}
	for _, method := range s.Methods {
		method = strings.ToUpper(strings.TrimSpace(method))
		if method == "" || seenMethods[method] {
			return fmt.Errorf("source coverage methods must be non-empty and unique")
		}
		seenMethods[method] = true
	}
	seenIDs := map[string]bool{}
	for _, id := range s.OperationIDs {
		if !enabledContractSourceOperationIdentifier(id) || seenIDs[id] {
			return fmt.Errorf("source coverage operation IDs must be non-empty, safe, and unique")
		}
		seenIDs[id] = true
	}
	if s.Partition {
		if len(s.Methods) == 0 && len(s.OperationIDs) == 0 {
			return fmt.Errorf("source partition requires one or more methods or operation IDs")
		}
		if len(s.Methods) > 0 && len(s.OperationIDs) > 0 {
			return fmt.Errorf("source partition cannot combine methods and operation IDs")
		}
		if len(s.OperationIDs) > 0 && len(s.OperationIDs) != s.Expected {
			return fmt.Errorf("source operation-ID partition count must match expected")
		}
	}
	if !s.Partition && len(s.OperationIDs) > 0 && len(s.OperationIDs) != s.Expected {
		return fmt.Errorf("non-partition source operation IDs must match expected count")
	}
	if s.SourceLock != "" && (s.Partition || len(s.OperationIDs) != s.Expected) {
		return fmt.Errorf("supplemental source coverage must be a non-partition exact operation selector")
	}
	return nil
}

// ReconcileSourceOperations proves that every locked operation belongs to one
// and only one partition lane, and that overlay lanes refer only to retained
// identities. It returns a typed error to the authoring layer; it cannot make
// a command executable.
func (c EnabledConnectorContract) ReconcileSourceOperations(operations []EnabledContractSourceOperation) error {
	if err := c.Validate(); err != nil {
		return err
	}
	partitionCounts := make(map[string]int)
	partitionMethods := make(map[string]string)
	partitionIDs := make(map[string]string)
	overlayIDs := make(map[string]bool)
	partitionUsesMethods := false
	partitionUsesIDs := false
	for _, lane := range c.Lanes {
		if lane.Source.SourceLock != "" {
			continue
		}
		if lane.Source.Partition {
			if len(lane.Source.OperationIDs) > 0 {
				partitionUsesIDs = true
				for _, id := range lane.Source.OperationIDs {
					if prior, exists := partitionIDs[id]; exists {
						return fmt.Errorf("source operation %q is claimed by both %q and %q", id, prior, lane.Name)
					}
					partitionIDs[id] = lane.Name
				}
			}
			if len(lane.Source.Methods) > 0 {
				partitionUsesMethods = true
			}
			for _, method := range lane.Source.Methods {
				method = strings.ToUpper(method)
				if prior, exists := partitionMethods[method]; exists {
					return fmt.Errorf("source method %q is claimed by both %q and %q", method, prior, lane.Name)
				}
				partitionMethods[method] = lane.Name
			}
		}
		if !lane.Source.Partition {
			for _, id := range lane.Source.OperationIDs {
				overlayIDs[id] = true
			}
		}
	}
	if partitionUsesMethods && partitionUsesIDs {
		return fmt.Errorf("source partitions cannot mix method and operation-ID selectors")
	}
	seenIDs := make(map[string]bool, len(operations))
	for _, operation := range operations {
		if !enabledContractSourceOperationIdentifier(operation.ID) || seenIDs[operation.ID] {
			return fmt.Errorf("source operation identity %q is invalid or duplicated", operation.ID)
		}
		seenIDs[operation.ID] = true
		lane, found := partitionIDs[operation.ID]
		if !found {
			lane, found = partitionMethods[strings.ToUpper(strings.TrimSpace(operation.Method))]
		}
		if !found {
			return fmt.Errorf("source operation %q method %q has no partition lane", operation.ID, operation.Method)
		}
		partitionCounts[lane]++
	}
	for _, lane := range c.Lanes {
		if lane.Source.Partition && partitionCounts[lane.Name] != lane.Source.Expected {
			return fmt.Errorf("partition lane %q reconciled %d source operations, want %d", lane.Name, partitionCounts[lane.Name], lane.Source.Expected)
		}
	}
	for id := range overlayIDs {
		if !seenIDs[id] {
			return fmt.Errorf("overlay references unknown source operation %q", id)
		}
	}
	for id := range partitionIDs {
		if !seenIDs[id] {
			return fmt.Errorf("partition references unknown source operation %q", id)
		}
	}
	return nil
}

// ReconcileSupplementalSourceOperations proves that an immutable supporting
// source document neither disappears behind the primary dialect nor leaks into
// a generic lane. It accepts only exact operation-ID selectors: a supplemental
// document cannot silently widen a primary method partition.
func (c EnabledConnectorContract) ReconcileSupplementalSourceOperations(sourceLock string, operations []EnabledContractSourceOperation) error {
	if err := c.Validate(); err != nil {
		return err
	}
	wantCount := 0
	knownLock := false
	for _, candidate := range c.SupplementalSourceLocks {
		if candidate.Path == sourceLock {
			wantCount = candidate.Operations
			knownLock = true
			break
		}
	}
	if !knownLock {
		return fmt.Errorf("supplemental source lock %q is not declared", sourceLock)
	}
	claimed := map[string]string{}
	for _, lane := range c.Lanes {
		if lane.Source.SourceLock != sourceLock {
			continue
		}
		for _, id := range lane.Source.OperationIDs {
			if previous, exists := claimed[id]; exists {
				return fmt.Errorf("supplemental source operation %q is claimed by both %q and %q", id, previous, lane.Name)
			}
			claimed[id] = lane.Name
		}
	}
	if len(operations) != wantCount {
		return fmt.Errorf("supplemental source lock %q reconciled %d operations, want %d", sourceLock, len(operations), wantCount)
	}
	seen := map[string]bool{}
	for _, operation := range operations {
		if !enabledContractSourceOperationIdentifier(operation.ID) || seen[operation.ID] {
			return fmt.Errorf("supplemental source operation identity %q is invalid or duplicated", operation.ID)
		}
		seen[operation.ID] = true
		if _, exists := claimed[operation.ID]; !exists {
			return fmt.Errorf("supplemental source operation %q has no lane claim", operation.ID)
		}
	}
	for id := range claimed {
		if !seen[id] {
			return fmt.Errorf("supplemental source lock %q claims unknown operation %q", sourceLock, id)
		}
	}
	return nil
}

func (c *EnabledConnectorContract) Clone() *EnabledConnectorContract {
	if c == nil {
		return nil
	}
	copy := *c
	copy.SupplementalSourceLocks = append([]EnabledContractSupplementalSourceLock(nil), c.SupplementalSourceLocks...)
	copy.Lanes = make([]EnabledConnectorLane, len(c.Lanes))
	for i, lane := range c.Lanes {
		copy.Lanes[i] = lane
		copy.Lanes[i].Citations = append([]EnabledContractCitation(nil), lane.Citations...)
		copy.Lanes[i].Artifacts = append([]string(nil), lane.Artifacts...)
		copy.Lanes[i].Source.Methods = append([]string(nil), lane.Source.Methods...)
		copy.Lanes[i].Source.OperationIDs = append([]string(nil), lane.Source.OperationIDs...)
		if lane.Transport != nil {
			transport := *lane.Transport
			transport.Modes = append([]string(nil), lane.Transport.Modes...)
			transport.Streams = append([]EnabledContractTransportStreamEvidence(nil), lane.Transport.Streams...)
			if lane.Transport.Destination != nil {
				destination := *lane.Transport.Destination
				transport.Destination = &destination
			}
			copy.Lanes[i].Transport = &transport
		}
		copy.Lanes[i].Warehouse = append([]EnabledContractWarehouseFlow(nil), lane.Warehouse...)
	}
	return &copy
}

func safeEnabledContractArtifact(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && value == path.Clean(value) && !strings.HasPrefix(value, "/") && value != "." && !strings.HasPrefix(value, "../")
}

func enabledContractIdentifier(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 256 || strings.ContainsAny(value, "/\\\x00\r\n\t ") {
		return false
	}
	return true
}

// enabledContractSourceOperationIdentifier validates retained provider source
// identities. Unlike connector names and artifact paths, provider operation
// IDs can legitimately contain a slash (for example GitHub's generated REST
// operation IDs); they are evidence keys only and never filesystem targets.
func enabledContractSourceOperationIdentifier(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && len(value) <= 256 && !strings.ContainsAny(value, "\\\x00\r\n\t ")
}

// LaneNames returns the normalized closed vocabulary for inspection callers.
func (c EnabledConnectorContract) LaneNames() []string {
	result := make([]string, 0, len(c.Lanes))
	for _, lane := range c.Lanes {
		result = append(result, lane.Name)
	}
	sort.Strings(result)
	return result
}

func (c EnabledConnectorContract) HasLane(name string) bool {
	return slices.Contains(c.LaneNames(), name)
}
