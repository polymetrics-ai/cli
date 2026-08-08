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
	manifestPath   string
	defsRoot       string
	sourceDefsRoot string
	artifactDir    string
	retrievedAt    string
	reportPath     string
	connectors     []string
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
	artifactInventory, err := parseBatchManifestArtifact(opts, candidate, rawArtifact)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, batchArtifactDropStage(err, "artifact_parse"), err)
	}
	sha := fmt.Sprintf("%x", sha256.Sum256(rawArtifact))
	surface, err := materializeAPISurface(bundle, candidate, opts.retrievedAt, sha, artifactInventory.Endpoints, artifactInventory.Sources)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "coverage", err)
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

func parseBatchManifestArtifact(opts batchMaterializeOptions, candidate BatchManifestConnector, raw []byte) (batchArtifactInventory, error) {
	source := batchArtifactSource{
		URL:       candidate.Artifact.URL,
		Kind:      candidate.Artifact.Kind,
		Version:   candidate.Artifact.Version,
		Retrieved: opts.retrievedAt,
		SHA256:    fmt.Sprintf("%x", sha256.Sum256(raw)),
	}
	fetch := batchArtifactSourceFetcher(opts, candidate)
	primary, primaryErr := parseBatchArtifactByKind(raw, source, fetch)
	if primaryErr == nil {
		if candidate.ProviderReferenceURL != "" && len(primary.Endpoints) < candidate.OperationsTotal && candidate.ProviderReferenceURL != candidate.Artifact.URL {
			referenceRaw, err := fetch(candidate.ProviderReferenceURL)
			if err != nil {
				return batchArtifactInventory{}, batchArtifactInventoryUnknown("primary artifact yielded %d operation(s), below the ledger's %d; official reference fallback %q could not be fetched: %v", len(primary.Endpoints), candidate.OperationsTotal, candidate.ProviderReferenceURL, err)
			}
			referenceSource := source
			referenceSource.URL = candidate.ProviderReferenceURL
			referenceSource.Kind = "official-reference"
			referenceSource.SHA256 = fmt.Sprintf("%x", sha256.Sum256(referenceRaw))
			fallback, fallbackErr := parseBatchArtifactByKind(referenceRaw, referenceSource, fetch)
			if fallbackErr != nil {
				return batchArtifactInventory{}, batchArtifactInventoryUnknown("primary artifact yielded %d operation(s), below the ledger's %d; official reference fallback %q could not be parsed: %v", len(primary.Endpoints), candidate.OperationsTotal, candidate.ProviderReferenceURL, fallbackErr)
			}
			primary = mergeBatchArtifactInventories(primary, fallback)
		}
		return primary, nil
	}
	if candidate.ProviderReferenceURL != "" && candidate.ProviderReferenceURL != candidate.Artifact.URL {
		referenceRaw, err := fetch(candidate.ProviderReferenceURL)
		if err != nil {
			return batchArtifactInventory{}, batchArtifactInventoryUnknown("primary artifact parse failed (%v), and official reference fallback %q could not be fetched: %v", primaryErr, candidate.ProviderReferenceURL, err)
		}
		referenceSource := source
		referenceSource.URL = candidate.ProviderReferenceURL
		referenceSource.Kind = "official-reference"
		referenceSource.SHA256 = fmt.Sprintf("%x", sha256.Sum256(referenceRaw))
		fallback, fallbackErr := parseBatchArtifactByKind(referenceRaw, referenceSource, fetch)
		if fallbackErr != nil {
			return batchArtifactInventory{}, batchArtifactInventoryUnknown("primary artifact parse failed (%v), and official reference fallback %q could not be parsed: %v", primaryErr, candidate.ProviderReferenceURL, fallbackErr)
		}
		return fallback, nil
	}
	return batchArtifactInventory{}, primaryErr
}

func parseBatchOpenAPIArtifactSource(raw []byte, source batchArtifactSource, fetch batchArtifactFetchFunc) (batchArtifactInventory, error) {
	inventory, err := parseBatchOpenAPIArtifactAt(raw, source.URL, fetch)
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

func parseBatchArtifactByKind(raw []byte, source batchArtifactSource, fetch batchArtifactFetchFunc) (batchArtifactInventory, error) {
	switch strings.ToLower(strings.TrimSpace(source.Kind)) {
	case "postman":
		return parseBatchPostmanArtifact(raw, source)
	case "openapi", "swagger":
		return parseBatchOpenAPIArtifactSource(raw, source, fetch)
	case "openapi_fragments", "html_reference", "official-reference":
		return parseBatchReferenceArtifact(raw, source, fetch)
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
		seenEndpoints[batchArtifactEndpointKey(endpoint.Method, endpoint.Path)] = index
	}
	for _, endpoint := range fallback.Endpoints {
		key := batchArtifactEndpointKey(endpoint.Method, endpoint.Path)
		if index, exists := seenEndpoints[key]; exists {
			if endpoint.SourceURL != "" && (endpoint.SourceURL != merged.Endpoints[index].SourceURL || endpoint.SourceCoordinate != merged.Endpoints[index].SourceCoordinate) {
				merged.Endpoints[index].Alternatives = append(merged.Endpoints[index].Alternatives, batchArtifactEndpointAlternative{
					SourceURL:        endpoint.SourceURL,
					SourceKind:       endpoint.SourceKind,
					SourceVersion:    endpoint.SourceVersion,
					SourceRetrieved:  endpoint.SourceRetrieved,
					SourceSHA256:     endpoint.SourceSHA256,
					SourceCoordinate: endpoint.SourceCoordinate,
				})
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
	root, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	var matches []string
	for _, extension := range []string{".json", ".yaml", ".yml", ".txt", ".md", ".html", ".htm"} {
		candidate := filepath.Join(root, connector+extension)
		info, err := os.Stat(candidate)
		if err == nil && info.Mode().IsRegular() {
			matches = append(matches, candidate)
			continue
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
	info, err := os.Stat(path)
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
			raw, err := fetchBatchMaterializeSource(rawURL)
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
	root, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(rawURL)))
	return filepath.Join(root, connector, "references", digest+".artifact"), nil
}

func fetchBatchMaterializeArtifact(rawURL string) ([]byte, error) {
	parsed, err := parseBatchArtifactURL(rawURL)
	if err != nil {
		return nil, err
	}
	return fetchBatchMaterializeURL(parsed, false)
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
		return nil, errors.New("artifact URL must not include a query")
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

func validateBatchArtifactRequestURL(ctx context.Context, parsed *url.URL, lookup batchArtifactLookupIPAddr) error {
	return validateBatchArtifactRequestURLWithQuery(ctx, parsed, lookup, false)
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
		for key := range parsed.Query() {
			if batchArtifactCredentialQueryParameter(key) {
				return errors.New("artifact reference query must not contain credential-shaped parameters")
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

func batchArtifactCredentialQueryParameter(key string) bool {
	normalized := strings.Map(func(r rune) rune {
		if r == '-' || r == '_' || r == '.' || r == ' ' {
			return -1
		}
		return r
	}, strings.ToLower(key))
	if normalized == "sig" {
		return true
	}
	for _, marker := range []string{"token", "secret", "password", "credential", "authorization", "apikey", "accesskey", "signature"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
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
	fetch      batchArtifactFetchFunc
	documents  map[string]batchArtifactDocument
	sources    []batchArtifactSource
	fetched    int
	totalBytes int
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
	switch {
	case openAPI != "" && strings.HasPrefix(openAPI, "3.") && swagger == "":
		source.Kind = "openapi"
		source.Version = openAPI
	case swagger == "2.0" && openAPI == "":
		source.Kind = "swagger"
		source.Version = swagger
	case openAPI == "" && swagger == "":
		return batchArtifactInventory{}, errors.New("artifact is not an OpenAPI or Swagger document")
	default:
		return batchArtifactInventory{}, fmt.Errorf("artifact must declare OpenAPI 3.x or Swagger 2.0 (openapi=%q swagger=%q)", openAPI, swagger)
	}
	document.Source = source
	resolver := newBatchArtifactResolver(document, fetch)
	endpoints := make([]batchArtifactEndpoint, 0)
	seen := map[string]bool{}
	if paths, ok := fields["paths"]; ok {
		pathEndpoints, err := batchArtifactEndpointsFromDocument(resolver, document, paths)
		if err != nil {
			return batchArtifactInventory{}, err
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

func parseBatchArtifactDocument(raw []byte, source batchArtifactSource) (batchArtifactDocument, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return batchArtifactDocument{}, err
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return batchArtifactDocument{}, err
		}
		return batchArtifactDocument{}, batchArtifactInventoryUnknown("artifact contains multiple YAML documents")
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

func newBatchArtifactResolver(root batchArtifactDocument, fetch batchArtifactFetchFunc) *batchArtifactResolver {
	resolver := &batchArtifactResolver{
		fetch:     fetch,
		documents: map[string]batchArtifactDocument{},
		sources:   []batchArtifactSource{},
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

func (resolver *batchArtifactResolver) resolvePathItem(document batchArtifactDocument, pathItem *yaml.Node, refs map[string]bool) (*yaml.Node, batchArtifactDocument, error) {
	pathItem, err := batchYAMLDeref(pathItem)
	if err != nil {
		return nil, document, err
	}
	fields, err := batchYAMLFields(pathItem)
	if err != nil {
		return nil, document, batchArtifactInventoryUnknown("path item is not a mapping")
	}
	refNode, hasRef := fields["$ref"]
	if !hasRef {
		return pathItem, document, nil
	}
	ref, err := batchYAMLFieldString(fields, "$ref")
	if err != nil || ref == "" {
		return nil, document, batchArtifactInventoryUnknown("path-item reference must be a non-empty string")
	}
	for _, key := range batchYAMLFieldNames(fields) {
		if key == "$ref" || key == "summary" || key == "description" || strings.HasPrefix(key, "x-") {
			continue
		}
		return nil, document, batchArtifactInventoryUnknown("path-item reference %q has unsupported sibling %q", ref, key)
	}
	cycleKey := resolver.documentKey(document) + "#" + ref
	if refs[cycleKey] {
		return nil, document, batchArtifactInventoryUnknown("path-item reference cycle at %q", ref)
	}
	refs[cycleKey] = true
	defer delete(refs, cycleKey)
	target, targetDocument, err := resolver.resolveReference(document, ref)
	if err != nil {
		return nil, document, err
	}
	if target == refNode {
		return nil, document, batchArtifactInventoryUnknown("path-item reference %q resolves to itself", ref)
	}
	return resolver.resolvePathItem(targetDocument, target, refs)
}

func (resolver *batchArtifactResolver) resolveReference(document batchArtifactDocument, reference string) (*yaml.Node, batchArtifactDocument, error) {
	parsed, err := url.Parse(reference)
	if err != nil {
		return nil, document, batchArtifactInventoryUnknown("path-item reference %q is not a valid URL reference", reference)
	}
	if parsed.Scheme == "" && parsed.Host == "" && parsed.Path == "" {
		node, err := resolveBatchArtifactJSONPointer(document.Root, parsed.Fragment, reference)
		return node, document, err
	}
	if document.Base == nil {
		return nil, document, batchArtifactInventoryUnknown("external path-item reference %q has no HTTPS base URL", reference)
	}
	resolvedURL := document.Base.ResolveReference(parsed)
	fragment := resolvedURL.Fragment
	resolvedURL.Fragment = ""
	if err := validateBatchArtifactURLObject(resolvedURL); err != nil {
		return nil, document, batchArtifactInventoryUnknown("external path-item reference %q is unsafe: %v", reference, err)
	}
	external, err := resolver.loadExternal(resolvedURL)
	if err != nil {
		return nil, document, err
	}
	node, err := resolveBatchArtifactJSONPointer(external.Root, fragment, reference)
	return node, external, err
}

func (resolver *batchArtifactResolver) loadExternal(resolvedURL *url.URL) (batchArtifactDocument, error) {
	key := resolvedURL.String()
	if document, ok := resolver.documents[key]; ok {
		return document, nil
	}
	if resolver.fetch == nil {
		return batchArtifactDocument{}, batchArtifactInventoryUnknown("external path-item reference %q cannot be fetched during materialization", key)
	}
	if resolver.fetched >= maxBatchArtifactReferenceDocuments {
		return batchArtifactDocument{}, batchArtifactInventoryUnknown("external path-item reference limit %d exceeded", maxBatchArtifactReferenceDocuments)
	}
	resolver.fetched++
	raw, err := resolver.fetch(key)
	if err != nil {
		return batchArtifactDocument{}, batchArtifactInventoryUnknown("external path-item reference %q fetch failed: %v", key, err)
	}
	if len(raw) > maxMaterializeArtifactBytes || resolver.totalBytes > maxBatchArtifactReferenceBytes-len(raw) {
		return batchArtifactDocument{}, batchArtifactInventoryUnknown("external path-item references exceed the bounded %d-byte source budget", maxBatchArtifactReferenceBytes)
	}
	resolver.totalBytes += len(raw)
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
			endpoint.SourceCoordinate = fmt.Sprintf("webhooks[%q].%s", name, strings.ToLower(originalMethod))
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints, nil
}

func batchArtifactPathItemEndpointsWithResolver(resolver *batchArtifactResolver, document batchArtifactDocument, path string, pathItem *yaml.Node) ([]batchArtifactEndpoint, error) {
	resolved, resolvedDocument, err := resolver.resolvePathItem(document, pathItem, map[string]bool{})
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
				endpoint.SourceCoordinate = fmt.Sprintf("paths[%q].%s", path, strings.ToLower(method))
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

// batchArtifactEndpointsFromPaths remains the local-only helper used by
// narrow parser tests and callers that do not have a source URL. Materialize
// uses the source-aware variant above so external references are traversed.
func batchArtifactEndpointsFromPaths(root, paths *yaml.Node) ([]batchArtifactEndpoint, error) {
	root, err := batchYAMLDeref(root)
	if err != nil {
		return nil, err
	}
	document := batchArtifactDocument{Root: root, Source: batchArtifactSource{Kind: "openapi"}}
	resolver := newBatchArtifactResolver(document, nil)
	return batchArtifactEndpointsFromDocument(resolver, document, paths)
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
		return batchArtifactInventory{}, errors.New("Postman collection has no request items")
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
		return batchArtifactInventory{}, errors.New("Postman collection has no callable HTTP requests")
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

func resolveBatchArtifactLocalReference(root *yaml.Node, reference string) (*yaml.Node, error) {
	if !strings.HasPrefix(reference, "#") {
		return nil, batchArtifactInventoryUnknown("external path-item reference %q cannot be exhaustively resolved", reference)
	}
	return resolveBatchArtifactJSONPointer(root, strings.TrimPrefix(reference, "#"), reference)
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
		value := fields[key]
		if strings.HasPrefix(key, "x-") {
			continue
		}
		switch key {
		case "callbacks":
			callbacks, err := batchYAMLFields(value)
			if err != nil || len(callbacks) > 0 {
				return batchArtifactEndpoint{}, batchArtifactInventoryUnknown("operation %s %s contains callbacks that cannot be represented as provider request paths", method, path)
			}
		case "tags", "summary", "description", "externalDocs", "operationId", "parameters", "requestBody", "responses", "deprecated", "security", "servers", "consumes", "produces", "schemes":
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
	artifactKeys := make(map[string]bool, len(artifactEndpoints))
	for _, endpoint := range artifactEndpoints {
		artifactKeys[batchArtifactEndpointKey(endpoint.Method, endpoint.Path)] = true
	}
	existingExact := make(map[string]engine.SurfaceEndpoint, len(bundle.Surface.Endpoints))
	for _, endpoint := range bundle.Surface.Endpoints {
		key := batchArtifactEndpointKey(endpoint.Method, endpoint.Path)
		if _, duplicate := existingExact[key]; duplicate {
			return engine.APISurface{}, fmt.Errorf("existing api surface duplicates %s %s", endpoint.Method, endpoint.Path)
		}
		existingExact[key] = endpoint
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
	for _, artifactEndpoint := range artifactEndpoints {
		endpointSourceURL := artifactEndpoint.SourceURL
		if endpointSourceURL == "" {
			endpointSourceURL = candidate.Artifact.URL
		}
		endpointArtifactID := artifactIDs[endpointSourceURL]
		if endpointArtifactID == "" {
			endpointSourceURL = candidate.Artifact.URL
			endpointArtifactID = artifactID
		}
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
		} else if existing, ok := existingExact[batchArtifactEndpointKey(artifactEndpoint.Method, artifactEndpoint.Path)]; ok {
			copyMaterializedClassifier(&endpoint, existing)
		} else {
			endpoint.Operation = defaultMaterializedOperation(artifactEndpoint)
		}
		if endpoint.CoveredBy == nil && endpoint.Excluded == nil && endpoint.Operation == nil {
			endpoint.Operation = defaultMaterializedOperation(artifactEndpoint)
		}
		surface.Endpoints = append(surface.Endpoints, endpoint)
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
	return method + "\x00" + path
}

func copyMaterializedClassifier(dst *engine.SurfaceEndpoint, src engine.SurfaceEndpoint) {
	if src.CoveredBy != nil {
		coverage := *src.CoveredBy
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
		if endpoint.CoveredBy.Stream != "" {
			coveredStreams[endpoint.CoveredBy.Stream] = true
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

func materializeOperationCatalog(bundle engine.Bundle, surface engine.APISurface, candidate BatchManifestConnector) ([]engine.OperationSpec, error) {
	operations := append([]engine.OperationSpec(nil), bundle.Operations...)
	usedIDs := make(map[string]bool, len(operations))
	for _, operation := range operations {
		usedIDs[operation.ID] = true
	}
	for _, endpoint := range surface.Endpoints {
		if endpoint.Operation == nil {
			continue
		}
		_, dependencyReason := materializedNamedDependency(endpoint.Method, endpoint.Path, endpoint.Operation.Reason)
		id := materializedOperationID(bundle.Name, endpoint.Method, endpoint.Path, usedIDs)
		operation := engine.OperationSpec{
			ID:           id,
			Summary:      fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path),
			Description:  endpoint.Operation.Reason,
			SourceURL:    endpoint.Provenance.SourceURL,
			Risk:         endpoint.Operation.Risk,
			Approval:     "not implemented: " + dependencyReason,
			OutputPolicy: "json",
		}
		if operation.Risk == "" {
			operation.Risk = "high"
		}
		method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
		switch method {
		case http.MethodGet:
			operation.Kind = "rest_read"
			operation.REST = &engine.RESTOperationSpec{Method: method, Path: endpoint.Path, MaxBytes: maxMaterializeArtifactBytes}
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			operation.Kind = "rest_write"
			operation.MutationClass = materializedMutationClass(method)
			operation.REST = &engine.RESTOperationSpec{Method: method, Path: endpoint.Path, MaxBytes: maxMaterializeArtifactBytes}
		default:
			operation.Kind = "composite"
			operation.Composite = &engine.CompositeOperationSpec{Steps: []string{fmt.Sprintf("%s %s", method, endpoint.Path)}}
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

func materializedOperationID(connector, method, path string, used map[string]bool) string {
	base := strings.ToLower(strings.TrimSpace(connector)) + "." + strings.ToLower(strings.TrimSpace(method)) + "." + materializedSlug(path)
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
	streamRefs := map[string][]engine.CLISurfaceEndpointRef{}
	writeRefs := map[string][]engine.CLISurfaceEndpointRef{}
	endpointByKey := map[string]engine.SurfaceEndpoint{}
	for _, endpoint := range surface.Endpoints {
		endpointByKey[batchArtifactEndpointKey(endpoint.Method, endpoint.Path)] = endpoint
		if endpoint.CoveredBy == nil {
			continue
		}
		ref := engine.CLISurfaceEndpointRef{Method: endpoint.Method, Path: endpoint.Path}
		if endpoint.CoveredBy.Stream != "" {
			streamRefs[endpoint.CoveredBy.Stream] = append(streamRefs[endpoint.CoveredBy.Stream], ref)
		}
		for _, write := range endpoint.CoveredBy.WriteTargets() {
			writeRefs[write] = append(writeRefs[write], ref)
		}
	}

	commands := make([]engine.CLICommand, 0, len(bundle.Streams)+len(bundle.Writes)+len(operations))
	readPaths := make([]string, 0, len(bundle.Streams))
	writePaths := make([]string, 0, len(bundle.Writes))
	usedPaths := map[string]bool{}
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
		commands = append(commands, command)
		usedPaths[path] = true
		readPaths = append(readPaths, path)
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
		commands = append(commands, command)
		usedPaths[path] = true
		writePaths = append(writePaths, path)
	}
	for _, endpoint := range surface.Endpoints {
		if endpoint.Operation == nil {
			continue
		}
		operationID := materializedOperationForEndpoint(operations, endpoint)
		if operationID == "" {
			return engine.CLISurface{}, fmt.Errorf("blocked operation %s %s has no generated operation metadata", endpoint.Method, endpoint.Path)
		}
		dependency, dependencyReason := materializedNamedDependency(endpoint.Method, endpoint.Path, endpoint.Operation.Reason)
		path := materializedOperationCommandPath(endpoint.Method, endpoint.Path)
		if usedPaths[path] {
			path += " operation"
		}
		for usedPaths[path] {
			path += " operation"
		}
		usedPaths[path] = true
		intent := "docs_only"
		paths := &readPaths
		method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
		if method == http.MethodGet {
			intent = "direct_read"
		} else if batchArtifactMutationMethod(method) {
			intent = "direct_write"
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

func materializedOperationForEndpoint(operations []engine.OperationSpec, endpoint engine.SurfaceEndpoint) string {
	method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
	for _, operation := range operations {
		if operation.REST != nil && strings.EqualFold(operation.REST.Method, method) && operation.REST.Path == endpoint.Path {
			return operation.ID
		}
		if operation.Composite != nil && len(operation.Composite.Steps) == 1 && operation.Composite.Steps[0] == fmt.Sprintf("%s %s", method, endpoint.Path) {
			return operation.ID
		}
	}
	return ""
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
