package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"polymetrics.ai/internal/connectors/engine"
)

const (
	batchMaterializeReportSchemaVersion   = 1
	maxMaterializeArtifactBytes           = 16 << 20
	materializeAvailabilityNotImplemented = "not_implemented"
	materializeNamedDependencyPrefix      = "named_dependency="
	materializeSurfaceDiscrepancy         = "present-in-surface-absent-from-artifact"
)

// BatchMaterializeReport records exactly which manifest candidates received a
// v2 provenance inventory. Like batch gate, it records every failure rather
// than aborting the rest of the batch at the first artifact mismatch.
type BatchMaterializeReport struct {
	SchemaVersion int                        `json:"schema_version"`
	Manifest      string                     `json:"manifest"`
	Candidates    int                        `json:"candidates"`
	Included      []BatchMaterializeIncluded `json:"included"`
	Dropped       []BatchGateDrop            `json:"dropped"`
}

// BatchMaterializeIncluded is the evidence retained for one generated bundle.
// The artifact version comes from the immutable survey manifest; v2's shared
// artifact table carries URL, retrieval date, and digest and is referenced by
// every endpoint-local provenance row.
type BatchMaterializeIncluded struct {
	Connector                string                `json:"connector"`
	EvidenceMode             string                `json:"evidence_mode,omitempty"`
	Artifact                 BatchMaterialArtifact `json:"artifact"`
	ArtifactOperations       int                   `json:"artifact_operations"`
	DeclaredOperations       int                   `json:"declared_operations"`
	OperationSplit           BatchOperationSplit   `json:"operation_split"`
	CLICommands              int                   `json:"cli_commands"`
	ImplementedCommands      int                   `json:"implemented_commands"`
	NamedDependencyCommands  int                   `json:"named_dependency_commands"`
	FlaggedDiscrepancies     int                   `json:"flagged_discrepancies"`
	RuntimePreflightCommands int                   `json:"runtime_preflight_commands"`
	OperationExecutors       int                   `json:"operation_executors"`
}

// BatchMaterialArtifact combines the survey's immutable version with the
// newly fetched public artifact evidence. The shared v2 schema intentionally
// owns URL/date/digest only; this report preserves the provider version rather
// than inventing another shared provenance field.
type BatchMaterialArtifact struct {
	Kind        string `json:"kind"`
	URL         string `json:"url"`
	Version     string `json:"version"`
	RetrievedAt string `json:"retrieved_at"`
	SHA256      string `json:"sha256"`
}

type batchMaterializeOptions struct {
	manifestPath            string
	defsRoot                string
	sourceDefsRoot          string
	artifactDir             string
	retrievedAt             string
	reportPath              string
	connectors              []string
	existingSurfaceEvidence bool
	cachedReferencesOnly    bool
}

// runBatchMaterialize is the one post-#3869 artifact-to-bundle authoring
// path. It reads public OpenAPI/Swagger documents from the manifest, upgrades
// the existing surface inventory to v2 provenance, and maps every documented
// operation into the bundle. It deliberately does not run the repository-wide
// runtime preflight: materialization is the cheap authoring phase, and batch
// gate performs the single end-of-run preflight over the staged result.
// Commands are marked implemented only when the real runtime can preflight
// them; unsupported operation metadata remains visible as a not_implemented
// command with a named dependency.
func runBatchMaterialize(args []string, stdout, stderr io.Writer) int {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			logln(stdout, batchUsage())
			return 0
		}
	}
	opts, err := parseBatchMaterializeOptions(args)
	if err != nil {
		logf(stderr, "connectorgen batch materialize: %v\n%s\n", err, batchUsage())
		return 2
	}
	manifest, err := readBatchManifest(opts.manifestPath)
	if err != nil {
		logf(stderr, "connectorgen batch materialize: %v\n", err)
		return 1
	}
	candidates, err := selectedManifestCandidates(manifest, opts.connectors)
	if err != nil {
		logf(stderr, "connectorgen batch materialize: %v\n", err)
		return 1
	}

	report := BatchMaterializeReport{
		SchemaVersion: batchMaterializeReportSchemaVersion,
		Manifest:      opts.manifestPath,
		Candidates:    len(candidates),
		Included:      []BatchMaterializeIncluded{},
		Dropped:       []BatchGateDrop{},
	}
	for _, candidate := range candidates {
		included, drop := materializeBatchCandidate(opts, candidate)
		if drop != nil {
			report.Dropped = append(report.Dropped, *drop)
			continue
		}
		report.Included = append(report.Included, included)
	}
	if err := writeBatchMaterializeReport(opts.reportPath, report); err != nil {
		logf(stderr, "connectorgen batch materialize: write report: %v\n", err)
		return 1
	}

	logf(stdout, "connectorgen batch materialize: %d connector(s) materialized, %d dropped; report %s\n",
		len(report.Included), len(report.Dropped), opts.reportPath)
	if len(report.Dropped) > 0 {
		return 1
	}
	return 0
}

func parseBatchMaterializeOptions(args []string) (batchMaterializeOptions, error) {
	opts := batchMaterializeOptions{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--existing-surface-evidence" {
			opts.existingSurfaceEvidence = true
			continue
		}
		if arg == "--cached-references-only" {
			opts.cachedReferencesOnly = true
			continue
		}
		if i+1 >= len(args) {
			return batchMaterializeOptions{}, fmt.Errorf("%s requires a value", arg)
		}
		value := args[i+1]
		i++
		switch arg {
		case "--manifest":
			opts.manifestPath = value
		case "--defs-root":
			opts.defsRoot = value
		case "--source-defs-root":
			opts.sourceDefsRoot = value
		case "--artifact-dir":
			opts.artifactDir = value
		case "--retrieved-at":
			opts.retrievedAt = value
		case "--report":
			opts.reportPath = value
		case "--connector":
			opts.connectors = append(opts.connectors, value)
		default:
			return batchMaterializeOptions{}, fmt.Errorf("unknown flag %q", arg)
		}
	}
	if opts.manifestPath == "" {
		return batchMaterializeOptions{}, errors.New("--manifest is required")
	}
	if opts.sourceDefsRoot == "" {
		return batchMaterializeOptions{}, errors.New("--source-defs-root is required")
	}
	if opts.retrievedAt == "" {
		return batchMaterializeOptions{}, errors.New("--retrieved-at is required")
	}
	if parsed, err := time.Parse(time.DateOnly, opts.retrievedAt); err != nil || parsed.Format(time.DateOnly) != opts.retrievedAt {
		return batchMaterializeOptions{}, fmt.Errorf("--retrieved-at must be an ISO full date")
	}
	if opts.reportPath == "" {
		return batchMaterializeOptions{}, errors.New("--report is required")
	}
	if opts.cachedReferencesOnly && opts.artifactDir == "" {
		return batchMaterializeOptions{}, errors.New("--cached-references-only requires --artifact-dir")
	}
	if opts.defsRoot == "" {
		root, err := repoRoot()
		if err != nil {
			return batchMaterializeOptions{}, fmt.Errorf("default defs root: %w", err)
		}
		opts.defsRoot = filepath.Join(root, "internal", "connectors", "defs")
	}
	return opts, nil
}

func selectedManifestCandidates(manifest BatchManifest, names []string) ([]BatchManifestConnector, error) {
	if len(names) == 0 {
		return append([]BatchManifestConnector(nil), manifest.Connectors...), nil
	}
	wanted := make(map[string]bool, len(names))
	for _, name := range names {
		if err := validateBatchConnectorName(name); err != nil {
			return nil, err
		}
		if wanted[name] {
			return nil, fmt.Errorf("connector %q was selected more than once", name)
		}
		wanted[name] = true
	}
	selected := make([]BatchManifestConnector, 0, len(names))
	for _, candidate := range manifest.Connectors {
		if wanted[candidate.Connector] {
			selected = append(selected, candidate)
			delete(wanted, candidate.Connector)
		}
	}
	if len(wanted) > 0 {
		missing := make([]string, 0, len(wanted))
		for name := range wanted {
			missing = append(missing, name)
		}
		sort.Strings(missing)
		return nil, fmt.Errorf("selected connector(s) absent from manifest: %s", strings.Join(missing, ", "))
	}
	return selected, nil
}

func materializeBatchCandidate(opts batchMaterializeOptions, candidate BatchManifestConnector) (BatchMaterializeIncluded, *BatchGateDrop) {
	normalizedCandidate, err := materializeUnversionedOfficialReferenceCandidate(opts, candidate)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "existing_surface_evidence", err)
	}
	candidate = normalizedCandidate
	sourceBundleDir, err := batchBundleDirectory(opts.sourceDefsRoot, candidate.Connector)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "source_bundle", err)
	}
	bundleDir, err := batchBundleDirectory(opts.defsRoot, candidate.Connector)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "bundle_path", err)
	}
	if err := rejectBatchMaterializePathOverlap(sourceBundleDir, bundleDir); err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "bundle_path", err)
	}
	if err := rejectBatchMaterializeDestination(bundleDir); err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "bundle_collision", err)
	}
	if !isBundleDir(sourceBundleDir) {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "source_bundle", errors.New("metadata.json is required for a materialization source bundle"))
	}
	sourceDefsRoot, err := filepath.Abs(opts.sourceDefsRoot)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "source_bundle", err)
	}
	bundle, err := engine.Load(os.DirFS(sourceDefsRoot), candidate.Connector)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "source_bundle", err)
	}
	if bundle.Surface == nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "api_surface", errors.New("api_surface.json is required before materialization"))
	}

	rawArtifact, err := readBatchMaterializeArtifact(opts, candidate)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, batchArtifactDropStage(err, "artifact_fetch"), err)
	}
	sha := fmt.Sprintf("%x", sha256.Sum256(rawArtifact))
	var artifactInventory batchArtifactInventory
	if opts.existingSurfaceEvidence {
		artifactInventory, err = existingSurfaceEvidenceInventory(bundle, candidate, opts.retrievedAt, sha, rawArtifact)
		if err != nil {
			return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "existing_surface_evidence", err)
		}
	} else {
		artifactInventory, err = parseBatchManifestArtifact(opts, candidate, rawArtifact)
		if err != nil {
			return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, batchArtifactDropStage(err, "artifact_parse"), err)
		}
	}
	surface, err := materializeAPISurface(bundle, candidate, opts.retrievedAt, sha, artifactInventory.Endpoints, artifactInventory.Sources)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "coverage", err)
	}
	if opts.existingSurfaceEvidence {
		surface.Scope = existingSurfaceEvidenceScope(bundle.Surface.Scope, candidate, candidate.OperationsTotal)
	}
	operations, err := materializeOperationCatalog(bundle, surface, candidate)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "operations", err)
	}
	cli, err := materializeCLISurface(bundle, surface, candidate, operations)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "cli_surface", err)
	}
	split, err := batchSurfaceSplit(&surface)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "api_surface", err)
	}

	surfaceRaw, err := json.MarshalIndent(surface, "", "  ")
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "write", fmt.Errorf("encode api_surface.json: %w", err))
	}
	cliRaw, err := json.MarshalIndent(cli, "", "  ")
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "write", fmt.Errorf("encode cli_surface.json: %w", err))
	}
	operationsRaw, err := json.MarshalIndent(struct {
		Operations []engine.OperationSpec `json:"operations"`
	}{Operations: operations}, "", "  ")
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "write", fmt.Errorf("encode operations.json: %w", err))
	}
	operationsRaw = append(operationsRaw, '\n')
	destination, err := createBatchMaterializeDestination(sourceBundleDir, bundleDir)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, batchMaterializeDestinationStage(err), err)
	}
	if err := writeBatchMaterializedFiles(bundleDir, append(surfaceRaw, '\n'), operationsRaw, append(cliRaw, '\n')); err != nil {
		destination.discard()
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "write", fmt.Errorf("write materialized bundle: %w", err))
	}

	implementedCommands, namedDependencyCommands := materializedCLICommandCounts(cli)
	flaggedDiscrepancies := materializedDiscrepancyCount(surface)
	return BatchMaterializeIncluded{
		Connector: candidate.Connector,
		EvidenceMode: func() string {
			if opts.existingSurfaceEvidence {
				return "existing_surface_evidence"
			}
			return ""
		}(),
		Artifact: BatchMaterialArtifact{
			Kind:        candidate.Artifact.Kind,
			URL:         candidate.Artifact.URL,
			Version:     candidate.Artifact.Version,
			RetrievedAt: opts.retrievedAt,
			SHA256:      sha,
		},
		ArtifactOperations:       len(artifactInventory.Endpoints),
		DeclaredOperations:       split.total(),
		OperationSplit:           split,
		CLICommands:              len(cli.Commands),
		ImplementedCommands:      implementedCommands,
		NamedDependencyCommands:  namedDependencyCommands,
		FlaggedDiscrepancies:     flaggedDiscrepancies,
		RuntimePreflightCommands: 0,
		OperationExecutors:       len(operations),
	}, nil
}

// materializeUnversionedOfficialReferenceCandidate keeps the manifest's
// ordinary version requirement intact while giving a provider-published
// unversioned HTML reference one constrained route. The direct-evidence path
// below separately proves the source surface exactly equals the immutable
// operation count before any generated JSON is written.
func materializeUnversionedOfficialReferenceCandidate(opts batchMaterializeOptions, candidate BatchManifestConnector) (BatchManifestConnector, error) {
	if !candidate.Artifact.UnversionedOfficialReference {
		return candidate, nil
	}
	if !opts.existingSurfaceEvidence {
		return BatchManifestConnector{}, errors.New("declared unversioned official reference requires --existing-surface-evidence")
	}
	candidate.Artifact.Version = batchNoVersionMarker
	return candidate, nil
}

// existingSurfaceEvidenceInventory is the explicit direct-authoring route for
// a connector whose preserved source surface has already been exhaustively
// counted against the immutable provider survey. The cited official artifact
// is still fetched and hashed; this route merely avoids pretending that an
// index page is an OpenAPI document or replaying an unrelated documentation
// crawl. A mismatch is refused rather than filling gaps from the old surface.
func existingSurfaceEvidenceInventory(bundle engine.Bundle, candidate BatchManifestConnector, retrievedAt, sha string, rawArtifact []byte) (batchArtifactInventory, error) {
	if len(rawArtifact) == 0 {
		return batchArtifactInventory{}, errors.New("cited official artifact is empty")
	}
	if bundle.Surface == nil {
		return batchArtifactInventory{}, errors.New("api_surface.json is required")
	}
	if len(bundle.Surface.Endpoints) != candidate.OperationsTotal {
		return batchArtifactInventory{}, fmt.Errorf("existing source surface has %d operation(s), not the immutable ledger count %d", len(bundle.Surface.Endpoints), candidate.OperationsTotal)
	}
	endpoints := make([]batchArtifactEndpoint, 0, len(bundle.Surface.Endpoints))
	seen := make(map[string]bool, len(bundle.Surface.Endpoints))
	for _, endpoint := range bundle.Surface.Endpoints {
		method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
		path := strings.TrimSpace(endpoint.Path)
		if method == "" || path == "" {
			return batchArtifactInventory{}, fmt.Errorf("existing source surface contains an incomplete endpoint %q %q", endpoint.Method, endpoint.Path)
		}
		key := batchArtifactEndpointKey(method, path)
		if seen[key] {
			continue
		}
		seen[key] = true
		endpoints = append(endpoints, batchArtifactEndpoint{
			Method:           method,
			Path:             path,
			Summary:          fmt.Sprintf("%s %s", method, path),
			SourceURL:        candidate.Artifact.URL,
			SourceKind:       candidate.Artifact.Kind,
			SourceVersion:    candidate.Artifact.Version,
			SourceRetrieved:  retrievedAt,
			SourceSHA256:     sha,
			SourceCoordinate: fmt.Sprintf("existing-complete-surface[%s %s]", method, path),
		})
	}
	return batchArtifactInventory{
		Endpoints: endpoints,
		Sources: []batchArtifactSource{{
			URL:       candidate.Artifact.URL,
			Kind:      candidate.Artifact.Kind,
			Version:   candidate.Artifact.Version,
			Retrieved: retrievedAt,
			SHA256:    sha,
		}},
	}, nil
}

func existingSurfaceEvidenceScope(previous string, candidate BatchManifestConnector, operations int) string {
	scope := fmt.Sprintf("Exact existing-surface evidence fallback: the preserved source inventory contains %d official-survey operation entries, exactly matching the immutable count. Shared executable bindings for the same provider method/path are merged into one normalized endpoint. The cited official %s artifact is retained as the v2 provenance root; no operation was inferred from a documentation crawl.", operations, candidate.Artifact.Kind)
	if candidate.Artifact.UnversionedOfficialReference {
		scope += " The provider publishes no version marker; provenance records that documented absence."
	}
	if prior := strings.TrimSpace(previous); prior != "" {
		scope += " Prior source-surface scope: " + prior
	}
	return scope
}

func parseBatchManifestArtifact(opts batchMaterializeOptions, candidate BatchManifestConnector, raw []byte) (batchArtifactInventory, error) {
	source := batchArtifactSource{
		URL:       candidate.Artifact.URL,
		Kind:      candidate.Artifact.Kind,
		Version:   candidate.Artifact.Version,
		Retrieved: opts.retrievedAt,
		SHA256:    fmt.Sprintf("%x", sha256.Sum256(raw)),
	}
	fetch := batchArtifactSourceFetcher(opts, candidate)
	if strings.EqualFold(strings.TrimSpace(source.Kind), "html_reference") {
		if rootInventory, complete, rootErr := completeBatchHTMLReferenceRoot(candidate, raw, source); rootErr == nil && complete {
			return requireBatchArtifactInventoryTotal(candidate, rootInventory)
		}
	}
	budget := newBatchArtifactReferenceBudget(raw)
	primary, primaryErr := parseBatchArtifactByKindWithBudget(raw, source, fetch, budget)
	if primaryErr == nil {
		if candidate.ProviderReferenceURL != "" && len(primary.Endpoints) < candidate.OperationsTotal && candidate.ProviderReferenceURL != candidate.Artifact.URL {
			fallback, fallbackErr := parseBatchManifestReferenceFallback(candidate, source, fetch, budget)
			if fallbackErr != nil {
				return batchArtifactInventory{}, batchArtifactInventoryUnknown("primary artifact yielded %d operation(s), below the ledger's %d; official reference fallback %q failed: %v", len(primary.Endpoints), candidate.OperationsTotal, candidate.ProviderReferenceURL, fallbackErr)
			}
			primary = mergeBatchArtifactInventories(primary, fallback)
		}
		return requireBatchArtifactInventoryTotal(candidate, primary)
	}
	if candidate.ProviderReferenceURL != "" && candidate.ProviderReferenceURL != candidate.Artifact.URL {
		fallback, fallbackErr := parseBatchManifestReferenceFallback(candidate, source, fetch, budget)
		if fallbackErr != nil {
			return batchArtifactInventory{}, batchArtifactInventoryUnknown("primary artifact parse failed (%v), and official reference fallback %q failed: %v", primaryErr, candidate.ProviderReferenceURL, fallbackErr)
		}
		return requireBatchArtifactInventoryTotal(candidate, fallback)
	}
	return batchArtifactInventory{}, primaryErr
}

func parseBatchManifestReferenceFallback(candidate BatchManifestConnector, source batchArtifactSource, fetch batchArtifactFetchFunc, budget *batchArtifactReferenceBudget) (batchArtifactInventory, error) {
	if !budget.hasDocumentCapacity() {
		return batchArtifactInventory{}, batchArtifactInventoryUnknown("candidate reference traversal exceeds the %d-document limit", maxBatchArtifactReferenceDocuments)
	}
	referenceRaw, err := fetch(candidate.ProviderReferenceURL)
	if err != nil {
		return batchArtifactInventory{}, err
	}
	if err := budget.addFetched(referenceRaw); err != nil {
		if err == errBatchArtifactReferenceDocumentLimit {
			return batchArtifactInventory{}, batchArtifactInventoryUnknown("candidate reference traversal exceeds the %d-document limit", maxBatchArtifactReferenceDocuments)
		}
		return batchArtifactInventory{}, batchArtifactInventoryUnknown("candidate reference traversal exceeds the bounded %d-byte source budget", maxBatchArtifactReferenceBytes)
	}
	referenceSource := source
	referenceSource.URL = candidate.ProviderReferenceURL
	referenceSource.Kind = "official-reference"
	referenceSource.SHA256 = fmt.Sprintf("%x", sha256.Sum256(referenceRaw))
	return parseBatchArtifactByKindWithBudget(referenceRaw, referenceSource, fetch, budget)
}

func requireBatchArtifactInventoryTotal(candidate BatchManifestConnector, inventory batchArtifactInventory) (batchArtifactInventory, error) {
	if len(inventory.Endpoints) < candidate.OperationsTotal {
		return batchArtifactInventory{}, batchArtifactInventoryUnknown("artifact inventory yielded %d operation(s), below the ledger's %d", len(inventory.Endpoints), candidate.OperationsTotal)
	}
	return inventory, nil
}

func parseBatchOpenAPIArtifactSourceWithBudget(raw []byte, source batchArtifactSource, fetch batchArtifactFetchFunc, budget *batchArtifactReferenceBudget) (batchArtifactInventory, error) {
	inventory, err := parseBatchOpenAPIArtifactAtWithBudget(raw, source.URL, fetch, budget)
	if err != nil {
		return batchArtifactInventory{}, err
	}
	for index := range inventory.Sources {
		if inventory.Sources[index].Retrieved == "" {
			inventory.Sources[index].Retrieved = source.Retrieved
		}
		if inventory.Sources[index].Kind == "" || inventory.Sources[index].Kind == "openapi" && source.Kind == "swagger" {
			inventory.Sources[index].Kind = source.Kind
		}
		if inventory.Sources[index].Version == "" {
			inventory.Sources[index].Version = source.Version
		}
	}
	for index := range inventory.Endpoints {
		endpoint := &inventory.Endpoints[index]
		if endpoint.SourceURL == "" {
			endpoint.SourceURL = source.URL
		}
		if endpoint.SourceKind == "" {
			endpoint.SourceKind = source.Kind
		}
		if endpoint.SourceVersion == "" {
			endpoint.SourceVersion = source.Version
		}
		if endpoint.SourceRetrieved == "" {
			endpoint.SourceRetrieved = source.Retrieved
		}
		if endpoint.SourceSHA256 == "" && endpoint.SourceURL == source.URL {
			endpoint.SourceSHA256 = source.SHA256
		}
	}
	return inventory, nil
}

type batchArtifactEndpointAlternative struct {
	SourceURL        string
	SourceKind       string
	SourceVersion    string
	SourceRetrieved  string
	SourceSHA256     string
	SourceCoordinate string
}

func batchArtifactEndpointPrimaryAlternative(endpoint batchArtifactEndpoint) batchArtifactEndpointAlternative {
	return batchArtifactEndpointAlternative{
		SourceURL:        endpoint.SourceURL,
		SourceKind:       endpoint.SourceKind,
		SourceVersion:    endpoint.SourceVersion,
		SourceRetrieved:  endpoint.SourceRetrieved,
		SourceSHA256:     endpoint.SourceSHA256,
		SourceCoordinate: endpoint.SourceCoordinate,
	}
}

func appendBatchArtifactEndpointAlternative(endpoint *batchArtifactEndpoint, alternative batchArtifactEndpointAlternative) {
	if alternative.SourceURL == "" || alternative == batchArtifactEndpointPrimaryAlternative(*endpoint) {
		return
	}
	for _, existing := range endpoint.Alternatives {
		if existing == alternative {
			return
		}
	}
	endpoint.Alternatives = append(endpoint.Alternatives, alternative)
}

func normalizeBatchArtifactEndpointAlternatives(endpoint *batchArtifactEndpoint) {
	alternatives := endpoint.Alternatives
	endpoint.Alternatives = nil
	for _, alternative := range alternatives {
		appendBatchArtifactEndpointAlternative(endpoint, alternative)
	}
}

func parseBatchArtifactByKindWithBudget(raw []byte, source batchArtifactSource, fetch batchArtifactFetchFunc, budget *batchArtifactReferenceBudget) (batchArtifactInventory, error) {
	switch strings.ToLower(strings.TrimSpace(source.Kind)) {
	case "postman":
		return parseBatchPostmanArtifact(raw, source)
	case "openapi", "swagger":
		inventory, openAPIErr := parseBatchOpenAPIArtifactSourceWithBudget(raw, source, fetch, budget)
		if openAPIErr == nil {
			return inventory, nil
		}
		inventory, discoveryErr := parseBatchGoogleDiscoveryArtifact(raw, source)
		if discoveryErr == nil {
			return inventory, nil
		}
		return batchArtifactInventory{}, openAPIErr
	case "openapi_fragments", "html_reference", "official-reference":
		return parseBatchReferenceArtifactWithBudget(raw, source, fetch, budget)
	default:
		return batchArtifactInventory{}, batchArtifactInventoryUnknown("unsupported artifact kind %q", source.Kind)
	}
}

func mergeBatchArtifactInventories(primary, fallback batchArtifactInventory) batchArtifactInventory {
	merged := batchArtifactInventory{
		Endpoints: append([]batchArtifactEndpoint(nil), primary.Endpoints...),
		Sources:   append([]batchArtifactSource(nil), primary.Sources...),
	}
	seenSources := map[string]bool{}
	for _, source := range merged.Sources {
		seenSources[source.URL] = true
	}
	for _, source := range fallback.Sources {
		if source.URL != "" && !seenSources[source.URL] {
			seenSources[source.URL] = true
			merged.Sources = append(merged.Sources, source)
		}
	}
	seenEndpoints := map[string]int{}
	for index, endpoint := range merged.Endpoints {
		normalizeBatchArtifactEndpointAlternatives(&merged.Endpoints[index])
		seenEndpoints[batchArtifactEndpointKey(endpoint.Method, endpoint.Path)] = index
	}
	for _, endpoint := range fallback.Endpoints {
		normalizeBatchArtifactEndpointAlternatives(&endpoint)
		key := batchArtifactEndpointKey(endpoint.Method, endpoint.Path)
		if index, exists := seenEndpoints[key]; exists {
			appendBatchArtifactEndpointAlternative(&merged.Endpoints[index], batchArtifactEndpointPrimaryAlternative(endpoint))
			for _, alternative := range endpoint.Alternatives {
				appendBatchArtifactEndpointAlternative(&merged.Endpoints[index], alternative)
			}
			continue
		}
		seenEndpoints[key] = len(merged.Endpoints)
		merged.Endpoints = append(merged.Endpoints, endpoint)
	}
	sort.Slice(merged.Endpoints, func(i, j int) bool {
		if merged.Endpoints[i].Path != merged.Endpoints[j].Path {
			return merged.Endpoints[i].Path < merged.Endpoints[j].Path
		}
		return batchArtifactMethodRank(merged.Endpoints[i].Method) < batchArtifactMethodRank(merged.Endpoints[j].Method)
	})
	return merged
}

type batchBundleCollisionError struct {
	path string
}

func (err *batchBundleCollisionError) Error() string {
	return fmt.Sprintf("destination bundle %s already exists", err.path)
}

func rejectBatchMaterializeDestination(path string) error {
	_, err := os.Lstat(path)
	if err == nil {
		return &batchBundleCollisionError{path: path}
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return fmt.Errorf("inspect destination bundle: %w", err)
}

func rejectBatchMaterializePathOverlap(source, destination string) error {
	if batchMaterializePathContains(source, destination) || batchMaterializePathContains(destination, source) {
		return errors.New("source and destination bundle paths must not overlap")
	}
	return nil
}

func batchMaterializePathContains(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

type batchMaterializeDestination struct {
	path string
	info os.FileInfo
}

func createBatchMaterializeDestination(source, destination string) (*batchMaterializeDestination, error) {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return nil, fmt.Errorf("create destination root: %w", err)
	}
	if err := os.Mkdir(destination, 0o755); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, &batchBundleCollisionError{path: destination}
		}
		return nil, fmt.Errorf("create destination bundle: %w", err)
	}
	info, err := os.Lstat(destination)
	if err != nil {
		return nil, fmt.Errorf("inspect created destination bundle: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("created destination bundle is not a non-symlink directory")
	}
	created := &batchMaterializeDestination{path: destination, info: info}
	if err := copyBatchMaterializeSource(source, destination); err != nil {
		created.discard()
		return nil, err
	}
	return created, nil
}

func copyBatchMaterializeSource(source, destination string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return fmt.Errorf("inspect source bundle: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("source bundle must be a non-symlink directory")
	}
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("source bundle contains symlink %s", path)
		}
		rel, err := filepath.Rel(source, path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return errors.New("source bundle entry escapes its root")
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(destination, rel)
		targetRel, err := filepath.Rel(destination, target)
		if err != nil || targetRel == ".." || strings.HasPrefix(targetRel, ".."+string(filepath.Separator)) {
			return errors.New("destination bundle entry escapes its root")
		}
		if entry.IsDir() {
			if err := os.Mkdir(target, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
				return err
			}
			return nil
		}
		entryInfo, err := entry.Info()
		if err != nil {
			return err
		}
		if !entryInfo.Mode().IsRegular() {
			return fmt.Errorf("source bundle entry %s is not a regular file", path)
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = input.Close() }()
		output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(output, input); err != nil {
			_ = output.Close()
			return err
		}
		if err := output.Close(); err != nil {
			return err
		}
		return nil
	})
}

func (destination *batchMaterializeDestination) discard() {
	if destination == nil || destination.info == nil {
		return
	}
	info, err := os.Lstat(destination.path)
	if err != nil || !info.IsDir() || !os.SameFile(destination.info, info) {
		return
	}
	_ = os.RemoveAll(destination.path)
}

func batchMaterializeDestinationStage(err error) string {
	var collision *batchBundleCollisionError
	if errors.As(err, &collision) {
		return "bundle_collision"
	}
	return "write"
}

func writeBatchMaterializedFiles(bundleDir string, surface, operations, cli []byte) error {
	for _, file := range []struct {
		name string
		raw  []byte
	}{
		{name: "api_surface.json", raw: surface},
		{name: "operations.json", raw: operations},
		{name: "cli_surface.json", raw: cli},
	} {
		if err := writeBatchFile(filepath.Join(bundleDir, file.name), file.raw); err != nil {
			return fmt.Errorf("%s: %w", file.name, err)
		}
	}
	return nil
}

func readBatchMaterializeArtifact(opts batchMaterializeOptions, candidate BatchManifestConnector) ([]byte, error) {
	if opts.artifactDir != "" {
		return readBatchMaterializeArtifactCache(opts.artifactDir, candidate.Connector)
	}
	return fetchBatchMaterializeArtifact(candidate.Artifact.URL)
}

func readBatchMaterializeArtifactCache(dir, connector string) ([]byte, error) {
	if err := validateBatchConnectorName(connector); err != nil {
		return nil, err
	}
	root, err := batchArtifactCacheRoot(dir)
	if err != nil {
		return nil, err
	}
	var matches []string
	for _, extension := range []string{".json", ".yaml", ".yml", ".txt", ".md", ".html", ".htm"} {
		candidate := filepath.Join(root, connector+extension)
		info, err := os.Lstat(candidate)
		if err == nil && info.Mode().IsRegular() {
			matches = append(matches, candidate)
			continue
		}
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("artifact cache file %s must not be a symlink", candidate)
		}
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	if len(matches) != 1 {
		if len(matches) == 0 {
			return nil, fmt.Errorf("artifact cache has no %s.{json,yaml,yml,txt,md,html,htm}", connector)
		}
		return nil, fmt.Errorf("artifact cache has multiple files for %s", connector)
	}
	return readBoundedArtifactFile(matches[0])
}

func readBoundedArtifactFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	if info.Size() > maxMaterializeArtifactBytes {
		return nil, fmt.Errorf("%s exceeds %d-byte artifact limit", path, maxMaterializeArtifactBytes)
	}
	return os.ReadFile(path)
}

func batchArtifactSourceFetcher(opts batchMaterializeOptions, candidate BatchManifestConnector) batchArtifactFetchFunc {
	return func(rawURL string) ([]byte, error) {
		if opts.artifactDir != "" {
			cachePath, err := batchArtifactReferenceCachePath(opts.artifactDir, candidate.Connector, rawURL)
			if err != nil {
				return nil, err
			}
			if raw, cacheErr := readBoundedArtifactFile(cachePath); cacheErr == nil {
				return raw, nil
			} else if !errors.Is(cacheErr, os.ErrNotExist) {
				return nil, fmt.Errorf("read cached external artifact %q: %w", rawURL, cacheErr)
			}
			if opts.cachedReferencesOnly {
				return nil, fmt.Errorf("cached-references-only source pass has no cached external artifact %q", rawURL)
			}
			raw, err := fetchBatchMaterializeSource(rawURL)
			if err != nil {
				return nil, err
			}
			cachePath, err = batchArtifactReferenceCachePath(opts.artifactDir, candidate.Connector, rawURL)
			if err != nil {
				return nil, err
			}
			if err := writeBatchFile(cachePath, raw); err != nil {
				return nil, fmt.Errorf("cache external artifact %q: %w", rawURL, err)
			}
			return raw, nil
		}
		return fetchBatchMaterializeSource(rawURL)
	}
}

func batchArtifactReferenceCachePath(dir, connector, rawURL string) (string, error) {
	if err := validateBatchConnectorName(connector); err != nil {
		return "", err
	}
	root, err := batchArtifactCacheRoot(dir)
	if err != nil {
		return "", err
	}
	references, err := batchArtifactCacheDirectory(root, connector, "references")
	if err != nil {
		return "", err
	}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(rawURL)))
	return filepath.Join(references, digest+".artifact"), nil
}

func batchArtifactCacheRoot(dir string) (string, error) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(root)
	if err != nil {
		return "", fmt.Errorf("inspect artifact cache root: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("artifact cache root must be a non-symlink directory")
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve artifact cache root: %w", err)
	}
	return filepath.Clean(resolved), nil
}

func batchArtifactCacheDirectory(root string, components ...string) (string, error) {
	directory := root
	for _, component := range components {
		if component == "" || component != filepath.Base(component) {
			return "", errors.New("artifact cache directory component is unsafe")
		}
		next := filepath.Join(directory, component)
		rel, err := filepath.Rel(root, next)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return "", errors.New("artifact cache directory escapes its root")
		}
		info, err := os.Lstat(next)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(next, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
				return "", err
			}
			info, err = os.Lstat(next)
		}
		if err != nil {
			return "", err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("artifact cache component %s must be a non-symlink directory", next)
		}
		directory = next
	}
	return directory, nil
}

func fetchBatchMaterializeArtifact(rawURL string) ([]byte, error) {
	parsed, err := parseBatchArtifactURL(rawURL)
	if err != nil {
		return nil, err
	}
	return fetchBatchMaterializeURL(parsed, true)
}

func fetchBatchMaterializeSource(rawURL string) ([]byte, error) {
	parsed, err := parseBatchReferenceURL(rawURL)
	if err != nil {
		return nil, err
	}
	return fetchBatchMaterializeURL(parsed, true)
}

func fetchBatchMaterializeURL(parsed *url.URL, allowQuery bool) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	lookup := batchArtifactLookupIPAddr(net.DefaultResolver.LookupIPAddr)
	if err := validateBatchArtifactRequestURLWithQuery(ctx, parsed, lookup, allowQuery); err != nil {
		return nil, fmt.Errorf("validate artifact destination: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build artifact request: %w", err)
	}
	response, err := newBatchArtifactHTTPClientWithQuery(lookup, allowQuery).Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch artifact: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if err := validateBatchArtifactResponse(response); err != nil {
		return nil, err
	}
	if response.ContentLength > maxMaterializeArtifactBytes {
		return nil, fmt.Errorf("fetch artifact: content length %d exceeds %d-byte artifact limit", response.ContentLength, maxMaterializeArtifactBytes)
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, maxMaterializeArtifactBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read artifact response: %w", err)
	}
	if len(raw) > maxMaterializeArtifactBytes {
		return nil, fmt.Errorf("fetch artifact: response exceeds %d-byte artifact limit", maxMaterializeArtifactBytes)
	}
	return raw, nil
}

func validateBatchArtifactResponse(response *http.Response) error {
	if response == nil {
		return errors.New("fetch artifact: empty HTTP response")
	}
	if response.StatusCode == http.StatusPartialContent {
		return batchArtifactInventoryUnknown("artifact response is incomplete: received HTTP 206 Partial Content")
	}
	if len(response.Header.Values("Content-Range")) > 0 {
		return batchArtifactInventoryUnknown("artifact response is incomplete: response carries Content-Range")
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch artifact: provider returned HTTP %d", response.StatusCode)
	}
	return nil
}

type batchArtifactLookupIPAddr func(context.Context, string) ([]net.IPAddr, error)

type batchArtifactDialContext func(context.Context, string, string) (net.Conn, error)

var batchArtifactNonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("100:0:0:1::/64"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:2::/48"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("fec0::/10"),
}

func parseBatchArtifactURL(raw string) (*url.URL, error) {
	if raw == "" || raw != strings.TrimSpace(raw) {
		return nil, errors.New("artifact URL must be a non-empty absolute HTTPS URL")
	}
	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() || parsed.Scheme != "https" || parsed.Host == "" {
		return nil, errors.New("artifact URL must be an absolute HTTPS URL")
	}
	if parsed.User != nil {
		return nil, errors.New("artifact URL must not include userinfo")
	}
	if parsed.RawQuery != "" || parsed.ForceQuery {
		query, err := url.ParseQuery(parsed.RawQuery)
		if err != nil {
			return nil, errors.New("artifact URL query must be well-formed")
		}
		for key := range query {
			if !batchArtifactReferenceQueryParameterAllowed(key) {
				return nil, errors.New("artifact URL query must use only non-sensitive parameters")
			}
		}
	}
	if parsed.Fragment != "" || strings.Contains(raw, "#") {
		return nil, errors.New("artifact URL must not include a fragment")
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" || strings.Contains(host, "%") || host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return nil, errors.New("artifact URL must name an ordinary HTTPS host")
	}
	if literal, err := netip.ParseAddr(host); err == nil && !batchArtifactIPIsPublic(literal.Unmap()) {
		return nil, errors.New("artifact URL destination must be public")
	}
	return parsed, nil
}

func parseBatchReferenceURL(raw string) (*url.URL, error) {
	if raw == "" || raw != strings.TrimSpace(raw) {
		return nil, errors.New("official reference URL must be a non-empty absolute HTTPS URL")
	}
	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() || parsed.Scheme != "https" || parsed.Host == "" {
		return nil, errors.New("official reference URL must be an absolute HTTPS URL")
	}
	if parsed.User != nil || parsed.Fragment != "" || strings.Contains(raw, "#") {
		return nil, errors.New("official reference URL must not include userinfo or fragments")
	}
	if err := validateBatchArtifactURLObjectWithQuery(parsed, true); err != nil {
		return nil, err
	}
	return parsed, nil
}

func validateBatchArtifactRequestURLWithQuery(ctx context.Context, parsed *url.URL, lookup batchArtifactLookupIPAddr, allowQuery bool) error {
	if err := validateBatchArtifactURLObjectWithQuery(parsed, allowQuery); err != nil {
		return err
	}
	_, err := batchArtifactPublicAddresses(ctx, parsed.Hostname(), lookup)
	return err
}

func validateBatchArtifactURLObject(parsed *url.URL) error {
	return validateBatchArtifactURLObjectWithQuery(parsed, false)
}

func validateBatchArtifactURLObjectWithQuery(parsed *url.URL, allowQuery bool) error {
	if parsed == nil || !parsed.IsAbs() || parsed.Scheme != "https" || parsed.Host == "" {
		return errors.New("artifact request URL must be an absolute HTTPS URL")
	}
	if parsed.User != nil {
		return errors.New("artifact request URL must not include userinfo")
	}
	if (!allowQuery && (parsed.RawQuery != "" || parsed.ForceQuery)) || parsed.Fragment != "" {
		return errors.New("artifact request URL must not include query or fragment components")
	}
	if allowQuery {
		query, err := url.ParseQuery(parsed.RawQuery)
		if err != nil {
			return errors.New("artifact reference query must be well-formed")
		}
		for key := range query {
			if !batchArtifactReferenceQueryParameterAllowed(key) {
				return errors.New("artifact reference query must use only non-sensitive parameters")
			}
		}
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" || strings.Contains(host, "%") || host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return errors.New("artifact request URL must name an ordinary HTTPS host")
	}
	if literal, err := netip.ParseAddr(host); err == nil && !batchArtifactIPIsPublic(literal.Unmap()) {
		return errors.New("artifact request URL destination must be public")
	}
	return nil
}

func batchArtifactReferenceQueryParameterAllowed(key string) bool {
	switch strings.ToLower(key) {
	case "_v", "api-version", "api_version", "converttoopenapi", "download", "environment", "export", "format", "lang", "locale", "reduce", "resolved", "segregateauth", "slug", "v", "ver", "version", "versiontag", "view":
		return true
	default:
		return false
	}
}

func newBatchArtifactHTTPClient(lookup batchArtifactLookupIPAddr) *http.Client {
	return newBatchArtifactHTTPClientWithQuery(lookup, false)
}

func newBatchArtifactHTTPClientWithQuery(lookup batchArtifactLookupIPAddr, allowQuery bool) *http.Client {
	dialer := &net.Dialer{}
	transport := &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialBatchArtifactAddress(ctx, network, address, lookup, dialer.DialContext)
		},
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
		CheckRedirect: func(request *http.Request, _ []*http.Request) error {
			return validateBatchArtifactRequestURLWithQuery(request.Context(), request.URL, lookup, allowQuery)
		},
	}
}

func dialBatchArtifactAddress(ctx context.Context, network, address string, lookup batchArtifactLookupIPAddr, dial batchArtifactDialContext) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("artifact dial network %q is not TCP", network)
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil || host == "" || port == "" {
		return nil, fmt.Errorf("artifact dial address is invalid")
	}
	addresses, err := batchArtifactPublicAddresses(ctx, host, lookup)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, candidate := range addresses {
		connection, err := dial(ctx, network, net.JoinHostPort(candidate.IP.String(), port))
		if err == nil {
			return connection, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("dial public artifact destination: %w", lastErr)
}

func batchArtifactPublicAddresses(ctx context.Context, host string, lookup batchArtifactLookupIPAddr) ([]net.IPAddr, error) {
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return nil, errors.New("artifact destination is not public")
	}
	if literal, err := netip.ParseAddr(host); err == nil {
		literal = literal.Unmap()
		if !batchArtifactIPIsPublic(literal) {
			return nil, errors.New("artifact destination is not public")
		}
		return []net.IPAddr{{IP: net.IP(literal.AsSlice())}}, nil
	}
	resolved, err := lookup(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve artifact destination: %w", err)
	}
	if len(resolved) == 0 {
		return nil, errors.New("artifact destination has no resolved addresses")
	}
	addresses := make([]net.IPAddr, 0, len(resolved))
	seen := map[netip.Addr]bool{}
	for _, resolvedAddress := range resolved {
		if resolvedAddress.Zone != "" {
			return nil, errors.New("artifact destination is not public")
		}
		address, ok := netip.AddrFromSlice(resolvedAddress.IP)
		if !ok {
			return nil, errors.New("artifact destination has an invalid resolved address")
		}
		address = address.Unmap()
		if !batchArtifactIPIsPublic(address) {
			return nil, errors.New("artifact destination is not public")
		}
		if seen[address] {
			continue
		}
		seen[address] = true
		addresses = append(addresses, net.IPAddr{IP: net.IP(address.AsSlice())})
	}
	if len(addresses) == 0 {
		return nil, errors.New("artifact destination has no resolved addresses")
	}
	return addresses, nil
}

func batchArtifactIPIsPublic(address netip.Addr) bool {
	if !address.IsValid() || !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() || address.IsMulticast() || address.IsUnspecified() {
		return false
	}
	for _, prefix := range batchArtifactNonPublicPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

type batchArtifactEndpoint struct {
	Method           string
	Path             string
	Summary          string
	SourceURL        string
	SourceKind       string
	SourceVersion    string
	SourceRetrieved  string
	SourceSHA256     string
	SourceCoordinate string
	Alternatives     []batchArtifactEndpointAlternative
	Webhook          bool
}

// batchArtifactSource is one provider-owned document used to establish one or
// more normalized operations. References are retained as separate sources so
// each endpoint can cite the exact document that supplied its path item.
type batchArtifactSource struct {
	URL       string
	Kind      string
	Version   string
	Retrieved string
	SHA256    string
}

type batchArtifactInventory struct {
	Endpoints []batchArtifactEndpoint
	Sources   []batchArtifactSource
}

type batchArtifactFetchFunc func(string) ([]byte, error)

type batchArtifactDocument struct {
	Root   *yaml.Node
	Source batchArtifactSource
	Base   *url.URL
}

const (
	maxBatchArtifactReferenceDocuments = 64
	maxBatchArtifactReferenceBytes     = 64 << 20
)

type batchArtifactResolver struct {
	fetch     batchArtifactFetchFunc
	documents map[string]batchArtifactDocument
	sources   []batchArtifactSource
	budget    *batchArtifactReferenceBudget
}

type batchArtifactReferenceBudget struct {
	documents  int
	totalBytes int
}

var (
	errBatchArtifactReferenceDocumentLimit = errors.New("artifact reference document limit exceeded")
	errBatchArtifactReferenceByteLimit     = errors.New("artifact reference byte budget exceeded")
)

func newBatchArtifactReferenceBudget(root []byte) *batchArtifactReferenceBudget {
	return &batchArtifactReferenceBudget{documents: 1, totalBytes: len(root)}
}

func (budget *batchArtifactReferenceBudget) hasDocumentCapacity() bool {
	return budget != nil && budget.documents < maxBatchArtifactReferenceDocuments
}

func (budget *batchArtifactReferenceBudget) addFetched(raw []byte) error {
	if !budget.hasDocumentCapacity() {
		return errBatchArtifactReferenceDocumentLimit
	}
	if len(raw) > maxMaterializeArtifactBytes || budget.totalBytes > maxBatchArtifactReferenceBytes-len(raw) {
		return errBatchArtifactReferenceByteLimit
	}
	budget.documents++
	budget.totalBytes += len(raw)
	return nil
}

func parseBatchOpenAPIArtifact(raw []byte) ([]batchArtifactEndpoint, error) {
	inventory, err := parseBatchOpenAPIArtifactAt(raw, "", nil)
	if err != nil {
		return nil, err
	}
	return inventory.Endpoints, nil
}

// parseBatchOpenAPIArtifactAt is the source-aware parser used by materialize
// and by traversal tests. fetch is called only for validated HTTPS external
// references and is bounded by the caller's artifact fetcher.
func parseBatchOpenAPIArtifactAt(raw []byte, sourceURL string, fetch batchArtifactFetchFunc) (batchArtifactInventory, error) {
	return parseBatchOpenAPIArtifactAtWithBudget(raw, sourceURL, fetch, newBatchArtifactReferenceBudget(raw))
}

func parseBatchOpenAPIArtifactAtWithBudget(raw []byte, sourceURL string, fetch batchArtifactFetchFunc, budget *batchArtifactReferenceBudget) (batchArtifactInventory, error) {
	source := batchArtifactSource{URL: sourceURL, Kind: "openapi", SHA256: fmt.Sprintf("%x", sha256.Sum256(raw))}
	document, err := parseBatchArtifactDocument(raw, source)
	if err != nil {
		return batchArtifactInventory{}, fmt.Errorf("decode OpenAPI or Swagger artifact: %w", err)
	}
	fields, err := batchYAMLFields(document.Root)
	if err != nil {
		return batchArtifactInventory{}, errors.New("artifact root must be a mapping")
	}
	openAPI, err := batchYAMLFieldString(fields, "openapi")
	if err != nil {
		return batchArtifactInventory{}, fmt.Errorf("artifact openapi field: %w", err)
	}
	swagger, err := batchYAMLFieldString(fields, "swagger")
	if err != nil {
		return batchArtifactInventory{}, fmt.Errorf("artifact swagger field: %w", err)
	}
	swaggerBasePath := ""
	openAPIServerBasePath := ""
	switch {
	case openAPI != "" && strings.HasPrefix(openAPI, "3.") && swagger == "":
		source.Kind = "openapi"
		source.Version = openAPI
		openAPIServerBasePath, err = batchOpenAPIServerBasePath(fields)
		if err != nil {
			return batchArtifactInventory{}, err
		}
	case swagger == "2.0" && openAPI == "":
		source.Kind = "swagger"
		source.Version = swagger
		swaggerBasePath, err = batchYAMLFieldString(fields, "basePath")
		if err != nil {
			return batchArtifactInventory{}, fmt.Errorf("artifact basePath field: %w", err)
		}
		if err := validateBatchSwaggerBasePath(swaggerBasePath); err != nil {
			return batchArtifactInventory{}, err
		}
	case openAPI == "" && swagger == "":
		return batchArtifactInventory{}, errors.New("artifact is not an OpenAPI or Swagger document")
	default:
		return batchArtifactInventory{}, fmt.Errorf("artifact must declare OpenAPI 3.x or Swagger 2.0 (openapi=%q swagger=%q)", openAPI, swagger)
	}
	document.Source = source
	resolver := newBatchArtifactResolverWithBudget(document, fetch, budget)
	endpoints := make([]batchArtifactEndpoint, 0)
	seen := map[string]bool{}
	if paths, ok := fields["paths"]; ok {
		pathEndpoints, err := batchArtifactEndpointsFromDocument(resolver, document, paths)
		if err != nil {
			return batchArtifactInventory{}, err
		}
		for index := range pathEndpoints {
			pathEndpoints[index].Path = batchSwaggerBasePathEndpointPath(swaggerBasePath, pathEndpoints[index].Path)
			pathEndpoints[index].Path = batchSwaggerBasePathEndpointPath(openAPIServerBasePath, pathEndpoints[index].Path)
		}
		for _, endpoint := range pathEndpoints {
			key := batchArtifactEndpointKey(endpoint.Method, endpoint.Path)
			if seen[key] {
				return batchArtifactInventory{}, batchArtifactInventoryUnknown("duplicate operation %s %s", endpoint.Method, endpoint.Path)
			}
			seen[key] = true
			endpoints = append(endpoints, endpoint)
		}
	}
	if webhooks, ok := fields["webhooks"]; ok {
		webhookEndpoints, err := batchArtifactWebhookEndpoints(resolver, document, webhooks)
		if err != nil {
			return batchArtifactInventory{}, err
		}
		for _, endpoint := range webhookEndpoints {
			key := batchArtifactEndpointKey(endpoint.Method, endpoint.Path)
			if seen[key] {
				return batchArtifactInventory{}, batchArtifactInventoryUnknown("duplicate operation %s %s", endpoint.Method, endpoint.Path)
			}
			seen[key] = true
			endpoints = append(endpoints, endpoint)
		}
	}
	if len(endpoints) == 0 {
		return batchArtifactInventory{}, errors.New("artifact has no HTTP or webhook operations")
	}
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Path != endpoints[j].Path {
			return endpoints[i].Path < endpoints[j].Path
		}
		return batchArtifactMethodRank(endpoints[i].Method) < batchArtifactMethodRank(endpoints[j].Method)
	})
	return batchArtifactInventory{Endpoints: endpoints, Sources: resolver.sources}, nil
}

func validateBatchSwaggerBasePath(basePath string) error {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" || basePath == "/" {
		return nil
	}
	if !strings.HasPrefix(basePath, "/") || strings.ContainsAny(basePath, "?#\\\x00") {
		return batchArtifactInventoryUnknown("artifact Swagger basePath %q must be an absolute request path", basePath)
	}
	return nil
}

func batchSwaggerBasePathEndpointPath(basePath, endpointPath string) string {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" || basePath == "/" {
		return endpointPath
	}
	return strings.TrimRight(basePath, "/") + endpointPath
}

// batchOpenAPIServerBasePath admits a server URL path only when every declared
// OpenAPI server has the same literal request base path. That keeps a server
// alternative from silently changing the materialized endpoint inventory.
func batchOpenAPIServerBasePath(fields map[string]*yaml.Node) (string, error) {
	servers, exists := fields["servers"]
	if !exists {
		return "", nil
	}
	servers, err := batchYAMLDeref(servers)
	if err != nil {
		return "", err
	}
	if servers == nil || servers.Kind != yaml.SequenceNode {
		return "", batchArtifactInventoryUnknown("artifact OpenAPI servers must be a sequence")
	}
	if len(servers.Content) == 0 {
		return "", nil
	}

	basePath := ""
	for index, server := range servers.Content {
		serverFields, err := batchYAMLFields(server)
		if err != nil {
			return "", batchArtifactInventoryUnknown("artifact OpenAPI server %d is not a mapping", index+1)
		}
		serverURL, err := batchYAMLFieldString(serverFields, "url")
		if err != nil || serverURL == "" {
			return "", batchArtifactInventoryUnknown("artifact OpenAPI server %d has no valid URL", index+1)
		}
		candidate, err := batchOpenAPIServerURLPath(serverURL)
		if err != nil {
			return "", err
		}
		if index == 0 {
			basePath = candidate
			continue
		}
		if basePath != candidate {
			return "", batchArtifactInventoryUnknown("artifact OpenAPI servers have ambiguous request base paths %q and %q", basePath, candidate)
		}
	}
	return basePath, nil
}

func batchOpenAPIServerURLPath(serverURL string) (string, error) {
	serverURL = strings.TrimSpace(serverURL)
	if strings.ContainsAny(serverURL, "\r\n\t\\\x00") {
		return "", batchArtifactInventoryUnknown("artifact OpenAPI server URL %q is not a valid absolute URL", serverURL)
	}
	schemeEnd := strings.Index(serverURL, "://")
	if schemeEnd < 1 {
		return "", batchArtifactInventoryUnknown("artifact OpenAPI server URL %q is not an absolute HTTP URL", serverURL)
	}
	scheme := strings.ToLower(serverURL[:schemeEnd])
	if scheme != "https" && scheme != "http" {
		return "", batchArtifactInventoryUnknown("artifact OpenAPI server URL %q is not an HTTP URL", serverURL)
	}
	authorityAndPath := serverURL[schemeEnd+3:]
	pathOffset := strings.IndexByte(authorityAndPath, '/')
	if pathOffset < 0 {
		if authorityAndPath == "" || strings.ContainsAny(authorityAndPath, "?#") {
			return "", batchArtifactInventoryUnknown("artifact OpenAPI server URL %q is not a valid absolute URL", serverURL)
		}
		return "", nil
	}
	authority := authorityAndPath[:pathOffset]
	basePath := authorityAndPath[pathOffset:]
	if authority == "" || strings.ContainsAny(authority, "?#") {
		return "", batchArtifactInventoryUnknown("artifact OpenAPI server URL %q is not a valid absolute URL", serverURL)
	}
	if strings.ContainsAny(basePath, "?#") || !validBatchOpenAPIServerBasePath(basePath) {
		return "", batchArtifactInventoryUnknown("artifact OpenAPI server URL %q has a non-literal request path", serverURL)
	}
	if err := validateBatchSwaggerBasePath(basePath); err != nil {
		return "", batchArtifactInventoryUnknown("artifact OpenAPI server URL %q has an invalid request path", serverURL)
	}
	return basePath, nil
}

// validBatchOpenAPIServerBasePath allows only ordinary OpenAPI server-variable
// names within an otherwise literal request path. A provider may put a required
// account or space segment in servers[].url while retaining root-relative paths;
// that segment is part of the documented request path, not an unsafe URL escape.
func validBatchOpenAPIServerBasePath(path string) bool {
	for index := 0; index < len(path); index++ {
		switch path[index] {
		case '{':
			end := strings.IndexByte(path[index+1:], '}')
			if end < 0 {
				return false
			}
			end += index + 1
			if end == index+1 {
				return false
			}
			for _, character := range path[index+1 : end] {
				if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') && (character < '0' || character > '9') && character != '_' && character != '-' && character != '.' {
					return false
				}
			}
			index = end
		case '}':
			return false
		}
	}
	return true
}

func parseBatchArtifactDocument(raw []byte, source batchArtifactSource) (batchArtifactDocument, error) {
	document, err := decodeBatchArtifactYAML(raw)
	if err != nil {
		if normalized, ok := normalizeBatchArtifactJSON(raw); ok {
			document, err = decodeBatchArtifactYAML(normalized)
		}
		if err != nil {
			return batchArtifactDocument{}, err
		}
	}
	root, err := batchYAMLDeref(yamlDocumentRoot(&document))
	if err != nil {
		return batchArtifactDocument{}, err
	}
	if _, err := batchYAMLFields(root); err != nil {
		return batchArtifactDocument{}, errors.New("artifact root must be a mapping")
	}
	if source.SHA256 == "" {
		source.SHA256 = fmt.Sprintf("%x", sha256.Sum256(raw))
	}
	var base *url.URL
	if source.URL != "" {
		base, _ = url.Parse(source.URL)
	}
	return batchArtifactDocument{Root: root, Source: source, Base: base}, nil
}

func decodeBatchArtifactYAML(raw []byte) (yaml.Node, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return yaml.Node{}, err
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return yaml.Node{}, err
		}
		return yaml.Node{}, batchArtifactInventoryUnknown("artifact contains multiple YAML documents")
	}
	return document, nil
}

// normalizeBatchArtifactJSON lets valid provider JSON use its own Unicode
// escape syntax when yaml.v3 rejects a valid surrogate pair. It is only a
// fallback after YAML decoding fails, and UseNumber prevents number coercion.
func normalizeBatchArtifactJSON(raw []byte) ([]byte, bool) {
	if !json.Valid(raw) {
		return nil, false
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, false
	}
	normalized, err := json.Marshal(value)
	if err != nil {
		return nil, false
	}
	return normalized, true
}

func newBatchArtifactResolverWithBudget(root batchArtifactDocument, fetch batchArtifactFetchFunc, budget *batchArtifactReferenceBudget) *batchArtifactResolver {
	if budget == nil {
		budget = newBatchArtifactReferenceBudget(nil)
	}
	resolver := &batchArtifactResolver{
		fetch:     fetch,
		documents: map[string]batchArtifactDocument{},
		sources:   []batchArtifactSource{},
		budget:    budget,
	}
	key := root.Source.URL
	if key == "" {
		key = "<root>"
	}
	resolver.documents[key] = root
	if root.Source.URL != "" {
		resolver.sources = append(resolver.sources, root.Source)
	}
	return resolver
}

func (resolver *batchArtifactResolver) documentKey(document batchArtifactDocument) string {
	if document.Source.URL == "" {
		return "<root>"
	}
	return document.Source.URL
}

func (resolver *batchArtifactResolver) resolvePathItem(document batchArtifactDocument, pathItem *yaml.Node, refs map[string]bool) (*yaml.Node, batchArtifactDocument, string, bool, error) {
	pathItem, err := batchYAMLDeref(pathItem)
	if err != nil {
		return nil, document, "", false, err
	}
	fields, err := batchYAMLFields(pathItem)
	if err != nil {
		return nil, document, "", false, batchArtifactInventoryUnknown("path item is not a mapping")
	}
	refNode, hasRef := fields["$ref"]
	if !hasRef {
		return pathItem, document, "", false, nil
	}
	ref, err := batchYAMLFieldString(fields, "$ref")
	if err != nil || ref == "" {
		return nil, document, "", false, batchArtifactInventoryUnknown("path-item reference must be a non-empty string")
	}
	for _, key := range batchYAMLFieldNames(fields) {
		if key == "$ref" || key == "summary" || key == "description" || strings.HasPrefix(key, "x-") {
			continue
		}
		return nil, document, "", false, batchArtifactInventoryUnknown("path-item reference %q has unsupported sibling %q", ref, key)
	}
	cycleKey := resolver.documentKey(document) + "#" + ref
	if refs[cycleKey] {
		return nil, document, "", false, batchArtifactInventoryUnknown("path-item reference cycle at %q", ref)
	}
	refs[cycleKey] = true
	defer delete(refs, cycleKey)
	target, targetDocument, pointer, err := resolver.resolveReference(document, ref)
	if err != nil {
		return nil, document, "", false, err
	}
	if target == refNode {
		return nil, document, "", false, batchArtifactInventoryUnknown("path-item reference %q resolves to itself", ref)
	}
	resolved, resolvedDocument, resolvedPointer, resolvedByReference, err := resolver.resolvePathItem(targetDocument, target, refs)
	if err != nil {
		return nil, document, "", false, err
	}
	if resolvedByReference {
		return resolved, resolvedDocument, resolvedPointer, true, nil
	}
	return resolved, resolvedDocument, pointer, true, nil
}

func (resolver *batchArtifactResolver) resolveReference(document batchArtifactDocument, reference string) (*yaml.Node, batchArtifactDocument, string, error) {
	parsed, err := url.Parse(reference)
	if err != nil {
		return nil, document, "", batchArtifactInventoryUnknown("path-item reference %q is not a valid URL reference", reference)
	}
	if parsed.Scheme == "" && parsed.Host == "" && parsed.Path == "" {
		node, err := resolveBatchArtifactJSONPointer(document.Root, parsed.Fragment, reference)
		return node, document, parsed.Fragment, err
	}
	if document.Base == nil {
		return nil, document, "", batchArtifactInventoryUnknown("external path-item reference %q has no HTTPS base URL", reference)
	}
	resolvedURL := document.Base.ResolveReference(parsed)
	fragment := resolvedURL.Fragment
	resolvedURL.Fragment = ""
	if err := validateBatchArtifactURLObject(resolvedURL); err != nil {
		return nil, document, "", batchArtifactInventoryUnknown("external path-item reference %q is unsafe: %v", reference, err)
	}
	external, err := resolver.loadExternal(resolvedURL)
	if err != nil {
		return nil, document, "", err
	}
	node, err := resolveBatchArtifactJSONPointer(external.Root, fragment, reference)
	return node, external, fragment, err
}

func batchArtifactReferenceOperationCoordinate(pointer, method string) string {
	return "#" + pointer + "/" + strings.ToLower(method)
}

func (resolver *batchArtifactResolver) loadExternal(resolvedURL *url.URL) (batchArtifactDocument, error) {
	key := resolvedURL.String()
	if document, ok := resolver.documents[key]; ok {
		return document, nil
	}
	if resolver.fetch == nil {
		return batchArtifactDocument{}, batchArtifactInventoryUnknown("external path-item reference %q cannot be fetched during materialization", key)
	}
	if !resolver.budget.hasDocumentCapacity() {
		return batchArtifactDocument{}, batchArtifactInventoryUnknown("external path-item reference limit %d exceeded", maxBatchArtifactReferenceDocuments)
	}
	raw, err := resolver.fetch(key)
	if err != nil {
		return batchArtifactDocument{}, batchArtifactInventoryUnknown("external path-item reference %q fetch failed: %v", key, err)
	}
	if err := resolver.budget.addFetched(raw); err != nil {
		if err == errBatchArtifactReferenceDocumentLimit {
			return batchArtifactDocument{}, batchArtifactInventoryUnknown("external path-item reference limit %d exceeded", maxBatchArtifactReferenceDocuments)
		}
		return batchArtifactDocument{}, batchArtifactInventoryUnknown("external path-item references exceed the bounded %d-byte source budget", maxBatchArtifactReferenceBytes)
	}
	source := batchArtifactSource{URL: key, Kind: "referenced-document", SHA256: fmt.Sprintf("%x", sha256.Sum256(raw))}
	document, err := parseBatchArtifactDocument(raw, source)
	if err != nil {
		return batchArtifactDocument{}, batchArtifactInventoryUnknown("external path-item reference %q is not a single mapping document: %v", key, err)
	}
	if fields, fieldsErr := batchYAMLFields(document.Root); fieldsErr == nil {
		if version, versionErr := batchYAMLFieldString(fields, "openapi"); versionErr == nil && strings.HasPrefix(version, "3.") {
			document.Source.Kind, document.Source.Version = "openapi", version
		} else if version, versionErr := batchYAMLFieldString(fields, "swagger"); versionErr == nil && version == "2.0" {
			document.Source.Kind, document.Source.Version = "swagger", version
		}
	}
	resolver.documents[key] = document
	resolver.sources = append(resolver.sources, document.Source)
	return document, nil
}

func batchArtifactEndpointsFromDocument(resolver *batchArtifactResolver, document batchArtifactDocument, paths *yaml.Node) ([]batchArtifactEndpoint, error) {
	fields, err := batchYAMLFields(paths)
	if err != nil {
		return nil, errors.New("artifact paths must be a mapping")
	}
	endpoints := make([]batchArtifactEndpoint, 0)
	for _, path := range batchYAMLFieldNames(fields) {
		if strings.HasPrefix(path, "x-") {
			continue
		}
		if path == "" || path != strings.TrimSpace(path) || !strings.HasPrefix(path, "/") {
			return nil, fmt.Errorf("artifact path %q must be non-empty and connector-relative", path)
		}
		pathEndpoints, err := batchArtifactPathItemEndpointsWithResolver(resolver, document, path, fields[path])
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, pathEndpoints...)
	}
	return endpoints, nil
}

func batchArtifactWebhookEndpoints(resolver *batchArtifactResolver, document batchArtifactDocument, webhooks *yaml.Node) ([]batchArtifactEndpoint, error) {
	fields, err := batchYAMLFields(webhooks)
	if err != nil {
		return nil, batchArtifactInventoryUnknown("top-level webhooks is not a mapping")
	}
	endpoints := make([]batchArtifactEndpoint, 0)
	for _, name := range batchYAMLFieldNames(fields) {
		if strings.HasPrefix(name, "x-") {
			continue
		}
		if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "\r\n") {
			return nil, batchArtifactInventoryUnknown("top-level webhooks has an invalid name")
		}
		pathEndpoints, err := batchArtifactPathItemEndpointsWithResolver(resolver, document, name, fields[name])
		if err != nil {
			return nil, batchArtifactInventoryUnknown("top-level webhook %q: %v", name, err)
		}
		if len(pathEndpoints) == 0 {
			return nil, batchArtifactInventoryUnknown("top-level webhook %q has no HTTP operations", name)
		}
		for _, endpoint := range pathEndpoints {
			originalMethod := endpoint.Method
			endpoint.Method = "WEBHOOK"
			endpoint.Path = name + "#" + strings.ToUpper(originalMethod)
			endpoint.Webhook = true
			if endpoint.SourceCoordinate == fmt.Sprintf("paths[%q].%s", name, strings.ToLower(originalMethod)) {
				endpoint.SourceCoordinate = fmt.Sprintf("webhooks[%q].%s", name, strings.ToLower(originalMethod))
			}
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints, nil
}

func batchArtifactPathItemEndpointsWithResolver(resolver *batchArtifactResolver, document batchArtifactDocument, path string, pathItem *yaml.Node) ([]batchArtifactEndpoint, error) {
	resolved, resolvedDocument, resolvedPointer, resolvedByReference, err := resolver.resolvePathItem(document, pathItem, map[string]bool{})
	if err != nil {
		return nil, err
	}
	fields, err := batchYAMLFields(resolved)
	if err != nil {
		return nil, batchArtifactInventoryUnknown("path %q has a non-mapping path item", path)
	}
	endpoints := make([]batchArtifactEndpoint, 0)
	for _, key := range batchYAMLFieldNames(fields) {
		value := fields[key]
		method, isHTTPMethod := batchArtifactHTTPMethodForPathItemKey(key)
		if isHTTPMethod {
			endpoint, err := batchArtifactEndpointFromOperation(path, method, value)
			if err != nil {
				return nil, err
			}
			endpoint.SourceURL = resolvedDocument.Source.URL
			endpoint.SourceKind = resolvedDocument.Source.Kind
			endpoint.SourceVersion = resolvedDocument.Source.Version
			endpoint.SourceRetrieved = resolvedDocument.Source.Retrieved
			endpoint.SourceSHA256 = resolvedDocument.Source.SHA256
			if endpoint.SourceCoordinate == "" {
				if resolvedByReference {
					endpoint.SourceCoordinate = batchArtifactReferenceOperationCoordinate(resolvedPointer, method)
				} else {
					endpoint.SourceCoordinate = fmt.Sprintf("paths[%q].%s", path, strings.ToLower(method))
				}
			}
			endpoints = append(endpoints, endpoint)
			continue
		}
		if strings.HasPrefix(key, "x-") {
			continue
		}
		switch key {
		case "$ref", "summary", "description", "servers", "parameters":
			continue
		default:
			return nil, batchArtifactInventoryUnknown("path %q has unsupported path-item field %q", path, key)
		}
	}
	return endpoints, nil
}

func parseBatchPostmanArtifact(raw []byte, source batchArtifactSource) (batchArtifactInventory, error) {
	var collection struct {
		Info struct {
			Schema  string          `json:"schema"`
			Version json.RawMessage `json:"version"`
		} `json:"info"`
		Item []json.RawMessage `json:"item"`
	}
	if err := json.Unmarshal(raw, &collection); err != nil {
		return batchArtifactInventory{}, fmt.Errorf("decode Postman collection: %w", err)
	}
	if len(collection.Item) == 0 {
		return batchArtifactInventory{}, errors.New("postman collection has no request items")
	}
	source.Kind = "postman"
	if source.Version == "" {
		if version := strings.TrimSpace(string(collection.Info.Version)); version != "" && version != "null" {
			var versionString string
			if json.Unmarshal(collection.Info.Version, &versionString) == nil {
				source.Version = versionString
			}
		}
	}
	if source.SHA256 == "" {
		source.SHA256 = fmt.Sprintf("%x", sha256.Sum256(raw))
	}
	endpoints := make([]batchArtifactEndpoint, 0)
	seen := map[string]bool{}
	var walk func([]json.RawMessage, string) error
	walk = func(items []json.RawMessage, coordinate string) error {
		for index, rawItem := range items {
			var item struct {
				Name    string            `json:"name"`
				Item    []json.RawMessage `json:"item"`
				Request json.RawMessage   `json:"request"`
			}
			if err := json.Unmarshal(rawItem, &item); err != nil {
				return batchArtifactInventoryUnknown("Postman item %s is not an object: %v", coordinate, err)
			}
			itemCoordinate := fmt.Sprintf("%s.item[%d]", coordinate, index)
			if len(item.Item) > 0 {
				if err := walk(item.Item, itemCoordinate); err != nil {
					return err
				}
			}
			if len(item.Request) == 0 || string(item.Request) == "null" {
				continue
			}
			method, path, err := parseBatchPostmanRequest(item.Request)
			if err != nil {
				return batchArtifactInventoryUnknown("Postman request %s: %v", itemCoordinate, err)
			}
			key := batchArtifactEndpointKey(method, path)
			if seen[key] {
				continue
			}
			seen[key] = true
			endpoints = append(endpoints, batchArtifactEndpoint{
				Method:           method,
				Path:             path,
				Summary:          strings.TrimSpace(item.Name),
				SourceURL:        source.URL,
				SourceKind:       source.Kind,
				SourceVersion:    source.Version,
				SourceRetrieved:  source.Retrieved,
				SourceSHA256:     source.SHA256,
				SourceCoordinate: itemCoordinate + ".request.url",
			})
		}
		return nil
	}
	if err := walk(collection.Item, "collection"); err != nil {
		return batchArtifactInventory{}, err
	}
	if len(endpoints) == 0 {
		return batchArtifactInventory{}, errors.New("postman collection has no callable HTTP requests")
	}
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Path != endpoints[j].Path {
			return endpoints[i].Path < endpoints[j].Path
		}
		return batchArtifactMethodRank(endpoints[i].Method) < batchArtifactMethodRank(endpoints[j].Method)
	})
	sources := []batchArtifactSource{}
	if source.URL != "" {
		sources = append(sources, source)
	}
	return batchArtifactInventory{Endpoints: endpoints, Sources: sources}, nil
}

func parseBatchPostmanRequest(raw json.RawMessage) (string, string, error) {
	var object struct {
		Method string          `json:"method"`
		URL    json.RawMessage `json:"url"`
	}
	if err := json.Unmarshal(raw, &object); err != nil {
		var rawURL string
		if string(raw) == "" || json.Unmarshal(raw, &rawURL) != nil {
			return "", "", errors.New("request must be an object or URL string")
		}
		return http.MethodGet, normalizeBatchPostmanPath(rawURL), nil
	}
	method := strings.ToUpper(strings.TrimSpace(object.Method))
	if method == "" {
		method = http.MethodGet
	}
	if len(object.URL) == 0 || string(object.URL) == "null" {
		return "", "", errors.New("request URL is required")
	}
	var rawURL string
	if json.Unmarshal(object.URL, &rawURL) == nil {
		return method, normalizeBatchPostmanPath(rawURL), nil
	}
	var structured struct {
		Raw  string   `json:"raw"`
		Path []string `json:"path"`
	}
	if err := json.Unmarshal(object.URL, &structured); err != nil {
		return "", "", errors.New("request URL must be a string or object")
	}
	if len(structured.Path) > 0 {
		return method, normalizeBatchPostmanPath(strings.Join(structured.Path, "/")), nil
	}
	if strings.TrimSpace(structured.Raw) == "" {
		return "", "", errors.New("request URL has neither raw nor path")
	}
	return method, normalizeBatchPostmanPath(structured.Raw), nil
}

func normalizeBatchPostmanPath(raw string) string {
	path := strings.TrimSpace(raw)
	if path == "" {
		return "/"
	}
	if strings.HasPrefix(path, "{{") {
		if end := strings.Index(path, "}}"); end >= 0 {
			path = path[end+2:]
		}
	}
	if parsed, err := url.Parse(path); err == nil && parsed.IsAbs() {
		path = parsed.EscapedPath()
	} else {
		if cut := strings.IndexAny(path, "?#"); cut >= 0 {
			path = path[:cut]
		}
	}
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	var builder strings.Builder
	for len(path) > 0 {
		open := strings.Index(path, "{{")
		if open < 0 {
			builder.WriteString(path)
			break
		}
		builder.WriteString(path[:open])
		close := strings.Index(path[open+2:], "}}")
		if close < 0 {
			builder.WriteString(path[open:])
			break
		}
		close += open + 2
		name := strings.TrimSpace(path[open+2 : close])
		if name == "" {
			builder.WriteString("{}")
		} else {
			builder.WriteByte('{')
			builder.WriteString(name)
			builder.WriteByte('}')
		}
		path = path[close+2:]
	}
	path = builder.String()
	segments := strings.Split(path, "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, ":") && len(segment) > 1 {
			segments[index] = "{" + segment[1:] + "}"
		}
	}
	return strings.Join(segments, "/")
}

func yamlDocumentRoot(document *yaml.Node) *yaml.Node {
	if document == nil {
		return nil
	}
	if document.Kind == yaml.DocumentNode && len(document.Content) == 1 {
		return document.Content[0]
	}
	return document
}

type batchArtifactInventoryUnknownError struct {
	reason string
}

func (err *batchArtifactInventoryUnknownError) Error() string {
	return "artifact operation inventory is unknown: " + err.reason
}

func batchArtifactInventoryUnknown(format string, args ...any) error {
	return &batchArtifactInventoryUnknownError{reason: fmt.Sprintf(format, args...)}
}

func batchArtifactDropStage(err error, fallback string) string {
	var unknown *batchArtifactInventoryUnknownError
	if errors.As(err, &unknown) {
		return "artifact_inventory_unknown"
	}
	return fallback
}

func batchYAMLDeref(node *yaml.Node) (*yaml.Node, error) {
	seen := map[*yaml.Node]bool{}
	for node != nil && node.Kind == yaml.AliasNode {
		if seen[node] || node.Alias == nil {
			return nil, errors.New("YAML alias cannot be resolved")
		}
		seen[node] = true
		node = node.Alias
	}
	return node, nil
}

func batchYAMLFields(node *yaml.Node) (map[string]*yaml.Node, error) {
	node, err := batchYAMLDeref(node)
	if err != nil {
		return nil, err
	}
	if node == nil || node.Kind != yaml.MappingNode || len(node.Content)%2 != 0 {
		return nil, errors.New("must be a YAML mapping")
	}
	fields := make(map[string]*yaml.Node, len(node.Content)/2)
	for i := 0; i < len(node.Content); i += 2 {
		key, err := batchYAMLDeref(node.Content[i])
		if err != nil {
			return nil, err
		}
		if key == nil || key.Kind != yaml.ScalarNode {
			return nil, errors.New("mapping key must be a string")
		}
		if _, exists := fields[key.Value]; exists {
			return nil, fmt.Errorf("duplicate mapping key %q", key.Value)
		}
		fields[key.Value] = node.Content[i+1]
	}
	return fields, nil
}

func batchYAMLFieldNames(fields map[string]*yaml.Node) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func batchYAMLFieldString(fields map[string]*yaml.Node, key string) (string, error) {
	node, exists := fields[key]
	if !exists {
		return "", nil
	}
	node, err := batchYAMLDeref(node)
	if err != nil {
		return "", err
	}
	if node == nil || node.Kind != yaml.ScalarNode || node.Tag == "!!null" {
		return "", errors.New("must be a string")
	}
	return strings.TrimSpace(node.Value), nil
}

func resolveBatchArtifactJSONPointer(root *yaml.Node, pointer, reference string) (*yaml.Node, error) {
	pointer, err := url.PathUnescape(pointer)
	if err != nil || (pointer != "" && !strings.HasPrefix(pointer, "/")) {
		return nil, batchArtifactInventoryUnknown("path-item reference %q is not a JSON pointer", reference)
	}
	if pointer == "" {
		return batchYAMLDeref(root)
	}
	node := root
	for _, rawToken := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		token, err := batchJSONPointerToken(rawToken)
		if err != nil {
			return nil, batchArtifactInventoryUnknown("local path-item reference %q has an invalid token", reference)
		}
		node, err = batchYAMLDeref(node)
		if err != nil {
			return nil, err
		}
		switch node.Kind {
		case yaml.MappingNode:
			fields, err := batchYAMLFields(node)
			if err != nil {
				return nil, err
			}
			var found bool
			node, found = fields[token]
			if !found {
				return nil, batchArtifactInventoryUnknown("local path-item reference %q cannot be resolved", reference)
			}
		case yaml.SequenceNode:
			if token == "" || (len(token) > 1 && token[0] == '0') {
				return nil, batchArtifactInventoryUnknown("local path-item reference %q has an invalid array index", reference)
			}
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(node.Content) {
				return nil, batchArtifactInventoryUnknown("local path-item reference %q cannot be resolved", reference)
			}
			node = node.Content[index]
		default:
			return nil, batchArtifactInventoryUnknown("local path-item reference %q cannot be resolved", reference)
		}
	}
	return batchYAMLDeref(node)
}

func batchJSONPointerToken(raw string) (string, error) {
	var result strings.Builder
	for index := 0; index < len(raw); index++ {
		if raw[index] != '~' {
			result.WriteByte(raw[index])
			continue
		}
		if index+1 >= len(raw) {
			return "", errors.New("unterminated escape")
		}
		index++
		switch raw[index] {
		case '0':
			result.WriteByte('~')
		case '1':
			result.WriteByte('/')
		default:
			return "", errors.New("invalid escape")
		}
	}
	return result.String(), nil
}

func batchArtifactEndpointFromOperation(path, method string, operation *yaml.Node) (batchArtifactEndpoint, error) {
	operation, err := batchYAMLDeref(operation)
	if err != nil {
		return batchArtifactEndpoint{}, err
	}
	fields, err := batchYAMLFields(operation)
	if err != nil {
		return batchArtifactEndpoint{}, batchArtifactInventoryUnknown("operation %s %s is not a mapping", method, path)
	}
	for _, key := range batchYAMLFieldNames(fields) {
		if strings.HasPrefix(key, "x-") {
			continue
		}
		switch key {
		case "$ref":
			ref, err := batchYAMLFieldString(fields, "$ref")
			if err != nil || ref == "" {
				return batchArtifactEndpoint{}, batchArtifactInventoryUnknown("operation %s %s has an invalid reference", method, path)
			}
		case "tags", "summary", "description", "externalDocs", "operationId", "parameters", "requestBody", "responses", "deprecated", "security", "servers", "consumes", "produces", "schemes", "callbacks", "examples", "freeTier":
			// These fields describe the request or provider-initiated callback
			// behavior. The materialized surface inventories only the request
			// method/path, so callback deliveries never become invented endpoints.
			continue
		default:
			return batchArtifactEndpoint{}, batchArtifactInventoryUnknown("operation %s %s has unsupported field %q", method, path, key)
		}
	}
	summary, err := batchYAMLFieldString(fields, "summary")
	if err != nil {
		return batchArtifactEndpoint{}, batchArtifactInventoryUnknown("operation %s %s has a non-string summary", method, path)
	}
	if summary == "" {
		summary, err = batchYAMLFieldString(fields, "operationId")
		if err != nil {
			return batchArtifactEndpoint{}, batchArtifactInventoryUnknown("operation %s %s has a non-string operationId", method, path)
		}
	}
	if summary == "" {
		summary = fmt.Sprintf("%s %s", method, path)
	}
	return batchArtifactEndpoint{Method: method, Path: path, Summary: summary}, nil
}

func batchArtifactHTTPMethodForPathItemKey(key string) (string, bool) {
	switch key {
	case "get":
		return http.MethodGet, true
	case "head":
		return http.MethodHead, true
	case "post":
		return http.MethodPost, true
	case "put":
		return http.MethodPut, true
	case "patch":
		return http.MethodPatch, true
	case "delete":
		return http.MethodDelete, true
	case "options":
		return http.MethodOptions, true
	case "trace":
		return http.MethodTrace, true
	default:
		return "", false
	}
}

func batchArtifactMethodRank(method string) int {
	switch method {
	case http.MethodGet:
		return 0
	case http.MethodHead:
		return 1
	case http.MethodPost:
		return 2
	case http.MethodPut:
		return 3
	case http.MethodPatch:
		return 4
	case http.MethodDelete:
		return 5
	case http.MethodOptions:
		return 6
	case http.MethodTrace:
		return 7
	default:
		return 8
	}
}

func materializeAPISurface(bundle engine.Bundle, candidate BatchManifestConnector, retrievedAt, sha string, artifactEndpoints []batchArtifactEndpoint, sourceLists ...[]batchArtifactSource) (engine.APISurface, error) {
	if bundle.Surface == nil {
		return engine.APISurface{}, errors.New("api_surface.json is required")
	}
	sources := []batchArtifactSource{{
		URL:       candidate.Artifact.URL,
		Kind:      candidate.Artifact.Kind,
		Version:   candidate.Artifact.Version,
		Retrieved: retrievedAt,
		SHA256:    sha,
	}}
	if len(sourceLists) > 0 && len(sourceLists[0]) > 0 {
		provided := append([]batchArtifactSource(nil), sourceLists[0]...)
		sources = []batchArtifactSource{}
		rootFound := false
		for _, source := range provided {
			if source.URL == candidate.Artifact.URL {
				rootFound = true
				if source.SHA256 == "" {
					source.SHA256 = sha
				}
				if source.Retrieved == "" {
					source.Retrieved = retrievedAt
				}
			}
			sources = append(sources, source)
		}
		if !rootFound {
			root := batchArtifactSource{URL: candidate.Artifact.URL, Kind: candidate.Artifact.Kind, Version: candidate.Artifact.Version, Retrieved: retrievedAt, SHA256: sha}
			sources = append([]batchArtifactSource{root}, sources...)
		}
	}
	existingExact := make(map[string][]engine.SurfaceEndpoint, len(bundle.Surface.Endpoints))
	for _, endpoint := range bundle.Surface.Endpoints {
		key := batchArtifactEndpointKey(endpoint.Method, endpoint.Path)
		existingExact[key] = append(existingExact[key], endpoint)
	}
	normalizedArtifactEndpoints := make([]batchArtifactEndpoint, len(artifactEndpoints))
	for index, endpoint := range artifactEndpoints {
		endpoint.Path = materializedArtifactPathWithExistingBinding(endpoint.Method, endpoint.Path, existingExact)
		normalizedArtifactEndpoints[index] = endpoint
	}
	artifactEndpoints = normalizedArtifactEndpoints
	artifactKeys := make(map[string]bool, len(artifactEndpoints))
	for _, endpoint := range artifactEndpoints {
		artifactKeys[batchArtifactEndpointKey(endpoint.Method, endpoint.Path)] = true
	}
	artifactIDs := make(map[string]string, len(sources))
	usedArtifactIDs := make(map[string]bool, len(sources))
	artifacts := make([]engine.SurfaceArtifact, 0, len(sources))
	for _, source := range sources {
		if strings.TrimSpace(source.URL) == "" {
			continue
		}
		if _, exists := artifactIDs[source.URL]; exists {
			continue
		}
		id := fmt.Sprintf("%s-artifact-%s", candidate.Connector, retrievedAt)
		if source.URL != candidate.Artifact.URL {
			id = materializedReferenceArtifactID(candidate.Connector, source)
		}
		id = uniqueMaterializedArtifactID(id, usedArtifactIDs)
		artifactIDs[source.URL] = id
		usedArtifactIDs[id] = true
		artifacts = append(artifacts, engine.SurfaceArtifact{
			ID:          id,
			URL:         source.URL,
			Kind:        source.Kind,
			Version:     source.Version,
			RetrievedAt: firstNonEmpty(source.Retrieved, retrievedAt),
			SHA256:      source.SHA256,
		})
	}
	artifactID := artifactIDs[candidate.Artifact.URL]
	if artifactID == "" {
		return engine.APISurface{}, errors.New("root artifact source is missing from normalized source inventory")
	}
	surface := engine.APISurface{
		API:                    materializedAPIName(bundle.Surface.API, candidate.Artifact.Version),
		Docs:                   bundle.Surface.Docs,
		ReviewedAt:             retrievedAt,
		OperationLedgerVersion: 2,
		Scope:                  fmt.Sprintf("Provider-artifact inventory generated from the cited %s artifact (%d documented operations across %d provider-owned source document(s)). Every documented operation is represented; existing source-surface bindings absent from the artifact are retained and flagged with %s.", candidate.Artifact.Kind, len(artifactEndpoints), len(artifacts), materializeSurfaceDiscrepancy),
		Artifacts:              artifacts,
		Endpoints:              make([]engine.SurfaceEndpoint, 0, len(artifactEndpoints)),
	}
	artifactOccurrences := make(map[string]int, len(artifactEndpoints))
	for _, artifactEndpoint := range artifactEndpoints {
		artifactOccurrences[batchArtifactEndpointKey(artifactEndpoint.Method, artifactEndpoint.Path)]++
	}
	artifactIndexes := make(map[string]int, len(artifactOccurrences))
	for _, artifactEndpoint := range artifactEndpoints {
		key := batchArtifactEndpointKey(artifactEndpoint.Method, artifactEndpoint.Path)
		existingBindings := materializedExistingBindings(existingExact[key], artifactOccurrences[key], artifactIndexes[key])
		artifactIndexes[key]++
		if len(existingBindings) == 0 {
			existingBindings = []engine.SurfaceEndpoint{{}}
		}
		endpointSourceURL := artifactEndpoint.SourceURL
		if endpointSourceURL == "" {
			endpointSourceURL = candidate.Artifact.URL
		}
		endpointArtifactID := artifactIDs[endpointSourceURL]
		if endpointArtifactID == "" {
			endpointSourceURL = candidate.Artifact.URL
			endpointArtifactID = artifactID
		}
		for _, existing := range existingBindings {
			endpoint := engine.SurfaceEndpoint{
				Method: artifactEndpoint.Method,
				Path:   artifactEndpoint.Path,
				Provenance: &engine.SurfaceProvenance{
					Artifact:     endpointArtifactID,
					SourceURL:    endpointSourceURL,
					SourceKind:   artifactEndpoint.SourceKind,
					Version:      artifactEndpoint.SourceVersion,
					RetrievedAt:  firstNonEmpty(artifactEndpoint.SourceRetrieved, retrievedAt),
					SHA256:       artifactEndpoint.SourceSHA256,
					Coordinate:   artifactEndpoint.SourceCoordinate,
					Alternatives: materializedEndpointAlternatives(artifactEndpoint.Alternatives),
				},
			}
			if operation := batchProtocolMetadataOperation(artifactEndpoint.Method); operation != nil {
				endpoint.Operation = operation
			} else if existing.CoveredBy != nil || existing.Excluded != nil || existing.Operation != nil {
				copyMaterializedClassifier(&endpoint, existing)
			} else {
				endpoint.Operation = defaultMaterializedOperation(artifactEndpoint)
			}
			if endpoint.CoveredBy == nil && endpoint.Excluded == nil && endpoint.Operation == nil {
				endpoint.Operation = defaultMaterializedOperation(artifactEndpoint)
			}
			surface.Endpoints = append(surface.Endpoints, endpoint)
		}
	}
	for _, existing := range bundle.Surface.Endpoints {
		key := batchArtifactEndpointKey(existing.Method, existing.Path)
		if artifactKeys[key] {
			continue
		}
		endpoint := engine.SurfaceEndpoint{
			Method: existing.Method,
			Path:   existing.Path,
			Provenance: &engine.SurfaceProvenance{
				Artifact:    artifactID,
				SourceURL:   candidate.Artifact.URL,
				SourceKind:  candidate.Artifact.Kind,
				Version:     candidate.Artifact.Version,
				RetrievedAt: retrievedAt,
				SHA256:      sha,
				Coordinate:  fmt.Sprintf("existing-surface[%s %s]", existing.Method, existing.Path),
			},
			Discrepancy: materializeSurfaceDiscrepancy,
		}
		copyMaterializedClassifier(&endpoint, existing)
		if endpoint.CoveredBy == nil && endpoint.Excluded == nil && endpoint.Operation == nil {
			endpoint.Operation = defaultMaterializedOperation(batchArtifactEndpoint{Method: existing.Method, Path: existing.Path, Summary: fmt.Sprintf("%s %s", existing.Method, existing.Path)})
		}
		surface.Endpoints = append(surface.Endpoints, endpoint)
	}
	sort.Slice(surface.Endpoints, func(i, j int) bool {
		if surface.Endpoints[i].Path != surface.Endpoints[j].Path {
			return surface.Endpoints[i].Path < surface.Endpoints[j].Path
		}
		return batchArtifactMethodRank(surface.Endpoints[i].Method) < batchArtifactMethodRank(surface.Endpoints[j].Method)
	})
	if err := ensureMaterializedCoverage(bundle, surface); err != nil {
		return engine.APISurface{}, err
	}
	return surface, nil
}

// materializedExistingBindings preserves every source-surface binding for one
// documented method/path. Older bundles legitimately use duplicate endpoint
// rows when several ETL streams share one provider request. When the provider
// artifact itself documents repeated method/path actions, bindings are paired
// one-to-one instead of multiplying the provider inventory.
func materializedExistingBindings(bindings []engine.SurfaceEndpoint, artifactOccurrences, artifactIndex int) []engine.SurfaceEndpoint {
	if len(bindings) == 0 {
		return nil
	}
	if len(bindings) > artifactOccurrences {
		if artifactIndex == 0 {
			if merged, ok := materializedMergedCoverage(bindings); ok {
				return []engine.SurfaceEndpoint{merged}
			}
			return append([]engine.SurfaceEndpoint(nil), bindings...)
		}
		return nil
	}
	if artifactIndex < len(bindings) {
		return []engine.SurfaceEndpoint{bindings[artifactIndex]}
	}
	return nil
}

// Some Discovery documents include a leading API version even when the
// connector's declared base URL already carries that version. Preserve the
// connector-relative spelling only when the source bundle has an exact
// method/path binding after that one conservative normalization.
func materializedArtifactPathWithExistingBinding(method, path string, existing map[string][]engine.SurfaceEndpoint) string {
	if len(existing[batchArtifactEndpointKey(method, path)]) > 0 {
		return path
	}
	withoutVersion := materializedVersionlessDiscoveryPath(path)
	if withoutVersion != path && len(existing[batchArtifactEndpointKey(method, withoutVersion)]) > 0 {
		return withoutVersion
	}
	if suffix := materializedDynamicServerBaseSuffix(method, path, existing); suffix != "" {
		return suffix
	}
	return path
}

// materializedDynamicServerBaseSuffix preserves a connector-relative binding
// when an OpenAPI server base adds a dynamic path segment before a root-relative
// operation path. It is deliberately narrower than a generic suffix match: the
// retained binding must include an OpenAPI-style variable, and the longest
// matching binding wins only when it is unambiguous.
func materializedDynamicServerBaseSuffix(method, path string, existing map[string][]engine.SurfaceEndpoint) string {
	prefix := strings.ToUpper(strings.TrimSpace(method)) + "\x00"
	match := ""
	ambiguous := false
	for key := range existing {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		candidate := strings.TrimPrefix(key, prefix)
		if candidate == path || !strings.Contains(candidate, "{") || !strings.HasSuffix(path, candidate) {
			continue
		}
		if len(candidate) > len(match) {
			match = candidate
			ambiguous = false
			continue
		}
		if len(candidate) == len(match) && candidate != match {
			ambiguous = true
		}
	}
	if ambiguous {
		return ""
	}
	return match
}

func materializedVersionlessDiscoveryPath(path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	version, remainder, found := strings.Cut(trimmed, "/")
	if !found || len(version) < 2 || version[0] != 'v' {
		return path
	}
	for _, character := range version[1:] {
		if character < '0' || character > '9' {
			return path
		}
	}
	return "/" + remainder
}

func materializedMergedCoverage(bindings []engine.SurfaceEndpoint) (engine.SurfaceEndpoint, bool) {
	coverage := &engine.SurfaceCoverage{}
	for _, binding := range bindings {
		if binding.CoveredBy == nil || binding.Excluded != nil || binding.Operation != nil {
			return engine.SurfaceEndpoint{}, false
		}
		coverage.Streams = append(coverage.Streams, binding.CoveredBy.StreamTargets()...)
		coverage.Writes = append(coverage.Writes, binding.CoveredBy.WriteTargets()...)
		if binding.CoveredBy.DirectRead != "" {
			coverage.DirectReads = append(coverage.DirectReads, binding.CoveredBy.DirectRead)
		}
		coverage.DirectReads = append(coverage.DirectReads, binding.CoveredBy.DirectReads...)
	}
	coverage.Streams = materializedUniqueStrings(coverage.Streams)
	coverage.Writes = materializedUniqueStrings(coverage.Writes)
	coverage.DirectReads = materializedUniqueStrings(coverage.DirectReads)
	return engine.SurfaceEndpoint{CoveredBy: coverage}, true
}

func materializedUniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	sort.Strings(unique)
	return unique
}

func materializedReferenceArtifactID(connector string, source batchArtifactSource) string {
	contentDigest := source.SHA256
	if len(contentDigest) < 16 {
		contentDigest = fmt.Sprintf("%x", sha256.Sum256([]byte(source.URL)))
	}
	urlDigest := sha256.Sum256([]byte(source.URL))
	return fmt.Sprintf("%s-artifact-ref-%s-%x", connector, contentDigest[:16], urlDigest[:8])
}

func uniqueMaterializedArtifactID(base string, used map[string]bool) string {
	id := base
	for suffix := 2; used[id]; suffix++ {
		id = fmt.Sprintf("%s-%d", base, suffix)
	}
	return id
}

func materializedEndpointAlternatives(alternatives []batchArtifactEndpointAlternative) []engine.SurfaceProvenanceAlternative {
	if len(alternatives) == 0 {
		return nil
	}
	out := make([]engine.SurfaceProvenanceAlternative, 0, len(alternatives))
	for _, alternative := range alternatives {
		if strings.TrimSpace(alternative.SourceURL) == "" {
			continue
		}
		out = append(out, engine.SurfaceProvenanceAlternative{
			SourceURL:   alternative.SourceURL,
			SourceKind:  alternative.SourceKind,
			Version:     alternative.SourceVersion,
			RetrievedAt: alternative.SourceRetrieved,
			SHA256:      alternative.SourceSHA256,
			Coordinate:  alternative.SourceCoordinate,
		})
	}
	return out
}

func batchArtifactEndpointKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + "\x00" + path
}

func copyMaterializedClassifier(dst *engine.SurfaceEndpoint, src engine.SurfaceEndpoint) {
	if src.CoveredBy != nil {
		coverage := *src.CoveredBy
		coverage.Streams = append([]string(nil), src.CoveredBy.Streams...)
		coverage.Writes = append([]string(nil), src.CoveredBy.Writes...)
		coverage.DirectReads = append([]string(nil), src.CoveredBy.DirectReads...)
		dst.CoveredBy = &coverage
		return
	}
	if src.Excluded != nil {
		dst.Operation = materializedLegacyExclusion(*src.Excluded, dst.Method)
		return
	}
	if src.Operation != nil {
		operation := *src.Operation
		// v2 provenance is endpoint-local. Keeping the v1 operation field
		// would make the shared validator rightly reject this generated row.
		operation.SourceURL = ""
		dst.Operation = &operation
	}
}

// materializedLegacyExclusion upgrades a v1 classifier into v2's blocked
// operation ledger. V2 reserves excluded for no longer-supported legacy
// syntax; an explicit disallowed/deprecated operation retains the original
// reason and is counted as a justified exclusion by the batch report.
func materializedLegacyExclusion(exclusion engine.SurfaceExclusion, method string) *engine.SurfaceOperation {
	model := "disallowed"
	risk := "medium"
	if exclusion.Category == "deprecated" {
		model = "deprecated"
		risk = "low"
	}
	if exclusion.Category == "destructive_admin" || batchArtifactMutationMethod(method) {
		risk = "high"
	}
	reason := strings.TrimSpace(exclusion.Reason)
	if reason == "" {
		reason = fmt.Sprintf("The provider operation is intentionally excluded from this batch (%s).", exclusion.Category)
	}
	dependency, dependencyReason := materializedNamedDependency(method, "", reason)
	return &engine.SurfaceOperation{
		Model:            model,
		Status:           "blocked",
		Risk:             risk,
		BlockedByDefault: true,
		Reason:           reason,
		Notes:            materializedNamedDependencyNote(dependency, dependencyReason),
	}
}

func batchArtifactMutationMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func defaultMaterializedOperation(endpoint batchArtifactEndpoint) *engine.SurfaceOperation {
	dependency, dependencyReason := materializedNamedDependency(endpoint.Method, endpoint.Path, "")
	if endpoint.Webhook {
		return &engine.SurfaceOperation{
			Model:            "disallowed",
			Status:           "blocked",
			Risk:             "medium",
			BlockedByDefault: true,
			Reason:           "Documented OpenAPI top-level webhook is an inbound provider callback, not an outbound provider request path. It remains visible but is not executable until the runtime has a webhook receiver executor.",
			Notes:            materializedNamedDependencyNote(dependency, dependencyReason),
		}
	}
	switch endpoint.Method {
	case http.MethodGet, http.MethodHead:
		return &engine.SurfaceOperation{
			Model:            "direct_read",
			Status:           "blocked",
			Risk:             "low",
			BlockedByDefault: true,
			Reason:           "Documented provider read has no matching executable stream. It remains blocked because the current direct-read executor has no non-redacting output policy.",
			Notes:            materializedNamedDependencyNote(dependency, dependencyReason),
		}
	case http.MethodDelete:
		return &engine.SurfaceOperation{
			Model:            "destructive_action",
			Status:           "blocked",
			Risk:             "high",
			BlockedByDefault: true,
			Reason:           "Documented provider deletion is not promoted without an operation-specific plan, preview, approval, typed confirmation, and execution contract.",
			Notes:            materializedNamedDependencyNote(dependency, dependencyReason),
		}
	default:
		return &engine.SurfaceOperation{
			Model:            "disallowed",
			Status:           "blocked",
			Risk:             "high",
			BlockedByDefault: true,
			Reason:           "Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract in this batch.",
			Notes:            materializedNamedDependencyNote(dependency, dependencyReason),
		}
	}
}

func materializedNamedDependency(method, path, reason string) (string, string) {
	method = strings.ToUpper(strings.TrimSpace(method))
	lowerReason := strings.ToLower(reason)
	switch {
	case method == "WEBHOOK":
		return "engine.webhook_receiver_executor", "the runtime has no inbound webhook receiver executor for top-level webhook operations"
	case method == http.MethodGet || method == http.MethodHead:
		return "engine.direct_read_executor", "the direct-read executor has no non-redacting output policy for this provider operation"
	case strings.Contains(lowerReason, "wrapper") || strings.Contains(lowerReason, "envelope"):
		return "engine.rest_write_body_envelope", "the REST write executor lacks the provider-specific top-level body envelope required by this operation"
	case strings.Contains(lowerReason, "webhook"):
		return "review.webhook_url_mutation", "the provider webhook-URL mutation has no dedicated security-reviewed execution contract"
	case method == http.MethodOptions || method == http.MethodTrace:
		return "engine.protocol_metadata_executor", "the runtime has no provider command executor for protocol-metadata operations"
	default:
		return "engine.rest_write_operation_contract", "the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract"
	}
}

func materializedEndpointNamedDependency(endpoint engine.SurfaceEndpoint) (string, string) {
	reason := ""
	if endpoint.Operation != nil {
		reason = endpoint.Operation.Reason
		switch endpoint.Operation.Model {
		case "direct_read":
			method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
			if method != http.MethodGet && method != http.MethodHead {
				return "engine.direct_read_operation_contract", "the direct-read executor lacks a reviewed operation-specific request, output-policy, and CLI contract"
			}
		case "binary_read":
			return "engine.binary_download_operation_contract", "the binary-download executor lacks a reviewed operation-specific destination and CLI contract"
		}
	}
	return materializedNamedDependency(endpoint.Method, endpoint.Path, reason)
}

func materializedNamedDependencyNote(dependency, description string) string {
	return materializeNamedDependencyPrefix + dependency + ": " + description
}

func ensureMaterializedCoverage(bundle engine.Bundle, surface engine.APISurface) error {
	coveredStreams := map[string]bool{}
	coveredWrites := map[string]bool{}
	for _, endpoint := range surface.Endpoints {
		if endpoint.CoveredBy == nil {
			continue
		}
		for _, stream := range endpoint.CoveredBy.StreamTargets() {
			coveredStreams[stream] = true
		}
		for _, write := range endpoint.CoveredBy.WriteTargets() {
			coveredWrites[write] = true
		}
	}
	for _, stream := range bundle.Streams {
		if !coveredStreams[stream.Name] {
			return fmt.Errorf("stream %q has no matching endpoint in the cited artifact", stream.Name)
		}
	}
	for _, action := range bundle.Writes {
		if !coveredWrites[action.Name] {
			return fmt.Errorf("write action %q has no matching endpoint in the cited artifact", action.Name)
		}
	}
	return nil
}

func materializedAPIName(existing, artifactVersion string) string {
	const marker = "; provider artifact version: "
	base := strings.TrimSpace(existing)
	if before, _, found := strings.Cut(base, marker); found {
		base = before
	}
	if base == "" {
		base = "Provider API"
	}
	return base + marker + artifactVersion
}

type materializedEndpointOperationPlan struct {
	intent        string
	fallbackKind  string
	existingKinds []string
}

func (plan materializedEndpointOperationPlan) matches(kind string) bool {
	for _, expected := range plan.existingKinds {
		if kind == expected {
			return true
		}
	}
	return false
}

func materializedEndpointOperationPlanFor(endpoint engine.SurfaceEndpoint) materializedEndpointOperationPlan {
	method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
	model := ""
	if endpoint.Operation != nil {
		model = endpoint.Operation.Model
	}

	switch model {
	case "direct_read":
		switch method {
		case http.MethodGet:
			return materializedEndpointOperationPlan{intent: "direct_read", fallbackKind: "rest_read", existingKinds: []string{"rest_read"}}
		case http.MethodPost:
			return materializedEndpointOperationPlan{intent: "direct_read", fallbackKind: "composite", existingKinds: []string{"rest_read", "provider_search", "composite"}}
		default:
			return materializedEndpointOperationPlan{intent: "direct_read", fallbackKind: "composite", existingKinds: []string{"composite"}}
		}
	case "binary_read":
		if method == http.MethodGet {
			return materializedEndpointOperationPlan{intent: "binary_download", fallbackKind: "binary_download", existingKinds: []string{"binary_download"}}
		}
		return materializedEndpointOperationPlan{intent: "binary_download", fallbackKind: "composite", existingKinds: []string{"composite"}}
	}

	switch {
	case method == http.MethodGet:
		return materializedEndpointOperationPlan{intent: "direct_read", fallbackKind: "rest_read", existingKinds: []string{"rest_read"}}
	case batchArtifactMutationMethod(method):
		return materializedEndpointOperationPlan{intent: "direct_write", fallbackKind: "rest_write", existingKinds: []string{"rest_write"}}
	default:
		return materializedEndpointOperationPlan{intent: "docs_only", fallbackKind: "composite", existingKinds: []string{"composite"}}
	}
}

func materializeOperationCatalog(bundle engine.Bundle, surface engine.APISurface, candidate BatchManifestConnector) ([]engine.OperationSpec, error) {
	operations := materializedRetainedOperations(bundle.Operations, bundle.Name, surface)
	usedIDs := make(map[string]bool, len(operations))
	for _, operation := range operations {
		usedIDs[operation.ID] = true
	}
	for _, endpoint := range surface.Endpoints {
		if endpoint.Operation == nil {
			continue
		}
		if materializedOperationForEndpoint(operations, endpoint) != "" {
			continue
		}
		plan := materializedEndpointOperationPlanFor(endpoint)
		_, dependencyReason := materializedEndpointNamedDependency(endpoint)
		id := materializedOperationID(bundle.Name, endpoint.Method, endpoint.Path, usedIDs)
		sourceURL := candidate.Artifact.URL
		if endpoint.Provenance != nil && endpoint.Provenance.SourceURL != "" {
			sourceURL = endpoint.Provenance.SourceURL
		}
		operation := engine.OperationSpec{
			ID:           id,
			Summary:      fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path),
			Description:  endpoint.Operation.Reason,
			SourceURL:    sourceURL,
			Risk:         endpoint.Operation.Risk,
			Approval:     "not implemented: " + dependencyReason,
			OutputPolicy: "json",
		}
		if operation.Risk == "" {
			operation.Risk = "high"
		}
		method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
		switch plan.fallbackKind {
		case "rest_read":
			operation.Kind = "rest_read"
			operation.REST = &engine.RESTOperationSpec{Method: method, Path: endpoint.Path, MaxBytes: maxMaterializeArtifactBytes}
		case "rest_write":
			operation.Kind = "rest_write"
			operation.MutationClass = materializedMutationClass(method)
			operation.REST = &engine.RESTOperationSpec{Method: method, Path: endpoint.Path, MaxBytes: maxMaterializeArtifactBytes}
		case "binary_download":
			operation.Kind = "binary_download"
			operation.OutputPolicy = "binary_file_bounded"
			operation.Binary = &engine.BinaryOperationSpec{Method: method, Path: endpoint.Path, MaxBytes: maxMaterializeArtifactBytes}
		default:
			operation.Kind = "composite"
			operation.Composite = &engine.CompositeOperationSpec{Steps: []string{fmt.Sprintf("%s %s", method, endpoint.Path)}}
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

func materializedRetainedOperations(operations []engine.OperationSpec, connector string, surface engine.APISurface) []engine.OperationSpec {
	generatedEndpoints := make(map[string][]engine.SurfaceEndpoint, len(surface.Endpoints))
	for _, endpoint := range surface.Endpoints {
		if endpoint.Operation == nil {
			continue
		}
		base := materializedOperationIDBase(connector, endpoint.Method, endpoint.Path)
		generatedEndpoints[base] = append(generatedEndpoints[base], endpoint)
	}
	retained := make([]engine.OperationSpec, 0, len(operations))
	for _, operation := range operations {
		endpoints, generated := materializedGeneratedOperationEndpoints(operation.ID, generatedEndpoints)
		if generated && !materializedOperationMatchesAnyEndpoint(operation, endpoints) {
			continue
		}
		retained = append(retained, operation)
	}
	return retained
}

func materializedGeneratedOperationEndpoints(id string, endpointsByBase map[string][]engine.SurfaceEndpoint) ([]engine.SurfaceEndpoint, bool) {
	for base, endpoints := range endpointsByBase {
		if id == base {
			return endpoints, true
		}
		if !strings.HasPrefix(id, base+"-") {
			continue
		}
		suffix := strings.TrimPrefix(id, base+"-")
		if n, err := strconv.Atoi(suffix); err == nil && n > 1 {
			return endpoints, true
		}
	}
	return nil, false
}

func materializedOperationID(connector, method, path string, used map[string]bool) string {
	base := materializedOperationIDBase(connector, method, path)
	if !used[base] {
		used[base] = true
		return base
	}
	for index := 2; ; index++ {
		candidate := fmt.Sprintf("%s-%d", base, index)
		if !used[candidate] {
			used[candidate] = true
			return candidate
		}
	}
}

func materializedOperationIDBase(connector, method, path string) string {
	return strings.ToLower(strings.TrimSpace(connector)) + "." + materializedSlug(method) + "." + materializedSlug(path)
}

func materializedSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	separator := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			if separator && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(r)
			separator = false
			continue
		}
		separator = builder.Len() > 0
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "operation"
	}
	return result
}

func materializedMutationClass(method string) string {
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodDelete:
		return "delete"
	default:
		return "update"
	}
}

func materializeCLISurface(bundle engine.Bundle, surface engine.APISurface, candidate BatchManifestConnector, operations []engine.OperationSpec) (engine.CLISurface, error) {
	existingStreamCommands, existingWriteCommands, err := materializedExistingActionCommands(bundle)
	if err != nil {
		return engine.CLISurface{}, err
	}
	streamRefs := map[string][]engine.CLISurfaceEndpointRef{}
	writeRefs := map[string][]engine.CLISurfaceEndpointRef{}
	endpointByKey := map[string]engine.SurfaceEndpoint{}
	for _, endpoint := range surface.Endpoints {
		endpointByKey[batchArtifactEndpointKey(endpoint.Method, endpoint.Path)] = endpoint
		if endpoint.CoveredBy == nil {
			continue
		}
		ref := engine.CLISurfaceEndpointRef{Method: endpoint.Method, Path: endpoint.Path}
		for _, stream := range endpoint.CoveredBy.StreamTargets() {
			streamRefs[stream] = append(streamRefs[stream], ref)
		}
		for _, write := range endpoint.CoveredBy.WriteTargets() {
			writeRefs[write] = append(writeRefs[write], ref)
		}
	}

	commands := make([]engine.CLICommand, 0, len(bundle.Streams)+len(bundle.Writes)+len(operations))
	readPaths := make([]string, 0, len(bundle.Streams))
	writePaths := make([]string, 0, len(bundle.Writes))
	usedPaths := map[string]bool{}
	coveredDirectReads := map[string]bool{}
	for _, endpoint := range surface.Endpoints {
		for _, path := range coveredDirectReadTargets(endpoint.CoveredBy) {
			coveredDirectReads[path] = true
		}
	}
	if len(coveredDirectReads) > 0 {
		if bundle.CLISurface == nil {
			return engine.CLISurface{}, errors.New("api surface direct-read coverage has no source cli surface")
		}
		for _, command := range bundle.CLISurface.Commands {
			if !coveredDirectReads[command.Path] || command.Availability != "implemented" || (command.Intent != "direct_read" && command.Intent != "binary_download") {
				continue
			}
			if usedPaths[command.Path] {
				return engine.CLISurface{}, fmt.Errorf("implemented direct-read command %q conflicts with another generated command", command.Path)
			}
			commands = append(commands, command)
			readPaths = append(readPaths, command.Path)
			usedPaths[command.Path] = true
			delete(coveredDirectReads, command.Path)
		}
		if len(coveredDirectReads) > 0 {
			missing := make([]string, 0, len(coveredDirectReads))
			for path := range coveredDirectReads {
				missing = append(missing, path)
			}
			sort.Strings(missing)
			return engine.CLISurface{}, fmt.Errorf("api surface direct-read coverage has no implemented source command: %s", strings.Join(missing, ", "))
		}
	}
	for _, stream := range bundle.Streams {
		refs := sortedMaterializedReferences(streamRefs[stream.Name])
		if len(refs) == 0 {
			return engine.CLISurface{}, fmt.Errorf("stream %q has no materialized api_surface reference", stream.Name)
		}
		path := materializedCommandPath(stream.Name, "list")
		command := engine.CLICommand{
			Path:         path,
			Summary:      fmt.Sprintf("Run the %s ETL stream", strings.ReplaceAll(stream.Name, "_", " ")),
			Intent:       "etl",
			Availability: "implemented",
			Stream:       stream.Name,
			SourceURL:    candidate.Artifact.URL,
			APISurface:   refs,
			Examples:     []string{fmt.Sprintf("pm %s %s --json", bundle.Name, path)},
		}
		if note := materializedDiscrepancyNote(refs, endpointByKey); note != "" {
			command.Notes = note
		}
		if existing, ok := existingStreamCommands[stream.Name]; ok {
			command = existing
			command.APISurface = refs
			if command.SourceURL == "" {
				command.SourceURL = candidate.Artifact.URL
			}
		}
		if usedPaths[command.Path] {
			return engine.CLISurface{}, fmt.Errorf("stream command %q conflicts with another retained or generated command", command.Path)
		}
		commands = append(commands, command)
		usedPaths[command.Path] = true
		readPaths = append(readPaths, command.Path)
	}
	for _, action := range bundle.Writes {
		refs := sortedMaterializedReferences(writeRefs[action.Name])
		if len(refs) == 0 {
			return engine.CLISurface{}, fmt.Errorf("write action %q has no materialized api_surface reference", action.Name)
		}
		flags, representable, err := materializedWriteFlags(action)
		if err != nil {
			return engine.CLISurface{}, fmt.Errorf("write action %q flags: %w", action.Name, err)
		}
		path := materializedCommandPath(action.Name, "apply")
		command := engine.CLICommand{
			Path:         path,
			Summary:      fmt.Sprintf("Plan and execute the %s reverse-ETL action", strings.ReplaceAll(action.Name, "_", " ")),
			Intent:       "reverse_etl",
			Availability: "implemented",
			Write:        action.Name,
			SourceURL:    candidate.Artifact.URL,
			APISurface:   refs,
			Flags:        flags,
			Risk:         action.Risk,
			Approval:     "requires plan, preview, approval, and execute",
			Examples:     []string{fmt.Sprintf("pm %s %s --plan <plan-name>", bundle.Name, path)},
		}
		if !representable {
			command.Availability = materializeAvailabilityNotImplemented
			command.Notes = materializedNamedDependencyNote("engine.reverse_etl_scalar_flag_contract", "the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags")
		}
		if existing, ok := existingWriteCommands[action.Name]; ok {
			if existing.Availability == "partial" {
				// A partial command has an intentional connector-owned runtime
				// contract. Keep that full contract while refreshing its cited
				// provider endpoints.
				command = existing
				command.APISurface = refs
				if command.SourceURL == "" {
					command.SourceURL = candidate.Artifact.URL
				}
			} else {
				// Implemented command flags are derived from the current write
				// schema. Preserve its registered path, but do not carry stale
				// requiredness or field mappings across a materialization.
				command.Path = existing.Path
				command.SourceCLIPath = existing.SourceCLIPath
				if existing.Summary != "" {
					command.Summary = existing.Summary
				}
			}
		}
		if usedPaths[command.Path] {
			return engine.CLISurface{}, fmt.Errorf("write command %q conflicts with another retained or generated command", command.Path)
		}
		commands = append(commands, command)
		usedPaths[command.Path] = true
		writePaths = append(writePaths, command.Path)
	}
	for _, endpoint := range surface.Endpoints {
		if endpoint.Operation == nil {
			continue
		}
		operationID := materializedOperationForEndpoint(operations, endpoint)
		if operationID == "" {
			return engine.CLISurface{}, fmt.Errorf("blocked operation %s %s has no generated operation metadata", endpoint.Method, endpoint.Path)
		}
		dependency, dependencyReason := materializedEndpointNamedDependency(endpoint)
		path := materializedOperationCommandPath(endpoint.Method, endpoint.Path)
		if usedPaths[path] {
			path += " operation"
		}
		for usedPaths[path] {
			path += " operation"
		}
		usedPaths[path] = true
		plan := materializedEndpointOperationPlanFor(endpoint)
		intent := plan.intent
		paths := &readPaths
		if intent == "direct_write" {
			paths = &writePaths
		}
		command := engine.CLICommand{
			Path:         path,
			Summary:      fmt.Sprintf("Documented %s %s (not implemented)", endpoint.Method, endpoint.Path),
			Intent:       intent,
			Availability: materializeAvailabilityNotImplemented,
			SourceURL:    candidate.Artifact.URL,
			APISurface:   []engine.CLISurfaceEndpointRef{{Method: endpoint.Method, Path: endpoint.Path}},
			Operation:    operationID,
			Risk:         endpoint.Operation.Risk,
			Approval:     "not implemented: " + dependencyReason,
			Notes:        materializedNamedDependencyNote(dependency, dependencyReason),
			Examples:     []string{fmt.Sprintf("pm %s %s --help", bundle.Name, path)},
		}
		commands = append(commands, command)
		*paths = append(*paths, path)
	}
	sort.Slice(commands, func(i, j int) bool { return commands[i].Path < commands[j].Path })
	sort.Strings(readPaths)
	sort.Strings(writePaths)
	groups := make([]engine.CLICommandGroup, 0, 2)
	if len(readPaths) > 0 {
		groups = append(groups, engine.CLICommandGroup{ID: "read", Title: "Read streams", Commands: readPaths})
	}
	if len(writePaths) > 0 {
		groups = append(groups, engine.CLICommandGroup{ID: "write", Title: "Reverse ETL writes", Commands: writePaths})
	}
	return engine.CLISurface{
		Tagline:  fmt.Sprintf("Run %s's declared streams and reverse-ETL actions.", bundle.Metadata.DisplayName),
		Usage:    fmt.Sprintf("pm %s <command> [flags]", bundle.Name),
		Groups:   groups,
		Commands: commands,
	}, nil
}

// materializedExistingActionCommands keeps registered stream and reverse-ETL
// command paths stable while the materializer refreshes their endpoint
// references from the cited provider artifact. Direct-read commands are
// retained separately because their coverage refers to command paths rather
// than a stream or write target.
func materializedExistingActionCommands(bundle engine.Bundle) (map[string]engine.CLICommand, map[string]engine.CLICommand, error) {
	streams := make(map[string]bool, len(bundle.Streams))
	for _, stream := range bundle.Streams {
		streams[stream.Name] = true
	}
	writes := make(map[string]bool, len(bundle.Writes))
	for _, action := range bundle.Writes {
		writes[action.Name] = true
	}

	streamCommands := map[string]engine.CLICommand{}
	writeCommands := map[string]engine.CLICommand{}
	if bundle.CLISurface == nil {
		return streamCommands, writeCommands, nil
	}
	for _, command := range bundle.CLISurface.Commands {
		switch command.Intent {
		case "etl":
			if command.Stream == "" || !streams[command.Stream] {
				continue
			}
			if _, exists := streamCommands[command.Stream]; exists {
				return nil, nil, fmt.Errorf("source cli surface maps stream %q to more than one command", command.Stream)
			}
			streamCommands[command.Stream] = command
		case "reverse_etl":
			if command.Write == "" || !writes[command.Write] {
				continue
			}
			if _, exists := writeCommands[command.Write]; exists {
				return nil, nil, fmt.Errorf("source cli surface maps write action %q to more than one command", command.Write)
			}
			writeCommands[command.Write] = command
		}
	}
	return streamCommands, writeCommands, nil
}

func materializedOperationForEndpoint(operations []engine.OperationSpec, endpoint engine.SurfaceEndpoint) string {
	for _, operation := range operations {
		if materializedOperationMatchesEndpoint(operation, endpoint) {
			return operation.ID
		}
	}
	return ""
}

func materializedOperationMatchesAnyEndpoint(operation engine.OperationSpec, endpoints []engine.SurfaceEndpoint) bool {
	for _, endpoint := range endpoints {
		if materializedOperationMatchesEndpoint(operation, endpoint) {
			return true
		}
	}
	return false
}

func materializedOperationMatchesEndpoint(operation engine.OperationSpec, endpoint engine.SurfaceEndpoint) bool {
	plan := materializedEndpointOperationPlanFor(endpoint)
	if !plan.matches(operation.Kind) {
		return false
	}
	method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
	if operation.REST != nil && strings.EqualFold(operation.REST.Method, method) && operation.REST.Path == endpoint.Path {
		return true
	}
	if operation.Binary != nil && strings.EqualFold(operation.Binary.Method, method) && operation.Binary.Path == endpoint.Path {
		return true
	}
	return operation.Composite != nil && len(operation.Composite.Steps) == 1 && operation.Composite.Steps[0] == fmt.Sprintf("%s %s", method, endpoint.Path)
}

func materializedOperationCommandPath(method, path string) string {
	parts := []string{"api", strings.ToLower(strings.TrimSpace(method))}
	for _, part := range strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '{' || r == '}'
	}) {
		if token := materializedSlug(part); token != "" {
			parts = append(parts, token)
		}
	}
	return strings.Join(parts, " ")
}

func materializedDiscrepancyNote(refs []engine.CLISurfaceEndpointRef, endpoints map[string]engine.SurfaceEndpoint) string {
	for _, ref := range refs {
		if endpoint, ok := endpoints[batchArtifactEndpointKey(ref.Method, ref.Path)]; ok && endpoint.Discrepancy != "" {
			return "discrepancy=" + endpoint.Discrepancy
		}
	}
	return ""
}

func materializedDiscrepancyCount(surface engine.APISurface) int {
	count := 0
	for _, endpoint := range surface.Endpoints {
		if endpoint.Discrepancy == materializeSurfaceDiscrepancy {
			count++
		}
	}
	return count
}

func materializedCLICommandCounts(surface engine.CLISurface) (implemented, namedDependency int) {
	for _, command := range surface.Commands {
		if command.Availability == "implemented" {
			implemented++
		}
		if command.Availability == materializeAvailabilityNotImplemented && strings.HasPrefix(strings.TrimSpace(command.Notes), materializeNamedDependencyPrefix) {
			namedDependency++
		}
	}
	return implemented, namedDependency
}

// materializedWriteFlags derives the smallest complete flag contract for a
// reverse-ETL command. It intentionally does not flatten object or
// object-array inputs into made-up scalar flags: the caller receives a
// visible not_implemented command with a named dependency until the command
// surface has a faithful structured-input contract.
func materializedWriteFlags(action engine.WriteAction) ([]engine.CLIFlag, bool, error) {
	if len(action.RecordSchema) == 0 {
		return nil, false, errors.New("record_schema is required")
	}
	schema, err := parseCLIRecordSchema(action.RecordSchema)
	if err != nil {
		return nil, false, fmt.Errorf("parse record_schema: %w", err)
	}
	required := map[string]bool{}
	for _, path := range schema.requiredMappingPaths("") {
		required[path] = true
	}
	for _, path := range action.PathFields {
		required[path] = true
	}
	paths := make([]string, 0, len(required))
	for path := range required {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	flags := make([]engine.CLIFlag, 0, len(paths))
	for _, path := range paths {
		flag, ok, err := materializedWriteFlag(schema, path)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return nil, false, nil
		}
		flags = append(flags, flag)
	}
	return flags, true, nil
}

func materializedWriteFlag(schema *cliRecordSchemaNode, path string) (engine.CLIFlag, bool, error) {
	// Nested object/array paths would need structured JSON flags. The command
	// surface has no such escape hatch, so leave that action on its existing
	// reverse-ETL route rather than misrepresenting the input contract.
	if strings.Contains(path, ".") {
		return engine.CLIFlag{}, false, nil
	}
	node, err := schema.recordPath(path)
	if err != nil {
		return engine.CLIFlag{}, false, err
	}
	flagType, ok := materializedCLIFlagType(node)
	if !ok {
		return engine.CLIFlag{}, false, nil
	}
	return engine.CLIFlag{
		Name:     path,
		Type:     flagType,
		Summary:  fmt.Sprintf("Required %s record field.", strings.ReplaceAll(path, "_", " ")),
		MapsTo:   "record." + path,
		Required: true,
	}, true, nil
}

func materializedCLIFlagType(node *cliRecordSchemaNode) (string, bool) {
	if node == nil {
		return "", false
	}
	types := node.effectiveTypes()
	if len(types) != 1 {
		return "", false
	}
	switch types[0] {
	case "string":
		return "string", true
	case "integer":
		return "integer", true
	case "number":
		return "number", true
	case "boolean":
		return "boolean", true
	case "array":
		if node.items == nil || len(node.items.effectiveTypes()) != 1 || node.items.effectiveTypes()[0] != "string" {
			return "", false
		}
		return "string_array", true
	default:
		return "", false
	}
}

func sortedMaterializedReferences(refs []engine.CLISurfaceEndpointRef) []engine.CLISurfaceEndpointRef {
	out := append([]engine.CLISurfaceEndpointRef(nil), refs...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return batchArtifactMethodRank(out[i].Method) < batchArtifactMethodRank(out[j].Method)
	})
	return out
}

func materializedCommandPath(name, suffix string) string {
	return strings.ReplaceAll(name, "_", " ") + " " + suffix
}

func writeBatchMaterializeReport(path string, report BatchMaterializeReport) error {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return writeBatchFile(path, append(raw, '\n'))
}
