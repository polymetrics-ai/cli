package amazonsqs_test

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

	"polymetrics.ai/internal/connectors"
	native "polymetrics.ai/internal/connectors/native/amazon-sqs"
)

func TestNameAndMetadata(t *testing.T) {
	c := native.New()
	if c.Name() != "amazon-sqs" {
		t.Fatalf("Name() = %q, want amazon-sqs", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
	if !caps.Write {
		t.Fatalf("amazon-sqs parity connector must expose typed write/admin actions, got Write=false")
	}
}

// TestNoInitRegistration is the required grep-guard (mirrors
// native/postgres's and native/faker's TestNoInitRegistration): the native
// package must NOT call RegisterFactory from anywhere in
// its own source, nor declare an init() function. The registration flip
// (wiring native/amazon-sqs into the production registry) is a wave6
// change; this wave only builds and tests the package standalone.
func TestNoInitRegistration(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed; cannot locate package directory")
	}
	dir := filepath.Dir(thisFile)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", dir, err)
	}

	found := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			// The grep-guard covers the package's own production source, not
			// its tests (this very test file legitimately mentions the
			// forbidden identifiers in prose/identifiers above).
			continue
		}
		found = true
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", e.Name(), err)
		}
		src := string(raw)
		if strings.Contains(src, "RegisterFactory(") {
			t.Fatalf("%s calls RegisterFactory — native/amazon-sqs must NOT self-register (registration flip is wave6)", e.Name())
		}
		if strings.Contains(src, "func init()") {
			t.Fatalf("%s declares an init() function — native/amazon-sqs must perform no registration side effects", e.Name())
		}
	}
	if !found {
		t.Fatal("no non-test .go source files found in native/amazon-sqs; grep-guard did not actually scan anything")
	}
}

// TestConnectorSatisfiesCoreInterfaces compile/runtime-asserts the shape
// required by API-CONTRACT.md / design §B.7 Tier-3: Connector and
// DefinitionProvider. StatefulReader/CDCReader are deliberately NOT
// asserted: legacy amazon-sqs implements neither (SQS's ReceiveMessage has
// no timestamp/offset filter and legacy has no CDC path), so this native
// port carries the identical interface surface forward.
func TestConnectorSatisfiesCoreInterfaces(t *testing.T) {
	c := native.New()
	var _ connectors.Connector = c
	if _, ok := any(c).(connectors.DefinitionProvider); !ok {
		t.Fatal("native amazon-sqs connector must implement connectors.DefinitionProvider (engine.Base)")
	}
}

func fixtureConfig() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
}

func TestCheckFixtureModeOK(t *testing.T) {
	c := native.New()
	if err := c.Check(context.Background(), fixtureConfig()); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
}

func TestCheckRespectsContextCancellation(t *testing.T) {
	c := native.New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Check(ctx, fixtureConfig()); err == nil {
		t.Fatal("Check with a cancelled context: want error, got nil")
	}
}

func TestCheckRequiresQueueURLAndRegion(t *testing.T) {
	c := native.New()
	cases := []struct {
		name string
		cfg  connectors.RuntimeConfig
	}{
		{
			name: "missing queue_url",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"region": "us-east-1"},
				Secrets: map[string]string{"access_key": "a", "secret_key": "s"},
			},
		},
		{
			name: "missing region",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"queue_url": "https://sqs.example.com/1/q"},
				Secrets: map[string]string{"access_key": "a", "secret_key": "s"},
			},
		},
		{
			name: "missing secrets",
			cfg: connectors.RuntimeConfig{
				Config: map[string]string{"queue_url": "https://sqs.example.com/1/q", "region": "us-east-1"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := c.Check(context.Background(), tc.cfg); err == nil {
				t.Fatalf("Check(%s): want error, got nil", tc.name)
			}
		})
	}
}

func TestCheckRejectsUnsafeURLs(t *testing.T) {
	c := native.New()
	baseConfig := map[string]string{"queue_url": "https://sqs.example.com/123/orders", "region": "us-east-1"}
	cases := []struct {
		name   string
		config map[string]string
	}{
		{name: "queue query", config: map[string]string{"queue_url": "https://sqs.example.com/123/orders?Action=PurgeQueue", "region": "us-east-1"}},
		{name: "queue userinfo", config: map[string]string{"queue_url": "https://user:pass@sqs.example.com/123/orders", "region": "us-east-1"}},
		{name: "queue nonlocal http", config: map[string]string{"queue_url": "http://sqs.example.com/123/orders", "region": "us-east-1"}},
		{name: "endpoint query", config: map[string]string{"queue_url": baseConfig["queue_url"], "endpoint_url": "https://sqs.example.com/?Action=PurgeQueue", "region": "us-east-1"}},
		{name: "endpoint path", config: map[string]string{"queue_url": baseConfig["queue_url"], "endpoint_url": "https://sqs.example.com/custom", "region": "us-east-1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := c.Check(context.Background(), connectors.RuntimeConfig{Config: tc.config, Secrets: map[string]string{"access_key": "a", "secret_key": "s"}})
			if err == nil {
				t.Fatal("Check: want unsafe URL validation error, got nil")
			}
		})
	}
}

func TestCatalogHasMessagesStream(t *testing.T) {
	c := native.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) != 1 || cat.Streams[0].Name != "messages" {
		t.Fatalf("Catalog Streams = %+v, want exactly one 'messages' stream", cat.Streams)
	}
	if len(cat.Streams[0].PrimaryKey) != 1 || cat.Streams[0].PrimaryKey[0] != "message_id" {
		t.Fatalf("messages PrimaryKey = %v, want [message_id]", cat.Streams[0].PrimaryKey)
	}
}

func TestReadFixtureEmitsTwoMessages(t *testing.T) {
	c := native.New()
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: fixtureConfig()}, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0]["message_id"] != "message_fixture_1" {
		t.Fatalf("got[0] = %+v, unexpected shape", got[0])
	}
}

func TestReadDefaultsToMessagesStream(t *testing.T) {
	c := native.New()
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Config: fixtureConfig()}, func(connectors.Record) error {
		n++
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n != 2 {
		t.Fatalf("n = %d, want 2 (empty stream defaults to messages)", n)
	}
}

func TestReadUnknownStreamErrors(t *testing.T) {
	c := native.New()
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bogus", Config: fixtureConfig()}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with unknown stream: want error, got nil")
	}
}

func TestReadLiveSignsAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var sawAction string
	var sawQueueURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/123/test-queue" {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		sawAction = r.Form.Get("Action")
		sawQueueURL = r.Form.Get("QueueUrl")
		_, _ = w.Write([]byte(`<ReceiveMessageResponse><ReceiveMessageResult><Message><MessageId>m1</MessageId><ReceiptHandle>rh1</ReceiptHandle><MD5OfBody>md5</MD5OfBody><Body>{"kind":"order","id":1}</Body><Attribute><Name>SentTimestamp</Name><Value>1767225600000</Value></Attribute><MessageAttribute><Name>source</Name><Value><StringValue>checkout</StringValue><DataType>String</DataType></Value></MessageAttribute></Message></ReceiveMessageResult></ReceiveMessageResponse>`))
	}))
	defer srv.Close()

	c := native.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"queue_url": srv.URL + "/123/test-queue", "region": "us-east-1", "max_batch_size": "1"}, Secrets: map[string]string{"access_key": "AKIATEST", "secret_key": "test-secret"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.HasPrefix(sawAuth, "AWS4-HMAC-SHA256 ") || !strings.Contains(sawAuth, "Credential=AKIATEST/") {
		t.Fatalf("Authorization header was not SigV4: %q", sawAuth)
	}
	if sawAction != "ReceiveMessage" {
		t.Fatalf("Action = %q, want ReceiveMessage", sawAction)
	}
	if sawQueueURL != cfg.Config["queue_url"] {
		t.Fatalf("QueueUrl = %q, want configured queue URL", sawQueueURL)
	}
	if len(got) != 1 || got[0]["message_id"] != "m1" || got[0]["source"] != "checkout" || got[0]["body"] == nil {
		t.Fatalf("message not mapped: %+v", got)
	}
}

func TestReadStopsEarlyOnEmptyPoll(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`<ReceiveMessageResponse><ReceiveMessageResult></ReceiveMessageResult></ReceiveMessageResponse>`))
	}))
	defer srv.Close()

	c := native.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"queue_url": srv.URL + "/1/q", "region": "us-east-1", "max_polls": "5"},
		Secrets: map[string]string{"access_key": "a", "secret_key": "s"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (stop early on first empty poll)", calls)
	}
}

func TestWriteRejectsUnknownAction(t *testing.T) {
	c := native.New()
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("Write unknown action err = %v, want not found", err)
	}
}

func TestAPISurfaceCoversAllOfficialSQSOperations(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "..", "..", "defs", "amazon-sqs", "api_surface.json"))
	if err != nil {
		t.Fatalf("Read api_surface.json: %v", err)
	}
	var surface struct {
		Endpoints []struct {
			Method    string         `json:"method"`
			Path      string         `json:"path"`
			CoveredBy map[string]any `json:"covered_by"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("Unmarshal api_surface.json: %v", err)
	}
	want := []string{
		"SQS.AddPermission", "SQS.CancelMessageMoveTask", "SQS.ChangeMessageVisibility", "SQS.ChangeMessageVisibilityBatch",
		"SQS.CreateQueue", "SQS.DeleteMessage", "SQS.DeleteMessageBatch", "SQS.DeleteQueue", "SQS.GetQueueAttributes",
		"SQS.GetQueueUrl", "SQS.ListDeadLetterSourceQueues", "SQS.ListMessageMoveTasks", "SQS.ListQueues", "SQS.ListQueueTags",
		"SQS.PurgeQueue", "SQS.ReceiveMessage", "SQS.RemovePermission", "SQS.SendMessage", "SQS.SendMessageBatch",
		"SQS.SetQueueAttributes", "SQS.StartMessageMoveTask", "SQS.TagQueue", "SQS.UntagQueue",
	}
	got := map[string]int{}
	for _, ep := range surface.Endpoints {
		if ep.Method != http.MethodPost {
			t.Fatalf("endpoint %s method = %s, want POST", ep.Path, ep.Method)
		}
		if len(ep.CoveredBy) != 1 {
			t.Fatalf("endpoint %s covered_by = %#v, want exactly one implemented classifier", ep.Path, ep.CoveredBy)
		}
		got[ep.Path]++
	}
	if len(got) != len(want) {
		t.Fatalf("covered operation count = %d, want %d (%v)", len(got), len(want), got)
	}
	for _, op := range want {
		if got[op] != 1 {
			t.Fatalf("operation %s covered %d times, want once", op, got[op])
		}
	}
}

func TestManifestWriteActionsAndDestructiveConfirmations(t *testing.T) {
	c := native.New()
	manifest := c.Manifest()
	if len(manifest.WriteActions) != 16 {
		t.Fatalf("WriteActions = %d, want 16", len(manifest.WriteActions))
	}
	destructive := map[string]bool{
		"cancel_message_move_task": true,
		"delete_message":           true,
		"delete_message_batch":     true,
		"delete_queue":             true,
		"purge_queue":              true,
		"remove_permission":        true,
		"untag_queue":              true,
	}
	redacted := map[string][]string{
		"delete_message":           {"receipt_handle"},
		"send_message":             {"message_body", "message_attributes", "message_system_attributes"},
		"set_queue_attributes":     {"attribute_value", "attributes"},
		"cancel_message_move_task": {"task_handle"},
	}
	seen := map[string]bool{}
	for _, action := range manifest.WriteActions {
		seen[action.Name] = true
		if destructive[action.Name] && action.Confirm != "destructive" {
			t.Fatalf("action %s Confirm = %q, want destructive", action.Name, action.Confirm)
		}
		if action.Risk == "" || action.Method != http.MethodPost || !strings.HasPrefix(action.Path, "SQS.") {
			t.Fatalf("action %+v missing risk/method/path", action)
		}
		if want := redacted[action.Name]; len(want) > 0 && !hasAllStrings(action.RedactFields, want) {
			t.Fatalf("action %s RedactFields = %v, want at least %v", action.Name, action.RedactFields, want)
		}
	}
	for name := range destructive {
		if !seen[name] {
			t.Fatalf("destructive action %s not in manifest", name)
		}
	}
	if surface := c.CommandSurface(); surface == nil || len(surface.Commands) < 23 {
		t.Fatalf("CommandSurface commands = %v, want provider commands for all operations", surface)
	}
}

func TestOperationDirectReadListQueuesAndRedactsPolicy(t *testing.T) {
	var sawActions []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Fatalf("path = %q, want service root", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		sawActions = append(sawActions, r.Form.Get("Action"))
		switch r.Form.Get("Action") {
		case "ListQueues":
			_, _ = w.Write([]byte(`<ListQueuesResponse><ListQueuesResult><QueueUrl>https://sqs.us-east-1.amazonaws.com/123/orders</QueueUrl><NextToken>next</NextToken></ListQueuesResult></ListQueuesResponse>`))
		case "ListDeadLetterSourceQueues":
			_, _ = w.Write([]byte(`<ListDeadLetterSourceQueuesResponse><ListDeadLetterSourceQueuesResult><QueueUrl>https://sqs.us-east-1.amazonaws.com/123/orders-dlq</QueueUrl><NextToken>dead-letter-next</NextToken></ListDeadLetterSourceQueuesResult></ListDeadLetterSourceQueuesResponse>`))
		case "GetQueueAttributes":
			_, _ = w.Write([]byte(`<GetQueueAttributesResponse><GetQueueAttributesResult><Attribute><Name>Policy</Name><Value>{"Statement":"fixture"}</Value></Attribute><Attribute><Name>QueueArn</Name><Value>arn:aws:sqs:us-east-1:123:orders</Value></Attribute></GetQueueAttributesResult></GetQueueAttributesResponse>`))
		default:
			t.Fatalf("unexpected Action %q", r.Form.Get("Action"))
		}
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	res, err := c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{Operation: "list_queues", Config: cfg, Body: map[string]any{"max_results": 25}, RedactFields: []string{"", "policy"}})
	if err != nil {
		t.Fatalf("OperationDirectRead list_queues: %v", err)
	}
	body := res.Body.(map[string]any)
	if urls := body["queue_urls"].([]string); len(urls) != 1 || body["next_token"] != "next" {
		t.Fatalf("list_queues body = %#v", body)
	}
	res, err = c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{Operation: "list_dead_letter_source_queues", Config: cfg, Body: map[string]any{"max_results": 25}, RedactFields: []string{"policy"}})
	if err != nil {
		t.Fatalf("OperationDirectRead list_dead_letter_source_queues: %v", err)
	}
	body = res.Body.(map[string]any)
	if urls := body["queue_urls"].([]string); len(urls) != 1 || body["next_token"] != "dead-letter-next" {
		t.Fatalf("list_dead_letter_source_queues body = %#v", body)
	}
	res, err = c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{Operation: "get_queue_attributes", Config: cfg, RedactFields: []string{"policy"}})
	if err != nil {
		t.Fatalf("OperationDirectRead get_queue_attributes: %v", err)
	}
	attrs := res.Body.(map[string]any)["attributes"].(map[string]any)
	if attrs["Policy"] != "***" || attrs["QueueArn"] == "***" {
		t.Fatalf("attributes redaction = %#v", attrs)
	}
	if strings.Join(sawActions, ",") != "ListQueues,ListDeadLetterSourceQueues,GetQueueAttributes" {
		t.Fatalf("actions = %v", sawActions)
	}
}

func TestWriteSendMessageAndDeleteBatchChunking(t *testing.T) {
	var actions []string
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		actions = append(actions, r.Form.Get("Action"))
		bodies = append(bodies, r.Form.Encode())
		switch r.Form.Get("Action") {
		case "SendMessage":
			_, _ = w.Write([]byte(`<SendMessageResponse><SendMessageResult><MessageId>m1</MessageId></SendMessageResult></SendMessageResponse>`))
		case "DeleteMessageBatch":
			_, _ = w.Write([]byte(`<DeleteMessageBatchResponse><DeleteMessageBatchResult><Successful><Id>entry_1</Id></Successful></DeleteMessageBatchResult></DeleteMessageBatchResponse>`))
		default:
			t.Fatalf("unexpected Action %q", r.Form.Get("Action"))
		}
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	preview, err := c.DryRunWrite(context.Background(), connectors.WriteRequest{Action: "send_message", Config: cfg}, []connectors.Record{{"message_body": "hello", "message_group_id": "orders"}})
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	if preview.RecordsStaged != 1 || preview.Action != "send_message" {
		t.Fatalf("preview = %+v", preview)
	}
	res, err := c.Write(context.Background(), connectors.WriteRequest{Action: "send_message", Config: cfg}, []connectors.Record{{"message_body": "hello", "message_group_id": "orders"}})
	if err != nil {
		t.Fatalf("Write send_message: %v", err)
	}
	if res.RecordsWritten != 1 || !strings.Contains(bodies[0], "MessageBody=hello") || !strings.Contains(bodies[0], "MessageGroupId=orders") {
		t.Fatalf("send result=%+v body=%q", res, bodies[0])
	}

	records := make([]connectors.Record, 11)
	for i := range records {
		records[i] = connectors.Record{"receipt_handle": "rh"}
	}
	res, err = c.Write(context.Background(), connectors.WriteRequest{Action: "delete_message_batch", Config: cfg}, records)
	if err != nil {
		t.Fatalf("Write delete_message_batch: %v", err)
	}
	if res.RecordsWritten != 2 || countAction(actions, "DeleteMessageBatch") != 2 {
		t.Fatalf("delete batch result=%+v actions=%v", res, actions)
	}
	if !strings.Contains(bodies[1], "DeleteMessageBatchRequestEntry.10.ReceiptHandle=rh") || !strings.Contains(bodies[2], "DeleteMessageBatchRequestEntry.1.ReceiptHandle=rh") {
		t.Fatalf("batch bodies did not chunk at 10: %q / %q", bodies[1], bodies[2])
	}
}

func TestWriteNormalizesActionAndSendsQueueURL(t *testing.T) {
	var sawQueueURL string
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if action := r.Form.Get("Action"); action != "PurgeQueue" {
			t.Fatalf("Action = %q, want PurgeQueue", action)
		}
		sawQueueURL = r.Form.Get("QueueUrl")
		_, _ = w.Write([]byte(`<PurgeQueueResponse/>`))
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	preview, err := c.DryRunWrite(context.Background(), connectors.WriteRequest{Action: " purge_queue ", Config: cfg}, nil)
	if err != nil {
		t.Fatalf("DryRunWrite purge_queue: %v", err)
	}
	if preview.Action != "purge_queue" || preview.RecordsStaged != 0 || !strings.Contains(strings.Join(preview.Warnings, " "), "destructive") {
		t.Fatalf("preview = %+v, want normalized zero-staged destructive action", preview)
	}
	res, err := c.Write(context.Background(), connectors.WriteRequest{Action: " purge_queue ", Config: cfg}, nil)
	if err != nil {
		t.Fatalf("Write purge_queue empty: %v", err)
	}
	if res.RecordsWritten != 0 || res.RecordsFailed != 0 || calls != 0 {
		t.Fatalf("empty write result=%+v calls=%d, want no execution", res, calls)
	}
	res, err = c.Write(context.Background(), connectors.WriteRequest{Action: " purge_queue ", Config: cfg}, []connectors.Record{{}})
	if err != nil {
		t.Fatalf("Write purge_queue explicit record: %v", err)
	}
	if res.RecordsWritten != 1 || calls != 1 || sawQueueURL != cfg.Config["queue_url"] {
		t.Fatalf("result=%+v calls=%d QueueUrl=%q, want one write to configured queue", res, calls, sawQueueURL)
	}
}

func TestValidateWriteClosedSchemas(t *testing.T) {
	c := native.New()
	cfg := testRuntimeConfig("https://sqs.example.test")
	cases := []struct {
		name    string
		action  string
		records []connectors.Record
		want    string
	}{
		{
			name:    "unsupported field",
			action:  "send_message",
			records: []connectors.Record{{"message_body": "ok", "raw_action": "PurgeQueue"}},
			want:    "unsupported field",
		},
		{
			name:    "missing required field",
			action:  "send_message",
			records: []connectors.Record{{}},
			want:    "requires field",
		},
		{
			name:    "string field wrong type",
			action:  "send_message",
			records: []connectors.Record{{"message_body": 42}},
			want:    "must be string",
		},
		{
			name:    "integer field rejects string",
			action:  "send_message",
			records: []connectors.Record{{"message_body": "ok", "delay_seconds": "5"}},
			want:    "must be integer",
		},
		{
			name:    "integer field rejects fractional number",
			action:  "send_message",
			records: []connectors.Record{{"message_body": "ok", "delay_seconds": 1.5}},
			want:    "must be integer",
		},
		{
			name:    "integer field rejects out of range",
			action:  "send_message",
			records: []connectors.Record{{"message_body": "ok", "delay_seconds": 901}},
			want:    "must be between 0 and 900",
		},
		{
			name:    "array field rejects scalar",
			action:  "add_permission",
			records: []connectors.Record{{"label": "l", "aws_account_ids": "123", "actions": []string{"SendMessage"}}},
			want:    "must be array of strings",
		},
		{
			name:    "array field rejects non-string item",
			action:  "add_permission",
			records: []connectors.Record{{"label": "l", "aws_account_ids": []any{"123", 4}, "actions": []string{"SendMessage"}}},
			want:    "must be array of strings",
		},
		{
			name:    "map field rejects scalar",
			action:  "tag_queue",
			records: []connectors.Record{{"tag_key": "team", "tag_value": "platform", "tags": "team=platform"}},
			want:    "must be object with string values",
		},
		{
			name:    "map field rejects non-string value",
			action:  "tag_queue",
			records: []connectors.Record{{"tag_key": "team", "tag_value": "platform", "tags": map[string]any{"team": 7}}},
			want:    "must be object with string values",
		},
		{
			name:    "message attribute map rejects scalar",
			action:  "send_message",
			records: []connectors.Record{{"message_body": "ok", "message_attributes": "trace=abc"}},
			want:    "must be object",
		},
		{
			name:    "message attribute map rejects nested wrong type",
			action:  "send_message",
			records: []connectors.Record{{"message_body": "ok", "message_attributes": map[string]any{"trace": map[string]any{"string_value": 7}}}},
			want:    "field \"string_value\" must be string",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := c.ValidateWrite(context.Background(), connectors.WriteRequest{Action: tc.action, Config: cfg}, tc.records)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ValidateWrite err = %v, want %q", err, tc.want)
			}
		})
	}
	if err := c.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "purge_queue", Config: cfg}, nil); err != nil {
		t.Fatalf("ValidateWrite purge_queue empty record: %v", err)
	}
}

func TestWriteSerializesTagMapsWithKeyValue(t *testing.T) {
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		bodies = append(bodies, r.Form.Encode())
		switch r.Form.Get("Action") {
		case "CreateQueue":
			_, _ = w.Write([]byte(`<CreateQueueResponse><CreateQueueResult><QueueUrl>https://sqs.example.test/123/orders</QueueUrl></CreateQueueResult></CreateQueueResponse>`))
		case "TagQueue":
			_, _ = w.Write([]byte(`<TagQueueResponse/>`))
		default:
			t.Fatalf("unexpected Action %q", r.Form.Get("Action"))
		}
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Action: "create_queue", Config: cfg}, []connectors.Record{{"queue_name": "orders", "tags": map[string]any{"env": "prod"}}}); err != nil {
		t.Fatalf("Write create_queue: %v", err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Action: "tag_queue", Config: cfg}, []connectors.Record{{"tag_key": "team", "tag_value": "platform"}}); err != nil {
		t.Fatalf("Write tag_queue: %v", err)
	}
	if len(bodies) != 2 {
		t.Fatalf("bodies = %d, want 2", len(bodies))
	}
	if !strings.Contains(bodies[0], "Tag.1.Key=env") || !strings.Contains(bodies[0], "Tag.1.Value=prod") || strings.Contains(bodies[0], "Tag.1.Name") {
		t.Fatalf("create_queue tags encoded incorrectly: %q", bodies[0])
	}
	if !strings.Contains(bodies[1], "Tag.1.Key=team") || !strings.Contains(bodies[1], "Tag.1.Value=platform") || strings.Contains(bodies[1], "Tag.1.Name") {
		t.Fatalf("tag_queue tags encoded incorrectly: %q", bodies[1])
	}
}

func TestWriteBatchFailuresReturnError(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if action := r.Form.Get("Action"); action != "SendMessageBatch" {
			t.Fatalf("Action = %q, want SendMessageBatch", action)
		}
		_, _ = w.Write([]byte(`<SendMessageBatchResponse><SendMessageBatchResult><Successful><Id>entry_1</Id><MessageId>m1</MessageId></Successful><Failed><Id>entry_2</Id><SenderFault>true</SenderFault><Code>InvalidMessageContents</Code><Message>bad</Message></Failed></SendMessageBatchResult></SendMessageBatchResponse>`))
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	res, err := c.Write(context.Background(), connectors.WriteRequest{Action: "send_message_batch", Config: cfg}, []connectors.Record{{"message_body": "first"}, {"message_body": "second"}})
	if err == nil || !strings.Contains(err.Error(), "batch response reported 1 failed") {
		t.Fatalf("Write send_message_batch err = %v, want batch failure", err)
	}
	if res.RecordsWritten != 1 || res.RecordsFailed != 1 || calls != 1 {
		t.Fatalf("result=%+v calls=%d, want one success and one failure", res, calls)
	}
}

func testRuntimeConfig(endpoint string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"queue_url":    endpoint + "/123/orders",
			"endpoint_url": endpoint,
			"region":       "us-east-1",
		},
		Secrets: map[string]string{"access_key": "AKIATEST", "secret_key": "synthetic-signing-key"},
	}
}

func countAction(actions []string, want string) int {
	var count int
	for _, action := range actions {
		if action == want {
			count++
		}
	}
	return count
}

func hasAllStrings(values []string, wants []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, want := range wants {
		if !seen[want] {
			return false
		}
	}
	return true
}
