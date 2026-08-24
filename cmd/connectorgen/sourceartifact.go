package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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
or parity source lock, verifies the lock-selected identity, and stores the raw
bytes under sources/artifacts/. Byte identity is exact size plus SHA-256;
canonical_json identity key-sorts parsed JSON before SHA-256 comparison. It
records the fetched byte identity and detected form/version (or undetermined)
without changing a source lock. A denied, redirected, login, or drastically
undersized response is reported as wrong source rather than drift. Builds and
source-import remain offline: this is an explicit maintenance command only.

  <connector>       connector whose source lock is retained
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
	Artifact       sourceImportArtifact
	Raw            []byte
	RetainedSHA256 string
	RetainedBytes  int64
	Form           string
	Version        string
}

type sourceRetainLock struct {
	Connector string
	Artifacts []sourceImportArtifact
}

// sourceRetainWrongSourceError names an upstream response that cannot safely
// become source material. It is intentionally distinct from source drift.
type sourceRetainWrongSourceError struct {
	Reason string
}

func (err sourceRetainWrongSourceError) Error() string {
	return "wrong source: " + err.Reason
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
	lock, err := loadConnectorSourceRetainLock(opts.DefsDir, opts.Connector)
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
	payloads, err := sourceRetainFetchLockedArtifacts(context, lock, fetcher, defaultSourceImportLimits())
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

// loadConnectorSourceRetainLock parses only the fetch-and-preserve facts in a
// connector-owned source lock. Source import remains responsible for complete
// operation inventories, OpenAPI form pins, and aggregate version coverage.
func loadConnectorSourceRetainLock(defsDir, connector string) (sourceRetainLock, error) {
	sourcesDir, err := sourceImportConnectorSourcesDir(defsDir, connector)
	if err != nil {
		return sourceRetainLock{}, err
	}
	for _, name := range []string{connector + "-operation-source-lock.json", connector + "-parity-source-lock.json"} {
		path := filepath.Join(sourcesDir, name)
		info, statErr := os.Lstat(path)
		if os.IsNotExist(statErr) {
			continue
		}
		if statErr != nil {
			return sourceRetainLock{}, fmt.Errorf("inspect connector-owned source lock: %w", statErr)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return sourceRetainLock{}, fmt.Errorf("connector-owned source lock must be a regular non-symlink file")
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return sourceRetainLock{}, fmt.Errorf("read connector-owned source lock: %w", readErr)
		}
		lock, parseErr := parseSourceRetainLock(raw, connector)
		if parseErr != nil {
			return sourceRetainLock{}, parseErr
		}
		return lock, nil
	}
	return sourceRetainLock{}, fmt.Errorf("connector-owned source lock is missing (expected %s-operation-source-lock.json or %s-parity-source-lock.json)", connector, connector)
}

func parseSourceRetainLock(raw []byte, expectedConnector string) (sourceRetainLock, error) {
	var header struct {
		SchemaVersion int             `json:"schema_version"`
		Connector     string          `json:"connector"`
		Rest          json.RawMessage `json:"rest"`
		GraphQL       json.RawMessage `json:"graphql"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return sourceRetainLock{}, fmt.Errorf("parse source lock: %w", err)
	}
	if header.SchemaVersion <= 0 {
		return sourceRetainLock{}, fmt.Errorf("source lock has invalid schema version")
	}
	if header.Connector == "" || (expectedConnector != "" && header.Connector != expectedConnector) {
		return sourceRetainLock{}, fmt.Errorf("source lock connector %q does not match requested connector %q", header.Connector, expectedConnector)
	}
	if err := validateSourceImportConnector(header.Connector); err != nil {
		return sourceRetainLock{}, err
	}
	artifacts, err := sourceRetainRESTArtifacts(header.Rest)
	if err != nil {
		return sourceRetainLock{}, err
	}
	if artifact, found, graphqlErr := sourceRetainArtifactFromRaw(header.GraphQL); graphqlErr != nil {
		return sourceRetainLock{}, fmt.Errorf("source lock GraphQL artifact: %w", graphqlErr)
	} else if found {
		artifacts = append(artifacts, artifact)
	}
	artifacts, err = sourceRetainUniqueArtifacts(artifacts)
	if err != nil {
		return sourceRetainLock{}, err
	}
	if len(artifacts) == 0 {
		return sourceRetainLock{}, fmt.Errorf("source lock has no retainable provider artifacts")
	}
	return sourceRetainLock{Connector: header.Connector, Artifacts: artifacts}, nil
}

func sourceRetainRESTArtifacts(raw json.RawMessage) ([]sourceImportArtifact, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, fmt.Errorf("source lock has no REST artifact")
	}
	var rest map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rest); err != nil {
		return nil, fmt.Errorf("parse source lock REST artifact: %w", err)
	}
	if documentsRaw, found := rest["source_documents"]; found {
		var documents []struct {
			Kind     string          `json:"kind"`
			Artifact json.RawMessage `json:"artifact"`
		}
		if err := json.Unmarshal(documentsRaw, &documents); err != nil {
			return nil, fmt.Errorf("parse source lock REST documents: %w", err)
		}
		artifacts := make([]sourceImportArtifact, 0, len(documents))
		for _, document := range documents {
			if document.Kind == sourceImportDocumentKindUnavailable {
				continue
			}
			artifact, found, err := sourceRetainArtifactFromRaw(document.Artifact)
			if err != nil {
				return nil, fmt.Errorf("parse source lock REST document artifact: %w", err)
			}
			if !found {
				return nil, fmt.Errorf("source lock REST document has no artifact")
			}
			artifacts = append(artifacts, artifact)
		}
		return artifacts, nil
	}
	artifact, found, err := sourceRetainArtifactFromRaw(raw)
	if err != nil {
		return nil, fmt.Errorf("parse source lock REST artifact: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("source lock has no REST artifact")
	}
	return []sourceImportArtifact{artifact}, nil
}

func sourceRetainArtifactFromRaw(raw json.RawMessage) (sourceImportArtifact, bool, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return sourceImportArtifact{}, false, nil
	}
	var artifact sourceImportArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return sourceImportArtifact{}, false, err
	}
	if artifact.SourceURL == "" && artifact.SHA256 == "" && artifact.Bytes == 0 {
		return sourceImportArtifact{}, false, nil
	}
	// Parity locks predate the import-only identity_query field, but their URL
	// is already connector-owned. Retention may preserve that fixed bounded
	// query; source-import keeps its stricter versioned query admission.
	if !artifact.IdentityQuery {
		if parsed, parseErr := url.Parse(artifact.SourceURL); parseErr == nil && parsed.RawQuery != "" {
			artifact.IdentityQuery = true
		}
	}
	return artifact, true, nil
}

func sourceRetainUniqueArtifacts(artifacts []sourceImportArtifact) ([]sourceImportArtifact, error) {
	seen := make(map[string]bool, len(artifacts))
	unique := make([]sourceImportArtifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if err := validateSourceRetainArtifact(artifact); err != nil {
			return nil, err
		}
		key := sourceRetainArtifactLockKey(artifact)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, artifact)
		}
	}
	sort.Slice(unique, func(left, right int) bool {
		return sourceRetainArtifactLockKey(unique[left]) < sourceRetainArtifactLockKey(unique[right])
	})
	return unique, nil
}

func sourceRetainArtifactLockKey(artifact sourceImportArtifact) string {
	return sourceArtifactIdentity(artifact) + "\x00" + strings.ToLower(artifact.SHA256) + "\x00" + strings.ToLower(artifact.CanonicalSHA256) + "\x00" + artifact.SourceURL + "\x00" + fmt.Sprint(artifact.IdentityQuery)
}

func sourceRetainFetchLockedArtifacts(ctx context.Context, lock sourceRetainLock, fetcher sourceImportFetcher, limits sourceImportLimits) ([]sourceRetainPayload, error) {
	if fetcher == nil {
		return nil, fmt.Errorf("source retention has no fetcher")
	}
	payloads := make([]sourceRetainPayload, 0, len(lock.Artifacts))
	for _, artifact := range lock.Artifacts {
		raw, err := fetchSourceImportArtifact(ctx, fetcher, artifact)
		if err != nil {
			if sourceRetainFetchIsWrongSource(err) {
				return nil, fmt.Errorf("wrong source for locked artifact %s: %w", artifact.SourceURL, err)
			}
			return nil, fmt.Errorf("fetch locked source artifact %s: %w", artifact.SourceURL, err)
		}
		if int64(len(raw)) > limits.MaxArtifactBytes {
			return nil, fmt.Errorf("retained source artifact exceeds byte limit")
		}
		if err := sourceRetainClassifyFetchedArtifact(raw, artifact); err != nil {
			return nil, err
		}
		if err := validateSourceImportArtifactBytes(raw, artifact); err != nil {
			return nil, err
		}
		retainedDigest := sha256.Sum256(raw)
		form, version := sourceRetainDetectArtifactForm(raw)
		payloads = append(payloads, sourceRetainPayload{Artifact: artifact, Raw: raw, RetainedSHA256: hex.EncodeToString(retainedDigest[:]), RetainedBytes: int64(len(raw)), Form: form, Version: version})
	}
	return payloads, nil
}

func sourceRetainFetchIsWrongSource(err error) bool {
	var wrongSource sourceRetainWrongSourceError
	if errors.As(err, &wrongSource) {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "http 401") || strings.Contains(lower, "http 403") || strings.Contains(lower, "http 404") || strings.Contains(lower, "redirect") || strings.Contains(lower, "login") || strings.Contains(lower, "sign in")
}

func sourceRetainClassifyFetchedArtifact(raw []byte, artifact sourceImportArtifact) error {
	lower := bytes.ToLower(bytes.TrimSpace(raw))
	drasticallySmaller := int64(len(raw))*8 < artifact.Bytes
	if drasticallySmaller && (bytes.HasPrefix(lower, []byte("<!doctype html")) || bytes.HasPrefix(lower, []byte("<html"))) && (bytes.Contains(lower, []byte("login")) || bytes.Contains(lower, []byte("sign in")) || bytes.Contains(lower, []byte("sign-in"))) {
		return sourceRetainWrongSourceError{Reason: "received a login wall rather than the locked provider artifact"}
	}
	if drasticallySmaller {
		return sourceRetainWrongSourceError{Reason: fmt.Sprintf("received %d bytes where the lock records %d bytes; the URL likely resolves to a landing page or error response", len(raw), artifact.Bytes)}
	}
	return nil
}

func sourceRetainDetectArtifactForm(raw []byte) (string, string) {
	if _, form, err := parseSourceImportDocument(raw); err == nil {
		return form.Family, form.Version
	}
	var jsonDocument any
	if err := decodeSourceJSON(raw, &jsonDocument); err == nil {
		return "json", "undetermined"
	}
	var yamlDocument any
	if err := decodeSourceYAML(raw, &yamlDocument); err == nil {
		return "yaml", "undetermined"
	}
	return "undetermined", "undetermined"
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
	retainedDigest := sha256.Sum256(payload.Raw)
	if payload.RetainedBytes != int64(len(payload.Raw)) || !strings.EqualFold(payload.RetainedSHA256, hex.EncodeToString(retainedDigest[:])) {
		return fmt.Errorf("retained source artifact payload has inconsistent fetched-byte provenance")
	}
	artifactDir := filepath.Join(sourcesDir, sourceImportRetainedArtifactDirectory)
	if err := sourceRetainEnsureDirectory(artifactDir); err != nil {
		return fmt.Errorf("prepare retained source artifact directory: %w", err)
	}
	artifactPath := filepath.Join(artifactDir, strings.ToLower(payload.RetainedSHA256)+sourceImportRetainedArtifactExtension)
	info, err := os.Lstat(artifactPath)
	if err == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("existing retained source artifact must be a regular non-symlink file")
		}
		existing, readErr := os.ReadFile(artifactPath)
		if readErr != nil {
			return fmt.Errorf("read existing retained source artifact: %w", readErr)
		}
		existingDigest := sha256.Sum256(existing)
		if int64(len(existing)) != payload.RetainedBytes || !strings.EqualFold(hex.EncodeToString(existingDigest[:]), payload.RetainedSHA256) {
			return fmt.Errorf("existing retained source artifact does not match recorded fetched bytes; refusing replacement")
		}
		if validateErr := validateSourceImportArtifactBytes(existing, payload.Artifact); validateErr != nil {
			return fmt.Errorf("existing retained source artifact does not match source lock identity; refusing replacement: %w", validateErr)
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
		key := sourceRetainRecordKey(record)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("existing retained artifact manifest duplicates fetched-byte provenance")
		}
		seen[key] = record
	}
	for _, payload := range payloads {
		record := sourceImportRetainedArtifactRecord{
			SHA256:          strings.ToLower(payload.Artifact.SHA256),
			Bytes:           payload.Artifact.Bytes,
			SourceURL:       payload.Artifact.SourceURL,
			IdentityQuery:   payload.Artifact.IdentityQuery,
			Identity:        payload.Artifact.Identity,
			CanonicalSHA256: strings.ToLower(payload.Artifact.CanonicalSHA256),
			RetainedSHA256:  strings.ToLower(payload.RetainedSHA256),
			RetainedBytes:   payload.RetainedBytes,
			Form:            payload.Form,
			Version:         payload.Version,
			RetrievedAt:     opts.RetrievedAt.Format(time.RFC3339),
			License:         opts.License,
			Terms:           opts.Terms,
		}
		key := sourceRetainRecordKey(record)
		if existing, exists := seen[key]; exists {
			if existing.Bytes != payload.Artifact.Bytes || !strings.EqualFold(existing.CanonicalSHA256, record.CanonicalSHA256) {
				return fmt.Errorf("existing retained artifact manifest conflicts with locked bytes")
			}
			continue
		}
		manifest.Artifacts = append(manifest.Artifacts, record)
		seen[key] = record
	}
	sort.Slice(manifest.Artifacts, func(left, right int) bool {
		leftKey := sourceRetainRecordKey(manifest.Artifacts[left])
		rightKey := sourceRetainRecordKey(manifest.Artifacts[right])
		return leftKey < rightKey
	})
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode retained artifact manifest: %w", err)
	}
	return sourceRetainReplaceRegularFile(manifestPath, append(raw, '\n'))
}

func sourceRetainValidateManifestRecord(record sourceImportRetainedArtifactRecord) error {
	if err := validateSourceRetainArtifact(sourceImportArtifact{
		SourceURL:       record.SourceURL,
		SHA256:          record.SHA256,
		Bytes:           record.Bytes,
		IdentityQuery:   record.IdentityQuery,
		Identity:        record.Identity,
		CanonicalSHA256: record.CanonicalSHA256,
	}); err != nil {
		return fmt.Errorf("identity: %w", err)
	}
	retainedSHA256 := sourceRetainedArtifactRecordRawSHA256(record)
	if len(retainedSHA256) != sha256.Size*2 {
		return fmt.Errorf("retained byte SHA-256 is invalid")
	}
	if _, err := hex.DecodeString(retainedSHA256); err != nil {
		return fmt.Errorf("retained byte SHA-256 is invalid: %w", err)
	}
	if sourceRetainedArtifactRecordRawBytes(record) <= 0 {
		return fmt.Errorf("retained byte count is invalid")
	}
	if record.Form != "" || record.Version != "" {
		if record.Form == "" || record.Version == "" || record.Form != strings.TrimSpace(record.Form) || record.Version != strings.TrimSpace(record.Version) || len(record.Form) > 64 || len(record.Version) > 128 || strings.ContainsAny(record.Form+record.Version, "\r\n") {
			return fmt.Errorf("retained source form/version is invalid")
		}
	}
	if _, err := time.Parse(time.RFC3339, record.RetrievedAt); err != nil {
		return fmt.Errorf("retrieval timestamp is invalid: %w", err)
	}
	if err := validateSourceRetainProvenance("license", record.License); err != nil {
		return err
	}
	return validateSourceRetainProvenance("terms", record.Terms)
}

func sourceRetainRecordKey(record sourceImportRetainedArtifactRecord) string {
	return sourceRetainedArtifactRecordLockKey(record) + "\x00" + strings.ToLower(sourceRetainedArtifactRecordRawSHA256(record))
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
