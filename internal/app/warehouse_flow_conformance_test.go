package app

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"polymetrics.ai/internal/certificationcatalog"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// TestWarehouseMediatedFourPathConformance binds the four generated flow IDs
// to the real GitHub/PostgreSQL registry. It intentionally complements, rather
// than repeats, the opt-in binary route proofs: a changed declaration must fail
// the exact direction whose production descriptors can no longer be resolved.
func TestWarehouseMediatedFourPathConformance(t *testing.T) {
	contracts, err := warehouseMediatedFlowConformanceContracts()
	if err != nil {
		t.Fatal(err)
	}

	a := newWarehouseMediatedConformanceApp(t)
	results := make(map[string]warehouseMediatedConformanceResult, len(contracts))
	for _, contract := range contracts {
		contract := contract
		results[contract.FlowKind.ID] = warehouseMediatedConformanceResult{
			Status: warehouseMediatedConformanceUnexecuted,
			Reason: "direction has not yet resolved its real production transport",
		}
		t.Run(contract.FlowKind.ID, func(t *testing.T) {
			source := warehouseMediatedConformanceConnector(t, a, contract.SourceConnector)
			destination := warehouseMediatedConformanceConnector(t, a, contract.DestinationConnector)
			if got, want := source.Metadata().IntegrationType+"_source", contract.FlowKind.SourceRole; got != want {
				t.Fatalf("source role = %q, want generated role %q", got, want)
			}
			if got, want := destination.Metadata().IntegrationType+"_destination", contract.FlowKind.DestinationRole; got != want {
				t.Fatalf("destination role = %q, want generated role %q", got, want)
			}

			resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
				Source: source, Destination: destination, Stream: contract.Stream, Mode: contract.Mode,
			})
			if err != nil {
				t.Fatalf("production %s preflight = %v", contract.FlowKind.ID, err)
			}
			if got, want := resolved.Source.TransportExecutorReference(), contract.SourceExecutor; got != want {
				t.Fatalf("source executor = %+v, want %+v", got, want)
			}
			if got, want := resolved.Destination.TransportExecutorReference(), contract.DestinationExecutor; got != want {
				t.Fatalf("destination executor = %+v, want %+v", got, want)
			}
			if got, want := resolved.DestinationDescriptor.Acknowledgement, connectors.TransportAcknowledgementDurableWarehouse; got != want {
				t.Fatalf("destination acknowledgement = %q, want sealed warehouse acknowledgement %q", got, want)
			}
			connection := warehouseMediatedConformanceConnection(contract)
			a.state.Connections = append(a.state.Connections, connection)
			if !a.shouldRunTransport(connection, contract.Stream, SyncMode{ContractMode: contract.Mode}, source, destination) {
				t.Fatalf("production dispatch did not select %s after preflight", contract.FlowKind.ID)
			}
			results[contract.FlowKind.ID] = warehouseMediatedConformanceResult{Status: warehouseMediatedConformancePass}
		})
	}
	if !allWarehouseMediatedConformancePassed(contracts, results) {
		t.Fatalf("four-path conformance results = %#v, want every named direction passed", results)
	}
}

// TestWarehouseMediatedModeConformance records the executable state of the
// closed vocabulary on this branch. Change capture is deliberately not treated
// as a PostgreSQL destination mode: its implemented PostgreSQL CDC source
// produces a durable workset, while normal target preflight remains refused.
func TestWarehouseMediatedModeConformance(t *testing.T) {
	a := newWarehouseMediatedConformanceApp(t)
	modes := warehouseMediatedModeConformanceCases()
	if got, want := len(modes), len(synccontract.AllModes()); got != want {
		t.Fatalf("mode conformance cases = %d, want one for each closed mode (%d)", got, want)
	}

	results := make(map[synccontract.Mode]warehouseMediatedConformanceResult, len(modes))
	for _, mode := range synccontract.AllModes() {
		mode := mode
		contract, ok := modes[mode]
		if !ok {
			t.Fatalf("closed mode %q has no conformance contract", mode)
		}
		t.Run(string(mode), func(t *testing.T) {
			if mode == synccontract.ModeChangeCapture {
				results[mode] = assertPostgresChangeCaptureIsSourceOnly(t, a)
				return
			}
			source := warehouseMediatedConformanceConnector(t, a, contract.SourceConnector)
			destination := warehouseMediatedConformanceConnector(t, a, contract.DestinationConnector)
			resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
				Source: source, Destination: destination, Stream: contract.Stream, Mode: mode,
			})
			if err != nil {
				t.Fatalf("current executable mode %q preflight = %v", mode, err)
			}
			if resolved.ApplyStrategy.Mode != mode {
				t.Fatalf("resolved apply mode = %q, want %q", resolved.ApplyStrategy.Mode, mode)
			}
			results[mode] = warehouseMediatedConformanceResult{Status: warehouseMediatedConformancePass}
		})
	}

	for _, mode := range synccontract.AllModes() {
		result, ok := results[mode]
		if !ok || result.Status == warehouseMediatedConformanceUnexecuted || result.Reason == "" && result.Status != warehouseMediatedConformancePass {
			t.Fatalf("mode %q result = %#v, want pass or a concrete non-pass reason", mode, result)
		}
	}
	if got := warehouseMediatedConformancePassCount(results); got != len(synccontract.AllModes())-1 {
		t.Fatalf("mode pass roll-up = %d, want only the six executable transport modes", got)
	}
}

func TestWarehouseMediatedConformancePassRollupRejectsNonPassDirection(t *testing.T) {
	contracts, err := warehouseMediatedFlowConformanceContracts()
	if err != nil {
		t.Fatal(err)
	}
	results := make(map[string]warehouseMediatedConformanceResult, len(contracts))
	for _, contract := range contracts {
		results[contract.FlowKind.ID] = warehouseMediatedConformanceResult{Status: warehouseMediatedConformancePass}
	}
	results["api_to_database"] = warehouseMediatedConformanceResult{
		Status: warehouseMediatedConformanceUnexecuted,
		Reason: "database destination was not executed",
	}
	if allWarehouseMediatedConformancePassed(contracts, results) {
		t.Fatal("pass roll-up accepted an unexecuted direction")
	}
}

type warehouseMediatedFlowConformanceContract struct {
	FlowKind             certificationcatalog.FlowKind
	SourceConnector      string
	DestinationConnector string
	Stream               string
	Mode                 synccontract.Mode
	SourceExecutor       connectors.TransportExecutorReference
	DestinationExecutor  connectors.TransportExecutorReference
}

func warehouseMediatedFlowConformanceContracts() ([]warehouseMediatedFlowConformanceContract, error) {
	byID := map[string]warehouseMediatedFlowConformanceContract{
		"api_to_api": {
			SourceConnector: "github", DestinationConnector: "github", Stream: "issues", Mode: synccontract.ModeFullAppend,
			SourceExecutor:      declarativeStreamSourceReference,
			DestinationExecutor: connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyDeclarativeAPI, ID: "issue_label_destination"},
		},
		"api_to_database": {
			SourceConnector: "github", DestinationConnector: "postgres", Stream: "commits", Mode: synccontract.ModeFullAppend,
			SourceExecutor:      declarativeStreamSourceReference,
			DestinationExecutor: connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_managed_target"},
		},
		"database_to_api": {
			SourceConnector: "postgres", DestinationConnector: "github", Stream: "public.issue_label_events", Mode: synccontract.ModeFullAppend,
			SourceExecutor:      connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_polling_watermark"},
			DestinationExecutor: connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyDeclarativeAPI, ID: "issue_label_destination"},
		},
		"database_to_database": {
			SourceConnector: "postgres", DestinationConnector: "postgres", Stream: "snapshot", Mode: synccontract.ModeFullAppend,
			SourceExecutor:      connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_polling_watermark"},
			DestinationExecutor: connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_managed_target"},
		},
	}
	flowKinds := certificationcatalog.FlowKinds()
	if len(flowKinds) != len(byID) {
		return nil, fmt.Errorf("generated flow kind catalog has %d entries, want %d four-path contracts", len(flowKinds), len(byID))
	}
	contracts := make([]warehouseMediatedFlowConformanceContract, 0, len(flowKinds))
	for _, flowKind := range flowKinds {
		contract, ok := byID[flowKind.ID]
		if !ok {
			return nil, fmt.Errorf("generated flow kind %q has no warehouse-mediated conformance contract", flowKind.ID)
		}
		contract.FlowKind = flowKind
		contracts = append(contracts, contract)
	}
	return contracts, nil
}

type warehouseMediatedModeConformanceCase struct {
	SourceConnector      string
	DestinationConnector string
	Stream               string
}

func warehouseMediatedModeConformanceCases() map[synccontract.Mode]warehouseMediatedModeConformanceCase {
	return map[synccontract.Mode]warehouseMediatedModeConformanceCase{
		synccontract.ModeFullOverwrite:            {SourceConnector: "github", DestinationConnector: "github", Stream: "issues"},
		synccontract.ModeFullAppend:               {SourceConnector: "github", DestinationConnector: "postgres", Stream: "commits"},
		synccontract.ModeIncrementalAppend:        {SourceConnector: "postgres", DestinationConnector: "postgres", Stream: "snapshot"},
		synccontract.ModeIncrementalUpsert:        {SourceConnector: "postgres", DestinationConnector: "github", Stream: "public.issue_label_events"},
		synccontract.ModeIncrementalDedupe:        {SourceConnector: "github", DestinationConnector: "postgres", Stream: "pull_requests"},
		synccontract.ModeIncrementalDedupeHistory: {SourceConnector: "github", DestinationConnector: "postgres", Stream: "pull_requests"},
		synccontract.ModeChangeCapture:            {SourceConnector: "postgres", DestinationConnector: "warehouse", Stream: "runtime_catalog"},
	}
}

func assertPostgresChangeCaptureIsSourceOnly(t *testing.T, a *App) warehouseMediatedConformanceResult {
	t.Helper()
	postgres := warehouseMediatedConformanceConnector(t, a, "postgres")
	definition, ok := connectors.DefinitionOf(postgres)
	if !ok || !definition.Capabilities.CDC || definition.Changefeed == nil || !definition.Changefeed.IsImplemented() {
		t.Fatalf("PostgreSQL change-capture definition = %#v, want implemented source contract", definition)
	}
	if !slices.Contains(definition.Changefeed.Streams, "runtime_catalog") {
		t.Fatalf("PostgreSQL change-capture streams = %v, want runtime_catalog", definition.Changefeed.Streams)
	}
	destination, ok := connectors.DestinationTransportDescriptorOf(postgres)
	if !ok {
		t.Fatal("PostgreSQL normal destination transport is not declared")
	}
	if slices.Contains(destination.Modes, synccontract.ModeChangeCapture) {
		t.Fatal("PostgreSQL normal destination transport admitted change_capture")
	}
	_, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source: postgres, Destination: postgres, Stream: "snapshot", Mode: synccontract.ModeChangeCapture,
	})
	if err == nil || !strings.Contains(err.Error(), "does not support sync mode") {
		t.Fatalf("PostgreSQL destination-shaped change_capture preflight = %v, want pre-I/O refusal", err)
	}
	return warehouseMediatedConformanceResult{Status: warehouseMediatedConformanceRefused, Reason: err.Error()}
}

func newWarehouseMediatedConformanceApp(t *testing.T) *App {
	t.Helper()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if a.transportStage == nil {
		t.Fatal("Open() left the required warehouse stage unset")
	}
	return a
}

func warehouseMediatedConformanceConnector(t *testing.T, a *App, name string) connectors.Connector {
	t.Helper()
	connector, ok := a.registry.Get(name)
	if !ok {
		t.Fatalf("production connector %q is not registered", name)
	}
	return connector
}

func warehouseMediatedConformanceConnection(contract warehouseMediatedFlowConformanceContract) Connection {
	id := "four_path_" + strings.ReplaceAll(contract.FlowKind.ID, "_", "-")
	connection := Connection{
		ID:   id,
		Name: id,
		Source: EndpointConfig{
			Connector: contract.SourceConnector,
		},
		Destination: EndpointConfig{
			Connector: contract.DestinationConnector,
		},
		Streams: map[string]StreamConfig{
			contract.Stream: {
				SyncMode:         string(contract.Mode),
				CursorField:      "sequence",
				PrimaryKey:       []string{"id"},
				DestinationTable: "conformance_rows",
			},
		},
	}
	if contract.DestinationConnector == "github" {
		connection.Destination.Config = map[string]string{
			issueLabelTransportTargetIssueConfig: "200",
			issueLabelTransportLabelConfig:       "four-path-conformance",
		}
	}
	if contract.SourceConnector == "github" && contract.DestinationConnector == "github" {
		connection.Source.Config = map[string]string{issueLabelTransportSourceIssueConfig: "100"}
	}
	return connection
}

type warehouseMediatedConformanceStatus string

const (
	warehouseMediatedConformancePass       warehouseMediatedConformanceStatus = "pass"
	warehouseMediatedConformanceRefused    warehouseMediatedConformanceStatus = "refused"
	warehouseMediatedConformanceUnexecuted warehouseMediatedConformanceStatus = "unexecuted"
)

type warehouseMediatedConformanceResult struct {
	Status warehouseMediatedConformanceStatus
	Reason string
}

func allWarehouseMediatedConformancePassed(contracts []warehouseMediatedFlowConformanceContract, results map[string]warehouseMediatedConformanceResult) bool {
	if len(results) != len(contracts) {
		return false
	}
	for _, contract := range contracts {
		result, ok := results[contract.FlowKind.ID]
		if !ok || result.Status != warehouseMediatedConformancePass || result.Reason != "" {
			return false
		}
	}
	return true
}

func warehouseMediatedConformancePassCount(results map[synccontract.Mode]warehouseMediatedConformanceResult) int {
	passed := 0
	for _, result := range results {
		if result.Status == warehouseMediatedConformancePass {
			passed++
		}
	}
	return passed
}
