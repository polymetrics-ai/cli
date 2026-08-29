package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	_ "polymetrics.ai/internal/connectors/hooks/hookset"
)

const (
	batchManifestSchemaVersion = 1
	defaultBatchSize           = 5
	maxBatchSize               = 40
	defaultMinOperations       = 20
	defaultMaxOperations       = 250
	maxLedgerBytes             = 8 << 20
)

// batchLedger is the intentionally permissive decoder for the artifact survey.
// The survey is live and adds explanatory fields over time, so plan reads the
// fields it needs without rejecting unrelated forward-compatible evidence.
// Nullable counts are meaningful for an `unknown` survey record; a record must
// become fully measured only before it can become a batch candidate.
type batchLedger struct {
	SchemaVersion string              `json:"schema_version"`
	CreatedAt     string              `json:"created_at"`
	Records       []batchLedgerRecord `json:"records"`
}

type batchLedgerRecord struct {
	Connector          string `json:"connector"`
	Status             string `json:"status"`
	OperationsTotal    *int   `json:"operations_total"`
	OperationsRead     *int   `json:"operations_read"`
	OperationsWrite    *int   `json:"operations_write"`
	ArtifactURL        string `json:"artifact_url"`
	ArtifactKind       string `json:"artifact_kind"`
	ArtifactVersion    string `json:"artifact_version"`
	RetrievedAt        string `json:"retrieved_at"`
	AuthModel          string `json:"auth_model"`
	AccessModel        string `json:"access_model"`
	EvidenceSource     string `json:"evidence_source"`
	CountingNote       string `json:"counting_note"`
	ProcessedAt        string `json:"processed_at"`
	ScopeInCurrentDefs bool   `json:"scope_in_current_defs"`
}

// BatchManifest is the checked-in, immutable batch input. It captures the
// provider evidence used to decide a candidate before authoring mutates any
// bundle. No field is inferred from a connector name or operation count.
type BatchManifest struct {
	SchemaVersion int                      `json:"schema_version"`
	SourceLedger  BatchLedgerSource        `json:"source_ledger"`
	Selection     BatchSelection           `json:"selection"`
	Connectors    []BatchManifestConnector `json:"connectors"`
}

type BatchLedgerSource struct {
	SchemaVersion string `json:"schema_version"`
	CreatedAt     string `json:"created_at"`
	RecordCount   int    `json:"record_count"`
}

type BatchSelection struct {
	Mode          string `json:"mode"`
	RequestedSize int    `json:"requested_size"`
	MinOperations int    `json:"min_operations"`
	MaxOperations int    `json:"max_operations"`
	Criteria      string `json:"criteria"`
}

type BatchManifestConnector struct {
	Connector       string        `json:"connector"`
	OperationsTotal int           `json:"operations_total"`
	OperationsRead  int           `json:"operations_read"`
	OperationsWrite int           `json:"operations_write"`
	Artifact        BatchArtifact `json:"artifact"`
	AuthModel       string        `json:"auth_model"`
	AccessModel     string        `json:"access_model"`
	EvidenceSource  string        `json:"evidence_source"`
	CountingNote    string        `json:"counting_note"`
	ProcessedAt     string        `json:"processed_at"`
	SelectionReason string        `json:"selection_reason"`
}

type BatchArtifact struct {
	URL         string `json:"url"`
	Kind        string `json:"kind"`
	Version     string `json:"version"`
	RetrievedAt string `json:"retrieved_at"`
}

type batchPlanOptions struct {
	ledgerPath    string
	outPath       string
	size          int
	minOperations int
	maxOperations int
	connectors    []string
}

// runBatch dispatches the batch control-plane commands. These commands live
// alongside validate and surface-sync because they orchestrate those existing
// connector guarantees; they do not create a parallel generator or runtime
// rule set.
func runBatch(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		logln(stdout, batchUsage())
		return 0
	}

	switch args[1] {
	case "plan":
		return runBatchPlan(args[2:], stdout, stderr)
	case "materialize":
		return runBatchMaterialize(args[2:], stdout, stderr)
	case "gate":
		return runBatchGate(args[2:], stdout, stderr)
	case "-h", "--help", "help":
		logln(stdout, batchUsage())
		return 0
	default:
		logf(stderr, "connectorgen batch: unknown subcommand %q\n%s\n", args[1], batchUsage())
		return 2
	}
}

func batchUsage() string {
	return `usage:
  connectorgen batch plan --ledger <path> --out <path> [--size <1-40>] [--connector <name>] [--min-operations <n>] [--max-operations <n>]
  connectorgen batch materialize --manifest <path> --source-defs-root <path> --retrieved-at <YYYY-MM-DD> --report <path> [--defs-root <path>] [--artifact-dir <path>] [--connector <name>]
  connectorgen batch gate --manifest <path> --report <path> [--defs-root <path>] [--connector <name>]`
}

// runBatchPlan validates a live survey snapshot and writes the deterministic
// evidence manifest that a later authoring/gate run consumes. It intentionally
// does not fetch an artifact or change a connector directory: those actions
// require the merged v2 provenance and non-redacting output-policy contracts.
func runBatchPlan(args []string, stdout, stderr io.Writer) int {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			logln(stdout, batchUsage())
			return 0
		}
	}
	opts, err := parseBatchPlanOptions(args)
	if err != nil {
		logf(stderr, "connectorgen batch plan: %v\n%s\n", err, batchUsage())
		return 2
	}

	ledger, err := readBatchLedger(opts.ledgerPath)
	if err != nil {
		logf(stderr, "connectorgen batch plan: %v\n", err)
		return 1
	}
	selected, mode, err := selectBatchCandidates(ledger, opts)
	if err != nil {
		logf(stderr, "connectorgen batch plan: %v\n", err)
		return 1
	}

	manifest := newBatchManifest(ledger, opts, mode, selected)
	if err := writeBatchManifest(opts.outPath, manifest); err != nil {
		logf(stderr, "connectorgen batch plan: write manifest: %v\n", err)
		return 1
	}

	total := 0
	for _, record := range selected {
		total += *record.OperationsTotal
	}
	logf(stdout, "connectorgen batch plan: %d connector(s) selected, %d surveyed operation(s), manifest %s\n", len(selected), total, opts.outPath)
	return 0
}

func parseBatchPlanOptions(args []string) (batchPlanOptions, error) {
	opts := batchPlanOptions{
		size:          defaultBatchSize,
		minOperations: defaultMinOperations,
		maxOperations: defaultMaxOperations,
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		next := func() (string, error) {
			if i+1 >= len(args) {
				return "", fmt.Errorf("%s requires a value", arg)
			}
			i++
			return args[i], nil
		}
		switch arg {
		case "--ledger":
			value, err := next()
			if err != nil {
				return batchPlanOptions{}, err
			}
			opts.ledgerPath = value
		case "--out":
			value, err := next()
			if err != nil {
				return batchPlanOptions{}, err
			}
			opts.outPath = value
		case "--size":
			value, err := next()
			if err != nil {
				return batchPlanOptions{}, err
			}
			parsed, err := parseBoundedBatchInteger("--size", value, 1, maxBatchSize)
			if err != nil {
				return batchPlanOptions{}, err
			}
			opts.size = parsed
		case "--min-operations":
			value, err := next()
			if err != nil {
				return batchPlanOptions{}, err
			}
			parsed, err := parseBoundedBatchInteger("--min-operations", value, 0, int(^uint(0)>>1))
			if err != nil {
				return batchPlanOptions{}, err
			}
			opts.minOperations = parsed
		case "--max-operations":
			value, err := next()
			if err != nil {
				return batchPlanOptions{}, err
			}
			parsed, err := parseBoundedBatchInteger("--max-operations", value, 0, int(^uint(0)>>1))
			if err != nil {
				return batchPlanOptions{}, err
			}
			opts.maxOperations = parsed
		case "--connector":
			value, err := next()
			if err != nil {
				return batchPlanOptions{}, err
			}
			opts.connectors = append(opts.connectors, value)
		default:
			return batchPlanOptions{}, fmt.Errorf("unknown flag %q", arg)
		}
	}
	if opts.ledgerPath == "" {
		return batchPlanOptions{}, errors.New("--ledger is required")
	}
	if opts.outPath == "" {
		return batchPlanOptions{}, errors.New("--out is required")
	}
	if opts.minOperations > opts.maxOperations {
		return batchPlanOptions{}, fmt.Errorf("--min-operations %d exceeds --max-operations %d", opts.minOperations, opts.maxOperations)
	}
	if len(opts.connectors) > opts.size {
		return batchPlanOptions{}, fmt.Errorf("%d --connector values exceed --size %d", len(opts.connectors), opts.size)
	}
	return opts, nil
}

func parseBoundedBatchInteger(flag, raw string, min, max int) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("%s must be an integer from %d to %d", flag, min, max)
	}
	return value, nil
}

func readBatchLedger(path string) (batchLedger, error) {
	info, err := os.Stat(path)
	if err != nil {
		return batchLedger{}, fmt.Errorf("read ledger: %w", err)
	}
	if !info.Mode().IsRegular() {
		return batchLedger{}, fmt.Errorf("read ledger: %s is not a regular file", path)
	}
	if info.Size() > maxLedgerBytes {
		return batchLedger{}, fmt.Errorf("read ledger: %s exceeds %d-byte limit", path, maxLedgerBytes)
	}

	file, err := os.Open(path)
	if err != nil {
		return batchLedger{}, fmt.Errorf("open ledger: %w", err)
	}
	defer func() { _ = file.Close() }()

	var ledger batchLedger
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&ledger); err != nil {
		return batchLedger{}, fmt.Errorf("decode ledger: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return batchLedger{}, errors.New("decode ledger: multiple JSON values")
		}
		return batchLedger{}, fmt.Errorf("decode ledger trailing data: %w", err)
	}
	if strings.TrimSpace(ledger.SchemaVersion) == "" {
		return batchLedger{}, errors.New("ledger schema_version is required")
	}
	if strings.TrimSpace(ledger.CreatedAt) == "" {
		return batchLedger{}, errors.New("ledger created_at is required")
	}
	if len(ledger.Records) == 0 {
		return batchLedger{}, errors.New("ledger records is empty")
	}

	seen := make(map[string]bool, len(ledger.Records))
	for i, record := range ledger.Records {
		if err := validateBatchConnectorName(record.Connector); err != nil {
			return batchLedger{}, fmt.Errorf("ledger record %d: %w", i, err)
		}
		if seen[record.Connector] {
			return batchLedger{}, fmt.Errorf("ledger contains duplicate connector %q", record.Connector)
		}
		seen[record.Connector] = true
	}
	return ledger, nil
}

func validateBatchConnectorName(name string) error {
	if name == "" {
		return errors.New("connector is required")
	}
	if name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00\r\n") {
		return fmt.Errorf("connector %q is not a safe bundle name", name)
	}
	return nil
}

func selectBatchCandidates(ledger batchLedger, opts batchPlanOptions) ([]batchLedgerRecord, string, error) {
	byName := make(map[string]batchLedgerRecord, len(ledger.Records))
	for _, record := range ledger.Records {
		byName[record.Connector] = record
	}

	if len(opts.connectors) > 0 {
		seen := make(map[string]bool, len(opts.connectors))
		selected := make([]batchLedgerRecord, 0, len(opts.connectors))
		for _, name := range opts.connectors {
			if seen[name] {
				return nil, "", fmt.Errorf("connector %q was selected more than once", name)
			}
			seen[name] = true
			record, ok := byName[name]
			if !ok {
				return nil, "", fmt.Errorf("selected connector %q is absent from the ledger", name)
			}
			if err := validateBatchCandidate(record, opts); err != nil {
				return nil, "", fmt.Errorf("selected connector %q is ineligible: %w", name, err)
			}
			selected = append(selected, record)
		}
		sortBatchRecords(selected)
		return selected, "explicit", nil
	}

	eligible := make([]batchLedgerRecord, 0, len(ledger.Records))
	for _, record := range ledger.Records {
		if err := validateBatchCandidate(record, opts); err == nil {
			eligible = append(eligible, record)
		}
	}
	sortBatchRecords(eligible)
	if len(eligible) < opts.size {
		return nil, "", fmt.Errorf("only %d ledger records meet the machine-readable, versioned, public, %d-%d-operation criteria; need %d", len(eligible), opts.minOperations, opts.maxOperations, opts.size)
	}
	return eligible[:opts.size], "criteria", nil
}

func validateBatchCandidate(record batchLedgerRecord, opts batchPlanOptions) error {
	if record.Status != "done" {
		return fmt.Errorf("status is %q, want done", record.Status)
	}
	if record.ArtifactKind != "openapi" && record.ArtifactKind != "swagger" {
		return fmt.Errorf("artifact_kind is %q, want openapi or swagger", record.ArtifactKind)
	}
	if strings.TrimSpace(record.ArtifactVersion) == "" {
		return errors.New("artifact_version is required")
	}
	if strings.TrimSpace(record.RetrievedAt) == "" {
		return errors.New("retrieved_at is required")
	}
	if _, err := time.Parse("2006-01-02", record.RetrievedAt); err != nil {
		return fmt.Errorf("retrieved_at %q must be an ISO full date: %w", record.RetrievedAt, err)
	}
	if err := validateBatchArtifactURL(record.ArtifactURL); err != nil {
		return err
	}
	if strings.TrimSpace(record.AuthModel) == "" {
		return errors.New("auth_model is required")
	}
	if record.AccessModel != "public" {
		return fmt.Errorf("access_model is %q, want public", record.AccessModel)
	}
	if strings.TrimSpace(record.EvidenceSource) == "" {
		return errors.New("evidence_source is required")
	}
	if strings.TrimSpace(record.CountingNote) == "" {
		return errors.New("counting_note is required")
	}
	if strings.TrimSpace(record.ProcessedAt) == "" {
		return errors.New("processed_at is required")
	}
	if !record.ScopeInCurrentDefs {
		return errors.New("scope_in_current_defs is false")
	}
	if record.OperationsTotal == nil || record.OperationsRead == nil || record.OperationsWrite == nil {
		return errors.New("operations_total, operations_read, and operations_write must be measured")
	}
	if *record.OperationsTotal < opts.minOperations || *record.OperationsTotal > opts.maxOperations {
		return fmt.Errorf("operations_total %d is outside requested %d-%d range", *record.OperationsTotal, opts.minOperations, opts.maxOperations)
	}
	if *record.OperationsRead < 0 || *record.OperationsWrite < 0 || *record.OperationsTotal < 0 {
		return errors.New("operation counts must be non-negative")
	}
	return nil
}

func validateBatchArtifactURL(raw string) error {
	_, err := parseBatchArtifactURL(raw)
	return err
}

func sortBatchRecords(records []batchLedgerRecord) {
	sort.Slice(records, func(i, j int) bool {
		if *records[i].OperationsTotal != *records[j].OperationsTotal {
			return *records[i].OperationsTotal < *records[j].OperationsTotal
		}
		return records[i].Connector < records[j].Connector
	})
}

func newBatchManifest(ledger batchLedger, opts batchPlanOptions, mode string, records []batchLedgerRecord) BatchManifest {
	connectors := make([]BatchManifestConnector, 0, len(records))
	for _, record := range records {
		reason := "selected by machine-readable evidence: done ledger record, public versioned OpenAPI/Swagger artifact, recorded retrieval date, and bounded measured operation count"
		if mode == "explicit" {
			reason = "selected explicitly after provider-artifact quality review; the record independently satisfies the machine-readable, versioned, public, bounded-operation evidence gate"
		}
		connectors = append(connectors, BatchManifestConnector{
			Connector:       record.Connector,
			OperationsTotal: *record.OperationsTotal,
			OperationsRead:  *record.OperationsRead,
			OperationsWrite: *record.OperationsWrite,
			Artifact: BatchArtifact{
				URL:         record.ArtifactURL,
				Kind:        record.ArtifactKind,
				Version:     record.ArtifactVersion,
				RetrievedAt: record.RetrievedAt,
			},
			AuthModel:       record.AuthModel,
			AccessModel:     record.AccessModel,
			EvidenceSource:  record.EvidenceSource,
			CountingNote:    record.CountingNote,
			ProcessedAt:     record.ProcessedAt,
			SelectionReason: reason,
		})
	}
	return BatchManifest{
		SchemaVersion: batchManifestSchemaVersion,
		SourceLedger: BatchLedgerSource{
			SchemaVersion: ledger.SchemaVersion,
			CreatedAt:     ledger.CreatedAt,
			RecordCount:   len(ledger.Records),
		},
		Selection: BatchSelection{
			Mode:          mode,
			RequestedSize: opts.size,
			MinOperations: opts.minOperations,
			MaxOperations: opts.maxOperations,
			Criteria:      "status=done; artifact_kind=openapi|swagger; non-empty version; absolute HTTPS artifact URL; ISO full-date retrieved_at; public access; measured counts; in current defs scope",
		},
		Connectors: connectors,
	}
}

func writeBatchManifest(path string, manifest BatchManifest) error {
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return writeBatchFile(path, append(raw, '\n'))
}

func writeBatchFile(path string, raw []byte) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, ".connectorgen-batch-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if _, err := temp.Write(raw); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, absPath)
}

const batchGateReportSchemaVersion = 1

// BatchGateReport is the per-batch delivery record. A drop is data, not an
// early return: every candidate is inspected and its failure stage/reason is
// retained so the branch can remove only that bundle and continue with the
// remaining connectors.
type BatchGateReport struct {
	SchemaVersion      int                  `json:"schema_version"`
	Manifest           string               `json:"manifest"`
	Candidates         int                  `json:"candidates"`
	Included           []BatchGateIncluded  `json:"included"`
	Dropped            []BatchGateDrop      `json:"dropped"`
	ProvenanceRefusals int                  `json:"provenance_refusals"`
	SurveyedOperations BatchOperationTotals `json:"surveyed_operations"`
	DeclaredOperations int                  `json:"declared_operations"`
	OperationSplit     BatchOperationSplit  `json:"operation_split"`
}

// BatchOperationTotals is copied exactly from the manifest's ledger evidence.
// It is intentionally separate from DeclaredOperations, which comes from the
// finished bundle's api_surface.json and must never be estimated from it.
type BatchOperationTotals struct {
	Total int `json:"total"`
	Read  int `json:"read"`
	Write int `json:"write"`
}

// BatchOperationSplit accounts for every declared api-surface row. Executable
// means an endpoint has a real covered_by target; blocked means the provider
// operation remains blocked by default; excluded is a legacy exclusion or a
// blocked operation that the batch treats as an exclusion.
type BatchOperationSplit struct {
	Executable      int `json:"executable"`
	ProviderBlocked int `json:"provider_blocked"`
	Excluded        int `json:"excluded"`
}

func (s *BatchOperationSplit) add(other BatchOperationSplit) {
	s.Executable += other.Executable
	s.ProviderBlocked += other.ProviderBlocked
	s.Excluded += other.Excluded
}

func (s BatchOperationSplit) total() int {
	return s.Executable + s.ProviderBlocked + s.Excluded
}

type BatchGateIncluded struct {
	Connector          string              `json:"connector"`
	CommandsChecked    int                 `json:"commands_checked"`
	DeclaredOperations int                 `json:"declared_operations"`
	OperationSplit     BatchOperationSplit `json:"operation_split"`
	Warnings           []Finding           `json:"warnings"`
}

type BatchGateDrop struct {
	Connector string `json:"connector"`
	Stage     string `json:"stage"`
	Reason    string `json:"reason"`
}

type batchGateOptions struct {
	manifestPath string
	defsRoot     string
	reportPath   string
	connectors   []string
}

// runBatchGate applies the existing per-bundle guarantees one candidate at a
// time. It deliberately calls commandrunner.Preflight rather than reproducing
// its implementation rules, so a new executor is covered as soon as the
// runtime learns it.
func runBatchGate(args []string, stdout, stderr io.Writer) int {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			logln(stdout, batchUsage())
			return 0
		}
	}
	opts, err := parseBatchGateOptions(args)
	if err != nil {
		logf(stderr, "connectorgen batch gate: %v\n%s\n", err, batchUsage())
		return 2
	}
	manifest, err := readBatchManifest(opts.manifestPath)
	if err != nil {
		logf(stderr, "connectorgen batch gate: %v\n", err)
		return 1
	}
	candidates, err := selectedManifestCandidates(manifest, opts.connectors)
	if err != nil {
		logf(stderr, "connectorgen batch gate: %v\n", err)
		return 1
	}

	report := BatchGateReport{
		SchemaVersion: batchGateReportSchemaVersion,
		Manifest:      opts.manifestPath,
		Candidates:    len(candidates),
		Included:      []BatchGateIncluded{},
		Dropped:       []BatchGateDrop{},
	}
	for _, candidate := range candidates {
		report.SurveyedOperations.Total += candidate.OperationsTotal
		report.SurveyedOperations.Read += candidate.OperationsRead
		report.SurveyedOperations.Write += candidate.OperationsWrite

		included, drop := gateBatchConnector(opts.defsRoot, candidate)
		if drop != nil {
			report.Dropped = append(report.Dropped, *drop)
			if drop.Stage == "provenance" {
				report.ProvenanceRefusals++
			}
			continue
		}
		report.Included = append(report.Included, included)
		report.DeclaredOperations += included.DeclaredOperations
		report.OperationSplit.add(included.OperationSplit)
	}
	if err := writeBatchGateReport(opts.reportPath, report); err != nil {
		logf(stderr, "connectorgen batch gate: write report: %v\n", err)
		return 1
	}

	logf(stdout, "connectorgen batch gate: %d connector(s) included, %d dropped (%d provenance refusal(s)), %d declared operation(s); report %s\n",
		len(report.Included), len(report.Dropped), report.ProvenanceRefusals, report.DeclaredOperations, opts.reportPath)
	if len(report.Dropped) > 0 {
		return 1
	}
	return 0
}

func parseBatchGateOptions(args []string) (batchGateOptions, error) {
	opts := batchGateOptions{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if i+1 >= len(args) {
			return batchGateOptions{}, fmt.Errorf("%s requires a value", arg)
		}
		value := args[i+1]
		i++
		switch arg {
		case "--manifest":
			opts.manifestPath = value
		case "--defs-root":
			opts.defsRoot = value
		case "--report":
			opts.reportPath = value
		case "--connector":
			opts.connectors = append(opts.connectors, value)
		default:
			return batchGateOptions{}, fmt.Errorf("unknown flag %q", arg)
		}
	}
	if opts.manifestPath == "" {
		return batchGateOptions{}, errors.New("--manifest is required")
	}
	if opts.reportPath == "" {
		return batchGateOptions{}, errors.New("--report is required")
	}
	if opts.defsRoot == "" {
		root, err := repoRoot()
		if err != nil {
			return batchGateOptions{}, fmt.Errorf("default defs root: %w", err)
		}
		opts.defsRoot = filepath.Join(root, "internal", "connectors", "defs")
	}
	return opts, nil
}

func readBatchManifest(path string) (BatchManifest, error) {
	info, err := os.Stat(path)
	if err != nil {
		return BatchManifest{}, fmt.Errorf("read manifest: %w", err)
	}
	if !info.Mode().IsRegular() {
		return BatchManifest{}, fmt.Errorf("read manifest: %s is not a regular file", path)
	}
	if info.Size() > maxLedgerBytes {
		return BatchManifest{}, fmt.Errorf("read manifest: %s exceeds %d-byte limit", path, maxLedgerBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return BatchManifest{}, fmt.Errorf("open manifest: %w", err)
	}
	defer func() { _ = file.Close() }()

	var manifest BatchManifest
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return BatchManifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return BatchManifest{}, errors.New("decode manifest: multiple JSON values")
		}
		return BatchManifest{}, fmt.Errorf("decode manifest trailing data: %w", err)
	}
	if err := validateBatchManifest(manifest); err != nil {
		return BatchManifest{}, err
	}
	return manifest, nil
}

func validateBatchManifest(manifest BatchManifest) error {
	if manifest.SchemaVersion != batchManifestSchemaVersion {
		return fmt.Errorf("manifest schema_version %d is unsupported", manifest.SchemaVersion)
	}
	if strings.TrimSpace(manifest.SourceLedger.SchemaVersion) == "" || strings.TrimSpace(manifest.SourceLedger.CreatedAt) == "" {
		return errors.New("manifest source_ledger schema_version and created_at are required")
	}
	if len(manifest.Connectors) == 0 {
		return errors.New("manifest connectors is empty")
	}
	seen := make(map[string]bool, len(manifest.Connectors))
	for i, candidate := range manifest.Connectors {
		if err := validateBatchConnectorName(candidate.Connector); err != nil {
			return fmt.Errorf("manifest connector %d: %w", i, err)
		}
		if seen[candidate.Connector] {
			return fmt.Errorf("manifest contains duplicate connector %q", candidate.Connector)
		}
		seen[candidate.Connector] = true
		if candidate.OperationsTotal < 0 || candidate.OperationsRead < 0 || candidate.OperationsWrite < 0 {
			return fmt.Errorf("manifest connector %q has a negative operation count", candidate.Connector)
		}
		if err := validateBatchArtifactURL(candidate.Artifact.URL); err != nil {
			return fmt.Errorf("manifest connector %q: %w", candidate.Connector, err)
		}
		if candidate.Artifact.Kind != "openapi" && candidate.Artifact.Kind != "swagger" {
			return fmt.Errorf("manifest connector %q: artifact.kind %q is not openapi or swagger", candidate.Connector, candidate.Artifact.Kind)
		}
		if strings.TrimSpace(candidate.Artifact.Version) == "" {
			return fmt.Errorf("manifest connector %q: artifact.version is required", candidate.Connector)
		}
		if _, err := time.Parse("2006-01-02", candidate.Artifact.RetrievedAt); err != nil {
			return fmt.Errorf("manifest connector %q: artifact.retrieved_at must be an ISO full date: %w", candidate.Connector, err)
		}
		if strings.TrimSpace(candidate.AuthModel) == "" || candidate.AccessModel != "public" || strings.TrimSpace(candidate.EvidenceSource) == "" || strings.TrimSpace(candidate.CountingNote) == "" || strings.TrimSpace(candidate.ProcessedAt) == "" || strings.TrimSpace(candidate.SelectionReason) == "" {
			return fmt.Errorf("manifest connector %q is missing required evidence", candidate.Connector)
		}
	}
	return nil
}

func gateBatchConnector(defsRoot string, candidate BatchManifestConnector) (BatchGateIncluded, *BatchGateDrop) {
	bundleDir, err := batchBundleDirectory(defsRoot, candidate.Connector)
	if err != nil {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "bundle_path", err)
	}
	info, err := os.Stat(bundleDir)
	if err != nil {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "bundle_path", fmt.Errorf("bundle directory: %w", err))
	}
	if !info.IsDir() {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "bundle_path", errors.New("bundle path is not a directory"))
	}
	// validatePath treats a directory lacking metadata.json as a defs root. A
	// batch candidate is specifically one bundle, so make that malformed-bundle
	// case an explicit validation drop rather than accidentally scanning it as
	// an empty root.
	if !isBundleDir(bundleDir) {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "validate", errors.New("metadata.json is required for a batch bundle"))
	}
	bundle, err := engine.Load(os.DirFS(filepath.Dir(bundleDir)), filepath.Base(bundleDir))
	if err != nil {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "load", err)
	}
	if err := batchCandidateProvenance(bundle.Surface, candidate); err != nil {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "provenance", err)
	}

	validation, err := validatePath(bundleDir)
	if err != nil {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "validate", err)
	}
	if len(validation.Findings) > 0 {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "validate", batchFindingSummary(validation.Findings))
	}

	syncStats, err := syncBundle(bundleDir, true)
	if err != nil {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "surface_sync", err)
	}
	if syncStats.total() > 0 {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "surface_sync", fmt.Errorf("derived command metadata drift: filled %s; corrected %s", syncStats.Filled, syncStats.Corrected))
	}

	if bundle.CLISurface == nil || len(bundle.CLISurface.Commands) == 0 {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "cli_surface", errors.New("cli_surface.json with at least one reachable command is required"))
	}
	if err := batchNoRedactionDeclarations(bundle); err != nil {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "output_policy", err)
	}
	split, err := batchSurfaceSplit(bundle.Surface)
	if err != nil {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "api_surface", err)
	}
	checked, err := batchRuntimePreflight(bundle)
	if err != nil {
		return BatchGateIncluded{}, batchGateDrop(candidate.Connector, "runtime_preflight", err)
	}
	return BatchGateIncluded{
		Connector:          candidate.Connector,
		CommandsChecked:    checked,
		DeclaredOperations: split.total(),
		OperationSplit:     split,
		Warnings:           nonNilFindings(validation.Warnings),
	}, nil
}

func batchCandidateProvenance(surface *engine.APISurface, candidate BatchManifestConnector) error {
	if surface == nil {
		return errors.New("api_surface.json with complete v2 provenance is required for a batch candidate")
	}
	provenance := engine.ValidateSurfaceProvenance(surface)
	switch provenance.Status {
	case engine.SurfaceProvenanceLegacyUnverified:
		return fmt.Errorf("api_surface.json has legacy v%d provenance; batch candidates require complete v2 provenance matched to manifest artifact URL %q", provenance.LedgerVersion, candidate.Artifact.URL)
	case engine.SurfaceProvenanceInvalid:
		return fmt.Errorf("api_surface.json has incomplete v2 provenance: %s", batchProvenanceIssueSummary(provenance.Issues))
	case engine.SurfaceProvenanceComplete:
	default:
		return fmt.Errorf("api_surface.json provenance status %q is not complete", provenance.Status)
	}
	if len(surface.Artifacts) != 1 {
		return fmt.Errorf("api_surface.json has %d provenance artifacts; batch candidates require exactly one manifest artifact URL %q", len(surface.Artifacts), candidate.Artifact.URL)
	}
	artifact := surface.Artifacts[0]
	if artifact.URL != candidate.Artifact.URL {
		return fmt.Errorf("provenance artifact URL %q does not match manifest artifact URL %q", artifact.URL, candidate.Artifact.URL)
	}
	for i, endpoint := range surface.Endpoints {
		if endpoint.Provenance == nil {
			return fmt.Errorf("endpoint %d (%s %s) provenance is required", i, endpoint.Method, endpoint.Path)
		}
		if endpoint.Provenance.Artifact != artifact.ID {
			return fmt.Errorf("endpoint %d (%s %s) provenance artifact %q does not match manifest artifact %q", i, endpoint.Method, endpoint.Path, endpoint.Provenance.Artifact, artifact.ID)
		}
		if endpoint.Provenance.SourceURL != candidate.Artifact.URL {
			return fmt.Errorf("endpoint %d (%s %s) provenance source_url %q does not match manifest artifact URL %q", i, endpoint.Method, endpoint.Path, endpoint.Provenance.SourceURL, candidate.Artifact.URL)
		}
	}
	return nil
}

func batchProvenanceIssueSummary(issues []engine.SurfaceProvenanceIssue) string {
	if len(issues) == 0 {
		return "no validation details"
	}
	if len(issues) == 1 {
		return issues[0].Error()
	}
	return fmt.Sprintf("%s (and %d more issue(s))", issues[0], len(issues)-1)
}

func batchBundleDirectory(defsRoot, connector string) (string, error) {
	if err := validateBatchConnectorName(connector); err != nil {
		return "", err
	}
	root, err := filepath.Abs(defsRoot)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, connector)
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("bundle path escapes defs root")
	}
	return dir, nil
}

// batchNoRedactionDeclarations is a batch-specific safety policy, not a
// second runtime validator. The batch lane must not mint a declaration that
// strips provider output. Runtime support remains the source of truth for
// executability and is checked separately through commandrunner.Preflight.
func batchNoRedactionDeclarations(bundle engine.Bundle) error {
	for _, action := range bundle.Writes {
		if len(action.RedactFields) > 0 {
			return fmt.Errorf("write action %q declares redact_fields", action.Name)
		}
	}
	for _, operation := range bundle.Operations {
		if batchOutputPolicyRedacts(operation.OutputPolicy) {
			return fmt.Errorf("operation %q declares redacting output_policy %q", operation.ID, operation.OutputPolicy)
		}
		if operation.SensitivePolicy != nil && len(operation.SensitivePolicy.RedactFields) > 0 {
			return fmt.Errorf("operation %q declares redact_fields", operation.ID)
		}
	}
	for _, command := range bundle.CLISurface.Commands {
		if batchOutputPolicyRedacts(command.OutputPolicy) {
			return fmt.Errorf("command %q declares redacting output_policy %q", command.Path, command.OutputPolicy)
		}
		if len(command.RedactFields) > 0 {
			return fmt.Errorf("command %q declares redact_fields", command.Path)
		}
	}
	return nil
}

func batchOutputPolicyRedacts(policy string) bool {
	switch strings.TrimSpace(policy) {
	case "repository_contents_file_metadata", "repository_contents_directory":
		// These two legacy policies redact repository-content responses despite
		// their names not containing the word "redacted".
		return true
	}
	return strings.Contains(strings.ToLower(policy), "redact")
}

// batchRuntimePreflight uses the production command runner entry point. It
// must not gain parallel shape checks: validatePath handles static reporting,
// while this function proves the runtime agrees with every executable claim.
func batchRuntimePreflight(bundle engine.Bundle) (int, error) {
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	checked := 0
	for _, cmd := range bundle.CLISurface.Commands {
		if cmd.Availability != "implemented" {
			continue
		}
		if err := commandrunner.Preflight(connector, strings.Fields(cmd.Path)); err != nil {
			return checked, fmt.Errorf("command %q: %w", cmd.Path, err)
		}
		checked++
	}
	if checked == 0 {
		return 0, errors.New("cli_surface.json has no implemented command reachable by runtime preflight")
	}
	return checked, nil
}

func batchSurfaceSplit(surface *engine.APISurface) (BatchOperationSplit, error) {
	if surface == nil {
		return BatchOperationSplit{}, errors.New("api_surface.json is required")
	}
	split := BatchOperationSplit{}
	for i, endpoint := range surface.Endpoints {
		if strings.TrimSpace(endpoint.Method) == "" || strings.TrimSpace(endpoint.Path) == "" {
			return BatchOperationSplit{}, fmt.Errorf("endpoint %d lacks method or path", i)
		}
		if batchProtocolMetadataMethodVariant(endpoint.Method) {
			if !batchProtocolMetadataMethod(endpoint.Method) {
				return BatchOperationSplit{}, fmt.Errorf("endpoint %d (%s %s) protocol-metadata method must use exact canonical %s spelling", i, endpoint.Method, endpoint.Path, strings.ToUpper(strings.TrimSpace(endpoint.Method)))
			}
			if err := batchProtocolMetadataExclusion(endpoint); err != nil {
				return BatchOperationSplit{}, fmt.Errorf("endpoint %d (%s %s) protocol-metadata exclusion: %w", i, endpoint.Method, endpoint.Path, err)
			}
			split.Excluded++
			continue
		}
		switch {
		case endpoint.CoveredBy != nil:
			if endpoint.CoveredBy.Stream == "" && len(endpoint.CoveredBy.WriteTargets()) == 0 && endpoint.CoveredBy.DirectRead == "" && len(endpoint.CoveredBy.DirectReads) == 0 {
				return BatchOperationSplit{}, fmt.Errorf("endpoint %d (%s %s) has an empty covered_by classifier", i, endpoint.Method, endpoint.Path)
			}
			split.Executable++
		case endpoint.Excluded != nil:
			if strings.TrimSpace(endpoint.Excluded.Reason) == "" {
				return BatchOperationSplit{}, fmt.Errorf("endpoint %d (%s %s) exclusion lacks a reason", i, endpoint.Method, endpoint.Path)
			}
			split.Excluded++
		case endpoint.Operation != nil:
			if strings.TrimSpace(endpoint.Operation.Reason) == "" {
				return BatchOperationSplit{}, fmt.Errorf("endpoint %d (%s %s) blocked operation lacks a reason", i, endpoint.Method, endpoint.Path)
			}
			if batchOperationIsExcluded(endpoint.Operation.Model) {
				split.Excluded++
			} else {
				split.ProviderBlocked++
			}
		default:
			return BatchOperationSplit{}, fmt.Errorf("endpoint %d (%s %s) has no executable, blocked, or excluded classifier", i, endpoint.Method, endpoint.Path)
		}
	}
	return split, nil
}

func batchProtocolMetadataMethod(method string) bool {
	switch method {
	case http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func batchProtocolMetadataMethodVariant(method string) bool {
	method = strings.TrimSpace(method)
	return strings.EqualFold(method, http.MethodOptions) || strings.EqualFold(method, http.MethodTrace)
}

func batchProtocolMetadataOperation(method string) *engine.SurfaceOperation {
	if !batchProtocolMetadataMethod(method) {
		return nil
	}
	return &engine.SurfaceOperation{
		Model:            "local_workflow",
		Status:           "blocked",
		Risk:             "low",
		BlockedByDefault: true,
		Reason:           batchProtocolMetadataReason(method),
	}
}

func batchProtocolMetadataReason(method string) string {
	return fmt.Sprintf("The documented %s operation is protocol metadata, not a record-bearing read or state-changing provider mutation.", method)
}

func batchProtocolMetadataExclusion(endpoint engine.SurfaceEndpoint) error {
	expected := batchProtocolMetadataOperation(endpoint.Method)
	if expected == nil {
		return errors.New("is not a protocol-metadata operation")
	}
	if endpoint.CoveredBy != nil {
		return errors.New("must not use covered_by")
	}
	if endpoint.Excluded != nil {
		return errors.New("must use a v2 protocol-metadata operation")
	}
	if endpoint.Operation == nil {
		return errors.New("must use a method-specific protocol-metadata operation")
	}
	operation := endpoint.Operation
	if operation.Model != expected.Model || operation.Status != expected.Status || operation.Risk != expected.Risk || operation.BlockedByDefault != expected.BlockedByDefault || operation.Reason != expected.Reason {
		return fmt.Errorf("must use the method-specific protocol-metadata operation %q", expected.Reason)
	}
	return nil
}

func batchOperationIsExcluded(model string) bool {
	switch model {
	case "duplicate", "disallowed":
		return true
	default:
		return false
	}
}

func batchFindingSummary(findings []Finding) error {
	if len(findings) == 0 {
		return errors.New("validation failed without a finding")
	}
	first := findings[0]
	if len(findings) == 1 {
		return fmt.Errorf("%s: [%s] %s", first.File, first.Rule, first.Message)
	}
	return fmt.Errorf("%s: [%s] %s (and %d more finding(s))", first.File, first.Rule, first.Message, len(findings)-1)
}

func batchGateDrop(connector, stage string, err error) *BatchGateDrop {
	return &BatchGateDrop{Connector: connector, Stage: stage, Reason: err.Error()}
}

func nonNilFindings(findings []Finding) []Finding {
	if findings == nil {
		return []Finding{}
	}
	return findings
}

func writeBatchGateReport(path string, report BatchGateReport) error {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return writeBatchFile(path, append(raw, '\n'))
}
