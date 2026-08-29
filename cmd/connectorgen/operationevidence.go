package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/conformance"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	operationEvidenceSchemaVersion = 1
	operationEvidenceArtifactPath  = "internal/connectors/operation-evidence.json"
	operationEvidenceFixed100Path  = "internal/connectors/operation-evidence-fixed-100.json"

	operationEvidenceGapSourceTrace         = "source_trace"
	operationEvidenceGapCanonicalMapping    = "canonical_mapping"
	operationEvidenceGapRuntimeReachability = "runtime_reachability"
	operationEvidenceGapCLICommand          = "cli_command"
	operationEvidenceGapWebsiteRow          = "website_row"
	operationEvidenceGapFixtureProof        = "fixture_proof"
	operationEvidenceGapConformanceProof    = "conformance_proof"
	operationEvidenceGapReadOnlyConflict    = "read_only_conflict"

	operationEvidenceClassETL            = "etl"
	operationEvidenceClassReverseETL     = "reverse_etl"
	operationEvidenceClassDirectRead     = "direct_read"
	operationEvidenceClassDirectWrite    = "direct_write"
	operationEvidenceClassBinaryDownload = "binary_download"
	operationEvidenceClassBinaryUpload   = "binary_upload"
)

var operationEvidenceClasses = []string{
	operationEvidenceClassETL,
	operationEvidenceClassReverseETL,
	operationEvidenceClassDirectRead,
	operationEvidenceClassDirectWrite,
	operationEvidenceClassBinaryDownload,
	operationEvidenceClassBinaryUpload,
}

var operationEvidenceGraphQLRootField = regexp.MustCompile(`(?s)\{\s*([_A-Za-z][_0-9A-Za-z]*)`)

// operationEvidenceArtifact is the deterministic, source-first accounting
// artifact for every operation enumerated by a connector-owned source lock.
// It intentionally stores absence as provider evidence, not as a local N/A
// classification, so callers cannot silently suppress an unenumerable source.
type operationEvidenceArtifact struct {
	SchemaVersion         int                               `json:"schema_version"`
	GeneratedCommand      string                            `json:"generated_command"`
	Provenance            operationEvidenceProvenance       `json:"provenance"`
	Rows                  []operationEvidenceRow            `json:"rows"`
	MissingFoundations    []operationEvidenceRollup         `json:"missing_foundations"`
	IntentionallyReadOnly []operationEvidenceReadOnlyRollup `json:"intentionally_read_only,omitempty"`
}

// operationEvidenceProvenance binds every row set to the same source
// projection and relevant configuration identity used by certification. The
// source lock itself remains the operation-level provider trace.
type operationEvidenceProvenance struct {
	SourceProjectionSHA256 string `json:"source_projection_sha256"`
	RelevantConfigSHA256   string `json:"relevant_config_sha256"`
}

type operationEvidenceRow struct {
	Connector       string                                     `json:"connector"`
	SourceID        string                                     `json:"source_id"`
	Protocol        string                                     `json:"protocol"`
	Method          string                                     `json:"method,omitempty"`
	Path            string                                     `json:"path,omitempty"`
	Source          operationEvidenceSourceTrace               `json:"source"`
	Canonical       operationEvidenceCanonical                 `json:"canonical"`
	Runtime         operationEvidenceRuntime                   `json:"runtime"`
	CLI             operationEvidenceCommands                  `json:"cli"`
	Website         operationEvidenceCommands                  `json:"website"`
	Fixtures        operationEvidenceFixtures                  `json:"fixtures"`
	Conformance     operationEvidenceConformance               `json:"conformance"`
	Classifications map[string]operationEvidenceClassification `json:"classifications"`
	Foundations     []operationEvidenceFoundation              `json:"foundations"`
	ReadOnly        *operationEvidenceReadOnly                 `json:"read_only,omitempty"`
	Gaps            []operationEvidenceGap                     `json:"gaps"`
	Absence         *operationEvidenceAbsence                  `json:"absence"`
}

type operationEvidenceSourceTrace struct {
	Lock     string `json:"lock"`
	URL      string `json:"url"`
	SHA256   string `json:"sha256"`
	Bytes    int64  `json:"bytes"`
	Location string `json:"location"`
}

type operationEvidenceCanonical struct {
	State  string `json:"state"`
	Method string `json:"method,omitempty"`
	Path   string `json:"path,omitempty"`
}

type operationEvidenceRuntime struct {
	Enabled bool     `json:"enabled"`
	Targets []string `json:"targets"`
}

type operationEvidenceCommands struct {
	Paths []string `json:"paths"`
}

type operationEvidenceFixtures struct {
	Paths []string `json:"paths"`
}

type operationEvidenceConformance struct {
	Passed bool   `json:"passed"`
	Proof  string `json:"proof"`
}

type operationEvidenceClassification struct {
	Declared bool `json:"declared"`
	Enabled  bool `json:"enabled"`
}

type operationEvidenceFoundation struct {
	ID       string `json:"id"`
	Evidence string `json:"evidence"`
}

type operationEvidenceGap struct {
	Kind       string `json:"kind"`
	Foundation string `json:"foundation,omitempty"`
	Evidence   string `json:"evidence"`
}

type operationEvidenceAbsence struct {
	State    string `json:"state"`
	Reason   string `json:"reason"`
	Evidence string `json:"evidence"`
}

type operationEvidenceRollup struct {
	ID        string   `json:"id"`
	Evidence  string   `json:"evidence"`
	SourceIDs []string `json:"source_ids"`
}

type operationEvidenceReadOnly struct {
	Policy string `json:"policy"`
	Reason string `json:"reason"`
}

type operationEvidenceReadOnlyRollup struct {
	Connector string   `json:"connector"`
	Policy    string   `json:"policy"`
	SourceIDs []string `json:"source_ids"`
}

type operationEvidenceFixed100 struct {
	SchemaVersion int                         `json:"schema_version"`
	Source        string                      `json:"source"`
	Rows          []operationEvidenceFixedRow `json:"rows"`
}

type operationEvidenceFixedRow struct {
	SourceID        string   `json:"source_id"`
	SourceSHA256    string   `json:"source_sha256"`
	CanonicalMethod string   `json:"canonical_method"`
	CanonicalPath   string   `json:"canonical_path"`
	CLIPaths        []string `json:"cli_paths"`
	WebsitePaths    []string `json:"website_paths"`
	FixturePaths    []string `json:"fixture_paths"`
	Classifications []string `json:"classifications"`
}

type operationEvidenceOptions struct {
	repoRoot      string
	output        string
	fixed100      string
	check         bool
	writeFixed100 bool
}

// runOperationEvidence generates or verifies the operation evidence artifact.
// It consumes source locks, crosswalks, and dispositions read-only; their
// schema/parser remains owned by source-import.
func runOperationEvidence(args []string, stdout, stderr io.Writer) int {
	options, err := parseOperationEvidenceOptions(args)
	if err != nil {
		logf(stderr, "connectorgen operation-evidence: %v\n", err)
		return 2
	}
	artifact, err := buildOperationEvidence(options.repoRoot)
	if err != nil {
		logf(stderr, "connectorgen operation-evidence: %v\n", err)
		return 1
	}
	raw, err := marshalGeneratedJSON(artifact)
	if err != nil {
		logf(stderr, "connectorgen operation-evidence: encode artifact: %v\n", err)
		return 1
	}
	if options.writeFixed100 {
		fixed, err := buildOperationEvidenceFixed100(artifact)
		if err != nil {
			logf(stderr, "connectorgen operation-evidence: build fixed-100 reference: %v\n", err)
			return 1
		}
		fixedRaw, err := marshalGeneratedJSON(fixed)
		if err != nil {
			logf(stderr, "connectorgen operation-evidence: encode fixed-100 reference: %v\n", err)
			return 1
		}
		if err := writeGeneratedArtifact(options.fixed100, fixedRaw); err != nil {
			logf(stderr, "connectorgen operation-evidence: write fixed-100 reference: %v\n", err)
			return 1
		}
	}
	if options.check {
		current, err := os.ReadFile(options.output)
		if err != nil {
			logf(stderr, "connectorgen operation-evidence: generated artifact %q is missing; run `go run ./cmd/connectorgen operation-evidence --write-fixed-100`\n", filepath.ToSlash(options.output))
			return 1
		}
		if !bytes.Equal(current, raw) {
			logf(stderr, "connectorgen operation-evidence: generated artifact %q has drift; rerun `go run ./cmd/connectorgen operation-evidence --write-fixed-100`\n", filepath.ToSlash(options.output))
			return 1
		}
		fixed, err := readOperationEvidenceFixed100(options.fixed100)
		if err != nil {
			logf(stderr, "connectorgen operation-evidence: read fixed-100 reference: %v\n", err)
			return 1
		}
		if err := validateOperationEvidenceFixed100(artifact, fixed); err != nil {
			logf(stderr, "connectorgen operation-evidence: fixed-100 validation failed: %v\n", err)
			return 1
		}
		logf(stdout, "connectorgen operation-evidence: %s is current (%d rows; %d rollups; fixed-100 passed)\n", filepath.ToSlash(options.output), len(artifact.Rows), len(artifact.MissingFoundations))
		return 0
	}
	if err := writeGeneratedArtifact(options.output, raw); err != nil {
		logf(stderr, "connectorgen operation-evidence: write artifact: %v\n", err)
		return 1
	}
	logf(stdout, "connectorgen operation-evidence: wrote %s (%d rows; %d rollups)\n", filepath.ToSlash(options.output), len(artifact.Rows), len(artifact.MissingFoundations))
	return 0
}

func parseOperationEvidenceOptions(args []string) (operationEvidenceOptions, error) {
	options := operationEvidenceOptions{}
	for index := 1; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--output":
			if index+1 >= len(args) {
				return operationEvidenceOptions{}, errors.New("--output requires a path")
			}
			index++
			options.output = args[index]
		case "--fixed-100":
			if index+1 >= len(args) {
				return operationEvidenceOptions{}, errors.New("--fixed-100 requires a path")
			}
			index++
			options.fixed100 = args[index]
		case "--check":
			options.check = true
		case "--write-fixed-100":
			options.writeFixed100 = true
		default:
			if strings.HasPrefix(arg, "-") || options.repoRoot != "" {
				return operationEvidenceOptions{}, fmt.Errorf("unexpected argument %q", arg)
			}
			options.repoRoot = arg
		}
	}
	if options.repoRoot == "" {
		root, err := repoRoot()
		if err != nil {
			return operationEvidenceOptions{}, fmt.Errorf("resolve repository root: %w", err)
		}
		options.repoRoot = root
	}
	root, err := filepath.Abs(options.repoRoot)
	if err != nil {
		return operationEvidenceOptions{}, fmt.Errorf("resolve repository root: %w", err)
	}
	options.repoRoot = root
	if options.output == "" {
		options.output = filepath.Join(root, filepath.FromSlash(operationEvidenceArtifactPath))
	} else if !filepath.IsAbs(options.output) {
		options.output = filepath.Join(root, options.output)
	}
	if options.fixed100 == "" {
		options.fixed100 = filepath.Join(root, filepath.FromSlash(operationEvidenceFixed100Path))
	} else if !filepath.IsAbs(options.fixed100) {
		options.fixed100 = filepath.Join(root, options.fixed100)
	}
	return options, nil
}

func buildOperationEvidence(root string) (operationEvidenceArtifact, error) {
	defsRoot := filepath.Join(root, "internal", "connectors", "defs")
	provenance, err := readOperationEvidenceProvenance(root)
	if err != nil {
		return operationEvidenceArtifact{}, err
	}
	websiteRows, err := readOperationEvidenceWebsiteRows(filepath.Join(root, "website", "data", "connectors.generated.json"))
	if err != nil {
		return operationEvidenceArtifact{}, err
	}
	entries, err := os.ReadDir(defsRoot)
	if err != nil {
		return operationEvidenceArtifact{}, fmt.Errorf("read connector definitions: %w", err)
	}
	rows := make([]operationEvidenceRow, 0)
	seenSources := make(map[string]operationEvidenceSourceOperation)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		connector := entry.Name()
		lockPath := filepath.Join(defsRoot, connector, "sources", connector+"-operation-source-lock.json")
		if _, err := os.Stat(lockPath); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return operationEvidenceArtifact{}, fmt.Errorf("stat source lock for %q: %w", connector, err)
		}
		input, err := readOperationEvidenceSourceLock(lockPath, connector)
		if err != nil {
			return operationEvidenceArtifact{}, err
		}
		if input.Absence != nil {
			rows = append(rows, operationEvidenceAbsentRow(input))
			continue
		}
		bundle, loadErr := engine.Load(os.DirFS(defsRoot), connector)
		crosswalk := readOperationEvidenceCrosswalk(filepath.Join(defsRoot, connector, "sources", connector+"-operation-crosswalk.json"))
		dispositions := readOperationEvidenceDispositions(filepath.Join(defsRoot, connector, "sources", connector+"-declaration-disposition.json"))
		website := websiteRows[connector]
		var report conformance.Report
		if loadErr == nil {
			report = conformance.RunBundle(bundle)
		}
		for _, source := range input.Operations {
			identity := connector + "\x00" + source.ID
			if previous, found := seenSources[identity]; found {
				if previous != source {
					return operationEvidenceArtifact{}, fmt.Errorf("source lock for %q repeats source operation %q with conflicting evidence", connector, source.ID)
				}
				continue
			}
			seenSources[identity] = source
			rows = append(rows, projectOperationEvidenceRow(root, connector, source, bundle, loadErr, website, report, crosswalk[source.ID], dispositions[source.ID]))
		}
	}
	if len(rows) == 0 {
		return operationEvidenceArtifact{}, errors.New("no connector-owned operation source locks found")
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Connector != rows[j].Connector {
			return rows[i].Connector < rows[j].Connector
		}
		return rows[i].SourceID < rows[j].SourceID
	})
	return operationEvidenceArtifact{
		SchemaVersion:         operationEvidenceSchemaVersion,
		GeneratedCommand:      "go run ./cmd/connectorgen operation-evidence",
		Provenance:            provenance,
		Rows:                  rows,
		MissingFoundations:    operationEvidenceRollups(rows),
		IntentionallyReadOnly: operationEvidenceReadOnlyRollups(rows),
	}, nil
}

func readOperationEvidenceProvenance(root string) (operationEvidenceProvenance, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(certificationSubjectArtifactPath)))
	if err != nil {
		return operationEvidenceProvenance{}, fmt.Errorf("read certification source provenance: %w", err)
	}
	var artifact currentCertificationSubjectArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return operationEvidenceProvenance{}, fmt.Errorf("parse certification source provenance: %w", err)
	}
	if err := validateCertificationSubject(artifact.Subject); err != nil {
		return operationEvidenceProvenance{}, fmt.Errorf("validate certification source provenance: %w", err)
	}
	return operationEvidenceProvenance{
		SourceProjectionSHA256: artifact.Subject.SourceProjectionSHA256,
		RelevantConfigSHA256:   artifact.Subject.RelevantConfigSHA256,
	}, nil
}

type operationEvidenceSourceInput struct {
	Connector  string
	LockPath   string
	Operations []operationEvidenceSourceOperation
	Absence    *operationEvidenceAbsence
}

type operationEvidenceSourceOperation struct {
	ID                        string
	Protocol                  string
	Method                    string
	Path                      string
	Trace                     operationEvidenceSourceTrace
	SourceContractUnavailable bool
}

type operationEvidenceRawLock struct {
	SchemaVersion int    `json:"schema_version"`
	Connector     string `json:"connector"`
	State         string `json:"state"`
	Skip          *struct {
		AttemptedURL string `json:"attempted_url"`
		Reason       string `json:"reason"`
		Detail       string `json:"detail"`
	} `json:"skip"`
	Dynamic *struct {
		Reason string `json:"reason"`
		Detail string `json:"detail"`
	} `json:"dynamic"`
	Rest struct {
		SourceURL   string `json:"source_url"`
		SHA256      string `json:"sha256"`
		Bytes       int64  `json:"bytes"`
		SourceKind  string `json:"source_kind"`
		Supplements []struct {
			SourceURL string `json:"source_url"`
			SHA256    string `json:"sha256"`
			Bytes     int64  `json:"bytes"`
		} `json:"supplements"`
		SourceDocuments []json.RawMessage `json:"source_documents"`
		Documents       []struct {
			SourceURL string `json:"source_url"`
			SHA256    string `json:"sha256"`
			Bytes     int64  `json:"bytes"`
		} `json:"documents"`
		Operations []struct {
			ID             string `json:"id"`
			Protocol       string `json:"protocol"`
			Method         string `json:"method"`
			Path           string `json:"path"`
			SourceLocation string `json:"source_location"`
			SourceURL      string `json:"source_url"`
		} `json:"operations"`
		CoverageConfidence struct {
			Level string `json:"level"`
			Basis string `json:"basis"`
		} `json:"coverage_confidence"`
	} `json:"rest"`
	GraphQL struct {
		SourceURL        string `json:"source_url"`
		SHA256           string `json:"sha256"`
		Bytes            int64  `json:"bytes"`
		ProjectionSHA256 string `json:"projection_sha256"`
		ProjectionBytes  int64  `json:"projection_bytes"`
		QueryFields      []struct {
			Name string `json:"name"`
			Line int    `json:"line"`
		} `json:"query_fields"`
		MutationFields []struct {
			Name string `json:"name"`
			Line int    `json:"line"`
		} `json:"mutation_fields"`
	} `json:"graphql"`
}

func readOperationEvidenceSourceLock(path, connector string) (operationEvidenceSourceInput, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return operationEvidenceSourceInput{}, fmt.Errorf("read source lock for %q: %w", connector, err)
	}
	var lock operationEvidenceRawLock
	// This partial reader permits unknown fields for legacy evidence, but a
	// duplicate member must fail before state/inventory inspection. Otherwise a
	// last-member-wins decoder could suppress a v3 document inventory as absence.
	if err := decodeSourceJSON(raw, &lock); err != nil {
		return operationEvidenceSourceInput{}, fmt.Errorf("parse source lock for %q: %w", connector, err)
	}
	if lock.SchemaVersion != 2 && lock.SchemaVersion != 3 {
		return operationEvidenceSourceInput{}, fmt.Errorf("source lock for %q has unsupported schema_version %d", connector, lock.SchemaVersion)
	}
	if lock.Connector != connector {
		return operationEvidenceSourceInput{}, fmt.Errorf("source lock connector %q does not match directory %q", lock.Connector, connector)
	}
	input := operationEvidenceSourceInput{Connector: connector, LockPath: filepath.ToSlash(filepath.Join("sources", filepath.Base(path)))}
	// Validate the v2 wire before any provider-absence projection. Otherwise a
	// skipped or dynamic state could suppress the presence check and let a
	// reference-only field (including source_kind:null) hide in a byte-backed
	// legacy inventory.
	legacyReference, err := operationEvidenceLegacyReferenceWire(raw, lock.SchemaVersion)
	if err != nil {
		return operationEvidenceSourceInput{}, fmt.Errorf("parse source lock for %q: %w", connector, err)
	}
	// A v3 document-owned inventory is the strict source-import contract. It
	// cannot be hidden behind the historical provider-absence projection just
	// by adding an otherwise legacy state field. Empty v3 absence locks remain
	// evidence-only because they have no document operations to suppress.
	isProviderAbsence := lock.State == "skipped" || lock.State == "dynamic"
	hasV3DocumentInventory := lock.SchemaVersion == 3 && len(lock.Rest.SourceDocuments) != 0
	if isProviderAbsence && !hasV3DocumentInventory && !legacyReference {
		absence := operationEvidenceAbsence{State: lock.State}
		if lock.Skip != nil {
			absence.Reason = lock.Skip.Reason
			absence.Evidence = firstNonEmpty(lock.Skip.Detail, lock.Skip.AttemptedURL, lock.Rest.CoverageConfidence.Basis)
		}
		if lock.Dynamic != nil {
			absence.Reason = lock.Dynamic.Reason
			absence.Evidence = firstNonEmpty(lock.Dynamic.Detail, lock.Rest.CoverageConfidence.Basis, lock.Rest.SourceURL)
		}
		if absence.Reason == "" || absence.Evidence == "" {
			return operationEvidenceSourceInput{}, fmt.Errorf("provider absence in source lock for %q lacks reason or evidence", connector)
		}
		input.Absence = &absence
		return input, nil
	}
	// Version 3 and an explicit non-null v2 source-reference discriminator are
	// the strict source-import contracts. Historical byte-backed v2 evidence
	// still uses its tolerant projection, but reference-only fields cannot be
	// silently discarded or reinterpreted on that path.
	if lock.SchemaVersion == 3 || legacyReference {
		strictLock, err := parseSourceImportLock(raw, connector)
		if err != nil {
			return operationEvidenceSourceInput{}, fmt.Errorf("parse source lock for %q through source-import schema: %w", connector, err)
		}
		return operationEvidenceSourceInputFromImportLock(input, strictLock)
	}
	return operationEvidenceSourceInputFromLegacyLock(input, lock)
}

func operationEvidenceLegacyReferenceWire(raw []byte, schemaVersion int) (bool, error) {
	if schemaVersion != 2 {
		return false, nil
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return false, fmt.Errorf("decode source lock fields: %w", err)
	}
	var rest map[string]json.RawMessage
	if restRaw, exists := root["rest"]; exists {
		if err := json.Unmarshal(restRaw, &rest); err != nil {
			return false, fmt.Errorf("decode source lock rest fields: %w", err)
		}
	}
	if kindRaw, exists := rest["source_kind"]; exists {
		var kind string
		if err := json.Unmarshal(kindRaw, &kind); err != nil || kind == "" {
			return false, fmt.Errorf("source-reference discriminator rest.source_kind must be a non-empty string")
		}
		return true, nil
	}
	for _, field := range []string{"operations_found", "coverage_confidence"} {
		if _, exists := root[field]; exists {
			return false, fmt.Errorf("byte-backed v2 source lock cannot declare reference-only field %s", field)
		}
	}
	for _, field := range []string{"operation_counts", "supplements"} {
		if _, exists := rest[field]; exists {
			return false, fmt.Errorf("byte-backed v2 source lock cannot declare reference-only field rest.%s", field)
		}
	}
	if operationsRaw, exists := rest["operations"]; exists {
		var operations []map[string]json.RawMessage
		if err := json.Unmarshal(operationsRaw, &operations); err != nil {
			return false, fmt.Errorf("decode source lock REST operations: %w", err)
		}
		for _, operation := range operations {
			if _, exists := operation["source_url"]; exists {
				return false, fmt.Errorf("byte-backed v2 source lock cannot declare reference-only field rest.operations[].source_url")
			}
		}
	}
	return false, nil
}

func operationEvidenceSourceInputFromLegacyLock(input operationEvidenceSourceInput, lock operationEvidenceRawLock) (operationEvidenceSourceInput, error) {
	documents := make(map[string]operationEvidenceSourceTrace, len(lock.Rest.Documents))
	for _, document := range lock.Rest.Documents {
		documents[document.SourceURL] = operationEvidenceSourceTrace{Lock: input.LockPath, URL: document.SourceURL, SHA256: document.SHA256, Bytes: document.Bytes}
	}
	for _, operation := range lock.Rest.Operations {
		if operation.ID == "" || operation.Method == "" || operation.Path == "" {
			return operationEvidenceSourceInput{}, fmt.Errorf("source lock for %q contains an operation without identity", input.Connector)
		}
		trace := documents[operation.SourceURL]
		if trace.URL == "" {
			trace = operationEvidenceSourceTrace{Lock: input.LockPath, URL: firstNonEmpty(operation.SourceURL, lock.Rest.SourceURL), SHA256: lock.Rest.SHA256, Bytes: lock.Rest.Bytes}
		}
		trace.Lock = input.LockPath
		trace.Location = operation.SourceLocation
		input.Operations = append(input.Operations, operationEvidenceSourceOperation{ID: operation.ID, Protocol: firstNonEmpty(operation.Protocol, "rest"), Method: strings.ToUpper(operation.Method), Path: operation.Path, Trace: trace})
	}
	graphqlSHA := firstNonEmpty(lock.GraphQL.ProjectionSHA256, lock.GraphQL.SHA256)
	graphqlBytes := lock.GraphQL.ProjectionBytes
	if graphqlBytes <= 0 {
		graphqlBytes = lock.GraphQL.Bytes
	}
	for _, field := range lock.GraphQL.QueryFields {
		input.Operations = append(input.Operations, operationEvidenceSourceOperation{ID: input.Connector + ".graphql.query." + field.Name, Protocol: "graphql", Method: "POST", Path: "/graphql", Trace: operationEvidenceSourceTrace{Lock: input.LockPath, URL: lock.GraphQL.SourceURL, SHA256: graphqlSHA, Bytes: graphqlBytes, Location: fmt.Sprintf("Query.%s:%d", field.Name, field.Line)}})
	}
	for _, field := range lock.GraphQL.MutationFields {
		input.Operations = append(input.Operations, operationEvidenceSourceOperation{ID: input.Connector + ".graphql.mutation." + field.Name, Protocol: "graphql", Method: "POST", Path: "/graphql", Trace: operationEvidenceSourceTrace{Lock: input.LockPath, URL: lock.GraphQL.SourceURL, SHA256: graphqlSHA, Bytes: graphqlBytes, Location: fmt.Sprintf("Mutation.%s:%d", field.Name, field.Line)}})
	}
	if len(input.Operations) == 0 {
		return operationEvidenceSourceInput{}, fmt.Errorf("source lock for %q has no operations and no provider-evidenced absence", input.Connector)
	}
	return input, nil
}

func operationEvidenceSourceInputFromImportLock(input operationEvidenceSourceInput, lock sourceImportLock) (operationEvidenceSourceInput, error) {
	if lock.SchemaVersion == 3 {
		for _, document := range lock.Rest.SourceDocuments {
			operationEvidenceAppendRESTOperations(&input, document.Operations, document.sourceArtifact(), document.isSourceReference())
		}
	} else if lock.isLegacySourceReference() {
		artifacts, err := sourceImportLegacyReferenceArtifacts(lock)
		if err != nil {
			return operationEvidenceSourceInput{}, err
		}
		for _, operation := range lock.Rest.Operations {
			artifact, found := artifacts[operation.SourceURL]
			if !found {
				return operationEvidenceSourceInput{}, fmt.Errorf("source-reference operation %q cites an undeclared source URL", operation.ID)
			}
			operationEvidenceAppendRESTOperations(&input, []sourceImportRESTOperation{operation}, artifact, true)
		}
	} else {
		operationEvidenceAppendRESTOperations(&input, lock.Rest.Operations, lock.Rest.sourceImportArtifact, false)
	}
	graphqlSHA := firstNonEmpty(lock.GraphQL.ProjectionSHA256, lock.GraphQL.SHA256)
	graphqlBytes := lock.GraphQL.ProjectionBytes
	if graphqlBytes <= 0 {
		graphqlBytes = lock.GraphQL.Bytes
	}
	for _, field := range lock.GraphQL.QueryFields {
		input.Operations = append(input.Operations, operationEvidenceSourceOperation{ID: lock.Connector + ".graphql.query." + field.Name, Protocol: "graphql", Method: "POST", Path: "/graphql", Trace: operationEvidenceSourceTrace{Lock: input.LockPath, URL: lock.GraphQL.SourceURL, SHA256: graphqlSHA, Bytes: graphqlBytes, Location: fmt.Sprintf("Query.%s:%d", field.Name, field.Line)}})
	}
	for _, field := range lock.GraphQL.MutationFields {
		input.Operations = append(input.Operations, operationEvidenceSourceOperation{ID: lock.Connector + ".graphql.mutation." + field.Name, Protocol: "graphql", Method: "POST", Path: "/graphql", Trace: operationEvidenceSourceTrace{Lock: input.LockPath, URL: lock.GraphQL.SourceURL, SHA256: graphqlSHA, Bytes: graphqlBytes, Location: fmt.Sprintf("Mutation.%s:%d", field.Name, field.Line)}})
	}
	if len(input.Operations) == 0 {
		return operationEvidenceSourceInput{}, fmt.Errorf("source lock for %q has no operations and no provider-evidenced absence", lock.Connector)
	}
	return input, nil
}

func operationEvidenceAppendRESTOperations(input *operationEvidenceSourceInput, operations []sourceImportRESTOperation, artifact sourceImportArtifact, sourceContractUnavailable bool) {
	for _, operation := range operations {
		input.Operations = append(input.Operations, operationEvidenceSourceOperation{
			ID:       operation.ID,
			Protocol: firstNonEmpty(operation.Protocol, "rest"),
			Method:   strings.ToUpper(operation.Method),
			Path:     operation.Path,
			Trace: operationEvidenceSourceTrace{
				Lock:     input.LockPath,
				URL:      artifact.SourceURL,
				SHA256:   artifact.SHA256,
				Bytes:    artifact.Bytes,
				Location: operation.SourceLocation,
			},
			SourceContractUnavailable: sourceContractUnavailable,
		})
	}
}

type operationEvidenceCrosswalk struct {
	State  string
	Method string
	Path   string
}

func readOperationEvidenceCrosswalk(path string) map[string]operationEvidenceCrosswalk {
	raw, err := os.ReadFile(path)
	if err != nil {
		return map[string]operationEvidenceCrosswalk{}
	}
	var document struct {
		SourceOperations []struct {
			SourceID  string `json:"source_id"`
			Crosswalk struct {
				State      string `json:"state"`
				APISurface *struct {
					Method string `json:"method"`
					Path   string `json:"path"`
				} `json:"api_surface"`
			} `json:"crosswalk"`
		} `json:"source_operations"`
	}
	if json.Unmarshal(raw, &document) != nil {
		return map[string]operationEvidenceCrosswalk{}
	}
	values := make(map[string]operationEvidenceCrosswalk, len(document.SourceOperations))
	for _, operation := range document.SourceOperations {
		value := operationEvidenceCrosswalk{State: operation.Crosswalk.State}
		if operation.Crosswalk.APISurface != nil {
			value.Method = operation.Crosswalk.APISurface.Method
			value.Path = operation.Crosswalk.APISurface.Path
		}
		values[operation.SourceID] = value
	}
	return values
}

type operationEvidenceDisposition struct {
	ParityClass string
	Foundations []operationEvidenceFoundation
}

func readOperationEvidenceDispositions(path string) map[string]operationEvidenceDisposition {
	raw, err := os.ReadFile(path)
	if err != nil {
		return map[string]operationEvidenceDisposition{}
	}
	var document struct {
		LedgerDispositions []struct {
			ParityClass string `json:"parity_class"`
			Source      struct {
				SourceID string `json:"source_id"`
			} `json:"source"`
			ReverseETLEligibility json.RawMessage `json:"reverse_etl_eligibility"`
		} `json:"ledger_dispositions"`
	}
	if json.Unmarshal(raw, &document) != nil {
		return map[string]operationEvidenceDisposition{}
	}
	values := make(map[string]operationEvidenceDisposition, len(document.LedgerDispositions))
	for _, disposition := range document.LedgerDispositions {
		value := operationEvidenceDisposition{ParityClass: disposition.ParityClass}
		value.Foundations = append(value.Foundations, operationEvidenceFoundationsFromRaw(disposition.ReverseETLEligibility)...)
		values[disposition.Source.SourceID] = value
	}
	return values
}

func operationEvidenceFoundationsFromRaw(raw json.RawMessage) []operationEvidenceFoundation {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value struct {
		FoundationGap *struct {
			ID       string `json:"id"`
			Evidence string `json:"evidence"`
		} `json:"foundation_gap"`
	}
	if json.Unmarshal(raw, &value) != nil || value.FoundationGap == nil || value.FoundationGap.ID == "" {
		return nil
	}
	return []operationEvidenceFoundation{{ID: value.FoundationGap.ID, Evidence: value.FoundationGap.Evidence}}
}

type operationEvidenceWebsiteRow struct {
	Commands []operationEvidenceWebsiteCommand `json:"commands"`
}

type operationEvidenceWebsiteCommand struct {
	Path       string `json:"path"`
	Intent     string `json:"intent"`
	Stream     string `json:"stream"`
	Write      string `json:"write"`
	Operation  string `json:"operation"`
	APISurface []struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	} `json:"api_surface"`
}

func readOperationEvidenceWebsiteRows(path string) (map[string]operationEvidenceWebsiteRow, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read generated website connector data: %w", err)
	}
	var rows []struct {
		Slug       string                       `json:"slug"`
		CLISurface *operationEvidenceWebsiteRow `json:"cli_surface"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		var envelope struct {
			Rows []struct {
				Slug       string                       `json:"slug"`
				CLISurface *operationEvidenceWebsiteRow `json:"cli_surface"`
			} `json:"rows"`
		}
		if envelopeErr := json.Unmarshal(raw, &envelope); envelopeErr != nil {
			return nil, fmt.Errorf("parse generated website connector data: %w", err)
		}
		rows = envelope.Rows
	}
	result := make(map[string]operationEvidenceWebsiteRow, len(rows))
	for _, row := range rows {
		if row.Slug != "" && row.CLISurface != nil {
			result[row.Slug] = *row.CLISurface
		}
	}
	return result, nil
}

func operationEvidenceAbsentRow(input operationEvidenceSourceInput) operationEvidenceRow {
	return operationEvidenceRow{
		Connector:       input.Connector,
		SourceID:        input.Connector + ".provider-surface",
		Protocol:        "provider_surface",
		Source:          operationEvidenceSourceTrace{Lock: input.LockPath},
		Canonical:       operationEvidenceCanonical{State: "provider_absent"},
		Runtime:         operationEvidenceRuntime{Targets: []string{}},
		CLI:             operationEvidenceCommands{Paths: []string{}},
		Website:         operationEvidenceCommands{Paths: []string{}},
		Fixtures:        operationEvidenceFixtures{Paths: []string{}},
		Conformance:     operationEvidenceConformance{},
		Classifications: operationEvidenceEmptyClassifications(),
		Foundations:     []operationEvidenceFoundation{},
		Gaps:            []operationEvidenceGap{},
		Absence:         input.Absence,
	}
}

func projectOperationEvidenceRow(root, connector string, source operationEvidenceSourceOperation, bundle engine.Bundle, loadErr error, website operationEvidenceWebsiteRow, report conformance.Report, crosswalk operationEvidenceCrosswalk, disposition operationEvidenceDisposition) operationEvidenceRow {
	row := operationEvidenceRow{
		Connector:       connector,
		SourceID:        source.ID,
		Protocol:        source.Protocol,
		Method:          source.Method,
		Path:            source.Path,
		Source:          source.Trace,
		Canonical:       operationEvidenceCanonical{State: "missing"},
		Runtime:         operationEvidenceRuntime{Targets: []string{}},
		CLI:             operationEvidenceCommands{Paths: []string{}},
		Website:         operationEvidenceCommands{Paths: []string{}},
		Fixtures:        operationEvidenceFixtures{Paths: []string{}},
		Conformance:     operationEvidenceConformance{},
		Classifications: operationEvidenceEmptyClassifications(),
		Foundations:     append([]operationEvidenceFoundation(nil), disposition.Foundations...),
		Gaps:            []operationEvidenceGap{},
	}
	if row.Source.URL == "" || row.Source.SHA256 == "" || row.Source.Bytes <= 0 || row.Source.Location == "" {
		row.addGap(operationEvidenceGapSourceTrace, "source lock lacks a complete provider trace")
	}
	if source.SourceContractUnavailable {
		row.addGap(sourceContractUnavailableFoundation, "provider operation is cited by a closed source reference, but retained source bytes and execution-contract detail are source_contract_unavailable")
	}
	var endpoint *engine.SurfaceEndpoint
	if loadErr == nil {
		endpoint = operationEvidenceSurfaceEndpoint(bundle, source)
	}
	if crosswalk.State != "" {
		row.Canonical.State = crosswalk.State
		row.Canonical.Method = firstNonEmpty(crosswalk.Method, source.Method)
		row.Canonical.Path = firstNonEmpty(crosswalk.Path, source.Path)
	}
	if endpoint != nil {
		row.Canonical = operationEvidenceCanonical{State: "mapped", Method: endpoint.Method, Path: endpoint.Path}
	}
	if (row.Canonical.State != "mapped" && row.Canonical.State != "exact") || row.Canonical.Method == "" || row.Canonical.Path == "" {
		row.addGap(operationEvidenceGapCanonicalMapping, "source operation has no canonical api_surface/crosswalk mapping")
	}
	if loadErr != nil {
		row.addGap(operationEvidenceGapRuntimeReachability, "load connector bundle: "+loadErr.Error())
		return row.finalize()
	}
	targets := operationEvidenceTargetsFor(bundle, source, endpoint)
	row.Runtime.Targets = targets.names()
	commands := operationEvidenceMatchingCommands(bundle, source, endpoint, targets)
	websiteCommands := operationEvidenceMatchingWebsiteCommands(website, source, endpoint, targets)
	row.CLI.Paths = operationEvidenceCommandPaths(commands)
	row.Website.Paths = operationEvidenceWebsitePaths(websiteCommands)
	if declaration, declared, err := sourceReadOnlyOperationDeclaration(operationForSurfaceEndpoint(endpoint)); declared {
		if err != nil {
			row.addGap(operationEvidenceGapReadOnlyConflict, "invalid read-only declaration: "+err.Error())
			return row.finalize()
		}
		if sourceProjectionMutationMethod(source.Method) {
			row.addGap(operationEvidenceGapReadOnlyConflict, "read-only declaration cannot cover a mutating source operation")
			return row.finalize()
		}
		row.ReadOnly = &operationEvidenceReadOnly{Policy: declaration.Policy, Reason: declaration.Reason}
		if len(row.Runtime.Targets) > 0 || len(row.CLI.Paths) > 0 || len(row.Website.Paths) > 0 {
			row.addGap(operationEvidenceGapReadOnlyConflict, "read-only declaration has executable runtime, CLI, or website evidence")
			return row.finalize()
		}
		return row.finalize()
	}
	operationEvidenceClassify(&row, targets, commands, disposition.ParityClass)
	if source.SourceContractUnavailable {
		for _, class := range operationEvidenceClasses {
			value := row.Classifications[class]
			value.Enabled = false
			row.Classifications[class] = value
		}
	}
	if len(commands) == 0 {
		row.addGap(operationEvidenceGapCLICommand, "canonical operation has no generated CLI command")
	}
	if len(websiteCommands) == 0 {
		row.addGap(operationEvidenceGapWebsiteRow, "canonical operation has no generated website command row")
	}
	row.Runtime.Enabled = !source.SourceContractUnavailable && len(targets.names()) > 0 && operationEvidenceHasEnabledCommand(commands)
	if !row.Runtime.Enabled {
		evidence := "canonical operation has no enabled declaration-owned runtime command"
		if source.SourceContractUnavailable {
			evidence = "canonical declaration cannot become source-backed runtime evidence while the operation contract is source_contract_unavailable"
		}
		row.addGap(operationEvidenceGapRuntimeReachability, evidence)
	}
	row.Fixtures.Paths = operationEvidenceFixturePaths(root, connector, source, targets, commands)
	if len(row.Fixtures.Paths) == 0 {
		row.addGap(operationEvidenceGapFixtureProof, "canonical operation has no matching checked-in fixture")
	}
	row.Conformance = operationEvidenceProof(report, targets, commands)
	if !row.Conformance.Passed {
		row.addGap(operationEvidenceGapConformanceProof, "canonical operation has no passing matching conformance check")
	}
	return row.finalize()
}

func operationEvidenceEmptyClassifications() map[string]operationEvidenceClassification {
	classes := make(map[string]operationEvidenceClassification, len(operationEvidenceClasses))
	for _, class := range operationEvidenceClasses {
		classes[class] = operationEvidenceClassification{}
	}
	return classes
}

func (row *operationEvidenceRow) addGap(kind, evidence string) {
	for _, gap := range row.Gaps {
		if gap.Kind == kind && gap.Evidence == evidence {
			return
		}
	}
	row.Gaps = append(row.Gaps, operationEvidenceGap{Kind: kind, Evidence: evidence})
}

func (row operationEvidenceRow) finalize() operationEvidenceRow {
	row.Runtime.Targets = operationEvidenceSortedUnique(row.Runtime.Targets)
	row.CLI.Paths = operationEvidenceSortedUnique(row.CLI.Paths)
	row.Website.Paths = operationEvidenceSortedUnique(row.Website.Paths)
	row.Fixtures.Paths = operationEvidenceSortedUnique(row.Fixtures.Paths)
	sort.Slice(row.Foundations, func(i, j int) bool { return row.Foundations[i].ID < row.Foundations[j].ID })
	sort.Slice(row.Gaps, func(i, j int) bool {
		if row.Gaps[i].Kind != row.Gaps[j].Kind {
			return row.Gaps[i].Kind < row.Gaps[j].Kind
		}
		return row.Gaps[i].Evidence < row.Gaps[j].Evidence
	})
	return row
}

func operationEvidenceSurfaceEndpoint(bundle engine.Bundle, source operationEvidenceSourceOperation) *engine.SurfaceEndpoint {
	if bundle.Surface == nil {
		return nil
	}
	for index := range bundle.Surface.Endpoints {
		endpoint := &bundle.Surface.Endpoints[index]
		if strings.EqualFold(endpoint.Method, source.Method) && endpoint.Path == source.Path {
			return endpoint
		}
	}
	return nil
}

type operationEvidenceTargets struct {
	Streams    []string
	Writes     []string
	Operations []string
}

func (targets operationEvidenceTargets) names() []string {
	values := make([]string, 0, len(targets.Streams)+len(targets.Writes)+len(targets.Operations))
	for _, value := range targets.Streams {
		values = append(values, "stream:"+value)
	}
	for _, value := range targets.Writes {
		values = append(values, "write:"+value)
	}
	for _, value := range targets.Operations {
		values = append(values, "operation:"+value)
	}
	return operationEvidenceSortedUnique(values)
}

func operationEvidenceTargetsFor(bundle engine.Bundle, source operationEvidenceSourceOperation, endpoint *engine.SurfaceEndpoint) operationEvidenceTargets {
	targets := operationEvidenceTargets{}
	if endpoint != nil && endpoint.CoveredBy != nil {
		if endpoint.CoveredBy.Stream != "" {
			targets.Streams = append(targets.Streams, endpoint.CoveredBy.Stream)
		}
		targets.Writes = append(targets.Writes, endpoint.CoveredBy.WriteTargets()...)
		if source.Protocol != "graphql" {
			targets.Operations = append(targets.Operations, endpoint.CoveredBy.OperationTargets()...)
		}
	}
	for _, operation := range bundle.Operations {
		if operationEvidenceOperationMatchesSource(operation, source) {
			targets.Operations = append(targets.Operations, operation.ID)
		}
	}
	return operationEvidenceTargets{Streams: operationEvidenceSortedUnique(targets.Streams), Writes: operationEvidenceSortedUnique(targets.Writes), Operations: operationEvidenceSortedUnique(targets.Operations)}
}

func operationEvidenceOperationMatchesSource(operation engine.OperationSpec, source operationEvidenceSourceOperation) bool {
	if source.Protocol == "graphql" {
		return operationEvidenceGraphQLSourceField(source.ID) == operationEvidenceGraphQLRootFieldForOperation(operation)
	}
	if operation.REST != nil && strings.EqualFold(operation.REST.Method, source.Method) && operation.REST.Path == source.Path {
		return true
	}
	if operation.Binary != nil && strings.EqualFold(operation.Binary.Method, source.Method) && operation.Binary.Path == source.Path {
		return true
	}
	return false
}

func operationEvidenceMatchingCommands(bundle engine.Bundle, source operationEvidenceSourceOperation, endpoint *engine.SurfaceEndpoint, targets operationEvidenceTargets) []engine.CLICommand {
	if bundle.CLISurface == nil {
		return nil
	}
	streamSet := operationEvidenceSet(targets.Streams)
	writeSet := operationEvidenceSet(targets.Writes)
	operationSet := operationEvidenceSet(targets.Operations)
	matched := make([]engine.CLICommand, 0)
	for _, command := range bundle.CLISurface.Commands {
		if operationEvidenceCommandMatches(command, source, endpoint, streamSet, writeSet, operationSet) {
			matched = append(matched, command)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].Path < matched[j].Path })
	return matched
}

func operationEvidenceCommandMatches(command engine.CLICommand, source operationEvidenceSourceOperation, endpoint *engine.SurfaceEndpoint, streams, writes, operations map[string]bool) bool {
	if operations[command.Operation] {
		return true
	}
	// Every GraphQL operation shares POST /graphql. The physical endpoint is
	// therefore not an operation identity: matching it would claim every
	// GraphQL command as evidence for every GraphQL source field.
	if source.Protocol == "graphql" {
		return false
	}
	for _, reference := range command.APISurface {
		if strings.EqualFold(reference.Method, source.Method) && reference.Path == source.Path {
			return true
		}
	}
	if endpoint != nil && command.Operation != "" {
		for _, target := range endpoint.CoveredBy.OperationTargets() {
			if command.Operation == target {
				return true
			}
		}
	}
	// A command with an explicit api_surface must match that operation exactly.
	// Streams and writes can cover multiple provider operations, so using their
	// shared target as the first predicate would fabricate command evidence.
	return len(command.APISurface) == 0 && (streams[command.Stream] || writes[command.Write])
}

func operationEvidenceMatchingWebsiteCommands(website operationEvidenceWebsiteRow, source operationEvidenceSourceOperation, endpoint *engine.SurfaceEndpoint, targets operationEvidenceTargets) []operationEvidenceWebsiteCommand {
	streams := operationEvidenceSet(targets.Streams)
	writes := operationEvidenceSet(targets.Writes)
	operations := operationEvidenceSet(targets.Operations)
	matched := make([]operationEvidenceWebsiteCommand, 0)
	for _, command := range website.Commands {
		if operations[command.Operation] {
			matched = append(matched, command)
			continue
		}
		if source.Protocol == "graphql" {
			continue
		}
		for _, reference := range command.APISurface {
			if strings.EqualFold(reference.Method, source.Method) && reference.Path == source.Path {
				matched = append(matched, command)
				break
			}
		}
		if len(command.APISurface) == 0 && (streams[command.Stream] || writes[command.Write]) {
			matched = append(matched, command)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].Path < matched[j].Path })
	return matched
}

func operationEvidenceGraphQLSourceField(sourceID string) string {
	if _, field, found := strings.Cut(sourceID, ".graphql.query."); found {
		return field
	}
	if _, field, found := strings.Cut(sourceID, ".graphql.mutation."); found {
		return field
	}
	return ""
}

func operationEvidenceGraphQLRootFieldForOperation(operation engine.OperationSpec) string {
	if operation.GraphQL == nil {
		return ""
	}
	match := operationEvidenceGraphQLRootField.FindStringSubmatch(operation.GraphQL.Document)
	if len(match) != 2 {
		return ""
	}
	return match[1]
}

func operationEvidenceClassify(row *operationEvidenceRow, targets operationEvidenceTargets, commands []engine.CLICommand, dispositionClass string) {
	operationEvidenceSetClassification(row, operationEvidenceClassForDisposition(dispositionClass), false)
	// A provider operation can serve both a saved transport lane and an
	// interactive command lane. The runtime target owns the former; command
	// intent owns the latter. Keeping them independent prevents a bounded
	// direct command from erasing the same operation's ETL/reverse-ETL proof.
	targetEnabled := operationEvidenceHasEnabledCommand(commands)
	if len(targets.Streams) > 0 {
		operationEvidenceSetClassification(row, operationEvidenceClassETL, targetEnabled)
	}
	if len(targets.Writes) > 0 {
		operationEvidenceSetClassification(row, operationEvidenceClassReverseETL, targetEnabled)
	}
	for _, command := range commands {
		class := operationEvidenceClassForCommand(command.Intent, command.Operation)
		operationEvidenceSetClassification(row, class, command.Availability == "implemented")
	}
}

func operationEvidenceClassForDisposition(value string) string {
	switch value {
	case operationEvidenceClassETL, operationEvidenceClassReverseETL, operationEvidenceClassDirectRead, operationEvidenceClassDirectWrite, operationEvidenceClassBinaryDownload, operationEvidenceClassBinaryUpload:
		return value
	case "binary_read":
		return operationEvidenceClassBinaryDownload
	case "binary_write":
		return operationEvidenceClassBinaryUpload
	default:
		return ""
	}
}

func operationEvidenceClassForCommand(intent, operation string) string {
	switch intent {
	case "etl":
		return operationEvidenceClassETL
	case "reverse_etl":
		return operationEvidenceClassReverseETL
	case "direct_read":
		return operationEvidenceClassDirectRead
	case "binary_download":
		return operationEvidenceClassBinaryDownload
	case "binary_upload":
		return operationEvidenceClassBinaryUpload
	case "direct_write":
		return operationEvidenceClassDirectWrite
	default:
		return ""
	}
}

func operationEvidenceSetClassification(row *operationEvidenceRow, class string, enabled bool) {
	if class == "" {
		return
	}
	value := row.Classifications[class]
	value.Declared = true
	value.Enabled = value.Enabled || enabled
	row.Classifications[class] = value
}

func operationEvidenceCommandPaths(commands []engine.CLICommand) []string {
	paths := make([]string, 0, len(commands))
	for _, command := range commands {
		paths = append(paths, command.Path)
	}
	return operationEvidenceSortedUnique(paths)
}

func operationEvidenceWebsitePaths(commands []operationEvidenceWebsiteCommand) []string {
	paths := make([]string, 0, len(commands))
	for _, command := range commands {
		paths = append(paths, command.Path)
	}
	return operationEvidenceSortedUnique(paths)
}

func operationEvidenceHasEnabledCommand(commands []engine.CLICommand) bool {
	for _, command := range commands {
		if command.Availability == "implemented" {
			return true
		}
	}
	return false
}

func operationEvidenceFixturePaths(root, connector string, source operationEvidenceSourceOperation, targets operationEvidenceTargets, commands []engine.CLICommand) []string {
	base := filepath.Join(root, "internal", "connectors", "defs", connector, "fixtures")
	paths := make([]string, 0)
	for _, stream := range targets.Streams {
		directory := filepath.Join(base, "streams", stream)
		entries, err := os.ReadDir(directory)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				paths = append(paths, filepath.ToSlash(filepath.Join("fixtures", "streams", stream, entry.Name())))
			}
		}
	}
	for _, write := range targets.Writes {
		path := filepath.Join(base, "writes", write+".json")
		if _, err := os.Stat(path); err == nil {
			paths = append(paths, filepath.ToSlash(filepath.Join("fixtures", "writes", write+".json")))
		}
	}
	if len(paths) == 0 && source.Protocol != "" && operationEvidenceHasDirectCommand(commands) {
		if _, err := os.Stat(filepath.Join(base, "check.json")); err == nil {
			paths = append(paths, "fixtures/check.json")
		}
	}
	return operationEvidenceSortedUnique(paths)
}

func operationEvidenceHasDirectCommand(commands []engine.CLICommand) bool {
	for _, command := range commands {
		if command.Intent == "direct_read" || command.Intent == "direct_write" || command.Intent == "binary_download" || command.Intent == "binary_upload" {
			return true
		}
	}
	return false
}

func operationEvidenceProof(report conformance.Report, targets operationEvidenceTargets, commands []engine.CLICommand) operationEvidenceConformance {
	byName := make(map[string]conformance.CheckResult, len(report.Checks))
	for _, check := range report.Checks {
		byName[check.Name] = check
	}
	for _, stream := range targets.Streams {
		name := "read_fixture_nonempty:" + stream
		if check := byName[name]; check.Passed && !check.Skipped {
			return operationEvidenceConformance{Passed: true, Proof: name}
		}
	}
	for _, write := range targets.Writes {
		name := "write_request_shape:" + write
		if check := byName[name]; check.Passed && !check.Skipped {
			return operationEvidenceConformance{Passed: true, Proof: name}
		}
	}
	if operationEvidenceHasDirectCommand(commands) {
		if check := byName["check_fixture"]; check.Passed && !check.Skipped {
			return operationEvidenceConformance{Passed: true, Proof: "check_fixture"}
		}
	}
	return operationEvidenceConformance{}
}

func operationEvidenceRollups(rows []operationEvidenceRow) []operationEvidenceRollup {
	type grouped struct {
		evidence string
		sources  map[string]bool
	}
	groups := map[string]*grouped{}
	for _, row := range rows {
		for _, gap := range row.Gaps {
			id := gap.Kind
			if gap.Foundation != "" {
				id = gap.Foundation
			}
			group := groups[id]
			if group == nil {
				group = &grouped{evidence: gap.Evidence, sources: map[string]bool{}}
				groups[id] = group
			}
			group.sources[row.SourceID] = true
		}
		for _, foundation := range row.Foundations {
			group := groups[foundation.ID]
			if group == nil {
				group = &grouped{evidence: foundation.Evidence, sources: map[string]bool{}}
				groups[foundation.ID] = group
			}
			group.sources[row.SourceID] = true
		}
	}
	ids := make([]string, 0, len(groups))
	for id := range groups {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	rollups := make([]operationEvidenceRollup, 0, len(ids))
	for _, id := range ids {
		group := groups[id]
		sourceIDs := make([]string, 0, len(group.sources))
		for sourceID := range group.sources {
			sourceIDs = append(sourceIDs, sourceID)
		}
		sort.Strings(sourceIDs)
		rollups = append(rollups, operationEvidenceRollup{ID: id, Evidence: group.evidence, SourceIDs: sourceIDs})
	}
	return rollups
}

func operationEvidenceReadOnlyRollups(rows []operationEvidenceRow) []operationEvidenceReadOnlyRollup {
	type group struct {
		connector string
		policy    string
		sources   map[string]bool
	}
	groups := map[string]*group{}
	for _, row := range rows {
		if row.ReadOnly == nil {
			continue
		}
		key := row.Connector + "\x00" + row.ReadOnly.Policy
		if groups[key] == nil {
			groups[key] = &group{connector: row.Connector, policy: row.ReadOnly.Policy, sources: map[string]bool{}}
		}
		groups[key].sources[row.SourceID] = true
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rollups := make([]operationEvidenceReadOnlyRollup, 0, len(keys))
	for _, key := range keys {
		group := groups[key]
		sourceIDs := make([]string, 0, len(group.sources))
		for sourceID := range group.sources {
			sourceIDs = append(sourceIDs, sourceID)
		}
		sort.Strings(sourceIDs)
		rollups = append(rollups, operationEvidenceReadOnlyRollup{Connector: group.connector, Policy: group.policy, SourceIDs: sourceIDs})
	}
	return rollups
}

func buildOperationEvidenceFixed100(artifact operationEvidenceArtifact) (operationEvidenceFixed100, error) {
	rows := make([]operationEvidenceFixedRow, 0, 100)
	selected := make(map[string]bool, 100)
	appendRow := func(row operationEvidenceRow) {
		if selected[row.SourceID] || !operationEvidenceFixedEligible(row) {
			return
		}
		classes := make([]string, 0)
		for _, class := range operationEvidenceClasses {
			if value := row.Classifications[class]; value.Declared && value.Enabled {
				classes = append(classes, class)
			}
		}
		rows = append(rows, operationEvidenceFixedRow{SourceID: row.SourceID, SourceSHA256: row.Source.SHA256, CanonicalMethod: row.Canonical.Method, CanonicalPath: row.Canonical.Path, CLIPaths: append([]string(nil), row.CLI.Paths...), WebsitePaths: append([]string(nil), row.Website.Paths...), FixturePaths: append([]string(nil), row.Fixtures.Paths...), Classifications: classes})
		selected[row.SourceID] = true
	}
	// A sorted, per-capability sample prevents a large family (GitHub's
	// GraphQL writes today) from crowding every other executable surface out
	// of the independent hundred-operation regression cohort.
	for _, class := range operationEvidenceClasses {
		remaining := 20
		for _, row := range artifact.Rows {
			if remaining == 0 {
				break
			}
			classification := row.Classifications[class]
			if classification.Declared && classification.Enabled && operationEvidenceFixedEligible(row) && !selected[row.SourceID] {
				appendRow(row)
				remaining--
			}
		}
	}
	for _, row := range artifact.Rows {
		appendRow(row)
		if len(rows) == 100 {
			break
		}
	}
	if len(rows) != 100 {
		return operationEvidenceFixed100{}, fmt.Errorf("found %d complete source-locked operations, need 100", len(rows))
	}
	return operationEvidenceFixed100{SchemaVersion: operationEvidenceSchemaVersion, Source: "official provider source locks, stratified across available executable capability classes as a fixed 100-operation aggregate cohort", Rows: rows}, nil
}

func operationEvidenceFixedEligible(row operationEvidenceRow) bool {
	if row.Absence != nil || len(row.Gaps) != 0 || !row.Runtime.Enabled || !row.Conformance.Passed || len(row.Fixtures.Paths) == 0 || len(row.CLI.Paths) == 0 || len(row.Website.Paths) == 0 {
		return false
	}
	for _, class := range operationEvidenceClasses {
		if value := row.Classifications[class]; value.Declared && value.Enabled {
			return true
		}
	}
	return false
}

func readOperationEvidenceFixed100(path string) (operationEvidenceFixed100, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return operationEvidenceFixed100{}, err
	}
	var fixed operationEvidenceFixed100
	if err := json.Unmarshal(raw, &fixed); err != nil {
		return operationEvidenceFixed100{}, err
	}
	if fixed.SchemaVersion != operationEvidenceSchemaVersion || len(fixed.Rows) != 100 {
		return operationEvidenceFixed100{}, fmt.Errorf("invalid fixed-100 reference")
	}
	seen := map[string]bool{}
	for _, row := range fixed.Rows {
		if row.SourceID == "" || row.SourceSHA256 == "" || seen[row.SourceID] {
			return operationEvidenceFixed100{}, fmt.Errorf("fixed-100 reference has an invalid or duplicate source row")
		}
		seen[row.SourceID] = true
	}
	return fixed, nil
}

func validateOperationEvidenceFixed100(artifact operationEvidenceArtifact, fixed operationEvidenceFixed100) error {
	if fixed.SchemaVersion != operationEvidenceSchemaVersion || len(fixed.Rows) != 100 {
		return errors.New("fixed-100 reference must contain exactly 100 rows")
	}
	for _, expected := range fixed.Rows {
		actual, found := artifact.row(expected.SourceID)
		if !found {
			return fmt.Errorf("%s is absent from operation evidence", expected.SourceID)
		}
		if actual.Source.SHA256 != expected.SourceSHA256 || actual.Canonical.Method != expected.CanonicalMethod || actual.Canonical.Path != expected.CanonicalPath {
			return fmt.Errorf("%s source or canonical mapping regressed", expected.SourceID)
		}
		if !actual.Runtime.Enabled || len(actual.Gaps) != 0 || !actual.Conformance.Passed || !operationEvidenceIncludesAll(actual.CLI.Paths, expected.CLIPaths) || !operationEvidenceIncludesAll(actual.Website.Paths, expected.WebsitePaths) || !operationEvidenceIncludesAll(actual.Fixtures.Paths, expected.FixturePaths) {
			return fmt.Errorf("%s execution evidence regressed", expected.SourceID)
		}
		for _, class := range expected.Classifications {
			if actual := actual.Classifications[class]; !actual.Declared || !actual.Enabled {
				return fmt.Errorf("%s %s classification regressed", expected.SourceID, class)
			}
		}
	}
	return nil
}

func (artifact operationEvidenceArtifact) row(sourceID string) (operationEvidenceRow, bool) {
	for _, row := range artifact.Rows {
		if row.SourceID == sourceID {
			return row, true
		}
	}
	return operationEvidenceRow{}, false
}

func (artifact operationEvidenceArtifact) rowCount() int { return len(artifact.Rows) }

func (artifact operationEvidenceArtifact) sourceIDCount(sourceID string) int {
	count := 0
	for _, row := range artifact.Rows {
		if row.SourceID == sourceID {
			count++
		}
	}
	return count
}

func (artifact operationEvidenceArtifact) rollup(id string) (operationEvidenceRollup, bool) {
	for _, rollup := range artifact.MissingFoundations {
		if rollup.ID == id {
			return rollup, true
		}
	}
	return operationEvidenceRollup{}, false
}

func (artifact operationEvidenceArtifact) clone() operationEvidenceArtifact {
	raw, _ := json.Marshal(artifact)
	var clone operationEvidenceArtifact
	_ = json.Unmarshal(raw, &clone)
	return clone
}

func (artifact *operationEvidenceArtifact) replace(wanted operationEvidenceRow) {
	for index := range artifact.Rows {
		if artifact.Rows[index].SourceID == wanted.SourceID {
			artifact.Rows[index] = wanted
			return
		}
	}
}

func (row operationEvidenceRow) hasGap(kind string) bool {
	for _, gap := range row.Gaps {
		if gap.Kind == kind {
			return true
		}
	}
	return false
}

func operationEvidenceIncludesAll(actual, expected []string) bool {
	values := operationEvidenceSet(actual)
	for _, value := range expected {
		if !values[value] {
			return false
		}
	}
	return true
}

func operationEvidenceSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func operationEvidenceSortedUnique(values []string) []string {
	set := operationEvidenceSet(values)
	result := make([]string, 0, len(set))
	for value := range set {
		if value != "" {
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}
