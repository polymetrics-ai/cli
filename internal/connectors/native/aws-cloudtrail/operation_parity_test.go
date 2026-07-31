package awscloudtrail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestOperationLedgerCounts(t *testing.T) {
	bundle, err := engine.Load(os.DirFS("../../defs"), "aws-cloudtrail")
	if err != nil {
		t.Fatalf("load bundle: %v", err)
	}
	if got, want := len(bundle.Surface.Endpoints), 60; got != want {
		t.Fatalf("api_surface rows = %d, want %d", got, want)
	}
	if got, want := len(bundle.Streams), 19; got != want {
		t.Fatalf("streams = %d, want %d", got, want)
	}
	if got, want := len(bundle.Operations), 10; got != want {
		t.Fatalf("direct operations = %d, want %d", got, want)
	}
	if got, want := len(bundle.Writes), 31; got != want {
		t.Fatalf("write actions = %d, want %d", got, want)
	}
}

func TestOperationLedgerNoRawEscapes(t *testing.T) {
	bundle, err := engine.Load(os.DirFS("../../defs"), "aws-cloudtrail")
	if err != nil {
		t.Fatalf("load bundle: %v", err)
	}
	for _, op := range bundle.Operations {
		if op.Kind != "rest_read" {
			t.Fatalf("operation %s kind = %q, want rest_read", op.ID, op.Kind)
		}
		if op.REST == nil || strings.TrimSpace(op.REST.Path) != "/" || strings.ToUpper(op.REST.Method) != http.MethodPost {
			t.Fatalf("operation %s REST = %+v, want fixed POST /", op.ID, op.REST)
		}
		if len(op.REST.BodySchema) == 0 {
			t.Fatalf("operation %s has no closed body_schema", op.ID)
		}
	}
	for _, action := range bundle.Writes {
		if strings.TrimSpace(action.Path) != "/" || strings.ToUpper(action.Method) != http.MethodPost {
			t.Fatalf("write %s request = %s %s, want fixed POST /", action.Name, action.Method, action.Path)
		}
		if len(action.RecordSchema) == 0 {
			t.Fatalf("write %s has no record_schema", action.Name)
		}
	}
}

func TestNativeCloudTrailJSONRPCDispatchesOperationTarget(t *testing.T) {
	var gotTarget string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTarget = r.Header.Get("X-Amz-Target")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"QueryId":"11111111-1111-1111-1111-111111111111"}`))
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client()}
	_, err := c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "aws-cloudtrail.describe_query",
		Config:    fixtureRuntimeConfig(srv.URL),
		Body:      map[string]any{"QueryId": "11111111-1111-1111-1111-111111111111"},
	})
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	wantTarget := cloudTrailTarget("DescribeQuery")
	if gotTarget != wantTarget {
		t.Fatalf("X-Amz-Target = %q, want %q", gotTarget, wantTarget)
	}
	if gotBody["QueryId"] != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("body = %#v", gotBody)
	}
}

func TestNativeCloudTrailOperationDirectReadRedactsSensitiveDefaults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Events":[{"EventId":"evt-1","CloudTrailEvent":"sensitive-event-json","AccessKeyId":"sensitive-key"}]}`))
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client()}
	result, err := c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "aws-cloudtrail.lookup_events",
		Config:    fixtureRuntimeConfig(srv.URL),
	})
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	body := result.Body.(map[string]any)
	events := body["Events"].([]any)
	first := events[0].(map[string]any)
	if first["CloudTrailEvent"] != "[REDACTED]" || first["AccessKeyId"] != "[REDACTED]" {
		t.Fatalf("sensitive fields not redacted: %#v", first)
	}
}

func TestNativeCloudTrailWriteDispatchesActionTarget(t *testing.T) {
	var gotTarget string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTarget = r.Header.Get("X-Amz-Target")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client()}
	result, err := c.Write(context.Background(), connectors.WriteRequest{Action: "start_logging", Config: fixtureRuntimeConfig(srv.URL)}, []connectors.Record{{"Name": "trail-fixture"}})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("result = %+v", result)
	}
	wantTarget := cloudTrailTarget("StartLogging")
	if gotTarget != wantTarget {
		t.Fatalf("X-Amz-Target = %q, want %q", gotTarget, wantTarget)
	}
	if gotBody["Name"] != "trail-fixture" {
		t.Fatalf("body = %#v", gotBody)
	}
}

func fixtureRuntimeConfig(baseURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"aws_region_name": "us-east-1",
			"base_url":        baseURL,
		},
		Secrets: map[string]string{
			"aws_key_id":     "fixture-access-key",
			"aws_secret_key": "fixture-secret-key",
		},
	}
}
