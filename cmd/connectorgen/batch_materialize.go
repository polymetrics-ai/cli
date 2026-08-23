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
	artifactEndpoints, err := parseBatchOpenAPIArtifact(rawArtifact)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, batchArtifactDropStage(err, "artifact_parse"), err)
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
	materializedBundle := bundle
	materializedBundle.Surface = &surface
	materializedBundle.CLISurface = &cli
	split, err := batchSurfaceSplit(&surface)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "api_surface", err)
	}
	checked, err := batchRuntimePreflight(materializedBundle)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "runtime_preflight", err)
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
	destination, err := createBatchMaterializeDestination(sourceBundleDir, bundleDir)
	if err != nil {
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, batchMaterializeDestinationStage(err), err)
	}
	if err := writeBatchMaterializedFiles(bundleDir, append(surfaceRaw, '\n'), operationsRaw, append(cliRaw, '\n')); err != nil {
		destination.discard()
		return BatchMaterializeIncluded{}, batchGateDrop(candidate.Connector, "write", fmt.Errorf("write materialized bundle: %w", err))
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
		CLICommands:        checked,
		OperationExecutors: 0,
	}, nil
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
	parsed, err := parseBatchArtifactURL(rawURL)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	lookup := batchArtifactLookupIPAddr(net.DefaultResolver.LookupIPAddr)
	if err := validateBatchArtifactRequestURL(ctx, parsed, lookup); err != nil {
		return nil, fmt.Errorf("validate artifact destination: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build artifact request: %w", err)
	}
	response, err := newBatchArtifactHTTPClient(lookup).Do(request)
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

type batchArtifactURLPolicy struct {
	allowIdentityQuery bool
}

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
	return parseBatchArtifactURLWithPolicy(raw, batchArtifactURLPolicy{})
}

func parseBatchArtifactURLWithPolicy(raw string, policy batchArtifactURLPolicy) (*url.URL, error) {
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
	if !policy.allowIdentityQuery && (parsed.RawQuery != "" || parsed.ForceQuery) {
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

func validateBatchArtifactRequestURL(ctx context.Context, parsed *url.URL, lookup batchArtifactLookupIPAddr) error {
	return validateBatchArtifactRequestURLWithPolicy(ctx, parsed, lookup, batchArtifactURLPolicy{})
}

func validateBatchArtifactRequestURLWithPolicy(ctx context.Context, parsed *url.URL, lookup batchArtifactLookupIPAddr, policy batchArtifactURLPolicy) error {
	if err := validateBatchArtifactURLObject(parsed, policy); err != nil {
		return err
	}
	_, err := batchArtifactPublicAddresses(ctx, parsed.Hostname(), lookup)
	return err
}

func validateBatchArtifactURLObject(parsed *url.URL, policy batchArtifactURLPolicy) error {
	if parsed == nil || !parsed.IsAbs() || parsed.Scheme != "https" || parsed.Host == "" {
		return errors.New("artifact request URL must be an absolute HTTPS URL")
	}
	if parsed.User != nil {
		return errors.New("artifact request URL must not include userinfo")
	}
	if (!policy.allowIdentityQuery && (parsed.RawQuery != "" || parsed.ForceQuery)) || parsed.Fragment != "" {
		return errors.New("artifact request URL must not include query or fragment components")
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

func newBatchArtifactHTTPClient(lookup batchArtifactLookupIPAddr) *http.Client {
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
			return validateBatchArtifactRequestURL(request.Context(), request.URL, lookup)
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
	Method  string
	Path    string
	Summary string
}

func parseBatchOpenAPIArtifact(raw []byte) ([]batchArtifactEndpoint, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode OpenAPI or Swagger artifact: %w", err)
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("decode OpenAPI or Swagger artifact: %w", err)
		}
		return nil, batchArtifactInventoryUnknown("artifact contains multiple YAML documents")
	}
	root, err := batchYAMLDeref(yamlDocumentRoot(&document))
	if err != nil {
		return nil, err
	}
	fields, err := batchYAMLFields(root)
	if err != nil {
		return nil, errors.New("artifact root must be a mapping")
	}
	openAPI, err := batchYAMLFieldString(fields, "openapi")
	if err != nil {
		return nil, fmt.Errorf("artifact openapi field: %w", err)
	}
	swagger, err := batchYAMLFieldString(fields, "swagger")
	if err != nil {
		return nil, fmt.Errorf("artifact swagger field: %w", err)
	}
	if openAPI == "" && swagger == "" {
		return nil, errors.New("artifact is not an OpenAPI or Swagger document")
	}
	if webhooks, ok := fields["webhooks"]; ok {
		if err := batchArtifactWebhooksUnknown(webhooks); err != nil {
			return nil, err
		}
	}
	paths, ok := fields["paths"]
	if !ok {
		return nil, errors.New("artifact has no paths")
	}
	return batchArtifactEndpointsFromPaths(root, paths)
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

func batchArtifactWebhooksUnknown(node *yaml.Node) error {
	fields, err := batchYAMLFields(node)
	if err != nil {
		return batchArtifactInventoryUnknown("top-level webhooks is not a mapping")
	}
	names := make([]string, 0, len(fields))
	for _, name := range batchYAMLFieldNames(fields) {
		if strings.HasPrefix(name, "x-") {
			continue
		}
		if strings.TrimSpace(name) == "" {
			return batchArtifactInventoryUnknown("top-level webhooks has an empty name")
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil
	}
	return batchArtifactInventoryUnknown("top-level webhooks (%s) cannot be represented as provider request paths", strings.Join(names, ", "))
}

func batchArtifactEndpointsFromPaths(root, paths *yaml.Node) ([]batchArtifactEndpoint, error) {
	fields, err := batchYAMLFields(paths)
	if err != nil {
		return nil, errors.New("artifact paths must be a mapping")
	}
	endpoints := make([]batchArtifactEndpoint, 0)
	seen := map[string]bool{}
	for _, path := range batchYAMLFieldNames(fields) {
		if strings.HasPrefix(path, "x-") {
			continue
		}
		if path == "" || path != strings.TrimSpace(path) || !strings.HasPrefix(path, "/") {
			return nil, fmt.Errorf("artifact path %q must be non-empty and connector-relative", path)
		}
		pathEndpoints, err := batchArtifactPathItemEndpoints(root, path, fields[path])
		if err != nil {
			return nil, err
		}
		for _, endpoint := range pathEndpoints {
			key := batchArtifactEndpointKey(endpoint.Method, endpoint.Path)
			if seen[key] {
				return nil, batchArtifactInventoryUnknown("duplicate operation %s %s", endpoint.Method, endpoint.Path)
			}
			seen[key] = true
			endpoints = append(endpoints, endpoint)
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

func batchArtifactPathItemEndpoints(root *yaml.Node, path string, pathItem *yaml.Node) ([]batchArtifactEndpoint, error) {
	resolved, err := resolveBatchArtifactPathItem(root, pathItem, map[string]bool{})
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

func resolveBatchArtifactPathItem(root, pathItem *yaml.Node, refs map[string]bool) (*yaml.Node, error) {
	pathItem, err := batchYAMLDeref(pathItem)
	if err != nil {
		return nil, err
	}
	fields, err := batchYAMLFields(pathItem)
	if err != nil {
		return nil, batchArtifactInventoryUnknown("path item is not a mapping")
	}
	refNode, hasRef := fields["$ref"]
	if !hasRef {
		return pathItem, nil
	}
	ref, err := batchYAMLFieldString(fields, "$ref")
	if err != nil || ref == "" {
		return nil, batchArtifactInventoryUnknown("path-item reference must be a non-empty string")
	}
	for _, key := range batchYAMLFieldNames(fields) {
		if key == "$ref" || key == "summary" || key == "description" || strings.HasPrefix(key, "x-") {
			continue
		}
		return nil, batchArtifactInventoryUnknown("path-item reference %q has unsupported sibling %q", ref, key)
	}
	if refs[ref] {
		return nil, batchArtifactInventoryUnknown("path-item reference cycle at %q", ref)
	}
	refs[ref] = true
	defer delete(refs, ref)
	target, err := resolveBatchArtifactLocalReference(root, ref)
	if err != nil {
		return nil, err
	}
	if target == refNode {
		return nil, batchArtifactInventoryUnknown("path-item reference %q resolves to itself", ref)
	}
	return resolveBatchArtifactPathItem(root, target, refs)
}

func resolveBatchArtifactLocalReference(root *yaml.Node, reference string) (*yaml.Node, error) {
	if !strings.HasPrefix(reference, "#") {
		return nil, batchArtifactInventoryUnknown("external path-item reference %q cannot be exhaustively resolved", reference)
	}
	pointer, err := url.PathUnescape(strings.TrimPrefix(reference, "#"))
	if err != nil || pointer == "" || !strings.HasPrefix(pointer, "/") {
		return nil, batchArtifactInventoryUnknown("local path-item reference %q is not a JSON pointer", reference)
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

func materializeAPISurface(bundle engine.Bundle, candidate BatchManifestConnector, retrievedAt, sha string, artifactEndpoints []batchArtifactEndpoint) (engine.APISurface, error) {
	if bundle.Surface == nil {
		return engine.APISurface{}, errors.New("api_surface.json is required")
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
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.CoveredBy == nil {
			continue
		}
		if !artifactKeys[batchArtifactEndpointKey(endpoint.Method, endpoint.Path)] {
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
	if err := ensureMaterializedCoverage(bundle, surface); err != nil {
		return engine.APISurface{}, err
	}
	return surface, nil
}

func batchArtifactEndpointKey(method, path string) string {
	return method + "\x00" + path
}

func copyMaterializedClassifier(dst *engine.SurfaceEndpoint, src engine.SurfaceEndpoint) {
	if src.CoveredBy != nil {
		coverage := *src.CoveredBy
		coverage.Writes = append([]string(nil), src.CoveredBy.Writes...)
		coverage.DirectReads = append([]string(nil), src.CoveredBy.DirectReads...)
		coverage.Writes = append([]string(nil), src.CoveredBy.Writes...)
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
	switch method {
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
		for _, write := range endpoint.CoveredBy.WriteTargets() {
			writeRefs[write] = append(writeRefs[write], ref)
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
		// A reverse-ETL action is advertised only when every required field has a
		// faithful declaration-bound flag contract. Required containers use the
		// existing `json` flag type: commandrunner proves that it names one
		// top-level object/array field of this exact closed record schema before
		// it can form a plan. This is deliberately not a raw request-body route.
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
// reverse-ETL command. Scalar leaves keep their scalar flags. A required
// object or array is represented by one `json` flag on its top-level record
// property, which is the narrow structured-input form commandrunner preflights
// against the action's own closed schema. It never flattens a container into
// invented scalar flags or admits an arbitrary request body.
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
		materializedPath, err := materializedRequiredFlagPath(schema, path)
		if err != nil {
			return nil, false, err
		}
		required[materializedPath] = true
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

// materializedRequiredFlagPath turns a required descendant of a top-level
// object/array into that container's one typed JSON flag. The runtime only
// accepts JSON at this exact top-level record boundary, so keeping the
// normalization here makes the materializer and preflight agree.
func materializedRequiredFlagPath(schema *cliRecordSchemaNode, path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("required record path is empty")
	}
	topLevel, _, _ := strings.Cut(path, ".")
	node, err := schema.recordPath(topLevel)
	if err != nil {
		return "", err
	}
	if node.isObject() || node.isArray() {
		return topLevel, nil
	}
	if topLevel != path {
		return "", fmt.Errorf("required nested scalar path %q has no declared container", path)
	}
	return path, nil
}

func materializedWriteFlag(schema *cliRecordSchemaNode, path string) (engine.CLIFlag, bool, error) {
	if strings.Contains(path, ".") {
		return engine.CLIFlag{}, false, fmt.Errorf("materialized flag path %q must be top-level", path)
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
		if node.items != nil && len(node.items.effectiveTypes()) == 1 && node.items.effectiveTypes()[0] == "string" {
			return "string_array", true
		}
		return "json", true
	case "object":
		return "json", true
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
