package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/conformance"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	certificationSchemaVersion = 1
	capabilityMatrixPath       = "internal/connectors/certifications/capability-matrix.json"
	acceptedEvidenceDirectory  = "internal/connectors/certifications/evidence"

	executorAnnotationPrefix = "pmcert:executes"

	evidenceScopeCapability = "capability"
	evidenceScopeFlow       = "flow"
	evidenceScopeWorkflow   = "workflow"
	evidenceScopeSyncMode   = "sync_mode"
	evidenceStatusPassed    = "passed"
)

// functionKind is a single unit in the code-derived inventory. It deliberately
// carries the source that discovered it so reviewers can see which executable
// contract made the row required.
type functionKind struct {
	ID              string `json:"id"`
	Category        string `json:"category"`
	Name            string `json:"name"`
	DiscoverySource string `json:"discovery_source"`
	ExecutorSource  string `json:"executor_source,omitempty"`
}

// notApplicableReason explains why a generated row is outside a connector's
// contract. Code is a closed, specific classifier; generic "n/a" and
// "blocked" labels are rejected by validateCertificationCell.
type notApplicableReason struct {
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

// evidencePointer repeats the accepted record's publishable proof inside a
// generated certification shard. Transcript values are already repository-salted
// fingerprints, so a shard can be read and checked without loading an
// untrusted sidecar or ever becoming a credential/provider-data store.
type evidencePointer struct {
	Record          string                `json:"record"`
	Provider        string                `json:"provider"`
	ExecutedAt      string                `json:"executed_at"`
	RunID           string                `json:"run_id"`
	CredentialScope string                `json:"credential_scope"`
	CredentialNote  string                `json:"credential_note"`
	Proof           embeddedEvidenceProof `json:"proof"`
}

// certificationCell records the four facts that must all be true for an
// applicable function kind. A cell is non-applicable only with a named reason.
type certificationCell struct {
	FunctionKind    string               `json:"function_kind"`
	Applicable      bool                 `json:"applicable"`
	Declared        bool                 `json:"declared"`
	Implemented     bool                 `json:"implemented"`
	FixtureTested   bool                 `json:"fixture_tested"`
	LiveTested      bool                 `json:"live_tested"`
	FixtureEvidence []string             `json:"fixture_evidence"`
	LiveEvidence    []evidencePointer    `json:"live_evidence"`
	NotApplicable   *notApplicableReason `json:"not_applicable,omitempty"`
}

// capabilityConnector is the connector-local capability payload.
// CapabilityComplete intentionally does not claim final connector
// certification; reconstructed flow records add pair requirements later.
type capabilityConnector struct {
	Name               string              `json:"name"`
	IntegrationType    string              `json:"integration_type"`
	CapabilityComplete bool                `json:"capability_complete"`
	Cells              []certificationCell `json:"cells"`
}

type legacyCertificationFile struct {
	Path           string `json:"path"`
	Classification string `json:"classification"`
}

type legacyCertificationInventory struct {
	Ignored bool                      `json:"ignored"`
	Files   []legacyCertificationFile `json:"files"`
}

type kindBaseline struct {
	FunctionKind  string `json:"function_kind"`
	Connectors    int    `json:"connectors"`
	Applicable    int    `json:"applicable"`
	Declared      int    `json:"declared"`
	Implemented   int    `json:"implemented"`
	FixtureTested int    `json:"fixture_tested"`
	LiveTested    int    `json:"live_tested"`
	Complete      int    `json:"complete"`
}

type capabilityBaseline struct {
	Connectors         int            `json:"connectors"`
	CapabilityComplete int            `json:"capability_complete"`
	PerKind            []kindBaseline `json:"per_kind"`
}

// capabilityMatrix is the in-memory aggregate view used to assemble and
// validate connector-local shards. Fields are deliberately slices, not maps,
// so json.MarshalIndent produces stable, reviewable ordering.
type capabilityMatrix struct {
	SchemaVersion             int                          `json:"schema_version"`
	GeneratedCommand          string                       `json:"generated_command"`
	FunctionKinds             []functionKind               `json:"function_kinds"`
	Connectors                []capabilityConnector        `json:"connectors"`
	LegacyCertificationInputs legacyCertificationInventory `json:"legacy_certification_inputs"`
	Baseline                  capabilityBaseline           `json:"baseline"`
}

// certificationShard is the connector-owned proof-bearing record. It repeats
// only source inventories needed to validate this connector's cells; it
// deliberately excludes global baselines, counts, and position-dependent
// data. Pair sets are reconstructed from connector roles when needed rather
// than becoming a cross-connector write dependency.
type certificationShard struct {
	SchemaVersion    int                  `json:"schema_version"`
	GeneratedCommand string               `json:"generated_command"`
	Connector        capabilityConnector  `json:"connector"`
	FunctionKinds    []functionKind       `json:"function_kinds"`
	FlowKinds        []flowKind           `json:"flow_kinds"`
	WorkflowKinds    []workflowKind       `json:"workflow_kinds"`
	Workflow         connectorWorkflowSet `json:"workflow"`
	SyncModeKinds    []syncModeKind       `json:"sync_mode_kinds"`
	SyncPrimitives   []syncPrimitive      `json:"sync_primitives"`
	SyncModeCells    connectorSyncModeSet `json:"sync_mode_cells"`
	ConnectorRoles   connectorFlowRoles   `json:"connector_roles"`
	PairOverrides    []flowPairOverride   `json:"pair_overrides"`
}

// acceptedEvidence is the only record shape from which live_tested can be
// derived. No existing bundle certification contract or fixture filename uses
// this schema or directory, so filename matches cannot promote a connector.
type acceptedEvidence struct {
	SchemaVersion   int                   `json:"schema_version"`
	Scope           string                `json:"scope"`
	Status          string                `json:"status"`
	CredentialScope string                `json:"credential_scope"`
	CredentialNote  string                `json:"credential_note"`
	Connector       string                `json:"connector,omitempty"`
	FunctionKind    string                `json:"function_kind,omitempty"`
	WorkflowKind    string                `json:"workflow_kind,omitempty"`
	SyncMode        string                `json:"sync_mode,omitempty"`
	Primitive       string                `json:"primitive,omitempty"`
	Source          string                `json:"source,omitempty"`
	Destination     string                `json:"destination,omitempty"`
	FlowKind        string                `json:"flow_kind,omitempty"`
	Provider        string                `json:"provider"`
	ExecutedAt      string                `json:"executed_at"`
	RunID           string                `json:"run_id"`
	Proof           embeddedEvidenceProof `json:"proof"`

	recordPath string
}

type acceptedEvidenceScopeIdentity struct {
	Scope       string `json:"scope"`
	Connector   string `json:"connector"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type matrixConnectorSource struct {
	name            string
	integrationType string
	bundle          *engine.Bundle
	connector       connectors.Connector
	conformance     *conformance.Report
}

// scopedPostgresMatrixConnector preserves PostgreSQL's native unsupported
// Write stub for matrix introspection while sourcing every other method from
// the bundle the scoped generator has already validated.
type scopedPostgresMatrixConnector struct {
	*engine.Connector
}

func (scopedPostgresMatrixConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (scopedPostgresMatrixConnector) ReadCDC(context.Context, connectors.CDCReadRequest, func(connectors.CDCEvent) error) error {
	return nil
}

// runCertificationMatrix implements the source-controlled generation and
// byte-for-byte drift gate. It deliberately never runs a provider request.
func runCertificationMatrix(args []string, stdout, stderr io.Writer) int {
	root := "."
	check := false
	all := false
	connector := ""
	for index := 1; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--check":
			check = true
		case arg == "--all":
			all = true
		case strings.HasPrefix(arg, "--connector="):
			connector = strings.TrimPrefix(arg, "--connector=")
		case arg == "--connector":
			if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
				logln(stderr, "connectorgen certification-matrix: --connector requires a name")
				return 2
			}
			index++
			connector = args[index]
		case strings.HasPrefix(arg, "-"):
			logf(stderr, "connectorgen certification-matrix: unknown flag %q\n", arg)
			return 2
		case root == ".":
			root = arg
		default:
			logf(stderr, "connectorgen certification-matrix: unexpected extra argument %q\n", arg)
			return 2
		}
	}
	if check && (all || connector != "") {
		logln(stderr, "connectorgen certification-matrix: --check cannot be combined with --all or --connector")
		return 2
	}
	if !check && !all && connector == "" {
		logln(stderr, "connectorgen certification-matrix: require --connector <name> for an incremental run, or --all for deliberate regeneration")
		return 2
	}
	if connector != "" && !certificationConnectorAllowed(connector) {
		logf(stderr, "connectorgen certification-matrix: connector %q is not certification-allowlisted\n", connector)
		return 2
	}
	if all && connector != "" {
		logln(stderr, "connectorgen certification-matrix: --all and --connector cannot be combined")
		return 2
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		logf(stderr, "connectorgen certification-matrix: resolve repository root: %v\n", err)
		return 1
	}
	result, err := generateCertificationMatrix(absRoot, absRoot, check, all, connector)
	if err != nil {
		logf(stderr, "connectorgen certification-matrix: %v\n", err)
		return 1
	}
	if check {
		logf(stdout, "certification shards are current: connectors=%d capability_complete=%d certified=%d\n", result.capabilities.Baseline.Connectors, result.capabilities.Baseline.CapabilityComplete, result.flows.Baseline.Certified)
		return 0
	}
	if all {
		logf(stdout, "generated certification shards: scope=%s connectors=%d capability_complete=%d certified=%d\n", strings.Join(result.scope, ","), result.capabilities.Baseline.Connectors, result.capabilities.Baseline.CapabilityComplete, result.flows.Baseline.Certified)
		return 0
	}
	logf(stdout, "generated certification shard: connector=%s\n", connector)
	return 0
}

type certificationMatrixGenerationResult struct {
	scope        []string
	capabilities capabilityMatrix
	flows        flowMatrix
}

// generateCertificationMatrix is the generator core. Source and output roots
// are separate so the incremental-output contract is directly testable: a
// scoped run writes only its requested shard.
func generateCertificationMatrix(sourceRoot, outputRoot string, check, all bool, connector string) (certificationMatrixGenerationResult, error) {
	result := certificationMatrixGenerationResult{}
	statusPath := filepath.Join(outputRoot, certificationStatusPath)
	if check {
		if err := checkRetiredCertificationArtifactsAbsent(outputRoot); err != nil {
			return result, err
		}
		committedShards, err := readCertificationShards(outputRoot)
		if err != nil {
			return result, err
		}
		if _, _, err := reconstructCertificationMatrices(committedShards); err != nil {
			return result, fmt.Errorf("reconstruct committed shard union: %w", err)
		}
		if err := validateCertificationStatusArtifactFile(statusPath); err != nil {
			return result, err
		}
	}
	// A normal scoped run writes only its requested connector. The all-shard
	// sweep is reserved for the deliberate status-projection refresh and drift
	// gate, so concurrent connector lanes never touch each other's output.
	generationScope := certificationConnectorAllowlist
	if connector != "" {
		generationScope = []string{connector}
	}
	shards, err := buildCertificationShards(sourceRoot, generationScope)
	if err != nil {
		return result, err
	}
	payloads, err := marshalCertificationShards(shards)
	if err != nil {
		return result, fmt.Errorf("render shards: %w", err)
	}
	if check {
		capabilities, flows, err := reconstructCertificationMatrices(shards)
		if err != nil {
			return result, fmt.Errorf("reconstruct shard union: %w", err)
		}
		statusPayload, err := marshalGeneratedJSON(buildCertificationStatusArtifact(flows))
		if err != nil {
			return result, fmt.Errorf("render certification status: %w", err)
		}
		if err := checkCertificationShards(outputRoot, payloads); err != nil {
			return result, err
		}
		if err := checkCertificationStatusGeneratedArtifact(statusPath, statusPayload); err != nil {
			return result, err
		}
		result.scope = generationScope
		result.capabilities = capabilities
		result.flows = flows
		return result, nil
	}
	if err := writeCertificationShardScope(outputRoot, payloads, generationScope); err != nil {
		return result, fmt.Errorf("write certification shards: %w", err)
	}
	result.scope = generationScope
	if all {
		capabilities, flows, err := reconstructCertificationMatrices(shards)
		if err != nil {
			return result, fmt.Errorf("reconstruct shard union: %w", err)
		}
		statusPayload, err := marshalGeneratedJSON(buildCertificationStatusArtifact(flows))
		if err != nil {
			return result, fmt.Errorf("render certification status: %w", err)
		}
		if err := writeGeneratedArtifact(statusPath, statusPayload); err != nil {
			return result, fmt.Errorf("write certification status: %w", err)
		}
		if err := removeRetiredCertificationArtifacts(outputRoot); err != nil {
			return result, err
		}
		result.capabilities = capabilities
		result.flows = flows
	}
	return result, nil
}

func certificationShardPath(repoRoot, connector string) string {
	return filepath.Join(repoRoot, "internal", "connectors", "defs", connector, "certification-matrix.json")
}

func buildCertificationShards(repoRoot string, names []string) (map[string]certificationShard, error) {
	if err := validateCertificationConnectorScope(names); err != nil {
		return nil, err
	}
	capabilities, err := buildCapabilityMatrixForConnectors(repoRoot, certificationConnectorAllowlist)
	if err != nil {
		return nil, err
	}
	flows, err := buildFlowMatrixForConnectors(repoRoot, capabilities, certificationConnectorAllowlist)
	if err != nil {
		return nil, err
	}
	shards := make(map[string]certificationShard, len(names))
	for _, name := range names {
		connector, found := capabilityConnectorByName(capabilities, name)
		if !found {
			return nil, fmt.Errorf("certification scope %q has no capability connector", name)
		}
		workflow, found := workflowSetByConnector(flows.Workflows, name)
		if !found {
			return nil, fmt.Errorf("certification scope %q has no workflow cells", name)
		}
		syncCells, found := syncModeSetByConnector(flows.SyncModeCells, name)
		if !found {
			return nil, fmt.Errorf("certification scope %q has no sync-mode cells", name)
		}
		roles, found := rolesByConnector(flows.ConnectorRoles, name)
		if !found {
			return nil, fmt.Errorf("certification scope %q has no flow roles", name)
		}
		shards[name] = certificationShard{
			SchemaVersion:    certificationSchemaVersion,
			GeneratedCommand: "go run ./cmd/connectorgen certification-matrix --connector " + name,
			Connector:        connector,
			FunctionKinds:    append([]functionKind(nil), capabilities.FunctionKinds...),
			FlowKinds:        append([]flowKind(nil), flows.FlowKinds...),
			WorkflowKinds:    append([]workflowKind(nil), flows.WorkflowKinds...),
			Workflow:         workflow,
			SyncModeKinds:    append([]syncModeKind(nil), flows.SyncModeKinds...),
			SyncPrimitives:   append([]syncPrimitive(nil), flows.SyncPrimitives...),
			SyncModeCells:    syncCells,
			ConnectorRoles:   roles,
			PairOverrides:    pairOverridesForSource(flows.PairOverrides, name),
		}
	}
	return shards, nil
}

func validateCertificationConnectorScope(names []string) error {
	if len(names) == 0 {
		return errors.New("certification connector scope is empty")
	}
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if !certificationConnectorAllowed(name) {
			return fmt.Errorf("connector %q is not certification-allowlisted", name)
		}
		if seen[name] {
			return fmt.Errorf("certification connector scope duplicates %q", name)
		}
		seen[name] = true
	}
	return nil
}

func capabilityConnectorByName(matrix capabilityMatrix, name string) (capabilityConnector, bool) {
	for _, connector := range matrix.Connectors {
		if connector.Name == name {
			return connector, true
		}
	}
	return capabilityConnector{}, false
}

func workflowSetByConnector(sets []connectorWorkflowSet, name string) (connectorWorkflowSet, bool) {
	for _, set := range sets {
		if set.Connector == name {
			return set, true
		}
	}
	return connectorWorkflowSet{}, false
}

func syncModeSetByConnector(sets []connectorSyncModeSet, name string) (connectorSyncModeSet, bool) {
	for _, set := range sets {
		if set.Connector == name {
			return set, true
		}
	}
	return connectorSyncModeSet{}, false
}

func rolesByConnector(sets []connectorFlowRoles, name string) (connectorFlowRoles, bool) {
	for _, set := range sets {
		if set.Connector == name {
			return set, true
		}
	}
	return connectorFlowRoles{}, false
}

func pairOverridesForSource(overrides []flowPairOverride, source string) []flowPairOverride {
	out := make([]flowPairOverride, 0)
	for _, override := range overrides {
		if override.Source == source {
			out = append(out, override)
		}
	}
	return out
}

func marshalCertificationShards(shards map[string]certificationShard) (map[string][]byte, error) {
	payloads := make(map[string][]byte, len(shards))
	for _, name := range certificationConnectorAllowlist {
		shard, found := shards[name]
		if !found {
			continue
		}
		payload, err := marshalGeneratedJSON(shard)
		if err != nil {
			return nil, fmt.Errorf("marshal connector %q shard: %w", name, err)
		}
		payloads[name] = payload
	}
	return payloads, nil
}

func writeCertificationShardScope(repoRoot string, payloads map[string][]byte, names []string) error {
	if err := validateCertificationConnectorScope(names); err != nil {
		return err
	}
	for _, name := range names {
		payload, found := payloads[name]
		if !found {
			return fmt.Errorf("connector %q has no generated shard payload", name)
		}
		if err := writeGeneratedArtifact(certificationShardPath(repoRoot, name), payload); err != nil {
			return fmt.Errorf("connector %q: %w", name, err)
		}
	}
	return nil
}

func readCertificationShards(repoRoot string) (map[string]certificationShard, error) {
	shards := make(map[string]certificationShard, len(certificationConnectorAllowlist))
	for _, name := range certificationConnectorAllowlist {
		path := certificationShardPath(repoRoot, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("generated certification shard %q is missing; run `go run ./cmd/connectorgen certification-matrix --connector %s`", filepath.ToSlash(path), name)
			}
			return nil, fmt.Errorf("read certification shard %q: %w", filepath.ToSlash(path), err)
		}
		var shard certificationShard
		if err := decodeStrictJSON(raw, &shard); err != nil {
			return nil, fmt.Errorf("parse certification shard %q: %w", filepath.ToSlash(path), err)
		}
		if err := validateCertificationShard(shard); err != nil {
			return nil, fmt.Errorf("invalid certification shard %q: %w", filepath.ToSlash(path), err)
		}
		if shard.Connector.Name != name {
			return nil, fmt.Errorf("certification shard %q owns connector %q, want %q", filepath.ToSlash(path), shard.Connector.Name, name)
		}
		shards[name] = shard
	}
	return shards, nil
}

func checkCertificationShards(repoRoot string, expected map[string][]byte) error {
	if _, err := readCertificationShards(repoRoot); err != nil {
		return err
	}
	for _, name := range certificationConnectorAllowlist {
		payload, found := expected[name]
		if !found {
			return fmt.Errorf("connector %q has no expected certification shard", name)
		}
		path := certificationShardPath(repoRoot, name)
		existing, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read certification shard %q: %w", filepath.ToSlash(path), err)
		}
		if !bytes.Equal(existing, payload) {
			return fmt.Errorf("generated certification shard %q has drift; run `go run ./cmd/connectorgen certification-matrix --connector %s`", filepath.ToSlash(path), name)
		}
	}
	return nil
}

func checkRetiredCertificationArtifactsAbsent(repoRoot string) error {
	for _, relative := range []string{capabilityMatrixPath, flowMatrixPath} {
		path := filepath.Join(repoRoot, relative)
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("retired aggregate %q is present; run `go run ./cmd/connectorgen certification-matrix --all`", filepath.ToSlash(path))
		} else if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("inspect retired aggregate %q: %w", filepath.ToSlash(path), err)
		}
	}
	return nil
}

func removeRetiredCertificationArtifacts(repoRoot string) error {
	for _, relative := range []string{capabilityMatrixPath, flowMatrixPath} {
		path := filepath.Join(repoRoot, relative)
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove %q: %w", filepath.ToSlash(path), err)
		}
	}
	return nil
}

// reconstructCertificationMatrices derives the former aggregate views only in
// memory. It is used for validation and the compact runtime status projection;
// the aggregate JSON files are intentionally never written again.
func reconstructCertificationMatrices(shards map[string]certificationShard) (capabilityMatrix, flowMatrix, error) {
	if len(shards) == 0 {
		return capabilityMatrix{}, flowMatrix{}, errors.New("certification shard union is empty")
	}
	names := make([]string, 0, len(shards))
	for name := range shards {
		names = append(names, name)
	}
	sort.Strings(names)
	first := shards[names[0]]
	capabilities := capabilityMatrix{
		SchemaVersion:    certificationSchemaVersion,
		GeneratedCommand: "go run ./cmd/connectorgen certification-matrix --check",
		FunctionKinds:    append([]functionKind(nil), first.FunctionKinds...),
		Connectors:       make([]capabilityConnector, 0, len(names)),
		LegacyCertificationInputs: legacyCertificationInventory{
			Ignored: true,
			Files:   []legacyCertificationFile{},
		},
	}
	flows := flowMatrix{
		SchemaVersion:     certificationSchemaVersion,
		GeneratedCommand:  "go run ./cmd/connectorgen certification-matrix --check",
		Mediator:          localWarehouseMediator,
		FlowKinds:         append([]flowKind(nil), first.FlowKinds...),
		WorkflowKinds:     append([]workflowKind(nil), first.WorkflowKinds...),
		Workflows:         make([]connectorWorkflowSet, 0, len(names)),
		SyncModeKinds:     append([]syncModeKind(nil), first.SyncModeKinds...),
		SyncPrimitives:    append([]syncPrimitive(nil), first.SyncPrimitives...),
		SyncModeCells:     make([]connectorSyncModeSet, 0, len(names)),
		ConnectorRoles:    make([]connectorFlowRoles, 0, len(names)),
		PairOverrides:     []flowPairOverride{},
		ConnectorStatuses: []connectorCertificationStatus{},
	}
	for _, name := range names {
		shard := shards[name]
		if err := validateCertificationShard(shard); err != nil {
			return capabilityMatrix{}, flowMatrix{}, fmt.Errorf("connector %q: %w", name, err)
		}
		if shard.Connector.Name != name {
			return capabilityMatrix{}, flowMatrix{}, fmt.Errorf("connector %q shard has mismatched owner %q", name, shard.Connector.Name)
		}
		if !reflect.DeepEqual(first.FunctionKinds, shard.FunctionKinds) || !reflect.DeepEqual(first.FlowKinds, shard.FlowKinds) || !reflect.DeepEqual(first.WorkflowKinds, shard.WorkflowKinds) || !reflect.DeepEqual(first.SyncModeKinds, shard.SyncModeKinds) || !reflect.DeepEqual(first.SyncPrimitives, shard.SyncPrimitives) {
			return capabilityMatrix{}, flowMatrix{}, fmt.Errorf("connector %q shard inventory differs from %q", name, names[0])
		}
		capabilities.Connectors = append(capabilities.Connectors, shard.Connector)
		flows.Workflows = append(flows.Workflows, shard.Workflow)
		flows.SyncModeCells = append(flows.SyncModeCells, shard.SyncModeCells)
		flows.ConnectorRoles = append(flows.ConnectorRoles, shard.ConnectorRoles)
		flows.PairOverrides = append(flows.PairOverrides, shard.PairOverrides...)
	}
	capabilities.Baseline = deriveCapabilityBaseline(capabilities.Connectors, capabilities.FunctionKinds)
	rolesByConnector := make(map[string]map[string]connectorFlowRole, len(flows.ConnectorRoles))
	for _, roles := range flows.ConnectorRoles {
		byRole := make(map[string]connectorFlowRole, len(roles.Roles))
		for _, role := range roles.Roles {
			byRole[role.Role] = role
		}
		rolesByConnector[roles.Connector] = byRole
	}
	pairSets, err := buildFlowPairSets(flows.FlowKinds, rolesByConnector)
	if err != nil {
		return capabilityMatrix{}, flowMatrix{}, err
	}
	flows.PairSets = pairSets
	flows.ConnectorStatuses = deriveConnectorStatuses(capabilities, flows)
	flows.Baseline = deriveFlowBaseline(flows)
	if err := validateFlowMatrix(flows); err != nil {
		return capabilityMatrix{}, flowMatrix{}, err
	}
	return capabilities, flows, nil
}

func validateCertificationShard(shard certificationShard) error {
	if shard.SchemaVersion != certificationSchemaVersion {
		return fmt.Errorf("schema_version %d is unsupported", shard.SchemaVersion)
	}
	if !certificationConnectorAllowed(shard.Connector.Name) {
		return fmt.Errorf("connector %q is not certification-allowlisted", shard.Connector.Name)
	}
	if shard.GeneratedCommand != "go run ./cmd/connectorgen certification-matrix --connector "+shard.Connector.Name {
		return fmt.Errorf("generated_command %q does not identify connector %q", shard.GeneratedCommand, shard.Connector.Name)
	}
	for _, kind := range shard.FunctionKinds {
		if err := validateSymbolSourceAnchor(kind.DiscoverySource); err != nil {
			return fmt.Errorf("function kind %q discovery_source: %w", kind.ID, err)
		}
		if kind.ExecutorSource != "" {
			if err := validateSymbolSourceAnchor(kind.ExecutorSource); err != nil {
				return fmt.Errorf("function kind %q executor_source: %w", kind.ID, err)
			}
		}
	}
	for _, kind := range shard.WorkflowKinds {
		if err := validateSymbolSourceAnchor(kind.DiscoverySource); err != nil {
			return fmt.Errorf("workflow kind %q discovery_source: %w", kind.ID, err)
		}
	}
	for _, kind := range shard.SyncModeKinds {
		if err := validateSymbolSourceAnchor(kind.DiscoverySource); err != nil {
			return fmt.Errorf("sync mode %q discovery_source: %w", kind.ID, err)
		}
	}
	for _, primitive := range shard.SyncPrimitives {
		if err := validateSymbolSourceAnchor(primitive.DiscoverySource); err != nil {
			return fmt.Errorf("sync primitive %q discovery_source: %w", primitive.ID, err)
		}
	}
	capabilities := capabilityMatrix{FunctionKinds: shard.FunctionKinds, Connectors: []capabilityConnector{shard.Connector}}
	if err := validateCapabilityMatrix(capabilities); err != nil {
		return err
	}
	flows := flowMatrix{
		FlowKinds:         shard.FlowKinds,
		WorkflowKinds:     shard.WorkflowKinds,
		Workflows:         []connectorWorkflowSet{shard.Workflow},
		SyncModeKinds:     shard.SyncModeKinds,
		SyncPrimitives:    shard.SyncPrimitives,
		SyncModeCells:     []connectorSyncModeSet{shard.SyncModeCells},
		ConnectorRoles:    []connectorFlowRoles{shard.ConnectorRoles},
		PairOverrides:     []flowPairOverride{},
		ConnectorStatuses: []connectorCertificationStatus{},
	}
	// Full flow validation requires derived pair sets and statuses. Build them
	// from this singleton's roles before validating the complete structure.
	rolesByConnector := map[string]map[string]connectorFlowRole{shard.Connector.Name: {}}
	for _, role := range shard.ConnectorRoles.Roles {
		rolesByConnector[shard.Connector.Name][role.Role] = role
	}
	pairSets, err := buildFlowPairSets(flows.FlowKinds, rolesByConnector)
	if err != nil {
		return err
	}
	flows.SchemaVersion = certificationSchemaVersion
	flows.GeneratedCommand = "go run ./cmd/connectorgen certification-matrix --check"
	flows.Mediator = localWarehouseMediator
	flows.PairSets = pairSets
	flows.ConnectorStatuses = deriveConnectorStatuses(capabilities, flows)
	flows.Baseline = deriveFlowBaseline(flows)
	if err := validateFlowMatrix(flows); err != nil {
		return err
	}
	seenOverrides := make(map[string]bool, len(shard.PairOverrides))
	knownFlowKinds := make(map[string]flowKind, len(shard.FlowKinds))
	for _, kind := range shard.FlowKinds {
		knownFlowKinds[kind.ID] = kind
	}
	for _, override := range shard.PairOverrides {
		key := strings.Join([]string{override.FlowKind, override.Source, override.Destination}, "\x00")
		if seenOverrides[key] {
			return fmt.Errorf("pair override %q is duplicated", key)
		}
		seenOverrides[key] = true
		kind, known := knownFlowKinds[override.FlowKind]
		if override.Source != shard.Connector.Name || !certificationConnectorAllowed(override.Destination) || !known || override.Mediator != localWarehouseMediator {
			return fmt.Errorf("pair override %q has an invalid shard-local identity", key)
		}
		if err := validateConnectorFlowRole(override.DestinationRole); err != nil || override.DestinationRole.Role != kind.DestinationRole {
			return fmt.Errorf("pair override %q has invalid destination role context", key)
		}
		sourceRole, found := flowRoleForConnector([]connectorFlowRoles{shard.ConnectorRoles}, shard.Connector.Name, kind.SourceRole)
		if !found {
			return fmt.Errorf("pair override %q cannot target an inapplicable source or destination role", key)
		}
		base := baseFlowCell(sourceRole, override.DestinationRole)
		if !base.Applicable {
			return fmt.Errorf("pair override %q cannot target an inapplicable source or destination role", key)
		}
		if !flowCellMatchesRoleBase(override.Cell, base) {
			return fmt.Errorf("pair override %q facts do not match its source and destination roles", key)
		}
		if err := validateFlowCertificationCell(override.Cell); err != nil {
			return fmt.Errorf("pair override %q: %w", key, err)
		}
	}
	return nil
}

// buildCapabilityMatrix derives every matrix fact from source, registered
// runtime types, recorded fixtures, or an accepted evidence record. It remains
// inside connectorgen so the developer tool is the sole owner of the generated
// shard format.
func buildCapabilityMatrix(repoRoot string) (capabilityMatrix, error) {
	return buildCapabilityMatrixForConnectors(repoRoot, nil)
}

func buildCapabilityMatrixForConnectors(repoRoot string, names []string) (capabilityMatrix, error) {
	kinds, err := discoverFunctionKinds(repoRoot)
	if err != nil {
		return capabilityMatrix{}, err
	}
	bundles, err := loadSourceBundlesForConnectors(repoRoot, names)
	if err != nil {
		return capabilityMatrix{}, err
	}
	evidence, err := loadAcceptedEvidence(repoRoot, names)
	if err != nil {
		return capabilityMatrix{}, err
	}
	sources, err := matrixConnectorSourcesForNames(bundles, names)
	if err != nil {
		return capabilityMatrix{}, err
	}

	connectorsOut := make([]capabilityConnector, 0, len(sources))
	for _, source := range sources {
		cells := make([]certificationCell, 0, len(kinds))
		for _, kind := range kinds {
			cell, err := buildCertificationCell(repoRoot, source, kind, evidence)
			if err != nil {
				return capabilityMatrix{}, fmt.Errorf("connector %q %s: %w", source.name, kind.ID, err)
			}
			if err := validateCertificationCell(cell); err != nil {
				return capabilityMatrix{}, fmt.Errorf("connector %q %s: %w", source.name, kind.ID, err)
			}
			cells = append(cells, cell)
		}
		connectorsOut = append(connectorsOut, capabilityConnector{
			Name:               source.name,
			IntegrationType:    source.integrationType,
			CapabilityComplete: certificationComplete(cells),
			Cells:              cells,
		})
	}

	legacy, err := discoverLegacyCertificationInputs(repoRoot)
	if err != nil {
		return capabilityMatrix{}, err
	}
	matrix := capabilityMatrix{
		SchemaVersion:             certificationSchemaVersion,
		GeneratedCommand:          "go run ./cmd/connectorgen certification-matrix",
		FunctionKinds:             kinds,
		Connectors:                connectorsOut,
		LegacyCertificationInputs: legacy,
	}
	matrix.Baseline = deriveCapabilityBaseline(matrix.Connectors, matrix.FunctionKinds)
	return matrix, nil
}

func loadSourceBundles(repoRoot string) ([]engine.Bundle, error) {
	defsRoot := filepath.Join(repoRoot, "internal", "connectors", "defs")
	bundles, err := engine.LoadAll(os.DirFS(defsRoot))
	if err != nil {
		return nil, fmt.Errorf("load source connector bundles: %w", err)
	}
	return bundles, nil
}

func loadSourceBundlesForConnectors(repoRoot string, names []string) ([]engine.Bundle, error) {
	if len(names) == 0 {
		return loadSourceBundles(repoRoot)
	}
	defsRoot := filepath.Join(repoRoot, "internal", "connectors", "defs")
	sourceFS := os.DirFS(defsRoot)
	bundles := make([]engine.Bundle, 0, len(names))
	for _, name := range names {
		scopedFS, err := scopedRuntimeOperationEndpointLedgerForCertification(sourceFS, name)
		if err != nil {
			return nil, fmt.Errorf("scope source connector bundle %q runtime operation endpoint ledger: %w", name, err)
		}
		bundle, err := engine.Load(scopedFS, name)
		if err != nil {
			return nil, fmt.Errorf("load source connector bundle %q: %w", name, err)
		}
		bundles = append(bundles, bundle)
	}
	return bundles, nil
}

// scopedRuntimeOperationEndpointLedgerForCertification presents one connector's
// generated ledger entry to the certification generator. Runtime loading still
// validates the complete shipped ledger; this isolated view exists only so a
// connector-local certification shard is independent of unrelated entries.
func scopedRuntimeOperationEndpointLedgerForCertification(source fs.FS, connector string) (fs.FS, error) {
	raw, err := fs.ReadFile(source, engine.RuntimeOperationEndpointLedgerFile)
	if errors.Is(err, fs.ErrNotExist) {
		return source, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", engine.RuntimeOperationEndpointLedgerFile, err)
	}
	var entries map[string]json.RawMessage
	if err := decodeStrictJSON(raw, &entries); err != nil {
		return nil, fmt.Errorf("parse %s: %w", engine.RuntimeOperationEndpointLedgerFile, err)
	}
	scopedEntries := make(map[string]json.RawMessage, 1)
	if entry, found := entries[connector]; found {
		scopedEntries[connector] = entry
	}
	scopedRaw, err := json.Marshal(scopedEntries)
	if err != nil {
		return nil, fmt.Errorf("encode %s: %w", engine.RuntimeOperationEndpointLedgerFile, err)
	}
	return certificationRuntimeOperationEndpointLedgerFS{FS: source, raw: scopedRaw}, nil
}

type certificationRuntimeOperationEndpointLedgerFS struct {
	fs.FS
	raw []byte
}

func (f certificationRuntimeOperationEndpointLedgerFS) Open(name string) (fs.File, error) {
	if name == engine.RuntimeOperationEndpointLedgerFile {
		return &certificationRuntimeOperationEndpointLedgerFile{
			Reader: bytes.NewReader(f.raw),
			size:   int64(len(f.raw)),
		}, nil
	}
	return f.FS.Open(name)
}

type certificationRuntimeOperationEndpointLedgerFile struct {
	*bytes.Reader
	size int64
}

func (f *certificationRuntimeOperationEndpointLedgerFile) Close() error {
	return nil
}

func (f *certificationRuntimeOperationEndpointLedgerFile) Stat() (fs.FileInfo, error) {
	return certificationRuntimeOperationEndpointLedgerFileInfo{size: f.size}, nil
}

type certificationRuntimeOperationEndpointLedgerFileInfo struct {
	size int64
}

func (info certificationRuntimeOperationEndpointLedgerFileInfo) Name() string {
	return engine.RuntimeOperationEndpointLedgerFile
}

func (info certificationRuntimeOperationEndpointLedgerFileInfo) Size() int64 {
	return info.size
}

func (certificationRuntimeOperationEndpointLedgerFileInfo) Mode() fs.FileMode {
	return 0o444
}

func (certificationRuntimeOperationEndpointLedgerFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (certificationRuntimeOperationEndpointLedgerFileInfo) IsDir() bool {
	return false
}

func (certificationRuntimeOperationEndpointLedgerFileInfo) Sys() any {
	return nil
}

func matrixConnectorSourcesForNames(bundles []engine.Bundle, scope []string) ([]matrixConnectorSource, error) {
	bundleByName := make(map[string]*engine.Bundle, len(bundles))
	for i := range bundles {
		bundleByName[bundles[i].Name] = &bundles[i]
	}

	names := make(map[string]bool, len(bundleByName))
	var registry *connectors.Registry
	if len(scope) != 0 {
		for _, name := range scope {
			names[name] = true
		}
	} else {
		registry = bundleregistry.New()
		for name := range bundleByName {
			names[name] = true
		}
		for _, metadata := range registry.List() {
			names[metadata.Name] = true
		}
	}

	sortedNames := make([]string, 0, len(names))
	for name := range names {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	out := make([]matrixConnectorSource, 0, len(sortedNames))
	for _, name := range sortedNames {
		bundle := bundleByName[name]
		var connector connectors.Connector
		var registered bool
		if registry != nil {
			connector, registered = registry.Get(name)
		} else {
			connector, registered = scopedMatrixConnector(name, bundle)
		}
		integrationType := ""
		if bundle != nil {
			integrationType = bundle.Metadata.IntegrationType
		}
		if registered && integrationType == "" {
			integrationType = connector.Metadata().IntegrationType
		}
		if integrationType == "" {
			integrationType = "unknown"
		}

		var report *conformance.Report
		if bundle != nil {
			r := conformance.RunBundle(*bundle)
			report = &r
		}
		out = append(out, matrixConnectorSource{
			name:            name,
			integrationType: integrationType,
			bundle:          bundle,
			connector:       connector,
			conformance:     report,
		})
	}
	return out, nil
}

func scopedMatrixConnector(name string, bundle *engine.Bundle) (connectors.Connector, bool) {
	if bundle == nil {
		return nil, false
	}
	if name == "postgres" {
		// Certification has already loaded this bundle through its scoped ledger
		// filesystem. Calling the native factory here would reload defs.FS and
		// make an unrelated, non-allowlisted ledger entry able to break this
		// generator. Preserve PostgreSQL's native unsupported Write shape while
		// keeping the source bundle that the scoped generator actually validated.
		return scopedPostgresMatrixConnector{Connector: engine.New(*bundle, nil)}, true
	}
	return engine.New(*bundle, nil), true
}

func buildCertificationCell(repoRoot string, source matrixConnectorSource, kind functionKind, evidence []acceptedEvidence) (certificationCell, error) {
	declared := functionKindDeclared(source, kind)
	if !functionKindApplicable(source, kind, declared) {
		return certificationCell{
			FunctionKind:    kind.ID,
			NotApplicable:   nonApplicableFor(source, kind),
			FixtureEvidence: []string{},
			LiveEvidence:    []evidencePointer{},
		}, nil
	}

	liveEvidence := matchingCapabilityEvidence(evidence, source, kind.ID)
	implemented, err := functionKindImplemented(repoRoot, source, kind, evidence)
	if err != nil {
		return certificationCell{}, err
	}
	fixtureTested, fixtureEvidence := functionKindFixtureTested(source, kind)
	cell := certificationCell{
		FunctionKind:    kind.ID,
		Applicable:      true,
		Declared:        declared,
		Implemented:     implemented,
		FixtureTested:   fixtureTested,
		LiveTested:      len(liveEvidence) > 0,
		FixtureEvidence: fixtureEvidence,
		LiveEvidence:    liveEvidence,
	}
	return cell, nil
}

// functionKindApplicable distinguishes an absent capability from one which a
// connector exposes but has not honestly declared. A direct unsupported stub
// such as Postgres.Write is therefore an applicable, declared=false,
// implemented=false row rather than a hidden N/A exemption.
func functionKindApplicable(source matrixConnectorSource, kind functionKind, declared bool) bool {
	if declared {
		return true
	}
	if kind.Category != "capability" || source.connector == nil {
		return false
	}
	// Every declarative connector has the generic engine Write method so it can
	// serve bundles that do declare writes. Without a writes.json action that
	// method is deliberately unavailable for this bundle and is a genuine N/A,
	// unlike a native database connector's concrete Write stub.
	if kind.Name == "write" && isEngineConnector(source.connector) && source.bundle != nil && len(source.bundle.Writes) == 0 {
		return false
	}
	_, exposed := capabilityMethod(source.connector, kind.Name)
	return exposed
}

func functionKindDeclared(source matrixConnectorSource, kind functionKind) bool {
	switch kind.Category {
	case "capability":
		if source.bundle != nil {
			if declared, found := boolFieldForKind(source.bundle.Metadata.Capabilities, kind.Name); found {
				return declared
			}
		}
		if source.connector == nil {
			return false
		}
		declared, _ := boolFieldForKind(source.connector.Metadata().Capabilities, kind.Name)
		return declared
	case "operation":
		if source.bundle == nil {
			return false
		}
		for _, operation := range source.bundle.Operations {
			if operation.Kind == kind.Name {
				return true
			}
		}
	}
	return false
}

func nonApplicableFor(source matrixConnectorSource, kind functionKind) *notApplicableReason {
	if kind.Category == "operation" {
		return &notApplicableReason{
			Code:   "operation_kind_not_declared",
			Reason: fmt.Sprintf("connector %q does not declare operation kind %q", source.name, kind.Name),
		}
	}
	if kind.Name == "write" && isEngineConnector(source.connector) && source.bundle != nil && len(source.bundle.Writes) == 0 {
		return &notApplicableReason{
			Code:   "no_declared_write_actions",
			Reason: fmt.Sprintf("connector %q has no declared writes.json actions; the shared engine Write method is unavailable for this bundle", source.name),
		}
	}
	return &notApplicableReason{
		Code:   "capability_not_exposed",
		Reason: fmt.Sprintf("connector %q neither declares capability %q nor exposes its runtime method", source.name, kind.Name),
	}
}

func functionKindImplemented(repoRoot string, source matrixConnectorSource, kind functionKind, evidence []acceptedEvidence) (bool, error) {
	switch kind.Category {
	case "operation":
		return kind.ExecutorSource != "", nil
	case "capability":
		return capabilityImplemented(repoRoot, source, kind.Name, evidence)
	}
	return false, nil
}

func capabilityImplemented(repoRoot string, source matrixConnectorSource, capability string, evidence []acceptedEvidence) (bool, error) {
	if source.connector == nil {
		return false, nil
	}
	// A native database destination is dispatched through its definition-owned
	// warehouse transport, not Connector.Write. Preserve the direct Write stub
	// check below for every other shape. This public capability requires two
	// distinct evidence classes: an accepted capability:write record produced
	// from the aggregate certification profile, plus exact sync-mode records for
	// every declared destination mode. Mode records are deliberately not bound
	// into the capability cell.
	if capability == "write" && declaredNativeDatabaseDestination(source) {
		return len(matchingCapabilityEvidence(evidence, source, "capability:write")) > 0 && len(declaredNativeDatabaseDestinationModeEvidence(source, evidence)) > 0, nil
	}
	method, ok := capabilityMethod(source.connector, capability)
	if !ok {
		return false, nil
	}

	if capability == "write" && isEngineConnector(source.connector) && source.bundle != nil {
		return len(source.bundle.Writes) > 0, nil
	}
	unsupported, err := methodDirectlyReturnsUnsupported(repoRoot, source.connector, method)
	if err != nil {
		return false, err
	}
	return !unsupported, nil
}

// capabilityMethod discovers a real exported method instead of retaining a
// copied capability-to-interface table. Exact public names win; for contracts
// such as ReadCDC and MapSchema, the normalized capability vocabulary is used
// as a narrow fallback.
func capabilityMethod(connector connectors.Connector, capability string) (string, bool) {
	typ := reflect.TypeOf(connector)
	if typ == nil {
		return "", false
	}
	want := goIdentifier(capability)
	methodTypes := []reflect.Type{typ}
	if typ.Kind() != reflect.Pointer {
		methodTypes = append(methodTypes, reflect.PointerTo(typ))
	}
	for _, methodType := range methodTypes {
		if _, ok := methodType.MethodByName(want); ok {
			return want, true
		}
	}

	normalizedCapability := normalizeIdentifier(capability)
	best := ""
	for _, methodType := range methodTypes {
		for i := range methodType.NumMethod() {
			name := methodType.Method(i).Name
			if strings.Contains(normalizeIdentifier(name), normalizedCapability) {
				if best == "" || name < best {
					best = name
				}
			}
		}
	}
	if best != "" {
		return best, true
	}

	for _, methodType := range methodTypes {
		for i := range methodType.NumMethod() {
			name := methodType.Method(i).Name
			if identifierSharesMeaning(name, capability) {
				if best == "" || name < best {
					best = name
				}
			}
		}
	}
	return best, best != ""
}

func isEngineConnector(connector connectors.Connector) bool {
	typ := reflect.TypeOf(connector)
	if typ == nil {
		return false
	}
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ.PkgPath() == "polymetrics.ai/internal/connectors/engine" && typ.Name() == "Connector"
}

// methodDirectlyReturnsUnsupported inspects only the concrete registered
// method's source. It avoids calling Write/Read while generating a shard and
// therefore cannot accidentally make a network request or mutate a provider.
func methodDirectlyReturnsUnsupported(repoRoot string, connector connectors.Connector, method string) (bool, error) {
	typ := reflect.TypeOf(connector)
	if typ == nil {
		return false, nil
	}
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	const modulePrefix = "polymetrics.ai/"
	var dir string
	if typ.PkgPath() == "main" {
		// The scoped PostgreSQL matrix view is generator-local. `go run` records
		// its runtime type under main, unlike a test build which retains the
		// module path, so resolve its real source directory explicitly.
		dir = filepath.Join(repoRoot, "cmd", "connectorgen")
	} else {
		if !strings.HasPrefix(typ.PkgPath(), modulePrefix) {
			return false, nil
		}
		dir = filepath.Join(repoRoot, strings.TrimPrefix(typ.PkgPath(), modulePrefix))
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, fmt.Errorf("read runtime implementation source %q: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return false, fmt.Errorf("parse runtime implementation source %q: %w", path, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != method || receiverTypeName(fn) != typ.Name() {
				continue
			}
			return containsUnsupportedOperation(fn.Body), nil
		}
	}
	return false, nil
}

func receiverTypeName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return ""
	}
	return exprIdentifier(fn.Recv.List[0].Type)
}

func exprIdentifier(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return exprIdentifier(typed.X)
	case *ast.IndexExpr:
		return exprIdentifier(typed.X)
	case *ast.IndexListExpr:
		return exprIdentifier(typed.X)
	}
	return ""
}

func containsUnsupportedOperation(body *ast.BlockStmt) bool {
	unsupported := false
	ast.Inspect(body, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if ok && selector.Sel.Name == "ErrUnsupportedOperation" {
			unsupported = true
			return false
		}
		return !unsupported
	})
	return unsupported
}

func functionKindFixtureTested(source matrixConnectorSource, kind functionKind) (bool, []string) {
	if source.bundle == nil || source.conformance == nil {
		return false, []string{}
	}

	switch kind.ID {
	case "capability:check":
		if !conformanceCheckPassed(*source.conformance, "check_fixture") || !fixtureFileExists(source.bundle, "check.json") {
			return false, []string{}
		}
		return true, []string{fixtureEvidencePath(source.bundle.Name, "check.json")}
	case "capability:read":
		if len(source.bundle.Streams) == 0 {
			return false, []string{}
		}
		evidence := make([]string, 0, len(source.bundle.Streams))
		for _, stream := range source.bundle.Streams {
			if !conformanceCheckPassed(*source.conformance, "read_fixture_nonempty:"+stream.Name) {
				return false, []string{}
			}
			paths, ok := fixtureFiles(source.bundle, filepath.ToSlash(filepath.Join("streams", stream.Name, "*.json")))
			if !ok {
				return false, []string{}
			}
			for _, path := range paths {
				evidence = append(evidence, fixtureEvidencePath(source.bundle.Name, path))
			}
		}
		sort.Strings(evidence)
		return true, evidence
	case "capability:write":
		if len(source.bundle.Writes) == 0 {
			return false, []string{}
		}
		evidence := make([]string, 0, len(source.bundle.Writes))
		for _, action := range source.bundle.Writes {
			fixture := filepath.ToSlash(filepath.Join("writes", action.Name+".json"))
			if !conformanceCheckPassed(*source.conformance, "write_request_shape:"+action.Name) || !fixtureFileExists(source.bundle, fixture) {
				return false, []string{}
			}
			evidence = append(evidence, fixtureEvidencePath(source.bundle.Name, fixture))
		}
		sort.Strings(evidence)
		return true, evidence
	default:
		// The current conformance corpus has no operation-specific replay
		// contract for direct operations, query, CDC, catalog, or dynamic
		// schema. Treating a generic fixture as evidence for one would repeat
		// the reachability-vs-correctness error this matrix exists to prevent.
		return false, []string{}
	}
}

func conformanceCheckPassed(report conformance.Report, name string) bool {
	for _, check := range report.Checks {
		if check.Name == name {
			return check.Passed && !check.Skipped
		}
	}
	return false
}

func fixtureFileExists(bundle *engine.Bundle, name string) bool {
	if bundle.Fixtures == nil {
		return false
	}
	_, err := fs.Stat(bundle.Fixtures, filepath.ToSlash(name))
	return err == nil
}

func fixtureFiles(bundle *engine.Bundle, pattern string) ([]string, bool) {
	if bundle.Fixtures == nil {
		return nil, false
	}
	paths, err := fs.Glob(bundle.Fixtures, pattern)
	if err != nil || len(paths) == 0 {
		return nil, false
	}
	sort.Strings(paths)
	return paths, true
}

func fixtureEvidencePath(connectorName, path string) string {
	return filepath.ToSlash(filepath.Join("internal", "connectors", "defs", connectorName, "fixtures", path))
}

func matchingCapabilityEvidence(evidence []acceptedEvidence, source matrixConnectorSource, kind string) []evidencePointer {
	matched := make([]evidencePointer, 0)
	for _, item := range evidence {
		if item.Scope != evidenceScopeCapability || item.Status != evidenceStatusPassed || item.Connector != source.name || item.FunctionKind != kind {
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
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].Record == matched[j].Record {
			return matched[i].RunID < matched[j].RunID
		}
		return matched[i].Record < matched[j].Record
	})
	return matched
}

func certificationComplete(cells []certificationCell) bool {
	applicable := 0
	for _, cell := range cells {
		if !cell.Applicable {
			continue
		}
		applicable++
		if !cellComplete(cell) {
			return false
		}
	}
	return applicable > 0
}

func cellComplete(cell certificationCell) bool {
	return cell.Applicable && cell.Declared && cell.Implemented && cell.FixtureTested && cell.LiveTested && len(cell.LiveEvidence) > 0
}

func validateCertificationCell(cell certificationCell) error {
	if strings.TrimSpace(cell.FunctionKind) == "" {
		return errors.New("function_kind is required")
	}
	if !cell.Applicable {
		if cell.NotApplicable == nil {
			return errors.New("not_applicable reason is required when applicable=false")
		}
		if err := validateNotApplicableReason(*cell.NotApplicable); err != nil {
			return fmt.Errorf("not_applicable: %w", err)
		}
		if cell.Declared || cell.Implemented || cell.FixtureTested || cell.LiveTested || len(cell.FixtureEvidence) > 0 || len(cell.LiveEvidence) > 0 {
			return errors.New("not_applicable cell cannot carry affirmative evidence")
		}
		return nil
	}
	if cell.NotApplicable != nil {
		return errors.New("applicable cell cannot carry not_applicable reason")
	}
	if cell.FixtureTested && len(cell.FixtureEvidence) == 0 {
		return errors.New("fixture_tested cell requires fixture_evidence")
	}
	if cell.LiveTested && len(cell.LiveEvidence) == 0 {
		return errors.New("live_tested cell requires live_evidence")
	}
	if !cell.LiveTested && len(cell.LiveEvidence) > 0 {
		return errors.New("live_evidence requires live_tested=true")
	}
	for _, evidence := range cell.LiveEvidence {
		if err := validateEvidencePointer(evidence); err != nil {
			return fmt.Errorf("live_evidence: %w", err)
		}
	}
	return nil
}

func validateEvidencePointer(evidence evidencePointer) error {
	if strings.TrimSpace(evidence.Record) == "" {
		return errors.New("record is required")
	}
	if err := validateEvidenceProvider(evidence.Provider); err != nil {
		return fmt.Errorf("provider: %w", err)
	}
	if _, err := time.Parse(time.RFC3339, evidence.ExecutedAt); err != nil {
		return fmt.Errorf("executed_at must be RFC3339: %w", err)
	}
	if err := validateEvidenceRunID(evidence.RunID); err != nil {
		return fmt.Errorf("run_id: %w", err)
	}
	if err := validateFullParityCredential(evidence.CredentialScope, evidence.CredentialNote); err != nil {
		return fmt.Errorf("credential: %w", err)
	}
	if err := validateEmbeddedEvidenceProof(evidence.Proof); err != nil {
		return fmt.Errorf("proof: %w", err)
	}
	return nil
}

func validateNotApplicableReason(reason notApplicableReason) error {
	code := strings.ToLower(strings.TrimSpace(reason.Code))
	switch code {
	case "", "n/a", "na", "blocked", "not_applicable", "not-applicable":
		return fmt.Errorf("reason code %q is generic", reason.Code)
	}
	if strings.TrimSpace(reason.Reason) == "" {
		return errors.New("reason explanation is required")
	}
	return nil
}

func validateAcceptedEvidence(evidence acceptedEvidence) error {
	if evidence.SchemaVersion != certificationSchemaVersion {
		return fmt.Errorf("schema_version %d is unsupported", evidence.SchemaVersion)
	}
	if evidence.Status != evidenceStatusPassed {
		return fmt.Errorf("status %q is not accepted", evidence.Status)
	}
	if err := validateFullParityCredential(evidence.CredentialScope, evidence.CredentialNote); err != nil {
		return fmt.Errorf("credential: %w", err)
	}
	if err := validateEvidenceProvider(evidence.Provider); err != nil {
		return fmt.Errorf("provider: %w", err)
	}
	if _, err := time.Parse(time.RFC3339, evidence.ExecutedAt); err != nil {
		return fmt.Errorf("executed_at must be RFC3339: %w", err)
	}
	if err := validateEvidenceRunID(evidence.RunID); err != nil {
		return fmt.Errorf("run_id: %w", err)
	}
	if err := validateEmbeddedEvidenceProof(evidence.Proof); err != nil {
		return fmt.Errorf("proof: %w", err)
	}

	switch evidence.Scope {
	case evidenceScopeCapability:
		if strings.TrimSpace(evidence.Connector) == "" || strings.TrimSpace(evidence.FunctionKind) == "" {
			return errors.New("capability evidence requires connector and function_kind")
		}
		if !isSafeProofIdentifier(evidence.Connector) || !isSafeProofIdentifier(evidence.FunctionKind) {
			return errors.New("capability evidence connector and function_kind must be safe identifiers")
		}
		if evidence.Proof.Flow != nil {
			return errors.New("capability evidence cannot carry a flow proof")
		}
	case evidenceScopeFlow:
		if strings.TrimSpace(evidence.Source) == "" || strings.TrimSpace(evidence.Destination) == "" || strings.TrimSpace(evidence.FlowKind) == "" {
			return errors.New("flow evidence requires source, destination, and flow_kind")
		}
		if !isSafeProofIdentifier(evidence.Source) || !isSafeProofIdentifier(evidence.Destination) || !isSafeProofIdentifier(evidence.FlowKind) {
			return errors.New("flow evidence source, destination, and flow_kind must be safe identifiers")
		}
		if evidence.Proof.Flow == nil {
			return errors.New("flow evidence requires an embedded round-trip proof")
		}
	case evidenceScopeWorkflow:
		if strings.TrimSpace(evidence.Connector) == "" || strings.TrimSpace(evidence.WorkflowKind) == "" {
			return errors.New("workflow evidence requires connector and workflow_kind")
		}
		if !isSafeProofIdentifier(evidence.Connector) || !isSafeProofIdentifier(evidence.WorkflowKind) {
			return errors.New("workflow evidence connector and workflow_kind must be safe identifiers")
		}
		if evidence.Proof.Flow != nil {
			return errors.New("workflow evidence cannot carry a flow proof")
		}
	case evidenceScopeSyncMode:
		if strings.TrimSpace(evidence.Connector) == "" || strings.TrimSpace(evidence.SyncMode) == "" || strings.TrimSpace(evidence.Primitive) == "" {
			return errors.New("sync-mode evidence requires connector, sync_mode, and primitive")
		}
		if !isSafeProofIdentifier(evidence.Connector) || !isSafeProofIdentifier(evidence.SyncMode) || !isSafeProofIdentifier(evidence.Primitive) {
			return errors.New("sync-mode evidence connector, sync_mode, and primitive must be safe identifiers")
		}
		if evidence.Proof.Flow != nil {
			return errors.New("sync-mode evidence cannot carry a flow proof")
		}
	default:
		return fmt.Errorf("scope %q is unsupported", evidence.Scope)
	}
	return nil
}

func validateFullParityCredential(scope, note string) error {
	if scope != credentialScopeFullParity {
		return fmt.Errorf("scope %q must be %q", scope, credentialScopeFullParity)
	}
	if note != fullParityCredentialNote {
		return errors.New("note must state the full-parity subset behavior exactly")
	}
	return nil
}

func loadAcceptedEvidence(repoRoot string, scope []string) ([]acceptedEvidence, error) {
	dir := filepath.Join(repoRoot, acceptedEvidenceDirectory)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return []acceptedEvidence{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read accepted evidence directory: %w", err)
	}

	items := make([]acceptedEvidence, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if !isSafeProofIdentifier(strings.TrimSuffix(entry.Name(), ".json")) {
			return nil, fmt.Errorf("accepted evidence %q has an unsafe record name", entry.Name())
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return nil, fmt.Errorf("accepted evidence %q must not be a symlink", entry.Name())
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read accepted evidence %q: %w", entry.Name(), err)
		}
		if err := validateAcceptedEvidenceScopeIdentityKeys(raw); err != nil {
			return nil, fmt.Errorf("parse accepted evidence %q: %w", entry.Name(), err)
		}
		var identity acceptedEvidenceScopeIdentity
		if err := json.Unmarshal(raw, &identity); err != nil {
			return nil, fmt.Errorf("parse accepted evidence %q: %w", entry.Name(), err)
		}
		if len(scope) != 0 && acceptedEvidenceScopeIdentityIsConclusiveNonallowlisted(identity) {
			continue
		}
		var evidence acceptedEvidence
		if err := decodeStrictJSON(raw, &evidence); err != nil {
			return nil, fmt.Errorf("parse accepted evidence %q: %w", entry.Name(), err)
		}
		if err := validateAcceptedEvidence(evidence); err != nil {
			return nil, fmt.Errorf("accepted evidence %q: %w", entry.Name(), err)
		}
		if len(scope) != 0 && !acceptedEvidenceWithinScope(evidence, scope) {
			continue
		}
		relative, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return nil, fmt.Errorf("relativize accepted evidence %q: %w", entry.Name(), err)
		}
		evidence.recordPath = filepath.ToSlash(relative)
		items = append(items, evidence)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].recordPath < items[j].recordPath })
	return items, nil
}

func validateAcceptedEvidenceScopeIdentityKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	opening, ok := token.(json.Delim)
	if !ok || opening != '{' {
		return nil
	}

	seen := make(map[string]struct{}, 4)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		field, ok := token.(string)
		if !ok {
			return errors.New("accepted evidence object key is not a string")
		}
		if canonicalField, identityField := acceptedEvidenceScopeIdentityField(field); identityField {
			if _, found := seen[canonicalField]; found {
				return fmt.Errorf("duplicate accepted evidence identity field %q", canonicalField)
			}
			seen[canonicalField] = struct{}{}
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}

func acceptedEvidenceScopeIdentityField(field string) (string, bool) {
	for _, identityField := range []string{"scope", "connector", "source", "destination"} {
		if bytes.EqualFold([]byte(field), []byte(identityField)) {
			return identityField, true
		}
	}
	return "", false
}

func acceptedEvidenceWithinScope(evidence acceptedEvidence, scope []string) bool {
	return acceptedEvidenceScopeIdentityWithinScope(acceptedEvidenceScopeIdentity{
		Scope:       evidence.Scope,
		Connector:   evidence.Connector,
		Source:      evidence.Source,
		Destination: evidence.Destination,
	}, scope)
}

func acceptedEvidenceScopeIdentityWithinScope(identity acceptedEvidenceScopeIdentity, scope []string) bool {
	inScope := make(map[string]bool, len(scope))
	for _, name := range scope {
		inScope[name] = true
	}
	switch identity.Scope {
	case evidenceScopeCapability, evidenceScopeWorkflow, evidenceScopeSyncMode:
		return inScope[identity.Connector]
	case evidenceScopeFlow:
		return inScope[identity.Source] && certificationConnectorAllowed(identity.Destination)
	default:
		return false
	}
}

func acceptedEvidenceScopeIdentityIsConclusiveNonallowlisted(identity acceptedEvidenceScopeIdentity) bool {
	if certificationConnectorAllowed(identity.Connector) || certificationConnectorAllowed(identity.Source) || certificationConnectorAllowed(identity.Destination) {
		return false
	}
	switch identity.Scope {
	case evidenceScopeCapability, evidenceScopeWorkflow, evidenceScopeSyncMode:
		return identity.Connector != ""
	case evidenceScopeFlow:
		return identity.Source != "" && identity.Destination != ""
	default:
		return false
	}
}

func deriveCapabilityBaseline(connectorsIn []capabilityConnector, kinds []functionKind) capabilityBaseline {
	baseline := capabilityBaseline{
		Connectors: len(connectorsIn),
		PerKind:    make([]kindBaseline, 0, len(kinds)),
	}
	for _, connector := range connectorsIn {
		if connector.CapabilityComplete {
			baseline.CapabilityComplete++
		}
	}
	for _, kind := range kinds {
		totals := kindBaseline{FunctionKind: kind.ID, Connectors: len(connectorsIn)}
		for _, connector := range connectorsIn {
			cell, ok := capabilityCellForConnector(connector, kind.ID)
			if !ok || !cell.Applicable {
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
			if cellComplete(cell) {
				totals.Complete++
			}
		}
		baseline.PerKind = append(baseline.PerKind, totals)
	}
	return baseline
}

func capabilityCellFor(matrix capabilityMatrix, connectorName, kind string) (certificationCell, bool) {
	for _, connector := range matrix.Connectors {
		if connector.Name != connectorName {
			continue
		}
		return capabilityCellForConnector(connector, kind)
	}
	return certificationCell{}, false
}

func capabilityCellForConnector(connector capabilityConnector, kind string) (certificationCell, bool) {
	for _, cell := range connector.Cells {
		if cell.FunctionKind == kind {
			return cell, true
		}
	}
	return certificationCell{}, false
}

func discoverLegacyCertificationInputs(repoRoot string) (legacyCertificationInventory, error) {
	files := make([]legacyCertificationFile, 0)
	err := filepath.WalkDir(repoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), "certification.json") {
			return nil
		}
		relative, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		classification := "bundle_contract"
		if strings.Contains(filepath.ToSlash(relative), "/fixtures/") || strings.Contains(filepath.ToSlash(relative), "/schemas/") {
			classification = "fixture_or_schema"
		}
		files = append(files, legacyCertificationFile{
			Path:           filepath.ToSlash(relative),
			Classification: classification,
		})
		return nil
	})
	if err != nil {
		return legacyCertificationInventory{}, fmt.Errorf("scan legacy certification inputs: %w", err)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return legacyCertificationInventory{Ignored: true, Files: files}, nil
}

func discoverFunctionKinds(repoRoot string) ([]functionKind, error) {
	capabilitySources := []string{
		filepath.Join(repoRoot, "internal", "connectors", "connectors.go"),
		filepath.Join(repoRoot, "internal", "connectors", "engine", "bundle.go"),
	}
	capabilities := map[string]string{}
	for _, path := range capabilitySources {
		found, err := capabilityFieldsFromSource(repoRoot, path)
		if err != nil {
			return nil, err
		}
		for name, source := range found {
			if _, exists := capabilities[name]; !exists {
				capabilities[name] = source
			}
		}
	}

	operationSource := filepath.Join(repoRoot, "internal", "connectors", "engine", "bundle.go")
	operations, err := operationKindsFromSource(repoRoot, operationSource)
	if err != nil {
		return nil, err
	}
	executors, err := operationExecutorAnnotations(repoRoot)
	if err != nil {
		return nil, err
	}

	kinds := make([]functionKind, 0, len(capabilities)+len(operations))
	for name, source := range capabilities {
		kinds = append(kinds, functionKind{
			ID:              "capability:" + name,
			Category:        "capability",
			Name:            name,
			DiscoverySource: source,
		})
	}
	for name, source := range operations {
		kinds = append(kinds, functionKind{
			ID:              "operation:" + name,
			Category:        "operation",
			Name:            name,
			DiscoverySource: source,
			ExecutorSource:  executors[name],
		})
	}
	sort.Slice(kinds, func(i, j int) bool { return kinds[i].ID < kinds[j].ID })

	operationNames := map[string]bool{}
	for name := range operations {
		operationNames[name] = true
	}
	for name, source := range executors {
		if !operationNames[name] {
			return nil, fmt.Errorf("executor annotation %q names unknown operation kind %q", source, name)
		}
	}
	return kinds, nil
}

func capabilityFieldsFromSource(repoRoot, path string) (map[string]string, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse capability source %q: %w", path, err)
	}
	found := map[string]string{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "Capabilities" {
				continue
			}
			structure, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structure.Fields.List {
				for _, name := range field.Names {
					found[snakeIdentifier(name.Name)] = sourceSymbol(repoRoot, path, typeSpec.Name.Name+"."+name.Name)
				}
			}
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("capability source %q declares no Capabilities fields", path)
	}
	return found, nil
}

func operationKindsFromSource(repoRoot, path string) (map[string]string, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse operation source %q: %w", path, err)
	}
	found := map[string]string{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "expectedOperationBlock" {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			switchStmt, ok := node.(*ast.SwitchStmt)
			if !ok {
				return true
			}
			ident, ok := switchStmt.Tag.(*ast.Ident)
			if !ok || ident.Name != "kind" {
				return true
			}
			for _, clauseNode := range switchStmt.Body.List {
				clause, ok := clauseNode.(*ast.CaseClause)
				if !ok {
					continue
				}
				for _, expr := range clause.List {
					literal, ok := expr.(*ast.BasicLit)
					if !ok || literal.Kind != token.STRING {
						continue
					}
					value, err := strconv.Unquote(literal.Value)
					if err == nil {
						// The literal itself is not a source construct. The enclosing
						// function plus its operation discriminator is stable across
						// line insertions and distinguishes sibling case clauses.
						found[value] = sourceSymbol(repoRoot, path, fn.Name.Name+"(kind="+value+")")
					}
				}
			}
			return false
		})
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("operation source %q declares no expectedOperationBlock cases", path)
	}
	return found, nil
}

func operationExecutorAnnotations(repoRoot string) (map[string]string, error) {
	dir := filepath.Join(repoRoot, "internal", "connectors", "engine")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read engine source directory: %w", err)
	}
	annotations := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse engine source %q: %w", path, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			for _, kind := range executorKindsFromDoc(fn.Doc) {
				source := sourceSymbol(repoRoot, path, functionSymbol(fn))
				if existing, found := annotations[kind]; found && existing != source {
					return nil, fmt.Errorf("operation kind %q has multiple executor annotations (%s, %s)", kind, existing, source)
				}
				annotations[kind] = source
			}
		}
	}
	return annotations, nil
}

func executorKindsFromDoc(doc *ast.CommentGroup) []string {
	if doc == nil {
		return nil
	}
	var kinds []string
	for _, line := range strings.Split(doc.Text(), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, executorAnnotationPrefix) {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(line, executorAnnotationPrefix))
		for _, kind := range strings.Split(raw, ",") {
			kind = strings.TrimSpace(kind)
			if kind != "" {
				kinds = append(kinds, kind)
			}
		}
	}
	return kinds
}

func boolFieldForKind(value any, kind string) (bool, bool) {
	fieldName := goIdentifier(kind)
	reflected := reflect.ValueOf(value)
	for reflected.Kind() == reflect.Pointer {
		if reflected.IsNil() {
			return false, false
		}
		reflected = reflected.Elem()
	}
	field := reflected.FieldByName(fieldName)
	if !field.IsValid() || field.Kind() != reflect.Bool {
		return false, false
	}
	return field.Bool(), true
}

func goIdentifier(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		if strings.EqualFold(part, "cdc") {
			builder.WriteString("CDC")
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		builder.WriteString(part[1:])
	}
	return builder.String()
}

func snakeIdentifier(value string) string {
	var builder strings.Builder
	runes := []rune(value)
	for i, current := range runes {
		if i > 0 && isUpper(current) && (!isUpper(runes[i-1]) || (i+1 < len(runes) && !isUpper(runes[i+1]))) {
			builder.WriteByte('_')
		}
		builder.WriteRune(toLower(current))
	}
	return builder.String()
}

func normalizeIdentifier(value string) string {
	return strings.ReplaceAll(snakeIdentifier(value), "_", "")
}

func identifierSharesMeaning(method, capability string) bool {
	methodWords := strings.Split(snakeIdentifier(method), "_")
	capabilityWords := strings.Split(snakeIdentifier(capability), "_")
	for _, capabilityWord := range capabilityWords {
		if len(capabilityWord) < 4 {
			continue
		}
		for _, methodWord := range methodWords {
			if methodWord == capabilityWord {
				return true
			}
		}
	}
	return false
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func toLower(r rune) rune {
	if isUpper(r) {
		return r + ('a' - 'A')
	}
	return r
}

func sourceSymbol(repoRoot, path, symbol string) string {
	relative, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return filepath.ToSlash(path) + ":" + symbol
	}
	return filepath.ToSlash(relative) + ":" + symbol
}

func functionSymbol(fn *ast.FuncDecl) string {
	if receiver := receiverTypeName(fn); receiver != "" {
		return receiver + "." + fn.Name.Name
	}
	return fn.Name.Name
}

func validateSymbolSourceAnchor(anchor string) error {
	path, symbol, found := strings.Cut(anchor, ":")
	if !found || filepath.IsAbs(path) || strings.HasPrefix(filepath.ToSlash(path), "../") || !strings.HasSuffix(path, ".go") || strings.TrimSpace(symbol) == "" {
		return fmt.Errorf("source anchor %q must be relative/path.go:Symbol", anchor)
	}
	if _, err := strconv.Atoi(symbol); err == nil {
		return fmt.Errorf("source anchor %q uses a line number rather than a symbol", anchor)
	}
	if strings.HasPrefix(symbol, "expectedOperationBlock(kind=") && strings.HasSuffix(symbol, ")") {
		kind := strings.TrimSuffix(strings.TrimPrefix(symbol, "expectedOperationBlock(kind="), ")")
		if isSafeProofIdentifier(kind) {
			return nil
		}
	}
	parts := strings.Split(symbol, ".")
	if len(parts) > 2 {
		return fmt.Errorf("source anchor %q has an ambiguous symbol", anchor)
	}
	for _, part := range parts {
		if !isSourceIdentifier(part) {
			return fmt.Errorf("source anchor %q has an invalid symbol", anchor)
		}
	}
	return nil
}

func isSourceIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for index, r := range value {
		if r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			continue
		}
		if index > 0 && r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}

func marshalGeneratedJSON(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

func writeGeneratedArtifact(path string, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return err
	}
	return nil
}

func checkGeneratedArtifact(path string, generated []byte) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("generated artifact %q is missing; run `go run ./cmd/connectorgen certification-matrix`", filepath.ToSlash(path))
		}
		return fmt.Errorf("read generated artifact %q: %w", filepath.ToSlash(path), err)
	}
	if err := validateCapabilityMatrixArtifactJSON(existing); err != nil {
		return fmt.Errorf("generated certification artifact %q is invalid: %w", filepath.ToSlash(path), err)
	}
	if !bytes.Equal(existing, generated) {
		return fmt.Errorf("generated artifact %q has drift; run `go run ./cmd/connectorgen certification-matrix`", filepath.ToSlash(path))
	}
	return nil
}

// validateCapabilityMatrixArtifactJSON validates the committed, proof-bearing
// record independently of source generation. The later byte comparison is the
// code cross-check; it is intentionally not a substitute for this check.
func validateCapabilityMatrixArtifactJSON(raw []byte) error {
	var matrix capabilityMatrix
	if err := decodeStrictJSON(raw, &matrix); err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if matrix.SchemaVersion != certificationSchemaVersion {
		return fmt.Errorf("schema_version %d is unsupported", matrix.SchemaVersion)
	}
	return validateCapabilityMatrix(matrix)
}

func validateCapabilityMatrix(matrix capabilityMatrix) error {
	if matrix.SchemaVersion != 0 && matrix.SchemaVersion != certificationSchemaVersion {
		return fmt.Errorf("schema_version %d is unsupported", matrix.SchemaVersion)
	}
	seenKinds := make(map[string]bool, len(matrix.FunctionKinds))
	for _, kind := range matrix.FunctionKinds {
		if strings.TrimSpace(kind.ID) == "" {
			return errors.New("function_kinds contains an empty id")
		}
		if seenKinds[kind.ID] {
			return fmt.Errorf("function_kinds duplicates %q", kind.ID)
		}
		if kind.DiscoverySource != "" {
			if err := validateSymbolSourceAnchor(kind.DiscoverySource); err != nil {
				return fmt.Errorf("function kind %q discovery_source: %w", kind.ID, err)
			}
		}
		if kind.ExecutorSource != "" {
			if err := validateSymbolSourceAnchor(kind.ExecutorSource); err != nil {
				return fmt.Errorf("function kind %q executor_source: %w", kind.ID, err)
			}
		}
		seenKinds[kind.ID] = true
	}
	seenConnectors := make(map[string]bool, len(matrix.Connectors))
	for _, connector := range matrix.Connectors {
		if strings.TrimSpace(connector.Name) == "" {
			return errors.New("connectors contains an empty name")
		}
		if seenConnectors[connector.Name] {
			return fmt.Errorf("connectors duplicates %q", connector.Name)
		}
		seenConnectors[connector.Name] = true
		if len(connector.Cells) != len(matrix.FunctionKinds) {
			return fmt.Errorf("connector %q has %d cells for %d function kinds", connector.Name, len(connector.Cells), len(matrix.FunctionKinds))
		}
		seenCells := make(map[string]bool, len(connector.Cells))
		for _, cell := range connector.Cells {
			if !seenKinds[cell.FunctionKind] {
				return fmt.Errorf("connector %q has unknown function_kind %q", connector.Name, cell.FunctionKind)
			}
			if seenCells[cell.FunctionKind] {
				return fmt.Errorf("connector %q duplicates function_kind %q", connector.Name, cell.FunctionKind)
			}
			seenCells[cell.FunctionKind] = true
			if err := validateCertificationCell(cell); err != nil {
				return fmt.Errorf("connector %q function_kind %q: %w", connector.Name, cell.FunctionKind, err)
			}
		}
	}
	return nil
}
