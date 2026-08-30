package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

const (
	sourceOperationMappingCohortSchemaVersion = 1
	batch1SourceOperationMappingCohort        = "batch1-source-operation-mapping-r1"
	batch1SourceOperationCount                = 4341
)

var batch1SourceOperationMappingConnectors = []string{
	"asana",
	"bitbucket",
	"circleci",
	"dockerhub",
	"gitlab",
	"jira",
	"notion",
	"sentry",
	"stripe",
	"vercel",
}

// sourceOperationMappingCohortManifest freezes the Batch R1 source-lock
// denominator while leaving connector-local lane matrices in their owned
// directories. The cohort intentionally records only matrix input paths, not
// their rows or cells, so it cannot become a competing provider-fact source.
type sourceOperationMappingCohortManifest struct {
	SchemaVersion          int                                      `json:"schema_version"`
	Cohort                 string                                   `json:"cohort"`
	ExpectedConnectors     []string                                 `json:"expected_connectors"`
	SourceOperationCount   int                                      `json:"source_operation_count"`
	SourceOperationsSHA256 string                                   `json:"source_operations_sha256"`
	SourceLocks            []sourceOperationMappingCohortSourceLock `json:"source_locks"`
}

type sourceOperationMappingCohortSourceLock struct {
	Connector            string `json:"connector"`
	Path                 string `json:"path"`
	SHA256               string `json:"sha256"`
	SourceOperationCount int    `json:"source_operation_count"`
	SourceIDsSHA256      string `json:"source_ids_sha256"`
	MatrixPath           string `json:"matrix_path"`
}

type sourceOperationMappingCohortReport struct {
	Manifest          string    `json:"manifest"`
	ConnectorsChecked int       `json:"connectors_checked"`
	SourceOperations  int       `json:"source_operations"`
	Findings          []Finding `json:"findings"`
}

func runSourceOperationMappingCohort(args []string, stdout, stderr io.Writer) int {
	if len(args) == 2 && (args[1] == "-h" || args[1] == "--help") {
		logln(stdout, sourceOperationMappingCohortUsage())
		return 0
	}
	manifestPath, err := parseSourceOperationMappingCohortOptions(args)
	if err != nil {
		logf(stderr, "connectorgen source-operation-mapping-cohort: %v\n", err)
		return 2
	}
	root, err := repoRoot()
	if err != nil {
		logf(stderr, "connectorgen source-operation-mapping-cohort: resolve repository root: %v\n", err)
		return 1
	}
	report, err := sourceOperationMappingCohortPathCheck(root, manifestPath)
	if err != nil {
		logf(stderr, "connectorgen source-operation-mapping-cohort: %v\n", err)
		return 1
	}
	for _, finding := range report.Findings {
		logf(stdout, "%s: %s: %s\n", finding.Connector, finding.File, finding.Message)
	}
	logf(stdout, "connectorgen source-operation-mapping-cohort: %d connector(s), %d source operation(s), %d finding(s)\n", report.ConnectorsChecked, report.SourceOperations, len(report.Findings))
	if len(report.Findings) > 0 {
		return 1
	}
	return 0
}

func sourceOperationMappingCohortUsage() string {
	return "usage: connectorgen source-operation-mapping-cohort <manifest> --check"
}

func parseSourceOperationMappingCohortOptions(args []string) (string, error) {
	return parseSourceOperationMappingOptions(args)
}

func sourceOperationMappingCohortPathCheck(root, manifestPath string) (sourceOperationMappingCohortReport, error) {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return sourceOperationMappingCohortReport{}, fmt.Errorf("read cohort manifest %s: %w", manifestPath, err)
	}
	if err := engine.ValidateSourceOperationMappingCohort(raw); err != nil {
		return sourceOperationMappingCohortReport{}, fmt.Errorf("validate cohort manifest shape: %w", err)
	}
	var manifest sourceOperationMappingCohortManifest
	if err := decodeSourceStrictJSON(raw, &manifest); err != nil {
		return sourceOperationMappingCohortReport{}, fmt.Errorf("decode cohort manifest: %w", err)
	}
	if manifest.SchemaVersion != sourceOperationMappingCohortSchemaVersion {
		return sourceOperationMappingCohortReport{}, fmt.Errorf("unsupported cohort schema version %d", manifest.SchemaVersion)
	}

	repositoryRoot, err := sourceOperationMappingCohortRoot(root)
	if err != nil {
		return sourceOperationMappingCohortReport{}, err
	}
	report := sourceOperationMappingCohortReport{Manifest: manifestPath}
	add := func(connector, message string) {
		report.Findings = append(report.Findings, Finding{Connector: connector, File: manifestPath, Message: message})
	}
	if manifest.Cohort != batch1SourceOperationMappingCohort {
		add("", fmt.Sprintf("cohort %q does not equal fixed Batch R1 cohort %q", manifest.Cohort, batch1SourceOperationMappingCohort))
	}
	if !sourceOperationMappingExactStrings(manifest.ExpectedConnectors, batch1SourceOperationMappingConnectors) {
		add("", "expected_connectors does not retain exact Batch R1 connector membership")
	}
	if manifest.SourceOperationCount != batch1SourceOperationCount {
		add("", fmt.Sprintf("source operation count %d does not equal fixed Batch R1 count %d", manifest.SourceOperationCount, batch1SourceOperationCount))
	}
	if len(manifest.SourceLocks) != len(batch1SourceOperationMappingConnectors) {
		add("", fmt.Sprintf("source_locks has %d entries, want exact Batch R1 connector membership of %d", len(manifest.SourceLocks), len(batch1SourceOperationMappingConnectors)))
	}

	seenConnectors := make(map[string]struct{}, len(manifest.SourceLocks))
	allSourceIDs := make([]string, 0, batch1SourceOperationCount)
	for _, lock := range manifest.SourceLocks {
		if _, duplicate := seenConnectors[lock.Connector]; duplicate {
			add(lock.Connector, fmt.Sprintf("duplicate Batch R1 source lock connector %q", lock.Connector))
			continue
		}
		seenConnectors[lock.Connector] = struct{}{}
		if !sourceOperationMappingBatch1Connector(lock.Connector) {
			add(lock.Connector, fmt.Sprintf("connector %q is not in exact Batch R1 connector membership", lock.Connector))
			continue
		}
		lockPath, err := sourceOperationMappingCohortOwnedPath(repositoryRoot, lock.Connector, lock.Path, "-operation-source-lock.json")
		if err != nil {
			add(lock.Connector, fmt.Sprintf("source lock path %q: %v", lock.Path, err))
			continue
		}
		if err := sourceOperationMappingCohortMatrixPath(lock.Connector, lock.MatrixPath); err != nil {
			add(lock.Connector, fmt.Sprintf("connector-local matrix input %q: %v", lock.MatrixPath, err))
		}
		lockRaw, err := os.ReadFile(lockPath)
		if err != nil {
			return sourceOperationMappingCohortReport{}, fmt.Errorf("read source lock %s: %w", lock.Path, err)
		}
		if digest := sourceOperationMappingSHA256(lockRaw); digest != lock.SHA256 {
			add(lock.Connector, fmt.Sprintf("source lock SHA-256 %q does not match tracked %q", digest, lock.SHA256))
		}
		parsed, err := parseDeclarationAdmissionSourceLock(lockRaw, lock.Connector)
		if err != nil {
			return sourceOperationMappingCohortReport{}, fmt.Errorf("parse source lock %s: %w", lock.Path, err)
		}
		ids := make([]string, 0, len(parsed.Operations))
		for sourceOperationID := range parsed.Operations {
			ids = append(ids, sourceOperationID)
			allSourceIDs = append(allSourceIDs, lock.Connector+"\t"+sourceOperationID)
		}
		sort.Strings(ids)
		if len(ids) != lock.SourceOperationCount {
			add(lock.Connector, fmt.Sprintf("source operation count %d does not match tracked %d", len(ids), lock.SourceOperationCount))
		}
		if digest := sourceOperationMappingStringsSHA256(ids); digest != lock.SourceIDsSHA256 {
			add(lock.Connector, fmt.Sprintf("source ID SHA-256 %q does not match tracked %q", digest, lock.SourceIDsSHA256))
		}
		report.ConnectorsChecked++
	}
	for _, connector := range batch1SourceOperationMappingConnectors {
		if _, found := seenConnectors[connector]; !found {
			add(connector, "missing exact Batch R1 connector membership entry")
		}
	}
	sort.Strings(allSourceIDs)
	report.SourceOperations = len(allSourceIDs)
	if report.SourceOperations != manifest.SourceOperationCount {
		add("", fmt.Sprintf("source operation count %d does not match tracked cohort count %d", report.SourceOperations, manifest.SourceOperationCount))
	}
	if digest := sourceOperationMappingStringsSHA256(allSourceIDs); digest != manifest.SourceOperationsSHA256 {
		add("", fmt.Sprintf("source ID SHA-256 %q does not match tracked cohort %q", digest, manifest.SourceOperationsSHA256))
	}
	sort.Slice(report.Findings, func(i, j int) bool {
		left, right := report.Findings[i], report.Findings[j]
		if left.Connector != right.Connector {
			return left.Connector < right.Connector
		}
		if left.File != right.File {
			return left.File < right.File
		}
		return left.Message < right.Message
	})
	return report, nil
}

func sourceOperationMappingCohortRoot(root string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	return resolvedRoot, nil
}

func sourceOperationMappingCohortOwnedPath(root, connector, raw, suffix string) (string, error) {
	expected := filepath.ToSlash(filepath.Join("internal", "connectors", "defs", connector, "sources", connector+suffix))
	if raw != expected || filepath.IsAbs(raw) || strings.Contains(raw, "\\") || filepath.ToSlash(filepath.Clean(filepath.FromSlash(raw))) != raw {
		return "", fmt.Errorf("must be canonical connector-owned path %q", expected)
	}
	resolved, err := filepath.EvalSymlinks(filepath.Join(root, filepath.FromSlash(raw)))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("resolves outside repository root")
	}
	if filepath.ToSlash(relative) != expected {
		return "", fmt.Errorf("resolves outside canonical connector-owned path %q", expected)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("must resolve to a regular file")
	}
	return resolved, nil
}

func sourceOperationMappingCohortMatrixPath(connector, raw string) error {
	expected := filepath.ToSlash(filepath.Join("internal", "connectors", "defs", connector, "sources", connector+"-source-lane-matrix.json"))
	if raw != expected || filepath.IsAbs(raw) || strings.Contains(raw, "\\") || filepath.ToSlash(filepath.Clean(filepath.FromSlash(raw))) != raw {
		return fmt.Errorf("must be canonical connector-local input path %q", expected)
	}
	return nil
}

func sourceOperationMappingBatch1Connector(connector string) bool {
	for _, expected := range batch1SourceOperationMappingConnectors {
		if connector == expected {
			return true
		}
	}
	return false
}

func sourceOperationMappingExactStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sourceOperationMappingStringsSHA256(values []string) string {
	digest := sha256.New()
	for _, value := range values {
		_, _ = io.WriteString(digest, value)
		_, _ = io.WriteString(digest, "\n")
	}
	return hex.EncodeToString(digest.Sum(nil))
}

func sourceOperationMappingSHA256(raw []byte) string {
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:])
}
