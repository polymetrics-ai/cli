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

const declaredTransportEvidenceProvider = "postgres_container"

// runCertificationEvidence accepts a completed, passing declared-transport or
// change-capture report and emits proof at the requested contract scope. The
// selected bundle binds report executor references and apply strategies; this
// generic importer adds no runtime transport behavior.
func runCertificationEvidence(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		logln(stderr, "connectorgen certification-evidence: require transport, transport-capability-write, or change-capture")
		return 2
	}
	switch args[1] {
	case "transport":
		return runPostgresTransportCertificationEvidence(args[2:], stdout, stderr)
	case "transport-capability-write":
		return runPostgresTransportCapabilityWriteEvidence(args[2:], stdout, stderr)
	case "change-capture":
		return runChangeCaptureCertificationEvidence(args[2:], stdout, stderr)
	default:
		logln(stderr, "connectorgen certification-evidence: require transport, transport-capability-write, or change-capture")
		return 2
	}
}

func runPostgresTransportCertificationEvidence(args []string, stdout, stderr io.Writer) int {
	options, err := parsePostgresTransportEvidenceOptions(args)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 2
	}
	report, err := loadPostgresTransportEvidenceReport(options.reportPath)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	if err := validatePostgresTransportEvidenceReport(report, options.connector); err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	prepared, err := postgresTransportEvidencePreparedValue(options.secretEnv)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}

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
				SyncMode: mode.Mode, Primitive: primitive, Provider: declaredTransportEvidenceProvider,
				ExecutedAt:     report.CompletedAt.UTC().Format("2006-01-02T15:04:05Z"),
				RunID:          options.runID + "-" + mode.Mode + "-" + primitive,
				PMBinarySHA256: options.binarySHA, PMCommand: "pm connectors certify " + options.connector + " --full --write --from-env password=" + options.secretEnv,
				Passed: true, CredentialFullParity: true, PreparedValues: []string{prepared},
				DatabaseExchanges: []completedDatabaseExchange{{
					Operation: "postgres_transport_" + mode.Mode + "_" + primitive,
					Protocol:  "postgres_wire", Statement: "declared_transport_" + mode.Mode + "_" + mode.ApplyStrategy,
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

// runPostgresTransportCapabilityWriteEvidence records the broad write claim
// separately from the mode records. A sync_mode record proves exactly one
// mode; it is never eligible to satisfy this capability cell. The aggregate
// certificate report is eligible because validatePostgresTransportEvidenceReport
// requires every declared destination mode to have completed independently
// exercised target/read/checkpoint evidence.
func runPostgresTransportCapabilityWriteEvidence(args []string, stdout, stderr io.Writer) int {
	options, err := parsePostgresTransportEvidenceOptions(args)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 2
	}
	report, err := loadPostgresTransportEvidenceReport(options.reportPath)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	if err := validatePostgresTransportEvidenceReport(report, options.connector); err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	prepared, err := postgresTransportEvidencePreparedValue(options.secretEnv)
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: %v\n", err)
		return 1
	}
	modes := append([]certify.DeclaredTransportModeResult(nil), report.Capabilities.DeclaredTransport.Modes...)
	sort.Slice(modes, func(i, j int) bool { return modes[i].Mode < modes[j].Mode })
	body, err := json.Marshal(struct {
		SourceExecutor      string                                `json:"source_executor"`
		DestinationExecutor string                                `json:"destination_executor"`
		Modes               []certify.DeclaredTransportModeResult `json:"completed_destination_modes"`
	}{
		SourceExecutor:      report.Capabilities.DeclaredTransport.SourceExecutor,
		DestinationExecutor: report.Capabilities.DeclaredTransport.DestinationExecutor,
		Modes:               modes,
	})
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: encode transport capability write: %v\n", err)
		return 1
	}
	output := filepath.Join(options.repoRoot, acceptedEvidenceDirectory, options.recordPrefix+"-capability-write.json")
	_, err = writeProofBearingEvidence(options.repoRoot, output, completedLiveEvidence{
		SchemaVersion: certificationSchemaVersion, Scope: evidenceScopeCapability, Connector: options.connector,
		FunctionKind: "capability:write", Provider: declaredTransportEvidenceProvider,
		ExecutedAt:     report.CompletedAt.UTC().Format("2006-01-02T15:04:05Z"),
		RunID:          options.runID + "-capability-write",
		PMBinarySHA256: options.binarySHA, PMCommand: "pm connectors certify " + options.connector + " --full --write --from-env password=" + options.secretEnv,
		Passed: true, CredentialFullParity: true, PreparedValues: []string{prepared},
		DatabaseExchanges: []completedDatabaseExchange{{
			Operation: options.connector + "_transport_capability_write",
			Protocol:  "postgres_wire", Statement: "declared_transport_all_destination_modes",
			ResponseStatus: "completed", ResponseBody: body,
		}},
	})
	if err != nil {
		logf(stderr, "connectorgen certification-evidence: write transport capability write: %v\n", err)
		return 1
	}
	logln(stdout, "wrote declared transport capability evidence records: 1")
	return 0
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
	options, err := parsePostgresTransportEvidenceOptions(args)
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
	prepared, err := postgresTransportEvidencePreparedValue(options.secretEnv)
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
	completed := func(scope, functionKind, mode, primitive, suffix string) completedLiveEvidence {
		return completedLiveEvidence{
			SchemaVersion:  certificationSchemaVersion,
			Scope:          scope,
			Connector:      options.connector,
			FunctionKind:   functionKind,
			SyncMode:       mode,
			Primitive:      primitive,
			Provider:       declaredTransportEvidenceProvider,
			ExecutedAt:     report.CompletedAt.UTC().Format(time.RFC3339),
			RunID:          options.runID + "-" + suffix,
			PMBinarySHA256: options.binarySHA,
			PMCommand:      "pm etl run --sync-mode change_capture --from-env password=" + options.secretEnv,
			Passed:         true, CredentialFullParity: true, PreparedValues: []string{prepared},
			DatabaseExchanges: []completedDatabaseExchange{{
				Operation: options.connector + "_change_capture_" + suffix,
				Protocol:  "postgres_wire", Statement: "pgoutput_v2_receipt_before_acknowledgement",
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

type postgresTransportEvidenceOptions struct {
	connector    string
	reportPath   string
	binarySHA    string
	secretEnv    string
	runID        string
	recordPrefix string
	repoRoot     string
}

func parsePostgresTransportEvidenceOptions(args []string) (postgresTransportEvidenceOptions, error) {
	options := postgresTransportEvidenceOptions{repoRoot: "."}
	for index := 0; index < len(args); index++ {
		flag := args[index]
		if !strings.HasPrefix(flag, "--") || index+1 >= len(args) {
			return postgresTransportEvidenceOptions{}, fmt.Errorf("invalid flag %q", flag)
		}
		index++
		value := args[index]
		switch flag {
		case "--report":
			options.reportPath = value
		case "--connector":
			if !isSafeProofIdentifier(value) {
				return postgresTransportEvidenceOptions{}, fmt.Errorf("--connector must be a safe connector name")
			}
			options.connector = value
		case "--binary-sha":
			options.binarySHA = value
		case "--from-env":
			field, envName, ok := strings.Cut(value, "=")
			if !ok || field != "password" || !isSafeProofIdentifier(envName) {
				return postgresTransportEvidenceOptions{}, errors.New("--from-env must be password=<safe environment variable>")
			}
			options.secretEnv = envName
		case "--run-id":
			options.runID = value
		case "--record-prefix":
			options.recordPrefix = value
		case "--repo-root":
			options.repoRoot = value
		default:
			return postgresTransportEvidenceOptions{}, fmt.Errorf("unknown flag %q", flag)
		}
	}
	if options.connector == "" || strings.TrimSpace(options.reportPath) == "" || !isSHA256(options.binarySHA) || options.secretEnv == "" || !isSafeProofIdentifier(options.runID) || !isSafeProofIdentifier(options.recordPrefix) {
		return postgresTransportEvidenceOptions{}, errors.New("--connector, --report, lowercase --binary-sha, --from-env password=<ENV>, --run-id, and --record-prefix are required")
	}
	root, err := filepath.Abs(options.repoRoot)
	if err != nil {
		return postgresTransportEvidenceOptions{}, fmt.Errorf("resolve repository root: %w", err)
	}
	options.repoRoot = root
	return options, nil
}

func loadPostgresTransportEvidenceReport(path string) (certify.Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return certify.Report{}, fmt.Errorf("read report: %w", err)
	}
	var envelope struct {
		Kind   string         `json:"kind"`
		Report certify.Report `json:"report"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return certify.Report{}, fmt.Errorf("parse report: %w", err)
	}
	if envelope.Kind != "ConnectorCertification" {
		return certify.Report{}, fmt.Errorf("report kind %q is not ConnectorCertification", envelope.Kind)
	}
	return envelope.Report, nil
}

func validatePostgresTransportEvidenceReport(report certify.Report, connector string) error {
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

func postgresTransportEvidencePreparedValue(envName string) (string, error) {
	value := os.Getenv(envName)
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("environment variable %s is empty", envName)
	}
	return value, nil
}
