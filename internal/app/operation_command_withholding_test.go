package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors/engine"
)

const widgetNameSentinel = "WIDGET-SENTINEL-VALUE"

// operationWithholdingApp mirrors setupRestWriteDemoAppWithBundle but also
// returns the project root, so a test can assert against the persisted bytes
// rather than an in-memory copy of the plan.
func operationWithholdingApp(t *testing.T, ctx context.Context, baseURL string, mutate func(*engine.Bundle)) (*app.App, string) {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	bundle, err := engine.Load(os.DirFS("testdata/bundles"), restWriteDemoConnector)
	if err != nil {
		t.Fatalf("engine.Load(%s): %v", restWriteDemoConnector, err)
	}
	if mutate != nil {
		mutate(&bundle)
	}
	a.Registry().Register(engine.New(bundle, nil))
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "restwrite-local",
		Connector: restWriteDemoConnector,
		Config:    map[string]string{"base_url": baseURL},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}
	return a, root
}

func declareOperationRedactFields(fields ...string) func(*engine.Bundle) {
	return func(bundle *engine.Bundle) {
		for i := range bundle.Operations {
			if bundle.Operations[i].ID == "restwrite-demo.widget-update" {
				bundle.Operations[i].SensitivePolicy = &engine.SensitivePolicySpec{
					RedactFields: append([]string(nil), fields...),
				}
			}
		}
	}
}

// TestOperationBackedPlanWithholdsDeclaredFields covers the direct_write half
// of the withholding invariant: an operation's redact list is resolved from its
// own sensitive_policy, withheld from state.json, and replayed at execute.
func TestOperationBackedPlanWithholdsDeclaredFields(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["name"] != widgetNameSentinel {
			t.Fatalf("dispatched body = %#v, want the re-supplied sentinel", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"updated":true}`))
	}))
	defer server.Close()

	a, root := operationWithholdingApp(t, ctx, server.URL, declareOperationRedactFields("name"))

	flags := map[string][]string{"id": {"w_1"}, "name": {widgetNameSentinel}}
	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Connector:  restWriteDemoConnector,
		Credential: "restwrite-local",
		Path:       []string{"widget", "update"},
		Flags:      flags,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if got := plan.RedactFields; len(got) != 1 || got[0] != "name" {
		t.Fatalf("plan.RedactFields = %#v, want the operation's [name]", got)
	}
	if raw := stateBytes(t, root); strings.Contains(raw, widgetNameSentinel) {
		t.Fatalf("state.json contains the withheld operation body field")
	}
	if _, present := plan.ConnectorCommandRecord["name"]; present {
		t.Fatalf("ConnectorCommandRecord = %#v, want name withheld", plan.ConnectorCommandRecord)
	}

	if _, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil); err == nil {
		t.Fatal("preview with no re-supply = nil error, want a withheld-field error")
	}
	previewed, preview, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, flags)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan with re-supply: %v", err)
	}
	if preview.Digest == "" {
		t.Fatalf("preview = %#v, want a bound digest", preview)
	}
	if calls != 0 {
		t.Fatalf("preview reached the network; calls = %d", calls)
	}
	if raw := stateBytes(t, root); strings.Contains(raw, widgetNameSentinel) {
		t.Fatalf("previewing re-persisted the withheld field to state.json")
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        previewed.ID,
		ApprovalToken: previewed.ApprovalToken,
		WithheldFlags: flags,
	})
	if err != nil {
		t.Fatalf("RunReverseETL with re-supply: %v", err)
	}
	if calls != 1 || run.Status != "completed" {
		t.Fatalf("run/calls = %#v/%d, want one completed direct write", run, calls)
	}
	if raw := stateBytes(t, root); strings.Contains(raw, widgetNameSentinel) {
		t.Fatalf("executing persisted the withheld field to state.json")
	}
}

// TestOperationBackedPlanIgnoresSameNamedWriteAction is the collision case.
// asana ships 11 names that exist as BOTH a write action and an operation ID,
// so an operation-backed plan must resolve its withhold set from the operation
// alone. This test fails if anyone reintroduces a fallback between namespaces.
func TestOperationBackedPlanIgnoresSameNamedWriteAction(t *testing.T) {
	ctx := context.Background()
	a, root := operationWithholdingApp(t, ctx, "http://127.0.0.1:1", func(bundle *engine.Bundle) {
		declareOperationRedactFields("name")(bundle)
		// Same string in the other namespace, with a different redact list.
		bundle.Writes = append(bundle.Writes, engine.WriteAction{
			Name:         "restwrite-demo.widget-update",
			Method:       http.MethodPatch,
			Path:         "/api/widgets/{{ config.base_url }}",
			RedactFields: []string{"id"},
		})
	})

	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Connector:  restWriteDemoConnector,
		Credential: "restwrite-local",
		Path:       []string{"widget", "update"},
		Flags:      map[string][]string{"id": {"w_1"}, "name": {widgetNameSentinel}},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if got := plan.RedactFields; len(got) != 1 || got[0] != "name" {
		t.Fatalf("plan.RedactFields = %#v, want the operation's [name], not the write action's [id]", got)
	}
	if raw := stateBytes(t, root); strings.Contains(raw, widgetNameSentinel) {
		t.Fatalf("state.json contains the operation's declared field")
	}
	// The write action's list must not have been applied: id is a path param
	// the plan still needs, and it is not the operation's declared field.
	if got := plan.ConnectorCommandPathParams["id"]; got != "w_1" {
		t.Fatalf("ConnectorCommandPathParams id = %#v, want it untouched by the write action's redact list", got)
	}
}

// TestOperationWithoutSensitivePolicyRoundTrips is the negative control: an
// operation declaring no sensitive_policy withholds nothing and still executes
// with no re-supply.
func TestOperationWithoutSensitivePolicyRoundTrips(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"updated":true}`))
	}))
	defer server.Close()

	a, root := operationWithholdingApp(t, ctx, server.URL, nil)

	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Connector:  restWriteDemoConnector,
		Credential: "restwrite-local",
		Path:       []string{"widget", "update"},
		Flags:      map[string][]string{"id": {"w_1"}, "name": {"Ada"}},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if len(plan.RedactFields) != 0 {
		t.Fatalf("plan.RedactFields = %#v, want none declared", plan.RedactFields)
	}
	if got := plan.ConnectorCommandRecord["name"]; got != "Ada" {
		t.Fatalf("ConnectorCommandRecord name = %#v, want it kept", got)
	}
	if raw := stateBytes(t, root); !strings.Contains(raw, "Ada") {
		t.Fatalf("state.json dropped an undeclared field")
	}
	previewed, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan: %v", err)
	}
	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: previewed.ID, ApprovalToken: previewed.ApprovalToken})
	if err != nil {
		t.Fatalf("RunReverseETL: %v", err)
	}
	if calls != 1 || run.Status != "completed" {
		t.Fatalf("run/calls = %#v/%d, want one completed direct write", run, calls)
	}
}
