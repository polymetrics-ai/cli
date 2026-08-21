package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

func TestOpenRegistersDefinitionOwnedProductionTransports(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	postgres, ok := a.registry.Get("postgres")
	if !ok {
		t.Fatal("PostgreSQL connector is not registered")
	}
	github, ok := a.registry.Get("github")
	if !ok {
		t.Fatal("GitHub connector is not registered")
	}
	warehouse, ok := a.registry.Get("warehouse")
	if !ok {
		t.Fatal("local warehouse connector is not registered")
	}
	resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      postgres,
		Destination: github,
		Stream:      "snapshot",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("definition-owned PostgreSQL-to-GitHub preflight = %v", err)
	}
	if got, want := resolved.Source.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_polling_watermark"}); got != want {
		t.Fatalf("registered source reference = %+v, want %+v", got, want)
	}
	if got, want := resolved.Destination.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyDeclarativeAPI, ID: "issue_label_destination"}); got != want {
		t.Fatalf("registered destination reference = %+v, want %+v", got, want)
	}
	warehouseResolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      postgres,
		Destination: warehouse,
		Stream:      "snapshot",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("definition-owned PostgreSQL-to-warehouse preflight = %v", err)
	}
	if got, want := warehouseResolved.Destination.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "local_parquet_warehouse"}); got != want {
		t.Fatalf("registered warehouse destination reference = %+v, want %+v", got, want)
	}
	githubResolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      github,
		Destination: github,
		Stream:      "issues",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("definition-owned GitHub-to-GitHub preflight = %v", err)
	}
	if got, want := githubResolved.Source.TransportExecutorReference(), declarativeStreamSourceReference; got != want {
		t.Fatalf("registered GitHub source reference = %+v, want %+v", got, want)
	}
	postgresResolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      postgres,
		Destination: postgres,
		Stream:      "snapshot",
		Mode:        synccontract.ModeIncrementalUpsert,
	})
	if err != nil {
		t.Fatalf("definition-owned PostgreSQL-to-PostgreSQL preflight = %v", err)
	}
	if got, want := postgresResolved.Destination.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_managed_target"}); got != want {
		t.Fatalf("registered PostgreSQL destination reference = %+v, want %+v", got, want)
	}
	githubPostgresResolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      github,
		Destination: postgres,
		Stream:      "commits",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("definition-owned GitHub commits-to-PostgreSQL preflight = %v", err)
	}
	if got, want := githubPostgresResolved.Source.TransportExecutorReference(), declarativeStreamSourceReference; got != want {
		t.Fatalf("registered API source reference = %+v, want %+v", got, want)
	}
	if got, want := githubPostgresResolved.Destination.TransportExecutorReference(), postgresResolved.Destination.TransportExecutorReference(); got != want {
		t.Fatalf("registered API destination reference = %+v, want %+v", got, want)
	}
	if a.shouldRunTransport(Connection{}, "commits", SyncMode{ContractMode: synccontract.ModeFullAppend}, github, postgres) != true {
		t.Fatal("declared GitHub commits-to-PostgreSQL route was not selected for production dispatch")
	}
	for _, mode := range []synccontract.Mode{
		synccontract.ModeIncrementalDedupe,
		synccontract.ModeIncrementalDedupeHistory,
	} {
		t.Run("github_to_warehouse_"+string(mode), func(t *testing.T) {
			resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
				Source:      github,
				Destination: warehouse,
				Stream:      "pull_requests",
				Mode:        mode,
			})
			if err != nil {
				t.Fatalf("GitHub-to-warehouse %q preflight = %v", mode, err)
			}
			if got, want := resolved.Destination.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "local_parquet_warehouse"}); got != want {
				t.Fatalf("GitHub-to-warehouse %q destination = %+v, want %+v", mode, got, want)
			}
			if !a.shouldRunTransport(Connection{}, "pull_requests", SyncMode{ContractMode: mode}, github, warehouse) {
				t.Fatalf("declared GitHub-to-warehouse %q route was not selected for production dispatch", mode)
			}
		})
	}
	_, err = a.transports.Preflight(synctransport.PreflightRequest{
		Source:      github,
		Destination: warehouse,
		Stream:      "pull_requests",
		Mode:        synccontract.ModeIncrementalAppend,
	})
	if got, want := fmt.Sprint(err), `source transport does not support sync mode "incremental_append"`; got != want {
		t.Fatalf("GitHub-to-warehouse incremental_append preflight refusal = %q, want %q before executor I/O", got, want)
	}
	assertGitHubTransportEligibleStreamsMatchDefinition(t, github)
	_, err = a.transports.Preflight(synctransport.PreflightRequest{
		Source:      github,
		Destination: postgres,
		Stream:      "transport_ineligible_probe",
		Mode:        synccontract.ModeFullAppend,
	})
	var ineligible *synctransport.SourceStreamIneligibleError
	if !errors.As(err, &ineligible) {
		t.Fatalf("undeclared GitHub stream preflight = %v, want SourceStreamIneligibleError", err)
	}
	if postgres.Metadata().Capabilities.Write {
		t.Fatal("PostgreSQL published generic write capability for its closed managed destination")
	}
	if err := validateClosedTransportBatchSize(github, github, 2); err == nil {
		t.Fatal("closed issue-label destination accepted a batch larger than its one-record contract")
	}
	if err := validateClosedTransportBatchSize(github, postgres, 50); err != nil {
		t.Fatalf("GitHub managed-target transport rejected its bounded collection batch: %v", err)
	}
	if err := validateClosedTransportBatchSize(github, postgres, issueCollectionTransportMaxRecords+1); err == nil {
		t.Fatal("GitHub managed-target transport accepted an allocation-sized batch above its fixed bound")
	}
	if err := validateClosedTransportBatchSize(postgres, postgres, 1000); err != nil {
		t.Fatalf("PostgreSQL managed transport rejected its bounded database batch: %v", err)
	}
}

// TestAppCompositionRoutesLoadedSyntheticDefinitionConnector is the scalable
// registration proof for #4093. The connector's source and destination roles
// exist only in its loaded sync_transport.json; App discovers the test-only
// named hook through DefinitionFactoryProvider and the generic orchestrator
// runs the declared pair without a connector-name branch.
func TestAppCompositionRoutesLoadedSyntheticDefinitionConnector(t *testing.T) {
	bundle, err := engine.Load(syntheticTransportBundleFS(), "synthetic")
	if err != nil {
		t.Fatalf("load synthetic definition bundle: %v", err)
	}

	sourceReference := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeAPI, ID: "synthetic_snapshot_source"}
	destinationReference := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeAPI, ID: "synthetic_stage_destination"}
	source := &syntheticDefinitionSource{reference: sourceReference, page: synctransport.SourcePage{
		Records:             []connectors.Record{{"id": "synthetic-1", "value": "definition-owned"}},
		CandidateCheckpoint: syntheticDefinitionCheckpoint(),
	}}
	destination := &syntheticDefinitionDestination{reference: destinationReference, sink: "synthetic"}
	connector := &syntheticDefinitionConnector{Connector: engine.New(bundle, nil)}
	connector.factories = []synctransport.DefinitionFactory{
		{
			Reference:      sourceReference,
			SourceEvidence: connectors.ConformanceEvidenceReference{Suite: "synthetic_transport", RunID: "source_v1"},
			BuildSource: func(received connectors.Connector) (synctransport.SourceExecutor, error) {
				if received != connector {
					return nil, fmt.Errorf("synthetic source hook received another connector")
				}
				return source, nil
			},
		},
		{
			Reference:           destinationReference,
			DestinationEvidence: connectors.ConformanceEvidenceReference{Suite: "synthetic_transport", RunID: "destination_v1"},
			BuildDestination: func(received connectors.Connector) (synctransport.DestinationExecutor, error) {
				if received != connector {
					return nil, fmt.Errorf("synthetic destination hook received another connector")
				}
				return destination, nil
			},
		},
	}

	a := &App{registry: connectors.NewEmptyRegistry()}
	a.registry.Register(connector)
	if err := a.composeTransportRegistry(); err != nil {
		t.Fatalf("compose generic synthetic transport registry: %v", err)
	}
	if !a.shouldRunTransport(Connection{}, "widgets", SyncMode{ContractMode: synccontract.ModeFullAppend}, connector, connector) {
		t.Fatal("definition-owned synthetic connector was not selected for transport dispatch")
	}

	stage := &syntheticDefinitionStage{}
	commits := 0
	result, err := synctransport.NewOrchestrator(a.transports).Run(context.Background(), synctransport.RunRequest{
		ConnectionID: "synthetic-connection",
		Generation:   1,
		Source:       connector,
		Destination:  connector,
		Stream:       "widgets",
		Mode:         synccontract.ModeFullAppend,
		BatchSize:    1,
		Stage:        stage,
		Commit: func(checkpoint synccontract.CheckpointEnvelope) error {
			commits++
			if checkpoint.CommittedAt == nil {
				return fmt.Errorf("synthetic checkpoint was not acknowledged")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("run synthetic definition-owned transport: %v", err)
	}
	if source.reads != 1 || stage.stages != 1 || destination.plans != 1 || destination.applies != 1 || destination.readBacks != 1 || commits != 1 {
		t.Fatalf("synthetic route effects read/stage/plan/apply/read-back/commit = %d/%d/%d/%d/%d/%d, want 1 each", source.reads, stage.stages, destination.plans, destination.applies, destination.readBacks, commits)
	}
	if result.RecordsRead != 1 || result.RecordsStaged != 1 || result.RecordsApplied != 1 || result.CommittedCheckpoint == nil {
		t.Fatalf("synthetic transport result = %#v, want one record through a committed declaration-owned route", result)
	}
}

// TestDefinitionTransportFactoriesSelectDeclaredEvidence proves production
// composition takes reusable adapter evidence from the owning bundle rather
// than reusing the historical GitHub constant.
func TestDefinitionTransportFactoriesSelectDeclaredEvidence(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	factories, err := definitionTransportDefinitionFactories(a, a.registry)
	if err != nil {
		t.Fatal(err)
	}
	github, ok := a.registry.Get("github")
	if !ok {
		t.Fatal("GitHub connector was not registered")
	}
	source, ok := connectors.SourceTransportDescriptorOf(github)
	if !ok {
		t.Fatal("GitHub source transport declaration is unavailable")
	}
	destination, ok := connectors.DestinationTransportDescriptorOf(github)
	if !ok {
		t.Fatal("GitHub destination transport declaration is unavailable")
	}
	var sourceFactory, destinationFactory *synctransport.DefinitionFactory
	for index := range factories {
		factory := &factories[index]
		if factory.Reference == source.Executor && factory.BuildSource != nil {
			sourceFactory = factory
		}
		if factory.Reference == destination.Executor && factory.BuildDestination != nil {
			destinationFactory = factory
		}
	}
	if sourceFactory == nil || sourceFactory.SourceEvidence != source.Conformance {
		t.Fatalf("source factory evidence = %#v, want declaration %#v", sourceFactory, source.Conformance)
	}
	if destinationFactory == nil || destinationFactory.DestinationEvidence != destination.Conformance {
		t.Fatalf("destination factory evidence = %#v, want declaration %#v", destinationFactory, destination.Conformance)
	}
	registered, err := destinationFactory.BuildDestination(github)
	if err != nil {
		t.Fatalf("build GitHub destination through declared adapter factory: %v", err)
	}
	if got, want := registered.TransportExecutorReference(), issueLabelDestinationReference; got != want {
		t.Fatalf("GitHub declared destination adapter = %#v, want %#v", got, want)
	}
}

// TestDefinitionTransportFactoriesRegisterSharedSourceOnce proves a second
// bundle can select the reusable declarative source with its own evidence. A
// single executor reference is registered once; it dispatches by the request's
// declared connector rather than binding the first bundle the registry sees.
func TestDefinitionTransportFactoriesRegisterSharedSourceOnce(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	bundleFS := syntheticTransportBundleFS()
	bundleFS["synthetic/sync_transport.json"] = &fstest.MapFile{Data: []byte(`{
  "schema_version": 1,
  "source_transport": {
    "executor": {"family": "declarative_api", "id": "declarative_stream_source"},
    "eligible_streams": ["widgets"],
    "modes": ["full_append"],
    "delivery": {"idempotency": "keyed", "ordering": "source_ordered", "deletes": "not_available"},
    "conformance": {"suite": "synthetic_declarative_source", "run_id": "source_v1"}
  }
}`)}
	bundle, err := engine.Load(bundleFS, "synthetic")
	if err != nil {
		t.Fatal(err)
	}
	a.registry.Register(engine.New(bundle, nil))
	factories, err := definitionTransportDefinitionFactories(a, a.registry)
	if err != nil {
		t.Fatal(err)
	}
	factories = append(factories, localWarehouseTransportDefinitionFactories(a)...)
	connectorFactories, err := synctransport.DefinitionFactoriesFromRegistry(a.registry)
	if err != nil {
		t.Fatal(err)
	}
	factories = append(factories, connectorFactories...)
	verifier, err := synctransport.NewDefinitionConformanceVerifier(factories)
	if err != nil {
		t.Fatal(err)
	}
	transports := synctransport.NewRegistry(verifier)
	if err := synctransport.RegisterDeclaredTransports(a.registry, transports, factories); err != nil {
		t.Fatalf("register shared declarative source: %v", err)
	}
	postgres, ok := a.registry.Get("postgres")
	if !ok {
		t.Fatal("PostgreSQL connector was not registered")
	}
	if _, err := transports.Preflight(synctransport.PreflightRequest{
		Source: engine.New(bundle, nil), Destination: postgres, Stream: "widgets", Mode: synccontract.ModeFullAppend,
	}); err != nil {
		t.Fatalf("synthetic definition preflight through one shared source registration: %v", err)
	}
}

// TestDefinitionTransportFactoriesRunTypedDestinationFromDefinition proves a
// declarative connector can select one closed, named typed action as a
// destination without App, engine, orchestrator, or dispatch name-routing.
// The adapter may use only the action, input bindings, acknowledgement, and
// per-mode strategy declared in this bundle.
func TestDeclarativeTypedDestination_ReadBackProviderStateBeforeCheckpoint(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	var reads, writes int
	var readBackLocators []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/widgets":
			reads++
			if locator := request.URL.Query().Get("id"); locator != "" {
				readBackLocators = append(readBackLocators, locator)
				if locator != "receipt-widget-1" {
					writer.WriteHeader(http.StatusNotFound)
					return
				}
				_ = json.NewEncoder(writer).Encode(map[string]any{"data": []map[string]any{{"id": "widget-1", "value": "definition-owned"}}})
				return
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{"data": []map[string]any{{"id": "widget-1", "value": "definition-owned"}}})
		case request.Method == http.MethodPost && request.URL.Path == "/widgets/target":
			var record map[string]any
			if err := json.NewDecoder(request.Body).Decode(&record); err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			if !reflect.DeepEqual(record, map[string]any{"target_id": "widget-1", "value": "definition-owned"}) {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			writes++
			writer.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(writer).Encode(map[string]any{"ok": true, "receipt_id": "receipt-widget-1"})
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	bundleFS := syntheticTransportBundleFS()
	bundleFS["synthetic/metadata.json"] = &fstest.MapFile{Data: []byte(`{"name":"synthetic","display_name":"Synthetic typed destination","description":"test-only typed destination","integration_type":"api","release_stage":"ga","capabilities":{"check":true,"read":true,"write":true,"query":false,"cdc":false,"dynamic_schema":false}}`)}
	bundleFS["synthetic/streams.json"] = &fstest.MapFile{Data: []byte(`{"base":{"url":"{{ config.base_url }}","user_agent":"synthetic","headers":{},"auth":[],"pagination":{"type":"none"},"check":{"method":"GET","path":"/widgets"},"error_map":[]},"streams":[{"name":"widgets","path":"/widgets","query":{"id":{"template":"{{ query.id }}","omit_when_absent":true}},"records":{"path":"data"},"schema":"schemas/widgets.json"}]}`)}
	bundleFS["synthetic/api_surface.json"] = &fstest.MapFile{Data: []byte(`{"api":"synthetic","endpoints":[{"method":"GET","path":"/widgets","covered_by":{"stream":"widgets"}},{"method":"POST","path":"/widgets/target","covered_by":{"write":"apply_widget"}}]}`)}
	bundleFS["synthetic/schemas/widgets.json"] = &fstest.MapFile{Data: []byte(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","x-primary-key":["id"],"properties":{"id":{"type":"string"},"value":{"type":"string"}}}`)}
	bundleFS["synthetic/writes.json"] = &fstest.MapFile{Data: []byte(`{"actions":[{"name":"apply_widget","kind":"create","method":"POST","path":"/widgets/target","body_type":"json","body_fields":["target_id","value"],"idempotency_key_header":"Idempotency-Key","risk":"creates a synthetic widget target","record_schema":{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","required":["target_id","value"],"additionalProperties":false,"properties":{"target_id":{"type":"string"},"value":{"type":"string"}}}}]}`)}
	bundleFS["synthetic/sync_transport.json"] = &fstest.MapFile{Data: []byte(`{
  "schema_version": 1,
  "source_transport": {
    "executor": {"family": "declarative_api", "id": "declarative_stream_source"},
    "eligible_streams": ["widgets"],
    "modes": ["full_append"],
    "delivery": {"idempotency": "keyed", "ordering": "source_ordered", "deletes": "not_available"},
    "conformance": {"suite": "synthetic_typed_destination", "run_id": "source_v1"}
  },
  "destination_transport": {
    "executor": {"family": "declarative_api", "id": "declarative_typed_destination"},
    "eligible_actions": ["apply_widget"],
    "modes": ["full_append"],
    "delivery": {"idempotency": "keyed", "ordering": "source_ordered", "deletes": "not_available"},
    "conformance": {"suite": "synthetic_typed_destination", "run_id": "destination_v1"},
    "acknowledgement": "durable_warehouse",
    "apply_strategies": [{"mode": "full_append", "strategy": "append", "action": "apply_widget"}],
    "read_back": {
      "operation": "widgets",
	  "receipt_locator": {"response_index": 0, "body_field": "receipt_id", "query_parameter": "id", "max_value_bytes": 256, "max_pages": 1},
      "identity": [{"provider_field": "id", "expected_field": "target_id"}],
      "expected": [{"provider_field": "value", "expected_field": "value"}],
      "max_records": 1000,
      "max_attempts": 2,
      "timeout_milliseconds": 5000,
      "retry_delay_milliseconds": 1,
      "conformance": {"suite": "synthetic_typed_destination", "run_id": "destination_v1"}
    },
    "source_bindings": [{
      "executor": {"family": "declarative_api", "id": "declarative_stream_source"},
      "eligible_streams": ["widgets"],
      "record_mapping": {"kind": "input_fields", "inputs": [{"input": "target_id", "field": "id"}, {"input": "value", "field": "value"}]}
    }]
  }
}`)}
	bundle, err := engine.Load(bundleFS, "synthetic")
	if err != nil {
		t.Fatalf("load synthetic typed destination bundle: %v", err)
	}
	connector := engine.New(bundle, nil)
	a.registry.Register(connector)
	if err := a.composeTransportRegistry(); err != nil {
		t.Fatalf("compose synthetic typed destination: %v", err)
	}

	runtime := connectors.RuntimeConfig{ProjectDir: root, Config: map[string]string{"base_url": server.URL}}
	commits := 0
	result, err := synctransport.NewOrchestrator(a.transports).Run(context.Background(), synctransport.RunRequest{
		ConnectionID:       "synthetic-typed-destination",
		Generation:         1,
		Source:             connector,
		SourceRuntime:      runtime,
		Destination:        connector,
		DestinationRuntime: runtime,
		Stream:             "widgets",
		Mode:               synccontract.ModeFullAppend,
		BatchSize:          1,
		Resume: synccontract.ResumeExpectation{
			Source:           synccontract.SourceIdentity{Engine: "synthetic", AccountOrCluster: "test", ObjectScope: "widgets"},
			SourceGeneration: synccontract.OpaqueToken("synthetic-typed-destination-v1"),
		},
		Stage:    &syntheticDefinitionStage{owner: "synthetic-typed-destination"},
		Approval: syntheticTypedDestinationApproval(t, "synthetic", "apply_widget"),
		Commit: func(checkpoint synccontract.CheckpointEnvelope) error {
			commits++
			if checkpoint.CommittedAt == nil {
				return fmt.Errorf("typed destination checkpoint was not acknowledged")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("run synthetic typed destination: %v", err)
	}
	if reads != 2 || writes != 1 || commits != 1 || !reflect.DeepEqual(readBackLocators, []string{"receipt-widget-1"}) {
		t.Fatalf("synthetic typed destination effects read/write/commit = %d/%d/%d, want source read + provider read-back, one write, one commit", reads, writes, commits)
	}
	if result.RecordsRead != 1 || result.RecordsApplied != 1 || result.CommittedCheckpoint == nil {
		t.Fatalf("synthetic typed destination result = %#v, want one acknowledged action", result)
	}
}

// TestDeclarativeDestination_ClaimedKeyedWithoutIndependentProofIsRejected
// keeps PF-CF-B28 at the definition boundary: delivery.idempotency=keyed is
// only a claim. The executable action must independently declare the exact
// provider idempotency header that the sealed approval will bind.
func TestDeclarativeDestination_ClaimedKeyedWithoutIndependentProofIsRejected(t *testing.T) {
	files := declarativeTypedDestinationBundleFS("claimed-keyed-without-proof", "source-proof", "destination-proof")
	files["claimed-keyed-without-proof/writes.json"] = &fstest.MapFile{Data: []byte(strings.Replace(string(files["claimed-keyed-without-proof/writes.json"].Data), `"idempotency_key_header":"Idempotency-Key",`, "", 1))}
	bundle, err := engine.Load(files, "claimed-keyed-without-proof")
	if err != nil {
		t.Fatal(err)
	}
	_, err = declarativeTypedDestinationContractFor(engine.New(bundle, nil))
	if err == nil || !strings.Contains(err.Error(), "independent idempotency proof") {
		t.Fatalf("keyed declaration contract error = %v, want independent idempotency proof refusal", err)
	}
}

// TestDeclarativeDestination_IndependentProof proves the positive B28 path:
// the sealed proof is exact to the executor, action digest, and provider
// header, and a post-provider-success retry reuses that header rather than
// creating a second mutation.
func TestDeclarativeDestination_IndependentProof(t *testing.T) {
	name := "independent-idempotency-proof"
	files := declarativeTypedDestinationBundleFS(name, "source-proof", "destination-proof")
	bundle, err := engine.Load(files, name)
	if err != nil {
		t.Fatal(err)
	}
	connector := engine.New(bundle, nil)
	contract, err := declarativeTypedDestinationContractFor(connector)
	if err != nil {
		t.Fatalf("typed destination contract: %v", err)
	}
	digest, err := contract.actionDefinitionDigest("apply_widget")
	if err != nil {
		t.Fatal(err)
	}
	header, err := contract.idempotencyHeader("apply_widget")
	if err != nil {
		t.Fatal(err)
	}
	approval := synctransport.DestinationApproval{IdempotencyProof: synctransport.DestinationIdempotencyProof{
		Executor: contract.descriptor.Executor, ActionDefinitionSHA256: digest, EffectiveHeader: header,
	}}
	if err := validateDeclarativeTypedDestinationIdempotencyProof(approval, contract.descriptor.Executor, digest, header); err != nil {
		t.Fatalf("exact idempotency proof = %v", err)
	}
	for _, proof := range []synctransport.DestinationIdempotencyProof{
		{Executor: contract.descriptor.Executor, ActionDefinitionSHA256: strings.Repeat("0", 64), EffectiveHeader: header},
		{Executor: contract.descriptor.Executor, ActionDefinitionSHA256: digest, EffectiveHeader: "X-Other-Key"},
		{Executor: declarativeStreamSourceReference, ActionDefinitionSHA256: digest, EffectiveHeader: header},
	} {
		approval.IdempotencyProof = proof
		if err := validateDeclarativeTypedDestinationIdempotencyProof(approval, contract.descriptor.Executor, digest, header); err == nil {
			t.Fatalf("mismatched idempotency proof %#v was accepted", proof)
		}
	}

	providerCalls := 0
	mutations := 0
	keys := make(map[string]struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		providerCalls++
		key := request.Header.Get(header)
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, replay := keys[key]; !replay {
			keys[key] = struct{}{}
			mutations++
			// The provider committed the mutation but the response was ambiguous.
			// Its keyed retry must not execute the mutation again.
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	fixtureApproval := syntheticTypedDestinationApproval(t, name, "apply_widget")
	result, err := connector.Write(context.Background(), connectors.WriteRequest{
		Action: "apply_widget", Config: connectors.RuntimeConfig{ProjectDir: t.TempDir(), Config: map[string]string{"base_url": server.URL}}, Approval: fixtureApproval.Evidence,
	}, []connectors.Record{{"target_id": "widget-1", "value": "one"}})
	if err != nil {
		t.Fatalf("keyed post-success retry = %v", err)
	}
	if providerCalls != 2 || mutations != 1 || result.RecordsWritten != 1 {
		t.Fatalf("provider calls/mutations/writes = %d/%d/%d, want 2/1/1", providerCalls, mutations, result.RecordsWritten)
	}
}

// TestDeclarativeDestination_ReadBackUsesInternalReceiptLocator proves the
// receipt boundary end to end: a declaration-owned write receipt selects one
// declared, bounded read-back query while its complete locator never crosses
// the public acknowledgement boundary. The bad cases are deliberately
// admission/read-back failures before a checkpoint caller can proceed.
func TestDeclarativeDestination_ReadBackUsesInternalReceiptLocator(t *testing.T) {
	const privateLocator = "receipt-private-locator"

	t.Run("unsupported compound locator is refused before write", func(t *testing.T) {
		files := declarativeTypedDestinationBundleFS("readback-compound-refusal", "source", "destination")
		path := "readback-compound-refusal/sync_transport.json"
		files[path] = &fstest.MapFile{Data: []byte(strings.Replace(string(files[path].Data), `"response_index": 0`, `"response_index": 1`, 1))}
		bundle, err := engine.Load(files, "readback-compound-refusal")
		if err != nil {
			t.Fatalf("load bundle: %v", err)
		}
		if _, err := declarativeTypedDestinationContractFor(engine.New(bundle, nil)); err == nil || !strings.Contains(err.Error(), "response_index") {
			t.Fatalf("compound receipt locator contract error = %v, want pre-write response_index refusal", err)
		}
	})

	var writes, readBackAttempts int
	var readBackLocators []string
	var readBackPages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/widgets/target":
			writes++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"receipt_id":"receipt-private-locator","echo":"receipt-private-locator"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/widgets":
			locator := r.URL.Query().Get("id")
			readBackLocators = append(readBackLocators, locator)
			page := r.URL.Query().Get("page")
			readBackPages = append(readBackPages, page)
			if locator != privateLocator {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			switch page {
			case "1":
				readBackAttempts++
				_, _ = w.Write([]byte(`{"data":[{"id":"unrelated","value":"definition-owned"}]}`))
			case "2":
				if readBackAttempts == 1 {
					_, _ = w.Write([]byte(`{"data":[]}`))
					return
				}
				_, _ = w.Write([]byte(`{"data":[{"id":"widget-1","value":"definition-owned"}]}`))
			default:
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	const name = "readback-private-receipt"
	files := declarativeTypedDestinationBundleFS(name, "source", "destination")
	streamsPath := name + "/streams.json"
	files[streamsPath] = &fstest.MapFile{Data: []byte(strings.Replace(string(files[streamsPath].Data), `"pagination":{"type":"none"}`, `"pagination":{"type":"page_number","page_param":"page","size_param":"per_page","page_size":1}`, 1))}
	syncPath := name + "/sync_transport.json"
	files[syncPath] = &fstest.MapFile{Data: []byte(strings.Replace(string(files[syncPath].Data), `"max_pages": 1`, `"max_pages": 2`, 1))}
	bundle, err := engine.Load(files, name)
	if err != nil {
		t.Fatalf("load bundle: %v", err)
	}
	connector := engine.New(bundle, nil)
	executor := &declarativeTypedDestinationExecutor{}
	approval := syntheticTypedDestinationApproval(t, name, "apply_widget")
	strategy := connectors.DestinationApplyStrategy{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: "apply_widget"}
	plan, err := executor.PlanDestination(context.Background(), synctransport.DestinationPlanRequest{
		Connector: connector, Source: connector, Stream: "widgets", Mode: synccontract.ModeFullAppend, ApplyStrategy: strategy, Approval: approval,
	})
	if err != nil {
		t.Fatalf("plan destination: %v", err)
	}
	warehouseReceipt := synctransport.WarehouseReceipt{
		ID: "private-readback-workset", Owner: "private-readback-workset", Generation: 1, Stream: "widgets", Mode: synccontract.ModeFullAppend,
		CheckpointSHA256: "checkpoint", TombstonesSHA256: "tombstones", ManifestSHA256: "manifest", ContentSHA256: "content", ParquetSHA256: "parquet", Records: 1,
	}
	runtime := connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}, Secrets: map[string]string{"receipt_secret": privateLocator}}
	workset := synctransport.WarehouseWorkset{ID: warehouseReceipt.ID, Records: []connectors.Record{{"id": "widget-1", "value": "definition-owned"}}}
	applyRequest := synctransport.DestinationApplyRequest{
		ConnectionID: warehouseReceipt.Owner, Plan: plan, Receipt: warehouseReceipt, Workset: workset, Runtime: runtime,
		Source: connector, Destination: connector, Stream: "widgets", Mode: synccontract.ModeFullAppend, Approval: approval,
	}
	acknowledgement, err := executor.ApplyDestination(context.Background(), applyRequest)
	if err != nil {
		t.Fatalf("apply destination: %v", err)
	}
	if writes != 1 || bytes.Contains(acknowledgement.Output, []byte(privateLocator)) {
		t.Fatalf("write/public acknowledgement = %d/%q, want one write and redacted public output", writes, acknowledgement.Output)
	}
	privateReceipt, found := acknowledgement.PrivateReceipt()
	if !found || !bytes.Contains(privateReceipt, []byte(privateLocator)) {
		t.Fatal("durable acknowledgement did not retain the complete private receipt locator")
	}
	readBackRequest := synctransport.DestinationReadBackRequest{
		Plan: plan, Workset: workset, Runtime: runtime,
		Source: connector, Destination: connector, Stream: "widgets", Mode: synccontract.ModeFullAppend, Acknowledgement: acknowledgement,
	}
	if err := executor.ReadBackDestination(context.Background(), readBackRequest); err != nil {
		t.Fatalf("read back destination: %v", err)
	}
	if got, want := readBackLocators, []string{privateLocator, privateLocator, privateLocator, privateLocator}; !reflect.DeepEqual(got, want) || !reflect.DeepEqual(readBackPages, []string{"1", "2", "1", "2"}) || readBackAttempts != 2 {
		t.Fatalf("read-back locators/pages/attempts = %q/%q/%d, want one stable locator across two declared two-page attempts", got, readBackPages, readBackAttempts)
	}

	missingReceipt, err := synccontract.NewDurableDownstreamAcknowledgement(name, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	beforeMissing := len(readBackLocators)
	readBackRequest.Acknowledgement = missingReceipt
	if err := executor.ReadBackDestination(context.Background(), readBackRequest); err == nil || !strings.Contains(err.Error(), "private provider receipt") {
		t.Fatalf("missing private receipt read-back error = %v, want pre-I/O refusal", err)
	}
	if len(readBackLocators) != beforeMissing {
		t.Fatal("missing private receipt reached provider read-back")
	}

	foreignReceipt, err := connectors.NewDeclarativeTypedDestinationReadBackReceipt(plan.ActionDefinitionSHA256, connectors.DestinationReceiptLocator{
		ResponseIndex: 0, BodyField: "receipt_id", QueryParameter: "id", MaxValueBytes: 256, MaxPages: 1,
	}, []string{"foreign-receipt"}, 1000)
	if err != nil {
		t.Fatalf("construct foreign receipt: %v", err)
	}
	foreignAcknowledgement, err := missingReceipt.WithPrivateReceipt(foreignReceipt)
	if err != nil {
		t.Fatalf("attach foreign receipt: %v", err)
	}
	readBackRequest.Acknowledgement = foreignAcknowledgement
	if err := executor.ReadBackDestination(context.Background(), readBackRequest); err == nil {
		t.Fatal("foreign private receipt matched provider state")
	}
}

// TestDeclarativeTransportClone_PreservesLargeNumbers keeps source records
// lossless across the declaration-owned source -> stage -> destination
// boundary. In particular, this must not use JSON's default float64 decode
// path: provider JSON and native readers legitimately deliver distinct exact
// integer and json.Number values.
func TestDeclarativeTransportClone_PreservesLargeNumbers(t *testing.T) {
	source := connectors.Record{
		"signed":   int64(9007199254740993),
		"unsigned": ^uint64(0),
		"number":   json.Number("900719925474099312345678901234567890"),
		"raw":      json.RawMessage(`{"nested":9007199254740993}`),
		"nested": map[string]any{
			"items": []any{int64(-9007199254740993), json.Number("42.0"), []byte("bytes")},
		},
	}

	cloned, err := cloneTransportRecord(source)
	if err != nil {
		t.Fatalf("clone transport record: %v", err)
	}
	if got, ok := cloned["signed"].(int64); !ok || got != source["signed"] {
		t.Fatalf("signed clone = %#v (%T), want exact int64", cloned["signed"], cloned["signed"])
	}
	if got, ok := cloned["unsigned"].(uint64); !ok || got != source["unsigned"] {
		t.Fatalf("unsigned clone = %#v (%T), want exact uint64", cloned["unsigned"], cloned["unsigned"])
	}
	if got, ok := cloned["number"].(json.Number); !ok || got != source["number"] {
		t.Fatalf("number clone = %#v (%T), want exact json.Number", cloned["number"], cloned["number"])
	}
	if got, ok := cloned["raw"].(json.RawMessage); !ok || !bytes.Equal(got, source["raw"].(json.RawMessage)) {
		t.Fatalf("raw clone = %#v (%T), want exact raw JSON bytes", cloned["raw"], cloned["raw"])
	}
	nested := cloned["nested"].(map[string]any)
	nested["items"].([]any)[0] = int64(1)
	nested["items"].([]any)[2].([]byte)[0] = 'X'
	if got := source["nested"].(map[string]any)["items"].([]any); got[0] != int64(-9007199254740993) || string(got[2].([]byte)) != "bytes" {
		t.Fatalf("clone mutation changed source record: %#v", source)
	}
	if _, err := cloneTransportRecord(connectors.Record{"unsupported": make(chan struct{})}); err == nil {
		t.Fatal("clone accepted an unsupported mutable value")
	}
}

// TestDeclarativeReadBack_NumericSemanticEquality defines the destination
// identity policy: finite JSON/integer values compare by arbitrary-precision
// mathematical value, including scale/exponent variants (42 == 42.0), while
// strings and booleans retain their own distinct JSON types.
func TestDeclarativeReadBack_NumericSemanticEquality(t *testing.T) {
	policy := connectors.DestinationReadBackPolicy{
		Identity: []connectors.DestinationReadBackField{{ProviderField: "identity", ExpectedField: "target_identity"}},
		Expected: []connectors.DestinationReadBackField{{ProviderField: "state", ExpectedField: "target_state"}},
	}
	expected := []connectors.Record{{
		"target_identity": map[string]any{"id": int64(9007199254740993), "version": json.Number("42.0")},
		"target_state":    map[string]any{"nested": []any{json.Number("1e3"), int64(42)}},
	}}
	provider := []connectors.Record{{
		"identity": map[string]any{"id": json.Number("9007199254740993"), "version": int64(42)},
		"state":    map[string]any{"nested": []any{int64(1000), json.Number("42.0")}},
	}}
	if err := matchDeclarativeTypedDestinationProviderState(expected, provider, policy); err != nil {
		t.Fatalf("numeric semantic provider match: %v", err)
	}

	for _, testCase := range []struct {
		name  string
		value any
	}{
		{name: "unequal numeric", value: map[string]any{"nested": []any{int64(1000), int64(43)}}},
		{name: "string is not numeric", value: "42"},
		{name: "boolean is not numeric", value: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			changed := []connectors.Record{{
				"identity": provider[0]["identity"],
				"state":    testCase.value,
			}}
			if err := matchDeclarativeTypedDestinationProviderState(expected, changed, policy); err == nil {
				t.Fatalf("read-back accepted %s value %#v", testCase.name, testCase.value)
			}
		})
	}
}

func TestDeclarativeTypedDestination_MissingOrMismatchedReadBackDoesNotCommit(t *testing.T) {
	policy := connectors.DestinationReadBackPolicy{
		Identity: []connectors.DestinationReadBackField{{ProviderField: "id", ExpectedField: "target_id"}},
		Expected: []connectors.DestinationReadBackField{{ProviderField: "value", ExpectedField: "value"}},
	}
	expected := []connectors.Record{{"target_id": "widget-1", "value": "approved"}}
	for _, testCase := range []struct {
		name     string
		provider []connectors.Record
	}{
		{name: "missing", provider: nil},
		{name: "mismatch", provider: []connectors.Record{{"id": "widget-1", "value": "stale"}}},
		{name: "duplicate identity", provider: []connectors.Record{{"id": "widget-1", "value": "approved"}, {"id": "widget-1", "value": "approved"}}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			commits := 0
			if err := matchDeclarativeTypedDestinationProviderState(expected, testCase.provider, policy); err == nil {
				commits++
			}
			if commits != 0 {
				t.Fatal("missing, mismatched, or duplicate provider state admitted checkpoint commit")
			}
		})
	}
}

// TestPersistedConnectionSelectsDeclarativeTypedDestinationAction proves the
// application path, rather than a direct adapter call, owns multi-action
// selection. The action is persisted on the stream configuration; RunETL has
// no action input and can therefore neither substitute a different action nor
// turn this adapter into an arbitrary writer.
func TestPersistedConnectionSelectsDeclarativeTypedDestinationAction(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	const appendProviderResponse = "{\n  \"id\": \"provider-append-1\",\n  \"receipt_id\": \"widget-1\",\n  \"large_provider_id\": 9007199254740993,\n  \"rare_field\": {\"enabled\": true},\n  \"paid_tier\": \"enterprise\",\n  \"echoed_secret\": \"destination-secret\",\n  \"duplicate\": \"first\",\n  \"duplicate\": \"second\"\n}\n"
	// A configured scalar requires structured masking, so the public raw
	// result is the exact JSON encoder form of the masked object. BodyBytes
	// remains the verbatim provider byte count above; the companion Group-3
	// regression covers byte-for-byte raw preservation when no value is masked.
	const publicAppendProviderResponse = `{"duplicate":"second","echoed_secret":"[masked]","id":"provider-append-1","large_provider_id":9007199254740993,"paid_tier":"enterprise","rare_field":{"enabled":true},"receipt_id":"widget-1"}`
	var reads, appendWrites, replaceWrites, otherWrites int
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/widgets":
			reads++
			if locator := request.URL.Query().Get("id"); locator != "" && locator != "widget-1" {
				writer.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = writer.Write([]byte(`{"data":[{"id":"widget-1","value":"definition-owned"}]}`))
		case request.Method == http.MethodPost && request.URL.Path == "/widgets/append":
			appendWrites++
			writer.Header().Set("Content-Type", "application/json")
			writer.Header().Set("X-Provider-Request-ID", "append-request-1")
			writer.Header().Set("Authorization", "provider-response-secret")
			writer.Header().Set("X-Echoed-Credential", "destination-secret")
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(appendProviderResponse))
		case request.Method == http.MethodPost && request.URL.Path == "/widgets/replace":
			replaceWrites++
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"receipt_id":"widget-1"}`))
		case request.Method == http.MethodPost && request.URL.Path == "/widgets/other":
			otherWrites++
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"receipt_id":"widget-1"}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	primaryBundle, err := engine.Load(declarativeTypedDestinationMultiActionBundleFS("typed-multi", "multi-source", "multi-destination"), "typed-multi")
	if err != nil {
		t.Fatalf("load multi-action bundle: %v", err)
	}
	otherBundle, err := engine.Load(declarativeTypedDestinationActionBundleFS("typed-other", "other-source", "other-destination", "apply_other", "/widgets/other"), "typed-other")
	if err != nil {
		t.Fatalf("load other action bundle: %v", err)
	}
	a.registry.Register(engine.New(primaryBundle, nil))
	a.registry.Register(engine.New(otherBundle, nil))
	if err := a.composeTransportRegistry(); err != nil {
		t.Fatalf("compose production-shaped typed destinations: %v", err)
	}

	for _, connector := range []string{"typed-multi", "typed-other"} {
		if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: connector + "-source", Connector: connector, Config: map[string]string{"base_url": server.URL}}); err != nil {
			t.Fatalf("add %s source credential: %v", connector, err)
		}
		if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: connector + "-destination", Connector: connector, Config: map[string]string{"base_url": server.URL}, Secrets: map[string]string{"fixture_token": "destination-secret"}}); err != nil {
			t.Fatalf("add %s destination credential: %v", connector, err)
		}
	}

	create := func(name, connector, action string) Connection {
		t.Helper()
		connection, err := a.CreateConnection(ctx, CreateConnectionRequest{
			Name:        name,
			Source:      EndpointConfig{Connector: connector, Credential: connector + "-source"},
			Destination: EndpointConfig{Connector: connector, Credential: connector + "-destination"},
			Streams: map[string]StreamConfig{"widgets": {
				SyncMode: string(synccontract.ModeFullAppend), PrimaryKey: []string{"id"}, DestinationAction: action,
			}},
		})
		if err != nil {
			t.Fatalf("create %s connection: %v", name, err)
		}
		return connection
	}

	appendConnection := create("typed_multi_append", "typed-multi", "append_widget")
	replaceConnection := create("typed_multi_replace", "typed-multi", "replace_widget")
	otherConnection := create("typed_other", "typed-other", "apply_other")

	for _, testCase := range []struct {
		name       string
		connection Connection
		action     string
		writes     *int
	}{
		{name: "first declared action", connection: appendConnection, action: "append_widget", writes: &appendWrites},
		{name: "second declared action", connection: replaceConnection, action: "replace_widget", writes: &replaceWrites},
		{name: "one action in another connector", connection: otherConnection, action: "apply_other", writes: &otherWrites},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			plan, err := a.PlanDeclarativeTypedDestinationTransport(ctx, testCase.connection.Name, "widgets")
			if err != nil {
				t.Fatalf("plan persisted typed destination: %v", err)
			}
			if plan.Action != testCase.action || plan.TransportConnectionID != testCase.connection.ID || plan.TransportStream != "widgets" || plan.TransportActionDefinitionSHA256 == "" {
				t.Fatalf("typed destination plan = %#v, want persisted action %q and exact connection stream", plan, testCase.action)
			}
			previewed, preview, err := a.PreviewDeclarativeTypedDestinationTransport(ctx, plan.ID)
			if err != nil {
				t.Fatalf("preview persisted typed destination: %v", err)
			}
			if preview.Action != testCase.action || previewed.ApprovalToken == "" {
				t.Fatalf("typed destination preview = plan=%#v preview=%#v, want exact action and approval token", previewed, preview)
			}
			run, err := a.RunETL(ctx, RunETLRequest{
				Connection: testCase.connection.Name,
				Stream:     "widgets",
				BatchSize:  1,
				DestinationApproval: synctransport.DestinationApproval{
					PlanID: plan.ID, ApprovalToken: previewed.ApprovalToken,
					Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
				},
			})
			if err != nil {
				t.Fatalf("run persisted typed destination: %v", err)
			}
			if run.Status != "completed" || run.RecordsRead != 1 || run.RecordsLoaded != 1 || *testCase.writes != 1 {
				t.Fatalf("typed destination run = %#v writes=%d, want one acknowledged action", run, *testCase.writes)
			}
			if testCase.action != "append_widget" {
				return
			}
			if len(run.DestinationResults) != 1 {
				t.Fatalf("typed destination output count = %d, want one persisted action result", len(run.DestinationResults))
			}
			var output struct {
				RecordsWritten    int `json:"records_written"`
				ProviderResponses []struct {
					RecordIndex     int    `json:"record_index"`
					Status          int    `json:"status"`
					BodyPresent     bool   `json:"body_present"`
					BodyBytes       int    `json:"body_bytes"`
					BodyRaw         string `json:"body_raw"`
					BodyRawEncoding string `json:"body_raw_encoding"`
					Headers         map[string]struct {
						Values []string `json:"values,omitempty"`
						Masked bool     `json:"masked,omitempty"`
					} `json:"headers"`
					Body map[string]any `json:"body"`
				} `json:"provider_responses"`
			}
			if err := json.Unmarshal(run.DestinationResults[0], &output); err != nil {
				t.Fatalf("decode persisted typed destination output: %v", err)
			}
			if output.RecordsWritten != 1 || len(output.ProviderResponses) != 1 || output.ProviderResponses[0].RecordIndex != 0 || output.ProviderResponses[0].Status != http.StatusCreated {
				t.Fatal("persisted typed destination output did not retain the full provider result")
			}
			if got := output.ProviderResponses[0].Headers["X-Provider-Request-Id"].Values; !reflect.DeepEqual(got, []string{"append-request-1"}) {
				t.Fatal("ordinary provider header did not retain the request ID")
			}
			if got := output.ProviderResponses[0].Headers["Authorization"]; got.Masked || !reflect.DeepEqual(got.Values, []string{"provider-response-secret"}) {
				t.Fatal("unconfigured provider header was not preserved")
			}
			if got := output.ProviderResponses[0].Headers["X-Echoed-Credential"]; !reflect.DeepEqual(got.Values, []string{"[masked]"}) {
				t.Fatal("configured credential echo was not masked")
			}
			if got := output.ProviderResponses[0].Body; got["id"] != "provider-append-1" || got["receipt_id"] != "widget-1" || got["paid_tier"] != "enterprise" || !reflect.DeepEqual(got["rare_field"], map[string]any{"enabled": true}) || got["echoed_secret"] != "[masked]" {
				t.Fatal("ordinary provider body fields were not preserved")
			}
			provider := output.ProviderResponses[0]
			if !provider.BodyPresent {
				t.Fatal("provider raw body presence was lost")
			}
			if provider.BodyBytes != len(appendProviderResponse) {
				t.Fatalf("provider raw body byte count = %d, want exact %d", provider.BodyBytes, len(appendProviderResponse))
			}
			if provider.BodyRaw != publicAppendProviderResponse {
				t.Fatal("provider raw body lost bytes or credential masking")
			}
			if provider.BodyRawEncoding != "text" {
				t.Fatalf("provider raw body encoding = %q, want text", provider.BodyRawEncoding)
			}
			var rawOutput struct {
				ProviderResponses []struct {
					Body map[string]json.RawMessage `json:"body"`
				} `json:"provider_responses"`
			}
			if err := json.Unmarshal(run.DestinationResults[0], &rawOutput); err != nil || string(rawOutput.ProviderResponses[0].Body["large_provider_id"]) != "9007199254740993" {
				t.Fatal("large provider result identifier was not preserved as an exact JSON number")
			}
			persistedApp, err := Open(root)
			if err != nil {
				t.Fatalf("reopen app with persisted provider result: %v", err)
			}
			persisted, err := persistedApp.GetRun(run.ID)
			if err != nil {
				t.Fatalf("load persisted typed destination run: %v", err)
			}
			if len(persisted.DestinationResults) != len(run.DestinationResults) {
				t.Fatalf("persisted typed destination result count = %d, want %d", len(persisted.DestinationResults), len(run.DestinationResults))
			}
			var persistedOutput, returnedOutput any
			if err := json.Unmarshal(persisted.DestinationResults[0], &persistedOutput); err != nil {
				t.Fatalf("decode persisted typed destination result: %v", err)
			}
			if err := json.Unmarshal(run.DestinationResults[0], &returnedOutput); err != nil || !reflect.DeepEqual(persistedOutput, returnedOutput) {
				t.Fatal("persisted typed destination result did not match returned output")
			}
			cliEnvelope, err := json.Marshal(struct {
				Run Run `json:"run"`
			}{Run: persisted})
			if err != nil || !bytes.Contains(cliEnvelope, []byte(`"destination_results"`)) || !bytes.Contains(cliEnvelope, []byte(`"body_raw"`)) || !bytes.Contains(cliEnvelope, []byte(`"paid_tier":"enterprise"`)) || !bytes.Contains(cliEnvelope, []byte("[masked]")) || bytes.Contains(cliEnvelope, []byte("destination-secret")) {
				t.Fatal("CLI-shaped persisted ETL JSON lost provider truth or leaked a configured credential")
			}
		})
	}

	for _, testCase := range []struct {
		name      string
		connector string
		action    string
	}{
		{name: "missing action in multi-action connector", connector: "typed-multi", action: ""},
		{name: "other connector action", connector: "typed-multi", action: "apply_other"},
		{name: "unlisted action", connector: "typed-other", action: "replace_widget"},
	} {
		t.Run("refuses "+testCase.name, func(t *testing.T) {
			_, err := a.CreateConnection(ctx, CreateConnectionRequest{
				Name:        "reject_" + strings.ReplaceAll(testCase.name, " ", "_"),
				Source:      EndpointConfig{Connector: testCase.connector, Credential: testCase.connector + "-source"},
				Destination: EndpointConfig{Connector: testCase.connector, Credential: testCase.connector + "-destination"},
				Streams: map[string]StreamConfig{"widgets": {
					SyncMode: string(synccontract.ModeFullAppend), PrimaryKey: []string{"id"}, DestinationAction: testCase.action,
				}},
			})
			if err == nil {
				t.Fatal("persisted connection accepted an absent or foreign typed destination action")
			}
			if reads != 6 || appendWrites != 1 || replaceWrites != 1 || otherWrites != 1 {
				t.Fatalf("refused persisted selection caused I/O reads/append/replace/other=%d/%d/%d/%d, want 6/1/1/1 (source plus provider read-back)", reads, appendWrites, replaceWrites, otherWrites)
			}
		})
	}

	// Persisted state from before multi-action selection existed must also fail
	// closed on the real RunETL path. Mutating this synthetic state directly
	// models an old/manual state file; CreateConnection already refuses the same
	// absent selection above.
	for index := range a.state.Connections {
		if a.state.Connections[index].ID == appendConnection.ID {
			stream := a.state.Connections[index].Streams["widgets"]
			stream.DestinationAction = ""
			a.state.Connections[index].Streams["widgets"] = stream
			break
		}
	}
	if err := a.save(); err != nil {
		t.Fatalf("persist missing multi-action selection fixture: %v", err)
	}
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: appendConnection.Name, Stream: "widgets", BatchSize: 1}); err == nil {
		t.Fatal("RunETL accepted a persisted multi-action stream without destination_action")
	}
	if reads != 6 || appendWrites != 1 || replaceWrites != 1 || otherWrites != 1 {
		t.Fatalf("persisted missing action reached I/O reads/append/replace/other=%d/%d/%d/%d, want 6/1/1/1 (source plus provider read-back)", reads, appendWrites, replaceWrites, otherWrites)
	}
}

func TestDeclarativeTypedDestinationMappedNullDefersToSelectedActionSchema(t *testing.T) {
	name := "typed-nullable-mapping"
	files := declarativeTypedDestinationBundleFS(name, "nullable_source", "nullable_destination")
	writes := string(files[name+"/writes.json"].Data)
	writes = strings.Replace(writes, `"value":{"type":"string"}`, `"value":{"type":["string","null"]}`, 1)
	files[name+"/writes.json"] = &fstest.MapFile{Data: []byte(writes)}
	bundle, err := engine.Load(files, name)
	if err != nil {
		t.Fatalf("load nullable typed destination: %v", err)
	}
	connector := engine.New(bundle, nil)
	contract, err := declarativeTypedDestinationContractFor(connector)
	if err != nil {
		t.Fatalf("typed destination contract: %v", err)
	}
	strategy := connectors.DestinationApplyStrategy{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: "apply_widget"}
	binding, err := contract.plan(connector, "widgets", synccontract.ModeFullAppend, strategy)
	if err != nil {
		t.Fatalf("plan nullable typed destination: %v", err)
	}
	records, err := declarativeTypedDestinationRecords([]connectors.Record{{"id": "widget-1", "value": nil}}, binding)
	if err != nil {
		t.Fatalf("map present null source field: %v", err)
	}
	if value, found := records[0]["value"]; !found || value != nil {
		t.Fatalf("mapped nullable record = %#v, want present nil value", records[0])
	}
	if err := connector.ValidateWrite(context.Background(), connectors.WriteRequest{Action: strategy.Action}, records); err != nil {
		t.Fatalf("selected nullable schema rejected present null: %v", err)
	}
	if _, err := declarativeTypedDestinationRecords([]connectors.Record{{"id": "widget-1"}}, binding); err == nil {
		t.Fatal("mapping accepted an absent selected action field")
	}
}

func TestRunETLRejectsPersistedDestinationActionOnLegacyDestinationBeforeRead(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	application, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	source := &streamingSource{total: 1}
	destination := &batchDestination{}
	application.registry = connectors.NewEmptyRegistry()
	application.registry.Register(source)
	application.registry.Register(destination)
	if _, err := application.AddCredential(ctx, AddCredentialRequest{Name: "legacy-action-source", Connector: source.Name()}); err != nil {
		t.Fatalf("add source credential: %v", err)
	}
	if _, err := application.AddCredential(ctx, AddCredentialRequest{Name: "legacy-action-destination", Connector: destination.Name()}); err != nil {
		t.Fatalf("add destination credential: %v", err)
	}
	connection, err := application.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "persisted_legacy_destination_action",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "legacy-action-source"},
		Destination: EndpointConfig{Connector: destination.Name(), Credential: "legacy-action-destination"},
		Streams: map[string]StreamConfig{"records": {
			SyncMode: string(synccontract.ModeFullAppend), PrimaryKey: []string{"id"},
		}},
	})
	if err != nil {
		t.Fatalf("create legacy destination connection: %v", err)
	}
	for index := range application.state.Connections {
		if application.state.Connections[index].ID == connection.ID {
			stream := application.state.Connections[index].Streams["records"]
			stream.DestinationAction = "legacy_manual_action"
			application.state.Connections[index].Streams["records"] = stream
		}
	}
	if _, err := application.RunETL(ctx, RunETLRequest{Connection: connection.Name, Stream: "records", BatchSize: 1}); err == nil || !strings.Contains(err.Error(), "not a declarative typed destination") {
		t.Fatalf("RunETL legacy persisted destination action error = %v, want pre-read refusal", err)
	}
	if len(destination.batches) != 0 {
		t.Fatalf("legacy persisted destination action reached write path: batches=%v", destination.batches)
	}
}

func TestDeclarativeTypedDestinationRejectsActionDefinitionDriftBeforePreviewAndApply(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	application, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests++
	}))
	t.Cleanup(server.Close)
	name := "typed-definition-drift"
	initialBundle, err := engine.Load(declarativeTypedDestinationBundleFS(name, "drift_source", "drift_destination"), name)
	if err != nil {
		t.Fatalf("load initial typed destination: %v", err)
	}
	initial := engine.New(initialBundle, nil)
	application.registry.Register(initial)
	if err := application.composeTransportRegistry(); err != nil {
		t.Fatalf("compose initial typed destination: %v", err)
	}
	for _, endpoint := range []string{"source", "destination"} {
		if _, err := application.AddCredential(ctx, AddCredentialRequest{Name: name + "-" + endpoint, Connector: name, Config: map[string]string{"base_url": server.URL}}); err != nil {
			t.Fatalf("add %s credential: %v", endpoint, err)
		}
	}
	connection, err := application.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "typed_definition_drift",
		Source:      EndpointConfig{Connector: name, Credential: name + "-source"},
		Destination: EndpointConfig{Connector: name, Credential: name + "-destination"},
		Streams: map[string]StreamConfig{"widgets": {
			SyncMode: string(synccontract.ModeFullAppend), PrimaryKey: []string{"id"}, DestinationAction: "apply_widget",
		}},
	})
	if err != nil {
		t.Fatalf("create typed destination connection: %v", err)
	}
	persistedPlan, err := application.PlanDeclarativeTypedDestinationTransport(ctx, connection.Name, "widgets")
	if err != nil {
		t.Fatalf("plan typed destination: %v", err)
	}
	if persistedPlan.TransportActionDefinitionSHA256 == "" {
		t.Fatal("persisted typed destination plan has no action definition digest")
	}

	approval := syntheticTypedDestinationApproval(t, name, "apply_widget")
	resolved, err := application.transports.Preflight(synctransport.PreflightRequest{Source: initial, Destination: initial, Stream: "widgets", Mode: synccontract.ModeFullAppend, DestinationAction: "apply_widget"})
	if err != nil {
		t.Fatalf("preflight initial typed destination: %v", err)
	}
	transportPlan, err := resolved.Destination.PlanDestination(ctx, synctransport.DestinationPlanRequest{
		Connector: initial, Source: initial, Stream: "widgets", Mode: synccontract.ModeFullAppend,
		ApplyStrategy: resolved.ApplyStrategy, Approval: approval,
	})
	if err != nil || transportPlan.ActionDefinitionSHA256 == "" {
		t.Fatalf("plan typed destination apply contract = %#v err=%v", transportPlan, err)
	}

	driftFiles := declarativeTypedDestinationActionBundleFS(name, "drift_source", "drift_destination", "apply_widget", "/widgets/drift")
	driftBundle, err := engine.Load(driftFiles, name)
	if err != nil {
		t.Fatalf("load drifted typed destination: %v", err)
	}
	drifted := engine.New(driftBundle, nil)
	_, err = resolved.Destination.ApplyDestination(ctx, synctransport.DestinationApplyRequest{
		ConnectionID: "typed-definition-drift", Plan: transportPlan,
		Destination: drifted, Source: initial, Stream: "widgets", Mode: synccontract.ModeFullAppend,
		Approval: approval,
	})
	if err == nil || !strings.Contains(err.Error(), "definition changed") {
		t.Fatalf("apply drift error = %v, want pre-I/O definition drift refusal", err)
	}

	application.registry.Register(drifted)
	if err := application.composeTransportRegistry(); err != nil {
		t.Fatalf("compose drifted typed destination: %v", err)
	}
	if _, _, err := application.PreviewDeclarativeTypedDestinationTransport(ctx, persistedPlan.ID); err == nil || !strings.Contains(err.Error(), "does not bind") {
		t.Fatalf("preview drift error = %v, want plan binding refusal", err)
	}
	if requests != 0 {
		t.Fatalf("definition drift reached provider I/O; requests=%d, want 0", requests)
	}
}

func TestDeclarativeTypedDestinationRejectsEffectiveRequestDefinitionDriftBeforePreview(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	application, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests++
	}))
	t.Cleanup(server.Close)
	name := "typed-effective-definition-drift"
	requestDefinitionFiles := func(userAgent, providerDefault string) fstest.MapFS {
		files := declarativeTypedDestinationBundleFS(name, "effective_source", "effective_destination")
		files[name+"/spec.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","required":["base_url"],"properties":{"base_url":{"type":"string"},"provider_default":{"type":"string","default":%q}}}`, providerDefault))}
		streams := string(files[name+"/streams.json"].Data)
		streams = strings.Replace(streams, `"user_agent":"synthetic"`, fmt.Sprintf(`"user_agent":%q`, userAgent), 1)
		streams = strings.Replace(streams, `"headers":{}`, `"headers":{"X-Definition-Default":"{{ config.provider_default }}"}`, 1)
		files[name+"/streams.json"] = &fstest.MapFile{Data: []byte(streams)}
		return files
	}
	initialBundle, err := engine.Load(requestDefinitionFiles("synthetic-one", "one"), name)
	if err != nil {
		t.Fatalf("load initial request definition: %v", err)
	}
	initial := engine.New(initialBundle, nil)
	initialDigest, err := initial.DeclarativeTypedDestinationActionDigest("apply_widget")
	if err != nil {
		t.Fatalf("initial action digest: %v", err)
	}
	application.registry.Register(initial)
	if err := application.composeTransportRegistry(); err != nil {
		t.Fatalf("compose initial typed destination: %v", err)
	}
	for _, endpoint := range []string{"source", "destination"} {
		if _, err := application.AddCredential(ctx, AddCredentialRequest{Name: name + "-" + endpoint, Connector: name, Config: map[string]string{"base_url": server.URL}}); err != nil {
			t.Fatalf("add %s credential: %v", endpoint, err)
		}
	}
	connection, err := application.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "typed_effective_definition_drift",
		Source:      EndpointConfig{Connector: name, Credential: name + "-source"},
		Destination: EndpointConfig{Connector: name, Credential: name + "-destination"},
		Streams: map[string]StreamConfig{"widgets": {
			SyncMode: string(synccontract.ModeFullAppend), PrimaryKey: []string{"id"}, DestinationAction: "apply_widget",
		}},
	})
	if err != nil {
		t.Fatalf("create typed destination connection: %v", err)
	}
	plan, err := application.PlanDeclarativeTypedDestinationTransport(ctx, connection.Name, "widgets")
	if err != nil {
		t.Fatalf("plan typed destination: %v", err)
	}

	driftBundle, err := engine.Load(requestDefinitionFiles("synthetic-two", "two"), name)
	if err != nil {
		t.Fatalf("load drifted request definition: %v", err)
	}
	drifted := engine.New(driftBundle, nil)
	driftDigest, err := drifted.DeclarativeTypedDestinationActionDigest("apply_widget")
	if err != nil {
		t.Fatalf("drifted action digest: %v", err)
	}
	if driftDigest == initialDigest {
		t.Fatal("changed user agent and default preserved the typed action digest")
	}
	application.registry.Register(drifted)
	if err := application.composeTransportRegistry(); err != nil {
		t.Fatalf("compose drifted typed destination: %v", err)
	}
	if _, _, err := application.PreviewDeclarativeTypedDestinationTransport(ctx, plan.ID); err == nil || !strings.Contains(err.Error(), "does not bind") {
		t.Fatalf("preview request-definition drift error = %v, want pre-I/O binding refusal", err)
	}
	if requests != 0 {
		t.Fatalf("effective definition drift reached provider I/O; requests=%d, want 0", requests)
	}
}

func TestDeclarativeTypedDestinationRefusesRedirectsBeforeRetargeting(t *testing.T) {
	for _, status := range []int{
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect,
	} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			originCalls := 0
			targetCalls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/widgets/target":
					originCalls++
					http.Redirect(w, r, "/widgets/retargeted", status)
				case "/widgets/retargeted":
					targetCalls++
					w.WriteHeader(http.StatusCreated)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			name := fmt.Sprintf("typed-redirect-%d", status)
			files := declarativeTypedDestinationBundleFS(name, "redirect-source", "redirect-destination")
			writesPath := name + "/writes.json"
			files[writesPath] = &fstest.MapFile{Data: []byte(strings.Replace(string(files[writesPath].Data), `"risk":`, `"idempotency_key_header":"Idempotency-Key","risk":`, 1))}
			bundle, err := engine.Load(files, name)
			if err != nil {
				t.Fatalf("load typed destination bundle: %v", err)
			}
			connector := engine.New(bundle, nil)
			executor := &declarativeTypedDestinationExecutor{}
			approval := syntheticTypedDestinationApproval(t, name, "apply_widget")
			strategy := connectors.DestinationApplyStrategy{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: "apply_widget"}
			plan, err := executor.PlanDestination(context.Background(), synctransport.DestinationPlanRequest{
				Connector: connector, Source: connector, Stream: "widgets", Mode: synccontract.ModeFullAppend, ApplyStrategy: strategy, Approval: approval,
			})
			if err != nil {
				t.Fatalf("PlanDestination: %v", err)
			}
			receipt := synctransport.WarehouseReceipt{
				ID: "typed-redirect-workset", Owner: "typed-redirect-workset", Generation: 1, Stream: "widgets", Mode: synccontract.ModeFullAppend,
				CheckpointSHA256: "checkpoint", TombstonesSHA256: "tombstones", ManifestSHA256: "manifest", ContentSHA256: "content", ParquetSHA256: "parquet", Records: 1,
			}
			_, err = executor.ApplyDestination(context.Background(), synctransport.DestinationApplyRequest{
				ConnectionID: receipt.Owner, Plan: plan, Receipt: receipt,
				Workset: synctransport.WarehouseWorkset{ID: receipt.ID, Records: []connectors.Record{{"id": "widget-1", "value": "value"}}},
				Runtime: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}},
				Source:  connector, Destination: connector, Stream: "widgets", Mode: synccontract.ModeFullAppend, Approval: approval,
			})
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), "redirect") {
				t.Fatalf("ApplyDestination error = %v, want redirect refusal", err)
			}
			if originCalls != 1 || targetCalls != 0 {
				t.Fatalf("redirect calls origin=%d target=%d, want 1/0", originCalls, targetCalls)
			}
		})
	}
}

func TestDeclarativeTypedDestinationApprovalBindsActionDefinition(t *testing.T) {
	digest := strings.Repeat("a", 64)
	approval := synctransport.DestinationApproval{Target: connectors.WriteApprovalTarget{Scope: connectors.WriteApprovalScopeProject}}
	if err := validateDeclarativeTypedDestinationApprovalDefinition(approval, digest); err == nil || !strings.Contains(err.Error(), "does not bind") {
		t.Fatalf("missing action digest approval error = %v, want binding refusal", err)
	}
	approval.ActionDefinitionSHA256 = digest
	if err := validateDeclarativeTypedDestinationApprovalDefinition(approval, digest); err != nil {
		t.Fatalf("matching action digest approval: %v", err)
	}
	approval.Target.Scope = connectors.WriteApprovalScopeFixture
	approval.ActionDefinitionSHA256 = ""
	if err := validateDeclarativeTypedDestinationApprovalDefinition(approval, digest); err != nil {
		t.Fatalf("fixture action digest approval: %v", err)
	}
}

func TestDeclarativeTypedDestinationPersistsPartialProviderResultsOnFailedApply(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	application, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	posts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/widgets":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"widget-1","value":"first"},{"id":"widget-2","value":"second"}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/widgets/target":
			posts++
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Credential-Echo", "destination-secret")
			if posts == 1 {
				_, _ = w.Write([]byte(`{"provider_id":"first","echo":"destination-secret"}`))
				return
			}
			w.Header().Add("X-Provider-Receipt", "receipt-one")
			w.Header().Add("X-Provider-Receipt", "receipt-two")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"provider_id":"second","echo":"destination-secret"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	name := "typed-partial-results"
	bundle, err := engine.Load(declarativeTypedDestinationBundleFS(name, "partial_source", "partial_destination"), name)
	if err != nil {
		t.Fatalf("load partial-result typed destination: %v", err)
	}
	application.registry.Register(engine.New(bundle, nil))
	if err := application.composeTransportRegistry(); err != nil {
		t.Fatalf("compose partial-result typed destination: %v", err)
	}
	if _, err := application.AddCredential(ctx, AddCredentialRequest{Name: name + "-source", Connector: name, Config: map[string]string{"base_url": server.URL}}); err != nil {
		t.Fatalf("add source credential: %v", err)
	}
	if _, err := application.AddCredential(ctx, AddCredentialRequest{Name: name + "-destination", Connector: name, Config: map[string]string{"base_url": server.URL}, Secrets: map[string]string{"token": "destination-secret"}}); err != nil {
		t.Fatalf("add destination credential: %v", err)
	}
	connection, err := application.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "typed_partial_results",
		Source:      EndpointConfig{Connector: name, Credential: name + "-source"},
		Destination: EndpointConfig{Connector: name, Credential: name + "-destination"},
		Streams: map[string]StreamConfig{"widgets": {
			SyncMode: string(synccontract.ModeFullAppend), PrimaryKey: []string{"id"}, DestinationAction: "apply_widget",
		}},
	})
	if err != nil {
		t.Fatalf("create partial-result connection: %v", err)
	}
	plan, err := application.PlanDeclarativeTypedDestinationTransport(ctx, connection.Name, "widgets")
	if err != nil {
		t.Fatalf("plan partial-result connection: %v", err)
	}
	previewed, _, err := application.PreviewDeclarativeTypedDestinationTransport(ctx, plan.ID)
	if err != nil {
		t.Fatalf("preview partial-result connection: %v", err)
	}
	run, err := application.RunETL(ctx, RunETLRequest{
		Connection: connection.Name, Stream: "widgets", BatchSize: 2,
		DestinationApproval: synctransport.DestinationApproval{
			PlanID: plan.ID, ApprovalToken: previewed.ApprovalToken,
			Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP status 500") {
		t.Fatalf("RunETL error = %v, want failed provider response", err)
	}
	if run.Status != "failed" || run.RecordsLoaded != 0 || posts != 6 || len(run.DestinationResults) != 1 {
		t.Fatalf("failed partial-result run status/loaded/posts/results = %q/%d/%d/%d, want failed/0/6/1", run.Status, run.RecordsLoaded, posts, len(run.DestinationResults))
	}
	var output struct {
		RecordsWritten    int `json:"records_written"`
		RecordsFailed     int `json:"records_failed"`
		ProviderResponses []struct {
			RecordIndex int `json:"record_index"`
			Status      int `json:"status"`
			Headers     map[string]struct {
				Values []string `json:"values,omitempty"`
			} `json:"headers"`
			Body         any    `json:"body"`
			BodyEncoding string `json:"body_encoding"`
		} `json:"provider_responses"`
	}
	if err := json.Unmarshal(run.DestinationResults[0], &output); err != nil {
		t.Fatalf("decode failed partial provider result: %v", err)
	}
	var secondBody map[string]any
	secondBodyOK := false
	if len(output.ProviderResponses) == 2 {
		secondBody, secondBodyOK = output.ProviderResponses[1].Body.(map[string]any)
	}
	if output.RecordsWritten != 1 || output.RecordsFailed != 1 || len(output.ProviderResponses) != 2 || output.ProviderResponses[0].RecordIndex != 0 || output.ProviderResponses[1].RecordIndex != 1 || output.ProviderResponses[1].Status != http.StatusInternalServerError || output.ProviderResponses[1].BodyEncoding != "" || !reflect.DeepEqual(output.ProviderResponses[1].Headers["X-Provider-Receipt"].Values, []string{"receipt-one", "receipt-two"}) || !secondBodyOK || secondBody["provider_id"] != "second" || secondBody["echo"] != "[masked]" {
		t.Fatal("failed partial provider output did not retain ordered results")
	}
	serialized, err := json.Marshal(run)
	if err != nil || bytes.Contains(serialized, []byte("destination-secret")) {
		t.Fatal("failed partial provider run leaked a configured credential")
	}
}

// TestDeclarativeTypedDestinationSourceBindingsUseExactSelectedActionSchemaFields
// pins the declaration-time field boundary for the reusable destination
// adapter. Field spelling belongs to the selected writes.json action schema,
// which is the sole authority for provider-owned property names. This remains
// a schema/name check, never a dynamic record mapping or runtime action
// selector, and all refusals occur before source/provider I/O.
func TestDeclarativeTypedDestinationSourceBindingsUseExactSelectedActionSchemaFields(t *testing.T) {
	newConnector := func(t *testing.T, name, action, input string) *engine.Connector {
		t.Helper()
		files := declarativeTypedDestinationActionBundleFS(name, name+"_source", name+"_destination", action, "/widgets/"+name)
		for _, path := range []string{name + "/writes.json", name + "/sync_transport.json"} {
			files[path] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(string(files[path].Data), "target_id", input))}
		}
		bundle, err := engine.Load(files, name)
		if err != nil {
			t.Fatalf("load %s bundle: %v", name, err)
		}
		return engine.New(bundle, nil)
	}

	snake := newConnector(t, "schema-snake", "apply_snake", "target_id")
	camel := newConnector(t, "schema-camel", "apply_camel", "targetId")
	providerOwned := newConnector(t, "schema-provider-owned", "apply_provider_owned", "http")
	for _, testCase := range []struct {
		name      string
		connector *engine.Connector
		action    string
		input     string
	}{
		{name: "snake case action property", connector: snake, action: "apply_snake", input: "target_id"},
		{name: "camel case action property", connector: camel, action: "apply_camel", input: "targetId"},
		{name: "provider owned action property", connector: providerOwned, action: "apply_provider_owned", input: "http"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			contract, err := declarativeTypedDestinationContractFor(testCase.connector)
			if err != nil {
				t.Fatalf("load typed destination contract: %v", err)
			}
			binding, err := contract.plan(testCase.connector, "widgets", synccontract.ModeFullAppend, connectors.DestinationApplyStrategy{
				Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: testCase.action,
			})
			if err != nil {
				t.Fatalf("plan selected %s action: %v", testCase.action, err)
			}
			if got := binding.RecordMapping.Inputs[0].Input; got != testCase.input {
				t.Fatalf("selected action input = %q, want exact schema property %q", got, testCase.input)
			}
		})
	}

	app := &App{registry: connectors.NewEmptyRegistry()}
	app.registry.Register(snake)
	app.registry.Register(camel)
	app.registry.Register(providerOwned)
	if err := app.composeTransportRegistry(); err != nil {
		t.Fatalf("compose cross-connector typed destinations: %v", err)
	}
	for _, testCase := range []struct {
		name        string
		destination connectors.Connector
		action      string
	}{
		{name: "snake selected action", destination: snake, action: "apply_snake"},
		{name: "camel selected action", destination: camel, action: "apply_camel"},
		{name: "provider owned selected action", destination: providerOwned, action: "apply_provider_owned"},
	} {
		t.Run(testCase.name+" preflight", func(t *testing.T) {
			if _, err := app.transports.Preflight(synctransport.PreflightRequest{Source: testCase.destination, Destination: testCase.destination, Stream: "widgets", Mode: synccontract.ModeFullAppend, DestinationAction: testCase.action}); err != nil {
				t.Fatalf("preflight selected action %q: %v", testCase.action, err)
			}
		})
	}
	for _, testCase := range []struct {
		name        string
		destination connectors.Connector
		action      string
	}{
		{name: "cross connector action", destination: snake, action: "apply_camel"},
		{name: "undeclared action", destination: camel, action: "not_declared"},
	} {
		t.Run("refuses "+testCase.name, func(t *testing.T) {
			if _, err := app.transports.Preflight(synctransport.PreflightRequest{Source: testCase.destination, Destination: testCase.destination, Stream: "widgets", Mode: synccontract.ModeFullAppend, DestinationAction: testCase.action}); err == nil {
				t.Fatalf("preflight accepted %s %q", testCase.name, testCase.action)
			}
		})
	}

	// target_id is valid for the snake action but not the selected camel action.
	// The exact-schema check must refuse this cross-action mapping before any
	// workset or provider call.
	crossActionFiles := declarativeTypedDestinationActionBundleFS("schema-cross-action", "cross_source", "cross_destination", "apply_camel", "/widgets/cross")
	crossActionFiles["schema-cross-action/writes.json"] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(string(crossActionFiles["schema-cross-action/writes.json"].Data), "target_id", "targetId"))}
	crossActionBundle, err := engine.Load(crossActionFiles, "schema-cross-action")
	if err != nil {
		t.Fatalf("load cross-action bundle: %v", err)
	}
	crossActionContract, err := declarativeTypedDestinationContractFor(engine.New(crossActionBundle, nil))
	if err != nil {
		t.Fatalf("load cross-action contract: %v", err)
	}
	if _, err := crossActionContract.plan(crossActionContract.connector, "widgets", synccontract.ModeFullAppend, connectors.DestinationApplyStrategy{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: "apply_camel"}); err == nil {
		t.Fatal("camel action accepted a source binding for another action schema")
	}

	for _, testCase := range []struct {
		name          string
		input         string
		wantLoadError bool
	}{
		{name: "empty", input: "", wantLoadError: true},
		{name: "malformed", input: "target/id"},
		{name: "whitespace", input: " target_id "},
		{name: "generic", input: "genericHTTP"},
		{name: "shell", input: "shell"},
		{name: "undeclared-provider-name", input: "http"},
	} {
		t.Run("refuses "+testCase.name+" binding name", func(t *testing.T) {
			files := declarativeTypedDestinationActionBundleFS("invalid-"+testCase.name, "invalid_source", "invalid_destination", "apply_invalid", "/widgets/invalid")
			path := "invalid-" + testCase.name + "/sync_transport.json"
			files[path] = &fstest.MapFile{Data: []byte(strings.Replace(string(files[path].Data), `"input": "target_id"`, `"input": "`+testCase.input+`"`, 1))}
			bundle, err := engine.Load(files, "invalid-"+testCase.name)
			if testCase.wantLoadError {
				if err == nil {
					t.Fatalf("load accepted %s source binding input %q", testCase.name, testCase.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("load %s source binding input %q: %v", testCase.name, testCase.input, err)
			}
			contract, err := declarativeTypedDestinationContractFor(engine.New(bundle, nil))
			if err != nil {
				t.Fatalf("load %s contract: %v", testCase.name, err)
			}
			if _, err := contract.plan(contract.connector, "widgets", synccontract.ModeFullAppend, connectors.DestinationApplyStrategy{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: "apply_invalid"}); err == nil {
				t.Fatalf("plan accepted %s source binding input %q", testCase.name, testCase.input)
			}
		})
	}
}

func TestDeclarativeTypedDestinationPreflightRejectsIncompleteMappingAndFullOverwriteBeforeIO(t *testing.T) {
	ctx := context.Background()
	for _, testCase := range []struct {
		name     string
		syncMode string
		wantErr  string
		mutate   func(fstest.MapFS, string)
	}{
		{
			name:     "incomplete required mapping",
			syncMode: string(synccontract.ModeFullAppend),
			wantErr:  `required field "value" is not mapped`,
			mutate: func(files fstest.MapFS, name string) {
				path := name + "/sync_transport.json"
				files[path] = &fstest.MapFile{Data: []byte(strings.Replace(string(files[path].Data), `, {"input": "value", "field": "value"}`, "", 1))}
			},
		},
		{
			name:     "default full overwrite",
			syncMode: "",
			wantErr:  "does not implement full_overwrite",
			mutate: func(files fstest.MapFS, name string) {
				path := name + "/sync_transport.json"
				contents := strings.ReplaceAll(string(files[path].Data), `"full_append"`, `"full_overwrite"`)
				contents = strings.ReplaceAll(contents, `"strategy": "append"`, `"strategy": "replace"`)
				files[path] = &fstest.MapFile{Data: []byte(contents)}
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatal(err)
			}
			a, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			name := "preflight-" + strings.ReplaceAll(testCase.name, " ", "-")
			files := declarativeTypedDestinationBundleFS(name, name+"-source", name+"-destination")
			testCase.mutate(files, name)
			bundle, err := engine.Load(files, name)
			if err != nil {
				t.Fatalf("load %s bundle: %v", testCase.name, err)
			}

			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				requests++
			}))
			t.Cleanup(server.Close)
			a.registry.Register(engine.New(bundle, nil))
			if err := a.composeTransportRegistry(); err != nil {
				t.Fatalf("compose %s typed destination: %v", testCase.name, err)
			}
			for _, endpoint := range []string{"source", "destination"} {
				if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: name + "-" + endpoint, Connector: name, Config: map[string]string{"base_url": server.URL}}); err != nil {
					t.Fatalf("add %s credential: %v", endpoint, err)
				}
			}
			_, err = a.CreateConnection(ctx, CreateConnectionRequest{
				Name:        "reject_" + strings.ReplaceAll(testCase.name, " ", "_"),
				Source:      EndpointConfig{Connector: name, Credential: name + "-source"},
				Destination: EndpointConfig{Connector: name, Credential: name + "-destination"},
				Streams: map[string]StreamConfig{"widgets": {
					SyncMode: testCase.syncMode, PrimaryKey: []string{"id"}, DestinationAction: "apply_widget",
				}},
			})
			if err == nil || !strings.Contains(err.Error(), testCase.wantErr) {
				t.Fatalf("CreateConnection error = %v, want %q", err, testCase.wantErr)
			}
			if requests != 0 {
				t.Fatalf("CreateConnection reached source/provider I/O; requests = %d", requests)
			}
		})
	}
}

func TestDeclarativeTypedDestinationRejectsPersistedFullOverwriteDuringPreparation(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	application, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	name := "persisted-full-overwrite"
	files := declarativeTypedDestinationBundleFS(name, "overwrite_source", "overwrite_destination")
	transportPath := name + "/sync_transport.json"
	transport := strings.ReplaceAll(string(files[transportPath].Data), `"full_append"`, `"full_overwrite"`)
	files[transportPath] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(transport, `"strategy": "append"`, `"strategy": "replace"`))}
	bundle, err := engine.Load(files, name)
	if err != nil {
		t.Fatalf("load full-overwrite typed destination: %v", err)
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests++
	}))
	t.Cleanup(server.Close)
	application.registry.Register(engine.New(bundle, nil))
	if err := application.composeTransportRegistry(); err != nil {
		t.Fatalf("compose full-overwrite typed destination: %v", err)
	}
	for _, endpoint := range []string{"source", "destination"} {
		if _, err := application.AddCredential(ctx, AddCredentialRequest{Name: name + "-" + endpoint, Connector: name, Config: map[string]string{"base_url": server.URL}}); err != nil {
			t.Fatalf("add %s credential: %v", endpoint, err)
		}
	}
	application.state.Connections = append(application.state.Connections, Connection{
		ID: "connection_persisted_full_overwrite", Name: "persisted_full_overwrite",
		Source: EndpointConfig{Connector: name, Credential: name + "-source"}, Destination: EndpointConfig{Connector: name, Credential: name + "-destination"},
		Streams: map[string]StreamConfig{"widgets": {SyncMode: string(synccontract.ModeFullOverwrite), PrimaryKey: []string{"id"}, DestinationAction: "apply_widget"}},
	})
	if _, err := application.prepareDeclarativeTypedDestinationTransport(ctx, "persisted_full_overwrite", "widgets"); err == nil || !strings.Contains(err.Error(), "does not implement full_overwrite") {
		t.Fatalf("prepare persisted full-overwrite error = %v, want pre-I/O refusal", err)
	}
	if _, err := application.PlanDeclarativeTypedDestinationTransport(ctx, "persisted_full_overwrite", "widgets"); err == nil || !strings.Contains(err.Error(), "does not implement full_overwrite") {
		t.Fatalf("plan persisted full-overwrite error = %v, want pre-I/O refusal", err)
	}
	if requests != 0 {
		t.Fatalf("persisted full-overwrite preparation reached provider I/O; requests=%d, want 0", requests)
	}
}

// TestDefinitionTransportFactoriesSelectDistinctTypedDestinationEvidence proves
// a factory accepts exactly the evidence selected by each declaring bundle,
// rather than silently admitting the first destination that happened to load.
func TestDefinitionTransportFactoriesSelectDistinctTypedDestinationEvidence(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	firstBundle, err := engine.Load(declarativeTypedDestinationBundleFS("typed-destination-one", "source_one", "destination_one"), "typed-destination-one")
	if err != nil {
		t.Fatal(err)
	}
	secondBundle, err := engine.Load(declarativeTypedDestinationBundleFS("typed-destination-two", "source_two", "destination_two"), "typed-destination-two")
	if err != nil {
		t.Fatal(err)
	}
	first := engine.New(firstBundle, nil)
	second := engine.New(secondBundle, nil)
	a.registry.Register(first)
	a.registry.Register(second)

	factories, err := definitionTransportDefinitionFactories(a, a.registry)
	if err != nil {
		t.Fatal(err)
	}
	var typedFactory *synctransport.DefinitionFactory
	for index := range factories {
		if factories[index].Reference == declarativeTypedDestinationReference {
			typedFactory = &factories[index]
			break
		}
	}
	if typedFactory == nil {
		t.Fatal("declarative typed destination factory was not composed")
	}
	accepted := map[connectors.ConformanceEvidenceReference]bool{
		typedFactory.DestinationEvidence: true,
	}
	for _, evidence := range typedFactory.AcceptedDestinationEvidences {
		accepted[evidence] = true
	}
	for _, evidence := range []connectors.ConformanceEvidenceReference{
		{Suite: "typed_destination", RunID: "destination_one"},
		{Suite: "typed_destination", RunID: "destination_two"},
	} {
		if !accepted[evidence] {
			t.Fatalf("typed destination evidence %#+v was not collected from its declaring definition", evidence)
		}
	}

	if err := a.composeTransportRegistry(); err != nil {
		t.Fatalf("compose distinct typed destinations: %v", err)
	}
	for _, connector := range []connectors.Connector{first, second} {
		destination, ok := connectors.DestinationTransportDescriptorOf(connector)
		if !ok {
			t.Fatal("typed destination declaration is unavailable")
		}
		resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
			Source: connector, Destination: connector, Stream: "widgets", Mode: synccontract.ModeFullAppend,
		})
		if err != nil {
			t.Fatalf("preflight %q through shared typed destination = %v", connector.Name(), err)
		}
		if got := resolved.Destination.TransportExecutorReference(); got != declarativeTypedDestinationReference {
			t.Fatalf("%q destination executor = %#v, want %#v", connector.Name(), got, declarativeTypedDestinationReference)
		}
		request := synctransport.DestinationPlanRequest{
			Connector: connector, Source: connector, Stream: "widgets", Mode: synccontract.ModeFullAppend,
			ApplyStrategy: destination.ApplyStrategies[0],
		}
		if _, err := resolved.Destination.PlanDestination(context.Background(), request); err == nil {
			t.Fatalf("plan %q accepted missing approval before source I/O", connector.Name())
		}
		request.Approval = syntheticTypedDestinationApproval(t, connector.Name(), destination.ApplyStrategies[0].Action)
		if _, err := resolved.Destination.PlanDestination(context.Background(), request); err != nil {
			t.Fatalf("plan %q through shared typed destination = %v", connector.Name(), err)
		}
	}
}

// TestDefinitionTransportFactoriesRefuseTypedDestinationDeclarationsBeforeIO
// keeps the generic adapter closed: a declaration may select an already typed
// action, but it cannot turn a malformed mapping, a foreign adapter, an
// unknown executor, or capture mode into destination I/O.
func TestDefinitionTransportFactoriesRefuseTypedDestinationDeclarationsBeforeIO(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(fstest.MapFS, string)
	}{
		{
			name: "unknown executor",
			mutate: func(files fstest.MapFS, name string) {
				path := name + "/sync_transport.json"
				files[path] = &fstest.MapFile{Data: []byte(strings.Replace(string(files[path].Data), `"id": "declarative_typed_destination"`, `"id": "unknown_typed_destination"`, 1))}
			},
		},
		{
			name: "unavailable action",
			mutate: func(files fstest.MapFS, name string) {
				path := name + "/sync_transport.json"
				files[path] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(string(files[path].Data), "apply_widget", "missing_action"))}
			},
		},
		{
			name: "wrong source mapping role",
			mutate: func(files fstest.MapFS, name string) {
				path := name + "/sync_transport.json"
				files[path] = &fstest.MapFile{Data: []byte(strings.Replace(string(files[path].Data), `"record_mapping": {"kind": "input_fields", "inputs": [{"input": "target_id", "field": "id"}, {"input": "value", "field": "value"}]}`, `"record_mapping": {"kind": "config_match", "config_key": "configured_widget", "record_field": "id"}`, 1))}
			},
		},
		{
			name: "change capture destination",
			mutate: func(files fstest.MapFS, name string) {
				path := name + "/sync_transport.json"
				contents := strings.ReplaceAll(string(files[path].Data), `"full_append"`, `"change_capture"`)
				files[path] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(contents, `"strategy": "append"`, `"strategy": "change_apply"`))}
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			name := "typed-destination-refusal-" + strings.ReplaceAll(testCase.name, " ", "-")
			files := declarativeTypedDestinationBundleFS(name, "source_refusal", "destination_refusal")
			testCase.mutate(files, name)
			bundle, loadErr := engine.Load(files, name)
			if loadErr == nil {
				application := &App{registry: connectors.NewEmptyRegistry()}
				application.registry.Register(engine.New(bundle, nil))
				loadErr = application.composeTransportRegistry()
				if application.transports != nil {
					t.Fatal("refused typed destination mutated the transport registry")
				}
			}
			if loadErr == nil {
				t.Fatal("malformed typed destination declaration was accepted")
			}
		})
	}
}

func TestDeclarativeTypedDestinationAcceptsDefinitionOwningNativeConnector(t *testing.T) {
	name := "native-typed-destination"
	files := declarativeTypedDestinationBundleFS(name, "native_source", "native_destination")
	transport := string(files[name+"/sync_transport.json"].Data)
	sourceStart := strings.Index(transport, `  "source_transport":`)
	destinationStart := strings.Index(transport, `  "destination_transport":`)
	if sourceStart < 0 || destinationStart < 0 {
		t.Fatal("native typed destination fixture has no removable source declaration")
	}
	files[name+"/sync_transport.json"] = &fstest.MapFile{Data: []byte(transport[:sourceStart] + transport[destinationStart:])}
	bundle, err := engine.Load(files, name)
	if err != nil {
		t.Fatalf("load native typed destination bundle: %v", err)
	}
	sourceBundle, err := engine.Load(declarativeTypedDestinationBundleFS("native-typed-source", "native_source", "native_destination"), "native-typed-source")
	if err != nil {
		t.Fatalf("load native typed destination source: %v", err)
	}
	source := engine.New(sourceBundle, nil)
	native := &nativeTypedDestinationTestConnector{Base: engine.NewBase(bundle)}
	contract, err := declarativeTypedDestinationContractFor(native)
	if err != nil {
		t.Fatalf("native typed destination contract: %v", err)
	}
	strategy := connectors.DestinationApplyStrategy{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: "apply_widget"}
	if _, err := contract.plan(source, "widgets", synccontract.ModeFullAppend, strategy); err != nil {
		t.Fatalf("native typed destination plan: %v", err)
	}
	application := &App{registry: connectors.NewEmptyRegistry()}
	application.registry.Register(source)
	application.registry.Register(native)
	if err := application.composeTransportRegistry(); err != nil {
		t.Fatalf("compose native typed destination: %v", err)
	}
	if _, err := application.transports.Preflight(synctransport.PreflightRequest{Source: source, Destination: native, Stream: "widgets", Mode: synccontract.ModeFullAppend, DestinationAction: strategy.Action}); err != nil {
		t.Fatalf("preflight native typed destination: %v", err)
	}
}

func TestDeclarativeTypedDestinationRefusesInvalidWorksetsBeforeProviderWrite(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		mutate  func(fstest.MapFS, string)
		record  connectors.Record
		tombs   []synccontract.Tombstone
		planErr bool
	}{
		{
			name: "mapped record misses required action field",
			mutate: func(files fstest.MapFS, name string) {
				path := name + "/sync_transport.json"
				files[path] = &fstest.MapFile{Data: []byte(strings.Replace(string(files[path].Data), `"inputs": [{"input": "target_id", "field": "id"}, {"input": "value", "field": "value"}]`, `"inputs": [{"input": "target_id", "field": "id"}]`, 1))}
			},
			record:  connectors.Record{"id": "widget-1", "value": "not-mapped"},
			planErr: true,
		},
		{
			name:   "tombstone delete",
			record: connectors.Record{"id": "widget-1", "value": "definition-owned"},
			tombs:  []synccontract.Tombstone{{}},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatal(err)
			}
			a, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodPost {
					writes++
				}
				writer.WriteHeader(http.StatusNoContent)
			}))
			t.Cleanup(server.Close)

			name := "typed-destination-workset-" + strings.ReplaceAll(testCase.name, " ", "-")
			files := declarativeTypedDestinationBundleFS(name, "source_workset", "destination_workset")
			if testCase.mutate != nil {
				testCase.mutate(files, name)
			}
			bundle, err := engine.Load(files, name)
			if err != nil {
				t.Fatalf("load typed destination bundle: %v", err)
			}
			connector := engine.New(bundle, nil)
			a.registry.Register(connector)
			if err := a.composeTransportRegistry(); err != nil {
				t.Fatalf("compose typed destination: %v", err)
			}
			resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
				Source: connector, Destination: connector, Stream: "widgets", Mode: synccontract.ModeFullAppend,
			})
			if err != nil {
				t.Fatalf("preflight typed destination: %v", err)
			}
			approval := syntheticTypedDestinationApproval(t, connector.Name(), "apply_widget")
			plan, err := resolved.Destination.PlanDestination(context.Background(), synctransport.DestinationPlanRequest{
				Connector: connector, Source: connector, Stream: "widgets", Mode: synccontract.ModeFullAppend,
				ApplyStrategy: resolved.ApplyStrategy, Approval: approval,
			})
			if err != nil {
				if testCase.planErr {
					if writes != 0 {
						t.Fatalf("provider write calls = %d, want 0 after pre-I/O plan refusal", writes)
					}
					return
				}
				t.Fatalf("plan typed destination: %v", err)
			}
			if testCase.planErr {
				t.Fatal("invalid typed destination mapping was planned")
			}
			receipt := synctransport.WarehouseReceipt{
				ID: "typed-destination-workset", Owner: "typed-destination-workset", Generation: 1, Stream: "widgets", Mode: synccontract.ModeFullAppend,
				CheckpointSHA256: "checkpoint", TombstonesSHA256: "tombstones", ManifestSHA256: "manifest", ContentSHA256: "content", ParquetSHA256: "parquet",
				Records: 1, Tombstones: len(testCase.tombs),
			}
			_, err = resolved.Destination.ApplyDestination(context.Background(), synctransport.DestinationApplyRequest{
				ConnectionID: receipt.Owner, Plan: plan, Receipt: receipt,
				Workset: synctransport.WarehouseWorkset{ID: receipt.ID, Records: []connectors.Record{testCase.record}, Tombstones: testCase.tombs},
				Runtime: connectors.RuntimeConfig{ProjectDir: root, Config: map[string]string{"base_url": server.URL}},
				Source:  connector, Destination: connector, Stream: "widgets", Mode: synccontract.ModeFullAppend,
				Approval: approval,
			})
			if err == nil {
				t.Fatal("invalid typed destination workset was applied")
			}
			if writes != 0 {
				t.Fatalf("provider write calls = %d, want 0 after pre-I/O workset refusal", writes)
			}
		})
	}
}

func declarativeTypedDestinationBundleFS(name, sourceRunID, destinationRunID string) fstest.MapFS {
	base := syntheticTransportBundleFS()
	files := make(fstest.MapFS, len(base))
	for path, file := range base {
		files[name+"/"+strings.TrimPrefix(path, "synthetic/")] = file
	}
	files[name+"/metadata.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{"name":%q,"display_name":"Synthetic typed destination","description":"test-only typed destination","integration_type":"api","release_stage":"ga","capabilities":{"check":true,"read":true,"write":true,"query":false,"cdc":false,"dynamic_schema":false}}`, name))}
	files[name+"/streams.json"] = &fstest.MapFile{Data: []byte(`{"base":{"url":"{{ config.base_url }}","user_agent":"synthetic","headers":{},"auth":[],"pagination":{"type":"none"},"check":{"method":"GET","path":"/widgets"},"error_map":[]},"streams":[{"name":"widgets","path":"/widgets","query":{"id":{"template":"{{ query.id }}","omit_when_absent":true}},"records":{"path":"data"},"schema":"schemas/widgets.json"}]}`)}
	files[name+"/api_surface.json"] = &fstest.MapFile{Data: []byte(`{"api":"synthetic","endpoints":[{"method":"GET","path":"/widgets","covered_by":{"stream":"widgets"}},{"method":"POST","path":"/widgets/target","covered_by":{"write":"apply_widget"}}]}`)}
	files[name+"/schemas/widgets.json"] = &fstest.MapFile{Data: []byte(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","x-primary-key":["id"],"properties":{"id":{"type":"string"},"value":{"type":"string"}}}`)}
	files[name+"/writes.json"] = &fstest.MapFile{Data: []byte(`{"actions":[{"name":"apply_widget","kind":"create","method":"POST","path":"/widgets/target","body_type":"json","body_fields":["target_id","value"],"idempotency_key_header":"Idempotency-Key","risk":"creates a synthetic widget target","record_schema":{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","required":["target_id","value"],"additionalProperties":false,"properties":{"target_id":{"type":"string"},"value":{"type":"string"}}}}]}`)}
	files[name+"/sync_transport.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{
  "schema_version": 1,
  "source_transport": {
    "executor": {"family": "declarative_api", "id": "declarative_stream_source"},
    "eligible_streams": ["widgets"],
    "modes": ["full_append"],
    "delivery": {"idempotency": "keyed", "ordering": "source_ordered", "deletes": "not_available"},
    "conformance": {"suite": "typed_destination", "run_id": %q}
  },
  "destination_transport": {
    "executor": {"family": "declarative_api", "id": "declarative_typed_destination"},
    "eligible_actions": ["apply_widget"],
    "modes": ["full_append"],
    "delivery": {"idempotency": "keyed", "ordering": "source_ordered", "deletes": "not_available"},
    "conformance": {"suite": "typed_destination", "run_id": %q},
    "acknowledgement": "durable_warehouse",
    "apply_strategies": [{"mode": "full_append", "strategy": "append", "action": "apply_widget"}],
    "read_back": {
      "operation": "widgets",
	  "receipt_locator": {"response_index": 0, "body_field": "receipt_id", "query_parameter": "id", "max_value_bytes": 256, "max_pages": 1},
      "identity": [{"provider_field": "id", "expected_field": "target_id"}],
      "expected": [{"provider_field": "value", "expected_field": "value"}],
      "max_records": 1000,
      "max_attempts": 2,
      "timeout_milliseconds": 5000,
      "retry_delay_milliseconds": 1,
      "conformance": {"suite": "typed_destination", "run_id": %q}
    },
    "source_bindings": [{
      "executor": {"family": "declarative_api", "id": "declarative_stream_source"},
      "eligible_streams": ["widgets"],
      "record_mapping": {"kind": "input_fields", "inputs": [{"input": "target_id", "field": "id"}, {"input": "value", "field": "value"}]}
    }]
  }
}`, sourceRunID, destinationRunID, destinationRunID))}
	return files
}

func declarativeTypedDestinationActionBundleFS(name, sourceRunID, destinationRunID, action, path string) fstest.MapFS {
	files := declarativeTypedDestinationBundleFS(name, sourceRunID, destinationRunID)
	for _, filename := range []string{name + "/writes.json", name + "/api_surface.json", name + "/sync_transport.json"} {
		contents := strings.ReplaceAll(string(files[filename].Data), "apply_widget", action)
		contents = strings.ReplaceAll(contents, "/widgets/target", path)
		files[filename] = &fstest.MapFile{Data: []byte(contents)}
	}
	return files
}

func declarativeTypedDestinationMultiActionBundleFS(name, sourceRunID, destinationRunID string) fstest.MapFS {
	files := declarativeTypedDestinationBundleFS(name, sourceRunID, destinationRunID)
	files[name+"/writes.json"] = &fstest.MapFile{Data: []byte(`{"actions":[
  {"name":"append_widget","kind":"create","method":"POST","path":"/widgets/append","body_type":"json","body_fields":["target_id","value"],"idempotency_key_header":"Idempotency-Key","risk":"creates a synthetic widget target","confirm":"destructive","record_schema":{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","required":["target_id","value"],"additionalProperties":false,"properties":{"target_id":{"type":"string"},"value":{"type":"string"}}}},
  {"name":"replace_widget","kind":"update","method":"POST","path":"/widgets/replace","body_type":"json","body_fields":["target_id","value"],"idempotency_key_header":"Idempotency-Key","risk":"replaces a synthetic widget target","confirm":"destructive","record_schema":{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","required":["target_id","value"],"additionalProperties":false,"properties":{"target_id":{"type":"string"},"value":{"type":"string"}}}}
]}`)}
	files[name+"/api_surface.json"] = &fstest.MapFile{Data: []byte(`{"api":"synthetic","endpoints":[{"method":"GET","path":"/widgets","covered_by":{"stream":"widgets"}},{"method":"POST","path":"/widgets/append","covered_by":{"write":"append_widget"}},{"method":"POST","path":"/widgets/replace","covered_by":{"write":"replace_widget"}}]}`)}
	files[name+"/sync_transport.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{
  "schema_version": 1,
  "source_transport": {
    "executor": {"family": "declarative_api", "id": "declarative_stream_source"},
    "eligible_streams": ["widgets"],
    "modes": ["full_append"],
    "delivery": {"idempotency": "keyed", "ordering": "source_ordered", "deletes": "not_available"},
    "conformance": {"suite": "typed_destination", "run_id": %q}
  },
  "destination_transport": {
    "executor": {"family": "declarative_api", "id": "declarative_typed_destination"},
    "eligible_actions": ["append_widget", "replace_widget"],
    "modes": ["full_append"],
    "delivery": {"idempotency": "keyed", "ordering": "source_ordered", "deletes": "not_available"},
    "conformance": {"suite": "typed_destination", "run_id": %q},
    "acknowledgement": "durable_warehouse",
    "apply_strategies": [
      {"mode": "full_append", "strategy": "append", "action": "append_widget"},
      {"mode": "full_append", "strategy": "append", "action": "replace_widget"}
    ],
    "read_back": {
      "operation": "widgets",
	  "receipt_locator": {"response_index": 0, "body_field": "receipt_id", "query_parameter": "id", "max_value_bytes": 256, "max_pages": 1},
      "identity": [{"provider_field": "id", "expected_field": "target_id"}],
      "expected": [{"provider_field": "value", "expected_field": "value"}],
      "max_records": 1000,
      "max_attempts": 2,
      "timeout_milliseconds": 5000,
      "retry_delay_milliseconds": 1,
      "conformance": {"suite": "typed_destination", "run_id": %q}
    },
    "source_bindings": [{
      "executor": {"family": "declarative_api", "id": "declarative_stream_source"},
      "eligible_streams": ["widgets"],
      "record_mapping": {"kind": "input_fields", "inputs": [{"input": "target_id", "field": "id"}, {"input": "value", "field": "value"}]}
    }]
  }
}`, sourceRunID, destinationRunID, destinationRunID))}
	return files
}

func syntheticTypedDestinationApproval(t *testing.T, connector, action string) synctransport.DestinationApproval {
	t.Helper()
	authority, err := connectors.NewFixtureWriteApprovalAuthority()
	if err != nil {
		t.Fatalf("create fixture write approval authority: %v", err)
	}
	confirmation := connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
	target := connectors.WriteApprovalTarget{
		Connector:           connector,
		Operation:           action,
		Method:              http.MethodPost,
		MutationClass:       "transport",
		TargetDigest:        strings.Repeat("a", 64),
		CredentialRevision:  "fixture-credential-revision",
		ConfigurationDigest: "fixture-configuration-digest",
		Batchable:           true,
		Scope:               connectors.WriteApprovalScopeFixture,
		Confirmation:        confirmation,
	}
	request := connectors.WriteApprovalGrantRequest{
		PlanID:        "rplan_synthetic_typed_destination",
		PlanHash:      strings.Repeat("b", 64),
		Mode:          string(synccontract.ModeFullAppend),
		PreviewDigest: strings.Repeat("c", 64),
		ApprovalToken: "fixture-approval-token",
		Target:        target,
		Confirmation:  confirmation,
	}
	grant, err := authority.IssueWriteGrant(request)
	if err != nil {
		t.Fatalf("issue fixture write grant: %v", err)
	}
	evidence, err := authority.VerifyWriteGrant(grant, connectors.WriteApprovalExpectation{
		PlanID: request.PlanID, PlanHash: request.PlanHash, Mode: request.Mode,
		PreviewDigest: request.PreviewDigest, ApprovalToken: request.ApprovalToken,
		Target: target, Confirmation: confirmation,
	})
	if err != nil {
		t.Fatalf("verify fixture write grant: %v", err)
	}
	return synctransport.DestinationApproval{
		Evidence:      evidence,
		Target:        target,
		PreviewDigest: request.PreviewDigest,
		AuthorizeNextUnit: func(context.Context) error {
			return nil
		},
	}
}

type syntheticDefinitionConnector struct {
	*engine.Connector
	factories []synctransport.DefinitionFactory
}

type nativeTypedDestinationTestConnector struct {
	engine.Base
}

func (*nativeTypedDestinationTestConnector) ReadBackDeclarativeDestination(context.Context, connectors.DeclarativeTypedDestinationReadBackRequest) ([]connectors.Record, error) {
	return []connectors.Record{}, nil
}

func (*nativeTypedDestinationTestConnector) Check(context.Context, connectors.RuntimeConfig) error {
	return nil
}

func (*nativeTypedDestinationTestConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{}, nil
}

func (*nativeTypedDestinationTestConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return nil
}

func (*nativeTypedDestinationTestConnector) Write(_ context.Context, _ connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsWritten: len(records)}, nil
}

func (c *syntheticDefinitionConnector) SyncTransportDefinitionFactories() []synctransport.DefinitionFactory {
	return append([]synctransport.DefinitionFactory(nil), c.factories...)
}

type syntheticDefinitionSource struct {
	reference connectors.TransportExecutorReference
	page      synctransport.SourcePage
	reads     int
}

func (s *syntheticDefinitionSource) TransportExecutorReference() connectors.TransportExecutorReference {
	return s.reference
}

func (s *syntheticDefinitionSource) ReadTransport(ctx context.Context, _ synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	s.reads++
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(s.page)
}

type syntheticDefinitionDestination struct {
	reference connectors.TransportExecutorReference
	sink      string
	plans     int
	applies   int
	readBacks int
}

func (d *syntheticDefinitionDestination) TransportExecutorReference() connectors.TransportExecutorReference {
	return d.reference
}

func (d *syntheticDefinitionDestination) PlanDestination(_ context.Context, request synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	d.plans++
	return synctransport.DestinationPlan{ApplyStrategy: request.ApplyStrategy}, nil
}

func (d *syntheticDefinitionDestination) ApplyDestination(_ context.Context, request synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	d.applies++
	if got, want := request.Workset.Records, []connectors.Record{{"id": "synthetic-1", "value": "definition-owned"}}; !reflect.DeepEqual(got, want) {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("synthetic workset = %#v, want %#v", got, want)
	}
	return synccontract.NewDurableDownstreamAcknowledgement(d.sink, time.Now().UTC())
}

func (d *syntheticDefinitionDestination) ReadBackDestination(_ context.Context, request synctransport.DestinationReadBackRequest) error {
	d.readBacks++
	if request.Acknowledgement.Sink != d.sink || request.Acknowledgement.AcknowledgedAt.IsZero() {
		return fmt.Errorf("synthetic destination acknowledgement is not durable")
	}
	return nil
}

type syntheticDefinitionStage struct {
	stages  int
	owner   string
	workset synctransport.WarehouseWorkset
}

func (s *syntheticDefinitionStage) Stage(_ context.Context, request synctransport.WarehouseStageRequest) (synctransport.WarehouseReceipt, error) {
	s.stages++
	owner := s.owner
	if owner == "" {
		owner = "synthetic-connection"
	}
	s.workset = synctransport.WarehouseWorkset{
		ID:                  "synthetic-stage",
		Records:             request.Page.Records,
		Tombstones:          request.Page.Tombstones,
		CandidateCheckpoint: request.Page.CandidateCheckpoint.Clone(),
	}
	return synctransport.WarehouseReceipt{
		ID: "synthetic-stage", Owner: owner, Generation: request.Generation, Stream: request.Stream, Mode: request.Mode,
		CheckpointSHA256: "synthetic-checkpoint", TombstonesSHA256: "synthetic-tombstones", ManifestSHA256: "synthetic-manifest", ContentSHA256: "synthetic-content", ParquetSHA256: "synthetic-parquet",
		Records: len(request.Page.Records), Tombstones: len(request.Page.Tombstones),
	}, nil
}

func (s *syntheticDefinitionStage) Reopen(_ context.Context, receipt synctransport.WarehouseReceipt) (synctransport.WarehouseWorkset, error) {
	if receipt.ID != s.workset.ID {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("synthetic stage receipt %q is unknown", receipt.ID)
	}
	return s.workset, nil
}

func syntheticDefinitionCheckpoint() synccontract.CheckpointEnvelope {
	positionObserved := true
	observedAt := time.Now().UTC().Add(-time.Minute)
	return synccontract.CheckpointEnvelope{
		StateVersion: synccontract.StateVersion,
		Source:       synccontract.SourceIdentity{Engine: "synthetic", AccountOrCluster: "fixture", ObjectScope: "widgets"},
		Mechanism:    "synthetic_definition", SnapshotBarrier: &synccontract.SnapshotBarrier{Kind: "fixture", Token: synccontract.OpaqueToken("barrier")},
		Position: synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("1"), TieBreaker: synccontract.OpaqueToken("1")}, PositionObserved: &positionObserved,
		Partitions: []synccontract.PartitionState{}, SourceGeneration: synccontract.OpaqueToken("generation"), SchemaVersion: "synthetic-v1", ProtocolVersion: "synthetic-v1",
		Dedupe:       synccontract.DedupeIdentity{Kind: "synthetic", Value: synccontract.OpaqueToken("synthetic-1")},
		DedupeWindow: synccontract.DedupeWindow{Kind: "synthetic", Start: synccontract.OpaqueToken("start"), End: synccontract.OpaqueToken("end")},
		ObservedAt:   observedAt,
	}
}

func syntheticTransportBundleFS() fstest.MapFS {
	return fstest.MapFS{
		"synthetic/metadata.json":        &fstest.MapFile{Data: []byte(`{"name":"synthetic","display_name":"Synthetic","description":"test-only definition-owned connector","integration_type":"api","release_stage":"ga","capabilities":{"check":true,"read":true,"write":true,"query":false,"cdc":false,"dynamic_schema":false}}`)},
		"synthetic/spec.json":            &fstest.MapFile{Data: []byte(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","required":["base_url"],"properties":{"base_url":{"type":"string"}}}`)},
		"synthetic/streams.json":         &fstest.MapFile{Data: []byte(`{"base":{"url":"{{ config.base_url }}","user_agent":"synthetic","headers":{},"auth":[],"pagination":{"type":"none"},"check":{"method":"GET","path":"/ping"},"error_map":[]},"streams":[{"name":"widgets","path":"/widgets","records":{"path":"data"},"schema":"schemas/widgets.json"}]}`)},
		"synthetic/api_surface.json":     &fstest.MapFile{Data: []byte(`{"api":"synthetic","endpoints":[{"method":"GET","path":"/widgets","covered_by":{"stream":"widgets"}}]}`)},
		"synthetic/schemas/widgets.json": &fstest.MapFile{Data: []byte(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","x-primary-key":["id"],"properties":{"id":{"type":"string"}}}`)},
		"synthetic/docs.md":              &fstest.MapFile{Data: []byte("# Overview\n\nsynthetic\n\n## Auth setup\n\nnone\n\n## Streams notes\n\nnone\n\n## Write actions & risks\n\nnone\n\n## Known limits\n\nnone\n")},
		"synthetic/sync_transport.json":  &fstest.MapFile{Data: []byte(`{"schema_version":1,"source_transport":{"executor":{"family":"native_api","id":"synthetic_snapshot_source"},"eligible_streams":["widgets"],"modes":["full_append"],"delivery":{"idempotency":"keyed","ordering":"source_ordered","deletes":"not_available"},"conformance":{"suite":"synthetic_transport","run_id":"source_v1"}},"destination_transport":{"executor":{"family":"native_api","id":"synthetic_stage_destination"},"eligible_actions":["stage_append"],"modes":["full_append"],"delivery":{"idempotency":"keyed","ordering":"source_ordered","deletes":"not_available"},"conformance":{"suite":"synthetic_transport","run_id":"destination_v1"},"acknowledgement":"durable_warehouse","apply_strategies":[{"mode":"full_append","strategy":"append","action":"stage_append"}]}}`)},
	}
}

// TestOpenSelectsPostgresIssueLabelDestinationTransport is the route-level
// regression test for the database-to-API quadrant. The source/destination
// pair is already definition-preflightable; persisted connection dispatch must
// also admit PostgreSQL's polling-watermark rows to GitHub's two-action label
// destination rather than reserving it for a GitHub-to-GitHub demo.
func TestOpenSelectsPostgresIssueLabelDestinationTransport(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	postgres, ok := a.registry.Get("postgres")
	if !ok {
		t.Fatal("PostgreSQL connector is not registered")
	}
	github, ok := a.registry.Get("github")
	if !ok {
		t.Fatal("GitHub connector is not registered")
	}

	for _, testCase := range []struct {
		name       string
		mode       synccontract.Mode
		consentKey string
	}{
		{name: "append maps to add issue labels", mode: synccontract.ModeFullAppend},
		{name: "keyed maps to set issue labels", mode: synccontract.ModeIncrementalUpsert, consentKey: issueLabelTransportKeyedConsentConfig},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			connection := Connection{
				ID:     "conn_postgres_github_" + strings.ReplaceAll(string(testCase.mode), "_", "-"),
				Name:   "postgres_github_" + strings.ReplaceAll(string(testCase.mode), "_", "-"),
				Source: EndpointConfig{Connector: "postgres"},
				Destination: EndpointConfig{Connector: "github", Config: map[string]string{
					issueLabelTransportTargetIssueConfig: "200",
					issueLabelTransportLabelConfig:       "from-postgres",
					testCase.consentKey:                  "true",
				}},
				Streams: map[string]StreamConfig{
					"public.issue_label_events": {
						SyncMode:         string(testCase.mode),
						CursorField:      "sequence",
						PrimaryKey:       []string{"id"},
						DestinationTable: "issue_label_events",
					},
				},
			}
			if testCase.consentKey == "" {
				delete(connection.Destination.Config, "")
			}
			a.state.Connections = append(a.state.Connections, connection)
			if !a.shouldRunTransport(connection, "public.issue_label_events", SyncMode{ContractMode: testCase.mode}, postgres, github) {
				t.Fatalf("PostgreSQL-to-GitHub %q route was not selected for declared transport dispatch", testCase.mode)
			}
		})
	}
}

func TestOpenPreflightsEveryDeclaredPostgresDestinationMode(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	postgres, ok := a.registry.Get("postgres")
	if !ok {
		t.Fatal("PostgreSQL connector is not registered")
	}
	destination, ok := connectors.DestinationTransportDescriptorOf(postgres)
	if !ok {
		t.Fatal("PostgreSQL destination transport is not declared")
	}
	wantSource := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_polling_watermark"}
	wantDestination := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_managed_target"}
	var fullOverwrite synctransport.ResolvedTransport
	for _, mode := range destination.Modes {
		t.Run(string(mode), func(t *testing.T) {
			resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
				Source:      postgres,
				Destination: postgres,
				Stream:      "snapshot",
				Mode:        mode,
			})
			if err != nil {
				t.Fatalf("production PostgreSQL preflight for %q = %v", mode, err)
			}
			if got := resolved.Source.TransportExecutorReference(); got != wantSource {
				t.Fatalf("production source for %q = %+v, want %+v", mode, got, wantSource)
			}
			if got := resolved.Destination.TransportExecutorReference(); got != wantDestination {
				t.Fatalf("production destination for %q = %+v, want %+v", mode, got, wantDestination)
			}
			if mode == synccontract.ModeFullOverwrite {
				fullOverwrite = resolved
			}
		})
	}
	if fullOverwrite.Source == nil {
		t.Fatal("PostgreSQL destination declaration omitted full_overwrite")
	}
	var records []connectors.Record
	err = fullOverwrite.Source.ReadTransport(context.Background(), synctransport.SourceRequest{
		Connector: postgres,
		Runtime: connectors.RuntimeConfig{ProjectDir: root, Config: map[string]string{
			"mode": "fixture", "host": "fixture.internal", "database": "analytics", "username": "reader", "sslmode": "require",
		}},
		Stream: "public.users", CursorField: "updated_at", PrimaryKey: []string{"id"},
		Mode: synccontract.ModeFullOverwrite, BatchSize: 2,
		Resume: synccontract.ResumeExpectation{
			Source:           synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "fixture-credential", ObjectScope: "public.users"},
			SourceGeneration: synccontract.OpaqueToken("fixture-generation"),
		},
	}, func(page synctransport.SourcePage) error {
		records = append(records, page.Records...)
		return nil
	})
	if err != nil {
		t.Fatalf("production-composed PostgreSQL full_overwrite read = %v", err)
	}
	if got, want := len(records), 3; got != want {
		t.Fatalf("production-composed PostgreSQL full_overwrite records = %d, want %d", got, want)
	}
}

func TestOpenPostgresHistoryModeResolvesRegisteredExecutors(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	postgres, ok := a.registry.Get("postgres")
	if !ok {
		t.Fatal("PostgreSQL connector is not registered")
	}
	resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      postgres,
		Destination: postgres,
		Stream:      "snapshot",
		Mode:        synccontract.ModeIncrementalDedupeHistory,
	})
	if err != nil {
		t.Fatalf("PostgreSQL history preflight = %v", err)
	}
	if got, want := resolved.Source.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_polling_watermark"}); got != want {
		t.Fatalf("history source reference = %+v, want %+v", got, want)
	}
	if got, want := resolved.Destination.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_managed_target"}); got != want {
		t.Fatalf("history destination reference = %+v, want %+v", got, want)
	}
	if got, want := resolved.ApplyStrategy, (connectors.DestinationApplyStrategy{Mode: synccontract.ModeIncrementalDedupeHistory, Strategy: connectors.ApplyStrategyDedupeHistory, Action: "managed_incremental_dedupe_history"}); got != want {
		t.Fatalf("history apply strategy = %+v, want %+v", got, want)
	}
}

func TestOpenPostgresTransportDeclarationsAreExactModeIntersection(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	postgres, ok := a.registry.Get("postgres")
	if !ok {
		t.Fatal("PostgreSQL connector is not registered")
	}
	source, ok := connectors.SourceTransportDescriptorOf(postgres)
	if !ok {
		t.Fatal("PostgreSQL source transport is not declared")
	}
	destination, ok := connectors.DestinationTransportDescriptorOf(postgres)
	if !ok {
		t.Fatal("PostgreSQL destination transport is not declared")
	}
	want := []synccontract.Mode{
		synccontract.ModeFullOverwrite,
		synccontract.ModeFullAppend,
		synccontract.ModeIncrementalAppend,
		synccontract.ModeIncrementalUpsert,
		synccontract.ModeIncrementalDedupe,
		synccontract.ModeIncrementalDedupeHistory,
	}
	if fmt.Sprint(source.Modes) != fmt.Sprint(want) {
		t.Fatalf("PostgreSQL source modes = %v, want exact reachable intersection %v", source.Modes, want)
	}
	if fmt.Sprint(destination.Modes) != fmt.Sprint(want) {
		t.Fatalf("PostgreSQL destination modes = %v, want exact reachable intersection %v", destination.Modes, want)
	}
}

func assertGitHubTransportEligibleStreamsMatchDefinition(t *testing.T, github connectors.Connector) {
	t.Helper()
	definition, ok := connectors.DefinitionOf(github)
	if !ok {
		t.Fatal("GitHub connector has no definition")
	}
	descriptor, ok := connectors.SourceTransportDescriptorOf(github)
	if !ok {
		t.Fatal("GitHub connector has no source transport descriptor")
	}
	if len(descriptor.EligibleStreams) != len(definition.Streams) {
		t.Fatalf("GitHub eligible streams = %d, want all %d executable definition streams", len(descriptor.EligibleStreams), len(definition.Streams))
	}
	eligible := make(map[string]struct{}, len(descriptor.EligibleStreams))
	for _, stream := range descriptor.EligibleStreams {
		if stream == "*" {
			t.Fatal("GitHub transport eligibility must be a positive concrete allowlist, not a wildcard")
		}
		if _, duplicate := eligible[stream]; duplicate {
			t.Fatalf("GitHub eligible stream %q is duplicated", stream)
		}
		eligible[stream] = struct{}{}
	}
	for _, stream := range definition.Streams {
		if _, exists := eligible[stream.Name]; !exists {
			t.Errorf("GitHub executable stream %q is absent from transport eligibility", stream.Name)
		}
	}
}

func TestOpenComposedGitHubCommitsSourceEmitsEveryUnlimitedPageInBoundedBatches(t *testing.T) {
	providerRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		providerRequests++
		if request.URL.Path != "/repos/rails/rails/commits" {
			http.NotFound(w, request)
			return
		}
		page := request.URL.Query().Get("page")
		count := 100
		start := 0
		if page == "2" {
			count = 3
			start = 100
		}
		records := make([]map[string]any, 0, count)
		for index := range count {
			ordinal := start + index
			records = append(records, map[string]any{
				"sha": fmt.Sprintf("sha-%03d", ordinal),
				"commit": map[string]any{
					"message":   fmt.Sprintf("commit %d", ordinal),
					"author":    map[string]any{"name": "Ada", "email": "ada@example.test", "date": "2026-08-15T00:00:00Z"},
					"committer": map[string]any{"name": "Ada", "email": "ada@example.test", "date": "2026-08-15T00:00:00Z"},
				},
			})
		}
		if err := json.NewEncoder(w).Encode(records); err != nil {
			t.Errorf("encode provider page: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	github, ok := a.registry.Get("github")
	if !ok {
		t.Fatal("GitHub connector is not registered")
	}
	postgres, ok := a.registry.Get("postgres")
	if !ok {
		t.Fatal("PostgreSQL connector is not registered")
	}
	resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source: github, Destination: postgres, Stream: "commits", Mode: synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("production-composed commits preflight: %v", err)
	}
	runtime := connectors.RuntimeConfig{ProjectDir: root, Config: map[string]string{
		"base_url": server.URL, "owner": "rails", "repo": "rails", "public_access": "true", "max_pages": "unlimited",
	}}
	resume := synccontract.ResumeExpectation{
		Source:           synccontract.SourceIdentity{Engine: "github", AccountOrCluster: "public-test", ObjectScope: "commits"},
		SourceGeneration: synccontract.OpaqueToken("github-commits-generation"),
	}
	var pages []synctransport.SourcePage
	err = resolved.Source.ReadTransport(context.Background(), synctransport.SourceRequest{
		Connector: github, Runtime: runtime, Stream: "commits", Mode: synccontract.ModeFullAppend,
		BatchSize: 25, Resume: resume,
	}, func(page synctransport.SourcePage) error {
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("production-composed commits ReadTransport: %v", err)
	}
	if got, want := providerRequests, 2; got != want {
		t.Fatalf("provider requests = %d, want %d unlimited pages", got, want)
	}
	if got, want := len(pages), 5; got != want {
		t.Fatalf("transport pages = %d, want %d bounded batches", got, want)
	}
	total := 0
	for index, page := range pages {
		if len(page.Records) == 0 || len(page.Records) > 25 {
			t.Fatalf("transport page %d records = %d, want 1..25", index, len(page.Records))
		}
		if page.CandidateCheckpoint.Mechanism != "declarative_stream_engine_read" {
			t.Fatalf("transport page %d mechanism = %q", index, page.CandidateCheckpoint.Mechanism)
		}
		if index > 0 && bytes.Compare(pages[index-1].CandidateCheckpoint.Position.Primary, page.CandidateCheckpoint.Position.Primary) >= 0 {
			t.Fatalf("transport page %d primary position did not advance monotonically", index)
		}
		total += len(page.Records)
	}
	if got, want := total, 103; got != want {
		t.Fatalf("emitted records = %d, want %d", got, want)
	}

	interrupted := errors.New("simulated process death before durable acknowledgement")
	providerRequests = 0
	var attempted synccontract.CheckpointEnvelope
	err = resolved.Source.ReadTransport(context.Background(), synctransport.SourceRequest{
		Connector: github, Runtime: runtime, Stream: "commits", Mode: synccontract.ModeFullAppend,
		BatchSize: 25, Resume: resume,
	}, func(page synctransport.SourcePage) error {
		attempted = page.CandidateCheckpoint.Clone()
		return interrupted
	})
	if !errors.Is(err, interrupted) || providerRequests != 1 {
		t.Fatalf("interrupted commits read = (%v, requests=%d), want one attempted provider page and no acknowledgement", err, providerRequests)
	}
	providerRequests = 0
	var replayed synccontract.CheckpointEnvelope
	err = resolved.Source.ReadTransport(context.Background(), synctransport.SourceRequest{
		Connector: github, Runtime: runtime, Stream: "commits", Mode: synccontract.ModeFullAppend,
		BatchSize: 25, Resume: resume,
	}, func(page synctransport.SourcePage) error {
		replayed = page.CandidateCheckpoint.Clone()
		return interrupted
	})
	if !errors.Is(err, interrupted) || !checkpointPositionEqual(attempted.Position, replayed.Position) {
		t.Fatalf("unacknowledged commits replay = (%v, %+v), want the same candidate position %+v", err, replayed.Position, attempted.Position)
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement("postgres", attempted.ObservedAt.Add(1))
	if err != nil {
		t.Fatal(err)
	}
	var committed synccontract.CheckpointEnvelope
	if err := synccontract.CommitAfterDownstreamAcknowledgement(attempted, acknowledgement, func(checkpoint synccontract.CheckpointEnvelope) error {
		committed = checkpoint
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	providerRequests = 0
	resumedRecords := 0
	err = resolved.Source.ReadTransport(context.Background(), synctransport.SourceRequest{
		Connector: github, Runtime: runtime, Stream: "commits", Mode: synccontract.ModeFullAppend,
		BatchSize: 25, Resume: resume, Checkpoint: &committed,
	}, func(page synctransport.SourcePage) error {
		resumedRecords += len(page.Records)
		return nil
	})
	if err != nil || providerRequests != 2 || resumedRecords != 78 {
		t.Fatalf("acknowledged commits resume = (err=%v requests=%d records=%d), want two-page traversal and 78 records after the durable batch", err, providerRequests, resumedRecords)
	}
}

func TestOpenComposedGitHubCommitsHonorsTransportMaxPages(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		maxPages     string
		wantRequests int
		wantRecords  int
		wantBudget   bool
	}{
		{name: "omitted defaults to one page", wantRequests: 1, wantRecords: 100, wantBudget: true},
		{name: "positive cap", maxPages: "2", wantRequests: 2, wantRecords: 200, wantBudget: true},
		{name: "zero is unlimited", maxPages: "0", wantRequests: 3, wantRecords: 201},
		{name: "all is unlimited", maxPages: "all", wantRequests: 3, wantRecords: 201},
		{name: "unlimited is unlimited", maxPages: "unlimited", wantRequests: 3, wantRecords: 201},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			providerRequests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				providerRequests++
				page, _ := strconv.Atoi(request.URL.Query().Get("page"))
				if page == 0 {
					page = 1
				}
				count := 100
				if page == 3 {
					count = 1
				}
				records := make([]map[string]any, 0, count)
				for index := range count {
					records = append(records, map[string]any{
						"sha": fmt.Sprintf("sha-%d-%d", page, index),
						"commit": map[string]any{
							"message":   "bounded page",
							"author":    map[string]any{"date": "2026-08-15T00:00:00Z"},
							"committer": map[string]any{"date": "2026-08-15T00:00:00Z"},
						},
					})
				}
				if err := json.NewEncoder(w).Encode(records); err != nil {
					t.Errorf("encode provider page: %v", err)
				}
			}))
			defer server.Close()

			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatal(err)
			}
			a, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			github, _ := a.registry.Get("github")
			postgres, _ := a.registry.Get("postgres")
			resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
				Source: github, Destination: postgres, Stream: "commits", Mode: synccontract.ModeFullAppend,
			})
			if err != nil {
				t.Fatal(err)
			}
			config := map[string]string{"base_url": server.URL, "owner": "rails", "repo": "rails", "public_access": "true"}
			if testCase.maxPages != "" {
				config[declarativeTransportMaxPagesConfig] = testCase.maxPages
			}
			resume := synccontract.ResumeExpectation{
				Source:           synccontract.SourceIdentity{Engine: "github", AccountOrCluster: "pagination-test", ObjectScope: "commits"},
				SourceGeneration: synccontract.OpaqueToken("pagination-generation"),
			}
			records := 0
			err = resolved.Source.ReadTransport(context.Background(), synctransport.SourceRequest{
				Connector: github, Runtime: connectors.RuntimeConfig{ProjectDir: root, Config: config}, Stream: "commits",
				Mode: synccontract.ModeFullAppend, BatchSize: 1000, Resume: resume,
			}, func(page synctransport.SourcePage) error {
				records += len(page.Records)
				return nil
			})
			var budgetStop *synctransport.SourceBudgetStoppedError
			if testCase.wantBudget {
				if !errors.As(err, &budgetStop) || budgetStop == nil || budgetStop.Continuation.Kind == "" || len(budgetStop.Continuation.Token) == 0 {
					t.Fatalf("max_pages=%q read error = %T %v, want typed bounded-source continuation", testCase.maxPages, err, err)
				}
			} else if err != nil {
				t.Fatalf("max_pages=%q read error = %v, want exhausted scan", testCase.maxPages, err)
			}
			if providerRequests != testCase.wantRequests || records != testCase.wantRecords {
				t.Fatalf("max_pages=%q read = (requests=%d records=%d), want requests=%d records=%d", testCase.maxPages, providerRequests, records, testCase.wantRequests, testCase.wantRecords)
			}
		})
	}
}

// TestDeclarativeTransport_PageBudgetIsNotEOF keeps the source-side half of
// PF-CF-B27 production-shaped: an engine-declared three-page collection must
// distinguish an exhausted scan from a bounded prefix, and the opaque
// continuation must resume the next provider page without exposing a caller
// cursor or replaying an acknowledged batch.
func TestDeclarativeTransport_PageBudgetIsNotEOF(t *testing.T) {
	type outcomeSource interface {
		ReadTransportWithOutcome(context.Context, synctransport.SourceRequest, func(synctransport.SourcePage) error) (synctransport.SourceReadOutcome, error)
	}

	requestedPages := make([]int, 0, 8)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		page, _ := strconv.Atoi(request.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}
		requestedPages = append(requestedPages, page)
		count := 100
		if page == 3 {
			count = 1
		}
		records := make([]map[string]any, 0, count)
		for index := range count {
			records = append(records, map[string]any{
				"sha": fmt.Sprintf("sha-%d-%d", page, index),
				"commit": map[string]any{
					"message":   "budget outcome",
					"author":    map[string]any{"date": "2026-08-15T00:00:00Z"},
					"committer": map[string]any{"date": "2026-08-15T00:00:00Z"},
				},
			})
		}
		if err := json.NewEncoder(w).Encode(records); err != nil {
			t.Errorf("encode provider page: %v", err)
		}
	}))
	defer server.Close()

	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	github, _ := a.registry.Get("github")
	postgres, _ := a.registry.Get("postgres")
	resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source: github, Destination: postgres, Stream: "commits", Mode: synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatal(err)
	}
	source, ok := resolved.Source.(outcomeSource)
	if !ok {
		t.Fatalf("declarative source %T does not expose a typed read outcome", resolved.Source)
	}
	resume := synccontract.ResumeExpectation{
		Source:           synccontract.SourceIdentity{Engine: "github", AccountOrCluster: "budget-outcome", ObjectScope: "commits"},
		SourceGeneration: synccontract.OpaqueToken("budget-outcome-generation"),
	}
	baseConfig := map[string]string{"base_url": server.URL, "owner": "rails", "repo": "rails", "public_access": "true"}
	cappedConfig := func() map[string]string {
		config := cloneStringMap(baseConfig)
		config[declarativeTransportMaxPagesConfig] = "1"
		return config
	}

	for _, testCase := range []struct {
		name          string
		maxPages      string
		wantRequests  int
		wantRecords   int
		wantExhausted bool
	}{
		{name: "omitted default is an incomplete one-page prefix", wantRequests: 1, wantRecords: 100},
		{name: "explicit one-page prefix", maxPages: "1", wantRequests: 1, wantRecords: 100},
		{name: "explicit two-page prefix", maxPages: "2", wantRequests: 2, wantRecords: 200},
		{name: "unlimited scan is exhausted", maxPages: "unlimited", wantRequests: 3, wantRecords: 201, wantExhausted: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			requestedPages = requestedPages[:0]
			config := cloneStringMap(baseConfig)
			if testCase.maxPages != "" {
				config[declarativeTransportMaxPagesConfig] = testCase.maxPages
			}
			records := 0
			outcome, readErr := source.ReadTransportWithOutcome(context.Background(), synctransport.SourceRequest{
				Connector: github, Runtime: connectors.RuntimeConfig{ProjectDir: root, Config: config}, Stream: "commits",
				Mode: synccontract.ModeFullAppend, BatchSize: issueCollectionTransportMaxRecords, Resume: resume,
			}, func(page synctransport.SourcePage) error {
				records += len(page.Records)
				return nil
			})
			if readErr != nil {
				t.Fatalf("ReadTransportWithOutcome() error = %v", readErr)
			}
			if got := len(requestedPages); got != testCase.wantRequests {
				t.Fatalf("provider requests = %d, want %d", got, testCase.wantRequests)
			}
			if records != testCase.wantRecords {
				t.Fatalf("emitted records = %d, want %d", records, testCase.wantRecords)
			}
			if outcome.Exhausted != testCase.wantExhausted {
				t.Fatalf("outcome exhausted = %t, want %t", outcome.Exhausted, testCase.wantExhausted)
			}
			if outcome.Exhausted && outcome.Continuation != nil {
				t.Fatalf("exhausted outcome exposed a continuation: %#v", outcome.Continuation)
			}
			if !outcome.Exhausted && outcome.Continuation == nil {
				t.Fatal("budget-stopped outcome omitted its opaque continuation")
			}
		})
	}

	requestedPages = requestedPages[:0]
	var firstPage synctransport.SourcePage
	firstOutcome, err := source.ReadTransportWithOutcome(context.Background(), synctransport.SourceRequest{
		Connector: github,
		Runtime:   connectors.RuntimeConfig{ProjectDir: root, Config: cappedConfig()},
		Stream:    "commits", Mode: synccontract.ModeIncrementalDedupe, BatchSize: issueCollectionTransportMaxRecords, Resume: resume,
	}, func(page synctransport.SourcePage) error {
		firstPage = page
		return nil
	})
	if err != nil || firstOutcome.Exhausted || firstOutcome.Continuation == nil || len(firstPage.Records) != 100 {
		t.Fatalf("first incremental prefix = (outcome=%#v err=%v records=%d), want one bounded resumable page", firstOutcome, err, len(firstPage.Records))
	}
	ack, err := synccontract.NewDurableDownstreamAcknowledgement("postgres", firstPage.CandidateCheckpoint.ObservedAt.Add(time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	var checkpoint synccontract.CheckpointEnvelope
	if err := synccontract.CommitAfterDownstreamAcknowledgement(firstPage.CandidateCheckpoint, ack, func(committed synccontract.CheckpointEnvelope) error {
		checkpoint = committed
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	checkpoint.Continuation = firstOutcome.Continuation.Clone()

	requestedPages = requestedPages[:0]
	resumed := make([]string, 0, 100)
	_, err = source.ReadTransportWithOutcome(context.Background(), synctransport.SourceRequest{
		Connector: github,
		Runtime:   connectors.RuntimeConfig{ProjectDir: root, Config: cappedConfig()},
		Stream:    "commits", Mode: synccontract.ModeIncrementalDedupe, BatchSize: issueCollectionTransportMaxRecords,
		Resume: resume, Checkpoint: &checkpoint,
	}, func(page synctransport.SourcePage) error {
		for _, record := range page.Records {
			resumed = append(resumed, record["sha"].(string))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("resumed incremental read error = %v", err)
	}
	if got, want := len(resumed), 100; got != want {
		t.Fatalf("resumed records = %d, want page N+1 exactly once (%d)", got, want)
	}
	for _, sha := range resumed {
		if !strings.HasPrefix(sha, "sha-2-") {
			t.Fatalf("resumed records include an acknowledged-prefix value %q", sha)
		}
	}
	if got, want := requestedPages, []int{1, 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("resumed provider traversal = %v, want engine-owned replay through %v", got, want)
	}

	// The continuation is an engine-owned, definition-bound capability. A
	// corrupted persisted token must be rejected before it can select a
	// provider page, rather than falling back to page one or treating the
	// capped prefix as an exhausted scan.
	corrupt := checkpoint.Clone()
	corrupt.Continuation.Token = append(synccontract.OpaqueToken(nil), corrupt.Continuation.Token...)
	corrupt.Continuation.Token[0] ^= 0x01
	requestedPages = requestedPages[:0]
	_, err = source.ReadTransportWithOutcome(context.Background(), synctransport.SourceRequest{
		Connector: github,
		Runtime:   connectors.RuntimeConfig{ProjectDir: root, Config: cappedConfig()},
		Stream:    "commits", Mode: synccontract.ModeIncrementalDedupe, BatchSize: issueCollectionTransportMaxRecords,
		Resume: resume, Checkpoint: &corrupt,
	}, func(synctransport.SourcePage) error {
		t.Fatal("corrupt continuation reached the source emitter")
		return nil
	})
	if err == nil {
		t.Fatal("corrupt continuation was accepted")
	}
	if got := len(requestedPages); got != 0 {
		t.Fatalf("corrupt continuation made %d provider requests, want rejection before I/O", got)
	}
}

func TestOpenComposedGitHubCommitsTimesOutOneProviderPageWithoutCancellingTheRunContext(t *testing.T) {
	providerRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		providerRequests++
		<-request.Context().Done()
	}))
	defer server.Close()

	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	github, _ := a.registry.Get("github")
	postgres, _ := a.registry.Get("postgres")
	resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source: github, Destination: postgres, Stream: "commits", Mode: synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatal(err)
	}
	resume := synccontract.ResumeExpectation{
		Source:           synccontract.SourceIdentity{Engine: "github", AccountOrCluster: "deadline-test", ObjectScope: "commits"},
		SourceGeneration: synccontract.OpaqueToken("deadline-generation"),
	}
	runCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	started := time.Now()
	err = resolved.Source.ReadTransport(runCtx, synctransport.SourceRequest{
		Connector: github, Runtime: connectors.RuntimeConfig{ProjectDir: root, Config: map[string]string{
			"base_url": server.URL, "owner": "rails", "repo": "rails", "public_access": "true", "max_pages": "1",
		}},
		Stream: "commits", Mode: synccontract.ModeFullAppend, BatchSize: 100,
		Resume: resume, UnitDeadline: 20 * time.Millisecond,
	}, func(synctransport.SourcePage) error {
		t.Fatal("slow provider page reached the source emitter")
		return nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ReadTransport() error = %T %v, want the one-page deadline", err, err)
	}
	if elapsed := time.Since(started); elapsed >= 300*time.Millisecond {
		t.Fatalf("ReadTransport() elapsed = %s, want a page deadline rather than the one-second run context", elapsed)
	}
	if providerRequests != 1 {
		t.Fatalf("provider requests = %d, want one timed-out fetch", providerRequests)
	}
	if runCtx.Err() != nil {
		t.Fatalf("run context = %v, want the timed-out page to leave the run context usable for checkpoint resume", runCtx.Err())
	}
}

func TestLocalWarehouseDestinationExecutorWritesAndReadBacksConnectionOwnedParquet(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	warehouseConnector, ok := a.registry.Get("warehouse")
	if !ok {
		t.Fatal("local warehouse connector is not registered")
	}
	conn := Connection{
		ID:          "connection_transport_warehouse",
		Name:        "transport-warehouse",
		Source:      EndpointConfig{Connector: "postgres"},
		Destination: EndpointConfig{Connector: "warehouse"},
		Streams: map[string]StreamConfig{
			"snapshot": {DestinationTable: "snapshot_rows"},
		},
	}
	a.state.Connections = append(a.state.Connections, conn)
	executor, err := newLocalWarehouseDestinationExecutor(a, warehouseConnector)
	if err != nil {
		t.Fatal(err)
	}
	runtime := connectors.RuntimeConfig{ProjectDir: root, Config: map[string]string{"path": t.TempDir()}}
	strategy, err := localWarehouseApplyStrategy(synccontract.ModeFullAppend)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := executor.PlanDestination(context.Background(), synctransport.DestinationPlanRequest{
		Connector:     warehouseConnector,
		Runtime:       runtime,
		Stream:        "snapshot",
		Mode:          synccontract.ModeFullAppend,
		ApplyStrategy: strategy,
	})
	if err != nil {
		t.Fatalf("PlanDestination() error = %v", err)
	}
	receipt := synctransport.WarehouseReceipt{
		ID:               "stage_transport_warehouse",
		Owner:            conn.ID,
		Generation:       1,
		Stream:           "snapshot",
		Mode:             synccontract.ModeFullAppend,
		CheckpointSHA256: "checkpoint",
		TombstonesSHA256: "tombstones",
		ManifestSHA256:   "manifest",
		ContentSHA256:    "content",
		ParquetSHA256:    "parquet",
		Records:          1,
	}
	workset := synctransport.WarehouseWorkset{ID: receipt.ID, Records: []connectors.Record{{"id": "row-1", "name": "Ada"}}}
	ack, err := executor.ApplyDestination(context.Background(), synctransport.DestinationApplyRequest{
		ConnectionID: conn.ID,
		Plan:         plan,
		Receipt:      receipt,
		Workset:      workset,
		Runtime:      runtime,
	})
	if err != nil {
		t.Fatalf("ApplyDestination() error = %v", err)
	}
	location, err := a.warehouseLocation(runtime.Config["path"], conn)
	if err != nil {
		t.Fatal(err)
	}
	tablePath, err := location.TablePath("snapshot_rows")
	if err != nil {
		t.Fatal(err)
	}
	var rows []warehouse.Row
	if err := warehouse.ReadTable(context.Background(), tablePath, func(row warehouse.Row) error {
		rows = append(rows, row)
		return nil
	}); err != nil {
		t.Fatalf("read connection-owned table: %v", err)
	}
	if len(rows) != 1 || rows[0]["id"] != "row-1" {
		t.Fatalf("connection-owned table rows = %#v, want the reopened workset row", rows)
	}
	if err := executor.ReadBackDestination(context.Background(), synctransport.DestinationReadBackRequest{
		Plan:            plan,
		Workset:         workset,
		Acknowledgement: ack,
		Runtime:         runtime,
	}); err != nil {
		t.Fatalf("ReadBackDestination() error = %v", err)
	}
	if err := warehouse.WriteTable(context.Background(), tablePath, []warehouse.Row{{"id": "changed"}}); err != nil {
		t.Fatal(err)
	}
	if err := executor.ReadBackDestination(context.Background(), synctransport.DestinationReadBackRequest{
		Plan:            plan,
		Workset:         workset,
		Acknowledgement: ack,
		Runtime:         runtime,
	}); err == nil {
		t.Fatal("ReadBackDestination() accepted a table changed after acknowledgement")
	}
}

func TestLocalWarehouseDestinationExecutorRefusesChangeCapture(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	warehouseConnector, ok := a.registry.Get("warehouse")
	if !ok {
		t.Fatal("local warehouse connector is not registered")
	}
	descriptor, ok := connectors.DestinationTransportDescriptorOf(warehouseConnector)
	if !ok {
		t.Fatal("local warehouse transport declaration is absent")
	}
	if _, err := descriptor.ApplyStrategyFor(synccontract.ModeChangeCapture); err == nil {
		t.Fatal("local warehouse transport declared change_capture as a destination mode")
	}
	if _, err := localWarehouseApplyStrategy(synccontract.ModeChangeCapture); err == nil {
		t.Fatal("local warehouse destination executor exposed a change_capture strategy")
	}
}
