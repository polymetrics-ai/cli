package app_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"polymetrics.ai/internal/app"
)

func TestAsanaSavedTypedActionApproval206(t *testing.T) {
	var mu sync.Mutex
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 4097))
		if err != nil {
			t.Error(err)
		}
		mu.Lock()
		calls = append(calls, r.Method+" "+r.URL.RequestURI()+" "+string(body))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"data":{"gid":"created-local"}}`))
	}))
	defer server.Close()
	snapshot := func() []string { mu.Lock(); defer mu.Unlock(); return append([]string(nil), calls...) }
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	if _, err = a.AddCredential(ctx, app.AddCredentialRequest{Name: "asana-local", Connector: "asana", Config: map[string]string{"base_url": server.URL}, Secrets: map[string]string{"access_token": "synthetic-local-only"}}); err != nil {
		t.Fatal(err)
	}
	seedWarehouseTableRows(t, root, "tasks206", `{"data":{"name":"Alpha","workspace":"local-w"}}`, `{"data":{"name":"Bravo","workspace":"local-w"}}`)
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{Name: "asana206", SourceTable: "tasks206", DestinationConnector: "asana", DestinationCredential: "asana-local", Action: "create_task", Mappings: map[string]string{"data": "data"}})
	if err != nil {
		t.Fatal(err)
	}
	if plan.RecordCount != 2 {
		t.Fatalf("staged records=%d", plan.RecordCount)
	}
	approvalToken := plan.ApprovalToken
	_, _, err = a.PreviewReversePlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot()) != 0 {
		t.Fatal("plan/preview performed provider I/O")
	}
	reopened, err := app.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = reopened.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID}); err == nil {
		t.Fatal("missing approval accepted")
	}
	if len(snapshot()) != 0 {
		t.Fatal("missing approval sent provider request")
	}
	run, err := reopened.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: approvalToken})
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != "completed" || run.RecordsSucceeded != 2 {
		t.Fatalf("saved outcome=%+v", run)
	}
	want := []string{`POST /tasks {"data":{"name":"Alpha","workspace":"local-w"}}`, `POST /tasks {"data":{"name":"Bravo","workspace":"local-w"}}`}
	if !reflect.DeepEqual(snapshot(), want) {
		t.Fatalf("physical requests=%q want=%q", snapshot(), want)
	}
	_, _ = reopened.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: approvalToken})
	if !reflect.DeepEqual(snapshot(), want) {
		t.Fatal("completed approval replay sent a request")
	}
}
