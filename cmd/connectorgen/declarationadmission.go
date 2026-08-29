package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/failures"
	"polymetrics.ai/internal/safety"
)

var declarationAdmissionActionTemplateRE = regexp.MustCompile(`\{\{\s*(?:config|record)\.([-A-Za-z0-9_]+)\s*\}\}`)

const (
	declarationAdmissionSchemaVersion          = 2
	declarationAdmissionInventorySchemaVersion = 2

	declarationAdmissionStateImplemented = "implemented"
	declarationAdmissionStateDeferred    = "deferred"
	declarationAdmissionStateUnsupported = "unsupported_with_provider_evidence"

	declarationAdmissionLaneETL            = "etl"
	declarationAdmissionLaneReverseETL     = "reverse_etl"
	declarationAdmissionLaneDirectRead     = "direct_read"
	declarationAdmissionLaneDirectWrite    = "direct_write"
	declarationAdmissionLaneBinaryDownload = "binary_download"
	declarationAdmissionLaneBinaryUpload   = "binary_upload"
)

type declarationAdmissionDocument struct {
	SchemaVersion    int                                   `json:"schema_version"`
	Connector        string                                `json:"connector"`
	SourceOperations []declarationAdmissionSourceOperation `json:"source_operations"`
	Declarations     []declarationAdmissionDeclaration     `json:"declarations"`
}

type declarationAdmissionSourceCatalog struct {
	SchemaVersion    int                                   `json:"schema_version"`
	Cohort           string                                `json:"cohort"`
	SourceOperations []declarationAdmissionSourceOperation `json:"source_operations"`
}

type declarationAdmissionCatalog struct {
	SchemaVersion int                               `json:"schema_version"`
	Cohort        string                            `json:"cohort"`
	Declarations  []declarationAdmissionDeclaration `json:"declarations"`
}

type declarationAdmissionInventory struct {
	SchemaVersion int                                      `json:"schema_version"`
	Cohort        string                                   `json:"cohort"`
	Operations    []declarationAdmissionInventoryOperation `json:"operations"`
}

type declarationAdmissionInventoryOperation struct {
	Connector         string `json:"connector"`
	SourceID          string `json:"source_id"`
	SourceLock        string `json:"source_lock"`
	SourceOperationID string `json:"source_operation_id"`
}

type declarationAdmissionSourceOperation struct {
	Connector           string                      `json:"connector,omitempty"`
	ID                  string                      `json:"id"`
	Protocol            string                      `json:"protocol"`
	SourceURL           string                      `json:"source_url"`
	Location            string                      `json:"location"`
	ProviderOperationID string                      `json:"operation_id"`
	Method              string                      `json:"method"`
	BasePath            string                      `json:"base_path,omitempty"`
	Path                string                      `json:"path"`
	Binding             declarationAdmissionBinding `json:"binding"`
	DestructiveKind     string                      `json:"destructive_kind"`
}

type declarationAdmissionDeclaration struct {
	Connector   string                           `json:"connector,omitempty"`
	SourceID    string                           `json:"source_id"`
	Lane        string                           `json:"lane"`
	Command     string                           `json:"command"`
	State       string                           `json:"state"`
	Canonical   declarationAdmissionEndpoint     `json:"canonical"`
	Binding     declarationAdmissionBinding      `json:"binding"`
	Foundation  *declarationAdmissionFoundation  `json:"foundation_gap,omitempty"`
	Unsupported *declarationAdmissionUnsupported `json:"unsupported_disposition,omitempty"`
	Destructive *declarationAdmissionDestructive `json:"destructive,omitempty"`
}

type declarationAdmissionBinding struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type declarationAdmissionEndpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type declarationAdmissionFoundation struct {
	ID        string                       `json:"id"`
	Reason    string                       `json:"reason"`
	Component string                       `json:"component"`
	Evidence  string                       `json:"evidence"`
	Target    declarationAdmissionEndpoint `json:"target"`
}

type declarationAdmissionDestructive struct {
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
}

type declarationAdmissionUnsupported struct {
	Reason string                                `json:"reason"`
	Target declarationAdmissionUnsupportedTarget `json:"target"`
}

type declarationAdmissionUnsupportedTarget struct {
	SourceID            string `json:"source_id"`
	ProviderOperationID string `json:"operation_id"`
	Method              string `json:"method"`
	Path                string `json:"path"`
}

type declarationAdmissionReport struct {
	Findings          []Finding `json:"findings"`
	ConnectorsChecked int       `json:"connectors_checked"`
	SourceOperations  int       `json:"source_operations"`
}

type declarationAdmissionOptions struct {
	dir    string
	asJSON bool
}

// runDeclarationAdmission checks the required independent source cohort and
// its separate declaration catalog.
// It does not fetch provider material or run a provider operation. It reuses
// commandrunner's no-I/O preflight so executable and deferred declarations are
// checked by the same resolver as the CLI without becoming runtime or live
// certification.
func runDeclarationAdmission(args []string, stdout, stderr io.Writer) int {
	options, err := parseDeclarationAdmissionOptions(args)
	if err != nil {
		logf(stderr, "connectorgen declaration-admission: %v\n", err)
		return 2
	}
	report, err := declarationAdmissionPathCheck(options.dir)
	if err != nil {
		logf(stderr, "connectorgen declaration-admission: %v\n", err)
		return 1
	}
	if options.asJSON {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			logf(stderr, "connectorgen declaration-admission: encode report: %v\n", err)
			return 1
		}
	} else {
		for _, finding := range report.Findings {
			logf(stdout, "%s: %s: %s\n", finding.Connector, finding.File, finding.Message)
		}
		logf(stdout, "connectorgen declaration-admission: %d connector(s), %d source operation(s), %d finding(s)\n", report.ConnectorsChecked, report.SourceOperations, len(report.Findings))
	}
	if len(report.Findings) > 0 {
		return 1
	}
	return 0
}

func parseDeclarationAdmissionOptions(args []string) (declarationAdmissionOptions, error) {
	options := declarationAdmissionOptions{}
	for _, argument := range args[1:] {
		switch argument {
		case "--json":
			options.asJSON = true
		default:
			if strings.HasPrefix(argument, "-") || options.dir != "" {
				return declarationAdmissionOptions{}, fmt.Errorf("unexpected argument %q", argument)
			}
			options.dir = argument
		}
	}
	if options.dir != "" {
		return options, nil
	}
	root, err := repoRoot()
	if err != nil {
		return declarationAdmissionOptions{}, fmt.Errorf("resolve repository root: %w", err)
	}
	options.dir = filepath.Join(root, "internal", "connectors", "defs")
	return options, nil
}

func declarationAdmissionPathCheck(dir string) (declarationAdmissionReport, error) {
	inventoryFile := filepath.Join(dir, "declaration_admission_inventory.json")
	sourceFile := filepath.Join(dir, "declaration_admission_sources.json")
	declarationFile := filepath.Join(dir, "declaration_admissions.json")
	inventoryRaw, err := os.ReadFile(inventoryFile)
	if err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("read required declaration admission inventory %s: %w", inventoryFile, err)
	}
	if err := engine.ValidateDeclarationAdmissionInventory(inventoryRaw); err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("validate declaration admission inventory: %w", err)
	}
	sourceRaw, err := os.ReadFile(sourceFile)
	if err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("read required independent source cohort %s: %w", sourceFile, err)
	}
	if err := engine.ValidateDeclarationAdmissionSources(sourceRaw); err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("validate independent source cohort: %w", err)
	}
	declarationRaw, err := os.ReadFile(declarationFile)
	if err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("read required declaration catalog %s: %w", declarationFile, err)
	}
	if err := engine.ValidateDeclarationAdmission(declarationRaw); err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("validate declaration catalog: %w", err)
	}

	var inventory declarationAdmissionInventory
	if err := decodeSourceStrictJSON(inventoryRaw, &inventory); err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("parse declaration admission inventory: %w", err)
	}
	var sources declarationAdmissionSourceCatalog
	if err := decodeSourceStrictJSON(sourceRaw, &sources); err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("parse independent source cohort: %w", err)
	}
	var declarations declarationAdmissionCatalog
	if err := decodeSourceStrictJSON(declarationRaw, &declarations); err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("parse declaration catalog: %w", err)
	}

	report := declarationAdmissionReport{Findings: []Finding{}, SourceOperations: len(inventory.Operations)}
	addCatalogFinding := func(file, message string) {
		report.Findings = append(report.Findings, Finding{File: file, Rule: "declaration_admission", Message: message})
	}
	if inventory.SchemaVersion != declarationAdmissionInventorySchemaVersion || sources.SchemaVersion != declarationAdmissionSchemaVersion || declarations.SchemaVersion != declarationAdmissionSchemaVersion {
		addCatalogFinding("declaration_admission_inventory.json", "inventory and mutable catalogs must use schema_version 2")
	}
	if strings.TrimSpace(inventory.Cohort) == "" || inventory.Cohort != sources.Cohort || inventory.Cohort != declarations.Cohort {
		addCatalogFinding("declaration_admission_inventory.json", "source and declaration cohorts do not match the independent inventory")
	}

	documents := map[string]*declarationAdmissionDocument{}
	for _, selected := range inventory.Operations {
		if documents[selected.Connector] == nil {
			documents[selected.Connector] = &declarationAdmissionDocument{SchemaVersion: declarationAdmissionSchemaVersion, Connector: selected.Connector}
		}
	}
	for _, source := range sources.SourceOperations {
		document := documents[source.Connector]
		if document == nil {
			document = &declarationAdmissionDocument{SchemaVersion: declarationAdmissionSchemaVersion, Connector: source.Connector}
			documents[source.Connector] = document
		}
		document.SourceOperations = append(document.SourceOperations, source)
	}
	for _, declaration := range declarations.Declarations {
		document := documents[declaration.Connector]
		if document == nil {
			document = &declarationAdmissionDocument{SchemaVersion: declarationAdmissionSchemaVersion, Connector: declaration.Connector}
			documents[declaration.Connector] = document
		}
		document.Declarations = append(document.Declarations, declaration)
	}
	report.Findings = append(report.Findings, declarationAdmissionReviewedSourceFindings(dir, inventory, sources)...)

	connectorNames := make([]string, 0, len(documents))
	for connector := range documents {
		connectorNames = append(connectorNames, connector)
	}
	sort.Strings(connectorNames)
	report.ConnectorsChecked = len(connectorNames)
	sourceFS := os.DirFS(dir)
	for _, connector := range connectorNames {
		bundle, err := engine.Load(withoutCertificationOverlayFS{FS: sourceFS, connector: connector}, connector)
		if err != nil {
			report.Findings = append(report.Findings, declarationAdmissionFinding(connector, declarationFile, "load connector bundle: "+err.Error()))
			continue
		}
		report.Findings = append(report.Findings, declarationAdmissionFindings(bundle, *documents[connector])...)
	}
	sort.Slice(report.Findings, func(i, j int) bool {
		if report.Findings[i].Connector != report.Findings[j].Connector {
			return report.Findings[i].Connector < report.Findings[j].Connector
		}
		if report.Findings[i].File != report.Findings[j].File {
			return report.Findings[i].File < report.Findings[j].File
		}
		return report.Findings[i].Message < report.Findings[j].Message
	})
	return report, nil
}

type declarationAdmissionReviewedOperation struct {
	Protocol            string
	SourceURL           string
	CitationURL         string
	PublishedSourceURL  string
	Location            string
	DocumentID          string
	ProviderOperationID string
	Method              string
	Path                string
	SourceReference     bool
}

// declarationAdmissionReviewedSourceLock is the mapping-only projection of a
// connector-owned source lock. Retention fields deliberately do not survive
// this boundary: declaration admission proves operation provenance, while
// source-import separately proves retained bytes and hashes.
type declarationAdmissionReviewedSourceLock struct {
	SchemaVersion int
	Connector     string
	Operations    map[string]declarationAdmissionReviewedOperation
	Unavailable   []declarationAdmissionReviewedUnavailableDocument
}

type declarationAdmissionReviewedUnavailableDocument struct {
	ID        string
	SourceURL string
	Reason    string
}

type declarationAdmissionSourceLockWire struct {
	SchemaVersion      int                `json:"schema_version"`
	Connector          string             `json:"connector"`
	CapturedAt         json.RawMessage    `json:"captured_at,omitempty"`
	Rest               json.RawMessage    `json:"rest"`
	GraphQL            json.RawMessage    `json:"graphql,omitempty"`
	Counts             sourceImportCounts `json:"counts,omitempty"`
	OperationsFound    json.RawMessage    `json:"operations_found,omitempty"`
	CoverageConfidence json.RawMessage    `json:"coverage_confidence,omitempty"`
	SourceContract     json.RawMessage    `json:"source_contract,omitempty"`
}

type declarationAdmissionSourceArtifactWire struct {
	SourceURL       string          `json:"source_url"`
	SHA256          json.RawMessage `json:"sha256"`
	Bytes           json.RawMessage `json:"bytes"`
	OpenAPI         json.RawMessage `json:"openapi,omitempty"`
	Swagger         json.RawMessage `json:"swagger,omitempty"`
	IdentityQuery   json.RawMessage `json:"identity_query,omitempty"`
	Identity        json.RawMessage `json:"identity,omitempty"`
	CanonicalSHA256 json.RawMessage `json:"canonical_sha256,omitempty"`
}

type declarationAdmissionLegacyRESTWire struct {
	declarationAdmissionSourceArtifactWire
	Commit          json.RawMessage                          `json:"commit,omitempty"`
	InfoVersion     json.RawMessage                          `json:"info_version,omitempty"`
	SourceKind      string                                   `json:"source_kind,omitempty"`
	OperationCounts json.RawMessage                          `json:"operation_counts,omitempty"`
	Supplements     []declarationAdmissionRESTSupplementWire `json:"supplements,omitempty"`
	Operations      []declarationAdmissionRESTOperationWire  `json:"operations,omitempty"`
}

type declarationAdmissionRESTSupplementWire struct {
	declarationAdmissionSourceArtifactWire
	SourceLocation string `json:"source_location"`
	OperationCount int    `json:"operation_count"`
}

type declarationAdmissionRESTOperationWire struct {
	ID              string                                                    `json:"id"`
	Protocol        string                                                    `json:"protocol"`
	Method          string                                                    `json:"method"`
	Path            string                                                    `json:"path"`
	OperationID     string                                                    `json:"operation_id"`
	Deprecated      json.RawMessage                                           `json:"deprecated"`
	SourceLocation  string                                                    `json:"source_location"`
	SourceURL       string                                                    `json:"source_url,omitempty"`
	CitationURL     string                                                    `json:"citation_url,omitempty"`
	CitationBinding *declarationAdmissionRenderedReferenceCitationBindingWire `json:"citation_binding,omitempty"`
	SourceOperation json.RawMessage                                           `json:"source_operation,omitempty"`
}

// declarationAdmissionRenderedReferenceCitationBindingWire retains only the
// operation extraction identity needed by mapping admission. Source import
// separately validates the capture URL, byte count, and digest representation.
type declarationAdmissionRenderedReferenceCitationBindingWire struct {
	CaptureURL     json.RawMessage `json:"capture_url"`
	CaptureSHA256  json.RawMessage `json:"capture_sha256"`
	CaptureBytes   json.RawMessage `json:"capture_bytes"`
	SourceLocation string          `json:"source_location"`
}

type declarationAdmissionIgnoredArtifactWire struct {
	SourceURL       json.RawMessage `json:"source_url"`
	SHA256          json.RawMessage `json:"sha256"`
	Bytes           json.RawMessage `json:"bytes"`
	OpenAPI         json.RawMessage `json:"openapi,omitempty"`
	Swagger         json.RawMessage `json:"swagger,omitempty"`
	IdentityQuery   json.RawMessage `json:"identity_query,omitempty"`
	Identity        json.RawMessage `json:"identity,omitempty"`
	CanonicalSHA256 json.RawMessage `json:"canonical_sha256,omitempty"`
}

type declarationAdmissionPublishedSourceWire struct {
	SourceURL  string          `json:"source_url"`
	CaptureURL json.RawMessage `json:"capture_url"`
	SHA256     json.RawMessage `json:"sha256"`
	Bytes      json.RawMessage `json:"bytes"`
	Adapter    json.RawMessage `json:"adapter"`
}

type declarationAdmissionRESTDocumentWire struct {
	ID                string                                  `json:"id"`
	Kind              string                                  `json:"kind,omitempty"`
	ContentType       json.RawMessage                         `json:"content_type,omitempty"`
	Artifact          declarationAdmissionIgnoredArtifactWire `json:"artifact"`
	SourceReference   *declarationAdmissionSourceArtifactWire `json:"source_reference,omitempty"`
	PublishedSource   declarationAdmissionPublishedSourceWire `json:"published_source"`
	InfoVersion       json.RawMessage                         `json:"info_version,omitempty"`
	UnavailableReason json.RawMessage                         `json:"unavailable_reason,omitempty"`
	Operations        []declarationAdmissionRESTOperationWire `json:"operations"`
}

type declarationAdmissionV3RESTWire struct {
	Retrieval          json.RawMessage                        `json:"retrieval"`
	OpenAPIVersions    json.RawMessage                        `json:"openapi"`
	CoverageConfidence json.RawMessage                        `json:"coverage_confidence,omitempty"`
	SourceDocuments    []declarationAdmissionRESTDocumentWire `json:"source_documents"`
}

type declarationAdmissionGraphQLFieldWire struct {
	Root       string          `json:"root"`
	Name       string          `json:"name"`
	Line       int             `json:"line"`
	Signature  string          `json:"signature"`
	Arguments  json.RawMessage `json:"arguments"`
	ReturnType json.RawMessage `json:"return_type"`
	Deprecated json.RawMessage `json:"deprecated"`
	Preview    json.RawMessage `json:"preview"`
}

type declarationAdmissionGraphQLWire struct {
	declarationAdmissionSourceArtifactWire
	ProjectionSHA256 json.RawMessage                        `json:"projection_sha256,omitempty"`
	ProjectionBytes  json.RawMessage                        `json:"projection_bytes,omitempty"`
	QueryFields      []declarationAdmissionGraphQLFieldWire `json:"query_fields"`
	MutationFields   []declarationAdmissionGraphQLFieldWire `json:"mutation_fields"`
	TypeSystem       json.RawMessage                        `json:"type_system"`
}

func parseDeclarationAdmissionSourceLock(raw []byte, expectedConnector string) (declarationAdmissionReviewedSourceLock, error) {
	var wire declarationAdmissionSourceLockWire
	if err := decodeSourceStrictJSON(raw, &wire); err != nil {
		return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("parse source lock mapping evidence: %w", err)
	}
	if wire.SchemaVersion != 1 && wire.SchemaVersion != 2 && wire.SchemaVersion != 3 {
		return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock has unsupported schema version %d", wire.SchemaVersion)
	}
	if wire.Connector == "" {
		return declarationAdmissionReviewedSourceLock{}, errors.New("source lock has no connector")
	}
	if expectedConnector != "" && wire.Connector != expectedConnector {
		return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock connector %q does not match requested connector %q", wire.Connector, expectedConnector)
	}
	if err := validateSourceImportConnector(wire.Connector); err != nil {
		return declarationAdmissionReviewedSourceLock{}, err
	}

	lock := declarationAdmissionReviewedSourceLock{
		SchemaVersion: wire.SchemaVersion,
		Connector:     wire.Connector,
		Operations:    map[string]declarationAdmissionReviewedOperation{},
	}
	addOperation := func(id string, operation declarationAdmissionReviewedOperation) error {
		if id == "" {
			return errors.New("source lock has incomplete operation identity")
		}
		if _, duplicate := lock.Operations[id]; duplicate {
			return fmt.Errorf("source lock duplicates operation identity %q", id)
		}
		lock.Operations[id] = operation
		return nil
	}

	restCount := 0
	var legacyReference *declarationAdmissionLegacyRESTWire
	if wire.SchemaVersion < 3 {
		var rest declarationAdmissionLegacyRESTWire
		if len(wire.Rest) != 0 {
			if err := decodeSourceStrictJSON(wire.Rest, &rest); err != nil {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("parse source lock REST mapping evidence: %w", err)
			}
		}
		if len(rest.Operations) > 0 && rest.SourceKind == "" {
			if err := validateDeclarationAdmissionMappingSourceURL(rest.SourceURL); err != nil {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock has invalid REST source URL: %w", err)
			}
		}
		referenceSources := map[string]struct{}{}
		if rest.SourceKind != "" {
			if wire.SchemaVersion != 2 {
				return declarationAdmissionReviewedSourceLock{}, errors.New("source-reference locks require schema version 2")
			}
			if rest.SourceKind != sourceImportLegacySourceReferenceKind {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock has unsupported legacy REST source kind %q", rest.SourceKind)
			}
			if err := validateDeclarationAdmissionMappingSourceURL(rest.SourceURL); err != nil {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source-reference lock has invalid REST source URL: %w", err)
			}
			referenceSources[rest.SourceURL] = struct{}{}
			for _, supplement := range rest.Supplements {
				if err := validateDeclarationAdmissionMappingSourceURL(supplement.SourceURL); err != nil {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source-reference supplement has invalid source URL: %w", err)
				}
				if !sourceImportReferenceText(supplement.SourceLocation, 4096) || supplement.OperationCount <= 0 {
					return declarationAdmissionReviewedSourceLock{}, errors.New("source-reference supplement has invalid mapping evidence")
				}
				if _, duplicate := referenceSources[supplement.SourceURL]; duplicate {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source-reference lock duplicates source URL %q", supplement.SourceURL)
				}
				referenceSources[supplement.SourceURL] = struct{}{}
			}
		} else if len(rest.Supplements) != 0 {
			return declarationAdmissionReviewedSourceLock{}, errors.New("source lock supplements require an explicit source-reference kind")
		}
		for _, operation := range rest.Operations {
			sourceReference := rest.SourceKind != ""
			validateOperation := validateDeclarationAdmissionMappingRESTOperation
			if sourceReference {
				validateOperation = validateDeclarationAdmissionMappingReferenceRESTOperation
			}
			if err := validateOperation(operation); err != nil {
				return declarationAdmissionReviewedSourceLock{}, err
			}
			if operation.CitationBinding != nil {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock legacy REST operation %q declares a rendered-reference citation binding", operation.ID)
			}
			if sourceReference && operation.CitationURL != "" {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source-reference operation %q must not declare a citation URL", operation.ID)
			}
			if operation.CitationURL != "" {
				if err := validateDeclarationAdmissionMappingSourceURL(operation.CitationURL); err != nil {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock REST operation %q has invalid citation URL: %w", operation.ID, err)
				}
			}
			sourceURL := rest.SourceURL
			if sourceReference {
				if operation.SourceURL == "" {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source-reference operation %q has no source URL", operation.ID)
				}
				if _, declared := referenceSources[operation.SourceURL]; !declared {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source-reference operation %q cites an undeclared source URL", operation.ID)
				}
				sourceURL = operation.SourceURL
			} else if operation.SourceURL != "" {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock REST operation %q declares a reference-only source URL", operation.ID)
			}
			restCount++
			if err := addOperation(operation.ID, declarationAdmissionReviewedOperation{
				Protocol: operation.Protocol, SourceURL: sourceURL, Location: operation.SourceLocation,
				ProviderOperationID: operation.OperationID, Method: operation.Method, Path: operation.Path,
				SourceReference: sourceReference,
			}); err != nil {
				return declarationAdmissionReviewedSourceLock{}, err
			}
		}
		if rest.SourceKind != "" {
			legacyReference = &rest
		}
	} else {
		var rest declarationAdmissionV3RESTWire
		if len(wire.Rest) != 0 {
			if err := decodeSourceStrictJSON(wire.Rest, &rest); err != nil {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("parse source lock v3 REST mapping evidence: %w", err)
			}
		}
		openAPIVersions, err := validateDeclarationAdmissionMappingV3Envelope(rest)
		if err != nil {
			return declarationAdmissionReviewedSourceLock{}, err
		}
		seenDocuments := map[string]struct{}{}
		seenRoutes := map[string]string{}
		openAPIDocuments := 0
		requiresCoverageConfidence := false
		for index, document := range rest.SourceDocuments {
			if document.ID == "" || document.ID != strings.TrimSpace(document.ID) || document.ID != strings.ToLower(document.ID) || !sourceImportDocumentID(document.ID) {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock has invalid v3 REST document ID %q", document.ID)
			}
			if _, duplicate := seenDocuments[document.ID]; duplicate {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock duplicates v3 REST document ID %q", document.ID)
			}
			if index > 0 && rest.SourceDocuments[index-1].ID >= document.ID {
				return declarationAdmissionReviewedSourceLock{}, errors.New("source lock v3 REST source documents are not sorted")
			}
			seenDocuments[document.ID] = struct{}{}
			kind := document.Kind
			if kind == "" {
				kind = sourceImportDocumentKindOpenAPI
			}
			switch kind {
			case sourceImportDocumentKindOpenAPI, sourceImportDocumentKindRenderedReference, sourceImportDocumentKindBundle:
				if document.SourceReference != nil {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 REST document %q kind %q declares a source reference", document.ID, kind)
				}
				form, err := declarationAdmissionMappingArtifactForm(document.Artifact.OpenAPI, document.Artifact.Swagger)
				if err != nil {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 REST document %q has invalid source form: %w", document.ID, err)
				}
				switch kind {
				case sourceImportDocumentKindOpenAPI:
					if !form.isSwagger2() {
						openAPIDocuments++
						if !form.isOpenAPI() || !openAPIVersions[form.Version] {
							return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 REST document %q has an OpenAPI version outside the aggregate inventory", document.ID)
						}
					}
				case sourceImportDocumentKindRenderedReference, sourceImportDocumentKindBundle:
					requiresCoverageConfidence = true
					if form.Family != "" {
						return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 REST document %q kind %q must not declare an OpenAPI or Swagger version", document.ID, kind)
					}
					contentType, err := declarationAdmissionMappingString(document.ContentType)
					if err != nil || validateSourceImportDocumentContentType(contentType) != nil {
						return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 REST document %q has invalid content type", document.ID)
					}
					if kind == sourceImportDocumentKindBundle && !sourceImportBundleContentType(contentType) {
						return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 REST bundle document %q must declare an archive content type", document.ID)
					}
				}
			case sourceImportDocumentKindSourceReference:
				requiresCoverageConfidence = true
				if err := validateDeclarationAdmissionMappingSourceReferenceDocument(document); err != nil {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 source-reference document %q: %w", document.ID, err)
				}
				if len(document.Operations) == 0 {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 source-reference document %q has no operations", document.ID)
				}
			case sourceImportDocumentKindUnavailable:
				requiresCoverageConfidence = true
				if document.SourceReference != nil {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 unavailable document %q declares a source reference", document.ID)
				}
				if len(document.Operations) != 0 {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 unavailable document %q must not declare operations", document.ID)
				}
				reason, err := declarationAdmissionMappingString(document.UnavailableReason)
				if err != nil || !sourceImportReferenceText(reason, 1024) {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 unavailable document %q has invalid reason", document.ID)
				}
				if document.PublishedSource.SourceURL != "" {
					if err := validateDeclarationAdmissionMappingSourceURL(document.PublishedSource.SourceURL); err != nil {
						return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 unavailable document %q has invalid source URL: %w", document.ID, err)
					}
				}
				lock.Unavailable = append(lock.Unavailable, declarationAdmissionReviewedUnavailableDocument{
					ID:        document.ID,
					SourceURL: document.PublishedSource.SourceURL,
					Reason:    reason,
				})
				continue
			default:
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 REST document %q has unsupported kind %q", document.ID, kind)
			}
			if len(document.Operations) == 0 && kind == sourceImportDocumentKindOpenAPI {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 REST document %q has no operations", document.ID)
			}
			for _, operation := range document.Operations {
				sourceReference := kind == sourceImportDocumentKindSourceReference
				validateOperation := validateDeclarationAdmissionMappingRESTOperation
				if sourceReference {
					validateOperation = validateDeclarationAdmissionMappingReferenceRESTOperation
				}
				if err := validateOperation(operation); err != nil {
					return declarationAdmissionReviewedSourceLock{}, err
				}
				sourceURL := ""
				citationURL := ""
				publishedSourceURL := ""
				if sourceReference {
					if operation.CitationURL != "" || operation.SourceURL != "" || operation.CitationBinding != nil {
						return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 source-reference operation %q must inherit its document citation", operation.ID)
					}
					sourceURL = document.SourceReference.SourceURL
				} else {
					if operation.SourceURL != "" {
						return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 REST operation %q declares a reference-only source URL", operation.ID)
					}
					if kind == sourceImportDocumentKindRenderedReference {
						canonicalCitationBase, err := validateDeclarationAdmissionRenderedReferenceCitation(operation, document)
						if err != nil {
							return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 rendered-reference operation %q has invalid citation: %w", operation.ID, err)
						}
						sourceURL = canonicalCitationBase
						citationURL = operation.CitationURL
						publishedSourceURL = document.PublishedSource.SourceURL
					} else {
						if operation.CitationBinding != nil {
							return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 REST operation %q declares a rendered-reference citation binding for kind %q", operation.ID, kind)
						}
						if operation.CitationURL != "" {
							if err := validateDeclarationAdmissionMappingSourceURL(operation.CitationURL); err != nil {
								return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock REST operation %q has invalid citation URL: %w", operation.ID, err)
							}
						}
						sourceURL = operation.CitationURL
						if sourceURL == "" {
							sourceURL = document.PublishedSource.SourceURL
						}
					}
				}
				if kind != sourceImportDocumentKindRenderedReference {
					if err := validateDeclarationAdmissionMappingSourceURL(sourceURL); err != nil {
						return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock operation %q has invalid provider source URL: %w", operation.ID, err)
					}
				}
				route := strings.ToUpper(operation.Method) + "\x00" + operation.Path
				if previous, duplicate := seenRoutes[route]; duplicate {
					return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock v3 REST route %s %s occurs in both %q and %q", operation.Method, operation.Path, previous, document.ID)
				}
				seenRoutes[route] = document.ID
				restCount++
				if err := addOperation(operation.ID, declarationAdmissionReviewedOperation{
					Protocol: operation.Protocol, SourceURL: sourceURL, CitationURL: citationURL, PublishedSourceURL: publishedSourceURL,
					Location: operation.SourceLocation, DocumentID: document.ID,
					ProviderOperationID: operation.OperationID, Method: operation.Method, Path: operation.Path,
					SourceReference: sourceReference,
				}); err != nil {
					return declarationAdmissionReviewedSourceLock{}, err
				}
			}
		}
		if openAPIDocuments > 0 && len(openAPIVersions) == 0 {
			return declarationAdmissionReviewedSourceLock{}, errors.New("source lock has no v3 REST OpenAPI versions")
		}
		if openAPIDocuments == 0 && len(openAPIVersions) > 0 {
			return declarationAdmissionReviewedSourceLock{}, errors.New("source lock v3 REST OpenAPI versions require an OpenAPI document")
		}
		coverageConfidence, err := declarationAdmissionMappingCoverageConfidence(rest.CoverageConfidence)
		if err != nil {
			return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock has invalid v3 REST coverage confidence: %w", err)
		}
		if requiresCoverageConfidence && coverageConfidence == nil {
			return declarationAdmissionReviewedSourceLock{}, errors.New("source lock v3 REST non-OpenAPI documents require coverage confidence")
		}
		if coverageConfidence != nil {
			if err := validateSourceImportCoverageConfidence(*coverageConfidence); err != nil {
				return declarationAdmissionReviewedSourceLock{}, fmt.Errorf("source lock has invalid v3 REST coverage confidence: %w", err)
			}
		}
	}

	graphqlQueryCount, graphqlMutationCount, err := declarationAdmissionMappingGraphQLOperations(wire.GraphQL, wire.Connector, addOperation)
	if err != nil {
		return declarationAdmissionReviewedSourceLock{}, err
	}
	if legacyReference != nil {
		if err := validateDeclarationAdmissionMappingLegacySourceReference(*legacyReference, wire, graphqlQueryCount, graphqlMutationCount); err != nil {
			return declarationAdmissionReviewedSourceLock{}, err
		}
	}
	if wire.Counts.REST != restCount || wire.Counts.GraphQLQuery != graphqlQueryCount || wire.Counts.GraphQLMutation != graphqlMutationCount || wire.Counts.Total != restCount+graphqlQueryCount+graphqlMutationCount {
		return declarationAdmissionReviewedSourceLock{}, errors.New("source lock counts do not match mapping operation inventories")
	}
	return lock, nil
}

func validateDeclarationAdmissionMappingSourceURL(sourceURL string) error {
	canonical, err := safety.CanonicalProviderCitationURL(sourceURL)
	if err != nil {
		return err
	}
	if canonical != sourceURL {
		return errors.New("provider source URL is not canonical")
	}
	return nil
}

func validateDeclarationAdmissionMappingRESTOperation(operation declarationAdmissionRESTOperationWire) error {
	if operation.ID == "" || operation.ID != strings.TrimSpace(operation.ID) || operation.Protocol != "rest" || operation.Method == "" || operation.Method != strings.TrimSpace(operation.Method) || operation.Path == "" || operation.SourceLocation == "" || operation.SourceLocation != strings.TrimSpace(operation.SourceLocation) || operation.OperationID != strings.TrimSpace(operation.OperationID) {
		return fmt.Errorf("source lock has incomplete REST operation identity %q", operation.ID)
	}
	if err := validateSourceImportPath(operation.Path); err != nil {
		return fmt.Errorf("source lock REST operation %q has invalid path: %w", operation.ID, err)
	}
	return nil
}

// validateDeclarationAdmissionMappingReferenceRESTOperation deliberately keeps
// the closed identity contract for cited-only source references. Mapping may
// ignore retained bytes and hashes, but a reference has no later request
// schema from which to repair an ambiguous method, route, ID, or location.
func validateDeclarationAdmissionMappingReferenceRESTOperation(operation declarationAdmissionRESTOperationWire) error {
	reference := sourceImportRESTOperation{
		ID:             operation.ID,
		Protocol:       operation.Protocol,
		Method:         operation.Method,
		Path:           operation.Path,
		OperationID:    operation.OperationID,
		SourceLocation: operation.SourceLocation,
	}
	if err := validateSourceImportReferenceOperation(reference); err != nil {
		return fmt.Errorf("source-reference operation %q %w", operation.ID, err)
	}
	return nil
}

func validateDeclarationAdmissionMappingLegacySourceReference(rest declarationAdmissionLegacyRESTWire, wire declarationAdmissionSourceLockWire, graphqlQueryCount, graphqlMutationCount int) error {
	if rest.SourceKind != sourceImportLegacySourceReferenceKind {
		return fmt.Errorf("source lock has unsupported legacy REST source kind %q", rest.SourceKind)
	}
	if len(rest.Operations) == 0 {
		return errors.New("source-reference lock has no operation inventory")
	}
	if graphqlQueryCount != 0 || graphqlMutationCount != 0 || wire.Counts.GraphQLQuery != 0 || wire.Counts.GraphQLMutation != 0 {
		return errors.New("source-reference lock cannot declare a GraphQL inventory")
	}
	confidence, err := declarationAdmissionMappingCoverageConfidence(wire.CoverageConfidence)
	if err != nil || confidence == nil || confidence.Level != rest.SourceKind {
		return errors.New("source-reference lock coverage confidence must repeat its source kind")
	}
	if err := validateSourceImportCoverageConfidence(*confidence); err != nil {
		return fmt.Errorf("source-reference lock has invalid coverage confidence: %w", err)
	}
	operationsFound, err := declarationAdmissionMappingCounts(wire.OperationsFound)
	if err != nil || operationsFound != wire.Counts {
		return errors.New("source-reference lock operations_found does not match counts")
	}
	operationCounts, err := declarationAdmissionMappingOperationCounts(rest.OperationCounts)
	if err != nil || len(operationCounts) == 0 {
		return errors.New("source-reference lock has no operation counts")
	}
	if wire.Counts.REST != len(rest.Operations) || wire.Counts.Total != wire.Counts.REST {
		return errors.New("source-reference lock counts do not match REST operation inventory")
	}

	actualCounts := make(map[string]int, len(operationCounts))
	seenRoutes := make(map[string]struct{}, len(rest.Operations))
	for _, operation := range rest.Operations {
		actualCounts[operation.Method]++
		route := operation.Method + "\x00" + operation.Path
		if _, duplicate := seenRoutes[route]; duplicate {
			return fmt.Errorf("source-reference lock duplicates REST route %s %s", operation.Method, operation.Path)
		}
		seenRoutes[route] = struct{}{}
	}
	for _, supplement := range rest.Supplements {
		bound := 0
		for _, operation := range rest.Operations {
			if operation.SourceURL != supplement.SourceURL {
				continue
			}
			if operation.SourceLocation != supplement.SourceLocation {
				return fmt.Errorf("source-reference operation %q citation location %q does not match supplement %q location %q", operation.ID, operation.SourceLocation, supplement.SourceURL, supplement.SourceLocation)
			}
			bound++
		}
		if bound != supplement.OperationCount {
			return fmt.Errorf("source-reference supplement %q operation count %d does not match %d operations", supplement.SourceURL, supplement.OperationCount, bound)
		}
	}
	if len(actualCounts) != len(operationCounts) {
		return errors.New("source-reference lock operation counts do not match operation inventory")
	}
	for method, count := range operationCounts {
		if !sourceImportReferenceHTTPMethod(method) || count <= 0 || actualCounts[method] != count {
			return fmt.Errorf("source-reference lock operation count for %q does not match operation inventory", method)
		}
	}
	return nil
}

func validateDeclarationAdmissionMappingV3Envelope(rest declarationAdmissionV3RESTWire) (map[string]bool, error) {
	retrieval, err := declarationAdmissionMappingString(rest.Retrieval)
	if err != nil || retrieval == "" || retrieval != strings.TrimSpace(retrieval) || len(retrieval) > 1024 || strings.ContainsAny(retrieval, "\r\n") {
		return nil, errors.New("source lock has invalid v3 REST retrieval metadata")
	}
	if len(rest.SourceDocuments) == 0 {
		return nil, errors.New("source lock has no v3 REST source documents")
	}
	if len(rest.SourceDocuments) > defaultSourceImportDocuments {
		return nil, fmt.Errorf("source lock v3 document count exceeds %d", defaultSourceImportDocuments)
	}
	versions, err := declarationAdmissionMappingStringSlice(rest.OpenAPIVersions)
	if err != nil {
		return nil, fmt.Errorf("source lock has invalid v3 REST OpenAPI versions: %w", err)
	}
	if len(versions) > 0 && !sort.StringsAreSorted(versions) {
		return nil, errors.New("source lock v3 REST OpenAPI versions are not sorted")
	}
	seen := make(map[string]bool, len(versions))
	for _, version := range versions {
		major, minor, ok := sourceOpenAPIMajorMinor(version)
		if !ok || major != 3 || (minor != 0 && minor != 1) || seen[version] {
			return nil, fmt.Errorf("source lock has invalid or duplicate v3 REST OpenAPI version %q", version)
		}
		seen[version] = true
	}
	return seen, nil
}

func validateDeclarationAdmissionMappingSourceReferenceDocument(document declarationAdmissionRESTDocumentWire) error {
	if document.SourceReference == nil {
		return errors.New("source reference is required")
	}
	contentType, contentTypeErr := declarationAdmissionMappingString(document.ContentType)
	unavailableReason, unavailableReasonErr := declarationAdmissionMappingString(document.UnavailableReason)
	if declarationAdmissionMappingArtifactDeclaresSourceForm(document.Artifact) || declarationAdmissionMappingPublishedSourceDeclaresPublication(document.PublishedSource) || contentTypeErr != nil || unavailableReasonErr != nil || contentType != "" || unavailableReason != "" {
		return errors.New("source reference cannot mix with retained artifact, publication, content type, or unavailable capture fields")
	}
	if err := validateDeclarationAdmissionMappingSourceURL(document.SourceReference.SourceURL); err != nil {
		return fmt.Errorf("has invalid source URL: %w", err)
	}
	if _, err := declarationAdmissionMappingArtifactForm(document.SourceReference.OpenAPI, document.SourceReference.Swagger); err != nil {
		return fmt.Errorf("has invalid source form: %w", err)
	}
	return nil
}

// declarationAdmissionMappingArtifactDeclaresSourceForm distinguishes a
// second source form from malformed retention representation. A cited-only
// document cannot also declare an artifact URL or form pin, but mapping does
// not reject an otherwise-empty sibling merely because its hash/byte capture
// fields are absent or malformed.
func declarationAdmissionMappingArtifactDeclaresSourceForm(artifact declarationAdmissionIgnoredArtifactWire) bool {
	sourceURL, sourceURLErr := declarationAdmissionMappingString(artifact.SourceURL)
	if sourceURLErr != nil || sourceURL != "" {
		return true
	}
	form, formErr := declarationAdmissionMappingArtifactForm(artifact.OpenAPI, artifact.Swagger)
	return formErr != nil || form.Family != ""
}

func declarationAdmissionMappingPublishedSourceDeclaresPublication(source declarationAdmissionPublishedSourceWire) bool {
	return source.SourceURL != ""
}

func declarationAdmissionMappingArtifactForm(openAPIRaw, swaggerRaw json.RawMessage) (sourceDocumentForm, error) {
	openAPI, err := declarationAdmissionMappingString(openAPIRaw)
	if err != nil {
		return sourceDocumentForm{}, err
	}
	swagger, err := declarationAdmissionMappingString(swaggerRaw)
	if err != nil {
		return sourceDocumentForm{}, err
	}
	if openAPI != "" && swagger != "" {
		return sourceDocumentForm{}, errors.New("source lock has ambiguous OpenAPI and Swagger form pins")
	}
	if openAPI != "" {
		major, minor, ok := sourceOpenAPIMajorMinor(openAPI)
		if !ok || major != 3 || (minor != 0 && minor != 1) {
			return sourceDocumentForm{}, fmt.Errorf("source lock has unsupported OpenAPI form pin %q", openAPI)
		}
		return sourceDocumentForm{Family: sourceImportDocumentKindOpenAPI, Version: openAPI}, nil
	}
	if swagger != "" {
		if swagger != "2.0" {
			return sourceDocumentForm{}, fmt.Errorf("source lock has unsupported Swagger form pin %q", swagger)
		}
		return sourceDocumentForm{Family: "swagger", Version: swagger}, nil
	}
	return sourceDocumentForm{}, nil
}

func declarationAdmissionMappingString(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "null" {
		return "", nil
	}
	var value string
	if err := decodeSourceStrictJSON(raw, &value); err != nil {
		return "", err
	}
	return value, nil
}

func declarationAdmissionMappingStringSlice(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "null" {
		return nil, nil
	}
	var values []string
	if err := decodeSourceStrictJSON(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func declarationAdmissionMappingCoverageConfidence(raw json.RawMessage) (*sourceImportCoverageConfidence, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "null" {
		return nil, nil
	}
	var confidence sourceImportCoverageConfidence
	if err := decodeSourceStrictJSON(raw, &confidence); err != nil {
		return nil, err
	}
	return &confidence, nil
}

func declarationAdmissionMappingCounts(raw json.RawMessage) (sourceImportCounts, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "null" {
		return sourceImportCounts{}, errors.New("missing counts")
	}
	var counts sourceImportCounts
	if err := decodeSourceStrictJSON(raw, &counts); err != nil {
		return sourceImportCounts{}, err
	}
	return counts, nil
}

func declarationAdmissionMappingOperationCounts(raw json.RawMessage) (map[string]int, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "null" {
		return nil, errors.New("missing operation counts")
	}
	var counts map[string]int
	if err := decodeSourceStrictJSON(raw, &counts); err != nil {
		return nil, err
	}
	return counts, nil
}

func validateDeclarationAdmissionRenderedReferenceCitation(operation declarationAdmissionRESTOperationWire, document declarationAdmissionRESTDocumentWire) (string, error) {
	if err := validateDeclarationAdmissionMappingSourceURL(document.PublishedSource.SourceURL); err != nil {
		return "", fmt.Errorf("published source URL: %w", err)
	}
	raw := operation.CitationURL
	if raw == "" || raw != strings.TrimSpace(raw) || strings.ContainsAny(raw, "\r\n") {
		return "", errors.New("citation must be a non-empty canonical absolute HTTPS URL")
	}
	citation, err := url.Parse(raw)
	if err != nil || !citation.IsAbs() || !strings.EqualFold(citation.Scheme, "https") || citation.Host == "" || citation.User != nil {
		return "", errors.New("citation must be a canonical absolute HTTPS URL without userinfo")
	}
	published, err := url.Parse(document.PublishedSource.SourceURL)
	if err != nil || !sourceImportURLsShareOrigin(citation, published) {
		return "", errors.New("citation must use the published-source origin")
	}
	canonicalCitation := *citation
	canonicalCitation.Fragment = ""
	canonicalCitation.RawFragment = ""
	canonicalBase, err := safety.CanonicalProviderCitationURL(canonicalCitation.String())
	if err != nil || canonicalBase != canonicalCitation.String() {
		return "", errors.New("citation base URL is not canonical")
	}
	if operation.CitationBinding != nil {
		if operation.CitationBinding.SourceLocation == "" || operation.CitationBinding.SourceLocation != operation.SourceLocation {
			return "", errors.New("citation capture extraction binding does not match the locked operation location")
		}
	}
	if fragment := strings.TrimSpace(citation.Fragment); fragment != "" {
		if fragment != strings.TrimPrefix(operation.SourceLocation, "#") {
			return "", fmt.Errorf("citation fragment %q does not match locked operation extraction location %q", fragment, operation.SourceLocation)
		}
		return canonicalBase, nil
	}
	if operation.CitationBinding == nil {
		return "", errors.New("citation must include an operation-specific fragment or capture extraction binding")
	}
	return canonicalBase, nil
}

func declarationAdmissionMappingGraphQLOperations(raw json.RawMessage, connector string, addOperation func(string, declarationAdmissionReviewedOperation) error) (int, int, error) {
	var graphql declarationAdmissionGraphQLWire
	if len(raw) != 0 {
		if err := decodeSourceStrictJSON(raw, &graphql); err != nil {
			return 0, 0, fmt.Errorf("parse source lock GraphQL mapping evidence: %w", err)
		}
	}
	queryCount := len(graphql.QueryFields)
	mutationCount := len(graphql.MutationFields)
	if queryCount+mutationCount > 0 {
		if err := validateDeclarationAdmissionMappingSourceURL(graphql.SourceURL); err != nil {
			return 0, 0, fmt.Errorf("source lock has invalid GraphQL source URL: %w", err)
		}
	}
	for _, group := range []struct {
		kind   string
		root   string
		fields []declarationAdmissionGraphQLFieldWire
	}{{"query", "Query", graphql.QueryFields}, {"mutation", "Mutation", graphql.MutationFields}} {
		for _, field := range group.fields {
			if field.Root != group.root || field.Name == "" || field.Name != strings.TrimSpace(field.Name) || field.Line <= 0 || field.Signature == "" || field.Signature != strings.TrimSpace(field.Signature) {
				return 0, 0, fmt.Errorf("source lock has incomplete GraphQL root identity %q", group.root+"."+field.Name)
			}
			id := fmt.Sprintf("%s.graphql.%s.%s", connector, group.kind, field.Name)
			if err := addOperation(id, declarationAdmissionReviewedOperation{
				Protocol: "graphql", SourceURL: graphql.SourceURL,
				Location:            fmt.Sprintf("graphql.%s_fields[%q]@line:%d", group.kind, field.Name, field.Line),
				ProviderOperationID: field.Root + "." + field.Name, Method: "GRAPHQL", Path: field.Name,
			}); err != nil {
				return 0, 0, err
			}
		}
	}
	return queryCount, mutationCount, nil
}

func declarationAdmissionReviewedSourceFindings(dir string, inventory declarationAdmissionInventory, sources declarationAdmissionSourceCatalog) []Finding {
	findings := []Finding{}
	add := func(connector, message string) {
		findings = append(findings, declarationAdmissionFinding(connector, "declaration_admission_inventory.json", message))
	}
	rows := make(map[string]declarationAdmissionSourceOperation, len(sources.SourceOperations))
	for _, source := range sources.SourceOperations {
		key := source.Connector + "\x00" + source.ID
		if _, duplicate := rows[key]; duplicate {
			continue
		}
		rows[key] = source
	}
	selectedRows := make(map[string]struct{}, len(inventory.Operations))
	selectedOperations := make(map[string]string, len(inventory.Operations))
	lockCache := map[string]declarationAdmissionReviewedSourceLock{}
	for _, selected := range inventory.Operations {
		rowKey := selected.Connector + "\x00" + selected.SourceID
		if _, duplicate := selectedRows[rowKey]; duplicate {
			add(selected.Connector, fmt.Sprintf("inventory duplicates source identity %q", selected.SourceID))
			continue
		}
		selectedRows[rowKey] = struct{}{}
		selectionKey := selected.Connector + "\x00" + selected.SourceLock + "\x00" + selected.SourceOperationID
		if previous, duplicate := selectedOperations[selectionKey]; duplicate {
			add(selected.Connector, fmt.Sprintf("inventory source identities %q and %q select one reviewed source operation", previous, selected.SourceID))
			continue
		}
		selectedOperations[selectionKey] = selected.SourceID
		row, exists := rows[rowKey]
		if !exists {
			add(selected.Connector, fmt.Sprintf("inventory source operation %q has no compact source row", selected.SourceID))
			continue
		}
		lockPath, err := declarationAdmissionOwnedSourceLockPath(dir, selected)
		if err != nil {
			add(selected.Connector, fmt.Sprintf("source operation %q has invalid reviewed source lock: %v", selected.SourceID, err))
			continue
		}
		lock, cached := lockCache[lockPath]
		if !cached {
			raw, readErr := os.ReadFile(lockPath)
			if readErr != nil {
				add(selected.Connector, fmt.Sprintf("source operation %q cannot read reviewed source lock: %v", selected.SourceID, readErr))
				continue
			}
			lock, err = parseDeclarationAdmissionSourceLock(raw, selected.Connector)
			if err != nil {
				add(selected.Connector, fmt.Sprintf("source operation %q has invalid reviewed source lock: %v", selected.SourceID, err))
				continue
			}
			lockCache[lockPath] = lock
		}
		reviewed, found := declarationAdmissionReviewedOperationFromLock(lock, selected.SourceOperationID)
		if !found {
			add(selected.Connector, fmt.Sprintf("source operation %q selects nonexistent reviewed operation %q", selected.SourceID, selected.SourceOperationID))
			continue
		}
		if message := declarationAdmissionReviewedOperationMismatch(row, reviewed); message != "" {
			add(selected.Connector, fmt.Sprintf("source operation %q does not exactly match reviewed operation %q: %s", selected.SourceID, selected.SourceOperationID, message))
		}
	}
	for _, source := range sources.SourceOperations {
		if _, selected := selectedRows[source.Connector+"\x00"+source.ID]; !selected {
			add(source.Connector, fmt.Sprintf("compact source operation %q is absent from the independent inventory", source.ID))
		}
	}
	return findings
}

func declarationAdmissionOwnedSourceLockPath(dir string, selected declarationAdmissionInventoryOperation) (string, error) {
	raw := selected.SourceLock
	if raw == "" || filepath.IsAbs(raw) || strings.Contains(raw, "\\") || filepath.ToSlash(filepath.Clean(filepath.FromSlash(raw))) != raw {
		return "", errors.New("source_lock must be one canonical relative path")
	}
	prefix := selected.Connector + "/sources/"
	if !strings.HasPrefix(raw, prefix) || !strings.HasSuffix(raw, "-operation-source-lock.json") {
		return "", fmt.Errorf("source_lock must be owned beneath %s", prefix)
	}
	root, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", fmt.Errorf("resolve defs root: %w", err)
	}
	path, err := filepath.EvalSymlinks(filepath.Join(dir, filepath.FromSlash(raw)))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("source_lock resolves outside the defs root")
	}
	if !strings.HasPrefix(filepath.ToSlash(relative), prefix) {
		return "", fmt.Errorf("source_lock resolves outside connector-owned %s", prefix)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("source_lock must resolve to one regular file")
	}
	return path, nil
}

func declarationAdmissionReviewedOperationFromLock(lock declarationAdmissionReviewedSourceLock, operationID string) (declarationAdmissionReviewedOperation, bool) {
	operation, found := lock.Operations[operationID]
	return operation, found
}

func declarationAdmissionReviewedOperationMismatch(row declarationAdmissionSourceOperation, reviewed declarationAdmissionReviewedOperation) string {
	if row.Protocol != reviewed.Protocol {
		return fmt.Sprintf("protocol %q does not equal %q", row.Protocol, reviewed.Protocol)
	}
	if row.SourceURL != reviewed.SourceURL {
		return "source URL is not the reviewed document URL"
	}
	if row.Location != reviewed.Location {
		return "document location is not the exact reviewed operation location"
	}
	if row.ProviderOperationID != reviewed.ProviderOperationID {
		return "provider operation identity differs"
	}
	if row.BasePath != "" || !strings.EqualFold(row.Method, reviewed.Method) || row.Path != reviewed.Path {
		return "method/path is not the exact reviewed operation endpoint"
	}
	return ""
}

func declarationAdmissionFindings(bundle engine.Bundle, document declarationAdmissionDocument) []Finding {
	findings := []Finding{}
	file := "declaration_admissions.json"
	add := func(message string) {
		findings = append(findings, Finding{Connector: bundle.Name, File: file, Rule: "declaration_admission", Message: message})
	}
	if document.SchemaVersion != declarationAdmissionSchemaVersion {
		add(fmt.Sprintf("schema_version = %d, want %d", document.SchemaVersion, declarationAdmissionSchemaVersion))
	}
	if document.Connector != bundle.Name {
		add(fmt.Sprintf("connector %q does not match bundle %q", document.Connector, bundle.Name))
	}
	if len(document.SourceOperations) == 0 {
		add("source declaration has no operations")
	}
	sources := make(map[string]declarationAdmissionSourceOperation, len(document.SourceOperations))
	identities := make(map[string]string, len(document.SourceOperations))
	bindings := make(map[string]string, len(document.SourceOperations))
	for _, source := range document.SourceOperations {
		if source.Connector != "" && source.Connector != bundle.Name {
			add(fmt.Sprintf("source operation %q belongs to connector %q", source.ID, source.Connector))
			continue
		}
		if source.ID == "" {
			add("source operation has no identity")
			continue
		}
		if _, duplicate := sources[source.ID]; duplicate {
			add("duplicate source operation " + source.ID)
			continue
		}
		sources[source.ID] = source
		canonicalSourceURL, citationErr := safety.CanonicalProviderCitationURL(source.SourceURL)
		if citationErr != nil {
			add("source operation " + source.ID + " has no valid provider source URL")
		} else if canonicalSourceURL != source.SourceURL {
			add("source operation " + source.ID + " has no canonical provider source URL")
		}
		if strings.TrimSpace(source.Location) == "" {
			add("source operation " + source.ID + " has no exact provider operation citation")
		}
		if source.ProviderOperationID != strings.TrimSpace(source.ProviderOperationID) {
			add("source operation " + source.ID + " has a noncanonical provider operation identity")
		}
		if source.Protocol != "rest" && source.Protocol != "graphql" {
			add("source operation " + source.ID + " has an invalid protocol")
		}
		if !declarationAdmissionBindingValid(source.Binding) {
			add("source operation " + source.ID + " has no exact canonical binding")
		}
		switch source.DestructiveKind {
		case "none", "delete", "destructive":
		default:
			add("source operation " + source.ID + " has an invalid destructive semantic")
		}
		_, _, err := declarationAdmissionCanonicalSourceEndpoint(source)
		if err != nil {
			add("source operation " + source.ID + ": " + err.Error())
			continue
		}
		if canonicalSourceURL != "" {
			identity := declarationAdmissionSourceIdentity(source, canonicalSourceURL)
			if previous, duplicate := identities[identity]; duplicate {
				add(fmt.Sprintf("duplicate exact provider operation identity for source operations %s and %s", previous, source.ID))
			} else {
				identities[identity] = source.ID
			}
		}
		binding := strings.Join([]string{source.Binding.Kind, source.Binding.ID}, "\x00")
		if previous, duplicate := bindings[binding]; duplicate {
			add(fmt.Sprintf("source operations %s and %s claim one canonical binding", previous, source.ID))
		} else {
			bindings[binding] = source.ID
		}
	}
	declared := make(map[string]bool, len(document.Declarations))
	for _, declaration := range document.Declarations {
		if declaration.Connector != "" && declaration.Connector != bundle.Name {
			add(fmt.Sprintf("declaration for source operation %q belongs to connector %q", declaration.SourceID, declaration.Connector))
			continue
		}
		source, exists := sources[declaration.SourceID]
		if declaration.SourceID == "" || !exists {
			add("declaration references unknown source operation " + declaration.SourceID)
			continue
		}
		if declared[declaration.SourceID] {
			add("duplicate declaration for source operation " + declaration.SourceID)
			continue
		}
		declared[declaration.SourceID] = true
		declarationAdmissionCheckRow(&findings, bundle, file, source, declaration)
	}
	for sourceID := range sources {
		if !declared[sourceID] {
			add("source operation " + sourceID + " has no declaration")
		}
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].Message < findings[j].Message })
	return findings
}

// declarationAdmissionSourceIdentity is the compact-row provenance key after
// declarationAdmissionReviewedSourceFindings has bound that row to the exact
// operation selected from its connector-owned reviewed source lock.
func declarationAdmissionSourceIdentity(source declarationAdmissionSourceOperation, canonicalSourceURL string) string {
	method, path, err := declarationAdmissionCanonicalSourceEndpoint(source)
	if err != nil {
		return ""
	}
	return strings.Join([]string{
		canonicalSourceURL,
		source.Location,
		source.Protocol,
		source.ProviderOperationID,
		method,
		path,
	}, "\x00")
}

func declarationAdmissionCheckRow(findings *[]Finding, bundle engine.Bundle, file string, source declarationAdmissionSourceOperation, declaration declarationAdmissionDeclaration) {
	add := func(message string) {
		*findings = append(*findings, Finding{Connector: bundle.Name, File: file, Rule: "declaration_admission", Message: "source operation " + source.ID + ": " + message})
	}
	if !declarationAdmissionLaneValid(declaration.Lane) {
		add("lane is not one of the admission lanes")
		return
	}
	canonicalMethod, effectivePath, err := declarationAdmissionCanonicalSourceEndpoint(source)
	if err != nil {
		add(err.Error())
		return
	}
	if !strings.EqualFold(declaration.Canonical.Method, canonicalMethod) || declaration.Canonical.Path != effectivePath {
		add("base-path mismatch or stale canonical endpoint")
	}
	if !declarationAdmissionBindingValid(declaration.Binding) || !declarationAdmissionBindingsMatch(declaration.Binding, source.Binding) {
		add("canonical binding does not match the independent source operation")
	}
	switch source.DestructiveKind {
	case "none":
		if declaration.Destructive != nil {
			if declaration.Destructive.Kind == "delete" {
				add("delete semantics do not match the independent source semantic")
			} else {
				add("destructive metadata does not match the independent source semantic")
			}
		}
	case "delete", "destructive":
		if declaration.Destructive == nil {
			add(source.DestructiveKind + " operation lacks destructive metadata")
		} else if declaration.Destructive.Kind != source.DestructiveKind {
			add("delete semantics do not match the independent source semantic")
		} else if strings.TrimSpace(declaration.Destructive.Reason) == "" {
			add("destructive metadata lacks a reason")
		}
	}
	commandPath, commandPathErr := commandrunner.CommandPathSegments(declaration.Command)
	if declaration.Command == "" {
		add("has no discoverable command mapping")
		return
	}
	if commandPathErr != nil {
		add("has no canonical round-trippable command path: " + commandPathErr.Error())
		return
	}
	command, matches := declarationAdmissionCommand(bundle, declaration.Command)
	if matches != 1 {
		add("has no unique discoverable command mapping")
		return
	}
	if declarationAdmissionLaneForIntent(command.Intent) != declaration.Lane {
		add("lane does not match command intent")
	}
	if !declarationAdmissionCommandCitesEndpoint(command, declaration.Canonical) || !declarationAdmissionSurfaceHasEndpoint(bundle.Surface, declaration.Canonical) {
		add("command does not map to the canonical API surface endpoint")
	}
	switch declaration.State {
	case declarationAdmissionStateImplemented:
		if declaration.Foundation != nil || command.Foundation != nil || declaration.Unsupported != nil || command.Unsupported != nil {
			add("implemented declaration must not retain a foundation gap or unsupported disposition")
		}
		resolved, resolveErr := engine.ResolveImplementedCommandPath(bundle, declaration.Command)
		if command.Availability != declarationAdmissionStateImplemented || resolveErr != nil ||
			!strings.EqualFold(resolved.Method, declaration.Canonical.Method) || resolved.Path != declaration.Canonical.Path ||
			!declarationAdmissionRuntimeBindingMatches(resolved.Binding, source.Binding) {
			add("implemented declaration has no valid runtime binding")
		} else if err := commandrunner.Preflight(engine.New(bundle, nil), commandPath); err != nil {
			add("implemented declaration fails runtime preflight: " + err.Error())
		}
		if !declarationAdmissionResolvedDestructiveMatches(resolved, source.DestructiveKind) {
			add("implemented destructive declaration does not retain destructive runtime metadata")
		}
	case declarationAdmissionStateDeferred:
		if declaration.Unsupported != nil || command.Unsupported != nil {
			add("deferred declaration must not retain an unsupported disposition")
		}
		if !declarationAdmissionFoundationSpecific(declaration.Foundation) {
			add("deferred declaration requires a specific missing implementation component with evidence")
		} else if message := declarationAdmissionFoundationEvidenceFinding(bundle, declaration, declaration.Foundation); message != "" {
			add(message)
		}
		if !declarationAdmissionFoundationMatches(declaration.Foundation, command.Foundation) {
			add("deferred declaration requires the same named foundation gap on its command")
		}
		if !declarationAdmissionFoundationTargetMatchesSource(command.Foundation, source) {
			add("deferred command foundation target does not retain the exact source identity and binding")
		}
		if declaration.Foundation != nil && !declarationAdmissionEndpointsMatch(declaration.Foundation.Target, declaration.Canonical) {
			add("deferred declaration foundation target does not match the canonical API surface endpoint")
		}
		if command.Availability != declarationAdmissionStateDeferred {
			add("deferred declaration command is not deferred")
		}
		if message := declarationAdmissionDeferredPreflight(bundle, commandPath); message != "" {
			add(message)
		}
	case declarationAdmissionStateUnsupported:
		if declaration.Foundation != nil || command.Foundation != nil {
			add("provider-evidenced unsupported declaration must not claim a missing foundation")
		}
		if source.Binding.Kind != connectors.CommandBindingCommand || source.Binding.ID != declaration.Command ||
			declaration.Binding.Kind != connectors.CommandBindingCommand || declaration.Binding.ID != declaration.Command {
			add("provider-evidenced unsupported declaration may bind only its discoverable command projection")
		}
		if command.Stream != "" || command.Write != "" || command.Operation != "" {
			add("provider-evidenced unsupported command must not claim an executable target")
		}
		if !declarationAdmissionUnsupportedMatches(declaration.Unsupported, command.Unsupported, source) {
			add("provider-evidenced unsupported disposition does not retain the exact source target and reason")
		}
		if command.Availability != declarationAdmissionStateUnsupported {
			add("provider-evidenced unsupported declaration command has the wrong availability")
		}
		err := commandrunner.Preflight(engine.New(bundle, nil), commandPath)
		var blocked *commandrunner.BlockedCommandError
		if !errors.As(err, &blocked) || blocked.Failure == nil || blocked.Failure.Code() != "provider_evidenced_unsupported" {
			add("provider-evidenced unsupported command does not return its typed terminal refusal")
		}
	default:
		add("state must be implemented, deferred, or provider-evidenced unsupported")
	}
}

func declarationAdmissionUnsupportedMatches(declaration *declarationAdmissionUnsupported, command *engine.CommandUnsupportedDisposition, source declarationAdmissionSourceOperation) bool {
	if declaration == nil || command == nil || strings.TrimSpace(declaration.Reason) == "" || declaration.Reason != command.Reason {
		return false
	}
	method, path, err := declarationAdmissionCanonicalSourceEndpoint(source)
	if err != nil {
		return false
	}
	want := declarationAdmissionUnsupportedTarget{
		SourceID: source.ID, ProviderOperationID: source.ProviderOperationID, Method: method, Path: path,
	}
	return declaration.Target == want && command.Target.SourceID == want.SourceID &&
		command.Target.ProviderOperationID == want.ProviderOperationID && strings.EqualFold(command.Target.Method, want.Method) && command.Target.Path == want.Path
}

func declarationAdmissionFinding(connector, filename, message string) Finding {
	return Finding{Connector: connector, File: filepath.ToSlash(filepath.Join("sources", filepath.Base(filename))), Rule: "declaration_admission", Message: message}
}

func declarationAdmissionLaneValid(lane string) bool {
	switch lane {
	case declarationAdmissionLaneETL, declarationAdmissionLaneReverseETL, declarationAdmissionLaneDirectRead, declarationAdmissionLaneDirectWrite, declarationAdmissionLaneBinaryDownload, declarationAdmissionLaneBinaryUpload:
		return true
	default:
		return false
	}
}

func declarationAdmissionLaneForIntent(intent string) string {
	switch intent {
	case declarationAdmissionLaneETL, declarationAdmissionLaneReverseETL, declarationAdmissionLaneDirectRead, declarationAdmissionLaneDirectWrite, declarationAdmissionLaneBinaryDownload, declarationAdmissionLaneBinaryUpload:
		return intent
	default:
		return ""
	}
}

func declarationAdmissionEffectivePath(basePath, operationPath string) (string, error) {
	if !strings.HasPrefix(operationPath, "/") {
		return "", errors.New("source path must be a connector-relative absolute path")
	}
	effective := operationPath
	if basePath == "" || basePath == "/" {
		if err := engine.ValidateCommandEndpoint("GET", effective); err != nil {
			return "", errors.New("source path must be a canonical connector-relative absolute path")
		}
		return effective, nil
	}
	if err := engine.ValidateCommandEndpoint("GET", basePath); err != nil {
		return "", errors.New("base path must be a canonical connector-relative absolute path")
	}
	effective = strings.TrimRight(basePath, "/") + operationPath
	if err := engine.ValidateCommandEndpoint("GET", effective); err != nil {
		return "", errors.New("source path must be a canonical connector-relative absolute path")
	}
	return effective, nil
}

func declarationAdmissionCanonicalSourceEndpoint(source declarationAdmissionSourceOperation) (string, string, error) {
	method := strings.ToUpper(strings.TrimSpace(source.Method))
	if source.Protocol == "graphql" && method == "GRAPHQL" {
		if source.BasePath != "" && source.BasePath != "/" {
			return "", "", errors.New("GraphQL operation identity must not declare an HTTP base path")
		}
		if err := engine.ValidateCommandEndpoint(method, source.Path); err != nil {
			return "", "", err
		}
		return method, source.Path, nil
	}
	if source.Protocol == "rest" && method == "GRAPHQL" {
		return "", "", errors.New("REST source operation cannot declare a GraphQL operation identity")
	}
	path, err := declarationAdmissionEffectivePath(source.BasePath, source.Path)
	if err != nil {
		return "", "", err
	}
	if err := engine.ValidateCommandEndpoint(method, path); err != nil {
		return "", "", err
	}
	return method, path, nil
}

func declarationAdmissionCommand(bundle engine.Bundle, path string) (engine.CLICommand, int) {
	if bundle.CLISurface == nil {
		return engine.CLICommand{}, 0
	}
	var command engine.CLICommand
	count := 0
	for _, candidate := range bundle.CLISurface.Commands {
		if candidate.Path == path {
			command = candidate
			count++
		}
	}
	return command, count
}

func declarationAdmissionCommandCitesEndpoint(command engine.CLICommand, endpoint declarationAdmissionEndpoint) bool {
	return len(command.APISurface) == 1 && strings.EqualFold(command.APISurface[0].Method, endpoint.Method) && command.APISurface[0].Path == endpoint.Path
}

func declarationAdmissionSurfaceHasEndpoint(surface *engine.APISurface, endpoint declarationAdmissionEndpoint) bool {
	if surface == nil {
		return false
	}
	matches := 0
	for _, candidate := range surface.Endpoints {
		if strings.EqualFold(candidate.Method, endpoint.Method) && candidate.Path == endpoint.Path {
			matches++
			if candidate.Excluded != nil || (candidate.Operation != nil && candidate.Operation.Model == "disallowed") {
				return false
			}
		}
	}
	return matches == 1
}

func declarationAdmissionFoundationMatches(declaration *declarationAdmissionFoundation, command *engine.CommandFoundation) bool {
	return declarationAdmissionFoundationSpecific(declaration) && command != nil && declaration.ID == command.ID && declaration.Reason == command.Reason &&
		declaration.Component == command.Component && declaration.Evidence == command.Evidence &&
		declarationAdmissionEndpointsMatch(declaration.Target, declarationAdmissionEndpoint{Method: command.Target.Method, Path: command.Target.Path})
}

func declarationAdmissionFoundationSpecific(foundation *declarationAdmissionFoundation) bool {
	return foundation != nil && strings.TrimSpace(foundation.ID) != "" && strings.TrimSpace(foundation.Reason) != "" &&
		declarationAdmissionFoundationComponentValid(foundation.Component) && declarationAdmissionFoundationEvidenceValid(foundation.Component, foundation.Evidence) &&
		strings.TrimSpace(foundation.Target.Method) != "" && strings.TrimSpace(foundation.Target.Path) != ""
}

// declarationAdmissionFoundationComponentValid accepts missing runtime pieces,
// never a provider policy, operation kind, source-retention choice, or a live
// certification state. Those facts can explain risk, but cannot remove a
// cited source operation from the declaration and command surface.
func declarationAdmissionFoundationComponentValid(component string) bool {
	return connectors.ValidCommandFoundationComponent(component)
}

func declarationAdmissionFoundationEvidenceValid(component, evidence string) bool {
	return connectors.ValidCommandFoundationEvidence(component, evidence)
}

func declarationAdmissionFoundationEvidenceFinding(bundle engine.Bundle, declaration declarationAdmissionDeclaration, foundation *declarationAdmissionFoundation) string {
	if foundation.Component != "typed_write_action" {
		return ""
	}
	if declaration.Lane != declarationAdmissionLaneReverseETL && declaration.Lane != declarationAdmissionLaneBinaryUpload {
		return "typed_write_action foundation does not apply to this lane"
	}
	for _, action := range bundle.Writes {
		if strings.EqualFold(action.Method, declaration.Canonical.Method) && declarationAdmissionActionPath(action.Path) == declaration.Canonical.Path {
			return "typed_write_action foundation is stale: a declared write action already maps to the canonical endpoint"
		}
	}
	return ""
}

func declarationAdmissionActionPath(path string) string {
	return declarationAdmissionActionTemplateRE.ReplaceAllString(path, "{$1}")
}

func declarationAdmissionEndpointsMatch(left, right declarationAdmissionEndpoint) bool {
	return strings.EqualFold(left.Method, right.Method) && left.Path == right.Path
}

func declarationAdmissionBindingValid(binding declarationAdmissionBinding) bool {
	if binding.ID == "" || binding.ID != strings.TrimSpace(binding.ID) {
		return false
	}
	switch binding.Kind {
	case connectors.CommandBindingCommand, connectors.CommandBindingStream, connectors.CommandBindingWrite, connectors.CommandBindingOperation:
		return true
	default:
		return false
	}
}

func declarationAdmissionBindingsMatch(left, right declarationAdmissionBinding) bool {
	return left.Kind == right.Kind && left.ID == right.ID
}

func declarationAdmissionRuntimeBindingMatches(runtime connectors.CommandBindingIdentity, source declarationAdmissionBinding) bool {
	return runtime.Kind == source.Kind && runtime.ID == source.ID
}

func declarationAdmissionFoundationTargetMatchesSource(foundation *engine.CommandFoundation, source declarationAdmissionSourceOperation) bool {
	if foundation == nil {
		return false
	}
	method, effectivePath, err := declarationAdmissionCanonicalSourceEndpoint(source)
	if err != nil {
		return false
	}
	target := foundation.Target
	return target.SourceID == source.ID && target.ProviderOperationID == source.ProviderOperationID &&
		target.Binding.Kind == source.Binding.Kind && target.Binding.ID == source.Binding.ID &&
		target.DestructiveKind == source.DestructiveKind && strings.EqualFold(target.Method, method) && target.Path == effectivePath
}

func declarationAdmissionResolvedDestructiveMatches(resolved engine.ResolvedCommandBinding, semantic string) bool {
	switch semantic {
	case "none":
		return !resolved.Destructive.RequiresApproval()
	case "delete":
		return strings.EqualFold(resolved.Destructive.MutationClass, "delete") && resolved.Destructive.RequiresApproval()
	case "destructive":
		return resolved.Destructive.RequiresApproval()
	default:
		return false
	}
}

func declarationAdmissionDeferredPreflight(bundle engine.Bundle, path []string) string {
	err := commandrunner.Preflight(engine.New(bundle, nil), path)
	var blocked *commandrunner.BlockedCommandError
	if !errors.As(err, &blocked) {
		return fmt.Sprintf("deferred target does not pass runtime preflight: %v", err)
	}
	if blocked.Failure == nil || blocked.Failure.Domain() != failures.DomainSystem || blocked.Failure.Code() != "missing_foundation" {
		return fmt.Sprintf("deferred target does not resolve to system/missing_foundation: %v", blocked)
	}
	return ""
}
