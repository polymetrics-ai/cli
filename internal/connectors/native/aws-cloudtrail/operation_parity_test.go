package awscloudtrail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
	if got, want := len(bundle.Operations), 0; got != want {
		t.Fatalf("implemented direct operations = %d, want %d", got, want)
	}
	if got, want := len(bundle.Writes), 0; got != want {
		t.Fatalf("implemented write actions = %d, want %d", got, want)
	}
	blocked := 0
	coveredStreams := 0
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.CoveredBy != nil && endpoint.CoveredBy.Stream != "" {
			coveredStreams++
		}
		if endpoint.Operation != nil && endpoint.Operation.Status == "blocked" {
			blocked++
		}
	}
	if got, want := coveredStreams, 19; got != want {
		t.Fatalf("stream-covered operations = %d, want %d", got, want)
	}
	if got, want := blocked, 41; got != want {
		t.Fatalf("blocked/planned operations = %d, want %d", got, want)
	}
}

func TestOperationLedgerNoRawEscapes(t *testing.T) {
	bundle, err := engine.Load(os.DirFS("../../defs"), "aws-cloudtrail")
	if err != nil {
		t.Fatalf("load bundle: %v", err)
	}
	for _, stream := range bundle.Streams {
		if stream.Path != "/" || stream.Method != http.MethodPost {
			t.Fatalf("stream %s request = %s %s, want fixed POST /", stream.Name, stream.Method, stream.Path)
		}
	}
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.CoveredBy != nil && endpoint.CoveredBy.Stream != "" && endpoint.Method != "READ_POST" {
			t.Fatalf("stream ledger endpoint %s method = %q, want READ_POST logical read-over-POST marker", endpoint.Path, endpoint.Method)
		}
	}
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.Operation == nil {
			continue
		}
		if endpoint.Operation.Status != "blocked" || !endpoint.Operation.BlockedByDefault {
			t.Fatalf("blocked endpoint %+v is not blocked by default", endpoint)
		}
	}
}

func TestNativeCloudTrailReadDispatchesStreamTarget(t *testing.T) {
	var gotTarget string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTarget = r.Header.Get("X-Amz-Target")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trailList":[{"Name":"trail-fixture"}]}`))
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client()}
	var records []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{
		Stream: "describe_trails",
		Config: fixtureRuntimeConfig(srv.URL),
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	wantTarget := cloudTrailTarget("DescribeTrails")
	if gotTarget != wantTarget {
		t.Fatalf("X-Amz-Target = %q, want %q", gotTarget, wantTarget)
	}
	if len(records) != 1 || records[0]["Name"] != "trail-fixture" {
		t.Fatalf("records = %#v", records)
	}
	if len(gotBody) != 0 {
		t.Fatalf("body = %#v, want empty DescribeTrails body by default", gotBody)
	}
}

func TestNativeCloudTrailDirectAndWritesAreBlockedInScopeCorrectedSurface(t *testing.T) {
	c := Connector{}
	if _, err := c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{Operation: "aws-cloudtrail.lookup_events"}); err == nil {
		t.Fatal("OperationDirectRead unexpectedly succeeded for blocked direct operation")
	}
	if err := c.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "start_logging"}, []connectors.Record{{"Name": "trail-fixture"}}); err == nil {
		t.Fatal("ValidateWrite unexpectedly accepted blocked write action")
	}
	if _, err := c.DryRunWrite(context.Background(), connectors.WriteRequest{Action: "start_logging"}, []connectors.Record{{"Name": "trail-fixture"}}); err == nil {
		t.Fatal("DryRunWrite unexpectedly accepted blocked write action")
	}
	if got, err := c.Write(context.Background(), connectors.WriteRequest{Action: "start_logging"}, []connectors.Record{{"Name": "trail-fixture"}}); err == nil || got.RecordsFailed != 1 {
		t.Fatalf("Write result = %+v err = %v, want blocked failure", got, err)
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
