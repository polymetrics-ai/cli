package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

const (
	flowMatrixPath           = "internal/connectors/certifications/flow-matrix.json"
	certificationStatusPath  = "internal/connectors/certifications/status.json"
	workflowAnnotationPrefix = "pmcert:workflow"
)

type flowKind struct {
	ID              string `json:"id"`
	SourceRole      string `json:"source_role"`
	DestinationRole string `json:"destination_role"`
}

// workflowKind is discovered from the live pm command handlers, not copied
// into a hand-maintained certification list.
type workflowKind struct {
	ID              string `json:"id"`
	DiscoverySource string `json:"discovery_source"`
}

type workflowCertificationCell struct {
	WorkflowKind    string               `json:"workflow_kind"`
	Applicable      bool                 `json:"applicable"`
	Declared        bool                 `json:"declared"`
	Implemented     bool                 `json:"implemented"`
	FixtureTested   bool                 `json:"fixture_tested"`
	LiveTested      bool                 `json:"live_tested"`
	FixtureEvidence []string             `json:"fixture_evidence"`
	LiveEvidence    []evidencePointer    `json:"live_evidence"`
	NotApplicable   *notApplicableReason `json:"not_applicable,omitempty"`
}

type connectorWorkflowSet struct {
	Connector string                      `json:"connector"`
	Complete  bool                        `json:"complete"`
	Cells     []workflowCertificationCell `json:"cells"`
}

type workflowKindBaseline struct {
	WorkflowKind  string `json:"workflow_kind"`
	Connectors    int    `json:"connectors"`
	Applicable    int    `json:"applicable"`
	Declared      int    `json:"declared"`
	Implemented   int    `json:"implemented"`
	FixtureTested int    `json:"fixture_tested"`
	LiveTested    int    `json:"live_tested"`
	Complete      int    `json:"complete"`
}

// syncPrimitive is derived from the connector registry's concrete integration
// types and the public Read/Write capability vocabulary. It deliberately
// distinguishes database CDC from a generic connector-level checkbox.
type syncPrimitive struct {
	ID                 string `json:"id"`
	IntegrationType    string `json:"integration_type"`
	Capability         string `json:"capability"`
	WarehouseDirection string `json:"warehouse_direction"`
	DiscoverySource    string `json:"discovery_source"`
}

type syncModeKind struct {
	ID              string `json:"id"`
	DiscoverySource string `json:"discovery_source"`
}

type syncModeCertificationCell struct {
	SyncMode        string               `json:"sync_mode"`
	Primitive       string               `json:"primitive"`
	Applicable      bool                 `json:"applicable"`
	Declared        bool                 `json:"declared"`
	Implemented     bool                 `json:"implemented"`
	FixtureTested   bool                 `json:"fixture_tested"`
	LiveTested      bool                 `json:"live_tested"`
	FixtureEvidence []string             `json:"fixture_evidence"`
	LiveEvidence    []evidencePointer    `json:"live_evidence"`
	NotApplicable   *notApplicableReason `json:"not_applicable,omitempty"`
}

type connectorSyncModeSet struct {
	Connector string                      `json:"connector"`
	Complete  bool                        `json:"complete"`
	Cells     []syncModeCertificationCell `json:"cells"`
}

type syncModePrimitiveBaseline struct {
	SyncMode      string `json:"sync_mode"`
	Primitive     string `json:"primitive"`
	Connectors    int    `json:"connectors"`
	Applicable    int    `json:"applicable"`
	Declared      int    `json:"declared"`
	Implemented   int    `json:"implemented"`
	FixtureTested int    `json:"fixture_tested"`
	LiveTested    int    `json:"live_tested"`
	Complete      int    `json:"complete"`
}

// connectorFlowRole is an explicit declaration about one endpoint position.
// A role can be applicable but unimplemented: that is how the matrix records
// the currently unavailable API durability and database write paths rather
// than quietly treating those flows as absent.
type connectorFlowRole struct {
	Role          string               `json:"role"`
	Applicable    bool                 `json:"applicable"`
	Declared      bool                 `json:"declared"`
	Implemented   bool                 `json:"implemented"`
	NotApplicable *notApplicableReason `json:"not_applicable,omitempty"`
}

type connectorFlowRoles struct {
	Connector string              `json:"connector"`
	Roles     []connectorFlowRole `json:"roles"`
}

type flowCertificationCell struct {
	Applicable      bool                 `json:"applicable"`
	Declared        bool                 `json:"declared"`
	Implemented     bool                 `json:"implemented"`
	FixtureTested   bool                 `json:"fixture_tested"`
	LiveTested      bool                 `json:"live_tested"`
	FixtureEvidence []string             `json:"fixture_evidence"`
	LiveEvidence    []evidencePointer    `json:"live_evidence"`
	NotApplicable   *notApplicableReason `json:"not_applicable,omitempty"`
}

// flowPairSet is an in-memory compact, non-overlapping set of exact pair cells.
// The resolver never treats a set as a connector-level status: its source and
// destination membership identify the concrete unit of certification.
type flowPairSet struct {
	FlowKind              string                `json:"flow_kind"`
	Mediator              string                `json:"mediator"`
	SourceConnectors      []string              `json:"source_connectors"`
	DestinationConnectors []string              `json:"destination_connectors"`
	Cell                  flowCertificationCell `json:"cell"`
}

// flowPairOverride is an exact pair record. It can only add proof to one
// source/destination pair after a real round trip; it cannot promote a whole
// product of endpoints by sharing one result.
type flowPairOverride struct {
	FlowKind        string                `json:"flow_kind"`
	Source          string                `json:"source"`
	Destination     string                `json:"destination"`
	Mediator        string                `json:"mediator"`
	DestinationRole connectorFlowRole     `json:"destination_role"`
	Cell            flowCertificationCell `json:"cell"`
}

type flowKindBaseline struct {
	FlowKind      string `json:"flow_kind"`
	Pairs         int    `json:"pairs"`
	Applicable    int    `json:"applicable"`
	Declared      int    `json:"declared"`
	Implemented   int    `json:"implemented"`
	FixtureTested int    `json:"fixture_tested"`
	LiveTested    int    `json:"live_tested"`
	Complete      int    `json:"complete"`
}

type flowBaseline struct {
	Connectors int                         `json:"connectors"`
	Certified  int                         `json:"certified"`
	Workflows  []workflowKindBaseline      `json:"workflows"`
	SyncModes  []syncModePrimitiveBaseline `json:"sync_modes"`
	PerKind    []flowKindBaseline          `json:"per_kind"`
}

// connectorCertificationStatus is deliberately small and binary. It is the
// user-facing projection derived from proof-bearing shards through an in-memory
// aggregate, not a separate hand-maintained taxonomy.
type connectorCertificationStatus struct {
	Connector string `json:"connector"`
	Certified bool   `json:"certified"`
	Label     string `json:"label"`
	Warning   string `json:"warning,omitempty"`
}

// certificationStatusArtifact is the small generated projection embedded by
// pm for point-of-use status. Its source remains the proof-bearing shards
// reconstructed in memory; certification-matrix --check validates this artifact
// before comparing it to the freshly generated projection.
type certificationStatusArtifact struct {
	SchemaVersion      int                            `json:"schema_version"`
	GeneratedCommand   string                         `json:"generated_command"`
	CertificationScope []string                       `json:"certification_scope"`
	Connectors         []connectorCertificationStatus `json:"connectors"`
}

type flowMatrix struct {
	SchemaVersion     int                            `json:"schema_version"`
	GeneratedCommand  string                         `json:"generated_command"`
	Mediator          string                         `json:"mediator"`
	FlowKinds         []flowKind                     `json:"flow_kinds"`
	WorkflowKinds     []workflowKind                 `json:"workflow_kinds"`
	Workflows         []connectorWorkflowSet         `json:"workflows"`
	SyncModeKinds     []syncModeKind                 `json:"sync_mode_kinds"`
	SyncPrimitives    []syncPrimitive                `json:"sync_primitives"`
	SyncModeCells     []connectorSyncModeSet         `json:"sync_mode_cells"`
	ConnectorRoles    []connectorFlowRoles           `json:"connector_roles"`
	PairSets          []flowPairSet                  `json:"pair_sets"`
	PairOverrides     []flowPairOverride             `json:"pair_overrides"`
	ConnectorStatuses []connectorCertificationStatus `json:"connector_statuses"`
	Baseline          flowBaseline                   `json:"baseline"`
}

type resolvedFlowPair struct {
	FlowKind    string
	Source      string
	Destination string
	Mediator    string
	Cell        flowCertificationCell
}

func discoveredFlowKinds() []flowKind {
	return []flowKind{
		{ID: "api_to_api", SourceRole: "api_source", DestinationRole: "api_destination"},
		{ID: "api_to_database", SourceRole: "api_source", DestinationRole: "database_destination"},
		{ID: "database_to_api", SourceRole: "database_source", DestinationRole: "api_destination"},
		{ID: "database_to_database", SourceRole: "database_source", DestinationRole: "database_destination"},
	}
}

func discoverWorkflowKinds(repoRoot string) ([]workflowKind, error) {
	dir := filepath.Join(repoRoot, "internal", "cli")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read workflow command source: %w", err)
	}
	found := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse workflow command source %q: %w", path, err)
		}
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			for _, id := range workflowKindsFromDoc(function.Doc) {
				if !isSafeProofIdentifier(id) {
					return nil, fmt.Errorf("workflow annotation %q is invalid", id)
				}
				source := sourceSymbol(repoRoot, path, functionSymbol(function))
				if existing, exists := found[id]; exists {
					return nil, fmt.Errorf("workflow annotation %q has multiple handlers (%s, %s)", id, existing, source)
				}
				found[id] = source
			}
		}
	}
	if len(found) == 0 {
		return nil, errors.New("no pmcert:workflow annotations found")
	}
	kinds := make([]workflowKind, 0, len(found))
	for id, source := range found {
		kinds = append(kinds, workflowKind{ID: id, DiscoverySource: source})
	}
	sort.Slice(kinds, func(i, j int) bool { return kinds[i].ID < kinds[j].ID })
	return kinds, nil
}

func discoverSyncModes(repoRoot string) ([]syncModeKind, error) {
	source, err := syncContractAllModesSource(repoRoot)
	if err != nil {
		return nil, err
	}
	modes := synccontract.AllModes()
	out := make([]syncModeKind, 0, len(modes))
	for _, mode := range modes {
		out = append(out, syncModeKind{
			ID:              string(mode),
			DiscoverySource: source,
		})
	}
	return out, nil
}

func syncContractAllModesSource(repoRoot string) (string, error) {
	dir := filepath.Join(repoRoot, "internal", "synccontract")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read synccontract source directory: %w", err)
	}
	found := ""
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return "", fmt.Errorf("parse synccontract source %q: %w", path, err)
		}
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Recv != nil || function.Name.Name != "AllModes" {
				continue
			}
			source := sourceSymbol(repoRoot, path, functionSymbol(function))
			if found != "" {
				return "", fmt.Errorf("synccontract source declares AllModes multiple times (%s, %s)", found, source)
			}
			found = source
		}
	}
	if found == "" {
		return "", errors.New("synccontract source declares no AllModes function")
	}
	return found, nil
}

func discoverSyncPrimitives(repoRoot string) ([]syncPrimitive, error) {
	metadataSource, err := connectorMetadataSource(repoRoot)
	if err != nil {
		return nil, err
	}
	primitives := make([]syncPrimitive, 0, 4)
	for _, integrationType := range []string{"api", "database"} {
		for _, capability := range []string{"read", "write"} {
			direction := "into_warehouse"
			if capability == "write" {
				direction = "from_warehouse"
			}
			primitives = append(primitives, syncPrimitive{
				ID:                 integrationType + "_" + capability + "_" + direction,
				IntegrationType:    integrationType,
				Capability:         capability,
				WarehouseDirection: direction,
				DiscoverySource:    metadataSource,
			})
		}
	}
	return primitives, nil
}

func connectorMetadataSource(repoRoot string) (string, error) {
	dir := filepath.Join(repoRoot, "internal", "connectors")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read connector source directory: %w", err)
	}
	found := ""
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return "", fmt.Errorf("parse connector source %q: %w", path, err)
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != "Connector" {
					continue
				}
				iface, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}
				for _, method := range iface.Methods.List {
					for _, name := range method.Names {
						if name.Name != "Metadata" {
							continue
						}
						source := sourceSymbol(repoRoot, path, typeSpec.Name.Name+"."+name.Name)
						if found != "" {
							return "", fmt.Errorf("connector source declares Connector.Metadata multiple times (%s, %s)", found, source)
						}
						found = source
					}
				}
			}
		}
	}
	if found == "" {
		return "", errors.New("connector source declares no Connector.Metadata method")
	}
	return found, nil
}

func workflowKindsFromDoc(doc *ast.CommentGroup) []string {
	if doc == nil {
		return nil
	}
	var kinds []string
	for _, line := range strings.Split(doc.Text(), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, workflowAnnotationPrefix) {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(line, workflowAnnotationPrefix))
		for _, kind := range strings.Split(raw, ",") {
			if kind = strings.TrimSpace(kind); kind != "" {
				kinds = append(kinds, kind)
			}
		}
	}
	return kinds
}

func buildConnectorWorkflows(sources []matrixConnectorSource, capabilities capabilityMatrix, kinds []workflowKind, evidence []acceptedEvidence) ([]connectorWorkflowSet, error) {
	out := make([]connectorWorkflowSet, 0, len(sources))
	for _, source := range sources {
		read, hasRead := capabilityCellFor(capabilities, source.name, "capability:read")
		write, hasWrite := capabilityCellFor(capabilities, source.name, "capability:write")
		if !hasRead || !hasWrite {
			return nil, fmt.Errorf("connector %q is missing read or write capability rows", source.name)
		}
		cells := make([]workflowCertificationCell, 0, len(kinds))
		for _, kind := range kinds {
			cell, err := workflowCellFor(source.name, kind.ID, read, write, evidence)
			if err != nil {
				return nil, err
			}
			if err := validateWorkflowCertificationCell(cell); err != nil {
				return nil, fmt.Errorf("connector %q workflow %q: %w", source.name, kind.ID, err)
			}
			cells = append(cells, cell)
		}
		out = append(out, connectorWorkflowSet{Connector: source.name, Complete: workflowCellsComplete(cells), Cells: cells})
	}
	return out, nil
}

func workflowCellFor(connector, workflow string, read, write certificationCell, evidence []acceptedEvidence) (workflowCertificationCell, error) {
	fromCapability := func(kind string, capability certificationCell) workflowCertificationCell {
		if !capability.Applicable {
			return workflowCertificationCell{
				WorkflowKind:    kind,
				FixtureEvidence: []string{},
				LiveEvidence:    []evidencePointer{},
				NotApplicable: &notApplicableReason{
					Code:   "connector_capability_not_applicable",
					Reason: fmt.Sprintf("connector %q has no applicable %s capability for workflow %q", connector, capability.FunctionKind, kind),
				},
			}
		}
		live := matchingWorkflowEvidence(evidence, connector, kind)
		return workflowCertificationCell{
			WorkflowKind:    kind,
			Applicable:      true,
			Declared:        capability.Declared,
			Implemented:     capability.Implemented,
			FixtureTested:   capability.FixtureTested,
			LiveTested:      len(live) > 0,
			FixtureEvidence: append([]string(nil), capability.FixtureEvidence...),
			LiveEvidence:    live,
		}
	}

	switch workflow {
	case "etl":
		return fromCapability(workflow, read), nil
	case "reverse_etl":
		return fromCapability(workflow, write), nil
	case "flow_authoring", "schedule":
		if !read.Applicable && !write.Applicable {
			return workflowCertificationCell{
				WorkflowKind:    workflow,
				FixtureEvidence: []string{},
				LiveEvidence:    []evidencePointer{},
				NotApplicable: &notApplicableReason{
					Code:   "connector_has_no_flow_endpoint",
					Reason: fmt.Sprintf("connector %q exposes neither an applicable read nor write endpoint for %s", connector, workflow),
				},
			}, nil
		}
		live := matchingWorkflowEvidence(evidence, connector, workflow)
		return workflowCertificationCell{
			WorkflowKind:    workflow,
			Applicable:      true,
			Declared:        (read.Applicable && read.Declared) || (write.Applicable && write.Declared),
			Implemented:     (read.Applicable && read.Implemented) || (write.Applicable && write.Implemented),
			FixtureTested:   false,
			LiveTested:      len(live) > 0,
			FixtureEvidence: []string{},
			LiveEvidence:    live,
		}, nil
	default:
		return workflowCertificationCell{}, fmt.Errorf("workflow annotation %q has no certification semantics", workflow)
	}
}

func matchingWorkflowEvidence(evidence []acceptedEvidence, connector, workflow string) []evidencePointer {
	matched := make([]evidencePointer, 0)
	for _, item := range evidence {
		if item.Scope != evidenceScopeWorkflow || item.Status != evidenceStatusPassed || item.Connector != connector || item.WorkflowKind != workflow {
			continue
		}
		matched = append(matched, evidencePointer{
			Record:          item.recordPath,
			Provider:        item.Provider,
			ExecutedAt:      item.ExecutedAt,
			RunID:           item.RunID,
			CredentialScope: item.CredentialScope,
			CredentialNote:  item.CredentialNote,
			Proof:           item.Proof,
		})
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].Record < matched[j].Record })
	return matched
}

func validateWorkflowCertificationCell(cell workflowCertificationCell) error {
	if !isSafeProofIdentifier(cell.WorkflowKind) {
		return errors.New("workflow_kind is invalid")
	}
	if !cell.Applicable {
		if cell.NotApplicable == nil {
			return errors.New("not_applicable reason is required when applicable=false")
		}
		if cell.Declared || cell.Implemented || cell.FixtureTested || cell.LiveTested || len(cell.FixtureEvidence) != 0 || len(cell.LiveEvidence) != 0 {
			return errors.New("not_applicable workflow cell cannot carry affirmative evidence")
		}
		return validateNotApplicableReason(*cell.NotApplicable)
	}
	if cell.NotApplicable != nil {
		return errors.New("applicable workflow cell cannot carry not_applicable reason")
	}
	if cell.FixtureTested && len(cell.FixtureEvidence) == 0 {
		return errors.New("fixture_tested workflow cell requires fixture_evidence")
	}
	if cell.LiveTested && len(cell.LiveEvidence) == 0 {
		return errors.New("live_tested workflow cell requires live_evidence")
	}
	if !cell.LiveTested && len(cell.LiveEvidence) != 0 {
		return errors.New("live_evidence requires live_tested=true")
	}
	for _, item := range cell.LiveEvidence {
		if err := validateEvidencePointer(item); err != nil {
			return fmt.Errorf("live_evidence: %w", err)
		}
	}
	return nil
}

func workflowCellComplete(cell workflowCertificationCell) bool {
	return cell.Applicable && cell.Declared && cell.Implemented && cell.FixtureTested && cell.LiveTested && len(cell.LiveEvidence) > 0
}

func workflowCellsComplete(cells []workflowCertificationCell) bool {
	applicable := 0
	for _, cell := range cells {
		if !cell.Applicable {
			continue
		}
		applicable++
		if !workflowCellComplete(cell) {
			return false
		}
	}
	return applicable > 0
}

func buildConnectorSyncModeCells(sources []matrixConnectorSource, capabilities capabilityMatrix, modes []syncModeKind, primitives []syncPrimitive, evidence []acceptedEvidence) ([]connectorSyncModeSet, error) {
	out := make([]connectorSyncModeSet, 0, len(sources))
	for _, source := range sources {
		read, hasRead := capabilityCellFor(capabilities, source.name, "capability:read")
		write, hasWrite := capabilityCellFor(capabilities, source.name, "capability:write")
		cdc, hasCDC := capabilityCellFor(capabilities, source.name, "capability:cdc")
		if !hasRead || !hasWrite || !hasCDC {
			return nil, fmt.Errorf("connector %q is missing read, write, or CDC capability rows", source.name)
		}
		durable := false
		if source.connector != nil {
			_, durable = source.connector.(synccontract.DurableETLDestination)
		}
		cells := make([]syncModeCertificationCell, 0, len(modes)*len(primitives))
		for _, mode := range modes {
			for _, primitive := range primitives {
				cell := syncModeCellFor(source, mode.ID, primitive, read, write, cdc, durable, evidence)
				if err := validateSyncModeCertificationCell(cell); err != nil {
					return nil, fmt.Errorf("connector %q mode %q primitive %q: %w", source.name, mode.ID, primitive.ID, err)
				}
				cells = append(cells, cell)
			}
		}
		out = append(out, connectorSyncModeSet{Connector: source.name, Complete: syncModeCellsComplete(cells), Cells: cells})
	}
	return out, nil
}

func syncModeCellFor(source matrixConnectorSource, mode string, primitive syncPrimitive, read, write, cdc certificationCell, durable bool, evidence []acceptedEvidence) syncModeCertificationCell {
	base := syncModeCertificationCell{
		SyncMode:        mode,
		Primitive:       primitive.ID,
		FixtureEvidence: []string{},
		LiveEvidence:    []evidencePointer{},
	}
	if source.integrationType != primitive.IntegrationType {
		base.NotApplicable = &notApplicableReason{
			Code:   "primitive_requires_" + primitive.IntegrationType + "_connector",
			Reason: fmt.Sprintf("connector %q integration_type %q cannot execute primitive %q", source.name, source.integrationType, primitive.ID),
		}
		return base
	}
	if mode == string(synccontract.ModeChangeCapture) && primitive.ID != "database_read_into_warehouse" {
		base.NotApplicable = &notApplicableReason{
			Code:   "change_capture_requires_database_read",
			Reason: fmt.Sprintf("sync mode %q applies only to the database_read_into_warehouse primitive", mode),
		}
		return base
	}

	capability := read
	if primitive.Capability == "write" {
		capability = write
	}
	if mode == string(synccontract.ModeChangeCapture) {
		capability = cdc
	}
	live := matchingSyncModeEvidence(evidence, source.name, mode, primitive.ID)
	base.Applicable = true
	// A connector-level read/write bit says that a connector has a broad
	// operation path. It does not admit any particular warehouse sync mode.
	// Certification mode cells must follow the definition-owned source or
	// destination transport role that the real preflight will resolve.
	admitted := syncModeTransportAdmits(source, primitive, synccontract.Mode(mode))
	declaredPair := declaredNativeDatabaseTransportPair(source, synccontract.Mode(mode))
	if declaredPair {
		// A declared native-database pair does not use the connector's direct
		// Write method. Mark implementation only after an accepted exact-mode
		// live proof: the matrix must not infer an executable destination from a
		// descriptor or import a connector-native factory into shared tooling.
		base.Declared = admitted
		base.Implemented = admitted && len(live) > 0
	} else {
		base.Declared = capability.Declared && admitted
		base.Implemented = capability.Implemented && admitted
	}
	if primitive.Capability == "write" && !declaredPair {
		// A connector Write method is not a checkpointed ETL destination until
		// it can produce DurableETLDestination acknowledgement. This is why API
		// destinations and the database write stubs remain red today.
		base.Implemented = base.Implemented && durable
	}
	// Capability fixtures demonstrate a request path, not an individual sync
	// mode. The mode cell remains false until that exact mode has a recorded
	// exercise, preventing a generic read fixture from certifying upsert,
	// dedupe history, or CDC by implication.
	base.FixtureTested = false
	base.LiveTested = len(live) > 0
	base.LiveEvidence = live
	return base
}

// declaredNativeDatabaseTransportPair recognizes only a fully declared source
// and destination native-database contract. It deliberately makes no claim
// that such a pair is executable; implementation is admitted above only by a
// matching exact-mode live evidence record. This lets certification preserve a
// database connector's definition-owned contract without a connector-name
// branch or a shared import of native implementation code.
func declaredNativeDatabaseTransportPair(source matrixConnectorSource, mode synccontract.Mode) bool {
	if source.integrationType != "database" {
		return false
	}
	descriptor := syncTransportDescriptorFor(source)
	if descriptor == nil || descriptor.Source == nil || descriptor.Destination == nil ||
		descriptor.Source.Executor.Family != connectors.TransportExecutorFamilyNativeDatabase ||
		descriptor.Destination.Executor.Family != connectors.TransportExecutorFamilyNativeDatabase ||
		!syncTransportContainsMode(descriptor.Source.Modes, mode) ||
		!syncTransportContainsMode(descriptor.Destination.Modes, mode) {
		return false
	}
	return true
}

func syncModeTransportAdmits(source matrixConnectorSource, primitive syncPrimitive, mode synccontract.Mode) bool {
	descriptor := syncTransportDescriptorFor(source)
	if descriptor == nil {
		return false
	}
	if primitive.WarehouseDirection == "into_warehouse" {
		return descriptor.Source != nil && syncTransportContainsMode(descriptor.Source.Modes, mode)
	}
	if primitive.WarehouseDirection == "from_warehouse" {
		return descriptor.Destination != nil && syncTransportContainsMode(descriptor.Destination.Modes, mode)
	}
	return false
}

func syncTransportDescriptorFor(source matrixConnectorSource) *connectors.SyncTransportDescriptor {
	if source.bundle != nil && source.bundle.SyncTransport != nil {
		return source.bundle.SyncTransport
	}
	if source.connector == nil {
		return nil
	}
	descriptor, ok := connectors.SyncTransportDescriptorOf(source.connector)
	if !ok {
		return nil
	}
	return descriptor
}

func syncTransportContainsMode(modes []synccontract.Mode, want synccontract.Mode) bool {
	for _, mode := range modes {
		if mode == want {
			return true
		}
	}
	return false
}

func matchingSyncModeEvidence(evidence []acceptedEvidence, connector, mode, primitive string) []evidencePointer {
	matched := make([]evidencePointer, 0)
	for _, item := range evidence {
		if item.Scope != evidenceScopeSyncMode || item.Status != evidenceStatusPassed || item.Connector != connector || item.SyncMode != mode || item.Primitive != primitive {
			continue
		}
		matched = append(matched, evidencePointer{
			Record:          item.recordPath,
			Provider:        item.Provider,
			ExecutedAt:      item.ExecutedAt,
			RunID:           item.RunID,
			CredentialScope: item.CredentialScope,
			CredentialNote:  item.CredentialNote,
			Proof:           item.Proof,
		})
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].Record < matched[j].Record })
	return matched
}

func validateSyncModeCertificationCell(cell syncModeCertificationCell) error {
	if !isSafeProofIdentifier(cell.SyncMode) || !isSafeProofIdentifier(cell.Primitive) {
		return errors.New("sync_mode or primitive is invalid")
	}
	if !cell.Applicable {
		if cell.NotApplicable == nil {
			return errors.New("not_applicable reason is required when applicable=false")
		}
		if cell.Declared || cell.Implemented || cell.FixtureTested || cell.LiveTested || len(cell.FixtureEvidence) != 0 || len(cell.LiveEvidence) != 0 {
			return errors.New("not_applicable sync-mode cell cannot carry affirmative evidence")
		}
		return validateNotApplicableReason(*cell.NotApplicable)
	}
	if cell.NotApplicable != nil {
		return errors.New("applicable sync-mode cell cannot carry not_applicable reason")
	}
	if cell.FixtureTested && len(cell.FixtureEvidence) == 0 {
		return errors.New("fixture_tested sync-mode cell requires fixture_evidence")
	}
	if cell.LiveTested && len(cell.LiveEvidence) == 0 {
		return errors.New("live_tested sync-mode cell requires live_evidence")
	}
	if !cell.LiveTested && len(cell.LiveEvidence) != 0 {
		return errors.New("live_evidence requires live_tested=true")
	}
	for _, item := range cell.LiveEvidence {
		if err := validateEvidencePointer(item); err != nil {
			return fmt.Errorf("live_evidence: %w", err)
		}
	}
	return nil
}

func syncModeCellComplete(cell syncModeCertificationCell) bool {
	return cell.Applicable && cell.Declared && cell.Implemented && cell.FixtureTested && cell.LiveTested && len(cell.LiveEvidence) > 0
}

func syncModeCellsComplete(cells []syncModeCertificationCell) bool {
	applicable := 0
	for _, cell := range cells {
		if !cell.Applicable {
			continue
		}
		applicable++
		if !syncModeCellComplete(cell) {
			return false
		}
	}
	return applicable > 0
}

func buildFlowMatrixForConnectors(repoRoot string, capabilities capabilityMatrix, names []string) (flowMatrix, error) {
	bundles, err := loadSourceBundlesForConnectors(repoRoot, names)
	if err != nil {
		return flowMatrix{}, err
	}
	sources, err := matrixConnectorSourcesForNames(bundles, names)
	if err != nil {
		return flowMatrix{}, err
	}
	evidence, err := loadAcceptedEvidence(repoRoot, names)
	if err != nil {
		return flowMatrix{}, err
	}

	roles, rolesByConnector, err := buildConnectorFlowRoles(sources, capabilities)
	if err != nil {
		return flowMatrix{}, err
	}
	workflowKinds, err := discoverWorkflowKinds(repoRoot)
	if err != nil {
		return flowMatrix{}, err
	}
	workflows, err := buildConnectorWorkflows(sources, capabilities, workflowKinds, evidence)
	if err != nil {
		return flowMatrix{}, err
	}
	syncModes, err := discoverSyncModes(repoRoot)
	if err != nil {
		return flowMatrix{}, err
	}
	syncPrimitives, err := discoverSyncPrimitives(repoRoot)
	if err != nil {
		return flowMatrix{}, err
	}
	syncCells, err := buildConnectorSyncModeCells(sources, capabilities, syncModes, syncPrimitives, evidence)
	if err != nil {
		return flowMatrix{}, err
	}
	kinds := discoveredFlowKinds()
	pairSets, err := buildFlowPairSets(kinds, rolesByConnector)
	if err != nil {
		return flowMatrix{}, err
	}
	matrix := flowMatrix{
		SchemaVersion:     certificationSchemaVersion,
		GeneratedCommand:  "go run ./cmd/connectorgen certification-matrix",
		Mediator:          localWarehouseMediator,
		FlowKinds:         kinds,
		WorkflowKinds:     workflowKinds,
		Workflows:         workflows,
		SyncModeKinds:     syncModes,
		SyncPrimitives:    syncPrimitives,
		SyncModeCells:     syncCells,
		ConnectorRoles:    roles,
		PairSets:          pairSets,
		PairOverrides:     []flowPairOverride{},
		ConnectorStatuses: []connectorCertificationStatus{},
	}
	if err := applyFlowEvidence(&matrix, evidence); err != nil {
		return flowMatrix{}, err
	}
	matrix.ConnectorStatuses = deriveConnectorStatuses(capabilities, matrix)
	matrix.Baseline = deriveFlowBaseline(matrix)
	if err := validateFlowMatrix(matrix); err != nil {
		return flowMatrix{}, err
	}
	return matrix, nil
}

func buildCertificationStatusArtifact(matrix flowMatrix) certificationStatusArtifact {
	statuses := append([]connectorCertificationStatus(nil), matrix.ConnectorStatuses...)
	return certificationStatusArtifact{
		SchemaVersion:      certificationSchemaVersion,
		GeneratedCommand:   "go run ./cmd/connectorgen certification-matrix --all",
		CertificationScope: append([]string(nil), certificationConnectorAllowlist...),
		Connectors:         statuses,
	}
}

func buildConnectorFlowRoles(sources []matrixConnectorSource, capabilities capabilityMatrix) ([]connectorFlowRoles, map[string]map[string]connectorFlowRole, error) {
	roles := make([]connectorFlowRoles, 0, len(sources))
	byConnector := make(map[string]map[string]connectorFlowRole, len(sources))
	for _, source := range sources {
		read, foundRead := capabilityCellFor(capabilities, source.name, "capability:read")
		write, foundWrite := capabilityCellFor(capabilities, source.name, "capability:write")
		if !foundRead || !foundWrite {
			return nil, nil, fmt.Errorf("connector %q is missing read or write capability rows", source.name)
		}
		durable := false
		if source.connector != nil {
			_, durable = source.connector.(synccontract.DurableETLDestination)
		}

		entries := []connectorFlowRole{
			roleForIntegration(source.integrationType, "api_source", "api", read.Declared, read.Implemented),
			roleForIntegration(source.integrationType, "api_destination", "api", write.Declared, write.Implemented && durable),
			roleForIntegration(source.integrationType, "database_source", "database", read.Declared, read.Implemented),
			roleForIntegration(source.integrationType, "database_destination", "database", write.Declared, write.Implemented && durable),
		}
		for _, role := range entries {
			if err := validateConnectorFlowRole(role); err != nil {
				return nil, nil, fmt.Errorf("connector %q role %q: %w", source.name, role.Role, err)
			}
		}
		byRole := make(map[string]connectorFlowRole, len(entries))
		for _, entry := range entries {
			byRole[entry.Role] = entry
		}
		roles = append(roles, connectorFlowRoles{Connector: source.name, Roles: entries})
		byConnector[source.name] = byRole
	}
	return roles, byConnector, nil
}

func roleForIntegration(integrationType, role, expectedType string, declared, implemented bool) connectorFlowRole {
	if integrationType != expectedType {
		return connectorFlowRole{
			Role: role,
			NotApplicable: &notApplicableReason{
				Code:   "integration_type_not_" + expectedType,
				Reason: fmt.Sprintf("connector integration_type %q cannot occupy the %s role", integrationType, role),
			},
		}
	}
	return connectorFlowRole{
		Role:        role,
		Applicable:  true,
		Declared:    declared,
		Implemented: implemented,
	}
}

func validateConnectorFlowRole(role connectorFlowRole) error {
	if !isSafeProofIdentifier(role.Role) {
		return errors.New("role is invalid")
	}
	if !role.Applicable {
		if role.NotApplicable == nil {
			return errors.New("not_applicable reason is required when applicable=false")
		}
		if role.Declared || role.Implemented {
			return errors.New("not_applicable role cannot carry affirmative facts")
		}
		return validateNotApplicableReason(*role.NotApplicable)
	}
	if role.NotApplicable != nil {
		return errors.New("applicable role cannot carry not_applicable reason")
	}
	return nil
}

func buildFlowPairSets(kinds []flowKind, rolesByConnector map[string]map[string]connectorFlowRole) ([]flowPairSet, error) {
	names := make([]string, 0, len(rolesByConnector))
	for name := range rolesByConnector {
		names = append(names, name)
	}
	sort.Strings(names)

	pairSets := make([]flowPairSet, 0, len(kinds)*4)
	for _, kind := range kinds {
		sourceGroups := groupFlowRoleNames(names, rolesByConnector, kind.SourceRole)
		destinationGroups := groupFlowRoleNames(names, rolesByConnector, kind.DestinationRole)
		sourceKeys := sortedStringKeys(sourceGroups)
		destinationKeys := sortedStringKeys(destinationGroups)
		for _, sourceKey := range sourceKeys {
			for _, destinationKey := range destinationKeys {
				sourceNames := sourceGroups[sourceKey]
				destinationNames := destinationGroups[destinationKey]
				if len(sourceNames) == 0 || len(destinationNames) == 0 {
					continue
				}
				sourceRole := rolesByConnector[sourceNames[0]][kind.SourceRole]
				destinationRole := rolesByConnector[destinationNames[0]][kind.DestinationRole]
				cell := baseFlowCell(sourceRole, destinationRole)
				if err := validateFlowCertificationCell(cell); err != nil {
					return nil, fmt.Errorf("%s %s -> %s: %w", kind.ID, sourceKey, destinationKey, err)
				}
				pairSets = append(pairSets, flowPairSet{
					FlowKind:              kind.ID,
					Mediator:              localWarehouseMediator,
					SourceConnectors:      sourceNames,
					DestinationConnectors: destinationNames,
					Cell:                  cell,
				})
			}
		}
	}
	return pairSets, nil
}

func groupFlowRoleNames(names []string, rolesByConnector map[string]map[string]connectorFlowRole, roleName string) map[string][]string {
	groups := make(map[string][]string)
	for _, name := range names {
		role := rolesByConnector[name][roleName]
		key := flowRoleKey(role)
		groups[key] = append(groups[key], name)
	}
	return groups
}

func flowRoleKey(role connectorFlowRole) string {
	if role.NotApplicable != nil {
		return strings.Join([]string{role.Role, "false", role.NotApplicable.Code, role.NotApplicable.Reason}, "\x00")
	}
	return fmt.Sprintf("%s\x00true\x00%t\x00%t", role.Role, role.Declared, role.Implemented)
}

func sortedStringKeys(values map[string][]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func baseFlowCell(sourceRole, destinationRole connectorFlowRole) flowCertificationCell {
	if !sourceRole.Applicable && !destinationRole.Applicable {
		return flowCertificationCell{
			FixtureEvidence: []string{},
			LiveEvidence:    []evidencePointer{},
			NotApplicable: &notApplicableReason{
				Code:   "source_and_destination_roles_inapplicable",
				Reason: fmt.Sprintf("source %s and destination %s roles are not applicable", sourceRole.Role, destinationRole.Role),
			},
		}
	}
	if !sourceRole.Applicable {
		return flowCertificationCell{
			FixtureEvidence: []string{},
			LiveEvidence:    []evidencePointer{},
			NotApplicable: &notApplicableReason{
				Code:   "source_" + sourceRole.NotApplicable.Code,
				Reason: "source " + sourceRole.NotApplicable.Reason,
			},
		}
	}
	if !destinationRole.Applicable {
		return flowCertificationCell{
			FixtureEvidence: []string{},
			LiveEvidence:    []evidencePointer{},
			NotApplicable: &notApplicableReason{
				Code:   "destination_" + destinationRole.NotApplicable.Code,
				Reason: "destination " + destinationRole.NotApplicable.Reason,
			},
		}
	}
	return flowCertificationCell{
		Applicable:      true,
		Declared:        sourceRole.Declared && destinationRole.Declared,
		Implemented:     sourceRole.Implemented && destinationRole.Implemented,
		FixtureTested:   false,
		LiveTested:      false,
		FixtureEvidence: []string{},
		LiveEvidence:    []evidencePointer{},
	}
}

func validateFlowCertificationCell(cell flowCertificationCell) error {
	if !cell.Applicable {
		if cell.NotApplicable == nil {
			return errors.New("not_applicable reason is required when applicable=false")
		}
		if cell.Declared || cell.Implemented || cell.FixtureTested || cell.LiveTested || len(cell.FixtureEvidence) != 0 || len(cell.LiveEvidence) != 0 {
			return errors.New("not_applicable flow cell cannot carry affirmative evidence")
		}
		return validateNotApplicableReason(*cell.NotApplicable)
	}
	if cell.NotApplicable != nil {
		return errors.New("applicable flow cell cannot carry not_applicable reason")
	}
	if cell.FixtureTested && len(cell.FixtureEvidence) == 0 {
		return errors.New("fixture_tested flow cell requires fixture_evidence")
	}
	if cell.LiveTested && len(cell.LiveEvidence) == 0 {
		return errors.New("live_tested flow cell requires live_evidence")
	}
	if !cell.LiveTested && len(cell.LiveEvidence) != 0 {
		return errors.New("live_evidence requires live_tested=true")
	}
	for _, evidence := range cell.LiveEvidence {
		if err := validateEvidencePointer(evidence); err != nil {
			return fmt.Errorf("live_evidence: %w", err)
		}
		if evidence.Proof.Flow == nil {
			return errors.New("live_evidence requires an embedded flow round-trip proof")
		}
	}
	return nil
}

func flowCellComplete(cell flowCertificationCell) bool {
	return cell.Applicable && cell.Declared && cell.Implemented && cell.FixtureTested && cell.LiveTested && len(cell.LiveEvidence) > 0
}

func applyFlowEvidence(matrix *flowMatrix, evidence []acceptedEvidence) error {
	knownKinds := make(map[string]flowKind, len(matrix.FlowKinds))
	for _, kind := range matrix.FlowKinds {
		knownKinds[kind.ID] = kind
	}
	knownConnectors := make(map[string]bool, len(matrix.ConnectorRoles))
	for _, item := range matrix.ConnectorRoles {
		knownConnectors[item.Connector] = true
	}

	for _, item := range evidence {
		if item.Scope != evidenceScopeFlow || item.Status != evidenceStatusPassed {
			continue
		}
		kind, known := knownKinds[item.FlowKind]
		if !known {
			return fmt.Errorf("accepted flow evidence %q names unknown flow kind %q", item.recordPath, item.FlowKind)
		}
		if !knownConnectors[item.Source] || !knownConnectors[item.Destination] {
			return fmt.Errorf("accepted flow evidence %q names an unknown source or destination", item.recordPath)
		}
		resolved, ok := resolveFlowPair(*matrix, item.FlowKind, item.Source, item.Destination)
		if !ok {
			return fmt.Errorf("accepted flow evidence %q does not resolve to a pair cell", item.recordPath)
		}
		if !resolved.Cell.Applicable {
			return fmt.Errorf("accepted flow evidence %q targets non-applicable pair %s -> %s", item.recordPath, item.Source, item.Destination)
		}
		destinationRole, found := flowRoleForConnector(matrix.ConnectorRoles, item.Destination, kind.DestinationRole)
		if !found {
			return fmt.Errorf("accepted flow evidence %q has no destination role context", item.recordPath)
		}
		cell := resolved.Cell
		pointer := evidencePointer{
			Record:          item.recordPath,
			Provider:        item.Provider,
			ExecutedAt:      item.ExecutedAt,
			RunID:           item.RunID,
			CredentialScope: item.CredentialScope,
			CredentialNote:  item.CredentialNote,
			Proof:           item.Proof,
		}
		cell.LiveEvidence = append(cell.LiveEvidence, pointer)
		cell.LiveTested = len(cell.LiveEvidence) > 0
		if err := validateFlowCertificationCell(cell); err != nil {
			return fmt.Errorf("accepted flow evidence %q: %w", item.recordPath, err)
		}
		matrix.PairOverrides = upsertFlowPairOverride(matrix.PairOverrides, flowPairOverride{
			FlowKind:        item.FlowKind,
			Source:          item.Source,
			Destination:     item.Destination,
			Mediator:        localWarehouseMediator,
			DestinationRole: destinationRole,
			Cell:            cell,
		})
	}
	sort.Slice(matrix.PairOverrides, func(i, j int) bool {
		left := matrix.PairOverrides[i]
		right := matrix.PairOverrides[j]
		if left.FlowKind != right.FlowKind {
			return left.FlowKind < right.FlowKind
		}
		if left.Source != right.Source {
			return left.Source < right.Source
		}
		return left.Destination < right.Destination
	})
	return nil
}

func upsertFlowPairOverride(overrides []flowPairOverride, next flowPairOverride) []flowPairOverride {
	for index, current := range overrides {
		if current.FlowKind == next.FlowKind && current.Source == next.Source && current.Destination == next.Destination {
			overrides[index] = next
			return overrides
		}
	}
	return append(overrides, next)
}

func flowRoleForConnector(sets []connectorFlowRoles, connector, role string) (connectorFlowRole, bool) {
	for _, set := range sets {
		if set.Connector != connector {
			continue
		}
		for _, candidate := range set.Roles {
			if candidate.Role == role {
				return candidate, true
			}
		}
	}
	return connectorFlowRole{}, false
}

func connectorFlowRolesEqual(left, right connectorFlowRole) bool {
	if left.Role != right.Role || left.Applicable != right.Applicable || left.Declared != right.Declared || left.Implemented != right.Implemented {
		return false
	}
	if left.NotApplicable == nil || right.NotApplicable == nil {
		return left.NotApplicable == nil && right.NotApplicable == nil
	}
	return *left.NotApplicable == *right.NotApplicable
}

func flowCellMatchesRoleBase(cell, base flowCertificationCell) bool {
	return cell.Applicable == base.Applicable && cell.Declared == base.Declared && cell.Implemented == base.Implemented
}

func resolveFlowPair(matrix flowMatrix, flowKindID, source, destination string) (resolvedFlowPair, bool) {
	for _, override := range matrix.PairOverrides {
		if override.FlowKind == flowKindID && override.Source == source && override.Destination == destination {
			return resolvedFlowPair{FlowKind: flowKindID, Source: source, Destination: destination, Mediator: override.Mediator, Cell: override.Cell}, true
		}
	}
	for _, set := range matrix.PairSets {
		if set.FlowKind != flowKindID || !containsString(set.SourceConnectors, source) || !containsString(set.DestinationConnectors, destination) {
			continue
		}
		return resolvedFlowPair{FlowKind: flowKindID, Source: source, Destination: destination, Mediator: set.Mediator, Cell: set.Cell}, true
	}
	return resolvedFlowPair{}, false
}

func containsString(values []string, want string) bool {
	index := sort.SearchStrings(values, want)
	return index < len(values) && values[index] == want
}

func deriveFlowBaseline(matrix flowMatrix) flowBaseline {
	baseline := flowBaseline{
		Connectors: len(matrix.ConnectorRoles),
		Workflows:  deriveWorkflowBaseline(matrix.Workflows, matrix.WorkflowKinds),
		SyncModes:  deriveSyncModeBaseline(matrix.SyncModeCells, matrix.SyncModeKinds, matrix.SyncPrimitives),
		PerKind:    make([]flowKindBaseline, 0, len(matrix.FlowKinds)),
	}
	for _, kind := range matrix.FlowKinds {
		totals := flowKindBaseline{FlowKind: kind.ID}
		for _, set := range matrix.PairSets {
			if set.FlowKind != kind.ID {
				continue
			}
			count := len(set.SourceConnectors) * len(set.DestinationConnectors)
			addFlowCellTotals(&totals, set.Cell, count)
		}
		for _, override := range matrix.PairOverrides {
			if override.FlowKind != kind.ID {
				continue
			}
			base, ok := resolveFlowPair(flowMatrix{PairSets: matrix.PairSets}, kind.ID, override.Source, override.Destination)
			if !ok {
				continue
			}
			addFlowCellTotals(&totals, base.Cell, -1)
			addFlowCellTotals(&totals, override.Cell, 1)
		}
		baseline.PerKind = append(baseline.PerKind, totals)
	}
	for _, status := range matrix.ConnectorStatuses {
		if status.Certified {
			baseline.Certified++
		}
	}
	return baseline
}

func deriveSyncModeBaseline(sets []connectorSyncModeSet, modes []syncModeKind, primitives []syncPrimitive) []syncModePrimitiveBaseline {
	baseline := make([]syncModePrimitiveBaseline, 0, len(modes)*len(primitives))
	for _, mode := range modes {
		for _, primitive := range primitives {
			totals := syncModePrimitiveBaseline{SyncMode: mode.ID, Primitive: primitive.ID, Connectors: len(sets)}
			for _, set := range sets {
				for _, cell := range set.Cells {
					if cell.SyncMode != mode.ID || cell.Primitive != primitive.ID || !cell.Applicable {
						continue
					}
					totals.Applicable++
					if cell.Declared {
						totals.Declared++
					}
					if cell.Implemented {
						totals.Implemented++
					}
					if cell.FixtureTested {
						totals.FixtureTested++
					}
					if cell.LiveTested {
						totals.LiveTested++
					}
					if syncModeCellComplete(cell) {
						totals.Complete++
					}
				}
			}
			baseline = append(baseline, totals)
		}
	}
	return baseline
}

func deriveWorkflowBaseline(workflows []connectorWorkflowSet, kinds []workflowKind) []workflowKindBaseline {
	baseline := make([]workflowKindBaseline, 0, len(kinds))
	for _, kind := range kinds {
		totals := workflowKindBaseline{WorkflowKind: kind.ID, Connectors: len(workflows)}
		for _, connector := range workflows {
			for _, cell := range connector.Cells {
				if cell.WorkflowKind != kind.ID || !cell.Applicable {
					continue
				}
				totals.Applicable++
				if cell.Declared {
					totals.Declared++
				}
				if cell.Implemented {
					totals.Implemented++
				}
				if cell.FixtureTested {
					totals.FixtureTested++
				}
				if cell.LiveTested {
					totals.LiveTested++
				}
				if workflowCellComplete(cell) {
					totals.Complete++
				}
			}
		}
		baseline = append(baseline, totals)
	}
	return baseline
}

func addFlowCellTotals(totals *flowKindBaseline, cell flowCertificationCell, count int) {
	totals.Pairs += count
	if !cell.Applicable {
		return
	}
	totals.Applicable += count
	if cell.Declared {
		totals.Declared += count
	}
	if cell.Implemented {
		totals.Implemented += count
	}
	if cell.FixtureTested {
		totals.FixtureTested += count
	}
	if cell.LiveTested {
		totals.LiveTested += count
	}
	if flowCellComplete(cell) {
		totals.Complete += count
	}
}

func deriveConnectorStatuses(capabilities capabilityMatrix, flows flowMatrix) []connectorCertificationStatus {
	statuses := make([]connectorCertificationStatus, 0, len(capabilities.Connectors))
	for _, connector := range capabilities.Connectors {
		certified := connector.CapabilityComplete && connectorWorkflowComplete(connector.Name, flows) && connectorSyncModesComplete(connector.Name, flows) && connectorFlowComplete(connector.Name, flows)
		status := connectorCertificationStatus{Connector: connector.Name, Certified: certified}
		if certified {
			status.Label = "CERTIFIED"
		} else {
			status.Label = "COMMUNITY BUILD, UNCERTIFIED"
			status.Warning = "This connector is reachable but is a COMMUNITY BUILD, UNCERTIFIED."
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func connectorWorkflowComplete(connector string, matrix flowMatrix) bool {
	for _, set := range matrix.Workflows {
		if set.Connector == connector {
			return set.Complete
		}
	}
	return false
}

func connectorSyncModesComplete(connector string, matrix flowMatrix) bool {
	for _, set := range matrix.SyncModeCells {
		if set.Connector == connector {
			return set.Complete
		}
	}
	return false
}

func connectorFlowComplete(connector string, matrix flowMatrix) bool {
	foundApplicable := false
	for _, kind := range matrix.FlowKinds {
		for _, roles := range matrix.ConnectorRoles {
			for _, side := range []bool{true, false} {
				var source, destination string
				if side {
					source, destination = connector, roles.Connector
				} else {
					source, destination = roles.Connector, connector
				}
				pair, ok := resolveFlowPair(matrix, kind.ID, source, destination)
				if !ok || !pair.Cell.Applicable {
					continue
				}
				foundApplicable = true
				if !flowCellComplete(pair.Cell) {
					return false
				}
			}
		}
	}
	return foundApplicable
}

func validateFlowMatrix(matrix flowMatrix) error {
	if matrix.SchemaVersion != certificationSchemaVersion {
		return fmt.Errorf("schema_version %d is unsupported", matrix.SchemaVersion)
	}
	if matrix.Mediator != localWarehouseMediator {
		return fmt.Errorf("mediator %q is unsupported", matrix.Mediator)
	}
	knownKinds := make(map[string]flowKind, len(matrix.FlowKinds))
	for _, kind := range matrix.FlowKinds {
		if !isSafeProofIdentifier(kind.ID) || !isSafeProofIdentifier(kind.SourceRole) || !isSafeProofIdentifier(kind.DestinationRole) {
			return errors.New("flow kind is invalid")
		}
		if _, exists := knownKinds[kind.ID]; exists {
			return fmt.Errorf("flow kind %q is duplicated", kind.ID)
		}
		knownKinds[kind.ID] = kind
	}
	knownWorkflowKinds := make(map[string]bool, len(matrix.WorkflowKinds))
	for _, kind := range matrix.WorkflowKinds {
		if !isSafeProofIdentifier(kind.ID) || strings.TrimSpace(kind.DiscoverySource) == "" {
			return errors.New("workflow kind is invalid")
		}
		if knownWorkflowKinds[kind.ID] {
			return fmt.Errorf("workflow kind %q is duplicated", kind.ID)
		}
		knownWorkflowKinds[kind.ID] = true
	}
	knownSyncModes := make(map[string]bool, len(matrix.SyncModeKinds))
	for _, mode := range matrix.SyncModeKinds {
		if !isSafeProofIdentifier(mode.ID) || strings.TrimSpace(mode.DiscoverySource) == "" {
			return errors.New("sync mode is invalid")
		}
		if knownSyncModes[mode.ID] {
			return fmt.Errorf("sync mode %q is duplicated", mode.ID)
		}
		knownSyncModes[mode.ID] = true
	}
	knownPrimitives := make(map[string]syncPrimitive, len(matrix.SyncPrimitives))
	for _, primitive := range matrix.SyncPrimitives {
		if !isSafeProofIdentifier(primitive.ID) || (primitive.IntegrationType != "api" && primitive.IntegrationType != "database") || (primitive.Capability != "read" && primitive.Capability != "write") || (primitive.WarehouseDirection != "into_warehouse" && primitive.WarehouseDirection != "from_warehouse") || strings.TrimSpace(primitive.DiscoverySource) == "" {
			return errors.New("sync primitive is invalid")
		}
		if _, exists := knownPrimitives[primitive.ID]; exists {
			return fmt.Errorf("sync primitive %q is duplicated", primitive.ID)
		}
		knownPrimitives[primitive.ID] = primitive
	}
	if err := validateRequiredWarehousePrimitives(knownPrimitives); err != nil {
		return err
	}
	knownConnectors := make(map[string]bool, len(matrix.ConnectorRoles))
	for _, connector := range matrix.ConnectorRoles {
		if strings.TrimSpace(connector.Connector) == "" || knownConnectors[connector.Connector] {
			return fmt.Errorf("connector role declaration %q is invalid or duplicated", connector.Connector)
		}
		knownConnectors[connector.Connector] = true
		seenRoles := make(map[string]bool, len(connector.Roles))
		for _, role := range connector.Roles {
			if seenRoles[role.Role] {
				return fmt.Errorf("connector %q role %q is duplicated", connector.Connector, role.Role)
			}
			seenRoles[role.Role] = true
			if err := validateConnectorFlowRole(role); err != nil {
				return fmt.Errorf("connector %q role %q: %w", connector.Connector, role.Role, err)
			}
		}
		for _, kind := range matrix.FlowKinds {
			if !seenRoles[kind.SourceRole] || !seenRoles[kind.DestinationRole] {
				return fmt.Errorf("connector %q omits a required flow role", connector.Connector)
			}
		}
	}
	if err := validateConnectorWorkflowSets(matrix.Workflows, knownConnectors, knownWorkflowKinds); err != nil {
		return err
	}
	if err := validateConnectorSyncModeSets(matrix.SyncModeCells, knownConnectors, knownSyncModes, knownPrimitives); err != nil {
		return err
	}
	for _, set := range matrix.PairSets {
		if _, known := knownKinds[set.FlowKind]; !known || set.Mediator != localWarehouseMediator {
			return errors.New("pair set has an unknown flow kind or mediator")
		}
		if len(set.SourceConnectors) == 0 || len(set.DestinationConnectors) == 0 || !isSortedUnique(set.SourceConnectors) || !isSortedUnique(set.DestinationConnectors) {
			return errors.New("pair set connector membership is invalid")
		}
		for _, name := range append(append([]string{}, set.SourceConnectors...), set.DestinationConnectors...) {
			if !knownConnectors[name] {
				return fmt.Errorf("pair set names unknown connector %q", name)
			}
		}
		if err := validateFlowCertificationCell(set.Cell); err != nil {
			return fmt.Errorf("pair set %q: %w", set.FlowKind, err)
		}
	}
	if err := validateFlowPairCoverage(matrix, knownConnectors); err != nil {
		return err
	}
	seenOverrides := make(map[string]bool, len(matrix.PairOverrides))
	for _, override := range matrix.PairOverrides {
		key := strings.Join([]string{override.FlowKind, override.Source, override.Destination}, "\x00")
		if seenOverrides[key] {
			return fmt.Errorf("pair override %q is duplicated", key)
		}
		seenOverrides[key] = true
		kind, known := knownKinds[override.FlowKind]
		if override.Mediator != localWarehouseMediator || !known || !knownConnectors[override.Source] || !knownConnectors[override.Destination] {
			return errors.New("pair override identity is invalid")
		}
		if err := validateConnectorFlowRole(override.DestinationRole); err != nil || override.DestinationRole.Role != kind.DestinationRole {
			return errors.New("pair override destination role context is invalid")
		}
		destinationRole, found := flowRoleForConnector(matrix.ConnectorRoles, override.Destination, kind.DestinationRole)
		if !found || !connectorFlowRolesEqual(destinationRole, override.DestinationRole) {
			return errors.New("pair override destination role context does not match the destination connector")
		}
		base, ok := resolveFlowPair(flowMatrix{PairSets: matrix.PairSets}, override.FlowKind, override.Source, override.Destination)
		if !ok || !base.Cell.Applicable {
			return errors.New("pair override must target an applicable default pair")
		}
		if !flowCellMatchesRoleBase(override.Cell, base.Cell) {
			return errors.New("pair override facts do not match its source and destination roles")
		}
		if err := validateFlowCertificationCell(override.Cell); err != nil {
			return fmt.Errorf("pair override %q: %w", key, err)
		}
		if !override.Cell.Applicable {
			return errors.New("pair override cannot make an applicable pair non-applicable")
		}
	}
	if err := validateConnectorStatuses(matrix.ConnectorStatuses, knownConnectors); err != nil {
		return err
	}
	return nil
}

func validateRequiredWarehousePrimitives(primitives map[string]syncPrimitive) error {
	required := map[string]struct {
		integrationType string
		capability      string
		direction       string
	}{
		"api_read_into_warehouse":       {integrationType: "api", capability: "read", direction: "into_warehouse"},
		"api_write_from_warehouse":      {integrationType: "api", capability: "write", direction: "from_warehouse"},
		"database_read_into_warehouse":  {integrationType: "database", capability: "read", direction: "into_warehouse"},
		"database_write_from_warehouse": {integrationType: "database", capability: "write", direction: "from_warehouse"},
	}
	if len(primitives) != len(required) {
		return fmt.Errorf("sync primitive inventory has %d entries, want the four required warehouse-facing primitives", len(primitives))
	}
	for id, want := range required {
		got, ok := primitives[id]
		if !ok || got.IntegrationType != want.integrationType || got.Capability != want.capability || got.WarehouseDirection != want.direction {
			return fmt.Errorf("sync primitive %q is missing or has an invalid warehouse-facing mapping", id)
		}
	}
	return nil
}

func validateConnectorWorkflowSets(sets []connectorWorkflowSet, connectors map[string]bool, kinds map[string]bool) error {
	seen := make(map[string]bool, len(sets))
	for _, set := range sets {
		if !connectors[set.Connector] || seen[set.Connector] {
			return fmt.Errorf("workflow connector %q is unknown or duplicated", set.Connector)
		}
		seen[set.Connector] = true
		if len(set.Cells) != len(kinds) {
			return fmt.Errorf("workflow connector %q has %d cells for %d kinds", set.Connector, len(set.Cells), len(kinds))
		}
		seenCells := make(map[string]bool, len(set.Cells))
		for _, cell := range set.Cells {
			if !kinds[cell.WorkflowKind] || seenCells[cell.WorkflowKind] {
				return fmt.Errorf("workflow connector %q has unknown or duplicate kind %q", set.Connector, cell.WorkflowKind)
			}
			seenCells[cell.WorkflowKind] = true
			if err := validateWorkflowCertificationCell(cell); err != nil {
				return fmt.Errorf("workflow connector %q kind %q: %w", set.Connector, cell.WorkflowKind, err)
			}
		}
		if set.Complete != workflowCellsComplete(set.Cells) {
			return fmt.Errorf("workflow connector %q complete disagrees with cells", set.Connector)
		}
	}
	if len(seen) != len(connectors) {
		return errors.New("workflow sets omit one or more connectors")
	}
	return nil
}

func validateConnectorSyncModeSets(sets []connectorSyncModeSet, connectors map[string]bool, modes map[string]bool, primitives map[string]syncPrimitive) error {
	seen := make(map[string]bool, len(sets))
	expected := len(modes) * len(primitives)
	for _, set := range sets {
		if !connectors[set.Connector] || seen[set.Connector] {
			return fmt.Errorf("sync-mode connector %q is unknown or duplicated", set.Connector)
		}
		seen[set.Connector] = true
		if len(set.Cells) != expected {
			return fmt.Errorf("sync-mode connector %q has %d cells for %d mode/primitive combinations", set.Connector, len(set.Cells), expected)
		}
		seenCells := make(map[string]bool, len(set.Cells))
		for _, cell := range set.Cells {
			if !modes[cell.SyncMode] {
				return fmt.Errorf("sync-mode connector %q has unknown mode %q", set.Connector, cell.SyncMode)
			}
			if _, ok := primitives[cell.Primitive]; !ok {
				return fmt.Errorf("sync-mode connector %q has unknown primitive %q", set.Connector, cell.Primitive)
			}
			key := cell.SyncMode + "\x00" + cell.Primitive
			if seenCells[key] {
				return fmt.Errorf("sync-mode connector %q duplicates %q", set.Connector, key)
			}
			seenCells[key] = true
			if err := validateSyncModeCertificationCell(cell); err != nil {
				return fmt.Errorf("sync-mode connector %q %q: %w", set.Connector, key, err)
			}
		}
		if set.Complete != syncModeCellsComplete(set.Cells) {
			return fmt.Errorf("sync-mode connector %q complete disagrees with cells", set.Connector)
		}
	}
	if len(seen) != len(connectors) {
		return errors.New("sync-mode sets omit one or more connectors")
	}
	return nil
}

func validateConnectorStatuses(statuses []connectorCertificationStatus, connectors map[string]bool) error {
	if len(statuses) != len(connectors) {
		return errors.New("connector statuses omit one or more connectors")
	}
	seen := make(map[string]bool, len(statuses))
	for _, status := range statuses {
		if !connectors[status.Connector] || seen[status.Connector] {
			return fmt.Errorf("connector status %q is unknown or duplicated", status.Connector)
		}
		seen[status.Connector] = true
		if status.Certified {
			if status.Label != "CERTIFIED" || status.Warning != "" {
				return fmt.Errorf("certified connector status %q is malformed", status.Connector)
			}
			continue
		}
		if status.Label != "COMMUNITY BUILD, UNCERTIFIED" || status.Warning != "This connector is reachable but is a COMMUNITY BUILD, UNCERTIFIED." {
			return fmt.Errorf("uncertified connector status %q is malformed", status.Connector)
		}
	}
	return nil
}

func validateFlowPairCoverage(matrix flowMatrix, knownConnectors map[string]bool) error {
	names := make([]string, 0, len(knownConnectors))
	for name := range knownConnectors {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, kind := range matrix.FlowKinds {
		for _, source := range names {
			for _, destination := range names {
				count := 0
				for _, set := range matrix.PairSets {
					if set.FlowKind == kind.ID && containsString(set.SourceConnectors, source) && containsString(set.DestinationConnectors, destination) {
						count++
					}
				}
				if count != 1 {
					return fmt.Errorf("flow pair coverage %s %s -> %s has %d cells, want 1", kind.ID, source, destination, count)
				}
			}
		}
	}
	return nil
}

func isSortedUnique(values []string) bool {
	if len(values) == 0 {
		return false
	}
	for index := 1; index < len(values); index++ {
		if values[index-1] >= values[index] {
			return false
		}
	}
	return true
}

func validateCertificationStatusArtifactJSON(raw []byte) error {
	var artifact certificationStatusArtifact
	if err := decodeStrictJSON(raw, &artifact); err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if artifact.SchemaVersion != certificationSchemaVersion {
		return fmt.Errorf("schema_version %d is unsupported", artifact.SchemaVersion)
	}
	if artifact.GeneratedCommand != "go run ./cmd/connectorgen certification-matrix --all" {
		return fmt.Errorf("generated_command %q is unsupported", artifact.GeneratedCommand)
	}
	known, err := certificationStatusScope(artifact.CertificationScope)
	if err != nil {
		return err
	}
	return validateConnectorStatuses(artifact.Connectors, known)
}

func certificationStatusScope(scope []string) (map[string]bool, error) {
	if len(scope) != len(certificationConnectorAllowlist) {
		return nil, errors.New("certification status scope omits one or more allowlisted connectors")
	}
	known := make(map[string]bool, len(scope))
	for index, allowed := range certificationConnectorAllowlist {
		if scope[index] != allowed {
			return nil, fmt.Errorf("certification status scope connector %d = %q, want %q", index, scope[index], allowed)
		}
		known[allowed] = true
	}
	return known, nil
}

func validateCertificationStatusArtifactFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("generated artifact %q is missing; run `go run ./cmd/connectorgen certification-matrix --all`", filepath.ToSlash(path))
		}
		return fmt.Errorf("read generated artifact %q: %w", filepath.ToSlash(path), err)
	}
	if err := validateCertificationStatusArtifactJSON(raw); err != nil {
		return fmt.Errorf("generated certification artifact %q is invalid: %w", filepath.ToSlash(path), err)
	}
	return nil
}

func checkCertificationStatusGeneratedArtifact(path string, generated []byte) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("generated artifact %q is missing; run `go run ./cmd/connectorgen certification-matrix --all`", filepath.ToSlash(path))
		}
		return fmt.Errorf("read generated artifact %q: %w", filepath.ToSlash(path), err)
	}
	if err := validateCertificationStatusArtifactJSON(existing); err != nil {
		return fmt.Errorf("generated certification artifact %q is invalid: %w", filepath.ToSlash(path), err)
	}
	if !bytes.Equal(existing, generated) {
		return fmt.Errorf("generated artifact %q has drift; run `go run ./cmd/connectorgen certification-matrix --all`", filepath.ToSlash(path))
	}
	return nil
}
