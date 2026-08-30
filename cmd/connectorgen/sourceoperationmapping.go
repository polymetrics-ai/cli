package main

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

const sourceOperationMappingSchemaVersion = 1

// sourceOperationMappingManifest is an authoring-only, source-lock-bound
// denominator for multi-lane connector planning. It intentionally has no
// runtime fields: a lane mapping records source facts and admission state, not
// an executable route, credential, transport, or certification result.
type sourceOperationMappingManifest struct {
	SchemaVersion int                                  `json:"schema_version"`
	SourceLocks   []sourceOperationMappingSourceLock   `json:"source_locks"`
	Operations    []sourceOperationMappingOperation    `json:"operations"`
	Artifacts     []sourceOperationMappingArtifactLink `json:"artifacts"`
}

type sourceOperationMappingSourceLock struct {
	Connector string `json:"connector"`
	Path      string `json:"path"`
}

type sourceOperationMappingOperation struct {
	Connector            string                               `json:"connector"`
	SourceOperationID    string                               `json:"source_operation_id"`
	CanonicalOperationID string                               `json:"canonical_operation_id"`
	Citation             sourceOperationMappingSourceCitation `json:"citation"`
	Facts                sourceOperationMappingFacts          `json:"facts"`
	Cells                []sourceOperationMappingLaneCell     `json:"cells"`
}

type sourceOperationMappingSourceCitation struct {
	SourceURL string `json:"source_url"`
	Location  string `json:"location"`
}

type sourceOperationMappingFacts struct {
	Pagination    sourceOperationMappingEnumeratedFact     `json:"pagination"`
	RecordShape   sourceOperationMappingEnumeratedFact     `json:"record_shape"`
	Scope         sourceOperationMappingValuesFact         `json:"scope"`
	PathVariables sourceOperationMappingValuesFact         `json:"path_variables"`
	Media         sourceOperationMappingMediaFact          `json:"media"`
	EventCursor   sourceOperationMappingEnumeratedFact     `json:"event_cursor"`
	Mutation      sourceOperationMappingEnumeratedFact     `json:"mutation"`
	Applicability sourceOperationMappingApplicabilityFacts `json:"applicability"`
}

type sourceOperationMappingEnumeratedFact struct {
	Kind     string                               `json:"kind"`
	Citation sourceOperationMappingSourceCitation `json:"citation"`
}

type sourceOperationMappingValuesFact struct {
	Values   []string                             `json:"values"`
	Citation sourceOperationMappingSourceCitation `json:"citation"`
}

type sourceOperationMappingMediaFact struct {
	Request  []string                             `json:"request"`
	Response []string                             `json:"response"`
	Citation sourceOperationMappingSourceCitation `json:"citation"`
}

// sourceOperationMappingApplicabilityFacts is a strict cited sidecar for
// lane candidate facts. Each fact is bound to the exact locked operation node,
// so its source location cannot be borrowed from another operation in the same
// provider document. It is mapping evidence only, never runtime admission.
type sourceOperationMappingApplicabilityFacts struct {
	ETL            sourceOperationMappingEnumeratedFact `json:"etl"`
	BinaryDownload sourceOperationMappingEnumeratedFact `json:"binary_download"`
	BinaryUpload   sourceOperationMappingEnumeratedFact `json:"binary_upload"`
	SyncTransport  sourceOperationMappingEnumeratedFact `json:"sync_transport"`
}

type sourceOperationMappingLaneCell struct {
	Lane   string                            `json:"lane"`
	State  string                            `json:"state"`
	Reason *sourceOperationMappingCellReason `json:"reason,omitempty"`
}

type sourceOperationMappingCellReason struct {
	Kind     string                                `json:"kind"`
	ID       string                                `json:"id"`
	Citation *sourceOperationMappingSourceCitation `json:"citation,omitempty"`
}

type sourceOperationMappingArtifactLink struct {
	Path  string                                   `json:"path"`
	Cells []sourceOperationMappingArtifactCellLink `json:"cells"`
}

// sourceOperationMappingArtifactCellLink deliberately has no connector or
// source-row payload. A link can resolve one already declared source ID/lane
// cell but cannot become a second place that creates source operations.
type sourceOperationMappingArtifactCellLink struct {
	SourceOperationID string `json:"source_operation_id"`
	Lane              string `json:"lane"`
}

type sourceOperationMappingReport struct {
	Manifest            string    `json:"manifest"`
	ConnectorsChecked   int       `json:"connectors_checked"`
	SourceOperations    int       `json:"source_operations"`
	CanonicalOperations int       `json:"canonical_operations"`
	Cells               int       `json:"cells"`
	Findings            []Finding `json:"findings"`
}

type sourceOperationMappingLockedOperation struct {
	Connector string
	Operation declarationAdmissionReviewedOperation
}

type sourceOperationMappingResolvedOperation struct {
	Manifest sourceOperationMappingOperation
	Locked   sourceOperationMappingLockedOperation
}

// runSourceOperationMapping validates an explicit multi-lane manifest. The
// command is check-only by design so it cannot mutate a source lock, connector
// definition, generated artifact, or runtime surface.
func runSourceOperationMapping(args []string, stdout, stderr io.Writer) int {
	if len(args) == 2 && (args[1] == "-h" || args[1] == "--help") {
		logln(stdout, sourceOperationMappingUsage())
		return 0
	}
	manifestPath, err := parseSourceOperationMappingOptions(args)
	if err != nil {
		logf(stderr, "connectorgen source-operation-mapping: %v\n", err)
		return 2
	}
	report, err := sourceOperationMappingPathCheck(manifestPath)
	if err != nil {
		logf(stderr, "connectorgen source-operation-mapping: %v\n", err)
		return 1
	}
	for _, finding := range report.Findings {
		logf(stdout, "%s: %s: %s\n", finding.Connector, finding.File, finding.Message)
	}
	logf(stdout, "connectorgen source-operation-mapping: %d connector(s), %d source operation(s), %d canonical operation(s), %d cell(s), %d finding(s)\n",
		report.ConnectorsChecked, report.SourceOperations, report.CanonicalOperations, report.Cells, len(report.Findings))
	if len(report.Findings) > 0 {
		return 1
	}
	return 0
}

func sourceOperationMappingUsage() string {
	return "usage: connectorgen source-operation-mapping <manifest> --check"
}

func parseSourceOperationMappingOptions(args []string) (string, error) {
	var manifestPath string
	check := false
	for _, argument := range args[1:] {
		switch argument {
		case "--check":
			if check {
				return "", fmt.Errorf("--check may be specified only once")
			}
			check = true
		default:
			if strings.HasPrefix(argument, "-") {
				return "", fmt.Errorf("unknown flag %q", argument)
			}
			if manifestPath != "" {
				return "", fmt.Errorf("unexpected extra argument %q", argument)
			}
			manifestPath = argument
		}
	}
	if manifestPath == "" {
		return "", fmt.Errorf("manifest path is required")
	}
	if !check {
		return "", fmt.Errorf("--check is required; source-operation-mapping is validation only")
	}
	return manifestPath, nil
}

func sourceOperationMappingPathCheck(manifestPath string) (sourceOperationMappingReport, error) {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return sourceOperationMappingReport{}, fmt.Errorf("read manifest %s: %w", manifestPath, err)
	}
	if err := engine.ValidateSourceOperationMappingManifest(raw); err != nil {
		return sourceOperationMappingReport{}, fmt.Errorf("validate manifest shape: %w", err)
	}
	var manifest sourceOperationMappingManifest
	if err := decodeSourceStrictJSON(raw, &manifest); err != nil {
		return sourceOperationMappingReport{}, fmt.Errorf("decode manifest: %w", err)
	}
	if manifest.SchemaVersion != sourceOperationMappingSchemaVersion {
		return sourceOperationMappingReport{}, fmt.Errorf("unsupported schema version %d", manifest.SchemaVersion)
	}

	root, err := sourceOperationMappingManifestRoot(manifestPath)
	if err != nil {
		return sourceOperationMappingReport{}, err
	}
	report := sourceOperationMappingReport{Manifest: manifestPath}
	add := func(connector, message string) {
		report.Findings = append(report.Findings, Finding{Connector: connector, File: manifestPath, Message: message})
	}

	lockedOperations := make(map[string]sourceOperationMappingLockedOperation)
	loadedLocks := make(map[string]struct{}, len(manifest.SourceLocks))
	connectors := make(map[string]struct{}, len(manifest.SourceLocks))
	for _, sourceLock := range manifest.SourceLocks {
		if _, duplicate := loadedLocks[sourceLock.Path]; duplicate {
			add(sourceLock.Connector, fmt.Sprintf("duplicate source lock path %q", sourceLock.Path))
			continue
		}
		loadedLocks[sourceLock.Path] = struct{}{}
		lockPath, err := sourceOperationMappingOwnedSourceLockPath(root, sourceLock)
		if err != nil {
			return sourceOperationMappingReport{}, fmt.Errorf("source lock %q: %w", sourceLock.Path, err)
		}
		lockRaw, err := os.ReadFile(lockPath)
		if err != nil {
			return sourceOperationMappingReport{}, fmt.Errorf("read source lock %s: %w", sourceLock.Path, err)
		}
		lock, err := parseDeclarationAdmissionSourceLock(lockRaw, sourceLock.Connector)
		if err != nil {
			return sourceOperationMappingReport{}, fmt.Errorf("parse source lock %s: %w", sourceLock.Path, err)
		}
		connectors[sourceLock.Connector] = struct{}{}
		for sourceOperationID, operation := range lock.Operations {
			if previous, duplicate := lockedOperations[sourceOperationID]; duplicate {
				add(sourceLock.Connector, fmt.Sprintf("duplicate source operation ID %q in source locks for connectors %q and %q", sourceOperationID, previous.Connector, sourceLock.Connector))
				continue
			}
			lockedOperations[sourceOperationID] = sourceOperationMappingLockedOperation{Connector: sourceLock.Connector, Operation: operation}
		}
	}
	report.ConnectorsChecked = len(connectors)
	report.SourceOperations = len(lockedOperations)

	mappedSourceOperations := make(map[string]struct{}, len(manifest.Operations))
	declaredCells := make(map[string]struct{})
	seenSourceOperationIDs := make(map[string]struct{}, len(manifest.Operations))
	resolvedOperations := make(map[string]sourceOperationMappingResolvedOperation, len(manifest.Operations))
	for _, operation := range manifest.Operations {
		if !sourceOperationMappingText(operation.SourceOperationID, 1024) {
			add(operation.Connector, "source operation ID must be nonempty and canonical")
			continue
		}
		if !sourceOperationMappingText(operation.CanonicalOperationID, 1024) {
			add(operation.Connector, fmt.Sprintf("source operation %q canonical operation ID must be nonempty and canonical", operation.SourceOperationID))
		}
		if _, duplicate := seenSourceOperationIDs[operation.SourceOperationID]; duplicate {
			add(operation.Connector, fmt.Sprintf("duplicate source operation ID %q", operation.SourceOperationID))
		} else {
			seenSourceOperationIDs[operation.SourceOperationID] = struct{}{}
		}

		locked, found := lockedOperations[operation.SourceOperationID]
		if !found {
			add(operation.Connector, fmt.Sprintf("source operation %q is not present in the declared source locks", operation.SourceOperationID))
			continue
		}
		mappedSourceOperations[operation.SourceOperationID] = struct{}{}
		if _, duplicate := resolvedOperations[operation.SourceOperationID]; !duplicate {
			resolvedOperations[operation.SourceOperationID] = sourceOperationMappingResolvedOperation{Manifest: operation, Locked: locked}
		}
		if operation.Connector != locked.Connector {
			add(operation.Connector, fmt.Sprintf("source operation %q belongs to source-lock connector %q, not %q", operation.SourceOperationID, locked.Connector, operation.Connector))
		}
		if err := sourceOperationMappingValidateExactCitation(operation.Citation, locked.Operation); err != nil {
			add(operation.Connector, fmt.Sprintf("source operation %q citation: %v", operation.SourceOperationID, err))
		}
		for _, message := range sourceOperationMappingFactFindings(operation, locked.Operation) {
			add(operation.Connector, message)
		}

		seenLanes := make(map[string]struct{}, len(operation.Cells))
		cellsByLane := make(map[string]sourceOperationMappingLaneCell, len(operation.Cells))
		for _, cell := range operation.Cells {
			report.Cells++
			if _, duplicate := seenLanes[cell.Lane]; duplicate {
				add(operation.Connector, fmt.Sprintf("source operation %q has duplicate %s cell", operation.SourceOperationID, cell.Lane))
				continue
			}
			seenLanes[cell.Lane] = struct{}{}
			cellsByLane[cell.Lane] = cell
			for _, message := range sourceOperationMappingCellFindings(operation, cell, locked.Operation) {
				add(operation.Connector, message)
			}
			declaredCells[sourceOperationMappingCellKey(operation.SourceOperationID, cell.Lane)] = struct{}{}
		}
		for _, message := range sourceOperationMappingApplicabilityFindings(operation, locked.Operation, cellsByLane) {
			add(operation.Connector, message)
		}
		for _, message := range sourceOperationMappingMutationFindings(operation, locked.Operation, cellsByLane) {
			add(operation.Connector, message)
		}
	}
	report.CanonicalOperations = sourceOperationMappingCanonicalOperationCount(resolvedOperations, add)

	lockedIDs := make([]string, 0, len(lockedOperations))
	for sourceOperationID := range lockedOperations {
		lockedIDs = append(lockedIDs, sourceOperationID)
	}
	sort.Strings(lockedIDs)
	for _, sourceOperationID := range lockedIDs {
		if _, mapped := mappedSourceOperations[sourceOperationID]; !mapped {
			locked := lockedOperations[sourceOperationID]
			add(locked.Connector, fmt.Sprintf("source operation %q is absent from the mapping manifest", sourceOperationID))
		}
	}

	for _, artifact := range manifest.Artifacts {
		if !sourceOperationMappingArtifactPath(artifact.Path) {
			add("", fmt.Sprintf("artifact path %q must be one canonical relative path", artifact.Path))
		}
		for _, link := range artifact.Cells {
			if _, found := declaredCells[sourceOperationMappingCellKey(link.SourceOperationID, link.Lane)]; !found {
				add("", fmt.Sprintf("artifact %q references nonexistent mapping cell %s/%s", artifact.Path, link.SourceOperationID, link.Lane))
			}
		}
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

// sourceOperationMappingCanonicalOperationCount keeps the source denominator
// and the canonical-operation denominator distinct. Supplemental source locks
// may describe the same provider route with a different source ID and cited
// location; they remain independently accounted source rows only when they
// explicitly point to a self-representing canonical source operation.
func sourceOperationMappingCanonicalOperationCount(resolved map[string]sourceOperationMappingResolvedOperation, add func(string, string)) int {
	sourceOperationIDs := make([]string, 0, len(resolved))
	for sourceOperationID := range resolved {
		sourceOperationIDs = append(sourceOperationIDs, sourceOperationID)
	}
	sort.Strings(sourceOperationIDs)
	canonicalOperations := make(map[string]struct{}, len(resolved))
	for _, sourceOperationID := range sourceOperationIDs {
		resolvedOperation := resolved[sourceOperationID]
		canonicalOperationID := resolvedOperation.Manifest.CanonicalOperationID
		canonical, found := resolved[canonicalOperationID]
		if !found {
			add(resolvedOperation.Manifest.Connector, fmt.Sprintf("source operation %q canonical operation ID %q does not reference a source-lock-backed manifest operation", sourceOperationID, canonicalOperationID))
			continue
		}
		if canonical.Manifest.CanonicalOperationID != canonicalOperationID {
			add(resolvedOperation.Manifest.Connector, fmt.Sprintf("source operation %q canonical operation ID %q must reference a self-canonical source operation", sourceOperationID, canonicalOperationID))
			continue
		}
		if resolvedOperation.Manifest.Connector != canonical.Manifest.Connector {
			add(resolvedOperation.Manifest.Connector, fmt.Sprintf("source operation %q cannot deduplicate across connector %q", sourceOperationID, canonical.Manifest.Connector))
			continue
		}
		if message := sourceOperationMappingCanonicalIdentityMismatch(resolvedOperation.Locked.Operation, canonical.Locked.Operation); message != "" {
			add(resolvedOperation.Manifest.Connector, fmt.Sprintf("source operation %q canonical operation ID %q does not preserve source-lock operation identity: %s", sourceOperationID, canonicalOperationID, message))
			continue
		}
		canonicalOperations[canonicalOperationID] = struct{}{}
	}
	return len(canonicalOperations)
}

func sourceOperationMappingCanonicalIdentityMismatch(source, canonical declarationAdmissionReviewedOperation) string {
	if source.Protocol != canonical.Protocol {
		return fmt.Sprintf("protocol %q does not equal %q", source.Protocol, canonical.Protocol)
	}
	if !strings.EqualFold(source.Method, canonical.Method) {
		return fmt.Sprintf("method %q does not equal %q", source.Method, canonical.Method)
	}
	if source.Path != canonical.Path {
		return fmt.Sprintf("path %q does not equal %q", source.Path, canonical.Path)
	}
	if source.Protocol == "graphql" && source.ProviderOperationID != canonical.ProviderOperationID {
		return fmt.Sprintf("GraphQL provider operation identity %q does not equal %q", source.ProviderOperationID, canonical.ProviderOperationID)
	}
	return ""
}

func sourceOperationMappingManifestRoot(manifestPath string) (string, error) {
	absManifest, err := filepath.Abs(manifestPath)
	if err != nil {
		return "", fmt.Errorf("resolve manifest path: %w", err)
	}
	root, err := filepath.EvalSymlinks(filepath.Dir(absManifest))
	if err != nil {
		return "", fmt.Errorf("resolve manifest directory: %w", err)
	}
	return root, nil
}

func sourceOperationMappingOwnedSourceLockPath(root string, sourceLock sourceOperationMappingSourceLock) (string, error) {
	if err := validateSourceImportConnector(sourceLock.Connector); err != nil {
		return "", err
	}
	raw := sourceLock.Path
	if raw == "" || filepath.IsAbs(raw) || strings.Contains(raw, "\\") || filepath.ToSlash(filepath.Clean(filepath.FromSlash(raw))) != raw {
		return "", fmt.Errorf("path must be one canonical relative path")
	}
	prefix := sourceLock.Connector + "/sources/"
	if !strings.HasPrefix(raw, prefix) || !strings.HasSuffix(raw, "-operation-source-lock.json") {
		return "", fmt.Errorf("path must be owned beneath %s", prefix)
	}
	path, err := filepath.EvalSymlinks(filepath.Join(root, filepath.FromSlash(raw)))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path resolves outside the manifest directory")
	}
	if !strings.HasPrefix(filepath.ToSlash(relative), prefix) {
		return "", fmt.Errorf("path resolves outside connector-owned %s", prefix)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("path must resolve to one regular file")
	}
	return path, nil
}

func sourceOperationMappingValidateExactCitation(citation sourceOperationMappingSourceCitation, operation declarationAdmissionReviewedOperation) error {
	if err := sourceOperationMappingValidateCitation(citation, operation.SourceURL); err != nil {
		return err
	}
	if citation.Location != operation.Location {
		return fmt.Errorf("location %q is not the exact source-lock location %q", citation.Location, operation.Location)
	}
	return nil
}

func sourceOperationMappingFactFindings(operation sourceOperationMappingOperation, locked declarationAdmissionReviewedOperation) []string {
	findings := []string{}
	addCitation := func(name string, citation sourceOperationMappingSourceCitation) {
		if err := sourceOperationMappingValidateExactCitation(citation, locked); err != nil {
			findings = append(findings, fmt.Sprintf("source operation %q %s fact citation: %v", operation.SourceOperationID, name, err))
		}
	}
	addCitation("pagination", operation.Facts.Pagination.Citation)
	addCitation("record_shape", operation.Facts.RecordShape.Citation)
	addCitation("scope", operation.Facts.Scope.Citation)
	addCitation("path_variables", operation.Facts.PathVariables.Citation)
	addCitation("media", operation.Facts.Media.Citation)
	addCitation("event_cursor", operation.Facts.EventCursor.Citation)
	addCitation("mutation", operation.Facts.Mutation.Citation)
	addCitation("etl applicability", operation.Facts.Applicability.ETL.Citation)
	addCitation("binary_download applicability", operation.Facts.Applicability.BinaryDownload.Citation)
	addCitation("binary_upload applicability", operation.Facts.Applicability.BinaryUpload.Citation)
	addCitation("sync_transport applicability", operation.Facts.Applicability.SyncTransport.Citation)
	for _, group := range []struct {
		name   string
		values []string
	}{
		{name: "scope", values: operation.Facts.Scope.Values},
		{name: "path_variables", values: operation.Facts.PathVariables.Values},
		{name: "media request", values: operation.Facts.Media.Request},
		{name: "media response", values: operation.Facts.Media.Response},
	} {
		if message := sourceOperationMappingValuesFinding(operation.SourceOperationID, group.name, group.values); message != "" {
			findings = append(findings, message)
		}
	}
	return findings
}

func sourceOperationMappingApplicabilityFindings(operation sourceOperationMappingOperation, locked declarationAdmissionReviewedOperation, cells map[string]sourceOperationMappingLaneCell) []string {
	findings := []string{}
	add := func(format string, args ...any) {
		findings = append(findings, fmt.Sprintf(format, args...))
	}
	records := operation.Facts.RecordShape.Kind == "collection"
	pageable := operation.Facts.Pagination.Kind != "none"
	evented := operation.Facts.EventCursor.Kind != "none"

	check := func(lane string, fact sourceOperationMappingEnumeratedFact, sourceCandidate bool, contradiction string) {
		if sourceCandidate && fact.Kind != "applicable" {
			add("source operation %q %s applicability contradicts %s", operation.SourceOperationID, lane, contradiction)
		}
		if fact.Kind == "applicable" && !sourceCandidate {
			add("source operation %q %s applicability is not supported by %s", operation.SourceOperationID, lane, contradiction)
		}
		cell, found := cells[lane]
		if fact.Kind == "applicable" {
			if !found {
				if lane == "etl" && pageable {
					add("pageable source operation %q has no explicit etl disposition", operation.SourceOperationID)
					return
				}
				add("source operation %q has no explicit %s disposition", operation.SourceOperationID, lane)
				return
			}
			if cell.State == "not_applicable" {
				add("source operation %q %s cell contradicts source-applicable evidence", operation.SourceOperationID, lane)
			}
			return
		}
		if found && cell.State != "not_applicable" {
			add("source operation %q %s cell contradicts source-not-applicable evidence", operation.SourceOperationID, lane)
		}
	}

	check("etl", operation.Facts.Applicability.ETL, records && pageable, "collection pagination evidence")
	check("binary_download", operation.Facts.Applicability.BinaryDownload, sourceOperationMappingBinaryMedia(operation.Facts.Media.Response), "response media evidence")
	check("binary_upload", operation.Facts.Applicability.BinaryUpload, sourceOperationMappingBinaryMedia(operation.Facts.Media.Request), "request media evidence")
	check("sync_transport", operation.Facts.Applicability.SyncTransport, records && evented, "event/cursor evidence")
	return findings
}

func sourceOperationMappingMutationFindings(operation sourceOperationMappingOperation, locked declarationAdmissionReviewedOperation, cells map[string]sourceOperationMappingLaneCell) []string {
	findings := []string{}
	add := func(format string, args ...any) {
		findings = append(findings, fmt.Sprintf(format, args...))
	}
	mustBeMutation := false
	mustNotBeMutation := false
	switch locked.Protocol {
	case "graphql":
		if strings.HasPrefix(locked.ProviderOperationID, "Mutation.") {
			mustBeMutation = true
		} else {
			mustNotBeMutation = true
		}
	case "rest":
		switch strings.ToUpper(locked.Method) {
		case "PUT", "PATCH", "DELETE":
			mustBeMutation = true
		case "GET", "HEAD":
			mustNotBeMutation = true
		}
	}
	if mustBeMutation && operation.Facts.Mutation.Kind != "mutation" {
		add("source operation %q mutation fact contradicts locked %s identity", operation.SourceOperationID, sourceOperationMappingMutationIdentity(locked))
	}
	if mustNotBeMutation && operation.Facts.Mutation.Kind != "not_mutation" {
		add("source operation %q mutation fact contradicts locked %s identity", operation.SourceOperationID, sourceOperationMappingMutationIdentity(locked))
	}
	if operation.Facts.Mutation.Kind != "mutation" {
		return findings
	}
	for _, lane := range []string{"direct_write", "reverse_etl"} {
		cell, found := cells[lane]
		if !found {
			add("mutation source operation %q has no explicit %s disposition", operation.SourceOperationID, lane)
			continue
		}
		if cell.State == "not_applicable" {
			add("mutation source operation %q %s disposition contradicts source mutation evidence", operation.SourceOperationID, lane)
		}
	}
	return findings
}

func sourceOperationMappingMutationIdentity(operation declarationAdmissionReviewedOperation) string {
	if operation.Protocol == "graphql" {
		return "GraphQL root " + operation.ProviderOperationID
	}
	return "REST method " + strings.ToUpper(operation.Method)
}

func sourceOperationMappingBinaryMedia(media []string) bool {
	for _, value := range media {
		mediaType, _, err := mime.ParseMediaType(value)
		if err != nil || !sourceOperationMappingConcreteMediaType(mediaType) || sourceOperationMappingJSONMediaType(mediaType) {
			continue
		}
		// The manifest's media fact has already been bound to the exact source
		// operation node by sourceOperationMappingFactFindings. A concrete,
		// source-cited non-JSON type is therefore provider evidence for the
		// binary lane; do not replace that evidence with a closed MIME allow-list.
		return true
	}
	return false
}

func sourceOperationMappingConcreteMediaType(mediaType string) bool {
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	topLevel, subType, found := strings.Cut(mediaType, "/")
	return found && topLevel != "" && subType != "" && topLevel != "*" && subType != "*"
}

func sourceOperationMappingJSONMediaType(mediaType string) bool {
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	_, subType, found := strings.Cut(mediaType, "/")
	if !found {
		return false
	}
	return subType == "json" || strings.HasSuffix(subType, "+json")
}

func sourceOperationMappingCellFindings(operation sourceOperationMappingOperation, cell sourceOperationMappingLaneCell, locked declarationAdmissionReviewedOperation) []string {
	findings := []string{}
	add := func(format string, args ...any) {
		findings = append(findings, fmt.Sprintf(format, args...))
	}
	switch cell.State {
	case "implemented", "mapped_unproven":
		if cell.Reason != nil {
			add("source operation %q %s cell state %q must not declare a reason", operation.SourceOperationID, cell.Lane, cell.State)
		}
	case "missing_foundation":
		if cell.Reason == nil {
			add("source operation %q %s cell state missing_foundation requires a stable typed reason", operation.SourceOperationID, cell.Lane)
			break
		}
		if cell.Reason.Kind != "missing_foundation" {
			add("source operation %q %s missing_foundation cell reason kind must be missing_foundation", operation.SourceOperationID, cell.Lane)
		}
		if !sourceOperationMappingReasonID(cell.Reason.ID) {
			add("source operation %q %s missing_foundation cell reason ID must be stable and canonical", operation.SourceOperationID, cell.Lane)
		}
		if cell.Reason.Citation != nil && !sourceOperationMappingCitationText(*cell.Reason.Citation) {
			add("source operation %q %s missing_foundation cell reason citation must be canonical", operation.SourceOperationID, cell.Lane)
		}
	case "not_applicable":
		if cell.Reason == nil {
			add("source operation %q %s not_applicable cell requires source evidence", operation.SourceOperationID, cell.Lane)
			break
		}
		if cell.Reason.Kind != "provider_not_applicable" {
			add("source operation %q %s not_applicable cell reason kind must be provider_not_applicable", operation.SourceOperationID, cell.Lane)
		}
		if !sourceOperationMappingReasonID(cell.Reason.ID) {
			add("source operation %q %s not_applicable cell reason ID must be stable and canonical", operation.SourceOperationID, cell.Lane)
		}
		if cell.Reason.Citation == nil {
			add("source operation %q %s not_applicable cell requires a source citation", operation.SourceOperationID, cell.Lane)
		} else if err := sourceOperationMappingValidateExactCitation(*cell.Reason.Citation, locked); err != nil {
			add("source operation %q %s not_applicable cell source citation: %v", operation.SourceOperationID, cell.Lane, err)
		}
	}
	return findings
}

func sourceOperationMappingValidateCitation(citation sourceOperationMappingSourceCitation, sourceURL string) error {
	if citation.SourceURL != sourceURL {
		return fmt.Errorf("source URL %q does not equal source-lock URL %q", citation.SourceURL, sourceURL)
	}
	if !sourceOperationMappingCitationText(citation) {
		return fmt.Errorf("source URL and location must be nonempty and canonical")
	}
	return nil
}

func sourceOperationMappingCitationText(citation sourceOperationMappingSourceCitation) bool {
	return sourceOperationMappingText(citation.SourceURL, 4096) && sourceOperationMappingText(citation.Location, 4096)
}

func sourceOperationMappingReasonID(value string) bool {
	if !sourceOperationMappingText(value, 256) || strings.ContainsAny(value, " \t") {
		return false
	}
	return true
}

func sourceOperationMappingText(value string, limit int) bool {
	return sourceImportReferenceText(value, limit)
}

func sourceOperationMappingValuesFinding(sourceOperationID, fact string, values []string) string {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !sourceOperationMappingText(value, 256) {
			return fmt.Sprintf("source operation %q %s fact has a noncanonical value", sourceOperationID, fact)
		}
		if _, duplicate := seen[value]; duplicate {
			return fmt.Sprintf("source operation %q %s fact repeats value %q", sourceOperationID, fact, value)
		}
		seen[value] = struct{}{}
	}
	return ""
}

func sourceOperationMappingArtifactPath(path string) bool {
	if path == "" || filepath.IsAbs(path) || strings.Contains(path, "\\") || filepath.ToSlash(filepath.Clean(filepath.FromSlash(path))) != path {
		return false
	}
	return path != ".." && !strings.HasPrefix(path, "../")
}

func sourceOperationMappingCellKey(sourceOperationID, lane string) string {
	return sourceOperationID + "\x00" + lane
}
