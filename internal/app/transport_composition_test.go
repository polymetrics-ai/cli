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
	if sourceFactory == nil {
		t.Fatal("declarative source factory was not registered")
	}
	sourceEvidenceAccepted := sourceFactory.SourceEvidence == source.Conformance
	for _, accepted := range sourceFactory.AcceptedSourceEvidences {
		if accepted == source.Conformance {
			sourceEvidenceAccepted = true
			break
		}
	}
	if !sourceEvidenceAccepted {
		t.Fatalf("source factory evidence = %#v, want declaration %#v to be accepted", sourceFactory, source.Conformance)
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
func TestDefinitionTransportFactoriesRunTypedDestinationFromDefinition(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	var reads, writes int
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/widgets":
			reads++
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
			_ = json.NewEncoder(writer).Encode(map[string]any{"ok": true})
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	bundleFS := syntheticTransportBundleFS()
	bundleFS["synthetic/metadata.json"] = &fstest.MapFile{Data: []byte(`{"name":"synthetic","display_name":"Synthetic typed destination","description":"test-only typed destination","integration_type":"api","release_stage":"ga","capabilities":{"check":true,"read":true,"write":true,"query":false,"cdc":false,"dynamic_schema":false}}`)}
	bundleFS["synthetic/streams.json"] = &fstest.MapFile{Data: []byte(`{"base":{"url":"{{ config.base_url }}","user_agent":"synthetic","headers":{},"auth":[],"pagination":{"type":"none"},"check":{"method":"GET","path":"/widgets"},"error_map":[]},"streams":[{"name":"widgets","path":"/widgets","records":{"path":"data"},"schema":"schemas/widgets.json"}]}`)}
	bundleFS["synthetic/api_surface.json"] = &fstest.MapFile{Data: []byte(`{"api":"synthetic","endpoints":[{"method":"GET","path":"/widgets","covered_by":{"stream":"widgets"}},{"method":"POST","path":"/widgets/target","covered_by":{"write":"apply_widget"}}]}`)}
	bundleFS["synthetic/schemas/widgets.json"] = &fstest.MapFile{Data: []byte(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","x-primary-key":["id"],"properties":{"id":{"type":"string"},"value":{"type":"string"}}}`)}
	bundleFS["synthetic/writes.json"] = &fstest.MapFile{Data: []byte(`{"actions":[{"name":"apply_widget","kind":"create","method":"POST","path":"/widgets/target","body_type":"json","body_fields":["target_id","value"],"risk":"creates a synthetic widget target","record_schema":{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","required":["target_id","value"],"additionalProperties":false,"properties":{"target_id":{"type":"string"},"value":{"type":"string"}}}}]}`)}
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
	if reads != 1 || writes != 1 || commits != 1 {
		t.Fatalf("synthetic typed destination effects read/write/commit = %d/%d/%d, want 1 each", reads, writes, commits)
	}
	if result.RecordsRead != 1 || result.RecordsApplied != 1 || result.CommittedCheckpoint == nil {
		t.Fatalf("synthetic typed destination result = %#v, want one acknowledged action", result)
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

func TestDeclarativeTypedDestinationRefusesInvalidWorksetsBeforeProviderWrite(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(fstest.MapFS, string)
		record connectors.Record
		tombs  []synccontract.Tombstone
	}{
		{
			name: "mapped record misses required action field",
			mutate: func(files fstest.MapFS, name string) {
				path := name + "/sync_transport.json"
				files[path] = &fstest.MapFile{Data: []byte(strings.Replace(string(files[path].Data), `"inputs": [{"input": "target_id", "field": "id"}, {"input": "value", "field": "value"}]`, `"inputs": [{"input": "target_id", "field": "id"}]`, 1))}
			},
			record: connectors.Record{"id": "widget-1", "value": "not-mapped"},
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
				t.Fatalf("plan typed destination: %v", err)
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
	files[name+"/streams.json"] = &fstest.MapFile{Data: []byte(`{"base":{"url":"{{ config.base_url }}","user_agent":"synthetic","headers":{},"auth":[],"pagination":{"type":"none"},"check":{"method":"GET","path":"/widgets"},"error_map":[]},"streams":[{"name":"widgets","path":"/widgets","records":{"path":"data"},"schema":"schemas/widgets.json"}]}`)}
	files[name+"/api_surface.json"] = &fstest.MapFile{Data: []byte(`{"api":"synthetic","endpoints":[{"method":"GET","path":"/widgets","covered_by":{"stream":"widgets"}},{"method":"POST","path":"/widgets/target","covered_by":{"write":"apply_widget"}}]}`)}
	files[name+"/schemas/widgets.json"] = &fstest.MapFile{Data: []byte(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","x-primary-key":["id"],"properties":{"id":{"type":"string"},"value":{"type":"string"}}}`)}
	files[name+"/writes.json"] = &fstest.MapFile{Data: []byte(`{"actions":[{"name":"apply_widget","kind":"create","method":"POST","path":"/widgets/target","body_type":"json","body_fields":["target_id","value"],"risk":"creates a synthetic widget target","record_schema":{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","required":["target_id","value"],"additionalProperties":false,"properties":{"target_id":{"type":"string"},"value":{"type":"string"}}}}]}`)}
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
    "source_bindings": [{
      "executor": {"family": "declarative_api", "id": "declarative_stream_source"},
      "eligible_streams": ["widgets"],
      "record_mapping": {"kind": "input_fields", "inputs": [{"input": "target_id", "field": "id"}, {"input": "value", "field": "value"}]}
    }]
  }
}`, sourceRunID, destinationRunID))}
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
	}{
		{name: "omitted defaults to one page", wantRequests: 1, wantRecords: 100},
		{name: "positive cap", maxPages: "2", wantRequests: 2, wantRecords: 200},
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
			if err != nil || providerRequests != testCase.wantRequests || records != testCase.wantRecords {
				t.Fatalf("max_pages=%q read = (err=%v requests=%d records=%d), want requests=%d records=%d", testCase.maxPages, err, providerRequests, records, testCase.wantRequests, testCase.wantRecords)
			}
		})
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
