package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	deferredVisibilitySchemaVersion             = 1
	deferredVisibilityMappedUnprovenReasonID    = "deferred_visibility.mapped_unproven.v1"
	deferredVisibilityMissingFoundationReasonID = "deferred_visibility.missing_foundation.v1"
	deferredVisibilityProjectionAdmissionID     = "source.projection-admission.v1"
	deferredVisibilityRuntimeClaim              = "none"
	deferredVisibilityFoundationCatalogPath     = "docs/connector-canon/foundations/catalog.json"
	deferredVisibilityMissingFoundationFile     = "missing-foundation.json"
)

// deferredVisibilityOptions makes the command an explicit, read-only
// validation action. The command has no output-file flag by design: a report
// is a discovery diagnostic, not a connector declaration artifact.
type deferredVisibilityOptions struct {
	ManifestPath string
	Check        bool
	JSON         bool
}

// deferredVisibilityReport contains mapping evidence only. It deliberately
// has no command, operation, stream, write, transport, credential, executor,
// or descriptor field. A consumer must not use it to admit execution.
type deferredVisibilityReport struct {
	SchemaVersion           int                           `json:"schema_version"`
	Cohort                  string                        `json:"cohort"`
	MappingOnly             bool                          `json:"mapping_only"`
	PrimarySourceOperations int                           `json:"primary_source_operations"`
	SupplementalSourceRows  int                           `json:"supplemental_source_rows"`
	SourceRows              int                           `json:"source_rows"`
	MatrixCells             int                           `json:"matrix_cells"`
	DeferredCells           int                           `json:"deferred_cells"`
	ExecutableDeclarations  int                           `json:"executable_declarations"`
	LaneCounts              []deferredVisibilityLaneCount `json:"lane_counts"`
	Entries                 []deferredVisibilityEntry     `json:"entries"`
}

type deferredVisibilityLaneCount struct {
	Lane              string `json:"lane"`
	MappedUnproven    int    `json:"mapped_unproven"`
	MissingFoundation int    `json:"missing_foundation"`
	Implemented       int    `json:"implemented"`
	NotApplicable     int    `json:"not_applicable"`
}

type deferredVisibilityEntry struct {
	Connector         string                       `json:"connector"`
	SourceOperationID string                       `json:"source_operation_id"`
	Lane              string                       `json:"lane"`
	Visibility        string                       `json:"visibility"`
	SourceDisposition string                       `json:"source_disposition"`
	Source            deferredVisibilitySource     `json:"source"`
	SourceFact        map[string]any               `json:"source_fact"`
	Reason            deferredVisibilityReason     `json:"reason"`
	Capability        deferredVisibilityCapability `json:"capability"`
	RuntimeClaim      string                       `json:"runtime_claim"`
}

type deferredVisibilitySource struct {
	SourceLock       string `json:"source_lock"`
	SourceLockSHA256 string `json:"source_lock_sha256"`
	// SourceLockOperationID is the exact operation identity inside the cited
	// immutable source lock. It can differ from the connector-owned matrix
	// identity when a provider's source lock preserves a bare operation ID.
	// Both identities are reported so a deferred row stays auditable without
	// synthesizing a descriptor or a command binding.
	SourceLockOperationID string `json:"source_lock_operation_id"`
	CitationURL           string `json:"citation_url"`
	SourceLocation        string `json:"source_location"`
	Protocol              string `json:"protocol"`
	Method                string `json:"method"`
	Path                  string `json:"path"`
	ProviderOperationID   string `json:"provider_operation_id"`
}

type deferredVisibilityReason struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
}

type deferredVisibilityCapability struct {
	Kind        string `json:"kind"`
	ID          string `json:"id"`
	AtlasID     string `json:"atlas_id"`
	Requirement string `json:"requirement"`
}

type deferredVisibilityMatrix struct {
	Connector       string
	DefaultLockPath string
	FoundationAtlas map[string]any
	Rows            []deferredVisibilityMatrixRow
}

type deferredVisibilityMatrixRow struct {
	SourceID            string
	SourceLockPath      string
	ProviderOperationID string
	SourceURL           string
	SourceLocation      string
	SourceFact          map[string]any
	Method              string
	Path                string
	Cells               map[string]deferredVisibilityMatrixCell
}

type deferredVisibilityMatrixCell struct {
	Lane           string
	State          string
	Reason         string
	Mapping        map[string]any
	Citation       map[string]any
	SourceEvidence map[string]any
}

type deferredVisibilitySourceLock struct {
	RelativePath string
	SHA256       string
	Reviewed     declarationAdmissionReviewedSourceLock
}

type deferredVisibilityFoundationRecord struct {
	ID              string
	AtlasCapability string
	AffectedLane    string
	SourceIDs       map[string]bool
	Reason          string
}

func runDeferredVisibility(args []string, stdout, stderr io.Writer) int {
	if len(args) == 2 && (args[1] == "-h" || args[1] == "--help") {
		logln(stdout, deferredVisibilityUsage())
		return 0
	}
	opts, err := parseDeferredVisibilityOptions(args)
	if err != nil {
		logf(stderr, "connectorgen deferred-visibility: %v\n", err)
		return 2
	}
	root, err := repoRoot()
	if err != nil {
		logf(stderr, "connectorgen deferred-visibility: resolve repository root: %v\n", err)
		return 1
	}
	report, err := deferredVisibilityFromRepository(root, opts.ManifestPath)
	if err != nil {
		logf(stderr, "connectorgen deferred-visibility: %v\n", err)
		return 1
	}
	if opts.JSON {
		encoded, marshalErr := json.MarshalIndent(report, "", "  ")
		if marshalErr != nil {
			logf(stderr, "connectorgen deferred-visibility: encode report: %v\n", marshalErr)
			return 1
		}
		logln(stdout, string(encoded))
		return 0
	}
	logf(stdout, "connectorgen deferred-visibility: mapping-only; %d primary source operation(s), %d source row(s) including %d supplemental row(s), %d matrix cell(s), %d deferred cell(s), 0 executable declaration(s)\n", report.PrimarySourceOperations, report.SourceRows, report.SupplementalSourceRows, report.MatrixCells, report.DeferredCells)
	return 0
}

func deferredVisibilityUsage() string {
	return `usage: connectorgen deferred-visibility <manifest> --check [--json]

Reads frozen source locks, connector-local source-lane matrices, existing
connector gap ledgers, and the Foundation Atlas to report source-backed
deferred visibility. It does not write a descriptor, operation, stream, write,
CLI, transport, credential, executor, or runtime artifact, and it never makes
provider I/O.`
}

func parseDeferredVisibilityOptions(args []string) (deferredVisibilityOptions, error) {
	var options deferredVisibilityOptions
	for _, argument := range args[1:] {
		switch argument {
		case "--check":
			if options.Check {
				return deferredVisibilityOptions{}, fmt.Errorf("--check may be specified only once")
			}
			options.Check = true
		case "--json":
			if options.JSON {
				return deferredVisibilityOptions{}, fmt.Errorf("--json may be specified only once")
			}
			options.JSON = true
		default:
			if strings.HasPrefix(argument, "-") {
				return deferredVisibilityOptions{}, fmt.Errorf("unknown flag %q", argument)
			}
			if options.ManifestPath != "" {
				return deferredVisibilityOptions{}, fmt.Errorf("unexpected extra argument %q", argument)
			}
			options.ManifestPath = argument
		}
	}
	if options.ManifestPath == "" {
		return deferredVisibilityOptions{}, fmt.Errorf("manifest path is required")
	}
	if !options.Check {
		return deferredVisibilityOptions{}, fmt.Errorf("--check is required; deferred-visibility is validation only")
	}
	return options, nil
}

// deferredVisibilityFromRepository is intentionally a source-evidence reader,
// not a source-import/materialization path. It first proves the frozen primary
// cohort, then binds each matrix row to its connector-owned primary or
// explicitly named supplemental source lock.
func deferredVisibilityFromRepository(root, manifestPath string) (deferredVisibilityReport, error) {
	cohortReport, err := sourceOperationMappingCohortPathCheck(root, manifestPath)
	if err != nil {
		return deferredVisibilityReport{}, fmt.Errorf("validate frozen source cohort: %w", err)
	}
	if len(cohortReport.Findings) != 0 {
		return deferredVisibilityReport{}, fmt.Errorf("frozen source cohort has %d finding(s)", len(cohortReport.Findings))
	}
	manifest, err := sourceOperationMappingCohortReadManifest(manifestPath)
	if err != nil {
		return deferredVisibilityReport{}, err
	}
	atlas, err := deferredVisibilityFoundationAtlas(root)
	if err != nil {
		return deferredVisibilityReport{}, err
	}
	if !atlas[deferredVisibilityProjectionAdmissionID] {
		return deferredVisibilityReport{}, fmt.Errorf("Foundation Atlas does not define required mapping capability %q", deferredVisibilityProjectionAdmissionID)
	}

	report := deferredVisibilityReport{
		SchemaVersion:          deferredVisibilitySchemaVersion,
		Cohort:                 manifest.Cohort,
		MappingOnly:            true,
		ExecutableDeclarations: 0,
		LaneCounts:             deferredVisibilityEmptyLaneCounts(),
		Entries:                make([]deferredVisibilityEntry, 0),
	}
	countsByLane := make(map[string]*deferredVisibilityLaneCount, len(retainedSourceMappingLaneOrder))
	for index := range report.LaneCounts {
		countsByLane[report.LaneCounts[index].Lane] = &report.LaneCounts[index]
	}

	seenPrimary := make(map[string]bool, manifest.SourceOperationCount)
	seenRows := make(map[string]bool, manifest.SourceOperationCount)
	locks := map[string]deferredVisibilitySourceLock{}

	for _, cohortEntry := range manifest.SourceLocks {
		primaryRelative, err := deferredVisibilityRelativeSourceLockPath(cohortEntry.Connector, cohortEntry.Path)
		if err != nil {
			return deferredVisibilityReport{}, fmt.Errorf("%s primary source lock: %w", cohortEntry.Connector, err)
		}
		primary, err := deferredVisibilityLoadSourceLock(root, cohortEntry.Connector, primaryRelative, locks)
		if err != nil {
			return deferredVisibilityReport{}, err
		}
		primaryIDs := make(map[string]bool, len(primary.Reviewed.Operations))
		for sourceID := range primary.Reviewed.Operations {
			primaryIDs[sourceID] = true
		}
		if len(primaryIDs) != cohortEntry.SourceOperationCount {
			return deferredVisibilityReport{}, fmt.Errorf("%s primary source lock has %d operation(s), want frozen cohort count %d", cohortEntry.Connector, len(primaryIDs), cohortEntry.SourceOperationCount)
		}
		report.PrimarySourceOperations += len(primaryIDs)

		matrixPath, err := deferredVisibilityOwnedMatrixPath(root, cohortEntry.Connector, cohortEntry.MatrixPath)
		if err != nil {
			return deferredVisibilityReport{}, err
		}
		matrixRaw, err := os.ReadFile(matrixPath)
		if err != nil {
			return deferredVisibilityReport{}, fmt.Errorf("read %s source-lane matrix: %w", cohortEntry.Connector, err)
		}
		matrix, err := decodeDeferredVisibilityMatrix(matrixRaw, cohortEntry.Connector, primaryRelative)
		if err != nil {
			return deferredVisibilityReport{}, fmt.Errorf("decode %s source-lane matrix: %w", cohortEntry.Connector, err)
		}
		ledger, err := deferredVisibilityReadFoundationLedger(root, cohortEntry.Connector)
		if err != nil {
			return deferredVisibilityReport{}, err
		}

		for _, row := range matrix.Rows {
			rowKey := cohortEntry.Connector + "\x00" + row.SourceID
			if seenRows[rowKey] {
				return deferredVisibilityReport{}, fmt.Errorf("%s source-lane matrix duplicates source ID %q", cohortEntry.Connector, row.SourceID)
			}
			seenRows[rowKey] = true

			rowLock, err := deferredVisibilityLoadSourceLock(root, cohortEntry.Connector, row.SourceLockPath, locks)
			if err != nil {
				return deferredVisibilityReport{}, err
			}
			lockedSourceID, operation, resolveErr := deferredVisibilityResolveSourceOperation(row, rowLock.Reviewed)
			if resolveErr != nil {
				return deferredVisibilityReport{}, fmt.Errorf("%s source row %q: %w", cohortEntry.Connector, row.SourceID, resolveErr)
			}
			if err := deferredVisibilityValidateSourceFact(row, operation); err != nil {
				return deferredVisibilityReport{}, fmt.Errorf("%s source row %q: %w", cohortEntry.Connector, row.SourceID, err)
			}

			isPrimary := row.SourceLockPath == primaryRelative
			if isPrimary {
				if !primaryIDs[lockedSourceID] {
					return deferredVisibilityReport{}, fmt.Errorf("%s source row %q cites primary lock but is outside frozen source denominator", cohortEntry.Connector, row.SourceID)
				}
				seenPrimary[cohortEntry.Connector+"\x00"+lockedSourceID] = true
			} else {
				report.SupplementalSourceRows++
			}
			report.SourceRows++

			if deferredVisibilityRowSemanticMutation(row) {
				if row.Cells["direct_write"].State == "not_applicable" || row.Cells["reverse_etl"].State == "not_applicable" {
					return deferredVisibilityReport{}, fmt.Errorf("%s source row %q has source semantic mutation evidence but hides a direct_write or reverse_etl lane", cohortEntry.Connector, row.SourceID)
				}
			}

			for _, lane := range retainedSourceMappingLaneOrder {
				cell := row.Cells[lane]
				report.MatrixCells++
				counter := countsByLane[lane]
				switch cell.State {
				case "mapped_unproven":
					counter.MappedUnproven++
				case "missing_foundation":
					counter.MissingFoundation++
				case "implemented":
					counter.Implemented++
				case "not_applicable":
					counter.NotApplicable++
				default:
					return deferredVisibilityReport{}, fmt.Errorf("%s source row %q lane %q has unsupported state %q", cohortEntry.Connector, row.SourceID, lane, cell.State)
				}
				if cell.State != "mapped_unproven" && cell.State != "missing_foundation" {
					continue
				}
				entry, entryErr := deferredVisibilityEntryForCell(cohortEntry.Connector, row, rowLock, lockedSourceID, operation, cell, matrix.FoundationAtlas, ledger, atlas)
				if entryErr != nil {
					return deferredVisibilityReport{}, entryErr
				}
				report.Entries = append(report.Entries, entry)
			}
		}

		for sourceID := range primaryIDs {
			if !seenPrimary[cohortEntry.Connector+"\x00"+sourceID] {
				return deferredVisibilityReport{}, fmt.Errorf("%s frozen source ID %q is absent from its source-lane matrix", cohortEntry.Connector, sourceID)
			}
		}
	}
	if report.PrimarySourceOperations != manifest.SourceOperationCount {
		return deferredVisibilityReport{}, fmt.Errorf("primary source operations %d do not equal frozen cohort count %d", report.PrimarySourceOperations, manifest.SourceOperationCount)
	}
	report.DeferredCells = len(report.Entries)
	sort.Slice(report.Entries, func(i, j int) bool {
		left, right := report.Entries[i], report.Entries[j]
		if left.Connector != right.Connector {
			return left.Connector < right.Connector
		}
		if left.SourceOperationID != right.SourceOperationID {
			return left.SourceOperationID < right.SourceOperationID
		}
		return deferredVisibilityLaneIndex(left.Lane) < deferredVisibilityLaneIndex(right.Lane)
	})
	if err := deferredVisibilityValidateReport(report); err != nil {
		return deferredVisibilityReport{}, err
	}
	return report, nil
}

func deferredVisibilityEmptyLaneCounts() []deferredVisibilityLaneCount {
	counts := make([]deferredVisibilityLaneCount, 0, len(retainedSourceMappingLaneOrder))
	for _, lane := range retainedSourceMappingLaneOrder {
		counts = append(counts, deferredVisibilityLaneCount{Lane: lane})
	}
	return counts
}

func deferredVisibilityLaneIndex(lane string) int {
	for index, candidate := range retainedSourceMappingLaneOrder {
		if lane == candidate {
			return index
		}
	}
	return len(retainedSourceMappingLaneOrder)
}

func deferredVisibilityRelativeSourceLockPath(connector, raw string) (string, error) {
	prefix := filepath.ToSlash(filepath.Join("internal", "connectors", "defs", connector)) + "/"
	if !strings.HasPrefix(raw, prefix) {
		return "", fmt.Errorf("must be connector-owned path beneath %q", prefix)
	}
	return deferredVisibilityValidateRelativeSourceLockPath(connector, strings.TrimPrefix(raw, prefix))
}

func deferredVisibilityValidateRelativeSourceLockPath(connector, raw string) (string, error) {
	cleaned := filepath.ToSlash(filepath.Clean(filepath.FromSlash(raw)))
	if raw == "" || filepath.IsAbs(raw) || strings.Contains(raw, "\\") || cleaned != raw || !strings.HasPrefix(raw, "sources/") || !strings.HasSuffix(raw, "-operation-source-lock.json") {
		return "", fmt.Errorf("must be one canonical connector-owned source lock path beneath sources/")
	}
	base := strings.TrimSuffix(strings.TrimPrefix(raw, "sources/"), "-operation-source-lock.json")
	if base == "" || !strings.HasPrefix(base, connector) {
		return "", fmt.Errorf("must name a %q connector-owned source lock", connector)
	}
	return raw, nil
}

func deferredVisibilityOwnedMatrixPath(root, connector, raw string) (string, error) {
	if err := sourceOperationMappingCohortMatrixPath(connector, raw); err != nil {
		return "", fmt.Errorf("%s source-lane matrix: %w", connector, err)
	}
	return sourceOperationMappingCohortOwnedPath(root, connector, raw, "-source-lane-matrix.json")
}

func deferredVisibilityOwnedSourceLockPath(root, connector, relative string) (string, error) {
	cleaned, err := deferredVisibilityValidateRelativeSourceLockPath(connector, relative)
	if err != nil {
		return "", err
	}
	base := filepath.Join(root, "internal", "connectors", "defs", connector)
	resolvedBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		return "", fmt.Errorf("resolve connector definition root: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(filepath.Join(resolvedBase, filepath.FromSlash(cleaned)))
	if err != nil {
		return "", fmt.Errorf("resolve source lock %q: %w", relative, err)
	}
	relativeResolved, err := filepath.Rel(resolvedBase, resolved)
	if err != nil || relativeResolved == ".." || strings.HasPrefix(relativeResolved, ".."+string(filepath.Separator)) || filepath.ToSlash(relativeResolved) != cleaned {
		return "", fmt.Errorf("source lock %q resolves outside connector-owned sources", relative)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("source lock %q must be a regular file", relative)
	}
	return resolved, nil
}

func deferredVisibilityLoadSourceLock(root, connector, relative string, loaded map[string]deferredVisibilitySourceLock) (deferredVisibilitySourceLock, error) {
	validated, err := deferredVisibilityValidateRelativeSourceLockPath(connector, relative)
	if err != nil {
		return deferredVisibilitySourceLock{}, err
	}
	key := connector + "\x00" + validated
	if existing, found := loaded[key]; found {
		return existing, nil
	}
	path, err := deferredVisibilityOwnedSourceLockPath(root, connector, validated)
	if err != nil {
		return deferredVisibilitySourceLock{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return deferredVisibilitySourceLock{}, fmt.Errorf("read %s source lock %q: %w", connector, validated, err)
	}
	reviewed, err := parseDeclarationAdmissionSourceLock(raw, connector)
	if err != nil {
		return deferredVisibilitySourceLock{}, fmt.Errorf("parse %s source lock %q: %w", connector, validated, err)
	}
	if len(reviewed.Operations) == 0 {
		return deferredVisibilitySourceLock{}, fmt.Errorf("%s source lock %q has no operation identities", connector, validated)
	}
	loadedLock := deferredVisibilitySourceLock{RelativePath: validated, SHA256: sourceOperationMappingSHA256(raw), Reviewed: reviewed}
	loaded[key] = loadedLock
	return loadedLock, nil
}

// deferredVisibilityResolveSourceOperation binds a connector-local matrix row
// to exactly one retained source-lock operation. A matrix may carry its own
// stable source identity (for example, a connector-prefixed crosswalk ID)
// while the immutable provider lock retains the provider operation ID. That
// relationship must be declared by the matrix's source facts and match the
// locked method/path; this is not an operation-name rewrite or a provider
// exception.
func deferredVisibilityResolveSourceOperation(row deferredVisibilityMatrixRow, lock declarationAdmissionReviewedSourceLock) (string, declarationAdmissionReviewedOperation, error) {
	if operation, found := lock.Operations[row.SourceID]; found {
		if row.ProviderOperationID != "" && operation.ProviderOperationID != row.ProviderOperationID {
			return "", declarationAdmissionReviewedOperation{}, fmt.Errorf("matrix source identity resolves to provider operation %q, not declared %q", operation.ProviderOperationID, row.ProviderOperationID)
		}
		if operation.Method != row.Method || operation.Path != row.Path {
			return "", declarationAdmissionReviewedOperation{}, fmt.Errorf("matrix source identity route %s %s does not match cited source lock %s %s", row.Method, row.Path, operation.Method, operation.Path)
		}
		return row.SourceID, operation, nil
	}
	if row.ProviderOperationID == "" {
		return "", declarationAdmissionReviewedOperation{}, fmt.Errorf("is not present in cited source lock and declares no provider operation identity")
	}

	var matchedID string
	var matched declarationAdmissionReviewedOperation
	for candidateID, candidate := range lock.Operations {
		if candidate.ProviderOperationID != row.ProviderOperationID || candidate.Method != row.Method || candidate.Path != row.Path {
			continue
		}
		if matchedID != "" {
			return "", declarationAdmissionReviewedOperation{}, fmt.Errorf("provider operation identity %q with route %s %s is ambiguous in cited source lock", row.ProviderOperationID, row.Method, row.Path)
		}
		matchedID = candidateID
		matched = candidate
	}
	if matchedID == "" {
		return "", declarationAdmissionReviewedOperation{}, fmt.Errorf("is not present in cited source lock and provider operation identity %q with route %s %s does not resolve", row.ProviderOperationID, row.Method, row.Path)
	}
	return matchedID, matched, nil
}

func deferredVisibilityFoundationAtlas(root string) (map[string]bool, error) {
	path := filepath.Join(root, filepath.FromSlash(deferredVisibilityFoundationCatalogPath))
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Foundation Atlas: %w", err)
	}
	var document struct {
		Foundations []struct {
			ID string `json:"id"`
		} `json:"foundations"`
	}
	if err := decodeSourceJSON(raw, &document); err != nil {
		return nil, fmt.Errorf("decode Foundation Atlas: %w", err)
	}
	ids := make(map[string]bool, len(document.Foundations))
	for _, foundation := range document.Foundations {
		if foundation.ID == "" || ids[foundation.ID] {
			return nil, fmt.Errorf("Foundation Atlas has missing or duplicate capability ID %q", foundation.ID)
		}
		ids[foundation.ID] = true
	}
	return ids, nil
}

func decodeDeferredVisibilityMatrix(raw []byte, connector, defaultLockPath string) (deferredVisibilityMatrix, error) {
	var value any
	if err := decodeSourceJSON(raw, &value); err != nil {
		return deferredVisibilityMatrix{}, err
	}
	root, err := retainedSourceMappingObject(value, "source-lane matrix")
	if err != nil {
		return deferredVisibilityMatrix{}, err
	}
	version, err := retainedSourceMappingInteger(root["schema_version"], "schema_version")
	if err != nil || version != 1 {
		if err != nil {
			return deferredVisibilityMatrix{}, err
		}
		return deferredVisibilityMatrix{}, fmt.Errorf("unsupported source-lane matrix schema version %d", version)
	}
	actualConnector, err := retainedSourceMappingString(root["connector"], "connector")
	if err != nil {
		return deferredVisibilityMatrix{}, err
	}
	if actualConnector != connector {
		return deferredVisibilityMatrix{}, fmt.Errorf("connector %q does not match requested connector %q", actualConnector, connector)
	}
	lanes, err := retainedSourceMappingStringArray(root["lanes"], "lanes")
	if err != nil {
		return deferredVisibilityMatrix{}, err
	}
	if !retainedSourceMappingExactLaneSet(lanes) {
		return deferredVisibilityMatrix{}, fmt.Errorf("must declare exactly the seven fixed lanes")
	}
	if rawSourceLock, found := root["source_lock"]; found && rawSourceLock != nil {
		if object, objectErr := retainedSourceMappingObject(rawSourceLock, "source_lock"); objectErr == nil {
			if path, pathErr := retainedSourceMappingString(object["path"], "source_lock.path"); pathErr == nil {
				defaultLockPath = path
			} else {
				return deferredVisibilityMatrix{}, pathErr
			}
		} else {
			return deferredVisibilityMatrix{}, objectErr
		}
	}
	if _, err := deferredVisibilityValidateRelativeSourceLockPath(connector, defaultLockPath); err != nil {
		return deferredVisibilityMatrix{}, fmt.Errorf("default source lock: %w", err)
	}
	matrix := deferredVisibilityMatrix{Connector: connector, DefaultLockPath: defaultLockPath, Rows: make([]deferredVisibilityMatrixRow, 0)}
	if rawAtlas, found := root["foundation_atlas"]; found {
		atlas, atlasErr := retainedSourceMappingObject(rawAtlas, "foundation_atlas")
		if atlasErr != nil {
			return deferredVisibilityMatrix{}, atlasErr
		}
		matrix.FoundationAtlas = atlas
	}
	_, hasSourceOperations := root["source_operations"]
	_, hasOperations := root["operations"]
	if hasSourceOperations == hasOperations {
		return deferredVisibilityMatrix{}, fmt.Errorf("must declare exactly one of source_operations or operations")
	}
	if hasSourceOperations {
		rows, rowsErr := retainedSourceMappingArray(root["source_operations"], "source_operations")
		if rowsErr != nil {
			return deferredVisibilityMatrix{}, rowsErr
		}
		for index, rawRow := range rows {
			row, rowErr := decodeDeferredVisibilitySourceOperationsRow(rawRow, connector, defaultLockPath, index)
			if rowErr != nil {
				return deferredVisibilityMatrix{}, rowErr
			}
			matrix.Rows = append(matrix.Rows, row)
		}
	} else {
		rows, rowsErr := retainedSourceMappingArray(root["operations"], "operations")
		if rowsErr != nil {
			return deferredVisibilityMatrix{}, rowsErr
		}
		for index, rawRow := range rows {
			row, rowErr := decodeDeferredVisibilityOperationsRow(rawRow, connector, defaultLockPath, index)
			if rowErr != nil {
				return deferredVisibilityMatrix{}, rowErr
			}
			matrix.Rows = append(matrix.Rows, row)
		}
	}
	if len(matrix.Rows) == 0 {
		return deferredVisibilityMatrix{}, fmt.Errorf("has no source rows")
	}
	return matrix, nil
}

func decodeDeferredVisibilitySourceOperationsRow(raw any, connector, defaultLockPath string, index int) (deferredVisibilityMatrixRow, error) {
	field := fmt.Sprintf("source_operations[%d]", index)
	row, err := retainedSourceMappingObject(raw, field)
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	if _, found := row["source_operation_id"]; found {
		return deferredVisibilityMatrixRow{}, fmt.Errorf("%s must use source_id, not source_operation_id", field)
	}
	sourceID, err := retainedSourceMappingString(row["source_id"], field+".source_id")
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	facts, err := retainedSourceMappingObject(row["source_facts"], field+".source_facts")
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	if evidenceID, found := deferredVisibilityDirectString(facts, "source_id"); found && evidenceID != sourceID {
		return deferredVisibilityMatrixRow{}, fmt.Errorf("%s.source_facts.source_id %q does not match row source_id %q", field, evidenceID, sourceID)
	}
	method, err := retainedSourceMappingString(facts["method"], field+".source_facts.method")
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	path, err := retainedSourceMappingString(facts["path"], field+".source_facts.path")
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	lockPath := defaultLockPath
	if value, found := deferredVisibilityDirectString(facts, "source_lock"); found {
		lockPath = value
	}
	lockPath, err = deferredVisibilityValidateRelativeSourceLockPath(connector, lockPath)
	if err != nil {
		return deferredVisibilityMatrixRow{}, fmt.Errorf("%s source lock: %w", field, err)
	}
	if _, found := row["cells"]; found {
		return deferredVisibilityMatrixRow{}, fmt.Errorf("%s cannot mix cells with lanes", field)
	}
	cells, err := decodeDeferredVisibilityLanes(row["lanes"], field+".lanes")
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	providerOperationID, _ := deferredVisibilityDirectString(facts, "operation_id")
	sourceURL, sourceLocation := deferredVisibilitySourceFactCitation(facts)
	return deferredVisibilityMatrixRow{
		SourceID:            sourceID,
		SourceLockPath:      lockPath,
		ProviderOperationID: providerOperationID,
		SourceURL:           sourceURL,
		SourceLocation:      sourceLocation,
		SourceFact:          facts,
		Method:              method,
		Path:                path,
		Cells:               cells,
	}, nil
}

func decodeDeferredVisibilityOperationsRow(raw any, connector, defaultLockPath string, index int) (deferredVisibilityMatrixRow, error) {
	field := fmt.Sprintf("operations[%d]", index)
	row, err := retainedSourceMappingObject(raw, field)
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	_, hasSourceID := row["source_id"]
	_, hasSourceOperationID := row["source_operation_id"]
	if hasSourceID == hasSourceOperationID {
		return deferredVisibilityMatrixRow{}, fmt.Errorf("%s must declare exactly one of source_id or source_operation_id", field)
	}
	idField := "source_id"
	if hasSourceOperationID {
		idField = "source_operation_id"
	}
	sourceID, err := retainedSourceMappingString(row[idField], field+"."+idField)
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	facts, err := retainedSourceMappingObject(row["facts"], field+".facts")
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	method, err := retainedSourceMappingString(row["method"], field+".method")
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	path, err := retainedSourceMappingString(row["path"], field+".path")
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	if _, found := row["lanes"]; found {
		return deferredVisibilityMatrixRow{}, fmt.Errorf("%s cannot mix lanes with cells", field)
	}
	cells, err := decodeDeferredVisibilityCells(row["cells"], field+".cells")
	if err != nil {
		return deferredVisibilityMatrixRow{}, err
	}
	providerOperationID := deferredVisibilityFirstString(row, "provider_operation_id", "operation_id")
	if providerOperationID == "" {
		providerOperationID, _ = deferredVisibilityDirectString(facts, "operation_id")
	}
	sourceURL := deferredVisibilityFirstString(row, "source_url", "citation_url")
	sourceLocation := deferredVisibilityFirstString(row, "source_location", "location")
	if factURL, factLocation := deferredVisibilitySourceFactCitation(facts); sourceURL == "" || sourceLocation == "" {
		if sourceURL == "" {
			sourceURL = factURL
		}
		if sourceLocation == "" {
			sourceLocation = factLocation
		}
	}
	return deferredVisibilityMatrixRow{
		SourceID:            sourceID,
		SourceLockPath:      defaultLockPath,
		ProviderOperationID: providerOperationID,
		SourceURL:           sourceURL,
		SourceLocation:      sourceLocation,
		SourceFact:          facts,
		Method:              method,
		Path:                path,
		Cells:               cells,
	}, nil
}

// deferredVisibilitySourceFactCitation reads only declared source-fact
// evidence. Matrix variants place the citation either directly in facts or
// below a `citation` object; both remain data, rather than inferred URLs.
func deferredVisibilitySourceFactCitation(facts map[string]any) (string, string) {
	url := deferredVisibilityFirstString(facts, "source_url", "citation_url", "url")
	location := deferredVisibilityFirstString(facts, "source_location", "location")
	if citation, found, err := deferredVisibilityOptionalObject(facts, "citation", "source_facts"); err == nil && found {
		if url == "" {
			url = deferredVisibilityFirstString(citation, "source_url", "citation_url", "url")
		}
		if location == "" {
			location = deferredVisibilityFirstString(citation, "source_location", "location")
		}
	}
	return url, location
}

func decodeDeferredVisibilityLanes(raw any, field string) (map[string]deferredVisibilityMatrixCell, error) {
	lanes, err := retainedSourceMappingObject(raw, field)
	if err != nil {
		return nil, err
	}
	if len(lanes) != len(retainedSourceMappingLaneOrder) {
		return nil, fmt.Errorf("%s must declare exactly seven lane cells", field)
	}
	result := make(map[string]deferredVisibilityMatrixCell, len(lanes))
	for _, lane := range retainedSourceMappingLaneOrder {
		rawCell, found := lanes[lane]
		if !found {
			return nil, fmt.Errorf("%s omits lane %q", field, lane)
		}
		cell, err := retainedSourceMappingObject(rawCell, field+"."+lane)
		if err != nil {
			return nil, err
		}
		applicability, err := retainedSourceMappingString(cell["applicability"], field+"."+lane+".applicability")
		if err != nil {
			return nil, err
		}
		disposition, err := retainedSourceMappingString(cell["disposition"], field+"."+lane+".disposition")
		if err != nil {
			return nil, err
		}
		if applicability == "not_applicable" && disposition != "not_applicable" {
			return nil, fmt.Errorf("%s.%s has not_applicable applicability with disposition %q", field, lane, disposition)
		}
		if applicability != "not_applicable" && applicability != "applicable" && applicability != "source_candidate" {
			return nil, fmt.Errorf("%s.%s has unknown applicability %q", field, lane, applicability)
		}
		decoded, err := decodeDeferredVisibilityCell(lane, disposition, cell, field+"."+lane)
		if err != nil {
			return nil, err
		}
		result[lane] = decoded
	}
	for lane := range lanes {
		if !retainedSourceMappingKnownLane(lane) {
			return nil, fmt.Errorf("%s has unknown lane %q", field, lane)
		}
	}
	return result, nil
}

func decodeDeferredVisibilityCells(raw any, field string) (map[string]deferredVisibilityMatrixCell, error) {
	cells, err := retainedSourceMappingArray(raw, field)
	if err != nil {
		return nil, err
	}
	if len(cells) != len(retainedSourceMappingLaneOrder) {
		return nil, fmt.Errorf("%s must declare exactly seven lane cells", field)
	}
	result := make(map[string]deferredVisibilityMatrixCell, len(cells))
	for index, rawCell := range cells {
		cell, err := retainedSourceMappingObject(rawCell, fmt.Sprintf("%s[%d]", field, index))
		if err != nil {
			return nil, err
		}
		lane, err := retainedSourceMappingString(cell["lane"], fmt.Sprintf("%s[%d].lane", field, index))
		if err != nil {
			return nil, err
		}
		if !retainedSourceMappingKnownLane(lane) || result[lane].Lane != "" {
			return nil, fmt.Errorf("%s has unknown or duplicate lane %q", field, lane)
		}
		state, err := retainedSourceMappingString(cell["state"], fmt.Sprintf("%s[%d].state", field, index))
		if err != nil {
			return nil, err
		}
		decoded, err := decodeDeferredVisibilityCell(lane, state, cell, fmt.Sprintf("%s[%d]", field, index))
		if err != nil {
			return nil, err
		}
		result[lane] = decoded
	}
	return result, nil
}

func decodeDeferredVisibilityCell(lane, state string, raw map[string]any, field string) (deferredVisibilityMatrixCell, error) {
	switch state {
	case "not_applicable", "mapped_unproven", "missing_foundation", "implemented":
	default:
		return deferredVisibilityMatrixCell{}, fmt.Errorf("%s has unsupported state/disposition %q", field, state)
	}
	cell := deferredVisibilityMatrixCell{Lane: lane, State: state}
	if reason, found := deferredVisibilityDirectString(raw, "reason"); found {
		cell.Reason = reason
	}
	if mapping, found, err := deferredVisibilityOptionalObject(raw, "mapping", field); err != nil {
		return deferredVisibilityMatrixCell{}, err
	} else if found {
		cell.Mapping = mapping
	}
	if citation, found, err := deferredVisibilityOptionalObject(raw, "citation", field); err != nil {
		return deferredVisibilityMatrixCell{}, err
	} else if found {
		cell.Citation = citation
	}
	if evidence, found, err := deferredVisibilityOptionalObject(raw, "source_evidence", field); err != nil {
		return deferredVisibilityMatrixCell{}, err
	} else if found {
		cell.SourceEvidence = evidence
	}
	if err := deferredVisibilityNonExecutableMatrixClaim(cell, field); err != nil {
		return deferredVisibilityMatrixCell{}, err
	}
	return cell, nil
}

func deferredVisibilityOptionalObject(raw map[string]any, key, field string) (map[string]any, bool, error) {
	value, found := raw[key]
	if !found || value == nil {
		return nil, false, nil
	}
	object, err := retainedSourceMappingObject(value, field+"."+key)
	if err != nil {
		return nil, false, err
	}
	return object, true, nil
}

func deferredVisibilityDirectString(object map[string]any, key string) (string, bool) {
	value, found := object[key]
	if !found {
		return "", false
	}
	text, ok := value.(string)
	return text, ok && strings.TrimSpace(text) != ""
}

func deferredVisibilityNonExecutableMatrixClaim(cell deferredVisibilityMatrixCell, field string) error {
	if cell.Mapping == nil {
		return nil
	}
	claim, found := deferredVisibilityDirectString(cell.Mapping, "runtime_claim")
	if !found {
		return nil
	}
	normalized := strings.ToLower(claim)
	if strings.Contains(normalized, "implemented") && !strings.Contains(normalized, "not implemented") {
		return fmt.Errorf("%s mapping runtime_claim falsely asserts implementation", field)
	}
	if strings.Contains(normalized, "executable") && !strings.Contains(normalized, "not executable") {
		return fmt.Errorf("%s mapping runtime_claim falsely asserts execution", field)
	}
	return nil
}

func deferredVisibilityValidateSourceFact(row deferredVisibilityMatrixRow, operation declarationAdmissionReviewedOperation) error {
	if row.SourceID == "" || len(row.SourceFact) == 0 {
		return fmt.Errorf("missing source ID or source fact")
	}
	if row.Method != operation.Method || row.Path != operation.Path {
		return fmt.Errorf("source fact %s %s does not match cited source lock %s %s", row.Method, row.Path, operation.Method, operation.Path)
	}
	if row.ProviderOperationID != "" && row.ProviderOperationID != operation.ProviderOperationID {
		return fmt.Errorf("provider operation identity %q does not match cited source lock %q", row.ProviderOperationID, operation.ProviderOperationID)
	}
	if operation.SourceURL == "" || operation.Location == "" || operation.Protocol == "" {
		return fmt.Errorf("cited source lock has incomplete provider fact")
	}
	if row.SourceLocation != "" && row.SourceLocation != operation.Location {
		return fmt.Errorf("source fact citation location %q does not match cited source lock %q", row.SourceLocation, operation.Location)
	}
	if row.SourceURL != "" && !deferredVisibilityEqualAny(row.SourceURL, operation.SourceURL, operation.CitationURL, operation.PublishedSourceURL) {
		return fmt.Errorf("source fact URL %q does not match cited source lock", row.SourceURL)
	}
	locations := deferredVisibilityCitationValues("source_location", "location", row.SourceFact)
	for _, cell := range row.Cells {
		locations = append(locations, deferredVisibilityCitationValues("source_location", "location", cell.Citation, cell.SourceEvidence)...)
	}
	if !deferredVisibilityContains(locations, operation.Location) {
		return fmt.Errorf("source facts omit citation location %q from cited source lock", operation.Location)
	}
	urls := deferredVisibilityCitationValues("source_url", "url", "citation_url", row.SourceFact)
	for _, cell := range row.Cells {
		urls = append(urls, deferredVisibilityCitationValues("source_url", "url", "citation_url", cell.Citation, cell.SourceEvidence)...)
	}
	if len(urls) != 0 && !deferredVisibilityContainsAny(urls, operation.SourceURL, operation.CitationURL, operation.PublishedSourceURL) {
		return fmt.Errorf("source facts cite URL(s) that do not match the cited source lock")
	}
	return nil
}

func deferredVisibilityEqualAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if candidate != "" && value == candidate {
			return true
		}
	}
	return false
}

func deferredVisibilityCitationValues(keys ...any) []string {
	var names []string
	var values []any
	for _, key := range keys {
		switch value := key.(type) {
		case string:
			names = append(names, value)
		default:
			values = append(values, value)
		}
	}
	wanted := make(map[string]bool, len(names))
	for _, name := range names {
		wanted[name] = true
	}
	collected := make([]string, 0)
	var visit func(any)
	visit = func(value any) {
		switch typed := value.(type) {
		case map[string]any:
			for key, child := range typed {
				if wanted[key] {
					if text, ok := child.(string); ok && strings.TrimSpace(text) != "" {
						collected = append(collected, text)
					}
				}
				visit(child)
			}
		case []any:
			for _, child := range typed {
				visit(child)
			}
		}
	}
	for _, value := range values {
		visit(value)
	}
	return collected
}

func deferredVisibilityContains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func deferredVisibilityContainsAny(values []string, wanted ...string) bool {
	for _, candidate := range wanted {
		if candidate != "" && deferredVisibilityContains(values, candidate) {
			return true
		}
	}
	return false
}

// deferredVisibilityRowSemanticMutation honors semantic source facts either
// on the matrix row or in a lane's retained mapping/source-evidence record.
// It deliberately does not infer a mutation from HTTP method, so POST
// semantic reads remain direct-read evidence and a provider can define
// nonstandard action vocabulary without a hardcoded list.
func deferredVisibilityRowSemanticMutation(row deferredVisibilityMatrixRow) bool {
	if deferredVisibilitySourceSemanticMutation(row.SourceFact) {
		return true
	}
	for _, cell := range row.Cells {
		if deferredVisibilitySourceSemanticMutation(cell.Mapping) || deferredVisibilitySourceSemanticMutation(cell.SourceEvidence) {
			return true
		}
	}
	return false
}

func deferredVisibilitySourceSemanticMutation(fact map[string]any) bool {
	for _, path := range [][]string{
		{"operation_semantics", "state"},
		{"classification"},
		{"write", "kind"},
		{"action", "kind"},
		{"source_fact", "classification"},
		{"source_fact", "operation_semantics", "state"},
		{"source_fact", "write", "kind"},
		{"source_fact", "action", "kind"},
	} {
		if text, found := deferredVisibilityNestedString(fact, path...); found && strings.Contains(strings.ToLower(text), "mutation") {
			return true
		}
	}
	return false
}

func deferredVisibilityNestedString(object map[string]any, path ...string) (string, bool) {
	var current any = object
	for _, segment := range path {
		next, ok := current.(map[string]any)
		if !ok {
			return "", false
		}
		current = next[segment]
	}
	text, ok := current.(string)
	return text, ok && strings.TrimSpace(text) != ""
}

func deferredVisibilityEntryForCell(connector string, row deferredVisibilityMatrixRow, lock deferredVisibilitySourceLock, lockedSourceID string, operation declarationAdmissionReviewedOperation, cell deferredVisibilityMatrixCell, foundationAtlas map[string]any, ledger []deferredVisibilityFoundationRecord, atlas map[string]bool) (deferredVisibilityEntry, error) {
	if cell.Reason == "" {
		if kind, found := deferredVisibilityDirectString(cell.SourceEvidence, "kind"); found {
			cell.Reason = kind
		}
	}
	if cell.Reason == "" {
		cell.Reason = deferredVisibilitySourceFactDetail(row.SourceFact)
	}
	if cell.State == "mapped_unproven" && strings.TrimSpace(cell.Reason) == "" {
		return deferredVisibilityEntry{}, fmt.Errorf("%s source row %q lane %q has mapped_unproven disposition without a stable source reason", connector, row.SourceID, cell.Lane)
	}
	source := deferredVisibilitySource{
		SourceLock:            lock.RelativePath,
		SourceLockSHA256:      lock.SHA256,
		SourceLockOperationID: lockedSourceID,
		CitationURL:           deferredVisibilityCitationURL(operation),
		SourceLocation:        operation.Location,
		Protocol:              operation.Protocol,
		Method:                operation.Method,
		Path:                  operation.Path,
		ProviderOperationID:   operation.ProviderOperationID,
	}
	entry := deferredVisibilityEntry{
		Connector:         connector,
		SourceOperationID: row.SourceID,
		Lane:              cell.Lane,
		Visibility:        "deferred",
		SourceDisposition: cell.State,
		Source:            source,
		SourceFact:        row.SourceFact,
		RuntimeClaim:      deferredVisibilityRuntimeClaim,
	}
	if cell.State == "mapped_unproven" {
		entry.Reason = deferredVisibilityReason{ID: deferredVisibilityMappedUnprovenReasonID, Kind: "mapped_unproven", Detail: cell.Reason}
		entry.Capability = deferredVisibilityCapability{
			Kind:        "authoring_prerequisite",
			ID:          deferredVisibilityProjectionAdmissionID,
			AtlasID:     deferredVisibilityProjectionAdmissionID,
			Requirement: "field_complete_source_bound_declaration",
		}
		if !atlas[entry.Capability.AtlasID] {
			return deferredVisibilityEntry{}, fmt.Errorf("%s source row %q lane %q references unknown capability %q", connector, row.SourceID, cell.Lane, entry.Capability.AtlasID)
		}
		return entry, nil
	}

	foundation := deferredVisibilityResolveFoundation(row.SourceID, cell.Lane, cell, foundationAtlas, ledger)
	if foundation.ID == "" || foundation.AtlasCapability == "" || foundation.Reason == "" {
		return deferredVisibilityEntry{}, fmt.Errorf("%s source row %q lane %q missing_foundation disposition has no named foundation/capability/source reason", connector, row.SourceID, cell.Lane)
	}
	if !atlas[foundation.AtlasCapability] {
		return deferredVisibilityEntry{}, fmt.Errorf("%s source row %q lane %q references unknown Foundation Atlas capability %q", connector, row.SourceID, cell.Lane, foundation.AtlasCapability)
	}
	entry.Reason = deferredVisibilityReason{ID: deferredVisibilityMissingFoundationReasonID, Kind: "missing_foundation", Detail: foundation.Reason}
	entry.Capability = deferredVisibilityCapability{
		Kind:        "named_foundation",
		ID:          foundation.ID,
		AtlasID:     foundation.AtlasCapability,
		Requirement: foundation.Reason,
	}
	return entry, nil
}

// deferredVisibilitySourceFactDetail is a deterministic fallback for matrix
// forms that store semantic evidence once per operation rather than repeating
// a prose reason in every lane cell. It uses the source's declared semantic
// fact, never HTTP method or an operation-name heuristic.
func deferredVisibilitySourceFactDetail(fact map[string]any) string {
	for _, path := range [][]string{
		{"operation_semantics", "state"},
		{"write", "kind"},
		{"action", "kind"},
		{"classification"},
		{"pagination", "kind"},
		{"extractability", "kind"},
		{"source_summary"},
	} {
		if value, found := deferredVisibilityNestedString(fact, path...); found {
			return "source_fact." + strings.Join(path, ".") + "=" + value
		}
	}
	return ""
}

func deferredVisibilityCitationURL(operation declarationAdmissionReviewedOperation) string {
	if operation.CitationURL != "" {
		return operation.CitationURL
	}
	if operation.SourceURL != "" {
		return operation.SourceURL
	}
	return operation.PublishedSourceURL
}

func deferredVisibilityResolveFoundation(sourceID, lane string, cell deferredVisibilityMatrixCell, foundationAtlas map[string]any, ledger []deferredVisibilityFoundationRecord) deferredVisibilityFoundationRecord {
	resolved := deferredVisibilityFoundationRecord{AffectedLane: lane, SourceIDs: map[string]bool{sourceID: true}, Reason: cell.Reason}
	if cell.Mapping != nil {
		if id, found := deferredVisibilityDirectString(cell.Mapping, "foundation_id"); found {
			resolved.ID = id
		}
		if id, found := deferredVisibilityDirectString(cell.Mapping, "foundation_gap_id"); found && resolved.ID == "" {
			resolved.ID = id
		}
		if atlas, found := deferredVisibilityNestedString(cell.Mapping, "atlas_lookup", "consulted_id"); found {
			resolved.AtlasCapability = atlas
		}
		if atlas, found := deferredVisibilityDirectString(cell.Mapping, "consulted_atlas_id"); found && resolved.AtlasCapability == "" {
			resolved.AtlasCapability = atlas
		}
	}
	for _, candidate := range ledger {
		if candidate.AffectedLane != lane || !candidate.SourceIDs[sourceID] {
			continue
		}
		if resolved.ID == "" {
			resolved.ID = candidate.ID
		}
		if resolved.AtlasCapability == "" {
			resolved.AtlasCapability = candidate.AtlasCapability
		}
		if resolved.Reason == "" {
			resolved.Reason = candidate.Reason
		}
	}
	if gap := deferredVisibilityFoundationGap(foundationAtlas, sourceID, lane); gap.ID != "" || gap.AtlasCapability != "" || gap.Reason != "" {
		if resolved.ID == "" {
			resolved.ID = gap.ID
		}
		if resolved.AtlasCapability == "" {
			resolved.AtlasCapability = gap.AtlasCapability
		}
		if resolved.Reason == "" {
			resolved.Reason = gap.Reason
		}
	}
	return resolved
}

// deferredVisibilityFoundationGap recursively searches authoring-only matrix
// foundation evidence. It requires the exact source ID and lane, so an
// unrelated provider's similar webhook or transport fact cannot satisfy a
// deferred cell.
func deferredVisibilityFoundationGap(value any, sourceID, lane string) deferredVisibilityFoundationRecord {
	var found deferredVisibilityFoundationRecord
	var visit func(any)
	visit = func(current any) {
		if found.ID != "" || found.AtlasCapability != "" || found.Reason != "" {
			return
		}
		switch typed := current.(type) {
		case map[string]any:
			if deferredVisibilityGapMatches(typed, sourceID, lane) {
				found = deferredVisibilityFoundationRecord{
					ID:              deferredVisibilityFirstString(typed, "foundation_id", "foundation_gap_id", "gap_id"),
					AtlasCapability: deferredVisibilityFirstString(typed, "atlas_capability", "consulted_atlas_id"),
					AffectedLane:    lane,
					SourceIDs:       map[string]bool{sourceID: true},
					Reason:          deferredVisibilityFirstString(typed, "missing_capability", "reason", "insufficiency"),
				}
				return
			}
			for _, child := range typed {
				visit(child)
			}
		case []any:
			for _, child := range typed {
				visit(child)
			}
		}
	}
	visit(value)
	return found
}

func deferredVisibilityGapMatches(object map[string]any, sourceID, lane string) bool {
	if foundLane, found := deferredVisibilityDirectString(object, "lane"); !found || foundLane != lane {
		return false
	}
	if id, found := deferredVisibilityDirectString(object, "source_id"); found && id == sourceID {
		return true
	}
	values, found := object["source_ids"]
	if !found {
		return false
	}
	array, ok := values.([]any)
	if !ok {
		return false
	}
	for _, raw := range array {
		switch value := raw.(type) {
		case string:
			if value == sourceID {
				return true
			}
		case map[string]any:
			if id, exists := deferredVisibilityDirectString(value, "id"); exists && id == sourceID {
				return true
			}
		}
	}
	return false
}

func deferredVisibilityFirstString(object map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, found := deferredVisibilityDirectString(object, key); found {
			return value
		}
	}
	return ""
}

func deferredVisibilityReadFoundationLedger(root, connector string) ([]deferredVisibilityFoundationRecord, error) {
	path := filepath.Join(root, "internal", "connectors", "defs", connector, deferredVisibilityMissingFoundationFile)
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s missing-foundation ledger: %w", connector, err)
	}
	var document any
	if err := decodeSourceJSON(raw, &document); err != nil {
		return nil, fmt.Errorf("decode %s missing-foundation ledger: %w", connector, err)
	}
	rootObject, err := retainedSourceMappingObject(document, "missing-foundation ledger")
	if err != nil {
		return nil, err
	}
	actualConnector, err := retainedSourceMappingString(rootObject["connector"], "missing-foundation.connector")
	if err != nil {
		return nil, err
	}
	if actualConnector != connector {
		return nil, fmt.Errorf("missing-foundation connector %q does not match %q", actualConnector, connector)
	}
	value, exists := rootObject["foundations"]
	if !exists {
		return nil, nil
	}
	items, err := retainedSourceMappingArray(value, "missing-foundation.foundations")
	if err != nil {
		return nil, err
	}
	records := make([]deferredVisibilityFoundationRecord, 0, len(items))
	for index, rawFoundation := range items {
		field := fmt.Sprintf("missing-foundation.foundations[%d]", index)
		foundation, err := retainedSourceMappingObject(rawFoundation, field)
		if err != nil {
			return nil, err
		}
		id, err := retainedSourceMappingString(foundation["id"], field+".id")
		if err != nil {
			return nil, err
		}
		sourceIDValue, hasSourceIDs := foundation["source_ids"]
		if !hasSourceIDs || sourceIDValue == nil {
			// A connector may also track a historical/implemented concern in the
			// same authoring ledger. It is not a per-source deferred row and must
			// neither be treated as one nor constrain this bridge.
			continue
		}
		lane, hasLane := deferredVisibilityDirectString(foundation, "affected_lane")
		if !hasLane {
			// A source-specific historical record without one lane is not a
			// candidate for the seven-lane bridge. If a matrix later marks that
			// source/lane missing_foundation, resolution fails closed rather than
			// guessing a lane from its prose.
			continue
		}
		if !retainedSourceMappingKnownLane(lane) {
			return nil, fmt.Errorf("%s.affected_lane %q is unknown", field, lane)
		}
		sourceIDValues, sourceIDsErr := retainedSourceMappingArray(sourceIDValue, field+".source_ids")
		if sourceIDsErr != nil {
			return nil, sourceIDsErr
		}
		sourceIDs := make(map[string]bool, len(sourceIDValues))
		for sourceIndex, rawSourceID := range sourceIDValues {
			object, objectErr := retainedSourceMappingObject(rawSourceID, fmt.Sprintf("%s.source_ids[%d]", field, sourceIndex))
			if objectErr != nil {
				return nil, objectErr
			}
			sourceID, sourceErr := retainedSourceMappingString(object["id"], fmt.Sprintf("%s.source_ids[%d].id", field, sourceIndex))
			if sourceErr != nil {
				return nil, sourceErr
			}
			if sourceIDs[sourceID] {
				return nil, fmt.Errorf("%s duplicates source ID %q", field, sourceID)
			}
			sourceIDs[sourceID] = true
		}
		atlasCapability, _ := deferredVisibilityDirectString(foundation, "atlas_capability")
		reason, _ := deferredVisibilityDirectString(foundation, "reason")
		records = append(records, deferredVisibilityFoundationRecord{ID: id, AtlasCapability: atlasCapability, AffectedLane: lane, SourceIDs: sourceIDs, Reason: reason})
	}
	return records, nil
}

func deferredVisibilityValidateReport(report deferredVisibilityReport) error {
	if report.SchemaVersion != deferredVisibilitySchemaVersion || !report.MappingOnly || report.ExecutableDeclarations != 0 {
		return fmt.Errorf("report must be mapping-only with zero executable declarations")
	}
	if report.PrimarySourceOperations == 0 || report.SourceRows < report.PrimarySourceOperations || report.MatrixCells != report.SourceRows*len(retainedSourceMappingLaneOrder) || report.DeferredCells != len(report.Entries) {
		return fmt.Errorf("report has incomplete source/deferred accounting")
	}
	seen := make(map[string]bool, len(report.Entries))
	for _, entry := range report.Entries {
		key := entry.Connector + "\x00" + entry.SourceOperationID + "\x00" + entry.Lane
		if entry.Connector == "" || entry.SourceOperationID == "" || !retainedSourceMappingKnownLane(entry.Lane) || seen[key] {
			return fmt.Errorf("report has missing, unknown, or duplicate deferred entry identity")
		}
		seen[key] = true
		if entry.Visibility != "deferred" || (entry.SourceDisposition != "mapped_unproven" && entry.SourceDisposition != "missing_foundation") || entry.RuntimeClaim != deferredVisibilityRuntimeClaim {
			return fmt.Errorf("report entry %s has an invalid deferred/runtime status", key)
		}
		if entry.Source.SourceLock == "" || entry.Source.SourceLockSHA256 == "" || entry.Source.SourceLockOperationID == "" || entry.Source.CitationURL == "" || entry.Source.SourceLocation == "" || entry.Source.Method == "" || entry.Source.Path == "" || len(entry.SourceFact) == 0 || entry.Reason.ID == "" || entry.Reason.Kind == "" || entry.Reason.Detail == "" || entry.Capability.ID == "" || entry.Capability.AtlasID == "" {
			return fmt.Errorf("report entry %s has incomplete source, reason, or capability evidence", key)
		}
		if entry.SourceDisposition == "mapped_unproven" && (entry.Reason.ID != deferredVisibilityMappedUnprovenReasonID || entry.Capability.ID != deferredVisibilityProjectionAdmissionID) {
			return fmt.Errorf("report entry %s has unstable mapped-unproven reason/capability", key)
		}
		if entry.SourceDisposition == "missing_foundation" && (entry.Reason.ID != deferredVisibilityMissingFoundationReasonID || entry.Capability.Kind != "named_foundation") {
			return fmt.Errorf("report entry %s has unstable missing-foundation reason/capability", key)
		}
	}
	return nil
}
