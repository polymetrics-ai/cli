package engine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/transportpolicy"
)

// newWriteTestBundle builds a minimal Bundle wired against srv with a single
// write action, defaults filled in by the caller via action.
func newWriteTestBundle(srv *httptest.Server, action WriteAction) Bundle {
	if action.Name == "" {
		action.Name = "update_widget"
	}
	if action.Method == "" {
		action.Method = http.MethodPost
	}
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Writes: []WriteAction{
			action,
		},
	}
}

func TestWriteIdempotencyKeySeparatesDurableWorksets(t *testing.T) {
	action := WriteAction{Name: "apply_widget", IdempotencyKeyHeader: "Idempotency-Key"}
	first := writeIdempotencyKey("acme", action, "sealed-preview", "connection-a/workset-one/checkpoint", 0)
	retry := writeIdempotencyKey("acme", action, "sealed-preview", "connection-a/workset-one/checkpoint", 0)
	second := writeIdempotencyKey("acme", action, "sealed-preview", "connection-a/workset-two/checkpoint", 0)
	if first == "" {
		t.Fatal("keyed action did not derive an idempotency key")
	}
	// These calls model distinct durable worksets carrying the same record at
	// index zero. A preview/body/index-only key aliases them at the provider.
	if first != retry {
		t.Fatalf("same durable workset retry changed provider key: %q != %q", first, retry)
	}
	if first == second {
		t.Fatalf("distinct durable worksets derived the same provider key %q", first)
	}
}

func TestWriteIdempotencyHeaderBindsDeliveryOccurrence(t *testing.T) {
	keys := make([]string, 0, 3)
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		keys = append(keys, request.Header.Get("Idempotency-Key"))
		writer.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	bundle := newWriteTestBundle(srv, WriteAction{
		Name: "apply_widget", Kind: "update", Method: http.MethodPost,
		Path: "/widgets/{{ record.id }}", PathFields: []string{"id"},
		IdempotencyKeyHeader: "Idempotency-Key",
	})
	records := []connectors.Record{{"id": "same-provider-payload"}}
	for _, occurrence := range []string{"connection-a/workset-one/checkpoint", "connection-a/workset-one/checkpoint", "connection-a/workset-two/checkpoint"} {
		if _, err := Write(context.Background(), bundle, connectors.WriteRequest{Action: "apply_widget", DeliveryOccurrence: occurrence}, records, nil); err != nil {
			t.Fatalf("Write(%q): %v", occurrence, err)
		}
	}
	if len(keys) != 3 || keys[0] == "" || keys[0] != keys[1] || keys[0] == keys[2] {
		t.Fatalf("provider idempotency headers = %#v, want stable retry key and distinct durable-workset key", keys)
	}
}

func writeSpecWithDefaultBaseURL(t *testing.T, defaultURL string) *Schema {
	t.Helper()
	rawDefault, err := json.Marshal(defaultURL)
	if err != nil {
		t.Fatalf("json.Marshal defaultURL: %v", err)
	}
	spec, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {
			"base_url": {"type": "string", "default": ` + string(rawDefault) + `}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	return spec
}

func captureServer(t *testing.T, status int, body string) (*httptest.Server, *capturedRequest) {
	t.Helper()
	cap := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.method = r.Method
		cap.path = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		cap.body = raw
		cap.contentType = r.Header.Get("Content-Type")
		if status != 0 {
			w.WriteHeader(status)
		}
		if body != "" {
			_, _ = w.Write([]byte(body))
		}
	}))
	t.Cleanup(srv.Close)
	return srv, cap
}

type capturedRequest struct {
	method      string
	path        string
	body        []byte
	contentType string
}

func TestWriteRejectsDestructiveActionWithoutTypedApprovalEvidence(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "remove_widget",
		Kind:       "custom",
		Method:     http.MethodDelete,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
		BodyType:   "none",
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "remove_widget", Config: connectors.RuntimeConfig{
		CredentialRevision: "fixture-credential-revision", ConfigurationDigest: "fixture-configuration-digest",
		WriteApprovalScope: connectors.WriteApprovalScopeFixture,
	}}, []connectors.Record{{"id": "42"}}, nil)
	if err == nil {
		t.Fatal("Write() dispatched DELETE without typed approval evidence")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "approval") {
		t.Fatalf("Write() error = %v, want approval rejection", err)
	}
	if calls != 0 {
		t.Fatalf("DELETE dispatched before approval gate; calls=%d", calls)
	}
}

func TestApprovedDestructiveWriteRefusesRedirectToUnapprovedTarget(t *testing.T) {
	approvedCalls := 0
	unapprovedCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/approved/42":
			approvedCalls++
			http.Redirect(w, r, "/unapproved/42", http.StatusTemporaryRedirect)
		case "/unapproved/42":
			unapprovedCalls++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer srv.Close()
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "remove_widget",
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/approved/{{ record.id }}",
		PathFields: []string{"id"},
		BodyType:   "none",
	})
	records := []connectors.Record{{"id": "42"}}
	req := approvedWriteRequest(t, b, "remove_widget", records, nil)

	if _, err := Write(context.Background(), b, req, records, nil); err == nil || !strings.Contains(strings.ToLower(err.Error()), "redirect") {
		t.Fatalf("Write() error = %v, want redirect refusal", err)
	}
	if approvedCalls != 1 {
		t.Fatalf("approved target calls = %d, want 1", approvedCalls)
	}
	if unapprovedCalls != 0 {
		t.Fatalf("unapproved redirect target calls = %d, want 0", unapprovedCalls)
	}
}

func TestGateRejectsForgedAndReplayedDestructiveEvidence(t *testing.T) {
	prepared := PreparedWrite{Target: DestructiveTarget{
		Connector:     "acme",
		Operation:     "delete_widget",
		Method:        http.MethodDelete,
		MutationClass: "delete",
		Confirmation:  connectors.ConfirmationKindDestructive,
	}, CredentialRevision: "fixture-credential-revision", ConfigurationDigest: "fixture-configuration-digest", ApprovalScope: connectors.WriteApprovalScopeFixture, Batchable: false, RecordsStaged: 1, Action: "delete_widget", Definition: map[string]any{"kind": "delete"}, Requests: []PreparedRequest{{Method: http.MethodDelete, URL: "http://127.0.0.1/widgets/42", Target: "http://127.0.0.1/widgets/42"}}}
	preview, err := PreviewPreparedWrite(prepared)
	if err != nil {
		t.Fatalf("PreviewPreparedWrite() error = %v", err)
	}

	t.Run("forged", func(t *testing.T) {
		executed := false
		err := ExecutePreparedWrite(context.Background(), prepared, &connectors.WriteApprovalEvidence{}, preview.Digest, func(context.Context) error {
			executed = true
			return nil
		})
		if err == nil || executed {
			t.Fatal("GateDestructiveExecution() accepted caller-minted evidence")
		}
	})

	t.Run("caller-selected-authority", func(t *testing.T) {
		authority, err := connectors.NewUntrustedWriteApprovalAuthority(bytes.Repeat([]byte{0x31}, sha256.Size))
		if err != nil {
			t.Fatalf("NewUntrustedWriteApprovalAuthority() error = %v", err)
		}
		evidence := approvedEvidenceWithAuthority(t, authority, preview)
		executed := false
		err = ExecutePreparedWrite(context.Background(), prepared, evidence, preview.Digest, func(context.Context) error {
			executed = true
			return nil
		})
		if err == nil || executed {
			t.Fatal("ExecutePreparedWrite() trusted caller-selected approval authority")
		}
	})

	t.Run("replayed", func(t *testing.T) {
		evidence := approvedEvidenceForPreview(t, preview)
		copiedEvidence := *evidence
		var executions int
		for _, attemptEvidence := range []*connectors.WriteApprovalEvidence{evidence, &copiedEvidence} {
			_ = ExecutePreparedWrite(context.Background(), prepared, attemptEvidence, preview.Digest, func(context.Context) error {
				executions++
				return nil
			})
		}
		if executions != 1 {
			t.Fatalf("approved closure executions = %d, want exactly one", executions)
		}
	})
}

func TestGateBindsConfigurationAndBatchability(t *testing.T) {
	prepared := PreparedWrite{
		Target: DestructiveTarget{
			Connector: "acme", Operation: "delete_widget", Method: http.MethodDelete,
			MutationClass: "delete", Confirmation: connectors.ConfirmationKindDestructive,
		},
		CredentialRevision:  "fixture-credential-revision",
		ConfigurationDigest: "fixture-configuration-one",
		ApprovalScope:       connectors.WriteApprovalScopeFixture,
		Batchable:           false,
		RecordsStaged:       1,
		Action:              "delete_widget",
		Definition:          map[string]any{"kind": "delete"},
		Requests: []PreparedRequest{{
			Method: http.MethodDelete, URL: "http://127.0.0.1/widgets/42", Target: "http://127.0.0.1/widgets/42",
		}},
	}
	preview, err := PreviewPreparedWrite(prepared)
	if err != nil {
		t.Fatalf("PreviewPreparedWrite() error = %v", err)
	}
	for _, mutate := range []struct {
		name string
		fn   func(*PreparedWrite)
	}{
		{name: "configuration", fn: func(changed *PreparedWrite) { changed.ConfigurationDigest = "fixture-configuration-two" }},
		{name: "batchability", fn: func(changed *PreparedWrite) { changed.Batchable = true }},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			changed := prepared
			mutate.fn(&changed)
			executed := false
			err := ExecutePreparedWrite(context.Background(), changed, approvedEvidenceForPreview(t, preview), preview.Digest, func(context.Context) error {
				executed = true
				return nil
			})
			if err == nil || executed {
				t.Fatalf("ExecutePreparedWrite() accepted %s drift", mutate.name)
			}
		})
	}
}

func TestZeroRecordPreparedWriteNeverInvokesExecutor(t *testing.T) {
	prepared := PreparedWrite{
		Target: DestructiveTarget{
			Connector: "acme", Operation: "purge_widgets", Method: http.MethodPost,
			MutationClass: "delete", Confirmation: connectors.ConfirmationKindDestructive,
		},
		CredentialRevision:  "fixture-credential-revision",
		ConfigurationDigest: "fixture-configuration-digest",
		ApprovalScope:       connectors.WriteApprovalScopeFixture,
		Action:              "purge_widgets",
		Definition:          map[string]any{"kind": "delete"},
	}
	preview, err := PreviewPreparedWrite(prepared)
	if err != nil {
		t.Fatalf("PreviewPreparedWrite() error = %v", err)
	}
	executed := false
	err = ExecutePreparedWrite(context.Background(), prepared, approvedEvidenceForPreview(t, preview), preview.Digest, func(context.Context) error {
		executed = true
		return nil
	})
	if err != nil {
		t.Fatalf("ExecutePreparedWrite() error = %v", err)
	}
	if executed {
		t.Fatal("ExecutePreparedWrite() invoked executor for an empty prepared write")
	}
}

func TestWriteFailsClosedForUnknownConfirmationDeclaration(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	b := newWriteTestBundle(srv, WriteAction{
		Name:     "dangerous_widget_action",
		Kind:     "custom",
		Method:   http.MethodPost,
		Path:     "/widgets",
		BodyType: "json",
		Confirm:  "type-anything",
	})

	if _, err := Write(context.Background(), b, connectors.WriteRequest{Action: "dangerous_widget_action", Config: connectors.RuntimeConfig{
		CredentialRevision: "fixture-credential-revision", ConfigurationDigest: "fixture-configuration-digest",
		WriteApprovalScope: connectors.WriteApprovalScopeFixture,
	}}, []connectors.Record{{"id": "42"}}, nil); err == nil {
		t.Fatal("Write() accepted an unknown non-empty confirmation declaration")
	}
	if calls != 0 {
		t.Fatalf("write dispatched after unknown confirmation declaration; calls=%d", calls)
	}
}

func TestRestWriteOperationUsesSharedDestructiveExecutionGate(t *testing.T) {
	executed := false
	operation := OperationSpec{
		ID:            "acme.widgets.delete",
		Kind:          "rest_write",
		MutationClass: "delete",
		REST:          &RESTOperationSpec{Method: http.MethodDelete, Path: "/widgets/{id}"},
		Confirmation:  &ConfirmationSpec{Kind: "destructive"},
	}
	prepared := PreparedWrite{
		Target:              DestructiveTargetForOperation("acme", operation),
		CredentialRevision:  "fixture-credential-revision",
		ConfigurationDigest: "fixture-configuration-digest",
		ApprovalScope:       connectors.WriteApprovalScopeFixture,
		Batchable:           false,
		RecordsStaged:       1,
		Action:              operation.ID,
		Definition:          operation,
		Requests:            []PreparedRequest{{Method: http.MethodDelete, URL: "http://127.0.0.1/widgets/42", Target: "http://127.0.0.1/widgets/42"}},
	}
	preview, err := PreviewPreparedWrite(prepared)
	if err != nil {
		t.Fatalf("PreviewPreparedWrite(rest_write): %v", err)
	}
	if err := ExecutePreparedWrite(context.Background(), prepared, nil, preview.Digest, func(context.Context) error {
		executed = true
		return nil
	}); err == nil {
		t.Fatal("unapproved rest_write closure executed without approval evidence")
	}
	if executed {
		t.Fatal("unapproved rest_write closure was invoked")
	}

	evidence := approvedEvidenceForPreview(t, preview)
	err = ExecutePreparedWrite(context.Background(), prepared, evidence, preview.Digest, func(executeCtx context.Context) error {
		if !transportpolicy.IsDestructive(executeCtx) {
			return errors.New("destructive transport policy was not propagated")
		}
		executed = true
		return nil
	})
	if err != nil {
		t.Fatalf("GateDestructiveExecution(rest_write): %v", err)
	}
	if !executed {
		t.Fatal("approved rest_write closure was not executed")
	}
}

func TestRestWriteDestructiveFlagCannotBeOverriddenByMutationClass(t *testing.T) {
	target := DestructiveTargetForOperation("acme", OperationSpec{
		ID:            "acme.widgets.admin_delete",
		Kind:          "rest_write",
		MutationClass: "admin",
		Destructive:   true,
		REST:          &RESTOperationSpec{Method: http.MethodPost, Path: "/widgets/delete"},
	})
	if !target.RequiresApproval() {
		t.Fatal("destructive operation flag was downgraded by a non-empty mutation class")
	}
}

func TestSecretOperationTypedConfirmationPolicyReachesSharedWriteGate(t *testing.T) {
	target := DestructiveTargetForOperation("acme", OperationSpec{
		ID:            "acme.organization.migration",
		Kind:          "graphql_mutation",
		MutationClass: "secret",
		SensitivePolicy: &SensitivePolicySpec{
			InputMode:    "env",
			ApprovalMode: "typed_confirmation",
		},
		GraphQL: &GraphQLOperationSpec{Path: "/graphql"},
	})
	if target.Confirmation != connectors.ConfirmationKindDestructive {
		t.Fatalf("secret operation confirmation = %q, want typed destructive confirmation", target.Confirmation)
	}
	if !target.RequiresApproval() {
		t.Fatal("secret operation typed-confirmation policy did not require approval")
	}
}

func TestPreparedWriteRejectsRequestMethodOutsideTargetPolicy(t *testing.T) {
	prepared := PreparedWrite{
		Target: DestructiveTarget{Connector: "acme", Operation: "delete_widget", Method: http.MethodPost},
		Action: "delete_widget",
		Requests: []PreparedRequest{{
			Method: http.MethodDelete,
			URL:    "https://api.example.test/widgets/42",
		}},
	}
	if _, err := PreviewPreparedWrite(prepared); err == nil {
		t.Fatal("PreviewPreparedWrite() accepted a DELETE outside the normalized target policy")
	}
}

func (c *capturedRequest) form() url.Values {
	v, _ := url.ParseQuery(string(c.body))
	return v
}

func (c *capturedRequest) json() map[string]any {
	var m map[string]any
	_ = json.Unmarshal(c.body, &m)
	return m
}

func approvedWriteRequest(t *testing.T, b Bundle, action string, records []connectors.Record, h Hooks) connectors.WriteRequest {
	t.Helper()
	authority := fixtureWriteApprovalAuthority(t)
	revision, err := authority.CredentialRevision("fixture-credential", nil)
	if err != nil {
		t.Fatalf("CredentialRevision(%s): %v", action, err)
	}
	configurationDigest, err := authority.ConfigurationDigest("fixture-credential", nil)
	if err != nil {
		t.Fatalf("ConfigurationDigest(%s): %v", action, err)
	}
	req := connectors.WriteRequest{Action: action, Config: connectors.RuntimeConfig{
		CredentialRevision: revision, ConfigurationDigest: configurationDigest,
		WriteApprovalScope: connectors.WriteApprovalScopeFixture,
	}}
	preview, err := DryRunWrite(context.Background(), b, req, records, h)
	if err != nil {
		t.Fatalf("DryRunWrite(%s): %v", action, err)
	}
	req.Approval = approvedEvidenceWithAuthority(t, authority, preview)
	return req
}

func fixtureWriteApprovalAuthority(t *testing.T) *connectors.WriteApprovalAuthority {
	t.Helper()
	authority, err := connectors.NewFixtureWriteApprovalAuthority()
	if err != nil {
		t.Fatalf("NewFixtureWriteApprovalAuthority() error = %v", err)
	}
	return authority
}

func approvedEvidenceForPreview(t *testing.T, preview connectors.WritePreview) *connectors.WriteApprovalEvidence {
	t.Helper()
	return approvedEvidenceWithAuthority(t, fixtureWriteApprovalAuthority(t), preview)
}

func approvedEvidenceWithAuthority(t *testing.T, authority *connectors.WriteApprovalAuthority, preview connectors.WritePreview) *connectors.WriteApprovalEvidence {
	t.Helper()
	token := "fixture-approval-token"
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID:        "rplan_fixture",
		PlanHash:      strings.Repeat("a", 64),
		PreviewDigest: preview.Digest,
		ApprovalToken: token,
		Target:        preview.ApprovalTarget,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("IssueWriteGrant() error = %v", err)
	}
	evidence, err := authority.VerifyWriteGrant(grant, connectors.WriteApprovalExpectation{
		PlanID:        grant.PlanID,
		PlanHash:      grant.PlanHash,
		PreviewDigest: preview.Digest,
		ApprovalToken: token,
		Target:        preview.ApprovalTarget,
		Confirmation:  grant.Confirmation,
	})
	if err != nil {
		t.Fatalf("VerifyWriteGrant() error = %v", err)
	}
	return evidence
}

func TestDryRunWriteMaterializesSpecDefaultBaseURL(t *testing.T) {
	b := Bundle{
		Name: "acme",
		Spec: writeSpecWithDefaultBaseURL(t, "https://api.example.test/v1"),
		HTTP: HTTPBase{URL: "{{ config.base_url }}"},
		Writes: []WriteAction{{
			Name:       "update_widget",
			Kind:       "update",
			Method:     http.MethodPost,
			Path:       "/widgets/{{ record.id }}",
			PathFields: []string{"id"},
		}},
	}

	preview, err := DryRunWrite(context.Background(), b, connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{{
		"id": "42",
	}}, nil)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	if len(preview.Warnings) < 2 || preview.Warnings[1] != "resolved request: POST https://api.example.test/v1/widgets/42" {
		t.Fatalf("warnings = %#v", preview.Warnings)
	}
}

func TestWriteMaterializesSpecDefaultBaseURL(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{"ok":true}`)
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "update",
		Method:     http.MethodPatch,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
	})
	b.Spec = writeSpecWithDefaultBaseURL(t, srv.URL)
	b.HTTP.URL = "{{ config.base_url }}"

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{
		{"id": "42", "name": "new-name"},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("result = %+v", result)
	}
	if cap.path != "/widgets/42" {
		t.Fatalf("path = %q, want /widgets/42", cap.path)
	}
}

func TestWritePreservesConfiguredBaseURLOverride(t *testing.T) {
	defaultSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer defaultSrv.Close()
	srv, cap := captureServer(t, http.StatusOK, `{"ok":true}`)
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "update",
		Method:     http.MethodPatch,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
	})
	b.Spec = writeSpecWithDefaultBaseURL(t, defaultSrv.URL)
	b.HTTP.URL = "{{ config.base_url }}"

	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}}
	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_widget", Config: cfg}, []connectors.Record{
		{"id": "42", "name": "new-name"},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("result = %+v", result)
	}
	if cap.path != "/widgets/42" {
		t.Fatalf("path = %q, want /widgets/42", cap.path)
	}
}

// --- body construction: json default (record minus path_fields) ---

func TestWriteJSONBodyDefaultExcludesPathFields(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{"ok":true}`)
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "update",
		Method:     http.MethodPatch,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
	})

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_widget", Config: connectors.RuntimeConfig{}}, []connectors.Record{
		{"id": "42", "name": "new-name"},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("result = %+v", result)
	}
	if cap.method != http.MethodPatch {
		t.Fatalf("method = %q, want PATCH", cap.method)
	}
	if cap.path != "/widgets/42" {
		t.Fatalf("path = %q, want /widgets/42", cap.path)
	}
	got := cap.json()
	if _, ok := got["id"]; ok {
		t.Fatalf("body = %+v, id (a path_field) must not be in the body", got)
	}
	if got["name"] != "new-name" {
		t.Fatalf("body = %+v, want name=new-name", got)
	}
}

// --- body construction: form (stripe-shape) ---

func TestWriteFormBodyStripeShape(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{"id":"cus_1"}`)
	b := newWriteTestBundle(srv, WriteAction{
		Name:     "create_customer",
		Kind:     "create",
		Method:   http.MethodPost,
		Path:     "/customers",
		BodyType: "form",
		RecordSchema: json.RawMessage(`{
			"type": "object", "minProperties": 1,
			"properties": {
				"email": {"type": "string"}, "name": {"type": "string"},
				"description": {"type": "string"}, "phone": {"type": "string"}
			}
		}`),
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "create_customer"}, []connectors.Record{
		{"email": "ada@example.com", "name": "Ada Lovelace"},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if cap.contentType != "application/x-www-form-urlencoded" {
		t.Fatalf("content-type = %q, want form-urlencoded", cap.contentType)
	}
	form := cap.form()
	if form.Get("email") != "ada@example.com" || form.Get("name") != "Ada Lovelace" {
		t.Fatalf("form = %v", form)
	}
}

// --- body construction: none + body_fields (delete-with-body) ---

func TestWriteNoneBodyTypeSendsNoBody(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, "")
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
		BodyType:   "none",
	})

	records := []connectors.Record{{"id": "42"}}
	_, err := Write(context.Background(), b, approvedWriteRequest(t, b, "update_widget", records, nil), records, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len(cap.body) != 0 {
		t.Fatalf("body = %q, want empty for body_type none", string(cap.body))
	}
}

func TestWriteBodyFieldsAllowListForDeleteWithBody(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, "")
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "delete_file",
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/files/{{ record.path }}",
		PathFields: []string{"path"},
		BodyFields: []string{"message", "sha", "branch"},
	})

	records := []connectors.Record{{"path": "a.txt", "message": "remove file", "sha": "abc123", "branch": "main", "extra_untouched": "x"}}
	_, err := Write(context.Background(), b, approvedWriteRequest(t, b, "delete_file", records, nil), records, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	got := cap.json()
	if got["message"] != "remove file" || got["sha"] != "abc123" || got["branch"] != "main" {
		t.Fatalf("body = %+v, want body_fields present", got)
	}
	if _, ok := got["extra_untouched"]; ok {
		t.Fatalf("body = %+v, want only body_fields present", got)
	}
	if _, ok := got["path"]; ok {
		t.Fatalf("body = %+v, path is a path_field and must not appear in body", got)
	}
}

// --- body construction: GraphQL fixed document + declared variables ---

func TestWriteGraphQLBodyUsesFixedDocumentAndDeclaredVariables(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{"data":{"deleteIssue":{"clientMutationId":"pm"}}}`)
	b := newWriteTestBundle(srv, WriteAction{
		Name:     "delete_issue",
		Kind:     "delete",
		Method:   http.MethodPost,
		Path:     "/graphql",
		BodyType: "graphql",
		GraphQL: &GraphQLRequestSpec{
			Document:      "mutation DeleteIssue($issueId: ID!) { deleteIssue(input: {id: $issueId}) { clientMutationId } }",
			OperationName: "DeleteIssue",
			Variables: map[string]any{
				"issueId": "{{ record.issue_id }}",
			},
		},
		RecordSchema: json.RawMessage(`{
			"type": "object",
			"required": ["issue_id"],
			"properties": {"issue_id": {"type": "string"}}
		}`),
	})

	records := []connectors.Record{{"issue_id": "I_kwDO123"}}
	result, err := Write(context.Background(), b, approvedWriteRequest(t, b, "delete_issue", records, nil), records, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("result = %+v", result)
	}
	if cap.method != http.MethodPost || cap.path != "/graphql" {
		t.Fatalf("request = %s %s, want POST /graphql", cap.method, cap.path)
	}
	got := cap.json()
	if got["query"] != "mutation DeleteIssue($issueId: ID!) { deleteIssue(input: {id: $issueId}) { clientMutationId } }" {
		t.Fatalf("query = %v, want fixed bundle mutation", got["query"])
	}
	if got["operationName"] != "DeleteIssue" {
		t.Fatalf("operationName = %v, want DeleteIssue", got["operationName"])
	}
	vars, ok := got["variables"].(map[string]any)
	if !ok {
		t.Fatalf("variables = %#v, want object", got["variables"])
	}
	if vars["issueId"] != "I_kwDO123" {
		t.Fatalf("variables.issueId = %#v, want record issue id", vars["issueId"])
	}
}

func TestWriteGraphQLBodyIgnoresRecordQueryField(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{"data":{"closeIssue":{"clientMutationId":"pm"}}}`)
	b := newWriteTestBundle(srv, WriteAction{
		Name:     "close_issue",
		Kind:     "update",
		Method:   http.MethodPost,
		Path:     "/graphql",
		BodyType: "graphql",
		GraphQL: &GraphQLRequestSpec{
			Document:      "mutation CloseIssue($issueId: ID!) { closeIssue(input: {issueId: $issueId}) { clientMutationId } }",
			OperationName: "CloseIssue",
			Variables: map[string]any{
				"issueId": "{{ record.issue_id }}",
			},
		},
		RecordSchema: json.RawMessage(`{"type":"object","required":["issue_id"],"properties":{"issue_id":{"type":"string"},"query":{"type":"string"}}}`),
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "close_issue"}, []connectors.Record{
		{"issue_id": "I_kwDO123", "query": "mutation Unsafe { deleteRepository(input:{repositoryId:\"R\"}) { clientMutationId } }"},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	got := cap.json()
	if strings.Contains(got["query"].(string), "Unsafe") || strings.Contains(got["query"].(string), "deleteRepository") {
		t.Fatalf("query = %q, record-provided query must not override fixed bundle document", got["query"])
	}
}

func TestWriteGraphQLErrorsFailClosed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"cannot delete issue"}],"data":{"deleteIssue":null}}`))
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:     "delete_issue",
		Kind:     "delete",
		Method:   http.MethodPost,
		Path:     "/graphql",
		BodyType: "graphql",
		GraphQL: &GraphQLRequestSpec{
			Document:      "mutation DeleteIssue($issueId: ID!) { deleteIssue(input: {id: $issueId}) { clientMutationId } }",
			OperationName: "DeleteIssue",
			Variables: map[string]any{
				"issueId": "{{ record.issue_id }}",
			},
		},
		RecordSchema: json.RawMessage(`{"type":"object","required":["issue_id"],"properties":{"issue_id":{"type":"string"}}}`),
	})

	records := []connectors.Record{{"issue_id": "I_kwDO123"}}
	result, err := Write(context.Background(), b, approvedWriteRequest(t, b, "delete_issue", records, nil), records, nil)
	if err == nil {
		t.Fatalf("Write: want GraphQL errors[] to fail closed")
	}
	if result.RecordsWritten != 0 || result.RecordsFailed != 1 {
		t.Fatalf("result = %+v, want 0 written / 1 failed", result)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "graphql") || strings.Contains(err.Error(), "cannot delete issue") {
		t.Fatalf("error = %q, want a generic GraphQL failure", err.Error())
	}
	if len(result.ProviderResponses) != 1 || result.ProviderResponses[0].Status != http.StatusOK || result.ProviderResponses[0].BodyRaw != `{"errors":[{"message":"cannot delete issue"}],"data":{"deleteIssue":null}}` || result.ProviderResponses[0].BodyRawEncoding != "text" {
		t.Fatalf("GraphQL failed-write provider result = %#v, want complete response envelope", result.ProviderResponses)
	}
}

func TestWriteGraphQLPreservesExplicitNonJSONResponse(t *testing.T) {
	const providerResponse = `{"errors":[{"message":"provider text is not a GraphQL envelope"}]}`
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(providerResponse))
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:     "delete_issue",
		Kind:     "delete",
		Method:   http.MethodPost,
		Path:     "/graphql",
		BodyType: "graphql",
		GraphQL: &GraphQLRequestSpec{
			Document:      "mutation DeleteIssue($issueId: ID!) { deleteIssue(input: {id: $issueId}) { clientMutationId } }",
			OperationName: "DeleteIssue",
			Variables: map[string]any{
				"issueId": "{{ record.issue_id }}",
			},
		},
		RecordSchema: json.RawMessage(`{"type":"object","required":["issue_id"],"properties":{"issue_id":{"type":"string"}}}`),
	})
	records := []connectors.Record{{"issue_id": "I_kwDO123"}}
	result, err := Write(context.Background(), b, approvedWriteRequest(t, b, "delete_issue", records, nil), records, nil)
	if err != nil {
		t.Fatalf("Write() = %v", err)
	}
	if calls != 1 || result.RecordsWritten != 1 || len(result.ProviderResponses) != 1 {
		t.Fatalf("GraphQL non-JSON write = %#v calls=%d, want successful provider result", result, calls)
	}
	response := result.ProviderResponses[0]
	if !response.BodyPresent || response.BodyRaw != providerResponse || response.BodyRawEncoding != "text" || response.Body != providerResponse || response.BodyEncoding != "text" {
		t.Fatalf("GraphQL non-JSON provider response = %#v, want exact text response", response)
	}
}

// --- record_schema validation ---

func TestValidateWriteRecordSchemaErrorCarriesRecordIndex(t *testing.T) {
	b := Bundle{
		Name: "acme",
		Writes: []WriteAction{{
			Name:   "create_issue",
			Kind:   "create",
			Method: http.MethodPost,
			Path:   "/issues",
			RecordSchema: json.RawMessage(`{
				"type": "object", "required": ["title"],
				"properties": {"title": {"type": "string"}}
			}`),
		}},
	}

	err := ValidateWrite(context.Background(), b, connectors.WriteRequest{Action: "create_issue"}, []connectors.Record{
		{"title": "ok"},
		{"body": "missing title"},
	})
	if err == nil {
		t.Fatalf("ValidateWrite: want error for record 1 (0-indexed) missing required title")
	}
	if !strings.Contains(err.Error(), "1") {
		t.Fatalf("error = %q, want it to name the record index", err.Error())
	}
}

func TestValidateWriteRecordSchemaValidPasses(t *testing.T) {
	b := Bundle{
		Name: "acme",
		Writes: []WriteAction{{
			Name:   "create_issue",
			Kind:   "create",
			Method: http.MethodPost,
			Path:   "/issues",
			RecordSchema: json.RawMessage(`{
				"type": "object", "required": ["title"],
				"properties": {"title": {"type": "string"}}
			}`),
		}},
	}

	err := ValidateWrite(context.Background(), b, connectors.WriteRequest{Action: "create_issue"}, []connectors.Record{
		{"title": "ok"},
	})
	if err != nil {
		t.Fatalf("ValidateWrite: %v", err)
	}
}

// --- DryRunWrite ---

func TestWritePreviewRedactsResolvedSecretsButDigestBindsThem(t *testing.T) {
	b := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: "https://api.example.com/{{ secrets.client_secret }}"},
		Writes: []WriteAction{{
			Name:         "update_customer",
			Kind:         "update",
			Method:       http.MethodPost,
			Path:         "/customers/{{ record.id }}",
			PathFields:   []string{"id"},
			RecordSchema: json.RawMessage(`{"type": "object", "properties": {"id": {"type": "string"}, "name": {"type": "string"}}}`),
		}},
	}

	cfg := connectors.RuntimeConfig{Secrets: map[string]string{"client_secret": "fixture-preview-secret"}}
	preview, err := DryRunWrite(context.Background(), b, connectors.WriteRequest{Action: "update_customer", Config: cfg}, []connectors.Record{
		{"id": "cus_1", "name": "New Name"},
	}, nil)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	if preview.RecordsStaged != 1 {
		t.Fatalf("RecordsStaged = %d, want 1", preview.RecordsStaged)
	}
	if preview.Action != "update_customer" {
		t.Fatalf("Action = %q, want update_customer", preview.Action)
	}
	joined := strings.Join(preview.Warnings, " | ")
	if !strings.Contains(joined, "POST") || !strings.Contains(joined, "/customers/cus_1") {
		t.Fatalf("Warnings = %v, want resolved method+path", preview.Warnings)
	}
	if strings.Contains(joined, "fixture-preview-secret") || !strings.Contains(joined, "redacted") {
		t.Fatalf("Warnings = %v, want secret masked while ordinary customer ID remains", preview.Warnings)
	}

	changed, err := DryRunWrite(context.Background(), b, connectors.WriteRequest{Action: "update_customer", Config: connectors.RuntimeConfig{
		Secrets: map[string]string{"client_secret": "different-preview-secret"},
	}}, []connectors.Record{{"id": "cus_1", "name": "New Name"}}, nil)
	if err != nil {
		t.Fatalf("DryRunWrite(changed secret): %v", err)
	}
	if preview.Digest == changed.Digest {
		t.Fatal("preview digest did not bind the privately prepared secret-derived target")
	}
}

func TestWritePreviewRedactsUserinfoQueryAndDeclaredValuesButPreservesOrdinaryToken(t *testing.T) {
	b := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: "https://{{ secrets.user }}:{{ secrets.password }}@api.example.com"},
		Writes: []WriteAction{{
			Name:       "update_customer",
			Kind:       "update",
			Method:     http.MethodPost,
			Path:       "/customers/{{ record.private_id }}/{{ record.token }}",
			PathFields: []string{"private_id", "token"},
			RedactFields: []string{
				"private_id",
			},
			Query: map[string]QueryParam{
				"access": {Template: "{{ secrets.query_secret }}"},
			},
		}},
	}
	records := []connectors.Record{
		{"private_id": "private/id", "token": "ordinary-token-42"},
		{"private_id": "second-private-id", "token": "ordinary-token-43"},
	}
	preview, err := DryRunWrite(context.Background(), b, connectors.WriteRequest{Action: "update_customer", Config: connectors.RuntimeConfig{
		Secrets: map[string]string{
			"user":         "credential-user",
			"password":     "credential-password",
			"query_secret": "query secret/value",
		},
	}}, records, nil)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	joined := strings.Join(preview.Warnings, " | ")
	for _, secret := range []string{"credential-user", "credential-password", "query secret/value", "query+secret%2Fvalue", "private/id", "private%2Fid"} {
		if strings.Contains(joined, secret) {
			t.Fatalf("Warnings = %v, leaked %q", preview.Warnings, secret)
		}
	}
	if !strings.Contains(joined, "ordinary-token-42") {
		t.Fatalf("Warnings = %v, ordinary provider token-shaped ID was not preserved", preview.Warnings)
	}
}

func TestDryRunWriteDigestBindsCanonicalRequestAndCredentialRevision(t *testing.T) {
	records := []connectors.Record{{"id": "42", "keep": "yes", "drop": "no"}}
	base := Bundle{
		Name: "twilio",
		HTTP: HTTPBase{URL: "https://api.twilio.test/Accounts/{{ secrets.account_sid }}"},
		Writes: []WriteAction{{
			Name:       "delete_message",
			Kind:       "delete",
			Method:     http.MethodDelete,
			Path:       "/Messages/{{ record.id }}",
			PathFields: []string{"id"},
			BodyFields: []string{"keep"},
			Query:      map[string]QueryParam{"mode": {Template: "soft"}},
			Hook:       "twilio-v1",
		}},
	}

	preview := func(t *testing.T, bundle Bundle, sid string) connectors.WritePreview {
		t.Helper()
		authority := fixtureWriteApprovalAuthority(t)
		revision, err := authority.CredentialRevision("twilio-fixture", map[string]string{"account_sid": sid})
		if err != nil {
			t.Fatalf("CredentialRevision() error = %v", err)
		}
		configurationDigest, err := authority.ConfigurationDigest("twilio-fixture", nil)
		if err != nil {
			t.Fatalf("ConfigurationDigest() error = %v", err)
		}
		got, err := DryRunWrite(context.Background(), bundle, connectors.WriteRequest{
			Action: "delete_message",
			Config: connectors.RuntimeConfig{
				Secrets: map[string]string{"account_sid": sid}, CredentialRevision: revision,
				ConfigurationDigest: configurationDigest, WriteApprovalScope: connectors.WriteApprovalScopeProject,
			},
		}, records, nil)
		if err != nil {
			t.Fatalf("DryRunWrite() error = %v", err)
		}
		return got
	}

	original := preview(t, base, "AC-one")
	changedCredential := preview(t, base, "AC-two")
	if original.Digest == changedCredential.Digest {
		t.Fatal("preview digest did not bind the secret-derived account target")
	}

	changedQuery := base
	changedQuery.Writes = append([]WriteAction(nil), base.Writes...)
	changedQuery.Writes[0].Query = map[string]QueryParam{"mode": {Template: "hard"}}
	if original.Digest == preview(t, changedQuery, "AC-one").Digest {
		t.Fatal("preview digest did not bind the resolved query")
	}

	changedBody := base
	changedBody.Writes = append([]WriteAction(nil), base.Writes...)
	changedBody.Writes[0].BodyFields = []string{"drop"}
	if original.Digest == preview(t, changedBody, "AC-one").Digest {
		t.Fatal("preview digest did not bind body construction")
	}

	changedHook := base
	changedHook.Writes = append([]WriteAction(nil), base.Writes...)
	changedHook.Writes[0].Hook = "twilio-v2"
	if original.Digest == preview(t, changedHook, "AC-one").Digest {
		t.Fatal("preview digest did not bind hook identity")
	}
}

func TestDryRunWritePreviewResolvedPathRedactsConfiguredRecordFields(t *testing.T) {
	b := Bundle{
		Name: "clinical",
		HTTP: HTTPBase{URL: "https://api.example.com"},
		Writes: []WriteAction{{
			Name:         "update_patient",
			Kind:         "update",
			Method:       http.MethodPost,
			Path:         "/patients/{{ record.uuid }}",
			PathFields:   []string{"uuid"},
			RedactFields: []string{"uuid"},
			RecordSchema: json.RawMessage(`{"type": "object", "required": ["uuid"], "properties": {"uuid": {"type": "string"}}}`),
		}},
	}

	preview, err := DryRunWrite(context.Background(), b, connectors.WriteRequest{Action: "update_patient"}, []connectors.Record{
		{"uuid": "patient-raw-uuid"},
	}, nil)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	joined := strings.Join(preview.Warnings, " | ")
	if strings.Contains(joined, "patient-raw-uuid") || !strings.Contains(joined, "/patients/redacted") {
		t.Fatalf("Warnings = %v, want configured record path field masked", preview.Warnings)
	}
}

func TestDryRunWritePreviewResolvedPathRedactsNestedRecordFields(t *testing.T) {
	b := Bundle{
		Name: "clinical",
		HTTP: HTTPBase{URL: "https://api.example.com"},
		Writes: []WriteAction{{
			Name:         "update_patient",
			Kind:         "update",
			Method:       http.MethodPost,
			Path:         "/patients/{{ record.patient.uuid }}",
			PathFields:   []string{"patient.uuid"},
			RedactFields: []string{"patient.uuid"},
			RecordSchema: json.RawMessage(`{"type": "object", "required": ["patient"], "properties": {"patient": {"type": "object", "required": ["uuid"], "properties": {"uuid": {"type": "string"}}}}}`),
		}},
	}
	patient := map[string]any{"uuid": "patient-nested-uuid"}

	preview, err := DryRunWrite(context.Background(), b, connectors.WriteRequest{Action: "update_patient"}, []connectors.Record{
		{"patient": patient},
	}, nil)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	joined := strings.Join(preview.Warnings, " | ")
	if strings.Contains(joined, "patient-nested-uuid") || !strings.Contains(joined, "/patients/redacted") {
		t.Fatalf("Warnings = %v, want nested configured record path field masked", preview.Warnings)
	}
	if patient["uuid"] != "patient-nested-uuid" {
		t.Fatalf("DryRunWrite mutated caller record: %v", patient)
	}
}

// --- delete semantics: missing_ok_status ---

func TestWriteDelete_HappyDocumentedMissingOK404DoesNotCountAsWritten(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "delete_label",
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/labels/{{ record.name }}",
		PathFields: []string{"name"},
		Delete:     &DeleteSpec{Idempotent: true, MissingOkStatus: []int{404}},
	})

	records := []connectors.Record{{"name": "bug"}}
	result, err := Write(context.Background(), b, approvedWriteRequest(t, b, "delete_label", records, nil), records, nil)
	if err != nil {
		t.Fatalf("Write: %v (allow-listed missing response should remain an idempotent no-op)", err)
	}
	if result.RecordsWritten != 0 || result.RecordsFailed != 0 || result.RecordsUnchanged != 1 {
		t.Fatalf("result = %+v, want provider 404 counted as one unchanged record", result)
	}
}

func TestWriteMissingOKDeleteRequiresExactDeclaredJSONResponse(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		body    []byte
		wantErr bool
	}{
		{name: "one JSON value remains an unchanged record", body: []byte(`{"provider":"missing"}`)},
		{name: "empty JSON body"},
		{name: "whitespace JSON body", body: []byte(" \n\t "), wantErr: true},
		{name: "malformed JSON body", body: []byte(`{"provider":`), wantErr: true},
		{name: "multiple JSON values", body: []byte(`{"provider":"first"} {"provider":"second"}`), wantErr: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Provider-Receipt", "delete-receipt")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(testCase.body)
			}))
			t.Cleanup(srv.Close)
			bundle := newWriteTestBundle(srv, WriteAction{
				Name:       "delete_widget",
				Kind:       "delete",
				Method:     http.MethodDelete,
				Path:       "/widgets/{{ record.id }}",
				PathFields: []string{"id"},
				Delete:     &DeleteSpec{Idempotent: true, MissingOkStatus: []int{http.StatusNotFound}},
			})
			records := []connectors.Record{{"id": "widget-1"}}
			result, err := Write(context.Background(), bundle, approvedWriteRequest(t, bundle, "delete_widget", records, nil), records, nil)
			if testCase.wantErr {
				if err == nil || !strings.Contains(err.Error(), "provider response") {
					t.Fatalf("Write() error = %v, want declared JSON response failure", err)
				}
				if result.RecordsWritten != 0 || result.RecordsUnchanged != 0 || result.RecordsFailed != 1 {
					t.Fatalf("failed missing delete result = %#v, want one failed record", result)
				}
			} else {
				if err != nil {
					t.Fatalf("Write(): %v", err)
				}
				if result.RecordsWritten != 0 || result.RecordsUnchanged != 1 || result.RecordsFailed != 0 {
					t.Fatalf("unchanged missing delete result = %#v, want one unchanged record", result)
				}
			}
			if len(result.ProviderResponses) != 1 {
				t.Fatalf("provider responses = %#v, want one captured response", result.ProviderResponses)
			}
			provider := result.ProviderResponses[0]
			wantEncoding := "text"
			if len(testCase.body) == 0 {
				wantEncoding = ""
			}
			if provider.BodyPresent != (len(testCase.body) != 0) || provider.BodyBytes != len(testCase.body) || provider.BodyRaw != string(testCase.body) || provider.BodyRawEncoding != wantEncoding || provider.Status != http.StatusNotFound || !reflect.DeepEqual(provider.Headers["X-Provider-Receipt"].Values, []string{"delete-receipt"}) {
				t.Fatalf("provider response = %#v, want exact captured missing-delete response", provider)
			}
		})
	}
}

func TestWriteDeleteFailureAccountingExcludesPriorUnchangedRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/absent") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "delete_label",
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/labels/{{ record.name }}",
		PathFields: []string{"name"},
		Delete:     &DeleteSpec{Idempotent: true, MissingOkStatus: []int{404}},
	})

	records := []connectors.Record{{"name": "absent"}, {"name": "bad"}}
	result, err := Write(context.Background(), b, approvedWriteRequest(t, b, "delete_label", records, nil), records, nil)
	if err == nil {
		t.Fatal("Write: want the second delete failure")
	}
	if result.RecordsWritten != 0 || result.RecordsUnchanged != 1 || result.RecordsFailed != 1 {
		t.Fatalf("result = %+v, want 0 written / 1 unchanged / 1 failed", result)
	}
}

func TestWriteDelete_BadUndeclared404Fails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "delete_label",
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/labels/{{ record.name }}",
		PathFields: []string{"name"},
		Delete:     &DeleteSpec{Idempotent: false},
	})

	records := []connectors.Record{{"name": "bug"}}
	result, err := Write(context.Background(), b, approvedWriteRequest(t, b, "delete_label", records, nil), records, nil)
	if err == nil {
		t.Fatalf("Write: want error for 404 not in missing_ok_status")
	}
	if result.RecordsWritten != 0 || result.RecordsFailed != 1 {
		t.Fatalf("result = %+v, want 0 written / 1 failed", result)
	}
}

func TestWriteNonListedStatusFails(t *testing.T) {
	// A non-retryable client error (400) so the test does not incur real
	// connsdk retry/backoff sleeps (500 would trigger 4 retries by default).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "delete_label",
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/labels/{{ record.name }}",
		PathFields: []string{"name"},
		Delete:     &DeleteSpec{Idempotent: true, MissingOkStatus: []int{404}},
	})

	records := []connectors.Record{{"name": "bug"}}
	result, err := Write(context.Background(), b, approvedWriteRequest(t, b, "delete_label", records, nil), records, nil)
	if err == nil {
		t.Fatalf("Write: want error for 400 (not a missing_ok_status match)")
	}
	if result.RecordsWritten != 0 || result.RecordsFailed != 1 {
		t.Fatalf("result = %+v, want 0 written / 1 failed", result)
	}
}

func TestWriteErrorRedactsConfiguredRecordFieldsInHTTPPathAndBody(t *testing.T) {
	const rawPatientUUID = "patient/uuid with space"
	const rawAppointmentUUID = "appointment-raw-uuid"
	encodedPatientUUID := strings.ReplaceAll(url.QueryEscape(rawPatientUUID), "+", "%20")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"failed patient/uuid with space appointment-raw-uuid"}`))
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:         "update_appointment_status",
		Kind:         "update",
		Method:       http.MethodPost,
		Path:         "/openmrs/ws/rest/v1/appointment/{{ record.appointment.uuid }}/status/{{ record.uuid }}",
		PathFields:   []string{"appointment.uuid", "uuid"},
		RedactFields: []string{"appointment.uuid", "uuid"},
		RecordSchema: json.RawMessage(`{
			"type": "object",
			"required": ["uuid", "appointment"],
			"properties": {
				"uuid": {"type": "string"},
				"appointment": {"type": "object", "required": ["uuid"], "properties": {"uuid": {"type": "string"}}}
			}
		}`),
	})

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_appointment_status"}, []connectors.Record{{
		"uuid": rawPatientUUID,
		"appointment": map[string]any{
			"uuid": rawAppointmentUUID,
		},
	}}, nil)
	if err == nil {
		t.Fatalf("Write: want HTTP error")
	}
	if result.RecordsWritten != 0 || result.RecordsFailed != 1 {
		t.Fatalf("result = %+v, want 0 written / 1 failed", result)
	}
	msg := err.Error()
	for _, leaked := range []string{rawPatientUUID, encodedPatientUUID, rawAppointmentUUID} {
		if strings.Contains(msg, leaked) {
			t.Fatalf("write error leaked %q in %q", leaked, msg)
		}
	}
	if !strings.Contains(msg, "redacted") {
		t.Fatalf("write error %q missing redaction marker", msg)
	}
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != http.StatusBadRequest {
		t.Fatalf("errors.As HTTPError = %#v, want status 400", httpErr)
	}
}

func TestWriteErrorRedactsOverlappingConfiguredRecordFields(t *testing.T) {
	action := WriteAction{
		RedactFields: []string{"short_id", "long_id"},
	}
	err := redactWriteActionError(errors.New("write failed for abcdef and abc"), action, connectors.Record{
		"short_id": "abc",
		"long_id":  "abcdef",
	})
	if err == nil {
		t.Fatalf("redactWriteActionError: got nil error")
	}
	msg := err.Error()
	for _, leaked := range []string{"abcdef", "abc", "redacteddef"} {
		if strings.Contains(msg, leaked) {
			t.Fatalf("write error leaked %q in %q", leaked, msg)
		}
	}
}

func TestWriteRetainsOrderedProviderResponsesBeforeTrailingJSONFailure(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			_, _ = w.Write([]byte(`{"provider_id":"first"}`))
			return
		}
		_, _ = w.Write([]byte(`{"provider_id":"second"} trailing`))
	}))
	t.Cleanup(srv.Close)
	bundle := newWriteTestBundle(srv, WriteAction{
		Kind:       "update",
		Method:     http.MethodPost,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
		Confirm:    "destructive",
	})
	records := []connectors.Record{{"id": "first"}, {"id": "second"}}
	result, err := Write(context.Background(), bundle, approvedWriteRequest(t, bundle, "update_widget", records, nil), records, nil)
	if err == nil {
		t.Fatal("Write() error = nil, want trailing JSON failure")
	}
	if calls != 2 || result.RecordsWritten != 1 || result.RecordsFailed != 1 || len(result.ProviderResponses) != 2 {
		t.Fatalf("Write() result = %#v calls=%d, want first success and two correlated responses", result, calls)
	}
	if first := result.ProviderResponses[0]; first.RecordIndex != 0 || first.Body.(map[string]any)["provider_id"] != "first" {
		t.Fatalf("first provider response = %#v, want record zero success", first)
	}
	if second := result.ProviderResponses[1]; second.RecordIndex != 1 || second.BodyEncoding != "text" || second.Body != `{"provider_id":"second"} trailing` {
		t.Fatalf("second provider response = %#v, want complete raw trailing response", second)
	}
}

func TestWriteRetainsRawDeclaredJSONAlongsideParsedBody(t *testing.T) {
	const providerBody = "{\n  \"provider_id\": \"first\",\n  \"duplicate\": \"first\",\n  \"duplicate\": \"second\"\n}\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(providerBody))
	}))
	t.Cleanup(srv.Close)
	bundle := newWriteTestBundle(srv, WriteAction{Kind: "update", Method: http.MethodPost, Path: "/widgets/{{ record.id }}", PathFields: []string{"id"}})
	result, err := Write(context.Background(), bundle, connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{{"id": "widget-1"}}, nil)
	if err != nil {
		t.Fatalf("Write(): %v", err)
	}
	if result.RecordsWritten != 1 || len(result.ProviderResponses) != 1 {
		t.Fatalf("Write() result = %#v, want one written record with one provider response", result)
	}
	provider := result.ProviderResponses[0]
	if !provider.BodyPresent || provider.BodyBytes != len(providerBody) || provider.BodyRaw != providerBody || provider.BodyRawEncoding != "text" {
		t.Fatalf("provider response = %#v, want exact raw JSON body", provider)
	}
	body, ok := provider.Body.(map[string]any)
	if !ok || body["provider_id"] != "first" || body["duplicate"] != "second" {
		t.Fatalf("parsed provider body = %#v, want parsed JSON view alongside raw body", provider.Body)
	}
}

func TestWriteProviderResponseRejectsBodiesBeyondCaptureLimit(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		header http.Header
		body   []byte
	}{
		{name: "declared JSON", header: http.Header{"Content-Type": []string{"application/json"}, "X-Provider-Receipt": []string{"receipt-one", "receipt-two"}}, body: []byte(`{"id":1}`)},
		{name: "explicit text", header: http.Header{"Content-Type": []string{"text/plain"}, "X-Provider-Receipt": []string{"receipt-one", "receipt-two"}}, body: []byte("abcde")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			response := &connsdk.Response{Status: http.StatusOK, Header: testCase.header, Body: testCase.body}
			result, err := writeProviderResponseWithLimit(response, 3, 4)
			if err == nil || !strings.Contains(err.Error(), "too large") {
				t.Fatalf("writeProviderResponseWithLimit() error = %v, want capture-limit failure", err)
			}
			if result.RecordIndex != 3 || result.Status != http.StatusOK || !result.BodyPresent || result.BodyBytes != len(testCase.body) || result.BodyRaw != string(testCase.body) || result.BodyRawEncoding != "text" || result.Body != string(testCase.body) || result.BodyEncoding != "text" {
				t.Fatalf("provider response = %#v, want bounded raw response facts", result)
			}
			if !reflect.DeepEqual(result.Headers["Content-Type"].Values, testCase.header.Values("Content-Type")) || !reflect.DeepEqual(result.Headers["X-Provider-Receipt"].Values, []string{"receipt-one", "receipt-two"}) {
				t.Fatalf("provider response headers = %#v, want preserved headers", result.Headers)
			}
		})
	}
}

func TestWriteRequiresOneJSONValueForDeclaredResponses(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		status       int
		contentType  string
		response     []byte
		wantErr      bool
		wantBody     any
		wantEncoding string
	}{
		{
			name:         "empty declared JSON",
			status:       http.StatusOK,
			contentType:  "application/json",
			wantBody:     nil,
			wantEncoding: "",
		},
		{
			name:         "whitespace declared JSON",
			status:       http.StatusOK,
			contentType:  "application/json",
			response:     []byte(" \n\t "),
			wantErr:      true,
			wantBody:     " \n\t ",
			wantEncoding: "text",
		},
		{
			name:         "one JSON value with trailing whitespace",
			status:       http.StatusOK,
			contentType:  "application/json; charset=utf-8",
			response:     []byte("{\"provider_id\":\"one\"}\n\t"),
			wantBody:     map[string]any{"provider_id": "one"},
			wantEncoding: "",
		},
		{
			name:         "multiple declared JSON values",
			status:       http.StatusOK,
			contentType:  "application/json",
			response:     []byte("{\"provider_id\":\"one\"} {\"provider_id\":\"two\"}"),
			wantErr:      true,
			wantBody:     "{\"provider_id\":\"one\"} {\"provider_id\":\"two\"}",
			wantEncoding: "text",
		},
		{
			name:         "non JSON whitespace text",
			status:       http.StatusOK,
			contentType:  "text/plain",
			response:     []byte(" \n\t "),
			wantBody:     " \n\t ",
			wantEncoding: "text",
		},
		{
			name:         "bodyless status",
			status:       http.StatusNoContent,
			wantBody:     nil,
			wantEncoding: "",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("X-Provider-Receipt", "receipt-one")
				w.Header().Add("X-Provider-Receipt", "receipt-two")
				if testCase.contentType != "" {
					w.Header().Set("Content-Type", testCase.contentType)
				}
				w.WriteHeader(testCase.status)
				_, _ = w.Write(testCase.response)
			}))
			t.Cleanup(srv.Close)
			bundle := newWriteTestBundle(srv, WriteAction{Kind: "update", Method: http.MethodPost, Path: "/widgets/{{ record.id }}", PathFields: []string{"id"}})

			result, err := Write(context.Background(), bundle, connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{{"id": "widget-1"}}, nil)
			if testCase.wantErr {
				if err == nil {
					t.Fatal("Write() error = nil, want declared JSON response failure")
				}
				if !strings.Contains(err.Error(), "provider response") {
					t.Fatalf("Write() error = %v, want provider response failure", err)
				}
			} else if err != nil {
				t.Fatalf("Write(): %v", err)
			}

			if len(result.ProviderResponses) != 1 {
				t.Fatalf("provider responses = %#v, want one captured response", result.ProviderResponses)
			}
			provider := result.ProviderResponses[0]
			if provider.RecordIndex != 0 || provider.Status != testCase.status || !reflect.DeepEqual(provider.Headers["X-Provider-Receipt"].Values, []string{"receipt-one", "receipt-two"}) {
				t.Fatalf("provider response = %#v, want captured status and receipts", provider)
			}
			if provider.BodyEncoding != testCase.wantEncoding || !reflect.DeepEqual(provider.Body, testCase.wantBody) {
				t.Fatalf("provider response body = %#v encoding=%q, want %#v encoding=%q", provider.Body, provider.BodyEncoding, testCase.wantBody, testCase.wantEncoding)
			}
			if provider.BodyPresent != (len(testCase.response) > 0) {
				t.Fatalf("provider BodyPresent = %v, want transport-byte presence %v", provider.BodyPresent, len(testCase.response) > 0)
			}
			if testCase.wantErr {
				if result.RecordsWritten != 0 || result.RecordsFailed != 1 {
					t.Fatalf("failed write result = %#v, want one failed record", result)
				}
			} else if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
				t.Fatalf("successful write result = %#v, want one written record", result)
			}
		})
	}
}

func TestWriteRetainsTerminalProviderResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("X-Provider-Receipt", "receipt-one")
		w.Header().Add("X-Provider-Receipt", "receipt-two")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"provider_echo":"client-secret","provider_id":9007199254740993}`))
	}))
	t.Cleanup(srv.Close)
	bundle := newWriteTestBundle(srv, WriteAction{
		Kind:         "update",
		Method:       http.MethodPost,
		Path:         "/widgets/{{ record.id }}",
		PathFields:   []string{"id"},
		RedactFields: []string{"token"},
		Confirm:      "destructive",
		RecordSchema: json.RawMessage(`{"type":"object","required":["id","token"],"properties":{"id":{"type":"string"},"token":{"type":"string"}}}`),
	})
	records := []connectors.Record{{"id": "widget-1", "token": "client-secret"}}
	result, err := Write(context.Background(), bundle, approvedWriteRequest(t, bundle, "update_widget", records, nil), records, nil)
	if err == nil {
		t.Fatal("Write() error = nil, want terminal provider failure")
	}
	if strings.Contains(err.Error(), "client-secret") {
		t.Fatal("Write() error leaked a provider response")
	}
	if result.RecordsWritten != 0 || result.RecordsFailed != 1 || len(result.ProviderResponses) != 1 {
		t.Fatal("terminal write result did not retain one failed response")
	}
	provider := result.ProviderResponses[0]
	if provider.RecordIndex != 0 || provider.Status != http.StatusBadRequest || !reflect.DeepEqual(provider.Headers["X-Provider-Receipt"].Values, []string{"receipt-one", "receipt-two"}) {
		t.Fatal("terminal provider response did not retain status and receipt")
	}
	body, ok := provider.Body.(map[string]any)
	if !ok || body["provider_echo"] != "client-secret" || body["provider_id"] != json.Number("9007199254740993") {
		t.Fatal("terminal provider body did not retain exact provider facts")
	}
}

func TestWritePreservesExplicitNonJSONProviderResponses(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		contentType string
		response    []byte
		body        any
		encoding    string
	}{
		{name: "starts with t", contentType: "text/plain", response: []byte("thanks"), body: "thanks", encoding: "text"},
		{name: "starts with f", contentType: "text/plain", response: []byte("false"), body: "false", encoding: "text"},
		{name: "starts with n", contentType: "text/plain", response: []byte("null"), body: "null", encoding: "text"},
		{name: "starts with digit", contentType: "text/plain", response: []byte("123"), body: "123", encoding: "text"},
		{name: "starts with brace", contentType: "text/plain", response: []byte(`{"provider":"text"}`), body: `{"provider":"text"}`, encoding: "text"},
		{name: "starts with bracket", contentType: "text/plain", response: []byte(`["provider","text"]`), body: `["provider","text"]`, encoding: "text"},
		{name: "json-looking media type", contentType: "application/jsonish", response: []byte(`{"provider":"text"}`), body: `{"provider":"text"}`, encoding: "text"},
		{name: "binary", contentType: "application/octet-stream", response: []byte{0xff, 0x00, 0x80}, body: base64.StdEncoding.EncodeToString([]byte{0xff, 0x00, 0x80}), encoding: "base64"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", testCase.contentType)
				_, _ = w.Write(testCase.response)
			}))
			t.Cleanup(srv.Close)
			bundle := newWriteTestBundle(srv, WriteAction{Kind: "update", Method: http.MethodPost, Path: "/widgets/{{ record.id }}", PathFields: []string{"id"}})
			result, err := Write(context.Background(), bundle, connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{{"id": "widget-1"}}, nil)
			if err != nil {
				t.Fatalf("Write(): %v", err)
			}
			if result.RecordsWritten != 1 || len(result.ProviderResponses) != 1 || result.ProviderResponses[0].BodyEncoding != testCase.encoding || !reflect.DeepEqual(result.ProviderResponses[0].Body, testCase.body) {
				t.Fatal("non-JSON provider result did not retain raw response bytes")
			}
		})
	}
}

// --- accounting parity with legacy semantics (stripe/write.go:66) ---

func TestWriteAccountingFailFastRemainderCountsAsFailed(t *testing.T) {
	// A non-retryable client error (400) on the second record so the test
	// does not incur real connsdk retry/backoff sleeps.
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "update",
		Method:     http.MethodPost,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
	})

	records := []connectors.Record{
		{"id": "1"}, {"id": "2"}, {"id": "3"}, {"id": "4"},
	}
	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_widget"}, records, nil)
	if err == nil {
		t.Fatalf("Write: want error on second record's 400")
	}
	// Matches stripe/write.go:66 fail-fast semantics: RecordsWritten counts
	// successes up to the failure, RecordsFailed = len(records) - RecordsWritten.
	if result.RecordsWritten != 1 {
		t.Fatalf("RecordsWritten = %d, want 1", result.RecordsWritten)
	}
	if result.RecordsFailed != len(records)-result.RecordsWritten {
		t.Fatalf("RecordsFailed = %d, want %d", result.RecordsFailed, len(records)-result.RecordsWritten)
	}
}

func TestWriteValidationFailureReportsAllRecordsFailed(t *testing.T) {
	b := Bundle{
		Name: "acme",
		Writes: []WriteAction{{
			Name:   "create_issue",
			Kind:   "create",
			Method: http.MethodPost,
			Path:   "/issues",
			RecordSchema: json.RawMessage(`{
				"type": "object", "required": ["title"],
				"properties": {"title": {"type": "string"}}
			}`),
		}},
	}
	records := []connectors.Record{{"title": "ok"}, {"body": "no title"}}

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "create_issue"}, records, nil)
	if err == nil {
		t.Fatalf("Write: want validation error")
	}
	if result.RecordsFailed != len(records) {
		t.Fatalf("RecordsFailed = %d, want %d (validation fails before any network call)", result.RecordsFailed, len(records))
	}
}

// --- ctx cancellation mid-loop ---

func TestWriteCtxCancelMidLoopAccounting(t *testing.T) {
	// A post-receipt validator cancels after the first provider response has
	// been persisted and counted, so execution observes cancellation at the
	// next sealed physical-record boundary. This replaces the old WriteHook
	// timing seam: legacy execution hooks are refused because they could choose
	// an unpreviewed physical request after approval.
	ctx, cancel := context.WithCancel(context.Background())
	var lastPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "update",
		Method:     http.MethodPost,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
	})

	records := []connectors.Record{{"id": "1"}, {"id": "2"}, {"id": "3"}}
	h := &preparedWriteResponseValidatorFunc{fn: func(_ WriteAction, rec connectors.Record, _ *connsdk.Response) error {
		if rec["id"] == "1" {
			cancel()
		}
		return nil
	}}
	result, err := Write(ctx, b, connectors.WriteRequest{Action: "update_widget"}, records, h)
	if err == nil {
		t.Fatalf("Write: want context.Canceled surfaced")
	}
	if result.RecordsWritten != 1 {
		t.Fatalf("RecordsWritten = %d, want 1 (record 1 completes; ctx is cancelled before record 2's own request is attempted)", result.RecordsWritten)
	}
	if result.RecordsFailed != len(records)-result.RecordsWritten {
		t.Fatalf("RecordsFailed = %d, want %d", result.RecordsFailed, len(records)-result.RecordsWritten)
	}
	if lastPath != "/widgets/1" {
		t.Fatalf("last observed request path = %q, want /widgets/1 (only record 1 ever reached the server)", lastPath)
	}
}

// --- WriteHook ---

func TestLegacyWriteHookClaimIsRefusedBeforeUnpreviewedTransport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("declarative HTTP request should not happen when WriteHook handles the write")
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:   "merge_pull_request",
		Kind:   "custom",
		Method: http.MethodPut,
		Path:   "/pulls/{{ record.pull_number }}/merge",
	})

	calls := 0
	h := &writeHookFunc{claims: true, fn: func(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (bool, []*connsdk.Response, error) {
		calls++
		return true, nil, nil
	}}

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "merge_pull_request"}, []connectors.Record{
		{"pull_number": 7},
	}, h)
	if err == nil || !strings.Contains(err.Error(), "without an exact prepared-request plan") {
		t.Fatalf("Write error = %v, want pre-I/O legacy-hook refusal", err)
	}
	if calls != 0 || result.RecordsWritten != 0 || result.RecordsFailed != 1 {
		t.Fatalf("legacy hook/result = %d / %+v, want no hook call and one refused record", calls, result)
	}
}

// TestPreparedWriteHookSealsEveryPhysicalRequestAndRetainsTerminalReceipts
// proves the compound-write approval boundary: the hook selects only named
// declaration-owned actions, preparation captures both physical requests in
// their execution order, and the response-bound follow-up receives its path
// value only from the sealed first receipt. A caller mutation after planning
// cannot reach either wire payload, and the failed terminal receipt remains in
// the result beside the successful creation receipt.
func TestPreparedWriteHookSealsEveryPhysicalRequestAndRetainsTerminalReceipts(t *testing.T) {
	var paths []string
	var bodies []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode %s body: %v", r.URL.Path, err)
		}
		paths = append(paths, r.URL.Path)
		bodies = append(bodies, body)
		switch r.URL.Path {
		case "/widgets":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":73,"provider":"created"}`))
		case "/widgets/73":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"terminal update failure"}`))
		default:
			t.Fatalf("unexpected physical request %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)
	b := Bundle{
		Name: "sealed-compound", HTTP: HTTPBase{URL: srv.URL},
		Writes: []WriteAction{
			{Name: "compound_widget", Kind: "custom", Method: http.MethodPost, Path: "/compound", Hook: "sealed"},
			{Name: "create_widget", Kind: "create", Method: http.MethodPost, Path: "/widgets", BodyFields: []string{"payload"}},
			{Name: "update_widget", Kind: "update", Method: http.MethodPatch, Path: "/widgets/{{ record.id }}", PathFields: []string{"id"}, BodyFields: []string{"payload"}},
		},
	}
	hook := &sealedCompoundWriteHook{}
	records := []connectors.Record{{"payload": map[string]any{"name": "approved"}}}
	prepared, err := prepareDeclarativeWrite(context.Background(), b, connectors.WriteRequest{Action: "compound_widget"}, records, hook)
	if err != nil {
		t.Fatalf("prepareDeclarativeWrite: %v", err)
	}
	if len(prepared.Requests) != 2 {
		t.Fatalf("prepared requests = %#v, want both physical requests", prepared.Requests)
	}
	if got := []string{prepared.Requests[0].Action, prepared.Requests[1].Action}; !reflect.DeepEqual(got, []string{"create_widget", "update_widget"}) {
		t.Fatalf("prepared action order = %v, want create_widget then update_widget", got)
	}
	if binding := prepared.Requests[1].ResponseBinding; binding == nil || binding.SourceStep != 0 || binding.Field != "id" || binding.TargetField != "id" {
		t.Fatalf("prepared follow-up binding = %#v, want sealed first-response id binding", binding)
	}
	preview, err := PreviewPreparedWrite(prepared)
	if err != nil {
		t.Fatalf("PreviewPreparedWrite: %v", err)
	}
	if preview.Digest == "" {
		t.Fatal("preview digest is empty; physical plan was not sealed")
	}

	action, err := findWriteAction(b, "compound_widget")
	if err != nil {
		t.Fatalf("findWriteAction: %v", err)
	}
	var result connectors.WriteResult
	err = ExecutePreparedWrite(context.Background(), prepared, nil, preview.Digest, func(ctx context.Context) error {
		result, err = executeApprovedWrite(ctx, b, action, connectors.WriteRequest{Action: "compound_widget"}, records, prepared, preview.Digest, hook)
		return err
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP status 400") {
		t.Fatalf("ExecutePreparedWrite error = %v, want terminal compound failure", err)
	}
	if got := records[0]["payload"].(map[string]any)["name"]; got != "mutated-after-preview" {
		t.Fatalf("fixture did not mutate caller record: %#v", records)
	}
	if !reflect.DeepEqual(paths, []string{"/widgets", "/widgets/73"}) {
		t.Fatalf("physical request paths = %v, want sealed ordered create then response-bound update", paths)
	}
	for index, body := range bodies {
		if body["payload"].(map[string]any)["name"] != "approved" {
			t.Fatalf("physical request %d body = %#v, want sealed approved payload", index, body)
		}
	}
	if result.RecordsWritten != 0 || result.RecordsFailed != 1 || len(result.ProviderResponses) != 2 {
		t.Fatalf("Write result = %#v, want failed record with both compound receipts", result)
	}
	if result.ProviderResponses[0].RecordIndex != 0 || result.ProviderResponses[0].Status != http.StatusCreated || result.ProviderResponses[1].RecordIndex != 0 || result.ProviderResponses[1].Status != http.StatusBadRequest {
		t.Fatalf("compound receipts = %#v, want ordered statuses 201 then 400 for record zero", result.ProviderResponses)
	}
	if got := result.ProviderResponses[1].Body.(map[string]any)["message"]; got != "terminal update failure" {
		t.Fatalf("terminal provider receipt = %#v, want exact provider failure", result.ProviderResponses[1])
	}
}

func TestWriteHookNotHandledFallsBackToDeclarative(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, "")
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "update",
		Method:     http.MethodPost,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
	})

	h := &writeHookFunc{claims: false, fn: func(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (bool, []*connsdk.Response, error) {
		return false, nil, nil
	}}

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{
		{"id": "42", "name": "x"},
	}, h)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if result.RecordsWritten != 1 {
		t.Fatalf("result = %+v, want 1", result)
	}
	if cap.path != "/widgets/42" {
		t.Fatalf("path = %q, want declarative fallback to run the real request", cap.path)
	}
}

// --- form body with non-string field values ---

func TestWriteFormBodyStringifiesNonStringValues(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, "")
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "update_widget_form",
		Kind:       "update",
		Method:     http.MethodPost,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
		BodyType:   "form",
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_widget_form"}, []connectors.Record{
		{"id": "42", "quantity": 3, "active": true},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	form := cap.form()
	if form.Get("quantity") != "3" {
		t.Fatalf("form.quantity = %q, want 3", form.Get("quantity"))
	}
	if form.Get("active") != "true" {
		t.Fatalf("form.active = %q, want true", form.Get("active"))
	}
}

// --- unknown action ---

func TestWriteUnknownActionErrors(t *testing.T) {
	b := Bundle{Name: "acme", Writes: []WriteAction{{Name: "known_action", Method: http.MethodPost, Path: "/x"}}}
	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "does-not-exist"}, []connectors.Record{{}}, nil)
	if err == nil {
		t.Fatalf("Write: want error for unknown action")
	}
}

func TestWriteJSONArrayBodySendsTopLevelArray(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{"ok":true}`)
	b := newWriteTestBundle(srv, WriteAction{
		Name:      "upload_schema",
		Kind:      "create",
		Method:    http.MethodPost,
		Path:      "/entity-schema",
		BodyType:  "json_array",
		BodyField: "selected_fields",
		BodySchema: json.RawMessage(`{
			"type": "array",
			"items": {"type": "object", "required": ["name"], "properties": {"name": {"type": "string"}}, "additionalProperties": false}
		}`),
		RecordSchema: json.RawMessage(`{
			"type": "object",
			"required": ["selected_fields"],
			"properties": {"selected_fields": {"type": "array", "items": {"type": "object"}}}
		}`),
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "upload_schema"}, []connectors.Record{
		{"selected_fields": []any{map[string]any{"name": "Account"}}},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if cap.contentType != "application/json" {
		t.Fatalf("content-type = %q, want application/json", cap.contentType)
	}
	var got []map[string]any
	if err := json.Unmarshal(cap.body, &got); err != nil {
		t.Fatalf("body is not a top-level array: %s: %v", string(cap.body), err)
	}
	if len(got) != 1 || got[0]["name"] != "Account" {
		t.Fatalf("body = %+v, want selected_fields root array", got)
	}
}

func TestWriteJSONArrayBodySchemaMismatchFailsBeforeNetwork(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits++ }))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:         "upload_schema",
		Kind:         "create",
		Method:       http.MethodPost,
		Path:         "/entity-schema",
		BodyType:     "json_array",
		BodyField:    "selected_fields",
		BodySchema:   json.RawMessage(`{"type":"array","items":{"type":"object","required":["name"],"properties":{"name":{"type":"string"}},"additionalProperties":false}}`),
		RecordSchema: json.RawMessage(`{"type":"object","required":["selected_fields"],"properties":{"selected_fields":{"type":"array","items":{"type":"object"}}}}`),
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "upload_schema"}, []connectors.Record{
		{"selected_fields": []any{map[string]any{"extra": "not allowed"}}},
	}, nil)
	if err == nil {
		t.Fatalf("Write: want schema validation error")
	}
	if hits != 0 {
		t.Fatalf("server hits = %d, want 0 before schema-valid body", hits)
	}
}

func TestWriteMultipartBodySendsDeclaredParts(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/media.txt"
	if err := os.WriteFile(filePath, []byte("hello media"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	var sawField, sawFile string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Fatalf("Content-Type = %q, want multipart boundary", r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		sawField = r.MultipartForm.Value["source"][0]
		fh := r.MultipartForm.File["mediaFile"][0]
		f, err := fh.Open()
		if err != nil {
			t.Fatalf("Open part: %v", err)
		}
		defer func() { _ = f.Close() }()
		raw, _ := io.ReadAll(f)
		sawFile = string(raw)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	b := newWriteTestBundle(srv, WriteAction{
		Name:       "upload_media",
		Kind:       "update",
		Method:     http.MethodPut,
		Path:       "/calls/{{ record.id }}/media",
		BodyType:   "multipart",
		PathFields: []string{"id"},
		Multipart: &MultipartSpec{MaxBytes: 1024, Parts: []MultipartPartSpec{
			{Name: "source", Type: "field", Field: "source", Required: true},
			{Name: "mediaFile", Type: "file", Field: "media_file_path", ContentType: "text/plain", Required: true, MaxBytes: 1024},
		}},
		RecordSchema: json.RawMessage(`{"type":"object","required":["id","source","media_file_path"],"properties":{"id":{"type":"string"},"source":{"type":"string"},"media_file_path":{"type":"string"}}}`),
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "upload_media", Config: connectors.RuntimeConfig{ProjectDir: dir}}, []connectors.Record{
		{"id": "call-1", "source": "recorder", "media_file_path": "media.txt"},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if sawField != "recorder" || sawFile != "hello media" {
		t.Fatalf("multipart parts source=%q file=%q", sawField, sawFile)
	}
}

func TestWriteMultipartRejectsContentThatDoesNotMatchApproval(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/media.txt", []byte("evil"), 0o600); err != nil {
		t.Fatal(err)
	}
	approved := sha256.Sum256([]byte("safe"))
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	bundle := newWriteTestBundle(server, WriteAction{
		Name:     "upload_media",
		Method:   http.MethodPost,
		Path:     "/upload",
		BodyType: "multipart",
		Multipart: &MultipartSpec{MaxBytes: 4, Parts: []MultipartPartSpec{
			{Name: "mediaFile", Type: "file", Field: "media_file_path", Required: true, MaxBytes: 4},
		}},
		RecordSchema: json.RawMessage(`{"type":"object","required":["media_file_path"],"properties":{"media_file_path":{"type":"string"}}}`),
	})
	_, err := Write(context.Background(), bundle, connectors.WriteRequest{
		Action: "upload_media",
		Config: connectors.RuntimeConfig{
			ProjectDir: dir,
			ApprovedPayloadSHA256: map[string]string{
				connectors.PayloadApprovalKey(0, "media_file_path"): hex.EncodeToString(approved[:]),
			},
		},
	}, []connectors.Record{{"media_file_path": "media.txt"}}, nil)
	if err == nil || !strings.Contains(err.Error(), "changed since approval") {
		t.Fatalf("Write error = %v, want approval mismatch", err)
	}
	if hits != 0 {
		t.Fatalf("server hits = %d, want zero", hits)
	}
}

// --- test-only hook adapter ---

type writeHookFunc struct {
	claims bool
	fn     func(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (bool, []*connsdk.Response, error)
}

func (w *writeHookFunc) ConnectorName() string { return "write-hook-func-test" }
func (w *writeHookFunc) ExecuteWrite(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (bool, []*connsdk.Response, error) {
	return w.fn(ctx, action, rec, rt)
}

func (w *writeHookFunc) HandlesWriteAction(WriteAction) bool { return w.claims }

type preparedWriteResponseValidatorFunc struct {
	fn func(WriteAction, connectors.Record, *connsdk.Response) error
}

func (*preparedWriteResponseValidatorFunc) ConnectorName() string {
	return "prepared-response-validator-test"
}
func (h *preparedWriteResponseValidatorFunc) ValidatePreparedWriteResponse(action WriteAction, rec connectors.Record, response *connsdk.Response) error {
	return h.fn(action, rec, response)
}

type sealedCompoundWriteHook struct{}

func (*sealedCompoundWriteHook) ConnectorName() string { return "sealed-compound-write-test" }

func (h *sealedCompoundWriteHook) MapWriteRecord(_ WriteAction, rec connectors.Record) (connectors.Record, bool, error) {
	pinned := connectors.Record(copyRecordMap(map[string]any(rec)))
	rec["payload"].(map[string]any)["name"] = "mutated-after-preview"
	return pinned, true, nil
}

func (*sealedCompoundWriteHook) PrepareWrite(action WriteAction, records []connectors.Record) (PreparedWriteHookPlan, bool, error) {
	if action.Name != "compound_widget" {
		return PreparedWriteHookPlan{}, false, nil
	}
	if len(records) != 1 {
		return PreparedWriteHookPlan{}, true, errors.New("test hook requires one record")
	}
	payload := records[0]["payload"]
	return PreparedWriteHookPlan{Records: []PreparedWriteHookRecord{{Steps: []PreparedWriteHookStep{
		{Action: "create_widget", Record: connectors.Record{"payload": payload}},
		// id is a schema/interpolation witness only. The engine replaces it
		// from the sealed create receipt before it can reach the wire.
		{Action: "update_widget", Record: connectors.Record{"id": 0, "payload": payload}, ResponseBinding: &PreparedWriteResponseBinding{SourceStep: 0, Field: "id", TargetField: "id"}},
	}}}}, true, nil
}

// --- array cardinality reach ----------------------------------------------
//
// One keyword pair added to the shared dialect must reach every request-building
// path with no per-site change. That property is what turns a single rule into
// 25 unblocked Airtable operations, so it is asserted directly rather than
// assumed: record_schema here, json_array body_schema below, and operation
// rest.body_schema in direct_read_test.go.

func TestWriteRecordSchemaMinItemsRejectsEmptyArray(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{}`)
	b := newWriteTestBundle(srv, WriteAction{
		Name:   "delete_records",
		Kind:   "delete",
		Method: http.MethodDelete,
		Path:   "/v0/base/table",
		Risk:   "critical",
		RecordSchema: json.RawMessage(`{
			"type": "object",
			"required": ["records"],
			"additionalProperties": false,
			"properties": {
				"records": {"type": "array", "minItems": 1, "items": {"type": "string"}}
			}
		}`),
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "delete_records"},
		[]connectors.Record{{"records": []any{}}}, nil)
	if err == nil {
		t.Fatal("empty documented array: want validation error, got nil")
	}
	if !strings.Contains(err.Error(), "minItems") {
		t.Fatalf("error should name minItems, got %v", err)
	}
	if cap.method != "" {
		t.Fatalf("request must not be issued for an invalid record, got %s %s", cap.method, cap.path)
	}

	validRecords := []connectors.Record{{"records": []any{"rec1"}}}
	if _, err := Write(context.Background(), b, approvedWriteRequest(t, b, "delete_records", validRecords, nil), validRecords, nil); err != nil {
		t.Fatalf("non-empty array: unexpected error: %v", err)
	}
}

func TestWriteJSONArrayBodySchemaMinItems(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{}`)
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "upsert_records",
		Kind:       "upsert",
		Method:     http.MethodPut,
		Path:       "/v0/acct/table/upsertRecords",
		Risk:       "critical",
		BodyType:   "json_array",
		BodyField:  "records",
		BodySchema: json.RawMessage(`{"type":"array","minItems":1,"items":{"type":"object"}}`),
		RecordSchema: json.RawMessage(`{
			"type": "object",
			"required": ["records"],
			"properties": {"records": {"type": "array"}}
		}`),
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "upsert_records"},
		[]connectors.Record{{"records": []any{}}}, nil)
	if err == nil {
		t.Fatal("empty json_array body: want validation error, got nil")
	}
	if !strings.Contains(err.Error(), "minItems") {
		t.Fatalf("error should name minItems, got %v", err)
	}
	if cap.method != "" {
		t.Fatalf("request must not be issued, got %s %s", cap.method, cap.path)
	}
}

// --- bounded base64 upload -------------------------------------------------
//
// Airtable's ledger blocks POST /v0/{baseId}/{recordId}/{field}/uploadAttachment
// "until an Airtable-owned executor validates official base64 encoding and
// decoded-size bounds before transmission". The only alternative would be a raw
// body escape hatch, which is banned outright.

func base64UploadAction(spec *Base64UploadSpec) WriteAction {
	return WriteAction{
		Name:         "upload_attachment",
		Kind:         "create",
		Method:       http.MethodPost,
		Path:         "/v0/base/rec/fld/uploadAttachment",
		Risk:         "medium",
		BodyType:     "base64_upload",
		Base64Upload: spec,
		RecordSchema: json.RawMessage(`{
			"type": "object",
			"required": ["file_path", "filename", "contentType"],
			"additionalProperties": false,
			"properties": {
				"file_path": {"type": "string"},
				"filename": {"type": "string"},
				"contentType": {"type": "string"}
			}
		}`),
	}
}

func writeTempPayload(t *testing.T, name string, content []byte) (projectDir, rel string) {
	t.Helper()
	dir := t.TempDir()
	rel = name
	if err := os.WriteFile(filepath.Join(dir, name), content, 0o600); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	return dir, rel
}

func TestWriteBase64UploadEncodesFileAndOmitsSourceField(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{"id":"att1"}`)
	dir, rel := writeTempPayload(t, "payload.txt", []byte("hello attachment"))

	b := newWriteTestBundle(srv, base64UploadAction(&Base64UploadSpec{
		SourceField:     "file_path",
		ContentField:    "file",
		MaxDecodedBytes: 1024,
	}))

	if _, err := Write(context.Background(), b, connectors.WriteRequest{
		Action: "upload_attachment",
		Config: connectors.RuntimeConfig{ProjectDir: dir},
	}, []connectors.Record{{"file_path": rel, "filename": "payload.txt", "contentType": "text/plain"}}, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(cap.body, &body); err != nil {
		t.Fatalf("decode body: %v (raw %s)", err, cap.body)
	}
	want := base64.StdEncoding.EncodeToString([]byte("hello attachment"))
	if got, _ := body["file"].(string); got != want {
		t.Fatalf("body[file] = %q, want %q", got, want)
	}
	// The source field carries a LOCAL FILESYSTEM PATH. Transmitting it would
	// leak the operator's directory layout to the provider.
	if _, present := body["file_path"]; present {
		t.Fatalf("source field must never reach the wire, body = %v", body)
	}
	// Ordinary declared fields still travel, governed by record_schema.
	if got, _ := body["filename"].(string); got != "payload.txt" {
		t.Fatalf("body[filename] = %q, want payload.txt", got)
	}
	if got, _ := body["contentType"].(string); got != "text/plain" {
		t.Fatalf("body[contentType] = %q, want text/plain", got)
	}
}

func TestWriteBase64UploadRejectsOversizePayload(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{}`)
	dir, rel := writeTempPayload(t, "big.bin", bytes.Repeat([]byte("x"), 64))

	b := newWriteTestBundle(srv, base64UploadAction(&Base64UploadSpec{
		SourceField:     "file_path",
		ContentField:    "file",
		MaxDecodedBytes: 32,
	}))

	_, err := Write(context.Background(), b, connectors.WriteRequest{
		Action: "upload_attachment",
		Config: connectors.RuntimeConfig{ProjectDir: dir},
	}, []connectors.Record{{"file_path": rel, "filename": "big.bin", "contentType": "application/octet-stream"}}, nil)
	if err == nil {
		t.Fatal("oversize payload: want error, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Fatalf("error should name the size bound, got %v", err)
	}
	// Rejected, never truncated: a truncated attachment is a silently corrupt
	// upload, which is worse than a failed one.
	if cap.method != "" {
		t.Fatalf("oversize payload must not be transmitted, got %s %s", cap.method, cap.path)
	}
}

func TestWriteBase64UploadRejectsPathEscape(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{}`)
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	// A symlink inside the project pointing out of it is the case a purely
	// lexical path check cannot catch.
	if err := os.Symlink(outside, filepath.Join(dir, "link.txt")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}

	b := newWriteTestBundle(srv, base64UploadAction(&Base64UploadSpec{
		SourceField:     "file_path",
		ContentField:    "file",
		MaxDecodedBytes: 1024,
	}))

	for _, raw := range []string{"../escape.txt", "link.txt"} {
		cap.method = ""
		_, err := Write(context.Background(), b, connectors.WriteRequest{
			Action: "upload_attachment",
			Config: connectors.RuntimeConfig{ProjectDir: dir},
		}, []connectors.Record{{"file_path": raw, "filename": "x", "contentType": "text/plain"}}, nil)
		if err == nil {
			t.Fatalf("%q: want containment error, got nil", raw)
		}
		if cap.method != "" {
			t.Fatalf("%q: must not be transmitted", raw)
		}
	}
}

func TestWriteBase64UploadStrictSourceRejectsSloppyEncoding(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{}`)
	action := base64UploadAction(&Base64UploadSpec{
		Source:          "base64",
		SourceField:     "content",
		ContentField:    "file",
		MaxDecodedBytes: 1024,
	})
	action.RecordSchema = json.RawMessage(`{
		"type": "object",
		"required": ["content"],
		"additionalProperties": false,
		"properties": {"content": {"type": "string"}}
	}`)
	b := newWriteTestBundle(srv, action)

	// "Official base64" means RFC 4648 standard alphabet, canonical padding, no
	// line breaks — exactly what StdEncoding.Strict() enforces.
	for _, bad := range []string{
		"aGVsbG8",             // missing padding
		"aGVs\nbG8=",          // embedded newline
		"aGVsbG8h_w==",        // URL-safe alphabet
		"not base64 at all!!", // not base64
	} {
		cap.method = ""
		_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "upload_attachment"},
			[]connectors.Record{{"content": bad}}, nil)
		if err == nil {
			t.Fatalf("%q: want strict-base64 rejection, got nil", bad)
		}
		if cap.method != "" {
			t.Fatalf("%q: must not be transmitted", bad)
		}
	}

	good := base64.StdEncoding.EncodeToString([]byte("hello"))
	if _, err := Write(context.Background(), b, connectors.WriteRequest{Action: "upload_attachment"},
		[]connectors.Record{{"content": good}}, nil); err != nil {
		t.Fatalf("canonical base64: unexpected error: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(cap.body, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got, _ := body["file"].(string); got != good {
		t.Fatalf("body[file] = %q, want %q", got, good)
	}
}

func TestWriteBase64UploadEnforcesEncodedBound(t *testing.T) {
	srv, _ := captureServer(t, http.StatusOK, `{}`)
	dir, rel := writeTempPayload(t, "p.bin", bytes.Repeat([]byte("y"), 40))

	// 40 decoded bytes encode to 56 base64 characters. Real APIs (Airtable's
	// 5 MB attachment cap among them) document the ENCODED limit, so both are
	// checked.
	b := newWriteTestBundle(srv, base64UploadAction(&Base64UploadSpec{
		SourceField:     "file_path",
		ContentField:    "file",
		MaxDecodedBytes: 1024,
		MaxEncodedBytes: 50,
	}))

	_, err := Write(context.Background(), b, connectors.WriteRequest{
		Action: "upload_attachment",
		Config: connectors.RuntimeConfig{ProjectDir: dir},
	}, []connectors.Record{{"file_path": rel, "filename": "p.bin", "contentType": "application/octet-stream"}}, nil)
	if err == nil {
		t.Fatal("over encoded bound: want error, got nil")
	}
	if !strings.Contains(err.Error(), "encoded") {
		t.Fatalf("error should name the encoded bound, got %v", err)
	}
}

func TestWriteBase64UploadHonorsApprovedPayloadDigest(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{}`)
	content := []byte("approved bytes")
	dir, rel := writeTempPayload(t, "file_path_payload.txt", content)

	b := newWriteTestBundle(srv, base64UploadAction(&Base64UploadSpec{
		SourceField:     "file_path",
		ContentField:    "file",
		MaxDecodedBytes: 1024,
	}))
	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:])
	rec := []connectors.Record{{"file_path": rel, "filename": "p.txt", "contentType": "text/plain"}}

	if _, err := Write(context.Background(), b, connectors.WriteRequest{
		Action: "upload_attachment",
		Config: connectors.RuntimeConfig{
			ProjectDir:            dir,
			ApprovedPayloadSHA256: map[string]string{connectors.PayloadApprovalKey(0, "file_path"): digest},
		},
	}, rec, nil); err != nil {
		t.Fatalf("matching digest: unexpected error: %v", err)
	}
	if cap.method == "" {
		t.Fatal("matching digest: request should have been issued")
	}

	cap.method = ""
	_, err := Write(context.Background(), b, connectors.WriteRequest{
		Action: "upload_attachment",
		Config: connectors.RuntimeConfig{
			ProjectDir:            dir,
			ApprovedPayloadSHA256: map[string]string{connectors.PayloadApprovalKey(0, "file_path"): strings.Repeat("0", 64)},
		},
	}, rec, nil)
	if err == nil {
		t.Fatal("substituted payload: want digest mismatch error, got nil")
	}
	if cap.method != "" {
		t.Fatal("substituted payload must not be transmitted")
	}

	cap.method = ""
	_, err = Write(context.Background(), b, connectors.WriteRequest{
		Action: "upload_attachment",
		Config: connectors.RuntimeConfig{
			ProjectDir:            dir,
			ApprovedPayloadSHA256: map[string]string{connectors.PayloadApprovalKey(0, "other"): digest},
		},
	}, rec, nil)
	if err == nil {
		t.Fatal("unapproved field: want error, got nil")
	}
}

func TestDryRunBase64UploadDoesNotReadTheFile(t *testing.T) {
	srv, _ := captureServer(t, http.StatusOK, `{}`)
	b := newWriteTestBundle(srv, base64UploadAction(&Base64UploadSpec{
		SourceField:     "file_path",
		ContentField:    "file",
		MaxDecodedBytes: 1024,
	}))

	preview, err := DryRunWrite(context.Background(), b, connectors.WriteRequest{
		Action: "upload_attachment",
		Config: connectors.RuntimeConfig{ProjectDir: t.TempDir()},
	}, []connectors.Record{{"file_path": "not-on-disk.txt", "filename": "x", "contentType": "text/plain"}}, nil)
	if err != nil {
		t.Fatalf("DryRunWrite must not touch the filesystem: %v", err)
	}
	if preview.RecordsStaged != 1 {
		t.Fatalf("RecordsStaged = %d, want 1", preview.RecordsStaged)
	}
}

func githubReleaseAssetUploadAction(t *testing.T, serverURL string) Bundle {
	t.Helper()
	bundle, err := Load(os.DirFS(filepath.Join("..", "defs")), "github")
	if err != nil {
		t.Fatalf("load installed GitHub bundle: %v", err)
	}
	var installed WriteAction
	for _, action := range bundle.Writes {
		if action.Name == "releases_release_id_assets2" {
			installed = action
			break
		}
	}
	if installed.Name == "" {
		t.Fatal("installed GitHub release-asset upload action is missing")
	}
	// The production declaration pins uploads.github.com. The test substitutes
	// only the origin so the real installed path/query/body contract can be
	// exercised without live credentials or provider I/O.
	installed.BaseURL = serverURL
	installed.AllowedBaseURLOrigins = []string{serverURL}
	bundle.HTTP.URL = serverURL
	bundle.HTTP.Auth = nil
	bundle.HTTP.Headers = nil
	bundle.Writes = []WriteAction{installed}
	return bundle
}

func TestGitHubReleaseAssetUpload_InstalledCommandSendsExactBytes(t *testing.T) {
	var calls int
	var gotBody []byte
	var gotQuery url.Values
	var gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		gotQuery = r.URL.Query()
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":9007199254740993,"name":"asset.bin"}`))
	}))
	defer server.Close()

	payload := []byte{0x00, 0xff, 'p', 'm', '\n'}
	projectDir, relative := writeTempPayload(t, "asset.bin", payload)
	bundle := githubReleaseAssetUploadAction(t, server.URL)
	result, err := Write(context.Background(), bundle, connectors.WriteRequest{
		Action: "releases_release_id_assets2",
		Config: connectors.RuntimeConfig{
			ProjectDir: projectDir,
			Config:     map[string]string{"owner": "octo cat", "repo": "hello/world"},
		},
	}, []connectors.Record{{
		"release_id": int64(42),
		"name":       "asset +1.bin",
		"label":      "release/one",
		"file_path":  relative,
	}}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if calls != 1 {
		t.Fatalf("provider calls = %d, want 1", calls)
	}
	if !bytes.Equal(gotBody, payload) {
		t.Fatalf("body = %x, want %x", gotBody, payload)
	}
	if gotContentType != "application/octet-stream" {
		t.Fatalf("Content-Type = %q, want application/octet-stream", gotContentType)
	}
	if gotQuery.Get("name") != "asset +1.bin" || gotQuery.Get("label") != "release/one" {
		t.Fatalf("query = %q", gotQuery.Encode())
	}
	if len(result.ProviderResponses) != 1 || result.ProviderResponses[0].Status != http.StatusCreated {
		t.Fatalf("provider responses = %#v", result.ProviderResponses)
	}
	if body, ok := result.ProviderResponses[0].Body.(map[string]any); !ok || body["id"] != json.Number("9007199254740993") {
		t.Fatalf("provider response body = %#v", result.ProviderResponses[0].Body)
	}
}

func TestGitHubReleaseAssetUpload_RejectsMissingChangedUnsafeOrOversizeFile(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	baseRecord := connectors.Record{"release_id": int64(42), "name": "asset.bin"}

	t.Run("missing", func(t *testing.T) {
		bundle := githubReleaseAssetUploadAction(t, server.URL)
		_, err := Write(context.Background(), bundle, connectors.WriteRequest{Action: "releases_release_id_assets2", Config: connectors.RuntimeConfig{Config: map[string]string{"owner": "o", "repo": "r"}}}, []connectors.Record{baseRecord}, nil)
		if err == nil || !strings.Contains(err.Error(), "file_path") {
			t.Fatalf("error = %v, want missing file_path", err)
		}
	})

	t.Run("unsafe", func(t *testing.T) {
		bundle := githubReleaseAssetUploadAction(t, server.URL)
		record := connectors.Record{"release_id": int64(42), "name": "asset.bin", "file_path": "../escape.bin"}
		_, err := Write(context.Background(), bundle, connectors.WriteRequest{Action: "releases_release_id_assets2", Config: connectors.RuntimeConfig{ProjectDir: t.TempDir(), Config: map[string]string{"owner": "o", "repo": "r"}}}, []connectors.Record{record}, nil)
		if err == nil || (!strings.Contains(strings.ToLower(err.Error()), "outside") && !strings.Contains(strings.ToLower(err.Error()), "project root")) {
			t.Fatalf("error = %v, want confinement refusal", err)
		}
	})

	t.Run("oversize", func(t *testing.T) {
		bundle := githubReleaseAssetUploadAction(t, server.URL)
		for index := range bundle.Writes {
			bundle.Writes[index].BinaryUpload.MaxBytes = 3
		}
		dir, path := writeTempPayload(t, "large.bin", []byte("four"))
		record := connectors.Record{"release_id": int64(42), "name": "asset.bin", "file_path": path}
		_, err := Write(context.Background(), bundle, connectors.WriteRequest{Action: "releases_release_id_assets2", Config: connectors.RuntimeConfig{ProjectDir: dir, Config: map[string]string{"owner": "o", "repo": "r"}}}, []connectors.Record{record}, nil)
		if err == nil || !strings.Contains(err.Error(), "too large") {
			t.Fatalf("error = %v, want size refusal", err)
		}
	})

	t.Run("changed", func(t *testing.T) {
		bundle := githubReleaseAssetUploadAction(t, server.URL)
		dir, path := writeTempPayload(t, "changed.bin", []byte("changed"))
		record := connectors.Record{"release_id": int64(42), "name": "asset.bin", "file_path": path}
		_, err := Write(context.Background(), bundle, connectors.WriteRequest{Action: "releases_release_id_assets2", Config: connectors.RuntimeConfig{
			ProjectDir: dir,
			Config:     map[string]string{"owner": "o", "repo": "r"},
			ApprovedPayloadSHA256: map[string]string{
				connectors.PayloadApprovalKey(0, "file_path"): strings.Repeat("0", 64),
			},
		}}, []connectors.Record{record}, nil)
		if err == nil || !strings.Contains(err.Error(), "approved digest") {
			t.Fatalf("error = %v, want approved digest refusal", err)
		}
	})

	if calls != 0 {
		t.Fatalf("invalid uploads reached provider %d time(s)", calls)
	}
}

func TestGitHubReleaseAssetUpload_EmptyFileAndTerminalFailures(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var body []byte
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()
		dir, path := writeTempPayload(t, "empty.bin", nil)
		bundle := githubReleaseAssetUploadAction(t, server.URL)
		_, err := Write(context.Background(), bundle, connectors.WriteRequest{Action: "releases_release_id_assets2", Config: connectors.RuntimeConfig{ProjectDir: dir, Config: map[string]string{"owner": "o", "repo": "r"}}}, []connectors.Record{{"release_id": int64(42), "name": "empty.bin", "file_path": path}}, nil)
		if err != nil {
			t.Fatalf("empty upload: %v", err)
		}
		if body == nil || len(body) != 0 {
			t.Fatalf("empty body = %#v, want present zero-byte body", body)
		}
	})

	t.Run("redirect and 4xx", func(t *testing.T) {
		var redirected int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/redirected" {
				redirected++
				w.WriteHeader(http.StatusCreated)
				return
			}
			if r.URL.Query().Get("name") == "redirect.bin" {
				http.Redirect(w, r, "/redirected", http.StatusTemporaryRedirect)
				return
			}
			w.Header().Add("X-Request-Id", "occurrence-1")
			w.Header().Add("X-Request-Id", "occurrence-2")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{"message":"duplicate asset"}`))
		}))
		defer server.Close()
		dir, path := writeTempPayload(t, "payload.bin", []byte("payload"))
		bundle := githubReleaseAssetUploadAction(t, server.URL)
		request := func(name string) (connectors.WriteResult, error) {
			return Write(context.Background(), bundle, connectors.WriteRequest{Action: "releases_release_id_assets2", Config: connectors.RuntimeConfig{ProjectDir: dir, Config: map[string]string{"owner": "o", "repo": "r"}}}, []connectors.Record{{"release_id": int64(42), "name": name, "file_path": path}}, nil)
		}
		if _, err := request("redirect.bin"); err == nil || !strings.Contains(strings.ToLower(err.Error()), "redirect") {
			t.Fatalf("redirect error = %v", err)
		}
		if redirected != 0 {
			t.Fatalf("redirect target calls = %d, want 0", redirected)
		}
		result, err := request("duplicate.bin")
		if err == nil {
			t.Fatal("4xx upload: want terminal error")
		}
		if len(result.ProviderResponses) != 1 || result.ProviderResponses[0].Status != http.StatusUnprocessableEntity {
			t.Fatalf("4xx receipt = %#v", result.ProviderResponses)
		}
		if got := result.ProviderResponses[0].Headers["X-Request-Id"].Values; !reflect.DeepEqual(got, []string{"occurrence-1", "occurrence-2"}) {
			t.Fatalf("X-Request-Id = %#v", got)
		}
	})
}

// TestGitHubReleaseAssetUpload_RejectsEnterpriseCrossOriginBeforeIO proves an
// Enterprise credential cannot be replayed to GitHub's public upload host.
// Both servers are armed: a passing error alone would be insufficient if the
// request had already disclosed the Authorization header to either host.
func TestGitHubReleaseAssetUpload_RejectsEnterpriseCrossOriginBeforeIO(t *testing.T) {
	var enterpriseCalls, uploadCalls int
	var enterpriseAuthorization, uploadAuthorization string
	enterprise := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enterpriseCalls++
		enterpriseAuthorization = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
	}))
	defer enterprise.Close()
	upload := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploadCalls++
		uploadAuthorization = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
	}))
	defer upload.Close()

	dir, relative := writeTempPayload(t, "enterprise-asset.bin", []byte("enterprise-safe"))
	bundle := githubReleaseAssetUploadAction(t, upload.URL)
	bundle.HTTP.URL = enterprise.URL
	bundle.HTTP.Auth = []AuthSpec{{Mode: "bearer", Token: "{{ secrets.token }}", When: "{{ secrets.token }}"}}
	_, err := Write(context.Background(), bundle, connectors.WriteRequest{
		Action: "releases_release_id_assets2",
		Config: connectors.RuntimeConfig{
			ProjectDir: dir,
			Config:     map[string]string{"owner": "octo", "repo": "enterprise"},
			Secrets:    map[string]string{"token": "NONSECRET-TEST-TOKEN"},
		},
	}, []connectors.Record{{"release_id": int64(42), "name": "enterprise-asset.bin", "file_path": relative}}, nil)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "origin") {
		t.Fatalf("Write enterprise upload error = %v, want alternate-origin refusal", err)
	}
	if enterpriseCalls != 0 || uploadCalls != 0 || enterpriseAuthorization != "" || uploadAuthorization != "" {
		t.Fatalf("enterprise/upload calls and authorization = %d/%d %q/%q, want no provider I/O", enterpriseCalls, uploadCalls, enterpriseAuthorization, uploadAuthorization)
	}
}

// TestGitHubReleaseAssetUpload_RequiresCreatedResponse keeps the provider's
// receipt but refuses every other successful-looking HTTP status. GitHub's
// upload contract is 201, not generic 2xx.
func TestGitHubReleaseAssetUpload_RequiresCreatedResponse(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls++
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"id":"wrong-success"}`))
			}))
			defer server.Close()
			dir, relative := writeTempPayload(t, "wrong-success.bin", []byte("receipt"))
			result, err := Write(context.Background(), githubReleaseAssetUploadAction(t, server.URL), connectors.WriteRequest{
				Action: "releases_release_id_assets2", Config: connectors.RuntimeConfig{ProjectDir: dir, Config: map[string]string{"owner": "o", "repo": "r"}},
			}, []connectors.Record{{"release_id": int64(42), "name": "wrong-success.bin", "file_path": relative}}, nil)
			if err == nil {
				t.Fatalf("Write status %d = nil error, want declared-201 refusal", status)
			}
			if calls != 1 || len(result.ProviderResponses) != 1 || result.ProviderResponses[0].Status != status || result.RecordsWritten != 0 || result.RecordsFailed != 1 {
				t.Fatalf("status %d result = %#v calls=%d, want retained failed receipt", status, result, calls)
			}
		})
	}
}

// TestGitHubReleaseAssetUpload_EnforcesDeclaredMediaPolicy is deliberately a
// runtime assertion. Merely parsing allowed_media_types would still allow a
// binary executor to send bytes under a media contract it cannot honour.
func TestGitHubReleaseAssetUpload_EnforcesDeclaredMediaPolicy(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	dir, relative := writeTempPayload(t, "policy.bin", []byte("policy"))
	bundle := githubReleaseAssetUploadAction(t, server.URL)
	bundle.Writes[0].BinaryUpload.AllowedMediaTypes = []string{"image/png"}
	_, err := Write(context.Background(), bundle, connectors.WriteRequest{
		Action: "releases_release_id_assets2", Config: connectors.RuntimeConfig{ProjectDir: dir, Config: map[string]string{"owner": "o", "repo": "r"}},
	}, []connectors.Record{{"release_id": int64(42), "name": "policy.bin", "file_path": relative}}, nil)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "media") {
		t.Fatalf("Write media-policy error = %v, want refusal", err)
	}
	if calls != 0 {
		t.Fatalf("unenforceable media policy reached provider %d time(s)", calls)
	}
}
