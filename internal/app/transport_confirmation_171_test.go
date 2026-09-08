package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestDefinitionOwnedTransportConfirmation171(t *testing.T) {
	for _, executor := range []string{"declarative_typed_destination", "declarative_single_attempt_destination"} {
		t.Run(executor, func(t *testing.T) {
			ctx := t.Context()
			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatal(err)
			}
			a, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			var reads, writes atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/widgets":
					reads.Add(1)
					_, _ = w.Write([]byte(`{"data":[{"id":"widget-1","value":"approved"}]}`))
				case r.Method == http.MethodPost && r.URL.Path == "/widgets/target":
					writes.Add(1)
					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write([]byte(`{"receipt_id":"widget-1"}`))
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			t.Cleanup(server.Close)
			const name = "confirmation-fixture"
			files := declarativeTypedDestinationBundleFS(name, "source", "destination")
			if executor == "declarative_single_attempt_destination" {
				var doc map[string]any
				if err := json.Unmarshal(files[name+"/sync_transport.json"].Data, &doc); err != nil {
					t.Fatal(err)
				}
				destination := doc["destination_transport"].(map[string]any)
				destination["executor"].(map[string]any)["id"] = executor
				destination["delivery"].(map[string]any)["idempotency"] = "single_attempt"
				delete(destination["apply_strategies"].([]any)[0].(map[string]any), "read_back")
				data, err := json.Marshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				files[name+"/sync_transport.json"] = &fstest.MapFile{Data: data}
			}
			bundle, err := engine.Load(files, name)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.registry.Register(engine.New(bundle, nil)); err != nil {
				t.Fatal(err)
			}
			if err := a.composeTransportRegistry(); err != nil {
				t.Fatal(err)
			}
			if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "local", Connector: name, Config: map[string]string{"base_url": server.URL}}); err != nil {
				t.Fatal(err)
			}
			connection, err := a.CreateConnection(ctx, CreateConnectionRequest{
				Name: "confirmation_flow", Source: EndpointConfig{Connector: name, Credential: "local"},
				Destination: EndpointConfig{Connector: name, Credential: "local"},
				Streams:     map[string]StreamConfig{"widgets": {SyncMode: string(synccontract.ModeFullAppend), PrimaryKey: []string{"id"}, DestinationAction: "apply_widget"}},
			})
			if err != nil {
				t.Fatal(err)
			}
			// The transport contract must not change the same standalone action.
			ordinary := ReversePlan{DestinationConnector: name, Action: "apply_widget"}
			if got := a.confirmationPolicyForPlan(ordinary); got.Kind != "" {
				t.Fatalf("ordinary action unexpectedly requires confirmation: %v", got.Kind)
			}
			beforeReads, beforeWrites := reads.Load(), writes.Load()
			plan, err := a.PlanDeclarativeTypedDestinationTransport(ctx, connection.Name, "widgets")
			if err != nil {
				t.Fatal(err)
			}
			if plan.PlanSeal == nil || plan.ConfirmationPolicy.Kind != connectors.ConfirmationKindDestructive || plan.ApprovalToken != "" {
				t.Fatal("creation did not seal destructive policy without granting execution")
			}
			// These are persisted-plan mutations; public preview must refuse each.
			for _, tc := range []struct {
				name string
				edit func(*ReversePlan)
			}{
				{"missing_seal", func(p *ReversePlan) { p.PlanSeal = nil }},
				{"removed_confirmation", func(p *ReversePlan) { p.ConfirmationPolicy = connectors.WriteConfirmation{} }},
				{"different_mode", func(p *ReversePlan) { p.Mode = "append" }},
				{"different_action", func(p *ReversePlan) { p.Action = "other" }},
				{"different_binding", func(p *ReversePlan) { p.TransportBindingSHA256 = "tampered" }},
			} {
				t.Run(tc.name, func(t *testing.T) {
					changed := plan
					tc.edit(&changed)
					a.state.ReversePlans[len(a.state.ReversePlans)-1] = changed
					if err := a.save(); err != nil {
						t.Fatal(err)
					}
					if _, _, err := a.PreviewDeclarativeTypedDestinationTransport(ctx, plan.ID); err == nil {
						t.Fatal("tampered persisted plan reached preview")
					}
				})
			}
			a.state.ReversePlans[len(a.state.ReversePlans)-1] = plan
			if err := a.save(); err != nil {
				t.Fatal(err)
			}
			previewed, preview, err := a.PreviewDeclarativeTypedDestinationTransport(ctx, plan.ID)
			if err != nil {
				t.Fatalf("valid definition-owned transport preview: %v", err)
			}
			if previewed.ApprovalToken == "" || preview.ApprovalTarget.Confirmation.Kind != connectors.ConfirmationKindDestructive {
				t.Fatal("preview did not preserve destructive approval target")
			}
			if _, err := a.AuthorizationScopeForReversePlan(ctx, plan.ID); err != nil {
				t.Fatalf("current transport authorization scope: %v", err)
			}
			if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection.Name, Stream: "widgets", BatchSize: 1, DestinationApproval: synctransport.DestinationApproval{PlanID: plan.ID, ApprovalToken: previewed.ApprovalToken}}); err == nil {
				t.Fatal("transport ran without destructive confirmation")
			}
			if reads.Load() != beforeReads || writes.Load() != beforeWrites {
				t.Fatal("planning, preview, or refused execution performed provider I/O")
			}
			run, err := a.RunETL(ctx, RunETLRequest{Connection: connection.Name, Stream: "widgets", BatchSize: 1, DestinationApproval: synctransport.DestinationApproval{PlanID: plan.ID, ApprovalToken: previewed.ApprovalToken, Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}}})
			if err != nil {
				t.Fatal(err)
			}
			if run.Status != "completed" || run.RecordsRead != 1 || run.RecordsLoaded != 1 || writes.Load()-beforeWrites != 1 || reads.Load() <= beforeReads || len(run.DestinationResults) != 1 {
				t.Fatal("approved transport did not return one acknowledged record and persisted result")
			}
		})
	}
}
