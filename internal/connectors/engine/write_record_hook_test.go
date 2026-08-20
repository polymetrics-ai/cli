package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

type statefulRecordMapper struct{ calls int }

func (h *statefulRecordMapper) ConnectorName() string { return "acme" }
func (h *statefulRecordMapper) MapWriteRecord(_ WriteAction, rec connectors.Record) (connectors.Record, bool, error) {
	h.calls++
	mapped := connectors.Record{"generation": h.calls}
	for key, value := range rec {
		mapped[key] = value
	}
	return mapped, true, nil
}

func TestApprovedWriteMapperRunsOnceAndExecutionMatchesDigest(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		if err := json.NewDecoder(request.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	bundle := Bundle{
		Name: "acme", HTTP: HTTPBase{URL: server.URL},
		Writes: []WriteAction{{Name: "create", Kind: "create", Method: http.MethodPost, Path: "/records", BodyFields: []string{"generation"}}},
	}
	hook := &statefulRecordMapper{}
	result, err := Write(context.Background(), bundle, connectors.WriteRequest{Action: "create"}, []connectors.Record{{"id": "record-1"}}, hook)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if hook.calls != 1 {
		t.Fatalf("mapper calls = %d, want exactly one before preview", hook.calls)
	}
	if result.RecordsWritten != 1 || received["generation"] != float64(1) {
		t.Fatalf("result/request = %#v / %#v, want approved generation 1", result, received)
	}
}

type recordPinHook struct {
	executes bool
	claims   bool
}

func (h recordPinHook) ConnectorName() string { return "acme" }

func (h recordPinHook) ExecuteWrite(context.Context, WriteAction, connectors.Record, *Runtime) (bool, error) {
	return h.executes, nil
}

func (h recordPinHook) HandlesWriteAction(WriteAction) bool { return h.claims }

func (h recordPinHook) MapWriteRecord(action WriteAction, rec connectors.Record) (connectors.Record, bool, error) {
	if action.Name != "archive" {
		return rec, false, nil
	}
	pinned := connectors.Record{}
	for key, value := range rec {
		pinned[key] = value
	}
	pinned["archived"] = true
	return pinned, true, nil
}

func destructivePinBundle() Bundle {
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: "https://api.acme.test"},
		Writes: []WriteAction{{
			Name:       "archive",
			Kind:       "update",
			Method:     http.MethodPatch,
			Path:       "/repos/acme/widgets",
			BodyFields: []string{"archived"},
			Confirm:    string(connectors.ConfirmationKindDestructive),
			Hook:       "acme",
		}},
	}
}

func destructivePinRequest(t *testing.T) connectors.WriteRequest {
	t.Helper()
	authority := fixtureWriteApprovalAuthority(t)
	revision, err := authority.CredentialRevision("acme-fixture", nil)
	if err != nil {
		t.Fatalf("CredentialRevision() error = %v", err)
	}
	digest, err := authority.ConfigurationDigest("acme-fixture", nil)
	if err != nil {
		t.Fatalf("ConfigurationDigest() error = %v", err)
	}
	return connectors.WriteRequest{
		Action: "archive",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  revision,
			ConfigurationDigest: digest,
			WriteApprovalScope:  connectors.WriteApprovalScopeProject,
		},
	}
}

// A destructive action whose hook overrides execution has no exact preview: the
// operator would approve the declarative body and the hook would send another.
// The engine refuses to prepare it at all, and this pins that refusal so the
// alternative below stays the only way to pin a constant on a destructive write.
func TestPrepareDeclarativeWriteRefusesDestructiveHookExecutedAction(t *testing.T) {
	_, err := prepareDeclarativeWrite(context.Background(), destructivePinBundle(), destructivePinRequest(t),
		[]connectors.Record{{}}, recordPinHook{executes: true, claims: true})
	if err == nil {
		t.Fatal("prepareDeclarativeWrite() prepared a destructive hook-executed action")
	}
	if !strings.Contains(err.Error(), "exact prepared-request preview") {
		t.Fatalf("prepareDeclarativeWrite() error = %v, want the exact-preview refusal", err)
	}
}

// A hook that only pins a record field leaves execution declarative, so the
// prepared body an operator approves is the body that runs. That is what lets a
// name-pinned action such as github's repo archive carry a typed confirmation.
func TestPrepareDeclarativeWritePinsRecordFieldsIntoTheApprovedBody(t *testing.T) {
	prepared, err := prepareDeclarativeWrite(context.Background(), destructivePinBundle(), destructivePinRequest(t),
		[]connectors.Record{{}}, recordPinHook{})
	if err != nil {
		t.Fatalf("prepareDeclarativeWrite() error = %v", err)
	}
	if len(prepared.Requests) != 1 {
		t.Fatalf("prepared requests = %d, want 1", len(prepared.Requests))
	}
	if got := prepared.Requests[0].Body; got != `{"archived":true}` {
		t.Fatalf("prepared body = %s, want the pinned field", got)
	}
	if !prepared.Target.RequiresApproval() {
		t.Fatal("prepared target dropped the declared destructive confirmation")
	}
}

// The mapping must not reach back into the caller's records: reverse ETL stages
// rows once and replays them across preview and execution.
func TestApplyWriteRecordHookLeavesCallerRecordsUntouched(t *testing.T) {
	original := connectors.Record{"id": "1"}
	mapped, err := applyWriteRecordHook(recordPinHook{}, destructivePinBundle().Writes[0], []connectors.Record{original})
	if err != nil {
		t.Fatalf("applyWriteRecordHook() error = %v", err)
	}
	if mapped[0]["archived"] != true {
		t.Fatalf("mapped record = %+v, want the pinned field", mapped[0])
	}
	if _, ok := original["archived"]; ok {
		t.Fatalf("caller record = %+v, want untouched", original)
	}
}

// A connector with no WriteRecordHook must see the records it staged, unchanged.
func TestApplyWriteRecordHookIsANoOpWithoutTheInterface(t *testing.T) {
	records := []connectors.Record{{"id": "1"}}
	mapped, err := applyWriteRecordHook(nil, destructivePinBundle().Writes[0], records)
	if err != nil {
		t.Fatalf("applyWriteRecordHook() error = %v", err)
	}
	if len(mapped) != 1 || mapped[0]["id"] != "1" || len(mapped[0]) != 1 {
		t.Fatalf("mapped records = %+v, want the staged records unchanged", mapped)
	}
}
