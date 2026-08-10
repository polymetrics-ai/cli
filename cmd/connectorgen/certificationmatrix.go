package main

import (
	"bytes"
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

// evidencePointer repeats the accepted record's publishable proof inside the
// generated matrix. Transcript values are already repository-salted
// fingerprints, so a matrix can be read and checked without loading an
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

// capabilityConnector is the per-connector portion of the capability matrix.
// CapabilityComplete intentionally does not claim final connector
// certification; flow-matrix generation adds pair requirements later.
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

// capabilityMatrix is the generated artifact. Fields are deliberately slices,
// not maps, so json.MarshalIndent produces stable, reviewable ordering.
type capabilityMatrix struct {
	SchemaVersion             int                          `json:"schema_version"`
	GeneratedCommand          string                       `json:"generated_command"`
	FunctionKinds             []functionKind               `json:"function_kinds"`
	Connectors                []capabilityConnector        `json:"connectors"`
	LegacyCertificationInputs legacyCertificationInventory `json:"legacy_certification_inputs"`
	Baseline                  capabilityBaseline           `json:"baseline"`
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

type matrixConnectorSource struct {
	name            string
	integrationType string
	bundle          *engine.Bundle
	connector       connectors.Connector
	conformance     *conformance.Report
}

// runCertificationMatrix implements the source-controlled generation and
// byte-for-byte drift gate. It deliberately never runs a provider request.
func runCertificationMatrix(args []string, stdout, stderr io.Writer) int {
	root := "."
	check := false
	for _, arg := range args[1:] {
		switch {
		case arg == "--check":
			check = true
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

	absRoot, err := filepath.Abs(root)
	if err != nil {
		logf(stderr, "connectorgen certification-matrix: resolve repository root: %v\n", err)
		return 1
	}
	path := filepath.Join(absRoot, capabilityMatrixPath)
	flowPath := filepath.Join(absRoot, flowMatrixPath)
	statusPath := filepath.Join(absRoot, certificationStatusPath)
	// A certification checker deliberately reads and validates the committed
	// record before it inspects current source. This ordering makes a malformed
	// or proofless certificate an error in its own right rather than something
	// code reachability could paper over.
	if check {
		if err := validateCapabilityMatrixArtifactFile(path); err != nil {
			logf(stderr, "connectorgen certification-matrix: %v\n", err)
			return 1
		}
		if err := validateFlowMatrixArtifactFile(flowPath); err != nil {
			logf(stderr, "connectorgen certification-matrix: %v\n", err)
			return 1
		}
		if err := validateCertificationStatusArtifactFile(statusPath); err != nil {
			logf(stderr, "connectorgen certification-matrix: %v\n", err)
			return 1
		}
	}
	matrix, err := buildCapabilityMatrix(absRoot)
	if err != nil {
		logf(stderr, "connectorgen certification-matrix: %v\n", err)
		return 1
	}
	flows, err := buildFlowMatrix(absRoot, matrix)
	if err != nil {
		logf(stderr, "connectorgen certification-matrix: build flow matrix: %v\n", err)
		return 1
	}
	payload, err := marshalGeneratedJSON(matrix)
	if err != nil {
		logf(stderr, "connectorgen certification-matrix: render capability matrix: %v\n", err)
		return 1
	}
	flowPayload, err := marshalGeneratedJSON(flows)
	if err != nil {
		logf(stderr, "connectorgen certification-matrix: render flow matrix: %v\n", err)
		return 1
	}
	statusPayload, err := marshalGeneratedJSON(buildCertificationStatusArtifact(flows))
	if err != nil {
		logf(stderr, "connectorgen certification-matrix: render certification status: %v\n", err)
		return 1
	}
	if check {
		if err := checkGeneratedArtifact(path, payload); err != nil {
			logf(stderr, "connectorgen certification-matrix: %v\n", err)
			return 1
		}
		if err := checkFlowGeneratedArtifact(flowPath, flowPayload); err != nil {
			logf(stderr, "connectorgen certification-matrix: %v\n", err)
			return 1
		}
		if err := checkCertificationStatusGeneratedArtifact(statusPath, statusPayload); err != nil {
			logf(stderr, "connectorgen certification-matrix: %v\n", err)
			return 1
		}
		logf(stdout, "certification matrices are current: connectors=%d capability_complete=%d certified=%d\n", matrix.Baseline.Connectors, matrix.Baseline.CapabilityComplete, flows.Baseline.Certified)
		return 0
	}
	if err := writeGeneratedArtifact(path, payload); err != nil {
		logf(stderr, "connectorgen certification-matrix: write capability matrix: %v\n", err)
		return 1
	}
	if err := writeGeneratedArtifact(flowPath, flowPayload); err != nil {
		logf(stderr, "connectorgen certification-matrix: write flow matrix: %v\n", err)
		return 1
	}
	if err := writeGeneratedArtifact(statusPath, statusPayload); err != nil {
		logf(stderr, "connectorgen certification-matrix: write certification status: %v\n", err)
		return 1
	}
	logf(stdout, "generated certification matrices: connectors=%d capability_complete=%d certified=%d\n", matrix.Baseline.Connectors, matrix.Baseline.CapabilityComplete, flows.Baseline.Certified)
	return 0
}

// buildCapabilityMatrix derives every matrix fact from source, registered
// runtime types, recorded fixtures, or an accepted evidence record. It is
// intentionally exported only inside connectorgen to keep the developer tool
// as the sole owner of its generated artifact format.
func buildCapabilityMatrix(repoRoot string) (capabilityMatrix, error) {
	kinds, err := discoverFunctionKinds(repoRoot)
	if err != nil {
		return capabilityMatrix{}, err
	}
	bundles, err := loadSourceBundles(repoRoot)
	if err != nil {
		return capabilityMatrix{}, err
	}
	evidence, err := loadAcceptedEvidence(repoRoot)
	if err != nil {
		return capabilityMatrix{}, err
	}
	sources, err := matrixConnectorSources(bundles)
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

func matrixConnectorSources(bundles []engine.Bundle) ([]matrixConnectorSource, error) {
	bundleByName := make(map[string]*engine.Bundle, len(bundles))
	for i := range bundles {
		bundleByName[bundles[i].Name] = &bundles[i]
	}

	registry := bundleregistry.New()
	names := make(map[string]bool, len(bundleByName)+len(registry.List()))
	for name := range bundleByName {
		names[name] = true
	}
	for _, metadata := range registry.List() {
		names[metadata.Name] = true
	}

	sortedNames := make([]string, 0, len(names))
	for name := range names {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	out := make([]matrixConnectorSource, 0, len(sortedNames))
	for _, name := range sortedNames {
		bundle := bundleByName[name]
		connector, registered := registry.Get(name)
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

	implemented, err := functionKindImplemented(repoRoot, source, kind)
	if err != nil {
		return certificationCell{}, err
	}
	fixtureTested, fixtureEvidence := functionKindFixtureTested(source, kind)
	liveEvidence := matchingCapabilityEvidence(evidence, source.name, kind.ID)
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

func functionKindImplemented(repoRoot string, source matrixConnectorSource, kind functionKind) (bool, error) {
	switch kind.Category {
	case "operation":
		return kind.ExecutorSource != "", nil
	case "capability":
		return capabilityImplemented(repoRoot, source, kind.Name)
	}
	return false, nil
}

func capabilityImplemented(repoRoot string, source matrixConnectorSource, capability string) (bool, error) {
	if source.connector == nil {
		return false, nil
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
// method's source. It avoids calling Write/Read while generating a matrix and
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
	if !strings.HasPrefix(typ.PkgPath(), modulePrefix) {
		return false, nil
	}
	dir := filepath.Join(repoRoot, strings.TrimPrefix(typ.PkgPath(), modulePrefix))
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

func matchingCapabilityEvidence(evidence []acceptedEvidence, connectorName, kind string) []evidencePointer {
	matched := make([]evidencePointer, 0)
	for _, item := range evidence {
		if item.Scope != evidenceScopeCapability || item.Status != evidenceStatusPassed || item.Connector != connectorName || item.FunctionKind != kind {
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

func capabilityComplete(cells []certificationCell) bool {
	return certificationComplete(cells)
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
	if strings.TrimSpace(evidence.Provider) == "" {
		return errors.New("provider is required")
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
		if evidence.Proof.Flow != nil {
			return errors.New("capability evidence cannot carry a flow proof")
		}
	case evidenceScopeFlow:
		if strings.TrimSpace(evidence.Source) == "" || strings.TrimSpace(evidence.Destination) == "" || strings.TrimSpace(evidence.FlowKind) == "" {
			return errors.New("flow evidence requires source, destination, and flow_kind")
		}
		if evidence.Proof.Flow == nil {
			return errors.New("flow evidence requires an embedded round-trip proof")
		}
	case evidenceScopeWorkflow:
		if strings.TrimSpace(evidence.Connector) == "" || strings.TrimSpace(evidence.WorkflowKind) == "" {
			return errors.New("workflow evidence requires connector and workflow_kind")
		}
		if evidence.Proof.Flow != nil {
			return errors.New("workflow evidence cannot carry a flow proof")
		}
	case evidenceScopeSyncMode:
		if strings.TrimSpace(evidence.Connector) == "" || strings.TrimSpace(evidence.SyncMode) == "" || strings.TrimSpace(evidence.Primitive) == "" {
			return errors.New("sync-mode evidence requires connector, sync_mode, and primitive")
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

func loadAcceptedEvidence(repoRoot string) ([]acceptedEvidence, error) {
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
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".json") == false {
			continue
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return nil, fmt.Errorf("accepted evidence %q must not be a symlink", entry.Name())
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read accepted evidence %q: %w", entry.Name(), err)
		}
		var evidence acceptedEvidence
		if err := decodeStrictJSON(raw, &evidence); err != nil {
			return nil, fmt.Errorf("parse accepted evidence %q: %w", entry.Name(), err)
		}
		if err := validateAcceptedEvidence(evidence); err != nil {
			return nil, fmt.Errorf("accepted evidence %q: %w", entry.Name(), err)
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
					found[snakeIdentifier(name.Name)] = sourceAt(repoRoot, path, fileSet.Position(name.Pos()).Line)
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
						found[value] = sourceAt(repoRoot, path, fileSet.Position(literal.Pos()).Line)
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
				source := sourceAt(repoRoot, path, fileSet.Position(fn.Pos()).Line)
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

func sourceAt(repoRoot, path string, line int) string {
	relative, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return fmt.Sprintf("%s:%d", filepath.ToSlash(relative), line)
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

func validateCapabilityMatrixArtifactFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("generated artifact %q is missing; run `go run ./cmd/connectorgen certification-matrix`", filepath.ToSlash(path))
		}
		return fmt.Errorf("read generated artifact %q: %w", filepath.ToSlash(path), err)
	}
	if err := validateCapabilityMatrixArtifactJSON(raw); err != nil {
		return fmt.Errorf("generated certification artifact %q is invalid: %w", filepath.ToSlash(path), err)
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
	seenKinds := make(map[string]bool, len(matrix.FunctionKinds))
	for _, kind := range matrix.FunctionKinds {
		if strings.TrimSpace(kind.ID) == "" {
			return errors.New("function_kinds contains an empty id")
		}
		if seenKinds[kind.ID] {
			return fmt.Errorf("function_kinds duplicates %q", kind.ID)
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
