package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	stripeTransportDefsRoot = "../connectors/defs"
	stripeTransportToken    = "fixture-stripe-token"
)

// TestStripeDeclaredETLStreamsMaterializeDuckDB exercises every current Stripe
// stream declaration through the existing App.RunETL path and a temporary
// DuckDB project. It proves local legacy app behavior only; source-lane matrix
// states remain mapping-only until a source descriptor contract is available.
func TestStripeDeclaredETLStreamsMaterializeDuckDB(t *testing.T) {
	fixture := newStripeTransportFixture(t)
	application := fixture.openApplication(t)
	ctx := context.Background()

	for _, binding := range fixture.etlBindings {
		binding := binding
		t.Run(binding.Record, func(t *testing.T) {
			connection, err := application.CreateConnection(ctx, app.CreateConnectionRequest{
				Name:        "stripe_" + binding.Record + "_to_warehouse",
				Source:      app.EndpointConfig{Connector: "stripe", Credential: "stripe-local"},
				Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
				Streams: map[string]app.StreamConfig{binding.Record: {
					SyncMode:         "full_refresh_append",
					DestinationTable: "stripe_" + binding.Record + "_witness",
				}},
			})
			if err != nil {
				t.Fatalf("create %s declared full-refresh connection: %v", binding.Record, err)
			}

			fixture.resetCalls()
			run, err := application.RunETL(ctx, app.RunETLRequest{Connection: connection.Name, Stream: binding.Record, BatchSize: 1})
			if err != nil {
				t.Fatalf("RunETL(%s): %v", binding.Record, err)
			}
			if run.Status != "completed" || run.RecordsRead != 2 || run.RecordsLoaded != 2 {
				t.Fatalf("RunETL(%s) = %+v, want two staged declared-stream records", binding.Record, run)
			}

			rows, err := application.QueryTable(ctx, app.QueryTableRequest{Connection: connection.Name, Table: "stripe_" + binding.Record + "_witness", Limit: 10})
			if err != nil {
				t.Fatalf("query %s DuckDB table: %v", binding.Record, err)
			}
			if len(rows) != 2 {
				t.Fatalf("%s DuckDB rows = %#v, want two provider records", binding.Record, rows)
			}
			for _, row := range rows {
				if id, _ := row["id"].(string); id == "" {
					t.Fatalf("%s DuckDB row lacks id: %#v", binding.Record, row)
				}
			}
			fixture.assertPaginatedETLCalls(t, binding)
		})
	}
}

// TestStripeDeclaredETLNonSuccessDoesNotMaterialize uses one declared stream.
// A provider non-success must be terminal before any source row reaches DuckDB.
func TestStripeDeclaredETLNonSuccessDoesNotMaterialize(t *testing.T) {
	fixture := newStripeTransportFixture(t)
	if len(fixture.etlBindings) == 0 {
		t.Fatal("Stripe declaration has no ETL stream")
	}
	binding := fixture.etlBindings[0]
	fixture.failETLRecord = binding.Record
	application := fixture.openApplication(t)
	ctx := context.Background()

	connection, err := application.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "stripe_failed_" + binding.Record + "_to_warehouse",
		Source:      app.EndpointConfig{Connector: "stripe", Credential: "stripe-local"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
		Streams: map[string]app.StreamConfig{binding.Record: {
			SyncMode:         "full_refresh_append",
			DestinationTable: "stripe_" + binding.Record + "_failure_witness",
		}},
	})
	if err != nil {
		t.Fatalf("create failing %s full-refresh connection: %v", binding.Record, err)
	}

	fixture.resetCalls()
	run, err := application.RunETL(ctx, app.RunETLRequest{Connection: connection.Name, Stream: binding.Record, BatchSize: 1})
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Fatalf("RunETL(%s) non-success error = %v, want fixture 502", binding.Record, err)
	}
	if run.Status != "failed" || run.RecordsRead != 0 || run.RecordsLoaded != 0 {
		t.Fatalf("failed Stripe run = %+v, want terminal failure before warehouse materialization", run)
	}
	if calls := fixture.snapshotETLCalls(); len(calls) == 0 {
		t.Fatal("failed Stripe declaration made no provider request")
	} else {
		for _, call := range calls {
			if call.Record != binding.Record || call.Query.Get("starting_after") != "" {
				t.Fatalf("failed Stripe calls = %#v, want only first-page %s requests", calls, binding.Record)
			}
		}
	}
	if rows, queryErr := application.QueryTable(ctx, app.QueryTableRequest{Connection: connection.Name, Table: "stripe_" + binding.Record + "_failure_witness", Limit: 10}); queryErr == nil || len(rows) != 0 {
		t.Fatalf("failed Stripe stream materialized rows=%#v err=%v, want no warehouse result", rows, queryErr)
	}
}

// TestStripeDeclaredReverseETLPlanPreviewApprovalAndExecute exercises every
// currently declared Stripe customer action through the existing plan →
// preview → approval → execute path. It also verifies the typed delete gate
// refuses a mutation before the fixture observes it.
func TestStripeDeclaredReverseETLPlanPreviewApprovalAndExecute(t *testing.T) {
	fixture := newStripeTransportFixture(t)
	application := fixture.openApplication(t)
	ctx := context.Background()
	connection, sourceTable := fixture.materializeCustomerSource(t, application)

	for _, binding := range fixture.reverseBindings {
		binding := binding
		t.Run(binding.Record, func(t *testing.T) {
			fixture.resetCalls()
			plan, err := application.PlanReverseETL(ctx, app.PlanReverseETLRequest{
				Name:                  "stripe_" + binding.Record + "_witness",
				SourceTable:           sourceTable,
				SourceConnection:      connection,
				DestinationConnector:  "stripe",
				DestinationCredential: "stripe-local",
				Action:                binding.Record,
				Mappings:              stripeReverseWitnessMappings(t, binding.Record),
				Limit:                 1,
			})
			if err != nil {
				t.Fatalf("plan %s: %v", binding.Record, err)
			}
			if calls := fixture.snapshotWriteCalls(); len(calls) != 0 {
				t.Fatalf("plan %s made provider writes before preview/approval: %#v", binding.Record, calls)
			}

			approvalToken := plan.ApprovalToken
			previewedPlan, preview, err := application.PreviewReversePlan(ctx, plan.ID, nil)
			if err != nil {
				t.Fatalf("preview %s: %v", binding.Record, err)
			}
			if previewedPlan.ApprovalToken != "" {
				approvalToken = previewedPlan.ApprovalToken
			}
			if approvalToken == "" || preview.Digest == "" {
				t.Fatalf("preview %s = plan:%+v previewed:%+v preview:%+v, want an approval token and digest", binding.Record, plan, previewedPlan, preview)
			}
			if calls := fixture.snapshotWriteCalls(); len(calls) != 0 {
				t.Fatalf("preview %s made provider writes before approval: %#v", binding.Record, calls)
			}

			confirmation := connectors.WriteConfirmation{}
			if binding.Record == "delete_customer" {
				if _, err := application.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: approvalToken}); err == nil || !strings.Contains(strings.ToLower(err.Error()), "confirmation") {
					t.Fatalf("unconfirmed %s run error = %v, want typed destructive confirmation refusal", binding.Record, err)
				}
				if calls := fixture.snapshotWriteCalls(); len(calls) != 0 {
					t.Fatalf("unconfirmed %s reached provider: %#v", binding.Record, calls)
				}
				confirmation = connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
			}

			run, err := application.RunReverseETL(ctx, app.RunReverseETLRequest{
				PlanID:        plan.ID,
				ApprovalToken: approvalToken,
				Confirmation:  confirmation,
			})
			if err != nil {
				t.Fatalf("approved %s run: %v", binding.Record, err)
			}
			if run.Status != "completed" || run.RecordsStaged != 1 || run.RecordsSucceeded != 1 || run.RecordsFailed != 0 {
				t.Fatalf("approved %s run = %+v, want one completed declared write", binding.Record, run)
			}
			fixture.assertWriteCall(t, binding)
		})
	}
}

// TestStripePublicCommandsReachCredentialBoundaryWithBaseURLOverride makes
// every currently declared ETL/reverse-ETL public command pass ordinary command
// resolution with an intentionally unusable base URL. A missing credential
// stops each invocation before provider I/O; descriptor-specific failures are
// rejected because these legacy declarations carry no descriptor binding.
func TestStripePublicCommandsReachCredentialBoundaryWithBaseURLOverride(t *testing.T) {
	bundle := stripeTransportBundle(t)
	if bundle.CLISurface == nil {
		t.Fatal("Stripe CLI surface is absent")
	}
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("init Stripe command-boundary project: %v", err)
	}

	spy := &stripeNoNetworkTransportSpy{}
	oldTransport := http.DefaultTransport
	http.DefaultTransport = spy
	t.Cleanup(func() { http.DefaultTransport = oldTransport })

	public := make([]engine.CLICommand, 0, len(bundle.CLISurface.Commands))
	for _, command := range bundle.CLISurface.Commands {
		switch command.Intent {
		case "etl", "reverse_etl":
			if command.Availability != "implemented" || strings.TrimSpace(command.Path) == "" {
				t.Fatalf("Stripe public command = %+v, want a declared implemented command", command)
			}
			public = append(public, command)
		default:
			t.Fatalf("unexpected Stripe public command intent %q for %q", command.Intent, command.Path)
		}
	}
	if len(public) != len(bundle.Streams)+len(bundle.Writes) {
		t.Fatalf("Stripe public command cohort = %d, want every declared stream/write command (%d + %d)", len(public), len(bundle.Streams), len(bundle.Writes))
	}
	sort.Slice(public, func(i, j int) bool { return public[i].Path < public[j].Path })

	for _, command := range public {
		command := command
		t.Run(strings.ReplaceAll(command.Path, " ", "_"), func(t *testing.T) {
			args := append([]string{"stripe"}, strings.Fields(command.Path)...)
			args = append(args, "--root", root, "--config", "base_url=https://invalid.example")
			var stdout, stderr bytes.Buffer
			if code := cli.Run(args, &stdout, &stderr); code == 0 {
				t.Fatalf("Run(%v) unexpectedly succeeded; stdout=%s stderr=%s", args, stdout.String(), stderr.String())
			}
			output := strings.TrimSpace(stdout.String() + stderr.String())
			if !strings.Contains(output, "missing --credential") {
				t.Fatalf("Run(%v) = %q, want ordinary credential-bound refusal", args, output)
			}
			lower := strings.ToLower(output)
			for _, forbidden := range []string{"source-bound", "source descriptor", "source operation", "preflight"} {
				if strings.Contains(lower, forbidden) {
					t.Fatalf("Run(%v) reached a descriptor-specific refusal %q: %s", args, forbidden, output)
				}
			}
		})
	}
	if got := spy.requests.Load(); got != 0 {
		t.Fatalf("Stripe public command provider requests = %d, want zero", got)
	}
}

type stripeNoNetworkTransportSpy struct {
	requests atomic.Int64
}

func (spy *stripeNoNetworkTransportSpy) RoundTrip(*http.Request) (*http.Response, error) {
	spy.requests.Add(1)
	return nil, fmt.Errorf("unexpected Stripe provider I/O")
}

type stripeDeclaredBinding struct {
	Method     string
	Path       string
	Lane       string
	Record     string
	CLICommand string
}

type stripeTransportCall struct {
	Record string
	Method string
	Path   string
	Query  url.Values
	Form   url.Values
}

type stripeTransportFixture struct {
	t               *testing.T
	server          *httptest.Server
	etlBindings     []stripeDeclaredBinding
	reverseBindings []stripeDeclaredBinding
	byRoute         map[string]stripeDeclaredBinding

	mu            sync.Mutex
	failETLRecord string
	etlCalls      []stripeTransportCall
	writeCalls    []stripeTransportCall
}

func newStripeTransportFixture(t *testing.T) *stripeTransportFixture {
	t.Helper()
	etlBindings, reverseBindings := stripeDeclaredRuntimeBindings(t)
	if len(etlBindings) == 0 || len(reverseBindings) == 0 {
		t.Fatalf("Stripe declared runtime bindings = ETL:%d reverse:%d, want both non-empty", len(etlBindings), len(reverseBindings))
	}
	fixture := &stripeTransportFixture{
		t:               t,
		etlBindings:     etlBindings,
		reverseBindings: reverseBindings,
		byRoute:         make(map[string]stripeDeclaredBinding, len(etlBindings)+len(reverseBindings)),
	}
	for _, binding := range append(append([]stripeDeclaredBinding(nil), etlBindings...), reverseBindings...) {
		key := binding.Method + " " + binding.Path
		if previous, exists := fixture.byRoute[key]; exists {
			t.Fatalf("Stripe declared runtime bindings duplicate route %q for %s and %s", key, previous.Record, binding.Record)
		}
		fixture.byRoute[key] = binding
	}
	fixture.server = httptest.NewServer(http.HandlerFunc(fixture.serveHTTP))
	t.Cleanup(fixture.server.Close)
	return fixture
}

func stripeTransportBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundle, err := engine.Load(os.DirFS(stripeTransportDefsRoot), "stripe")
	if err != nil {
		t.Fatalf("load Stripe declaration bundle: %v", err)
	}
	return bundle
}

func stripeDeclaredRuntimeBindings(t *testing.T) ([]stripeDeclaredBinding, []stripeDeclaredBinding) {
	t.Helper()
	bundle := stripeTransportBundle(t)
	if bundle.CLISurface == nil {
		t.Fatal("Stripe CLI surface is absent")
	}

	streams := make(map[string]engine.StreamSpec, len(bundle.Streams))
	for _, stream := range bundle.Streams {
		if _, duplicate := streams[stream.Name]; duplicate {
			t.Fatalf("Stripe bundle repeats stream %q", stream.Name)
		}
		streams[stream.Name] = stream
	}
	writes := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, action := range bundle.Writes {
		if _, duplicate := writes[action.Name]; duplicate {
			t.Fatalf("Stripe bundle repeats write action %q", action.Name)
		}
		writes[action.Name] = action
	}

	streamCommands := make(map[string]engine.CLICommand, len(streams))
	writeCommands := make(map[string]engine.CLICommand, len(writes))
	for _, command := range bundle.CLISurface.Commands {
		if command.Availability != "implemented" {
			t.Fatalf("Stripe command %q availability = %q, want implemented legacy declaration", command.Path, command.Availability)
		}
		switch command.Intent {
		case "etl":
			if command.Stream == "" || command.Write != "" {
				t.Fatalf("Stripe ETL command %q = %+v, want one stream only", command.Path, command)
			}
			if _, duplicate := streamCommands[command.Stream]; duplicate {
				t.Fatalf("Stripe stream %q has duplicate public commands", command.Stream)
			}
			streamCommands[command.Stream] = command
		case "reverse_etl":
			if command.Write == "" || command.Stream != "" {
				t.Fatalf("Stripe reverse-ETL command %q = %+v, want one write only", command.Path, command)
			}
			if _, duplicate := writeCommands[command.Write]; duplicate {
				t.Fatalf("Stripe write %q has duplicate public commands", command.Write)
			}
			writeCommands[command.Write] = command
		default:
			t.Fatalf("Stripe command %q intent = %q, want ETL or reverse ETL", command.Path, command.Intent)
		}
	}

	etlBindings := make([]stripeDeclaredBinding, 0, len(streams))
	for name, stream := range streams {
		command, exists := streamCommands[name]
		if !exists || stream.Records.Path != "data" || stream.Path == "" {
			t.Fatalf("Stripe stream %q lacks a usable legacy command/record path: stream=%+v command=%+v", name, stream, command)
		}
		etlBindings = append(etlBindings, stripeDeclaredBinding{
			Method:     http.MethodGet,
			Path:       "/v1" + stream.Path,
			Lane:       "etl",
			Record:     name,
			CLICommand: command.Path,
		})
	}
	for name := range streamCommands {
		if _, exists := streams[name]; !exists {
			t.Fatalf("Stripe public ETL command references unknown stream %q", name)
		}
	}

	reverseBindings := make([]stripeDeclaredBinding, 0, len(writes))
	for name, action := range writes {
		command, exists := writeCommands[name]
		if !exists || action.Method == "" || action.Path == "" {
			t.Fatalf("Stripe write %q lacks a usable legacy command/action path: action=%+v command=%+v", name, action, command)
		}
		reverseBindings = append(reverseBindings, stripeDeclaredBinding{
			Method:     action.Method,
			Path:       "/v1" + action.Path,
			Lane:       "reverse_etl",
			Record:     name,
			CLICommand: command.Path,
		})
	}
	for name := range writeCommands {
		if _, exists := writes[name]; !exists {
			t.Fatalf("Stripe public reverse-ETL command references unknown write %q", name)
		}
	}

	sort.Slice(etlBindings, func(i, j int) bool { return etlBindings[i].Record < etlBindings[j].Record })
	sort.Slice(reverseBindings, func(i, j int) bool { return reverseBindings[i].Record < reverseBindings[j].Record })
	return etlBindings, reverseBindings
}

func stripeTransportPathMatches(template, actual string) bool {
	templateParts := strings.Split(strings.Trim(template, "/"), "/")
	actualParts := strings.Split(strings.Trim(actual, "/"), "/")
	if len(templateParts) != len(actualParts) {
		return false
	}
	for index, part := range templateParts {
		if strings.Contains(part, "{") {
			if actualParts[index] == "" {
				return false
			}
			continue
		}
		if part != actualParts[index] {
			return false
		}
	}
	return true
}

func (f *stripeTransportFixture) openApplication(t *testing.T) *app.App {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("initialize Stripe witness project: %v", err)
	}
	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("open Stripe witness project: %v", err)
	}
	ctx := context.Background()
	if _, err := application.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "stripe-local",
		Connector: "stripe",
		Config:    map[string]string{"base_url": f.server.URL + "/v1"},
		Secrets:   map[string]string{"client_secret": stripeTransportToken},
	}); err != nil {
		t.Fatalf("add local Stripe credential: %v", err)
	}
	if _, err := application.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "warehouse-local",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("add local warehouse credential: %v", err)
	}
	return application
}

func (f *stripeTransportFixture) materializeCustomerSource(t *testing.T, application *app.App) (string, string) {
	t.Helper()
	var customer stripeDeclaredBinding
	for _, binding := range f.etlBindings {
		if binding.Record == "customers" {
			customer = binding
			break
		}
	}
	if customer.Record == "" {
		t.Fatal("Stripe reverse witness requires the declared customers ETL stream")
	}
	const connectionName = "stripe_customers_reverse_source"
	const tableName = "stripe_customers_reverse_witness"
	connection, err := application.CreateConnection(context.Background(), app.CreateConnectionRequest{
		Name:        connectionName,
		Source:      app.EndpointConfig{Connector: "stripe", Credential: "stripe-local"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
		Streams: map[string]app.StreamConfig{customer.Record: {
			SyncMode:         "full_refresh_append",
			DestinationTable: tableName,
		}},
	})
	if err != nil {
		t.Fatalf("create Stripe customers connection: %v", err)
	}
	f.resetCalls()
	run, err := application.RunETL(context.Background(), app.RunETLRequest{Connection: connection.Name, Stream: customer.Record, BatchSize: 1})
	if err != nil || run.Status != "completed" || run.RecordsLoaded != 2 {
		t.Fatalf("materialize Stripe customers stream run=%+v err=%v", run, err)
	}
	f.assertPaginatedETLCalls(t, customer)
	f.resetCalls()
	return connection.Name, tableName
}

func stripeReverseWitnessMappings(t *testing.T, action string) map[string]string {
	t.Helper()
	switch action {
	case "create_customer":
		return map[string]string{"name": "name"}
	case "update_customer":
		return map[string]string{"id": "id", "name": "name"}
	case "delete_customer":
		return map[string]string{"id": "id"}
	default:
		t.Fatalf("Stripe declaration selected unsupported reverse witness action %q", action)
		return nil
	}
}

func (f *stripeTransportFixture) serveHTTP(w http.ResponseWriter, request *http.Request) {
	if got := request.Header.Get("Authorization"); got != "Bearer "+stripeTransportToken {
		f.failf("Stripe fixture authorization = %q, want bearer test token", got)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	binding, exists := f.byRoute[request.Method+" "+request.URL.Path]
	if !exists {
		for _, candidate := range f.reverseBindings {
			if candidate.Method == request.Method && stripeTransportPathMatches(candidate.Path, request.URL.Path) {
				binding = candidate
				exists = true
				break
			}
		}
	}
	if !exists {
		f.failf("unexpected Stripe fixture request %s %s", request.Method, request.URL.String())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch binding.Lane {
	case "etl":
		f.serveETL(w, request, binding)
	case "reverse_etl":
		f.serveReverseWrite(w, request, binding)
	default:
		f.failf("unexpected Stripe fixture binding lane %q", binding.Lane)
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *stripeTransportFixture) serveETL(w http.ResponseWriter, request *http.Request, binding stripeDeclaredBinding) {
	query := request.URL.Query()
	f.recordETLCall(stripeTransportCall{Record: binding.Record, Method: request.Method, Path: request.URL.Path, Query: query})
	if query.Get("limit") != "100" || query.Get("created[gte]") != "" {
		f.failf("Stripe %s query = %q, want declared limit=100 without an undeclared incremental start", binding.Record, request.URL.RawQuery)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if binding.Record == f.failureRecord() {
		f.writeJSON(w, http.StatusBadGateway, map[string]any{"error": map[string]any{"message": "fixture stream failure"}})
		return
	}

	page := 0
	if after := query.Get("starting_after"); after != "" {
		if after != f.recordID(binding, 0) {
			f.failf("Stripe %s continuation = %q, want first-page id %q", binding.Record, after, f.recordID(binding, 0))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		page = 1
	}
	f.writeJSON(w, http.StatusOK, map[string]any{
		"data":     []any{f.record(binding, page)},
		"has_more": page == 0,
	})
}

func (f *stripeTransportFixture) serveReverseWrite(w http.ResponseWriter, request *http.Request, binding stripeDeclaredBinding) {
	if err := request.ParseForm(); err != nil {
		f.failf("parse Stripe %s form: %v", binding.Record, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	call := stripeTransportCall{Record: binding.Record, Method: request.Method, Path: request.URL.Path, Query: request.URL.Query(), Form: request.PostForm}
	f.recordWriteCall(call)
	if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/x-www-form-urlencoded") && binding.Record != "delete_customer" {
		f.failf("Stripe %s content type = %q, want declared form encoding", binding.Record, request.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	switch binding.Record {
	case "create_customer":
		if request.URL.Path != "/v1/customers" || request.PostForm.Get("name") != "Fixture Customer" {
			f.failf("Stripe create request = path:%q form:%q, want declared customer form", request.URL.Path, request.PostForm.Encode())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case "update_customer":
		if request.URL.Path != "/v1/customers/"+f.customerID() || request.PostForm.Get("name") != "Fixture Customer" || request.PostForm.Get("id") != "" {
			f.failf("Stripe update request = path:%q form:%q, want id path plus form mutation", request.URL.Path, request.PostForm.Encode())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case "delete_customer":
		if request.URL.Path != "/v1/customers/"+f.customerID() || len(request.PostForm) != 0 {
			f.failf("Stripe delete request = path:%q form:%q, want bodyless delete", request.URL.Path, request.PostForm.Encode())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		f.failf("Stripe reverse fixture has unknown action %q", binding.Record)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	f.writeJSON(w, http.StatusOK, map[string]any{"id": f.customerID(), "object": "customer", "deleted": binding.Record == "delete_customer"})
}

func (f *stripeTransportFixture) record(binding stripeDeclaredBinding, page int) map[string]any {
	record := map[string]any{
		"id":      f.recordID(binding, page),
		"object":  strings.TrimSuffix(binding.Record, "s"),
		"created": 1700000000 + page,
	}
	if binding.Record == "customers" {
		record["name"] = "Fixture Customer"
		record["email"] = "fixture@example.test"
		record["description"] = "local declared Stripe fixture"
		record["phone"] = "+15550100"
	}
	return record
}

func (f *stripeTransportFixture) recordID(binding stripeDeclaredBinding, page int) string {
	if binding.Record == "customers" && page == 0 {
		return f.customerID()
	}
	prefix := strings.TrimSuffix(binding.Record, "s")
	return fmt.Sprintf("%s_fixture_%d", prefix, page+1)
}

func (*stripeTransportFixture) customerID() string {
	return "cus_fixture_customer"
}

func (f *stripeTransportFixture) assertPaginatedETLCalls(t *testing.T, binding stripeDeclaredBinding) {
	t.Helper()
	calls := f.snapshotETLCalls()
	if len(calls) != 2 {
		t.Fatalf("Stripe %s requests = %#v, want declared first and starting_after pages", binding.Record, calls)
	}
	for index, call := range calls {
		if call.Record != binding.Record || call.Method != http.MethodGet || call.Path != binding.Path || call.Query.Get("limit") != "100" {
			t.Fatalf("Stripe %s request %d = %#v, want exact declared GET", binding.Record, index, call)
		}
	}
	if calls[0].Query.Get("starting_after") != "" || calls[1].Query.Get("starting_after") != f.recordID(binding, 0) {
		t.Fatalf("Stripe %s pagination calls = %#v, want id-cursor continuation", binding.Record, calls)
	}
}

func (f *stripeTransportFixture) assertWriteCall(t *testing.T, binding stripeDeclaredBinding) {
	t.Helper()
	calls := f.snapshotWriteCalls()
	if len(calls) != 1 || calls[0].Record != binding.Record || calls[0].Method != binding.Method || calls[0].Path != f.expectedWritePath(binding) {
		t.Fatalf("Stripe %s provider writes = %#v, want one exact declared request", binding.Record, calls)
	}
}

func (f *stripeTransportFixture) expectedWritePath(binding stripeDeclaredBinding) string {
	if binding.Record == "create_customer" {
		return "/v1/customers"
	}
	return "/v1/customers/" + f.customerID()
}

func (f *stripeTransportFixture) resetCalls() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.etlCalls = nil
	f.writeCalls = nil
}

func (f *stripeTransportFixture) recordETLCall(call stripeTransportCall) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.etlCalls = append(f.etlCalls, call)
}

func (f *stripeTransportFixture) recordWriteCall(call stripeTransportCall) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writeCalls = append(f.writeCalls, call)
}

func (f *stripeTransportFixture) snapshotETLCalls() []stripeTransportCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]stripeTransportCall(nil), f.etlCalls...)
}

func (f *stripeTransportFixture) snapshotWriteCalls() []stripeTransportCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]stripeTransportCall(nil), f.writeCalls...)
}

func (f *stripeTransportFixture) failureRecord() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.failETLRecord
}

func (f *stripeTransportFixture) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		f.failf("encode Stripe fixture response: %v", err)
	}
}

func (f *stripeTransportFixture) failf(format string, args ...any) {
	f.t.Errorf(format, args...)
}
