package cli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	stripeTransportDefsRoot = "../connectors/defs"
	stripeTransportToken    = "fixture-stripe-source-token"
)

var stripeTransportPathParameter = regexp.MustCompile(`\{\{[^}]+\}\}|\{[^}]+\}`)

// TestStripeSourceBoundETLMaterializesDuckDB executes every current
// declaration-owned Stripe ETL binding through App.RunETL and a temporary
// DuckDB project. The matrix, source lock, stream definition, and CLI
// declaration choose the cohort; the test does not maintain a parallel source
// operation allow-list.
func TestStripeSourceBoundETLMaterializesDuckDB(t *testing.T) {
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
				t.Fatalf("create %s source-bound full-refresh connection: %v", binding.Record, err)
			}

			fixture.resetCalls()
			run, err := application.RunETL(ctx, app.RunETLRequest{Connection: connection.Name, Stream: binding.Record, BatchSize: 1})
			if err != nil {
				t.Fatalf("RunETL(%s): %v", binding.Record, err)
			}
			if run.Status != "completed" || run.RecordsRead != 2 || run.RecordsLoaded != 2 {
				t.Fatalf("RunETL(%s) = %+v, want two staged source records", binding.Record, run)
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
					t.Fatalf("%s DuckDB row lacks source id: %#v", binding.Record, row)
				}
			}
			fixture.assertPaginatedETLCalls(t, binding)
		})
	}
}

// TestStripeSourceBoundETLNonSuccessDoesNotMaterialize uses one source-bound
// stream selected from the same matrix cohort. A provider non-success must be
// terminal before any source row reaches DuckDB.
func TestStripeSourceBoundETLNonSuccessDoesNotMaterialize(t *testing.T) {
	fixture := newStripeTransportFixture(t)
	if len(fixture.etlBindings) == 0 {
		t.Fatal("Stripe matrix has no implemented ETL binding")
	}
	binding := fixture.etlBindings[0]
	fixture.failETLSourceID = binding.SourceID
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
		t.Fatal("failed Stripe source made no provider request")
	} else {
		for _, call := range calls {
			if call.SourceID != binding.SourceID || call.Query.Get("starting_after") != "" {
				t.Fatalf("failed Stripe source calls = %#v, want only first-page %s requests", calls, binding.SourceID)
			}
		}
	}
	if rows, queryErr := application.QueryTable(ctx, app.QueryTableRequest{Connection: connection.Name, Table: "stripe_" + binding.Record + "_failure_witness", Limit: 10}); queryErr == nil || len(rows) != 0 {
		t.Fatalf("failed Stripe source materialized rows=%#v err=%v, want no warehouse result", rows, queryErr)
	}
}

// TestStripeSourceBoundReverseETLPlanPreviewApprovalAndExecute proves every
// current source-bound Stripe customer action through the real plan → preview
// → approval → execute path. It also asserts that the typed destructive gate
// rejects delete before the fake provider observes a request.
func TestStripeSourceBoundReverseETLPlanPreviewApprovalAndExecute(t *testing.T) {
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
				t.Fatalf("approved %s run = %+v, want one completed source-bound write", binding.Record, run)
			}
			fixture.assertWriteCall(t, binding)
		})
	}
}

type stripeTransportMatrix struct {
	Operations []stripeTransportMatrixOperation `json:"operations"`
}

type stripeTransportMatrixOperation struct {
	SourceID string                      `json:"source_id"`
	Method   string                      `json:"method"`
	Path     string                      `json:"path"`
	Cells    []stripeTransportMatrixCell `json:"cells"`
}

type stripeTransportMatrixCell struct {
	Lane      string                           `json:"lane"`
	State     string                           `json:"state"`
	Reason    string                           `json:"reason"`
	Execution *stripeTransportExecutionBinding `json:"execution,omitempty"`
}

type stripeTransportExecutionBinding struct {
	Artifact   string `json:"artifact"`
	Record     string `json:"record"`
	CLICommand string `json:"cli_command"`
	Proof      string `json:"proof"`
}

type stripeTransportSourceLock struct {
	REST struct {
		Operations []stripeTransportSourceOperation `json:"operations"`
	} `json:"rest"`
}

type stripeTransportSourceOperation struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

type stripeRuntimeBinding struct {
	SourceID   string
	Method     string
	Path       string
	Lane       string
	Record     string
	CLICommand string
}

type stripeTransportCall struct {
	SourceID string
	Method   string
	Path     string
	Query    url.Values
	Form     url.Values
}

type stripeTransportFixture struct {
	t               *testing.T
	server          *httptest.Server
	etlBindings     []stripeRuntimeBinding
	reverseBindings []stripeRuntimeBinding
	byRoute         map[string]stripeRuntimeBinding

	mu              sync.Mutex
	failETLSourceID string
	etlCalls        []stripeTransportCall
	writeCalls      []stripeTransportCall
}

func newStripeTransportFixture(t *testing.T) *stripeTransportFixture {
	t.Helper()
	etlBindings, reverseBindings := stripeDeclaredRuntimeBindings(t)
	if len(etlBindings) == 0 || len(reverseBindings) == 0 {
		t.Fatalf("Stripe source-bound runtime bindings = ETL:%d reverse:%d, want both non-empty", len(etlBindings), len(reverseBindings))
	}
	fixture := &stripeTransportFixture{
		t:               t,
		etlBindings:     etlBindings,
		reverseBindings: reverseBindings,
		byRoute:         make(map[string]stripeRuntimeBinding, len(etlBindings)+len(reverseBindings)),
	}
	for _, binding := range append(append([]stripeRuntimeBinding(nil), etlBindings...), reverseBindings...) {
		key := binding.Method + " " + binding.Path
		if previous, exists := fixture.byRoute[key]; exists {
			t.Fatalf("Stripe runtime bindings duplicate route %q for %s and %s", key, previous.SourceID, binding.SourceID)
		}
		fixture.byRoute[key] = binding
	}
	fixture.server = httptest.NewServer(http.HandlerFunc(fixture.serveHTTP))
	t.Cleanup(fixture.server.Close)
	return fixture
}

func stripeDeclaredRuntimeBindings(t *testing.T) ([]stripeRuntimeBinding, []stripeRuntimeBinding) {
	t.Helper()
	bundle, err := engine.Load(os.DirFS(stripeTransportDefsRoot), "stripe")
	if err != nil {
		t.Fatalf("load Stripe declaration bundle: %v", err)
	}
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
	commands := make(map[string]engine.CLICommand, len(bundle.CLISurface.Commands))
	for _, command := range bundle.CLISurface.Commands {
		if _, duplicate := commands[command.Path]; duplicate {
			t.Fatalf("Stripe CLI repeats command %q", command.Path)
		}
		commands[command.Path] = command
	}

	lockRaw, err := os.ReadFile(filepath.Join(stripeTransportDefsRoot, "stripe", "sources", "stripe-operation-source-lock.json"))
	if err != nil {
		t.Fatalf("read Stripe source lock: %v", err)
	}
	var lock stripeTransportSourceLock
	if err := json.Unmarshal(lockRaw, &lock); err != nil {
		t.Fatalf("decode Stripe source lock: %v", err)
	}
	locked := make(map[string]stripeTransportSourceOperation, len(lock.REST.Operations))
	for _, operation := range lock.REST.Operations {
		if _, duplicate := locked[operation.ID]; duplicate {
			t.Fatalf("Stripe source lock repeats operation %q", operation.ID)
		}
		locked[operation.ID] = operation
	}

	matrixRaw, err := os.ReadFile(filepath.Join(stripeTransportDefsRoot, "stripe", "sources", "stripe-source-lane-matrix.json"))
	if err != nil {
		t.Fatalf("read Stripe source lane matrix: %v", err)
	}
	var matrix stripeTransportMatrix
	if err := json.Unmarshal(matrixRaw, &matrix); err != nil {
		t.Fatalf("decode Stripe source lane matrix: %v", err)
	}

	seen := make(map[string]struct{})
	var etlBindings, reverseBindings []stripeRuntimeBinding
	for _, row := range matrix.Operations {
		lockedOperation, exists := locked[row.SourceID]
		if !exists || row.Method != lockedOperation.Method || row.Path != lockedOperation.Path {
			t.Fatalf("Stripe matrix row %q does not preserve the locked source identity", row.SourceID)
		}
		for _, cell := range row.Cells {
			if cell.State != "implemented" {
				continue
			}
			if cell.Execution == nil || (cell.Lane != "etl" && cell.Lane != "reverse_etl") {
				t.Fatalf("Stripe implemented matrix cell %s/%s = %+v, want a source-bound ETL or reverse-ETL execution", row.SourceID, cell.Lane, cell)
			}
			key := row.SourceID + ":" + cell.Lane
			if _, duplicate := seen[key]; duplicate {
				t.Fatalf("Stripe matrix repeats implemented binding %q", key)
			}
			seen[key] = struct{}{}

			command, exists := commands[cell.Execution.CLICommand]
			if !exists || command.Availability != "implemented" || command.Intent != cell.Lane || command.SourceOperation != row.SourceID || command.SourceCLIPath != row.Method+" "+row.Path {
				t.Fatalf("Stripe matrix execution %s/%s CLI binding = %+v, want exact source-bound declared command", row.SourceID, cell.Lane, command)
			}
			binding := stripeRuntimeBinding{SourceID: row.SourceID, Method: row.Method, Path: row.Path, Lane: cell.Lane, Record: cell.Execution.Record, CLICommand: cell.Execution.CLICommand}
			switch cell.Lane {
			case "etl":
				stream, exists := streams[binding.Record]
				if !exists || cell.Execution.Artifact != "streams.json" || command.Stream != binding.Record || stream.Records.Path != "data" || row.Method != http.MethodGet || strings.TrimPrefix(row.Path, "/v1") != stream.Path {
					t.Fatalf("Stripe ETL binding %s = stream:%+v command:%+v, want an exact GET source/stream/CLI chain", row.SourceID, stream, command)
				}
				etlBindings = append(etlBindings, binding)
			case "reverse_etl":
				action, exists := writes[binding.Record]
				if !exists || cell.Execution.Artifact != "writes.json" || command.Write != binding.Record || action.Method != row.Method || stripeTransportCanonicalPath(action.Path) != stripeTransportCanonicalPath(strings.TrimPrefix(row.Path, "/v1")) {
					t.Fatalf("Stripe reverse-ETL binding %s = action:%+v command:%+v, want an exact source/write/CLI chain", row.SourceID, action, command)
				}
				reverseBindings = append(reverseBindings, binding)
			}
		}
	}
	sort.Slice(etlBindings, func(i, j int) bool { return etlBindings[i].Record < etlBindings[j].Record })
	sort.Slice(reverseBindings, func(i, j int) bool { return reverseBindings[i].Record < reverseBindings[j].Record })
	return etlBindings, reverseBindings
}

func stripeTransportCanonicalPath(path string) string {
	return stripeTransportPathParameter.ReplaceAllString(path, "{}")
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
	var customer stripeRuntimeBinding
	for _, binding := range f.etlBindings {
		if binding.Record == "customers" {
			customer = binding
			break
		}
	}
	if customer.Record == "" {
		t.Fatal("Stripe source-bound reverse witness requires the declared customers ETL stream")
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
		t.Fatalf("create source-bound Stripe customers connection: %v", err)
	}
	f.resetCalls()
	run, err := application.RunETL(context.Background(), app.RunETLRequest{Connection: connection.Name, Stream: customer.Record, BatchSize: 1})
	if err != nil || run.Status != "completed" || run.RecordsLoaded != 2 {
		t.Fatalf("materialize Stripe customers source run=%+v err=%v", run, err)
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
		t.Fatalf("Stripe source matrix selected unsupported reverse witness action %q", action)
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

func (f *stripeTransportFixture) serveETL(w http.ResponseWriter, request *http.Request, binding stripeRuntimeBinding) {
	query := request.URL.Query()
	f.recordETLCall(stripeTransportCall{SourceID: binding.SourceID, Method: request.Method, Path: request.URL.Path, Query: query})
	if query.Get("limit") != "100" || query.Get("created[gte]") != "" {
		f.failf("Stripe %s query = %q, want declared limit=100 without an undeclared incremental start", binding.SourceID, request.URL.RawQuery)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if binding.SourceID == f.failureSourceID() {
		f.writeJSON(w, http.StatusBadGateway, map[string]any{"error": map[string]any{"message": "fixture source failure"}})
		return
	}

	page := 0
	if after := query.Get("starting_after"); after != "" {
		if after != f.recordID(binding, 0) {
			f.failf("Stripe %s continuation = %q, want first-page id %q", binding.SourceID, after, f.recordID(binding, 0))
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

func (f *stripeTransportFixture) serveReverseWrite(w http.ResponseWriter, request *http.Request, binding stripeRuntimeBinding) {
	if err := request.ParseForm(); err != nil {
		f.failf("parse Stripe %s form: %v", binding.SourceID, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	call := stripeTransportCall{SourceID: binding.SourceID, Method: request.Method, Path: request.URL.Path, Query: request.URL.Query(), Form: request.PostForm}
	f.recordWriteCall(call)
	if strings.HasPrefix(request.Header.Get("Content-Type"), "application/x-www-form-urlencoded") == false && binding.Record != "delete_customer" {
		f.failf("Stripe %s content type = %q, want declared form encoding", binding.SourceID, request.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	switch binding.Record {
	case "create_customer":
		if request.URL.Path != "/v1/customers" || request.PostForm.Get("name") != "Fixture Customer" {
			f.failf("Stripe create request = path:%q form:%q, want source-bound customer form", request.URL.Path, request.PostForm.Encode())
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
			f.failf("Stripe delete request = path:%q form:%q, want source-bound bodyless delete", request.URL.Path, request.PostForm.Encode())
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

func (f *stripeTransportFixture) record(binding stripeRuntimeBinding, page int) map[string]any {
	record := map[string]any{
		"id":      f.recordID(binding, page),
		"object":  strings.TrimSuffix(binding.Record, "s"),
		"created": 1700000000 + page,
	}
	if binding.Record == "customers" {
		record["name"] = "Fixture Customer"
		record["email"] = "fixture@example.test"
		record["description"] = "local source-bound Stripe fixture"
		record["phone"] = "+15550100"
	}
	return record
}

func (f *stripeTransportFixture) recordID(binding stripeRuntimeBinding, page int) string {
	if binding.Record == "customers" && page == 0 {
		return f.customerID()
	}
	prefix := strings.TrimSuffix(binding.Record, "s")
	return fmt.Sprintf("%s_fixture_%d", prefix, page+1)
}

func (*stripeTransportFixture) customerID() string {
	return "cus_fixture_customer"
}

func (f *stripeTransportFixture) assertPaginatedETLCalls(t *testing.T, binding stripeRuntimeBinding) {
	t.Helper()
	calls := f.snapshotETLCalls()
	if len(calls) != 2 {
		t.Fatalf("Stripe %s requests = %#v, want declared first and starting_after pages", binding.SourceID, calls)
	}
	for index, call := range calls {
		if call.SourceID != binding.SourceID || call.Method != http.MethodGet || call.Path != binding.Path || call.Query.Get("limit") != "100" {
			t.Fatalf("Stripe %s request %d = %#v, want exact declared source GET", binding.SourceID, index, call)
		}
	}
	if calls[0].Query.Get("starting_after") != "" || calls[1].Query.Get("starting_after") != f.recordID(binding, 0) {
		t.Fatalf("Stripe %s pagination calls = %#v, want source id-cursor continuation", binding.SourceID, calls)
	}
}

func (f *stripeTransportFixture) assertWriteCall(t *testing.T, binding stripeRuntimeBinding) {
	t.Helper()
	calls := f.snapshotWriteCalls()
	if len(calls) != 1 || calls[0].SourceID != binding.SourceID || calls[0].Method != binding.Method || calls[0].Path != f.expectedWritePath(binding) {
		t.Fatalf("Stripe %s provider writes = %#v, want one exact source-bound request", binding.SourceID, calls)
	}
}

func (f *stripeTransportFixture) expectedWritePath(binding stripeRuntimeBinding) string {
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

func (f *stripeTransportFixture) failureSourceID() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.failETLSourceID
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
