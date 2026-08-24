package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const sourceRetainUsage = `connectorgen source-retain <connector> [--defs <dir>] --retrieved-at <RFC3339> --license <text> --terms <text>

Fetches only the artifacts already identified by a connector-owned operation
source lock, or by a legacy parity source lock when no operation lock exists.
It verifies their preexisting byte counts and SHA-256 values, and stores the
raw bytes under sources/artifacts/. It never changes either lock. Builds and
source-import remain offline: this is an explicit maintenance command only.

  <connector>       connector whose operation or legacy parity source lock is retained
  --defs <dir>      connector defs root (default internal/connectors/defs)
  --retrieved-at    UTC RFC3339 time at which the provider artifact was read
  --license <text>  known license or redistribution status, recorded as data
  --terms <text>    known provider terms or redistribution status, recorded as data`

type sourceRetainOptions struct {
	Connector   string
	DefsDir     string
	RetrievedAt time.Time
	License     string
	Terms       string
}

type sourceRetainPayload struct {
	Artifact sourceImportArtifact
	Raw      []byte
}

func runSourceRetain(args []string, stdout, stderr io.Writer) int {
	return runSourceRetainWithFetcher(args, stdout, stderr, nil)
}

func runSourceRetainWithFetcher(args []string, stdout, stderr io.Writer, fetcher sourceImportFetcher) int {
	for _, arg := range args[1:] {
		if arg == "--help" || arg == "-h" || arg == "help" {
			logln(stdout, sourceRetainUsage)
			return 0
		}
	}
	opts, err := parseSourceRetainOptions(args[1:])
	if err != nil {
		logln(stderr, "connectorgen source-retain:", err)
		logln(stderr, sourceRetainUsage)
		return 2
	}
	artifacts, err := loadConnectorSourceRetainArtifacts(opts.DefsDir, opts.Connector)
	if err != nil {
		logln(stderr, "connectorgen source-retain:", err)
		return 1
	}
	if fetcher == nil {
		fetcher = httpSourceImportFetcher{
			limits: defaultSourceImportLimits(),
			lookup: batchArtifactLookupIPAddr(net.DefaultResolver.LookupIPAddr),
		}
	}
	context, cancel := context.WithTimeout(context.Background(), defaultSourceImportCorpusTimeout)
	defer cancel()
	payloads, err := sourceRetainFetchArtifacts(context, artifacts, fetcher, defaultSourceImportLimits())
	if err != nil {
		logln(stderr, "connectorgen source-retain:", err)
		return 1
	}
	sourcesDir, err := sourceImportConnectorSourcesDir(opts.DefsDir, opts.Connector)
	if err != nil {
		logln(stderr, "connectorgen source-retain:", err)
		return 1
	}
	if err := sourceRetainWritePayloads(sourcesDir, opts, payloads); err != nil {
		logln(stderr, "connectorgen source-retain:", err)
		return 1
	}
	logf(stdout, "connectorgen source-retain: %s, %d artifact(s) retained\n", opts.Connector, len(payloads))
	return 0
}

// sourceRetainParityLock is the pre-operation-lock source provenance retained
// by the Batch 4/5 connectors. It is read only by source-retain so those
// already-pinned artifact identities can be mirrored without changing normal
// source-import's operation-lock contract.
type sourceRetainParityLock struct {
	SchemaVersion   int    `json:"schema_version"`
	Connector       string `json:"connector"`
	SourceRetrieval struct {
		Artifacts []sourceRetainParityArtifact `json:"artifacts"`
	} `json:"source_retrieval"`
}

type sourceRetainParityArtifact struct {
	SourceURL string `json:"source_url"`
	SHA256    string `json:"sha256"`
	Bytes     int64  `json:"bytes"`
}

// loadConnectorSourceRetainArtifacts prefers an operation source lock when it
// exists. The legacy parity fallback is deliberately limited to its already
// declared artifact identities; it neither creates an operation lock nor gives
// source-import a network or legacy-lock fallback.
func loadConnectorSourceRetainArtifacts(defsDir, connector string) ([]sourceImportArtifact, error) {
	sourcesDir, err := sourceImportConnectorSourcesDir(defsDir, connector)
	if err != nil {
		return nil, err
	}
	operationLockPath := filepath.Join(sourcesDir, connector+"-operation-source-lock.json")
	if _, err := os.Lstat(operationLockPath); err == nil {
		lock, loadErr := loadConnectorSourceImportLock(defsDir, connector)
		if loadErr != nil {
			return nil, loadErr
		}
		return sourceRetainLockArtifacts(lock), nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect connector-owned operation source lock: %w", err)
	}
	return loadConnectorParitySourceRetainArtifacts(sourcesDir, connector)
}

func loadConnectorParitySourceRetainArtifacts(sourcesDir, connector string) ([]sourceImportArtifact, error) {
	path := filepath.Join(sourcesDir, connector+"-parity-source-lock.json")
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, fmt.Errorf("resolve connector-owned parity source lock: %w", err)
	}
	if !sourceImportPathWithin(sourcesDir, resolvedPath) {
		return nil, fmt.Errorf("parity source lock is outside connector-owned sources directory")
	}
	raw, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("read connector-owned parity source lock: %w", err)
	}
	var lock sourceRetainParityLock
	if err := decodeSourceJSON(raw, &lock); err != nil {
		return nil, fmt.Errorf("parse parity source lock: %w", err)
	}
	if lock.SchemaVersion != 1 {
		return nil, fmt.Errorf("parity source lock has unsupported schema version %d", lock.SchemaVersion)
	}
	if lock.Connector != connector {
		return nil, fmt.Errorf("parity source lock connector %q does not match requested connector %q", lock.Connector, connector)
	}
	artifacts := make([]sourceImportArtifact, 0, len(lock.SourceRetrieval.Artifacts))
	for _, legacy := range lock.SourceRetrieval.Artifacts {
		parsed, err := url.Parse(legacy.SourceURL)
		if err != nil {
			return nil, fmt.Errorf("parse parity source artifact URL: %w", err)
		}
		artifact := sourceImportArtifact{
			SourceURL:     legacy.SourceURL,
			SHA256:        legacy.SHA256,
			Bytes:         legacy.Bytes,
			IdentityQuery: parsed.RawQuery != "",
		}
		if err := validateSourceImportArtifact(artifact); err != nil {
			return nil, fmt.Errorf("parity source artifact: %w", err)
		}
		artifacts = append(artifacts, artifact)
	}
	if len(artifacts) == 0 {
		return nil, fmt.Errorf("parity source lock has no retainable provider artifacts")
	}
	return sourceRetainUniqueArtifacts(artifacts), nil
}

func parseSourceRetainOptions(args []string) (sourceRetainOptions, error) {
	opts := sourceRetainOptions{DefsDir: filepath.Join("internal", "connectors", "defs")}
	var retrievedAt string
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--defs", "--retrieved-at", "--license", "--terms":
			if index+1 >= len(args) {
				return sourceRetainOptions{}, fmt.Errorf("%s requires a value", arg)
			}
			index++
			switch arg {
			case "--defs":
				opts.DefsDir = args[index]
			case "--retrieved-at":
				retrievedAt = args[index]
			case "--license":
				opts.License = args[index]
			case "--terms":
				opts.Terms = args[index]
			}
		default:
			if strings.HasPrefix(arg, "-") {
				return sourceRetainOptions{}, fmt.Errorf("unknown flag %q", arg)
			}
			if opts.Connector != "" {
				return sourceRetainOptions{}, fmt.Errorf("only one connector may be retained at a time")
			}
			opts.Connector = arg
		}
	}
	if err := validateSourceImportConnector(opts.Connector); err != nil {
		return sourceRetainOptions{}, err
	}
	if retrievedAt == "" {
		return sourceRetainOptions{}, fmt.Errorf("--retrieved-at is required")
	}
	parsedRetrievedAt, err := time.Parse(time.RFC3339, retrievedAt)
	if err != nil || parsedRetrievedAt.Format(time.RFC3339) != retrievedAt {
		return sourceRetainOptions{}, fmt.Errorf("--retrieved-at must be a normalized RFC3339 timestamp")
	}
	opts.RetrievedAt = parsedRetrievedAt.UTC()
	if err := validateSourceRetainProvenance("license", opts.License); err != nil {
		return sourceRetainOptions{}, err
	}
	if err := validateSourceRetainProvenance("terms", opts.Terms); err != nil {
		return sourceRetainOptions{}, err
	}
	return opts, nil
}

func validateSourceRetainProvenance(name, value string) error {
	if value == "" || value != strings.TrimSpace(value) || len(value) > 4096 || strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("--%s must be non-empty provenance text", name)
	}
	return nil
}

// sourceRetainLockArtifacts reports every raw provider artifact that the
// normal importer verifies. It intentionally derives identity from the lock,
// never a user-provided URL or artifact path. Version-three unavailable
// documents are excluded because source-import terminally refuses them rather
// than treating their optional diagnostic capture as an importable spec.
func sourceRetainLockArtifacts(lock sourceImportLock) []sourceImportArtifact {
	artifacts := make([]sourceImportArtifact, 0, len(lock.Rest.SourceDocuments)+2)
	if lock.SchemaVersion < 3 {
		artifacts = append(artifacts, lock.Rest.sourceImportArtifact)
	} else {
		for _, document := range lock.Rest.SourceDocuments {
			if !document.isUnavailable() {
				artifacts = append(artifacts, document.Artifact)
			}
		}
	}
	if len(lock.GraphQL.QueryFields)+len(lock.GraphQL.MutationFields) > 0 {
		artifacts = append(artifacts, lock.GraphQL.sourceImportArtifact)
	}
	return sourceRetainUniqueArtifacts(artifacts)
}

func sourceRetainUniqueArtifacts(artifacts []sourceImportArtifact) []sourceImportArtifact {
	seen := make(map[string]bool, len(artifacts))
	unique := make([]sourceImportArtifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		key := strings.ToLower(artifact.SHA256) + "\x00" + artifact.SourceURL + "\x00" + fmt.Sprint(artifact.IdentityQuery)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, artifact)
		}
	}
	sort.Slice(unique, func(left, right int) bool {
		if strings.EqualFold(unique[left].SHA256, unique[right].SHA256) {
			return unique[left].SourceURL < unique[right].SourceURL
		}
		return strings.ToLower(unique[left].SHA256) < strings.ToLower(unique[right].SHA256)
	})
	return unique
}

func sourceRetainFetchLockedArtifacts(ctx context.Context, lock sourceImportLock, fetcher sourceImportFetcher, limits sourceImportLimits) ([]sourceRetainPayload, error) {
	return sourceRetainFetchArtifacts(ctx, sourceRetainLockArtifacts(lock), fetcher, limits)
}

func sourceRetainFetchArtifacts(ctx context.Context, artifacts []sourceImportArtifact, fetcher sourceImportFetcher, limits sourceImportLimits) ([]sourceRetainPayload, error) {
	if fetcher == nil {
		return nil, fmt.Errorf("source retention has no fetcher")
	}
	payloads := make([]sourceRetainPayload, 0, len(artifacts))
	for _, artifact := range artifacts {
		raw, err := fetchSourceImportArtifact(ctx, fetcher, artifact)
		if err != nil {
			return nil, fmt.Errorf("fetch locked source artifact %s: %w", artifact.SourceURL, err)
		}
		if int64(len(raw)) > limits.MaxArtifactBytes {
			return nil, fmt.Errorf("retained source artifact exceeds byte limit")
		}
		if err := validateSourceImportArtifactBytes(raw, artifact); err != nil {
			return nil, err
		}
		payloads = append(payloads, sourceRetainPayload{Artifact: artifact, Raw: raw})
	}
	return payloads, nil
}

func sourceRetainWritePayloads(sourcesDir string, opts sourceRetainOptions, payloads []sourceRetainPayload) error {
	if len(payloads) == 0 {
		return fmt.Errorf("source lock has no retainable provider artifacts")
	}
	for _, payload := range payloads {
		if err := sourceRetainWriteArtifact(sourcesDir, payload); err != nil {
			return err
		}
	}
	return sourceRetainWriteManifest(sourcesDir, opts, payloads)
}

func sourceRetainWriteArtifact(sourcesDir string, payload sourceRetainPayload) error {
	if err := validateSourceImportArtifactBytes(payload.Raw, payload.Artifact); err != nil {
		return err
	}
	artifactDir := filepath.Join(sourcesDir, sourceImportRetainedArtifactDirectory)
	if err := sourceRetainEnsureDirectory(artifactDir); err != nil {
		return fmt.Errorf("prepare retained source artifact directory: %w", err)
	}
	artifactPath := filepath.Join(artifactDir, strings.ToLower(payload.Artifact.SHA256)+sourceImportRetainedArtifactExtension)
	info, err := os.Lstat(artifactPath)
	if err == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("existing retained source artifact must be a regular non-symlink file")
		}
		existing, readErr := os.ReadFile(artifactPath)
		if readErr != nil {
			return fmt.Errorf("read existing retained source artifact: %w", readErr)
		}
		if validateErr := validateSourceImportArtifactBytes(existing, payload.Artifact); validateErr != nil {
			return fmt.Errorf("existing retained source artifact does not match lock; refusing replacement: %w", validateErr)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("inspect retained source artifact: %w", err)
	}
	temp, err := os.CreateTemp(artifactDir, ".source-retain-*")
	if err != nil {
		return fmt.Errorf("stage retained source artifact: %w", err)
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return fmt.Errorf("set retained source artifact mode: %w", err)
	}
	if written, writeErr := temp.Write(payload.Raw); writeErr != nil || written != len(payload.Raw) {
		_ = temp.Close()
		if writeErr != nil {
			return fmt.Errorf("stage retained source artifact: %w", writeErr)
		}
		return fmt.Errorf("stage retained source artifact: %w", io.ErrShortWrite)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync retained source artifact: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close retained source artifact: %w", err)
	}
	if err := os.Link(tempPath, artifactPath); err != nil {
		return fmt.Errorf("publish retained source artifact without replacement: %w", err)
	}
	return sourceRetainSyncDirectory(artifactDir)
}

func sourceRetainWriteManifest(sourcesDir string, opts sourceRetainOptions, payloads []sourceRetainPayload) error {
	if err := sourceRetainEnsureDirectory(sourcesDir); err != nil {
		return fmt.Errorf("inspect connector-owned sources directory: %w", err)
	}
	manifestPath := filepath.Join(sourcesDir, opts.Connector+sourceImportRetainedArtifactManifest)
	manifest := sourceImportRetainedArtifactManifestDocument{SchemaVersion: 1, Connector: opts.Connector, Artifacts: []sourceImportRetainedArtifactRecord{}}
	info, err := os.Lstat(manifestPath)
	if err == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("retained artifact manifest must be a regular non-symlink file")
		}
		raw, readErr := os.ReadFile(manifestPath)
		if readErr != nil {
			return fmt.Errorf("read retained artifact manifest: %w", readErr)
		}
		if parseErr := decodeSourceStrictJSON(raw, &manifest); parseErr != nil {
			return fmt.Errorf("parse retained artifact manifest: %w", parseErr)
		}
		if manifest.SchemaVersion != 1 || manifest.Connector != opts.Connector {
			return fmt.Errorf("retained artifact manifest identity does not match connector")
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect retained artifact manifest: %w", err)
	}

	seen := make(map[string]sourceImportRetainedArtifactRecord, len(manifest.Artifacts))
	for _, record := range manifest.Artifacts {
		if err := sourceRetainValidateManifestRecord(record); err != nil {
			return fmt.Errorf("existing retained artifact manifest record: %w", err)
		}
		key := sourceRetainRecordKey(record.SHA256, record.SourceURL, record.IdentityQuery)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("existing retained artifact manifest duplicates provenance identity")
		}
		seen[key] = record
	}
	for _, payload := range payloads {
		key := sourceRetainRecordKey(payload.Artifact.SHA256, payload.Artifact.SourceURL, payload.Artifact.IdentityQuery)
		if existing, exists := seen[key]; exists {
			if existing.Bytes != payload.Artifact.Bytes {
				return fmt.Errorf("existing retained artifact manifest conflicts with locked bytes")
			}
			continue
		}
		record := sourceImportRetainedArtifactRecord{
			SHA256:        strings.ToLower(payload.Artifact.SHA256),
			Bytes:         payload.Artifact.Bytes,
			SourceURL:     payload.Artifact.SourceURL,
			IdentityQuery: payload.Artifact.IdentityQuery,
			RetrievedAt:   opts.RetrievedAt.Format(time.RFC3339),
			License:       opts.License,
			Terms:         opts.Terms,
		}
		manifest.Artifacts = append(manifest.Artifacts, record)
		seen[key] = record
	}
	sort.Slice(manifest.Artifacts, func(left, right int) bool {
		leftKey := sourceRetainRecordKey(manifest.Artifacts[left].SHA256, manifest.Artifacts[left].SourceURL, manifest.Artifacts[left].IdentityQuery)
		rightKey := sourceRetainRecordKey(manifest.Artifacts[right].SHA256, manifest.Artifacts[right].SourceURL, manifest.Artifacts[right].IdentityQuery)
		return leftKey < rightKey
	})
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode retained artifact manifest: %w", err)
	}
	return sourceRetainReplaceRegularFile(manifestPath, append(raw, '\n'))
}

func sourceRetainValidateManifestRecord(record sourceImportRetainedArtifactRecord) error {
	if err := validateSourceImportArtifact(sourceImportArtifact{
		SourceURL:     record.SourceURL,
		SHA256:        record.SHA256,
		Bytes:         record.Bytes,
		IdentityQuery: record.IdentityQuery,
	}); err != nil {
		return fmt.Errorf("identity: %w", err)
	}
	if _, err := time.Parse(time.RFC3339, record.RetrievedAt); err != nil {
		return fmt.Errorf("retrieval timestamp is invalid: %w", err)
	}
	if err := validateSourceRetainProvenance("license", record.License); err != nil {
		return err
	}
	return validateSourceRetainProvenance("terms", record.Terms)
}

func sourceRetainRecordKey(sha256, sourceURL string, identityQuery bool) string {
	return strings.ToLower(sha256) + "\x00" + sourceURL + "\x00" + fmt.Sprint(identityQuery)
}

func sourceRetainEnsureDirectory(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		if err := os.Mkdir(path, 0o755); err != nil && !os.IsExist(err) {
			return err
		}
		info, err = os.Lstat(path)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("path must be a non-symlink directory")
	}
	return nil
}

func sourceRetainReplaceRegularFile(path string, raw []byte) error {
	if info, err := os.Lstat(path); err == nil && (!info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0) {
		return fmt.Errorf("path must be a regular non-symlink file")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	directory := filepath.Dir(path)
	temp, err := os.CreateTemp(directory, ".source-retain-manifest-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return err
	}
	if written, writeErr := temp.Write(raw); writeErr != nil || written != len(raw) {
		_ = temp.Close()
		if writeErr != nil {
			return writeErr
		}
		return io.ErrShortWrite
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	return sourceRetainSyncDirectory(directory)
}

func sourceRetainSyncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	if err := directory.Sync(); err != nil {
		_ = directory.Close()
		return err
	}
	return directory.Close()
}
