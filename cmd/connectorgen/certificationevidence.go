package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/certify"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// runCertificationEvidence imports completed certification evidence. Database
// transport receipts and external HTTP proofs share the same definition-owned
// provider lookup and accepted-evidence validation path.
func runCertificationEvidence(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		logln(stderr, "connectorgen certification-evidence: require transport, change-capture, or report")
		return 2
	}
	switch args[1] {
	case "transport":
		return runTransportCertificationEvidence(args[2:], stdout, stderr)
	case "change-capture":
		return runChangeCaptureCertificationEvidence(args[2:], stdout, stderr)
	case "report":
		return runReportCertificationEvidence(args[2:], stdout, stderr)
	default:
		logln(stderr, "connectorgen certification-evidence: require transport, change-capture, or report")
		return 2
	}
}

func runTransportCertificationEvidence(args []string, stdout, stderr io.Writer) int {
	options, err := parseTransportEvidenceOptions(args)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 2
	}
	report, err := loadCertificationEvidenceReport(options.reportPath)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	if err := validateNativeDatabaseTransportEvidenceReport(report, options.connector); err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	prepared, err := certificationEvidencePreparedValue(options.secretEnv)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}

	evidenceImport, err := certificationEvidenceImport(options.connector)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	if evidenceImport.Database == nil {
		logln(stderr, "connectorgen certification-evidence: connector does not declare database proof metadata")
		return 1
	}
	provider := evidenceImport.Provider
	databaseProof := evidenceImport.Database
	modes := append([]certify.DeclaredTransportModeResult(nil), report.Capabilities.DeclaredTransport.Modes...)
	sort.Slice(modes, func(i, j int) bool { return modes[i].Mode < modes[j].Mode })
	written := 0
	for _, mode := range modes {
		for _, primitive := range []string{"database_read_into_warehouse", "database_write_from_warehouse"} {
			output := filepath.Join(options.repoRoot, acceptedEvidenceDirectory, options.recordPrefix+"-"+mode.Mode+"-"+primitive+".json")
			body, err := json.Marshal(struct {
				Mode                string `json:"mode"`
				ApplyStrategy       string `json:"apply_strategy"`
				RecordsRead         int    `json:"records_read"`
				RecordsLoaded       int    `json:"records_loaded"`
				CheckpointCommitted bool   `json:"checkpoint_committed"`
				TargetNamespace     string `json:"target_namespace"`
				TargetRelation      string `json:"target_relation"`
			}{
				Mode: mode.Mode, ApplyStrategy: mode.ApplyStrategy, RecordsRead: mode.RecordsRead,
				RecordsLoaded: mode.RecordsLoaded, CheckpointCommitted: mode.CheckpointCommitted,
				TargetNamespace: mode.TargetNamespace, TargetRelation: mode.TargetRelation,
			})
			if err != nil {
				logf(stderr, "connectorgen certification-evidence: encode %s/%s: %v\n", mode.Mode, primitive, err)
				return 1
			}
			_, err = writeProofBearingEvidence(options.repoRoot, output, completedLiveEvidence{
				SchemaVersion: certificationSchemaVersion, Scope: evidenceScopeSyncMode, Connector: options.connector,
				SyncMode: mode.Mode, Primitive: primitive, Provider: provider,
				ExecutedAt:     report.CompletedAt.UTC().Format("2006-01-02T15:04:05Z"),
				RunID:          options.runID + "-" + mode.Mode + "-" + primitive,
				PMBinarySHA256: options.binarySHA, PMCommand: "pm connectors certify " + options.connector + " --full --write --from-env password=" + options.secretEnv,
				Passed: true, PreparedValues: []string{prepared},
				DatabaseExchanges: []completedDatabaseExchange{{
					Operation: databaseProof.OperationPrefix + "_" + mode.Mode + "_" + primitive,
					Protocol:  databaseProof.Protocol, Statement: "declared_transport_" + mode.Mode + "_" + mode.ApplyStrategy,
					ResponseStatus: "completed", ResponseBody: body,
				}},
			})
			if err != nil {
				logf(stderr, "connectorgen certification-evidence: write %s/%s: %v\n", mode.Mode, primitive, err)
				return 1
			}
			written++
		}
	}
	logf(stdout, "wrote declared transport evidence records: %d\n", written)
	return 0
}

type reportEvidenceOptions struct {
	connector         string
	reportPath        string
	externalProofPath string
	recordPrefix      string
	repoRoot          string
}

func runReportCertificationEvidence(args []string, stdout, stderr io.Writer) int {
	options, err := parseReportEvidenceOptions(args)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 2
	}
	report, err := loadCertificationEvidenceReport(options.reportPath)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	if err := validateCompletedCertificationEvidenceReport(report, options.connector); err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	proof, err := certify.ReadExternalProof(options.repoRoot, options.externalProofPath)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	if proof.Connector != options.connector {
		logln(stderr, "connectorgen certification-evidence: external proof connector does not match report connector")
		return 1
	}
	if proof.RunID != externalProofRunIDForReport(report) {
		logln(stderr, "connectorgen certification-evidence: external proof does not belong to the completed report run")
		return 1
	}
	bundle, err := engine.Load(defs.FS, options.connector)
	if err != nil || bundle.Certification == nil || bundle.Certification.EvidenceImport == nil {
		logln(stderr, "connectorgen certification-evidence: connector does not declare report evidence import bindings")
		return 1
	}
	bindings := bundle.Certification.EvidenceImport.Bindings
	if len(bindings) == 0 {
		logln(stderr, "connectorgen certification-evidence: connector declares no report evidence import bindings")
		return 1
	}
	exchanges := importedHTTPExchanges(proof.HTTPExchanges)
	type pendingEvidence struct {
		path     string
		identity importedLiveEvidence
	}
	pending := make([]pendingEvidence, 0, len(bindings))
	paths := make(map[string]struct{}, len(bindings))
	for index, binding := range bindings {
		if binding.Scope == evidenceScopeFlow {
			logln(stderr, "connectorgen certification-evidence: flow evidence requires a delivery receipt in addition to an HTTP proof")
			return 1
		}
		if err := validateEvidenceBindingStages(report, *bundle.Certification, binding); err != nil {
			logf(stderr, "connectorgen certification-evidence: binding %d: %v\n", index+1, err)
			return 1
		}
		output := filepath.Join(options.repoRoot, acceptedEvidenceDirectory, options.recordPrefix+"-"+evidenceBindingSuffix(binding)+".json")
		if _, exists := paths[output]; exists {
			logf(stderr, "connectorgen certification-evidence: binding %d resolves to a duplicate evidence path\n", index+1)
			return 1
		}
		paths[output] = struct{}{}
		identity := importedLiveEvidence{
			SchemaVersion:          certificationSchemaVersion,
			Scope:                  binding.Scope,
			Connector:              options.connector,
			FunctionKind:           binding.FunctionKind,
			WorkflowKind:           binding.WorkflowKind,
			SyncMode:               binding.SyncMode,
			Primitive:              binding.Primitive,
			Source:                 binding.Source,
			Destination:            binding.Destination,
			FlowKind:               binding.FlowKind,
			Provider:               bundle.Certification.EvidenceImport.Provider,
			ExecutedAt:             report.CompletedAt.UTC().Format(time.RFC3339),
			RunID:                  proof.RunID + "-" + evidenceBindingSuffix(binding),
			PMBinarySHA256:         proof.PMBinarySHA256,
			PMCommandFingerprint:   proof.PMCommandFingerprint,
			CredentialFingerprints: append([]string(nil), proof.CredentialFingerprints...),
			HTTPExchanges:          exchanges,
		}
		pending = append(pending, pendingEvidence{path: output, identity: identity})
	}
	for index, record := range pending {
		if _, err := writeImportedProofBearingEvidence(options.repoRoot, record.path, record.identity); err != nil {
			logf(stderr, "connectorgen certification-evidence: write binding %d: %v\n", index+1, err)
			return 1
		}
	}
	logf(stdout, "wrote report evidence records: %d\n", len(pending))
	return 0
}

func parseReportEvidenceOptions(args []string) (reportEvidenceOptions, error) {
	options := reportEvidenceOptions{repoRoot: "."}
	for index := 0; index < len(args); index++ {
		flag := args[index]
		if !strings.HasPrefix(flag, "--") || index+1 >= len(args) {
			return reportEvidenceOptions{}, fmt.Errorf("invalid flag %q", flag)
		}
		index++
		value := args[index]
		switch flag {
		case "--connector":
			if !isSafeProofIdentifier(value) {
				return reportEvidenceOptions{}, errors.New("--connector must be a safe connector name")
			}
			options.connector = value
		case "--report":
			options.reportPath = value
		case "--external-proof":
			options.externalProofPath = value
		case "--record-prefix":
			options.recordPrefix = value
		case "--repo-root":
			options.repoRoot = value
		default:
			return reportEvidenceOptions{}, fmt.Errorf("unknown flag %q", flag)
		}
	}
	if options.connector == "" || strings.TrimSpace(options.reportPath) == "" || strings.TrimSpace(options.externalProofPath) == "" || !isSafeProofIdentifier(options.recordPrefix) {
		return reportEvidenceOptions{}, errors.New("--connector, --report, --external-proof, and --record-prefix are required")
	}
	root, err := filepath.Abs(options.repoRoot)
	if err != nil {
		return reportEvidenceOptions{}, fmt.Errorf("resolve repository root: %w", err)
	}
	options.repoRoot = root
	return options, nil
}

func validateCompletedCertificationEvidenceReport(report certify.Report, connector string) error {
	if report.Kind != "ConnectorCertification" || report.Connector != connector || !report.Passed || report.StartedAt.IsZero() || report.CompletedAt.IsZero() {
		return errors.New("report is not a completed passing connector certification")
	}
	if !report.FullParityVerified() {
		return errors.New("report does not prove the accepted full-parity credential scope")
	}
	return nil
}

// externalProofRunIDForReport mirrors the fresh-child writer's report-bound
// identifier. It prevents an importer from combining a passing report from
// one run with a redacted exchange transcript from another run.
func externalProofRunIDForReport(report certify.Report) string {
	return fmt.Sprintf("external-%d", report.StartedAt.UTC().UnixNano())
}

func importedHTTPExchanges(exchanges []certify.ImportedExternalHTTPExchange) []certifiedHTTPExchange {
	result := make([]certifiedHTTPExchange, 0, len(exchanges))
	for index, exchange := range exchanges {
		result = append(result, certifiedHTTPExchange{
			Operation: fmt.Sprintf("http_%03d", index+1),
			Request: certifiedHTTPRequest{
				Method: exchange.Request.Method, Target: exchange.Request.Target,
				Query: importedQuery(exchange.Request.Query), Headers: importedHTTPFields(exchange.Request.Headers),
				Body: importedHTTPBody(exchange.Request.Body),
			},
			Response: certifiedHTTPResponse{
				Status: exchange.Response.Status, Headers: importedHTTPFields(exchange.Response.Headers),
				Body: importedHTTPBody(exchange.Response.Body),
			},
		})
	}
	return result
}

func importedQuery(fields []certify.ImportedExternalProofField) []certifiedQuery {
	result := make([]certifiedQuery, len(fields))
	for i, field := range fields {
		result[i] = certifiedQuery{Name: field.Name, Value: field.Value}
	}
	return result
}

func importedHTTPFields(fields []certify.ImportedExternalProofField) []certifiedHTTPField {
	result := make([]certifiedHTTPField, len(fields))
	for i, field := range fields {
		result[i] = certifiedHTTPField{Name: field.Name, Value: field.Value}
	}
	return result
}

func importedHTTPBody(body certify.ImportedExternalProofBody) certifiedHTTPBody {
	return certifiedHTTPBody{Encoding: body.Encoding, Value: append(json.RawMessage(nil), body.Value...), OriginalBytes: body.OriginalBytes}
}

func validateEvidenceBindingStages(report certify.Report, certification engine.CertificationSpec, binding engine.CertificationEvidenceImportBinding) error {
	required := make(map[string]struct{})
	for _, stage := range binding.Stages {
		required[stage] = struct{}{}
	}
	for _, set := range binding.StageSets {
		var candidates []engine.CertificationCommandCandidate
		switch set {
		case "direct_read_candidates":
			candidates = certification.DirectReadCandidates
		case "binary_candidates":
			candidates = certification.BinaryCandidates
		case "graphql_live_candidates":
			if certification.GraphQL != nil {
				candidates = certification.GraphQL.LiveCandidates
			}
		}
		for _, candidate := range candidates {
			required[candidate.StageName] = struct{}{}
		}
	}
	completed := make(map[string]certify.StageResult, len(report.Stages))
	for _, stage := range report.Stages {
		completed[stage.Name] = stage
	}
	for stage := range required {
		result, ok := completed[stage]
		if !ok || !result.Passed || result.Resumed {
			return fmt.Errorf("required stage %q was not freshly completed and passing", stage)
		}
	}
	return nil
}

func evidenceBindingSuffix(binding engine.CertificationEvidenceImportBinding) string {
	value := binding.Scope
	switch binding.Scope {
	case evidenceScopeCapability:
		value += "-" + binding.FunctionKind
	case evidenceScopeWorkflow:
		value += "-" + binding.WorkflowKind
	case evidenceScopeSyncMode:
		value += "-" + binding.SyncMode + "-" + binding.Primitive
	case evidenceScopeFlow:
		value += "-" + binding.Source + "-" + binding.Destination + "-" + binding.FlowKind
	}
	return strings.NewReplacer(":", "_", ".", "_", "-", "_").Replace(value)
}

// changeCaptureEvidenceReport is the deliberately small receipt from
// the independently asserted pm binary test. It contains delivery facts only;
// the proof writer fingerprints the actual credential and command before a
// record becomes publishable.
type changeCaptureEvidenceReport struct {
	Kind                              string    `json:"kind"`
	Connector                         string    `json:"connector"`
	CompletedAt                       time.Time `json:"completed_at"`
	BoundedDurableStaging             bool      `json:"bounded_durable_staging"`
	WarehouseReceiptPersisted         bool      `json:"warehouse_receipt_persisted"`
	IndependentWarehouseReadback      bool      `json:"independent_warehouse_readback"`
	AcknowledgedAfterWarehouseReceipt bool      `json:"acknowledged_after_warehouse_receipt"`
}

func runChangeCaptureCertificationEvidence(args []string, stdout, stderr io.Writer) int {
	options, err := parseTransportEvidenceOptions(args)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 2
	}
	report, err := loadChangeCaptureEvidenceReport(options.reportPath)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	if err := validateChangeCaptureEvidenceReport(report, options.connector); err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	prepared, err := certificationEvidencePreparedValue(options.secretEnv)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	body, err := json.Marshal(struct {
		BoundedDurableStaging             bool `json:"bounded_durable_staging"`
		WarehouseReceiptPersisted         bool `json:"warehouse_receipt_persisted"`
		IndependentWarehouseReadback      bool `json:"independent_warehouse_readback"`
		AcknowledgedAfterWarehouseReceipt bool `json:"acknowledged_after_warehouse_receipt"`
	}{
		BoundedDurableStaging:             report.BoundedDurableStaging,
		WarehouseReceiptPersisted:         report.WarehouseReceiptPersisted,
		IndependentWarehouseReadback:      report.IndependentWarehouseReadback,
		AcknowledgedAfterWarehouseReceipt: report.AcknowledgedAfterWarehouseReceipt,
	})
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: encode change-capture report: %v\n", err)
		return 1
	}
	evidenceImport, err := certificationEvidenceImport(options.connector)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	if evidenceImport.Database == nil {
		logln(stderr, "connectorgen certification-evidence: connector does not declare database proof metadata")
		return 1
	}
	provider := evidenceImport.Provider
	databaseProof := evidenceImport.Database
	completed := func(scope, functionKind, mode, primitive, suffix string) completedLiveEvidence {
		return completedLiveEvidence{
			SchemaVersion:  certificationSchemaVersion,
			Scope:          scope,
			Connector:      options.connector,
			FunctionKind:   functionKind,
			SyncMode:       mode,
			Primitive:      primitive,
			Provider:       provider,
			ExecutedAt:     report.CompletedAt.UTC().Format(time.RFC3339),
			RunID:          options.runID + "-" + suffix,
			PMBinarySHA256: options.binarySHA,
			PMCommand:      "pm etl run --sync-mode change_capture --from-env password=" + options.secretEnv,
			Passed:         true, PreparedValues: []string{prepared},
			DatabaseExchanges: []completedDatabaseExchange{{
				Operation: options.connector + "_change_capture_" + suffix,
				Protocol:  databaseProof.Protocol, Statement: databaseProof.ChangeCaptureStatement,
				ResponseStatus: "completed", ResponseBody: body,
			}},
		}
	}
	records := []struct {
		name      string
		completed completedLiveEvidence
	}{
		{name: options.recordPrefix + "-capability-cdc.json", completed: completed(evidenceScopeCapability, "capability:cdc", "", "", "capability_cdc")},
		{name: options.recordPrefix + "-change_capture-database_read_into_warehouse.json", completed: completed(evidenceScopeSyncMode, "", "change_capture", "database_read_into_warehouse", "change_capture_database_read")},
	}
	for _, record := range records {
		if _, err := writeProofBearingEvidence(options.repoRoot, filepath.Join(options.repoRoot, acceptedEvidenceDirectory, record.name), record.completed); err != nil {
			logf(stderr, "connectorgen certification-evidence: write %s: %v\n", record.name, err)
			return 1
		}
	}
	logf(stdout, "wrote change-capture evidence records: %d\n", len(records))
	return 0
}

func loadChangeCaptureEvidenceReport(path string) (changeCaptureEvidenceReport, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return changeCaptureEvidenceReport{}, fmt.Errorf("read report: %w", err)
	}
	var report changeCaptureEvidenceReport
	if err := json.Unmarshal(raw, &report); err != nil {
		return changeCaptureEvidenceReport{}, fmt.Errorf("parse report: %w", err)
	}
	return report, nil
}

func validateChangeCaptureEvidenceReport(report changeCaptureEvidenceReport, connector string) error {
	if report.Kind != "ChangeCaptureCertification" || report.Connector != connector || report.CompletedAt.IsZero() {
		return errors.New("report is not a completed change-capture certification")
	}
	if !report.BoundedDurableStaging || !report.WarehouseReceiptPersisted || !report.IndependentWarehouseReadback || !report.AcknowledgedAfterWarehouseReceipt {
		return errors.New("report requires bounded staging, independent warehouse read-back, and acknowledgement after receipt")
	}
	bundle, err := engine.Load(defs.FS, connector)
	if err != nil || bundle.Metadata.IntegrationType != "database" || !bundle.Metadata.Capabilities.CDC || bundle.Changefeed == nil || bundle.Changefeed.Status != connectors.ChangefeedStatusImplemented {
		return errors.New("connector does not declare an implemented database change-capture contract")
	}
	return nil
}

type transportEvidenceOptions struct {
	connector    string
	reportPath   string
	binarySHA    string
	secretEnv    string
	runID        string
	recordPrefix string
	repoRoot     string
}

func parseTransportEvidenceOptions(args []string) (transportEvidenceOptions, error) {
	options := transportEvidenceOptions{repoRoot: "."}
	for index := 0; index < len(args); index++ {
		flag := args[index]
		if !strings.HasPrefix(flag, "--") || index+1 >= len(args) {
			return transportEvidenceOptions{}, fmt.Errorf("invalid flag %q", flag)
		}
		index++
		value := args[index]
		switch flag {
		case "--report":
			options.reportPath = value
		case "--connector":
			if !isSafeProofIdentifier(value) {
				return transportEvidenceOptions{}, fmt.Errorf("--connector must be a safe connector name")
			}
			options.connector = value
		case "--binary-sha":
			options.binarySHA = value
		case "--from-env":
			field, envName, ok := strings.Cut(value, "=")
			if !ok || field != "password" || !isSafeProofIdentifier(envName) {
				return transportEvidenceOptions{}, errors.New("--from-env must be password=<safe environment variable>")
			}
			options.secretEnv = envName
		case "--run-id":
			options.runID = value
		case "--record-prefix":
			options.recordPrefix = value
		case "--repo-root":
			options.repoRoot = value
		default:
			return transportEvidenceOptions{}, fmt.Errorf("unknown flag %q", flag)
		}
	}
	if options.connector == "" || strings.TrimSpace(options.reportPath) == "" || !isSHA256(options.binarySHA) || options.secretEnv == "" || !isSafeProofIdentifier(options.runID) || !isSafeProofIdentifier(options.recordPrefix) {
		return transportEvidenceOptions{}, errors.New("--connector, --report, lowercase --binary-sha, --from-env password=<ENV>, --run-id, and --record-prefix are required")
	}
	root, err := filepath.Abs(options.repoRoot)
	if err != nil {
		return transportEvidenceOptions{}, fmt.Errorf("resolve repository root: %w", err)
	}
	options.repoRoot = root
	return options, nil
}

func loadCertificationEvidenceReport(path string) (certify.Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return certify.Report{}, fmt.Errorf("read report: %w", err)
	}
	var envelope struct {
		Report json.RawMessage `json:"report"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return certify.Report{}, fmt.Errorf("parse report: %w", err)
	}
	payload := raw
	if len(envelope.Report) != 0 {
		payload = envelope.Report
	}
	var report certify.Report
	if err := json.Unmarshal(payload, &report); err != nil {
		return certify.Report{}, fmt.Errorf("parse report: %w", err)
	}
	if report.Kind != "ConnectorCertification" {
		return certify.Report{}, fmt.Errorf("report kind %q is not ConnectorCertification", report.Kind)
	}
	return report, nil
}

func certificationEvidenceImport(connector string) (*engine.CertificationEvidenceImportSpec, error) {
	bundle, err := engine.Load(defs.FS, connector)
	if err != nil || bundle.Certification == nil || bundle.Certification.EvidenceImport == nil {
		return nil, errors.New("connector does not declare an evidence import provider")
	}
	provider := bundle.Certification.EvidenceImport.Provider
	if !isSafeProofIdentifier(provider) {
		return nil, errors.New("connector declares an unsafe evidence import provider")
	}
	return bundle.Certification.EvidenceImport, nil
}

func validateNativeDatabaseTransportEvidenceReport(report certify.Report, connector string) error {
	if report.Connector != connector || !report.Passed || report.Capabilities.DeclaredTransport == nil {
		return errors.New("report is not a completed passing connector certification")
	}
	bundle, err := engine.Load(defs.FS, connector)
	if err != nil || bundle.SyncTransport == nil || bundle.SyncTransport.Source == nil || bundle.SyncTransport.Destination == nil {
		return errors.New("connector does not declare a complete source and destination transport pair")
	}
	if bundle.SyncTransport.Source.Executor.Family != connectors.TransportExecutorFamilyNativeDatabase || bundle.SyncTransport.Destination.Executor.Family != connectors.TransportExecutorFamilyNativeDatabase {
		return errors.New("connector does not declare a native-database transport pair")
	}
	transport := report.Capabilities.DeclaredTransport
	if transport.Result != "pass" || transport.SourceExecutor != bundle.SyncTransport.Source.Executor.ID || transport.DestinationExecutor != bundle.SyncTransport.Destination.Executor.ID {
		return errors.New("report did not execute the connector's declared transport pair")
	}
	expected := make(map[string]string, len(bundle.SyncTransport.Source.Modes))
	for _, strategy := range bundle.SyncTransport.Destination.ApplyStrategies {
		expected[string(strategy.Mode)] = string(strategy.Strategy)
	}
	if len(transport.Modes) != len(bundle.SyncTransport.Source.Modes) {
		return fmt.Errorf("report has %d transport modes, want %d declared modes", len(transport.Modes), len(bundle.SyncTransport.Source.Modes))
	}
	seen := make(map[string]bool, len(expected))
	for _, mode := range transport.Modes {
		if expected[mode.Mode] != mode.ApplyStrategy || seen[mode.Mode] || mode.RecordsRead <= 0 || mode.RecordsLoaded <= 0 || !mode.CheckpointCommitted || !isSafeProofIdentifier(mode.TargetNamespace) || !isSafeProofIdentifier(mode.TargetRelation) {
			return fmt.Errorf("report transport mode %q lacks completed declared target/read/checkpoint proof", mode.Mode)
		}
		seen[mode.Mode] = true
	}
	for _, mode := range bundle.SyncTransport.Source.Modes {
		if !seen[string(mode)] {
			return fmt.Errorf("report omitted declared transport mode %q", mode)
		}
	}
	return nil
}

func certificationEvidencePreparedValue(envName string) (string, error) {
	value := os.Getenv(envName)
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("environment variable %s is empty", envName)
	}
	return value, nil
}
