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
	if got, want := len(bundle.Streams), 9; got != want {
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
	if got, want := coveredStreams, 9; got != want {
		t.Fatalf("stream-covered operations = %d, want %d", got, want)
	}
	if got, want := blocked, 51; got != want {
		t.Fatalf("blocked/planned operations = %d, want %d", got, want)
	}
}

func TestPublishedStreamsNeedNoRequiredRequestFields(t *testing.T) {
	published := map[string]bool{}
	for _, stream := range cloudTrailPublishedStreams {
		published[stream] = true
		action, ok := cloudTrailStreamActions[stream]
		if !ok {
			t.Fatalf("published stream %s missing action", stream)
		}
		for _, field := range cloudTrailActionFields[action] {
			if field.Required {
				t.Fatalf("published stream %s action %s requires %s", stream, action, field.Name)
			}
		}
	}
	if len(cloudTrailStreamActions) != len(cloudTrailPublishedStreams) {
		t.Fatalf("stream action map has %d entries, want %d", len(cloudTrailStreamActions), len(cloudTrailPublishedStreams))
	}
	for stream := range cloudTrailStreamActions {
		if !published[stream] {
			t.Fatalf("unpublished stream %s is dispatchable", stream)
		}
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

func TestNativeCloudTrailCheckUsesImplementedTarget(t *testing.T) {
	var gotTarget string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTarget = r.Header.Get("X-Amz-Target")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trailList":[]}`))
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client()}
	if err := c.Check(context.Background(), fixtureRuntimeConfig(srv.URL)); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if got, want := gotTarget, cloudTrailTarget("DescribeTrails"); got != want {
		t.Fatalf("X-Amz-Target = %q, want %q", got, want)
	}
	if len(gotBody) != 0 {
		t.Fatalf("body = %#v, want empty DescribeTrails body by default", gotBody)
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

func TestNativeCloudTrailRejectsInvalidMaxPages(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "malformed", value: "ten"},
		{name: "negative", value: "-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"Channels":[{"Name":"unexpected"}]}`))
			}))
			defer srv.Close()

			cfg := fixtureRuntimeConfig(srv.URL)
			cfg.Config["max_pages"] = tt.value
			c := Connector{Client: srv.Client()}
			err := c.Read(context.Background(), connectors.ReadRequest{Stream: "list_channels", Config: cfg}, func(connectors.Record) error {
				t.Fatal("emit called for invalid max_pages")
				return nil
			})
			if err == nil {
				t.Fatalf("Read unexpectedly accepted max_pages=%s", tt.value)
			}
			if !strings.Contains(err.Error(), "max_pages") {
				t.Fatalf("Read error = %v, want max_pages validation", err)
			}
			if requests != 0 {
				t.Fatalf("requests = %d, want 0", requests)
			}
		})
	}
}

func TestNativeCloudTrailRejectsBlockedReadStreams(t *testing.T) {
	blocked := []string{
		"get_channel",
		"get_dashboard",
		"get_event_data_store",
		"get_event_selectors",
		"get_import",
		"get_resource_policy",
		"get_trail",
		"get_trail_status",
		"list_import_failures",
		"list_tags",
		"management_events",
		"read_only_events",
		"write_only_events",
		"console_logins",
	}
	c := Connector{}
	for _, stream := range blocked {
		t.Run(stream, func(t *testing.T) {
			err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream}, func(connectors.Record) error {
				t.Fatalf("emit called for blocked stream %s", stream)
				return nil
			})
			if err == nil {
				t.Fatalf("Read(%s) unexpectedly succeeded", stream)
			}
		})
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
