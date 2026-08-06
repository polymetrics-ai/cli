package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"polymetrics.ai/internal/connectors/engine"
)

const (
	batchMaterializeReportSchemaVersion = 1
	maxMaterializeArtifactBytes         = 16 << 20
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
	Connector          string                `json:"connector"`
	Artifact           BatchMaterialArtifact `json:"artifact"`
	ArtifactOperations int                   `json:"artifact_operations"`
	DeclaredOperations int                   `json:"declared_operations"`
	OperationSplit     BatchOperationSplit   `json:"operation_split"`
	CLICommands        int                   `json:"cli_commands"`
	OperationExecutors int                   `json:"operation_executors"`
}

// BatchMaterialArtifact combines the survey's immutable version with the
// newly fetched public artifact evidence. The shared v2 schema intentionally
// owns URL/date/digest only; this report preserves the provider version rather
// than inventing another shared provenance field.
type BatchMaterialArtifact struct {
	URL         string `json:"url"`
	Version     string `json:"version"`
	RetrievedAt string `json:"retrieved_at"`
	SHA256      string `json:"sha256"`
}

type batchMaterializeOptions struct {
	manifestPath string
	defsRoot     string
	artifactDir  string
	retrievedAt  string
	reportPath   string
	connectors   []string
}

// runBatchMaterialize is the one post-#3869 artifact-to-bundle authoring
// path. It reads public OpenAPI/Swagger documents from the manifest, upgrades
// the existing surface inventory to v2 provenance, and derives only
// stream/write command declarations that the real runtime can preflight.
// Direct reads are deliberately not promoted: current runtime policies for
// them are redacting, which this batch lane rejects.
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
	bundleDir, err := batchBundleDirectory(opts.defsRoot, candidate.Connector)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "bundle_path", err)
	}
	if !isBundleDir(bundleDir) {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "bundle_path", errors.New("metadata.json is required for a batch bundle"))
	}
	defsRoot, err := filepath.Abs(opts.defsRoot)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "bundle_path", err)
	}
	bundle, err := engine.Load(os.DirFS(defsRoot), candidate.Connector)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "load", err)
	}
	if bundle.Surface == nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "api_surface", errors.New("api_surface.json is required before materialization"))
	}

	rawArtifact, err := readBatchMaterializeArtifact(opts, candidate)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "artifact_fetch", err)
	}
	artifactEndpoints, err := parseBatchOpenAPIArtifact(rawArtifact)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "artifact_parse", err)
	}
	sha := fmt.Sprintf("%x", sha256.Sum256(rawArtifact))
	surface, err := materializeAPISurface(bundle, candidate, opts.retrievedAt, sha, artifactEndpoints)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "coverage", err)
	}
	cli, err := materializeCLISurface(bundle, surface, candidate)
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
	operationsRaw := []byte("{\n  \"operations\": []\n}\n")
	if err := writeBatchFile(filepath.Join(bundleDir, "api_surface.json"), append(surfaceRaw, '\n')); err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "write", fmt.Errorf("write api_surface.json: %w", err))
	}
	if err := writeBatchFile(filepath.Join(bundleDir, "operations.json"), operationsRaw); err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "write", fmt.Errorf("write operations.json: %w", err))
	}
	if err := writeBatchFile(filepath.Join(bundleDir, "cli_surface.json"), append(cliRaw, '\n')); err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "write", fmt.Errorf("write cli_surface.json: %w", err))
	}

	return BatchMaterializeIncluded{
		Connector: candidate.Connector,
		Artifact: BatchMaterialArtifact{
			URL:         candidate.Artifact.URL,
			Version:     candidate.Artifact.Version,
			RetrievedAt: opts.retrievedAt,
			SHA256:      sha,
		},
		ArtifactOperations: len(artifactEndpoints),
		DeclaredOperations: split.total(),
		OperationSplit:     split,
		CLICommands:        len(cli.Commands),
		OperationExecutors: 0,
	}, nil
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
	for _, extension := range []string{".json", ".yaml", ".yml"} {
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
			return nil, fmt.Errorf("artifact cache has no %s.{json,yaml,yml}", connector)
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

func fetchBatchMaterializeArtifact(rawURL string) ([]byte, error) {
	parsed, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build artifact request: %w", err)
	}
	if parsed.URL.Scheme != "https" || parsed.URL.Host == "" || parsed.URL.User != nil {
		return nil, errors.New("artifact URL must be absolute HTTPS without userinfo")
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(request *http.Request, _ []*http.Request) error {
			if request.URL.Scheme != "https" || request.URL.Host == "" || request.URL.User != nil {
				return errors.New("artifact redirect must remain absolute HTTPS without userinfo")
			}
			return nil
		},
	}
	response, err := client.Do(parsed)
	if err != nil {
		return nil, fmt.Errorf("fetch artifact: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetch artifact: provider returned HTTP %d", response.StatusCode)
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

type batchOpenAPIDocument struct {
	OpenAPI string                                `json:"openapi"`
	Swagger string                                `json:"swagger"`
	Paths   map[string]map[string]json.RawMessage `json:"paths"`
}

type batchArtifactEndpoint struct {
	Method  string
	Path    string
	Summary string
}

func parseBatchOpenAPIArtifact(raw []byte) ([]batchArtifactEndpoint, error) {
	var document batchOpenAPIDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return parseBatchYAMLOpenAPIArtifact(raw, err)
	}
	if strings.TrimSpace(document.OpenAPI) == "" && strings.TrimSpace(document.Swagger) == "" {
		return nil, errors.New("artifact is not an OpenAPI or Swagger document")
	}
	if len(document.Paths) == 0 {
		return nil, errors.New("artifact has no paths")
	}

	endpoints := make([]batchArtifactEndpoint, 0)
	for path, pathItem := range document.Paths {
		if strings.TrimSpace(path) == "" || !strings.HasPrefix(path, "/") {
			return nil, fmt.Errorf("artifact path %q must be non-empty and connector-relative", path)
		}
		for method, operationRaw := range pathItem {
			method = strings.ToUpper(strings.TrimSpace(method))
			if !batchArtifactHTTPMethod(method) {
				continue
			}
			var operation struct {
				OperationID string `json:"operationId"`
				Summary     string `json:"summary"`
			}
			if len(operationRaw) > 0 {
				if err := json.Unmarshal(operationRaw, &operation); err != nil {
					return nil, fmt.Errorf("artifact %s %s: decode operation: %w", method, path, err)
				}
			}
			summary := strings.TrimSpace(operation.Summary)
			if summary == "" {
				summary = strings.TrimSpace(operation.OperationID)
			}
			if summary == "" {
				summary = fmt.Sprintf("%s %s", method, path)
			}
			endpoints = append(endpoints, batchArtifactEndpoint{Method: method, Path: path, Summary: summary})
		}
	}
	if len(endpoints) == 0 {
		return nil, errors.New("artifact has no HTTP operations")
	}
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Path != endpoints[j].Path {
			return endpoints[i].Path < endpoints[j].Path
		}
		return batchArtifactMethodRank(endpoints[i].Method) < batchArtifactMethodRank(endpoints[j].Method)
	})
	return endpoints, nil
}

// parseBatchYAMLOpenAPIArtifact extracts only the OAS fields this materializer
// needs directly from the YAML node tree. Converting an entire provider YAML
// document through map[any]any is unsafe: valid response status-code keys such
// as 200 are numeric YAML map keys and json.Marshal rejects them. Reading the
// path item nodes also deliberately ignores response bodies/schemas, which are
// not operation inventory metadata.
func parseBatchYAMLOpenAPIArtifact(raw []byte, jsonErr error) ([]batchArtifactEndpoint, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(raw, &document); err != nil {
		return nil, fmt.Errorf("decode artifact as JSON (%v) or YAML (%v)", jsonErr, err)
	}
	root := yamlDocumentRoot(&document)
	if root == nil || root.Kind != yaml.MappingNode {
		return nil, errors.New("YAML artifact root must be a mapping")
	}
	if yamlMappingString(root, "openapi") == "" && yamlMappingString(root, "swagger") == "" {
		return nil, errors.New("artifact is not an OpenAPI or Swagger document")
	}
	paths := yamlMappingValue(root, "paths")
	if paths == nil || paths.Kind != yaml.MappingNode || len(paths.Content) == 0 {
		return nil, errors.New("artifact has no paths")
	}

	endpoints := make([]batchArtifactEndpoint, 0)
	for i := 0; i+1 < len(paths.Content); i += 2 {
		pathNode, pathItem := paths.Content[i], paths.Content[i+1]
		path := strings.TrimSpace(pathNode.Value)
		if path == "" || !strings.HasPrefix(path, "/") {
			return nil, fmt.Errorf("artifact path %q must be non-empty and connector-relative", path)
		}
		if pathItem.Kind != yaml.MappingNode {
			continue
		}
		for j := 0; j+1 < len(pathItem.Content); j += 2 {
			method := strings.ToUpper(strings.TrimSpace(pathItem.Content[j].Value))
			if !batchArtifactHTTPMethod(method) {
				continue
			}
			operation := pathItem.Content[j+1]
			summary := yamlMappingString(operation, "summary")
			if summary == "" {
				summary = yamlMappingString(operation, "operationId")
			}
			if summary == "" {
				summary = fmt.Sprintf("%s %s", method, path)
			}
			endpoints = append(endpoints, batchArtifactEndpoint{Method: method, Path: path, Summary: summary})
		}
	}
	if len(endpoints) == 0 {
		return nil, errors.New("artifact has no HTTP operations")
	}
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Path != endpoints[j].Path {
			return endpoints[i].Path < endpoints[j].Path
		}
		return batchArtifactMethodRank(endpoints[i].Method) < batchArtifactMethodRank(endpoints[j].Method)
	})
	return endpoints, nil
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

func yamlMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func yamlMappingString(node *yaml.Node, key string) string {
	value := yamlMappingValue(node, key)
	if value == nil || value.Kind != yaml.ScalarNode {
		return ""
	}
	return strings.TrimSpace(value.Value)
}

func batchArtifactHTTPMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
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
	default:
		return 7
	}
}

func materializeAPISurface(bundle engine.Bundle, candidate BatchManifestConnector, retrievedAt, sha string, artifactEndpoints []batchArtifactEndpoint) (engine.APISurface, error) {
	if bundle.Surface == nil {
		return engine.APISurface{}, errors.New("api_surface.json is required")
	}
	artifactKeys := make(map[string]bool, len(artifactEndpoints))
	for _, endpoint := range artifactEndpoints {
		artifactKeys[batchArtifactEndpointKey(endpoint.Method, endpoint.Path)] = true
	}
	existingExact := make(map[string]engine.SurfaceEndpoint, len(bundle.Surface.Endpoints))
	existingCanonical := make(map[string][]engine.SurfaceEndpoint, len(bundle.Surface.Endpoints))
	for _, endpoint := range bundle.Surface.Endpoints {
		key := batchArtifactEndpointKey(endpoint.Method, endpoint.Path)
		if _, duplicate := existingExact[key]; duplicate {
			return engine.APISurface{}, fmt.Errorf("existing api surface duplicates %s %s", endpoint.Method, endpoint.Path)
		}
		existingExact[key] = endpoint
		canonical := batchCanonicalEndpointKey(endpoint.Method, endpoint.Path)
		existingCanonical[canonical] = append(existingCanonical[canonical], endpoint)
	}
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.CoveredBy == nil {
			continue
		}
		canonical := batchCanonicalEndpointKey(endpoint.Method, endpoint.Path)
		if !artifactKeys[batchArtifactEndpointKey(endpoint.Method, endpoint.Path)] && len(existingArtifactEndpointsForKey(artifactEndpoints, canonical)) == 0 {
			return engine.APISurface{}, fmt.Errorf("executable coverage %s %s is absent from the cited artifact", endpoint.Method, endpoint.Path)
		}
	}

	artifactID := fmt.Sprintf("%s-artifact-%s", candidate.Connector, retrievedAt)
	surface := engine.APISurface{
		API:                    materializedAPIName(bundle.Surface.API, candidate.Artifact.Version),
		Docs:                   bundle.Surface.Docs,
		ReviewedAt:             retrievedAt,
		OperationLedgerVersion: 2,
		Scope:                  fmt.Sprintf("Provider-artifact inventory generated from the cited %s artifact (%d documented HTTP operations). Existing executable stream/write bindings are retained only when the fetched provider method/path still matches; every other operation is explicitly provider-blocked or excluded.", candidate.Artifact.Kind, len(artifactEndpoints)),
		Artifacts: []engine.SurfaceArtifact{{
			ID:          artifactID,
			URL:         candidate.Artifact.URL,
			RetrievedAt: retrievedAt,
			SHA256:      sha,
		}},
		Endpoints: make([]engine.SurfaceEndpoint, 0, len(artifactEndpoints)),
	}
	for _, artifactEndpoint := range artifactEndpoints {
		endpoint := engine.SurfaceEndpoint{
			Method: artifactEndpoint.Method,
			Path:   artifactEndpoint.Path,
			Provenance: &engine.SurfaceProvenance{
				Artifact:  artifactID,
				SourceURL: candidate.Artifact.URL,
			},
		}
		if existing, ok := existingExact[batchArtifactEndpointKey(artifactEndpoint.Method, artifactEndpoint.Path)]; ok {
			copyMaterializedClassifier(&endpoint, existing)
		} else {
			canonical := batchCanonicalEndpointKey(artifactEndpoint.Method, artifactEndpoint.Path)
			if candidates := existingCanonical[canonical]; len(candidates) == 1 {
				copyMaterializedClassifier(&endpoint, candidates[0])
			} else {
				endpoint.Operation = defaultMaterializedOperation(artifactEndpoint)
			}
		}
		if endpoint.CoveredBy == nil && endpoint.Excluded == nil && endpoint.Operation == nil {
			endpoint.Operation = defaultMaterializedOperation(artifactEndpoint)
		}
		surface.Endpoints = append(surface.Endpoints, endpoint)
	}
	if err := ensureMaterializedCoverage(bundle, surface); err != nil {
		return engine.APISurface{}, err
	}
	return surface, nil
}

func existingArtifactEndpointsForKey(endpoints []batchArtifactEndpoint, canonical string) []batchArtifactEndpoint {
	matched := make([]batchArtifactEndpoint, 0, 1)
	for _, endpoint := range endpoints {
		if batchCanonicalEndpointKey(endpoint.Method, endpoint.Path) == canonical {
			matched = append(matched, endpoint)
		}
	}
	return matched
}

func batchArtifactEndpointKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + "\x00" + strings.TrimSpace(path)
}

func batchCanonicalEndpointKey(method, path string) string {
	cleanPath := strings.TrimSpace(path)
	if cleanPath != "/" {
		cleanPath = strings.TrimSuffix(cleanPath, "/")
	}
	return strings.ToUpper(strings.TrimSpace(method)) + "\x00" + cleanPath
}

func copyMaterializedClassifier(dst *engine.SurfaceEndpoint, src engine.SurfaceEndpoint) {
	if src.CoveredBy != nil {
		coverage := *src.CoveredBy
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
	return &engine.SurfaceOperation{
		Model:            model,
		Status:           "blocked",
		Risk:             risk,
		BlockedByDefault: true,
		Reason:           reason,
	}
}

func batchArtifactMutationMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func defaultMaterializedOperation(endpoint batchArtifactEndpoint) *engine.SurfaceOperation {
	switch endpoint.Method {
	case http.MethodGet, http.MethodHead:
		return &engine.SurfaceOperation{
			Model:            "direct_read",
			Status:           "blocked",
			Risk:             "low",
			BlockedByDefault: true,
			Reason:           "Documented provider read has no matching executable stream. It remains blocked because the current direct-read executor has no non-redacting output policy.",
		}
	case http.MethodDelete:
		return &engine.SurfaceOperation{
			Model:            "destructive_action",
			Status:           "blocked",
			Risk:             "high",
			BlockedByDefault: true,
			Reason:           "Documented provider deletion is not promoted without an operation-specific plan, preview, approval, typed confirmation, and execution contract.",
		}
	default:
		return &engine.SurfaceOperation{
			Model:            "disallowed",
			Status:           "blocked",
			Risk:             "high",
			BlockedByDefault: true,
			Reason:           "Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract in this batch.",
		}
	}
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
		if endpoint.CoveredBy.Write != "" {
			coveredWrites[endpoint.CoveredBy.Write] = true
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

func materializeCLISurface(bundle engine.Bundle, surface engine.APISurface, candidate BatchManifestConnector) (engine.CLISurface, error) {
	streamRefs := map[string][]engine.CLISurfaceEndpointRef{}
	writeRefs := map[string][]engine.CLISurfaceEndpointRef{}
	for _, endpoint := range surface.Endpoints {
		if endpoint.CoveredBy == nil {
			continue
		}
		ref := engine.CLISurfaceEndpointRef{Method: endpoint.Method, Path: endpoint.Path}
		if endpoint.CoveredBy.Stream != "" {
			streamRefs[endpoint.CoveredBy.Stream] = append(streamRefs[endpoint.CoveredBy.Stream], ref)
		}
		if endpoint.CoveredBy.Write != "" {
			writeRefs[endpoint.CoveredBy.Write] = append(writeRefs[endpoint.CoveredBy.Write], ref)
		}
	}

	commands := make([]engine.CLICommand, 0, len(bundle.Streams)+len(bundle.Writes))
	readPaths := make([]string, 0, len(bundle.Streams))
	writePaths := make([]string, 0, len(bundle.Writes))
	for _, stream := range bundle.Streams {
		refs := sortedMaterializedReferences(streamRefs[stream.Name])
		if len(refs) == 0 {
			return engine.CLISurface{}, fmt.Errorf("stream %q has no materialized api_surface reference", stream.Name)
		}
		path := materializedCommandPath(stream.Name, "list")
		commands = append(commands, engine.CLICommand{
			Path:         path,
			Summary:      fmt.Sprintf("Run the %s ETL stream", strings.ReplaceAll(stream.Name, "_", " ")),
			Intent:       "etl",
			Availability: "implemented",
			Stream:       stream.Name,
			SourceURL:    candidate.Artifact.URL,
			APISurface:   refs,
			Examples:     []string{fmt.Sprintf("pm %s %s --json", bundle.Name, path)},
		})
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
		// A reverse-ETL action with a required object/array record cannot be
		// honestly exposed as a scalar CLI-flag command. The action remains
		// executable through the existing generic plan/preview/approval/execute
		// workflow; this connector namespace deliberately advertises only the
		// actions whose required record contract can be derived and validated.
		if !representable {
			continue
		}
		path := materializedCommandPath(action.Name, "apply")
		commands = append(commands, engine.CLICommand{
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
		})
		writePaths = append(writePaths, path)
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

// materializedWriteFlags derives the smallest complete flag contract for a
// reverse-ETL command. It intentionally refuses to flatten object or
// object-array inputs into made-up scalar flags: those actions are still
// available via the typed generic reverse-ETL workflow, but are not promoted
// as a connector command unless every required value has a faithful flag type.
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
