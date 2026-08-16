package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
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

func TestOpenRefusesPostgresUnpairedHistoryModeBeforeExecutorIO(t *testing.T) {
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
	_, err = a.transports.Preflight(synctransport.PreflightRequest{
		Source:      postgres,
		Destination: postgres,
		Stream:      "snapshot",
		Mode:        synccontract.ModeIncrementalDedupeHistory,
	})
	if got, want := fmt.Sprint(err), `source transport does not support sync mode "incremental_dedupe_history"`; got != want {
		t.Fatalf("history preflight refusal = %q, want %q before executor I/O", got, want)
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

func TestLocalWarehouseDestinationExecutorAppliesChangeCaptureTombstones(t *testing.T) {
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
		ID:          "connection_transport_change_capture",
		Name:        "transport-change-capture",
		Source:      EndpointConfig{Connector: "postgres"},
		Destination: EndpointConfig{Connector: "warehouse"},
		Streams: map[string]StreamConfig{
			"changes": {DestinationTable: "change_rows", PrimaryKey: []string{"id"}},
		},
	}
	a.state.Connections = append(a.state.Connections, conn)
	executor, err := newLocalWarehouseDestinationExecutor(a, warehouseConnector)
	if err != nil {
		t.Fatal(err)
	}
	runtime := connectors.RuntimeConfig{ProjectDir: root, Config: map[string]string{"path": t.TempDir()}}
	strategy, err := localWarehouseApplyStrategy(synccontract.ModeChangeCapture)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := executor.PlanDestination(context.Background(), synctransport.DestinationPlanRequest{
		Connector:     warehouseConnector,
		Runtime:       runtime,
		Stream:        "changes",
		Mode:          synccontract.ModeChangeCapture,
		ApplyStrategy: strategy,
	})
	if err != nil {
		t.Fatalf("PlanDestination() error = %v", err)
	}
	receipt := synctransport.WarehouseReceipt{
		ID:               "stage_transport_change_capture",
		Owner:            conn.ID,
		Generation:       1,
		Stream:           "changes",
		Mode:             synccontract.ModeChangeCapture,
		CheckpointSHA256: "checkpoint",
		TombstonesSHA256: "tombstones",
		ManifestSHA256:   "manifest",
		ContentSHA256:    "content",
		ParquetSHA256:    "parquet",
		Records:          2,
		Tombstones:       1,
	}
	workset := synctransport.WarehouseWorkset{
		ID:      receipt.ID,
		Records: []connectors.Record{{"id": "kept", "name": "Ada"}, {"id": "removed", "name": "Alan"}},
		Tombstones: []synccontract.Tombstone{{
			Operation:   synccontract.OperationDelete,
			EventID:     synccontract.OpaqueToken("event-1"),
			Key:         json.RawMessage(`{"id":"removed"}`),
			DeleteImage: synccontract.DeleteImageKeyOnly,
			Position: synccontract.CheckpointPosition{
				Primary:    synccontract.OpaqueToken("2"),
				TieBreaker: synccontract.OpaqueToken("2"),
			},
		}},
	}
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
	tablePath, err := location.TablePath("change_rows")
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	if err := warehouse.ReadTable(context.Background(), tablePath, func(row warehouse.Row) error {
		id, _ := row["id"].(string)
		ids = append(ids, id)
		return nil
	}); err != nil {
		t.Fatalf("read change-capture table: %v", err)
	}
	if len(ids) != 1 || ids[0] != "kept" {
		t.Fatalf("change-capture table ids = %#v, want only the non-tombstoned record", ids)
	}
	if err := executor.ReadBackDestination(context.Background(), synctransport.DestinationReadBackRequest{
		Plan:            plan,
		Workset:         workset,
		Acknowledgement: ack,
		Runtime:         runtime,
	}); err != nil {
		t.Fatalf("ReadBackDestination() error = %v", err)
	}
}
