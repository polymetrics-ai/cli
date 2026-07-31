package dynamodb

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func testConfig(endpoint string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config:  map[string]string{"endpoint": endpoint, "streams_endpoint": endpoint, "region": "us-east-1", "table_name": "fixture_table", "stream_arn": "arn:aws:dynamodb:us-east-1:123456789012:table/fixture/stream/2026", "shard_id": "shard-000"},
		Secrets: map[string]string{"access_key_id": "AKID", "secret_access_key": "SECRET"},
	}
}

func TestParityOperationRegistryCounts(t *testing.T) {
	b, err := engine.Load(os.DirFS(defsRoot(t)), "dynamodb")
	if err != nil {
		t.Fatal(err)
	}
	if len(readOperations) != 27 || len(b.Streams) != 27 {
		t.Fatalf("read streams: native=%d bundle=%d, want 27", len(readOperations), len(b.Streams))
	}
	if len(writeOperations) != 26 || len(b.Writes) != 26 {
		t.Fatalf("write actions: native=%d bundle=%d, want 26", len(writeOperations), len(b.Writes))
	}
	if len(b.Operations) != 3 {
		t.Fatalf("direct operations = %d, want 3", len(b.Operations))
	}
	if b.Surface == nil || len(b.Surface.Endpoints) != 61 {
		t.Fatalf("api surface endpoints = %d, want 61", len(b.Surface.Endpoints))
	}
	for _, action := range b.Writes {
		if action.Name == "put_resource_policy" && action.Confirm != "destructive" {
			t.Fatalf("put_resource_policy confirm = %q, want destructive", action.Confirm)
		}
	}
	counts := map[string]int{}
	for _, ep := range b.Surface.Endpoints {
		switch {
		case ep.CoveredBy != nil && ep.CoveredBy.Stream != "":
			counts["stream"]++
		case ep.CoveredBy != nil && ep.CoveredBy.Write != "":
			counts["write"]++
		case ep.CoveredBy != nil && ep.CoveredBy.DirectRead != "":
			counts["direct"]++
		case ep.Operation != nil && ep.Operation.Model == "binary_read":
			counts["binary_blocked"]++
		case ep.Operation != nil && ep.Operation.Model == "admin_reverse_etl":
			counts["binary_blocked"]++
		case ep.Operation != nil && ep.Operation.Model == "disallowed":
			counts["disallowed"]++
		}
	}
	want := map[string]int{"stream": 27, "write": 26, "direct": 3, "binary_blocked": 2, "disallowed": 3}
	for key, expected := range want {
		if counts[key] != expected {
			t.Fatalf("surface %s count = %d, want %d (all counts=%v)", key, counts[key], expected, counts)
		}
	}
}

func TestReadFixtureCoversEveryStream(t *testing.T) {
	c := New()
	for _, op := range readOperations {
		t.Run(op.Stream, func(t *testing.T) {
			count := 0
			err := c.Read(context.Background(), connectors.ReadRequest{Stream: op.Stream, Config: connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}}, func(connectors.Record) error {
				count++
				return nil
			})
			if err != nil {
				t.Fatalf("Read fixture: %v", err)
			}
			if count == 0 {
				t.Fatal("fixture read emitted zero records")
			}
		})
	}
}

func TestWriteActionsValidatePreviewAndExecuteAgainstReplay(t *testing.T) {
	seen := map[string]int{}
	bodies := map[string]map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-Amz-Target")
		seen[target]++
		if r.Method != http.MethodPost || r.URL.Path != "/" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		bodies[target] = body
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client(), Now: func() time.Time { return time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC) }}
	for _, op := range writeOperations {
		t.Run(op.Action, func(t *testing.T) {
			rec := writeFixtureRecord(t, op.Action)
			req := connectors.WriteRequest{Action: op.Action, Config: testConfig(srv.URL)}
			if err := c.ValidateWrite(context.Background(), req, []connectors.Record{rec}); err != nil {
				t.Fatalf("ValidateWrite: %v", err)
			}
			preview, err := c.DryRunWrite(context.Background(), req, []connectors.Record{rec})
			if err != nil {
				t.Fatalf("DryRunWrite: %v", err)
			}
			if preview.RecordsStaged != 1 || !strings.Contains(strings.Join(preview.Warnings, "\n"), op.Target) {
				t.Fatalf("preview = %+v, want target %s", preview, op.Target)
			}
			result, err := c.Write(context.Background(), req, []connectors.Record{rec})
			if err != nil {
				t.Fatalf("Write: %v", err)
			}
			if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
				t.Fatalf("Write result = %+v", result)
			}
		})
	}
	for _, op := range writeOperations {
		if seen[op.Target] != 1 {
			t.Fatalf("target %s seen %d times, want 1", op.Target, seen[op.Target])
		}
	}
	batchBody := bodies[dynamoTargetPrefix+"BatchWriteItem"]
	if _, raw := batchBody["request_items"]; raw {
		t.Fatalf("batch_write_item leaked raw request_items body: %v", batchBody)
	}
	if requestItems, ok := batchBody["RequestItems"].(map[string]any); !ok || len(requestItems) != 1 {
		t.Fatalf("batch_write_item body = %v, want typed RequestItems", batchBody)
	}
	transactBody := bodies[dynamoTargetPrefix+"TransactWriteItems"]
	if _, raw := transactBody["transact_items"]; raw {
		t.Fatalf("transact_write_items leaked raw transact_items body: %v", transactBody)
	}
	if items, ok := transactBody["TransactItems"].([]any); !ok || len(items) != 1 {
		t.Fatalf("transact_write_items body = %v, want typed TransactItems", transactBody)
	}
}

func TestOperationDirectReadBuildsClosedBodies(t *testing.T) {
	cases := []struct {
		operation string
		body      map[string]any
		target    string
	}{
		{"get_item", map[string]any{"table_name": "users", "key_name": "pk", "key_value": "user#1"}, dynamoTargetPrefix + "GetItem"},
		{"batch_get_item", map[string]any{"table_name": "users", "key_name": "pk", "key_type": "S", "key_values": []any{"user#1", "user#2"}}, dynamoTargetPrefix + "BatchGetItem"},
		{"transact_get_items", map[string]any{"table_name": "users", "key_name": "pk", "key_type": "S", "key_values": []any{"user#1"}}, dynamoTargetPrefix + "TransactGetItems"},
	}
	for _, tc := range cases {
		t.Run(tc.operation, func(t *testing.T) {
			body, target, err := directReadBody(connectors.OperationDirectReadRequest{Operation: tc.operation, Body: tc.body})
			if err != nil {
				t.Fatal(err)
			}
			if target != tc.target {
				t.Fatalf("target = %s, want %s", target, tc.target)
			}
			encoded, _ := json.Marshal(body)
			if strings.Contains(string(encoded), "Statement") || strings.Contains(string(encoded), "PartiQL") {
				t.Fatalf("direct body exposed raw statement: %s", encoded)
			}
		})
	}
	body, _, err := directReadBody(connectors.OperationDirectReadRequest{Operation: "get_item", Body: map[string]any{"table_name": "users", "key_name": "pk", "key_value": "user#1"}})
	if err != nil {
		t.Fatal(err)
	}
	key := body["Key"].(map[string]any)["pk"].(map[string]any)
	if key["S"] != "user#1" {
		t.Fatalf("omitted key_type defaulted to %v, want S", key)
	}
}

func TestOperationDirectReadRejectsMissingRequiredFields(t *testing.T) {
	cases := []struct {
		name      string
		operation string
		body      map[string]any
		want      string
	}{
		{"missing table", "get_item", map[string]any{"key_name": "pk", "key_value": "user#1"}, "table_name"},
		{"missing get key name", "get_item", map[string]any{"table_name": "users", "key_value": "user#1"}, "key_name and key_value"},
		{"missing get key value", "get_item", map[string]any{"table_name": "users", "key_name": "pk"}, "key_name and key_value"},
		{"missing batch values", "batch_get_item", map[string]any{"table_name": "users", "key_name": "pk"}, "key_values"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := directReadBody(connectors.OperationDirectReadRequest{Operation: tc.operation, Body: tc.body})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestOperationDirectReadEnforcesMaxBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Item":{"pk":{"S":"user#1"}}}`))
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client(), Now: func() time.Time { return time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC) }}
	_, err := c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "get_item",
		Config:    testConfig(srv.URL),
		Body:      map[string]any{"table_name": "users", "key_name": "pk", "key_value": "user#1"},
		MaxBytes:  8,
	})
	if err == nil || !strings.Contains(err.Error(), "max_bytes 8") {
		t.Fatalf("OperationDirectRead error = %v, want max_bytes limit", err)
	}
}

func TestReadListStreamsUsesDocumentedPagination(t *testing.T) {
	var bodies []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Amz-Target"); got != streamsTargetPrefix+"ListStreams" {
			t.Fatalf("target = %s, want ListStreams", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		bodies = append(bodies, body)
		if len(bodies) == 1 {
			_, _ = w.Write([]byte(`{"Streams":[{"StreamArn":"arn:stream:1"}],"LastEvaluatedStreamArn":"arn:stream:1"}`))
			return
		}
		_, _ = w.Write([]byte(`{"Streams":[{"StreamArn":"arn:stream:2"}]}`))
	}))
	defer srv.Close()

	cfg := testConfig(srv.URL)
	cfg.Config["page_size"] = "1"
	cfg.Config["max_pages"] = "2"
	c := Connector{Client: srv.Client(), Now: func() time.Time { return time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC) }}
	var records []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "streams_list_streams", Config: cfg}, func(rec connectors.Record) error {
		records = append(records, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read streams_list_streams: %v", err)
	}
	if len(records) != 2 || len(bodies) != 2 {
		t.Fatalf("records=%d bodies=%d, want 2", len(records), len(bodies))
	}
	if bodies[0]["Limit"] != float64(1) && bodies[0]["Limit"] != 1 {
		t.Fatalf("first ListStreams body = %v, want Limit", bodies[0])
	}
	if _, bad := bodies[0]["MaxResults"]; bad {
		t.Fatalf("first ListStreams body used MaxResults: %v", bodies[0])
	}
	if bodies[1]["ExclusiveStartStreamArn"] != "arn:stream:1" {
		t.Fatalf("second ListStreams body = %v, want ExclusiveStartStreamArn", bodies[1])
	}
}

func TestReadCDCUsesShardIteratorAndEmitsEvents(t *testing.T) {
	calls := []string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-Amz-Target")
		calls = append(calls, target)
		switch target {
		case streamsTargetPrefix + "GetShardIterator":
			_, _ = w.Write([]byte(`{"ShardIterator":"iterator-1"}`))
		case streamsTargetPrefix + "GetRecords":
			_, _ = w.Write([]byte(`{"Records":[{"eventName":"INSERT","dynamodb":{"NewImage":{"pk":{"S":"user#1"}}}}],"NextShardIterator":""}`))
		default:
			t.Fatalf("unexpected target %s", target)
		}
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client(), Now: func() time.Time { return time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC) }}
	var events []connectors.CDCEvent
	err := c.ReadCDC(context.Background(), connectors.CDCReadRequest{Config: testConfig(srv.URL)}, func(event connectors.CDCEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCDC: %v", err)
	}
	if len(events) != 1 || events[0].Operation != "INSERT" || events[0].Record["pk"] != "user#1" {
		t.Fatalf("events = %+v", events)
	}
	if strings.Join(calls, ",") != streamsTargetPrefix+"GetShardIterator,"+streamsTargetPrefix+"GetRecords" {
		t.Fatalf("calls = %v", calls)
	}
}

func defsRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "defs")
}

func writeFixtureRecord(t *testing.T, action string) connectors.Record {
	t.Helper()
	path := filepath.Join(defsRoot(t), "dynamodb", "fixtures", "writes", action+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fx struct {
		Record map[string]any `json:"record"`
	}
	if err := json.Unmarshal(raw, &fx); err != nil {
		t.Fatal(err)
	}
	return connectors.Record(fx.Record)
}
