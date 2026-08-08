package amazonsqs_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
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

func TestOperationDirectReadPreflightUsesClosedSQSContract(t *testing.T) {
	c := native.New()
	tests := []struct {
		name    string
		op      string
		method  string
		path    string
		cap     int
		policy  string
		wantErr string
	}{
		{name: "valid", op: "list_queues", method: http.MethodPost, path: "SQS.ListQueues", cap: 16 << 20, policy: "json_redacted"},
		{name: "unknown operation", op: "raw_action", method: http.MethodPost, path: "SQS.RawAction", cap: 16 << 20, policy: "json_redacted", wantErr: "operation \"raw_action\" not found"},
		{name: "method mismatch", op: "list_queues", method: http.MethodGet, path: "SQS.ListQueues", cap: 16 << 20, policy: "json_redacted", wantErr: "does not match declared operation method"},
		{name: "path mismatch", op: "list_queues", method: http.MethodPost, path: "SQS.RawAction", cap: 16 << 20, policy: "json_redacted", wantErr: "does not match declared operation path"},
		{name: "missing cap", op: "list_queues", method: http.MethodPost, path: "SQS.ListQueues", cap: 0, policy: "json_redacted", wantErr: "requires positive max_bytes"},
		{name: "policy mismatch", op: "list_queues", method: http.MethodPost, path: "SQS.ListQueues", cap: 16 << 20, policy: "json", wantErr: "not supported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.PreflightOperationDirectRead(tt.op, tt.method, tt.path, tt.cap, tt.policy)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("PreflightOperationDirectRead error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("PreflightOperationDirectRead error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestOperationDirectReadPreflightFailsClosedWithoutRuntimeLedger(t *testing.T) {
	c := directReadConnectorWithoutRuntimeLedger()

	err := c.PreflightOperationDirectRead("list_queues", http.MethodPost, "SQS.ListQueues", 16<<20, "json_redacted")
	if err == nil || !strings.Contains(err.Error(), "runtime operation endpoint ledger is unavailable") {
		t.Fatalf("PreflightOperationDirectRead without runtime ledger = %v, want fail-closed ledger rejection", err)
	}
}

func TestOperationDirectReadFailsClosedWithoutRuntimeLedger(t *testing.T) {
	c := directReadConnectorWithoutRuntimeLedger()

	_, err := c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{Operation: "list_queues"})
	if err == nil || !strings.Contains(err.Error(), "runtime operation endpoint ledger is unavailable") {
		t.Fatalf("OperationDirectRead without runtime ledger = %v, want fail-closed ledger rejection", err)
	}
}

func directReadConnectorWithoutRuntimeLedger() native.Connector {
	return native.Connector{Base: engine.NewBase(engine.Bundle{
		Name: "amazon-sqs",
		Operations: []engine.OperationSpec{{
			ID:   "list_queues",
			Kind: "rest_read",
			REST: &engine.RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "SQS.ListQueues",
				ContentType: "application/json",
				MaxBytes:    1 << 20,
				BodySchema:  json.RawMessage(`{"type":"object","additionalProperties":false}`),
			},
		}},
	})}
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

func TestCheckReportsSanitizedAWSErrorCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`<ErrorResponse><Error><Code>AccessDenied</Code><Message>sensitive provider detail</Message></Error></ErrorResponse>`))
	}))
	defer srv.Close()

	err := native.New().Check(context.Background(), testRuntimeConfig(srv.URL))
	if err == nil {
		t.Fatal("Check: want provider error, got nil")
	}
	text := err.Error()
	if !strings.Contains(text, "status 403 Forbidden") || !strings.Contains(text, "AccessDenied") {
		t.Fatalf("Check error = %q, want safe status and AWS error code", text)
	}
	if strings.Contains(text, "sensitive provider detail") || strings.Contains(text, "<ErrorResponse>") {
		t.Fatalf("Check error exposed provider response body: %q", text)
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

func TestCLIReverseExamplesContainRequiredFlags(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "..", "..", "defs", "amazon-sqs", "cli_surface.json"))
	if err != nil {
		t.Fatalf("Read cli_surface.json: %v", err)
	}
	var surface struct {
		Commands []struct {
			Path     string   `json:"path"`
			Intent   string   `json:"intent"`
			Examples []string `json:"examples"`
			Flags    []struct {
				Name     string `json:"name"`
				Required bool   `json:"required"`
			} `json:"flags"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("Unmarshal cli_surface.json: %v", err)
	}
	want := map[string][]string{
		"permission add":                  {"label", "aws-account-ids", "actions"},
		"message-move-task cancel":        {"task-handle"},
		"message change-visibility":       {"receipt-handle", "visibility-timeout"},
		"message change-visibility-batch": {"receipt-handle", "visibility-timeout"},
		"queue create":                    {"queue-name"},
		"message delete":                  {"receipt-handle"},
		"message delete-batch":            {"receipt-handle"},
		"queue delete":                    {},
		"queue purge":                     {},
		"permission remove":               {"label"},
		"message send":                    {"message-body"},
		"message send-batch":              {"message-body"},
		"queue set-attributes":            {"attribute-name", "attribute-value"},
		"message-move-task start":         {"source-arn"},
		"queue tag":                       {"tag-key", "tag-value"},
		"queue untag":                     {"tag-keys"},
	}
	seen := map[string]bool{}
	for _, command := range surface.Commands {
		if command.Intent != "reverse_etl" {
			continue
		}
		required, ok := want[command.Path]
		if !ok {
			t.Fatalf("unexpected reverse command %q", command.Path)
		}
		seen[command.Path] = true
		if len(command.Examples) != 1 {
			t.Fatalf("command %q examples = %v, want one", command.Path, command.Examples)
		}
		example := command.Examples[0]
		for _, name := range required {
			var marked bool
			for _, flag := range command.Flags {
				if flag.Name == name {
					marked = flag.Required
				}
			}
			if !marked {
				t.Fatalf("command %q flag --%s is not required", command.Path, name)
			}
			if !strings.Contains(example, "--"+name+" ") {
				t.Fatalf("command %q example %q omits --%s value", command.Path, example, name)
			}
		}
	}
	for path := range want {
		if !seen[path] {
			t.Fatalf("missing reverse command %q", path)
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
		"create_queue":             {"attributes"},
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
		case "ListQueueTags":
			_, _ = w.Write([]byte(`<ListQueueTagsResponse><ListQueueTagsResult><Tag><Key>api_key</Key><Value>abc</Value></Tag><Tag><Key>access_key</Key><Value>def</Value></Tag><Tag><Key>credential_id</Key><Value>ghi</Value></Tag><Tag><Key>next_token</Key><Value>nested-secret-token</Value></Tag><Tag><Key>environment</Key><Value>prod</Value></Tag></ListQueueTagsResult></ListQueueTagsResponse>`))
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
	res, err = c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{Operation: "list_queue_tags", Config: cfg})
	if err != nil {
		t.Fatalf("OperationDirectRead list_queue_tags: %v", err)
	}
	tags := res.Body.(map[string]any)["tags"].(map[string]any)
	if tags["api_key"] != "***" || tags["access_key"] != "***" || tags["credential_id"] != "***" || tags["next_token"] != "***" || tags["environment"] != "prod" {
		t.Fatalf("tag redaction = %#v", tags)
	}
	if strings.Join(sawActions, ",") != "ListQueues,ListDeadLetterSourceQueues,GetQueueAttributes,ListQueueTags" {
		t.Fatalf("actions = %v", sawActions)
	}
}

func TestOperationDirectReadClampsResponseToLedgerLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", (1<<20)+1)))
	}))
	defer srv.Close()

	_, err := native.New().OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "list_queues",
		Config:    testRuntimeConfig(srv.URL),
		MaxBytes:  16 << 20,
	})
	if err == nil || !strings.Contains(err.Error(), "exceeds limit 1048576") {
		t.Fatalf("OperationDirectRead response above ledger cap = %v, want 1 MiB cap rejection", err)
	}
}

func TestOperationDirectReadGetQueueURL(t *testing.T) {
	const queueURL = "https://sqs.us-east-1.amazonaws.com/000000000000/fixture-queue"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if r.Form.Get("Action") != "GetQueueUrl" || r.Form.Get("QueueName") != "fixture-queue" || r.Form.Get("QueueOwnerAWSAccountId") != "000000000000" {
			t.Fatalf("unexpected GetQueueUrl form: %s", r.Form.Encode())
		}
		_, _ = w.Write([]byte(`<GetQueueUrlResponse><GetQueueUrlResult><QueueUrl>` + queueURL + `</QueueUrl></GetQueueUrlResult></GetQueueUrlResponse>`))
	}))
	defer srv.Close()

	res, err := native.New().OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "get_queue_url",
		Config:    testRuntimeConfig(srv.URL),
		Body:      map[string]any{"queue_name": "fixture-queue", "queue_owner_aws_account_id": "000000000000"},
	})
	if err != nil {
		t.Fatalf("OperationDirectRead get_queue_url: %v", err)
	}
	if res.Path != "SQS.GetQueueUrl" || res.Body.(map[string]any)["queue_url"] != queueURL {
		t.Fatalf("get_queue_url result = %+v", res)
	}
}

func TestOperationDirectReadListMessageMoveTasksDecodesResults(t *testing.T) {
	sourceArn := "arn:aws:sqs:us-east-1:123456789012:orders-dlq"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if action := r.Form.Get("Action"); action != "ListMessageMoveTasks" {
			t.Fatalf("Action = %q, want ListMessageMoveTasks", action)
		}
		if got := r.Form.Get("SourceArn"); got != sourceArn {
			t.Fatalf("SourceArn = %q, want %q", got, sourceArn)
		}
		_, _ = w.Write([]byte(`<ListMessageMoveTasksResponse xmlns="http://queue.amazonaws.com/doc/2012-11-05/"><ListMessageMoveTasksResult><Result><TaskHandle>task-handle-fixture</TaskHandle><Status>RUNNING</Status><SourceArn>arn:aws:sqs:us-east-1:123456789012:orders-dlq</SourceArn><DestinationArn>arn:aws:sqs:us-east-1:123456789012:orders</DestinationArn><MaxNumberOfMessagesPerSecond>10</MaxNumberOfMessagesPerSecond><ApproximateNumberOfMessagesMoved>42</ApproximateNumberOfMessagesMoved><ApproximateNumberOfMessagesToMove>100</ApproximateNumberOfMessagesToMove><StartedTimestamp>1767225600</StartedTimestamp></Result></ListMessageMoveTasksResult><ResponseMetadata><RequestId>synthetic-request</RequestId></ResponseMetadata></ListMessageMoveTasksResponse>`))
	}))
	defer srv.Close()

	c := native.New()
	res, err := c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{Operation: "list_message_move_tasks", Config: testRuntimeConfig(srv.URL), Body: map[string]any{"source_arn": sourceArn}, RedactFields: []string{"task_handle"}})
	if err != nil {
		t.Fatalf("OperationDirectRead list_message_move_tasks: %v", err)
	}
	results := res.Body.(map[string]any)["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results = %#v, want one decoded task", results)
	}
	task := results[0].(map[string]any)
	if task["task_handle"] != "***" || task["status"] != "RUNNING" || task["source_arn"] != sourceArn || task["approximate_number_of_messages_moved"] != "42" {
		t.Fatalf("task = %#v, want decoded redacted ListMessageMoveTasks result", task)
	}
}

func TestOperationDirectReadRejectsUntypedBodies(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		t.Fatalf("unexpected SQS request for invalid direct-read body")
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	cases := []struct {
		name      string
		operation string
		body      map[string]any
		want      string
	}{
		{name: "unknown field", operation: "list_queues", body: map[string]any{"raw_action": "ListQueues"}, want: "unsupported field"},
		{name: "string max results", operation: "list_queues", body: map[string]any{"max_results": "25"}, want: "must be integer"},
		{name: "clamped low max results", operation: "list_queues", body: map[string]any{"max_results": 0}, want: "must be between 1 and 1000"},
		{name: "clamped high max results", operation: "list_message_move_tasks", body: map[string]any{"source_arn": "arn:aws:sqs:us-east-1:123456789012:orders-dlq", "max_results": 11}, want: "must be between 1 and 10"},
		{name: "string field wrong type", operation: "get_queue_url", body: map[string]any{"queue_name": 123}, want: "must be string"},
		{name: "array field wrong type", operation: "get_queue_attributes", body: map[string]any{"attribute_names": "All"}, want: "must be array of strings"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{Operation: tc.operation, Config: cfg, Body: tc.body})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("OperationDirectRead err = %v, want %q", err, tc.want)
			}
		})
	}
	if calls != 0 {
		t.Fatalf("invalid direct-read bodies made %d SQS requests, want 0", calls)
	}
}

func TestWriteExecutesPreviouslyUncoveredTypedActions(t *testing.T) {
	tests := []struct {
		action    string
		awsAction string
		record    connectors.Record
		wantForm  map[string]string
	}{
		{action: "add_permission", awsAction: "AddPermission", record: connectors.Record{"label": "fixture-permission", "aws_account_ids": []string{"000000000000"}, "actions": []string{"SendMessage"}}, wantForm: map[string]string{"Label": "fixture-permission", "AWSAccountId.1": "000000000000", "ActionName.1": "SendMessage"}},
		{action: "cancel_message_move_task", awsAction: "CancelMessageMoveTask", record: connectors.Record{"task_handle": "task-handle-placeholder"}, wantForm: map[string]string{"TaskHandle": "task-handle-placeholder"}},
		{action: "change_message_visibility", awsAction: "ChangeMessageVisibility", record: connectors.Record{"receipt_handle": "receipt-handle-placeholder", "visibility_timeout": 30}, wantForm: map[string]string{"ReceiptHandle": "receipt-handle-placeholder", "VisibilityTimeout": "30"}},
		{action: "delete_message", awsAction: "DeleteMessage", record: connectors.Record{"receipt_handle": "receipt-handle-placeholder"}, wantForm: map[string]string{"ReceiptHandle": "receipt-handle-placeholder"}},
		{action: "delete_queue", awsAction: "DeleteQueue", record: connectors.Record{}, wantForm: map[string]string{}},
		{action: "remove_permission", awsAction: "RemovePermission", record: connectors.Record{"label": "fixture-permission"}, wantForm: map[string]string{"Label": "fixture-permission"}},
		{action: "set_queue_attributes", awsAction: "SetQueueAttributes", record: connectors.Record{"attribute_name": "VisibilityTimeout", "attribute_value": "30"}, wantForm: map[string]string{"Attribute.1.Name": "VisibilityTimeout", "Attribute.1.Value": "30"}},
		{action: "start_message_move_task", awsAction: "StartMessageMoveTask", record: connectors.Record{"source_arn": "arn:aws:sqs:us-east-1:000000000000:fixture-dlq"}, wantForm: map[string]string{"SourceArn": "arn:aws:sqs:us-east-1:000000000000:fixture-dlq"}},
		{action: "untag_queue", awsAction: "UntagQueue", record: connectors.Record{"tag_keys": []string{"environment"}}, wantForm: map[string]string{"TagKey.1": "environment"}},
	}
	for _, tc := range tests {
		t.Run(tc.action, func(t *testing.T) {
			var calls int
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if err := r.ParseForm(); err != nil {
					t.Fatalf("ParseForm: %v", err)
				}
				if got := r.Form.Get("Action"); got != tc.awsAction {
					t.Fatalf("Action = %q, want %q", got, tc.awsAction)
				}
				for name, want := range tc.wantForm {
					if got := r.Form.Get(name); got != want {
						t.Fatalf("%s = %q, want %q; form=%s", name, got, want, r.Form.Encode())
					}
				}
				_, _ = w.Write([]byte(`<Response/>`))
			}))
			defer srv.Close()

			connector := native.New()
			records := []connectors.Record{tc.record}
			res, err := connector.Write(context.Background(), approvedSQSWriteRequest(t, connector, tc.action, testRuntimeConfig(srv.URL), records), records)
			if err != nil {
				t.Fatalf("Write %s: %v", tc.action, err)
			}
			if res.RecordsWritten != 1 || res.RecordsFailed != 0 || calls != 1 {
				t.Fatalf("result=%+v calls=%d, want one successful request", res, calls)
			}
		})
	}
}

func TestApprovedDestructiveWriteRefusesRedirectToUnapprovedTarget(t *testing.T) {
	var initialCalls atomic.Int32
	var redirectedCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/unapproved" {
			redirectedCalls.Add(1)
			_, _ = w.Write([]byte(`<Response/>`))
			return
		}
		initialCalls.Add(1)
		w.Header().Set("Location", "/unapproved")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer srv.Close()

	connector := native.New()
	records := []connectors.Record{{}}
	_, err := connector.Write(context.Background(), approvedSQSWriteRequest(t, connector, "delete_queue", testRuntimeConfig(srv.URL), records), records)
	if err == nil || !strings.Contains(err.Error(), "redirect refused") {
		t.Fatalf("Write delete_queue err = %v, want redirect refusal", err)
	}
	if got := initialCalls.Load(); got != 1 {
		t.Fatalf("initial requests = %d, want 1", got)
	}
	if got := redirectedCalls.Load(); got != 0 {
		t.Fatalf("redirected requests = %d, want 0", got)
	}
}

func TestWriteSendMessageAndDeleteBatchChunking(t *testing.T) {
	var actions []string
	var bodies []string
	var sentMessageBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		actions = append(actions, r.Form.Get("Action"))
		bodies = append(bodies, r.Form.Encode())
		switch r.Form.Get("Action") {
		case "SendMessage":
			sentMessageBody = r.Form.Get("MessageBody")
			_, _ = w.Write([]byte(`<SendMessageResponse><SendMessageResult><MessageId>m1</MessageId></SendMessageResult></SendMessageResponse>`))
		case "DeleteMessageBatch":
			entries := 0
			for key := range r.Form {
				if strings.HasPrefix(key, "DeleteMessageBatchRequestEntry.") && strings.HasSuffix(key, ".ReceiptHandle") {
					entries++
				}
			}
			var response strings.Builder
			response.WriteString(`<DeleteMessageBatchResponse><DeleteMessageBatchResult>`)
			for i := 1; i <= entries; i++ {
				response.WriteString(`<DeleteMessageBatchResultEntry><Id>`)
				response.WriteString(r.Form.Get("DeleteMessageBatchRequestEntry." + strconv.Itoa(i) + ".Id"))
				response.WriteString(`</Id></DeleteMessageBatchResultEntry>`)
			}
			response.WriteString(`</DeleteMessageBatchResult></DeleteMessageBatchResponse>`)
			_, _ = w.Write([]byte(response.String()))
		default:
			t.Fatalf("unexpected Action %q", r.Form.Get("Action"))
		}
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	preview, err := c.DryRunWrite(context.Background(), connectors.WriteRequest{Action: "send_message", Config: cfg}, []connectors.Record{{"message_body": "  hello  ", "message_group_id": "orders"}})
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	if preview.RecordsStaged != 1 || preview.Action != "send_message" {
		t.Fatalf("preview = %+v", preview)
	}
	res, err := c.Write(context.Background(), connectors.WriteRequest{Action: "send_message", Config: cfg}, []connectors.Record{{"message_body": "  hello  ", "message_group_id": "orders"}})
	if err != nil {
		t.Fatalf("Write send_message: %v", err)
	}
	if res.RecordsWritten != 1 || sentMessageBody != "  hello  " || !strings.Contains(bodies[0], "MessageGroupId=orders") {
		t.Fatalf("send result=%+v message_body=%q body=%q", res, sentMessageBody, bodies[0])
	}

	records := make([]connectors.Record, 11)
	for i := range records {
		records[i] = connectors.Record{"receipt_handle": "rh"}
	}
	res, err = c.Write(context.Background(), approvedSQSWriteRequest(t, c, "delete_message_batch", cfg, records), records)
	if err != nil {
		t.Fatalf("Write delete_message_batch: %v", err)
	}
	if res.RecordsWritten != 11 || countAction(actions, "DeleteMessageBatch") != 2 {
		t.Fatalf("delete batch result=%+v actions=%v", res, actions)
	}
	if !strings.Contains(bodies[1], "DeleteMessageBatchRequestEntry.10.ReceiptHandle=rh") || !strings.Contains(bodies[2], "DeleteMessageBatchRequestEntry.1.ReceiptHandle=rh") {
		t.Fatalf("batch bodies did not chunk at 10: %q / %q", bodies[1], bodies[2])
	}
}

func TestWriteChangeMessageVisibilityBatchRequiresTimeout(t *testing.T) {
	var calls int
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		body = r.Form.Encode()
		if action := r.Form.Get("Action"); action != "ChangeMessageVisibilityBatch" {
			t.Fatalf("Action = %q, want ChangeMessageVisibilityBatch", action)
		}
		_, _ = w.Write([]byte(`<ChangeMessageVisibilityBatchResponse><ChangeMessageVisibilityBatchResult><ChangeMessageVisibilityBatchResultEntry><Id>entry_1</Id></ChangeMessageVisibilityBatchResultEntry></ChangeMessageVisibilityBatchResult></ChangeMessageVisibilityBatchResponse>`))
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	res, err := c.Write(context.Background(), connectors.WriteRequest{Action: "change_message_visibility_batch", Config: cfg}, []connectors.Record{{"receipt_handle": "rh"}})
	if err == nil || !strings.Contains(err.Error(), "requires field \"visibility_timeout\"") {
		t.Fatalf("Write change_message_visibility_batch err = %v, want missing visibility_timeout", err)
	}
	if res.RecordsWritten != 0 || res.RecordsFailed != 1 || calls != 0 {
		t.Fatalf("invalid write result=%+v calls=%d, want validation failure before request", res, calls)
	}

	res, err = c.Write(context.Background(), connectors.WriteRequest{Action: "change_message_visibility_batch", Config: cfg}, []connectors.Record{{"receipt_handle": "rh", "visibility_timeout": 45}})
	if err != nil {
		t.Fatalf("Write change_message_visibility_batch: %v", err)
	}
	if res.RecordsWritten != 1 || res.RecordsFailed != 0 || calls != 1 {
		t.Fatalf("valid write result=%+v calls=%d, want one batch request", res, calls)
	}
	if !strings.Contains(body, "ChangeMessageVisibilityBatchRequestEntry.1.VisibilityTimeout=45") {
		t.Fatalf("batch body %q missing required visibility timeout", body)
	}
}

func TestWriteAllowsWhitespaceOnlyMessageBody(t *testing.T) {
	var sentMessageBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if action := r.Form.Get("Action"); action != "SendMessage" {
			t.Fatalf("Action = %q, want SendMessage", action)
		}
		sentMessageBody = r.Form.Get("MessageBody")
		_, _ = w.Write([]byte(`<SendMessageResponse><SendMessageResult><MessageId>m1</MessageId></SendMessageResult></SendMessageResponse>`))
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	res, err := c.Write(context.Background(), connectors.WriteRequest{Action: "send_message", Config: cfg}, []connectors.Record{{"message_body": "   "}})
	if err != nil {
		t.Fatalf("Write send_message: %v", err)
	}
	if res.RecordsWritten != 1 || sentMessageBody != "   " {
		t.Fatalf("result=%+v message_body=%q, want whitespace payload sent unchanged", res, sentMessageBody)
	}
}

func TestWritePreservesMessageAttributeWhitespace(t *testing.T) {
	var forms []url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		forms = append(forms, cloneValues(r.Form))
		switch r.Form.Get("Action") {
		case "SendMessage":
			_, _ = w.Write([]byte(`<SendMessageResponse><SendMessageResult><MessageId>m1</MessageId></SendMessageResult></SendMessageResponse>`))
		case "SendMessageBatch":
			_, _ = w.Write([]byte(`<SendMessageBatchResponse><SendMessageBatchResult><SendMessageBatchResultEntry><Id>entry_1</Id><MessageId>m1</MessageId></SendMessageBatchResultEntry></SendMessageBatchResult></SendMessageBatchResponse>`))
		default:
			t.Fatalf("unexpected Action %q", r.Form.Get("Action"))
		}
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	attrs := map[string]any{
		"nested": map[string]any{
			"data_type":    "String.custom",
			"string_value": " x ",
		},
		"trace": " v ",
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Action: "send_message", Config: cfg}, []connectors.Record{{"message_body": "body", "message_attributes": attrs}}); err != nil {
		t.Fatalf("Write send_message: %v", err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Action: "send_message_batch", Config: cfg}, []connectors.Record{{"message_body": "body", "message_attributes": map[string]any{"trace": " v "}}}); err != nil {
		t.Fatalf("Write send_message_batch: %v", err)
	}
	if len(forms) != 2 {
		t.Fatalf("forms = %d, want 2", len(forms))
	}
	if got := forms[0].Get("MessageAttribute.1.Value.StringValue"); got != " x " {
		t.Fatalf("nested string attribute = %q, want whitespace preserved", got)
	}
	if got := forms[0].Get("MessageAttribute.2.Value.StringValue"); got != " v " {
		t.Fatalf("plain string attribute = %q, want whitespace preserved", got)
	}
	if got := forms[1].Get("SendMessageBatchRequestEntry.1.MessageAttribute.1.Value.StringValue"); got != " v " {
		t.Fatalf("batch string attribute = %q, want whitespace preserved", got)
	}
}

func TestWriteWithNoSourceRowsDoesNotExecuteDestructiveZeroFieldAction(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		t.Fatalf("unexpected SQS request for empty source rows: %s", r.URL.String())
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	preview, err := c.DryRunWrite(context.Background(), connectors.WriteRequest{Action: "purge_queue", Config: cfg}, nil)
	if err != nil {
		t.Fatalf("DryRunWrite purge_queue empty rows: %v", err)
	}
	if preview.RecordsStaged != 0 {
		t.Fatalf("preview.RecordsStaged = %d, want 0", preview.RecordsStaged)
	}
	res, err := c.Write(context.Background(), approvedSQSWriteRequest(t, c, "purge_queue", cfg, nil), nil)
	if err != nil {
		t.Fatalf("Write purge_queue empty rows: %v", err)
	}
	if res.RecordsWritten != 0 || res.RecordsFailed != 0 || calls != 0 {
		t.Fatalf("result=%+v calls=%d, want no-op", res, calls)
	}
}

func TestDestructiveWriteUsesPreparedPreviewAndSharedGate(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_, _ = w.Write([]byte(`<DeleteMessageResponse/>`))
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	records := []connectors.Record{{"receipt_handle": "receipt-handle-placeholder"}}
	preview, err := c.DryRunWrite(context.Background(), connectors.WriteRequest{Action: "delete_message", Config: cfg}, records)
	if err != nil {
		t.Fatalf("DryRunWrite(delete_message) error = %v", err)
	}
	if preview.Digest == "" {
		t.Error("destructive SQS preview returned no digest")
	}

	if _, err := c.Write(context.Background(), connectors.WriteRequest{Action: "delete_message", Config: cfg}, records); err == nil {
		t.Error("destructive SQS write bypassed shared approval gate")
	}
	if calls != 0 {
		t.Fatalf("destructive SQS request dispatched without approval; calls=%d", calls)
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
	res, err := c.Write(context.Background(), approvedSQSWriteRequest(t, c, " purge_queue ", cfg, nil), nil)
	if err != nil {
		t.Fatalf("Write purge_queue empty: %v", err)
	}
	if res.RecordsWritten != 0 || res.RecordsFailed != 0 || calls != 0 {
		t.Fatalf("empty write result=%+v calls=%d, want no execution", res, calls)
	}
	records := []connectors.Record{{}}
	res, err = c.Write(context.Background(), approvedSQSWriteRequest(t, c, " purge_queue ", cfg, records), records)
	if err != nil {
		t.Fatalf("Write purge_queue explicit record: %v", err)
	}
	if res.RecordsWritten != 1 || calls != 1 || sawQueueURL != cfg.Config["queue_url"] {
		t.Fatalf("result=%+v calls=%d QueueUrl=%q, want one write to configured queue", res, calls, sawQueueURL)
	}
}

func TestDryRunRejectsInvalidMessageAttributes(t *testing.T) {
	c := native.New()
	cfg := testRuntimeConfig("https://sqs.example.test")
	tests := []struct {
		name   string
		field  string
		values map[string]any
		want   string
	}{
		{name: "empty shorthand value", field: "message_attributes", values: map[string]any{"trace": ""}, want: "value must not be empty"},
		{name: "missing data type", field: "message_attributes", values: map[string]any{"trace": map[string]any{"string_value": "value"}}, want: "data_type must not be empty"},
		{name: "missing scalar value", field: "message_attributes", values: map[string]any{"trace": map[string]any{"data_type": "String"}}, want: "requires non-empty string_value"},
		{name: "wrong scalar for String", field: "message_attributes", values: map[string]any{"trace": map[string]any{"data_type": "String", "binary_value": "value"}}, want: "requires non-empty string_value"},
		{name: "wrong scalar for Binary", field: "message_attributes", values: map[string]any{"trace": map[string]any{"data_type": "Binary", "string_value": "value"}}, want: "requires non-empty binary_value"},
		{name: "reserved string list", field: "message_attributes", values: map[string]any{"trace": map[string]any{"data_type": "String", "string_value": "value", "string_list_values": []any{"reserved"}}}, want: "reserved and unsupported"},
		{name: "reserved provider name", field: "message_attributes", values: map[string]any{"AWS.trace": "value"}, want: "invalid name"},
		{name: "invalid number syntax", field: "message_attributes", values: map[string]any{"quantity": map[string]any{"data_type": "Number", "string_value": "not-a-number"}}, want: "valid SQS number"},
		{name: "number exceeds precision", field: "message_attributes", values: map[string]any{"quantity": map[string]any{"data_type": "Number", "string_value": "123456789012345678901234567890123456789"}}, want: "at most 38 digits of precision"},
		{name: "number below range", field: "message_attributes", values: map[string]any{"quantity": map[string]any{"data_type": "Number", "string_value": "9e-129"}}, want: "between 1e-128 and 1e126"},
		{name: "number above range", field: "message_attributes", values: map[string]any{"quantity": map[string]any{"data_type": "Number", "string_value": "1.1e126"}}, want: "between 1e-128 and 1e126"},
		{name: "malformed binary base64", field: "message_attributes", values: map[string]any{"payload": map[string]any{"data_type": "Binary.fixture", "binary_value": "not-base64!"}}, want: "standard base64"},
		{name: "empty decoded binary", field: "message_attributes", values: map[string]any{"payload": map[string]any{"data_type": "Binary.fixture", "binary_value": "\n"}}, want: "decode to at least one byte"},
		{name: "unsupported system name", field: "message_system_attributes", values: map[string]any{"trace": "value"}, want: "only supports AWSTraceHeader"},
		{name: "invalid system type", field: "message_system_attributes", values: map[string]any{"AWSTraceHeader": map[string]any{"data_type": "String.custom", "string_value": "value"}}, want: "must use data_type String"},
		{name: "malformed trace header", field: "message_system_attributes", values: map[string]any{"AWSTraceHeader": "value"}, want: "valid AWS X-Ray trace header"},
		{name: "malformed trace parent", field: "message_system_attributes", values: map[string]any{"AWSTraceHeader": map[string]any{"data_type": "String", "string_value": "Root=1-5759e988-bd862e3fe1be46a994272793;Parent=invalid;Sampled=1"}}, want: "valid AWS X-Ray trace header"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.DryRunWrite(context.Background(), connectors.WriteRequest{Action: "send_message", Config: cfg}, []connectors.Record{{"message_body": "hello", tc.field: tc.values}})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("DryRunWrite err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestDryRunAcceptsValidSQSMessageAttributeValues(t *testing.T) {
	attributes := map[string]any{
		"minimum": map[string]any{"data_type": "Number.minimum", "string_value": "1e-128"},
		"maximum": map[string]any{"data_type": "Number.maximum", "string_value": "1e126"},
		"precise": map[string]any{"data_type": "Number", "string_value": "0.12345678901234567890123456789012345678"},
		"payload": map[string]any{"data_type": "Binary.fixture", "binary_value": "AQIDBA=="},
	}
	systemAttributes := map[string]any{
		"AWSTraceHeader": map[string]any{"data_type": "String", "string_value": "Root=1-5759e988-bd862e3fe1be46a994272793;Parent=53995c3f42cd8ad8;Sampled=1"},
	}
	preview, err := native.New().DryRunWrite(context.Background(), connectors.WriteRequest{Action: "send_message", Config: testRuntimeConfig("https://sqs.example.test")}, []connectors.Record{{
		"message_body":              "hello",
		"message_attributes":        attributes,
		"message_system_attributes": systemAttributes,
	}})
	if err != nil {
		t.Fatalf("DryRunWrite valid message attributes: %v", err)
	}
	if preview.RecordsStaged != 1 {
		t.Fatalf("preview = %+v, want one staged record", preview)
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
			name:    "change visibility batch missing timeout",
			action:  "change_message_visibility_batch",
			records: []connectors.Record{{"receipt_handle": "rh"}},
			want:    "requires field \"visibility_timeout\"",
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
			name:    "set queue attributes rejects partial scalar with map",
			action:  "set_queue_attributes",
			records: []connectors.Record{{"attributes": map[string]any{"Policy": "{}"}, "attribute_name": "VisibilityTimeout"}},
			want:    "requires fields \"attribute_name\" + \"attribute_value\" together",
		},
		{
			name:    "tag queue rejects partial scalar with map",
			action:  "tag_queue",
			records: []connectors.Record{{"tags": map[string]any{"team": "platform"}, "tag_key": "env"}},
			want:    "requires fields \"tag_key\" + \"tag_value\" together",
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
	validMapShapes := []struct {
		name    string
		action  string
		records []connectors.Record
	}{
		{name: "set queue attributes map", action: "set_queue_attributes", records: []connectors.Record{{"attributes": map[string]any{"Policy": "{}"}}}},
		{name: "tag queue tags map", action: "tag_queue", records: []connectors.Record{{"tags": map[string]any{"team": "platform"}}}},
	}
	for _, tc := range validMapShapes {
		t.Run(tc.name, func(t *testing.T) {
			if err := c.ValidateWrite(context.Background(), connectors.WriteRequest{Action: tc.action, Config: cfg}, tc.records); err != nil {
				t.Fatalf("ValidateWrite err = %v, want nil", err)
			}
		})
	}
	if err := c.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "purge_queue", Config: cfg}, nil); err != nil {
		t.Fatalf("ValidateWrite purge_queue empty record: %v", err)
	}
}

func TestWriteSchemaRequiresChangeVisibilityBatchTimeout(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "..", "..", "defs", "amazon-sqs", "writes.json"))
	if err != nil {
		t.Fatalf("Read writes.json: %v", err)
	}
	var doc struct {
		Actions []struct {
			Name         string `json:"name"`
			RecordSchema struct {
				Required []string `json:"required"`
			} `json:"record_schema"`
		} `json:"actions"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("Unmarshal writes.json: %v", err)
	}
	for _, action := range doc.Actions {
		if action.Name != "change_message_visibility_batch" {
			continue
		}
		want := []string{"receipt_handle", "visibility_timeout"}
		if len(action.RecordSchema.Required) != len(want) || !hasAllStrings(action.RecordSchema.Required, want) {
			t.Fatalf("change_message_visibility_batch required = %v, want %v", action.RecordSchema.Required, want)
		}
		return
	}
	t.Fatal("writes.json missing change_message_visibility_batch")
}

func TestWriteSchemasAllowMapOnlySetAttributesAndTags(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "..", "..", "defs", "amazon-sqs", "writes.json"))
	if err != nil {
		t.Fatalf("Read writes.json: %v", err)
	}
	var doc struct {
		Actions []struct {
			Name         string `json:"name"`
			RecordSchema struct {
				Required      []string                   `json:"required"`
				MinProperties int                        `json:"minProperties"`
				Properties    map[string]json.RawMessage `json:"properties"`
			} `json:"record_schema"`
		} `json:"actions"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("Unmarshal writes.json: %v", err)
	}
	actions := map[string]struct {
		required []string
		mapField string
	}{
		"set_queue_attributes": {required: []string{"attribute_name", "attribute_value"}, mapField: "attributes"},
		"tag_queue":            {required: []string{"tag_key", "tag_value"}, mapField: "tags"},
	}
	seen := map[string]bool{}
	for _, action := range doc.Actions {
		want, ok := actions[action.Name]
		if !ok {
			continue
		}
		seen[action.Name] = true
		if len(action.RecordSchema.Required) != 0 {
			t.Fatalf("action %s required = %v, want no single required shape", action.Name, action.RecordSchema.Required)
		}
		if action.RecordSchema.MinProperties != 1 {
			t.Fatalf("action %s minProperties = %d, want 1", action.Name, action.RecordSchema.MinProperties)
		}
		if _, ok := action.RecordSchema.Properties[want.mapField]; !ok {
			t.Fatalf("action %s properties missing %q", action.Name, want.mapField)
		}
		for _, required := range want.required {
			if _, ok := action.RecordSchema.Properties[required]; !ok {
				t.Fatalf("action %s properties missing legacy field %q", action.Name, required)
			}
		}
	}
	for name := range actions {
		if !seen[name] {
			t.Fatalf("writes.json missing action %s", name)
		}
	}
}

func TestAmazonSQSGeneratedDocsDescribeMapOnlyWrites(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	docsDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "docs", "connectors", "amazon-sqs")
	for _, name := range []string{"MANUAL.md", "SKILL.md"} {
		raw, err := os.ReadFile(filepath.Join(docsDir, name))
		if err != nil {
			t.Fatalf("Read %s: %v", name, err)
		}
		text := string(raw)
		for _, stale := range []string{"required fields: attribute_name, attribute_value", "required fields: tag_key, tag_value"} {
			if strings.Contains(text, stale) {
				t.Fatalf("%s contains stale scalar-only requirement %q", name, stale)
			}
		}
		for _, want := range []string{"required fields: attributes or attribute_name + attribute_value", "required fields: tags or tag_key + tag_value"} {
			if !strings.Contains(text, want) {
				t.Fatalf("%s missing %q", name, want)
			}
		}
		if strings.Contains(text, "required fields: receipt_handle\n    optional fields: id, visibility_timeout") || strings.Contains(text, "required fields: receipt_handle\n  - optional fields: id, visibility_timeout") {
			t.Fatalf("%s contains stale change_message_visibility_batch requirement", name)
		}
		if !strings.Contains(text, "required fields: receipt_handle, visibility_timeout") || !strings.Contains(text, "optional fields: id") {
			t.Fatalf("%s missing change_message_visibility_batch timeout requirement", name)
		}
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
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Action: "tag_queue", Config: cfg}, []connectors.Record{{"tags": map[string]any{"team": "platform"}}}); err != nil {
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

func TestWriteBatchResponsesMustAccountForEntries(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		wantErr     string
		wantWritten int
		wantFailed  int
	}{
		{
			name:        "reported failure",
			response:    `<SendMessageBatchResponse><SendMessageBatchResult><SendMessageBatchResultEntry><Id>entry_1</Id><MessageId>m1</MessageId></SendMessageBatchResultEntry><BatchResultErrorEntry><Id>entry_2</Id><SenderFault>true</SenderFault><Code>InvalidMessageContents</Code><Message>bad</Message></BatchResultErrorEntry></SendMessageBatchResult></SendMessageBatchResponse>`,
			wantErr:     "batch response reported 1 failed",
			wantWritten: 1,
			wantFailed:  1,
		},
		{
			name:       "malformed xml",
			response:   `<SendMessageBatchResponse><SendMessageBatchResult>`,
			wantErr:    "batch response parse failed",
			wantFailed: 2,
		},
		{
			name:       "unrecognized response",
			response:   `<ErrorResponse><Error><Code>BadResponse</Code></Error></ErrorResponse>`,
			wantErr:    "accounted for 0 of 2 entries",
			wantFailed: 2,
		},
		{
			name:       "missing entry result",
			response:   `<SendMessageBatchResponse><SendMessageBatchResult><SendMessageBatchResultEntry><Id>entry_1</Id><MessageId>m1</MessageId></SendMessageBatchResultEntry></SendMessageBatchResult></SendMessageBatchResponse>`,
			wantErr:    "accounted for 1 of 2 entries",
			wantFailed: 2,
		},
		{
			name:       "duplicate success id",
			response:   `<SendMessageBatchResponse><SendMessageBatchResult><SendMessageBatchResultEntry><Id>entry_1</Id><MessageId>m1</MessageId></SendMessageBatchResultEntry><SendMessageBatchResultEntry><Id>entry_1</Id><MessageId>m2</MessageId></SendMessageBatchResultEntry></SendMessageBatchResult></SendMessageBatchResponse>`,
			wantErr:    "duplicate batch response id",
			wantFailed: 2,
		},
		{
			name:       "unknown success id",
			response:   `<SendMessageBatchResponse><SendMessageBatchResult><SendMessageBatchResultEntry><Id>entry_1</Id><MessageId>m1</MessageId></SendMessageBatchResultEntry><SendMessageBatchResultEntry><Id>entry_99</Id><MessageId>m2</MessageId></SendMessageBatchResultEntry></SendMessageBatchResult></SendMessageBatchResponse>`,
			wantErr:    "unknown batch response id",
			wantFailed: 2,
		},
		{
			name:       "missing result id",
			response:   `<SendMessageBatchResponse><SendMessageBatchResult><SendMessageBatchResultEntry><MessageId>m1</MessageId></SendMessageBatchResultEntry><SendMessageBatchResultEntry><Id>entry_2</Id><MessageId>m2</MessageId></SendMessageBatchResultEntry></SendMessageBatchResult></SendMessageBatchResponse>`,
			wantErr:    "batch response parse failed",
			wantFailed: 2,
		},
		{
			name:       "wrapper only results",
			response:   `<SendMessageBatchResponse><SendMessageBatchResult><Successful><Id>entry_1</Id></Successful><Successful><Id>entry_2</Id></Successful></SendMessageBatchResult></SendMessageBatchResponse>`,
			wantErr:    "accounted for 0 of 2 entries",
			wantFailed: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if err := r.ParseForm(); err != nil {
					t.Fatalf("ParseForm: %v", err)
				}
				if action := r.Form.Get("Action"); action != "SendMessageBatch" {
					t.Fatalf("Action = %q, want SendMessageBatch", action)
				}
				_, _ = w.Write([]byte(tt.response))
			}))
			defer srv.Close()

			c := native.New()
			cfg := testRuntimeConfig(srv.URL)
			res, err := c.Write(context.Background(), connectors.WriteRequest{Action: "send_message_batch", Config: cfg}, []connectors.Record{{"message_body": "first"}, {"message_body": "second"}})
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Write send_message_batch err = %v, want %q", err, tt.wantErr)
			}
			if res.RecordsWritten != tt.wantWritten || res.RecordsFailed != tt.wantFailed || calls != 1 {
				t.Fatalf("result=%+v calls=%d, want written=%d failed=%d", res, calls, tt.wantWritten, tt.wantFailed)
			}
		})
	}
}

func testRuntimeConfig(endpoint string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"queue_url":    endpoint + "/123/orders",
			"endpoint_url": endpoint,
			"region":       "us-east-1",
		},
		Secrets:             map[string]string{"access_key": "AKIATEST", "secret_key": "synthetic-signing-key"},
		CredentialRevision:  "amazon-sqs-fixture-credential-revision",
		ConfigurationDigest: "amazon-sqs-fixture-configuration-digest",
		WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
	}
}

func approvedSQSWriteRequest(t *testing.T, connector native.Connector, action string, cfg connectors.RuntimeConfig, records []connectors.Record) connectors.WriteRequest {
	t.Helper()
	var destructive bool
	for _, declared := range connector.Manifest().WriteActions {
		if declared.Name == strings.TrimSpace(action) {
			destructive = connectors.ConfirmationForWriteAction(declared).Kind == connectors.ConfirmationKindDestructive
			break
		}
	}
	if !destructive {
		return connectors.WriteRequest{Action: action, Config: cfg}
	}
	authority, err := connectors.NewFixtureWriteApprovalAuthority()
	if err != nil {
		t.Fatalf("NewFixtureWriteApprovalAuthority() error = %v", err)
	}
	cfg.CredentialRevision, err = authority.CredentialRevision("amazon-sqs-fixture", cfg.Secrets)
	if err != nil {
		t.Fatalf("CredentialRevision(%s) error = %v", action, err)
	}
	cfg.ConfigurationDigest, err = authority.ConfigurationDigest("amazon-sqs-fixture", cfg.Config)
	if err != nil {
		t.Fatalf("ConfigurationDigest(%s) error = %v", action, err)
	}
	cfg.WriteApprovalScope = connectors.WriteApprovalScopeFixture
	req := connectors.WriteRequest{Action: action, Config: cfg}
	preview, err := connector.DryRunWrite(context.Background(), req, records)
	if err != nil {
		t.Fatalf("DryRunWrite(%s) error = %v", action, err)
	}
	token := "amazon-sqs-fixture-approval"
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID:        "rplan_amazon_sqs_fixture",
		PlanHash:      strings.Repeat("a", 64),
		PreviewDigest: preview.Digest,
		ApprovalToken: token,
		Target:        preview.ApprovalTarget,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("IssueWriteGrant(%s) error = %v", action, err)
	}
	req.Approval, err = authority.VerifyWriteGrant(grant, connectors.WriteApprovalExpectation{
		PlanID:        grant.PlanID,
		PlanHash:      grant.PlanHash,
		PreviewDigest: grant.PreviewDigest,
		ApprovalToken: token,
		Target:        grant.Target,
		Confirmation:  grant.Confirmation,
	})
	if err != nil {
		t.Fatalf("VerifyWriteGrant(%s) error = %v", action, err)
	}
	return req
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

func cloneValues(values url.Values) url.Values {
	out := make(url.Values, len(values))
	for key, item := range values {
		out[key] = append([]string(nil), item...)
	}
	return out
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

// TestOperationDirectReadReportsAndHonoursSQSPages closes the gap where the
// six amazon-sqs direct reads documented --page/--page-cursor, accepted them,
// and returned page one. The SQS Query API pages by an opaque NextToken, so
// the cursor must reach the wire and the returned page must say where it sits.
func TestOperationDirectReadReportsAndHonoursSQSPages(t *testing.T) {
	var sawTokens []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		sawTokens = append(sawTokens, r.Form.Get("NextToken"))
		if r.Form.Get("NextToken") == "page-two" {
			_, _ = w.Write([]byte(`<ListQueuesResponse><ListQueuesResult><QueueUrl>https://sqs.us-east-1.amazonaws.com/123/c</QueueUrl></ListQueuesResult></ListQueuesResponse>`))
			return
		}
		_, _ = w.Write([]byte(`<ListQueuesResponse><ListQueuesResult><QueueUrl>https://sqs.us-east-1.amazonaws.com/123/a</QueueUrl><QueueUrl>https://sqs.us-east-1.amazonaws.com/123/b</QueueUrl><NextToken>page-two</NextToken></ListQueuesResult></ListQueuesResponse>`))
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)

	first, err := c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{Operation: "list_queues", Config: cfg})
	if err != nil {
		t.Fatalf("OperationDirectRead list_queues: %v", err)
	}
	if first.Page.Records != 2 {
		t.Fatalf("page.records = %d, want 2", first.Page.Records)
	}
	if first.Page.Complete {
		t.Fatal("page.complete = true while a NextToken was returned")
	}
	if !first.Page.HasMore || first.Page.NextCursor != "page-two" {
		t.Fatalf("page has_more/next_cursor = %v/%q, want true/page-two", first.Page.HasMore, first.Page.NextCursor)
	}

	second, err := c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{Operation: "list_queues", Config: cfg, PageCursor: first.Page.NextCursor})
	if err != nil {
		t.Fatalf("OperationDirectRead list_queues(cursor): %v", err)
	}
	if len(sawTokens) != 2 || sawTokens[1] != "page-two" {
		t.Fatalf("NextToken values sent = %#v, want the cursor on the second request", sawTokens)
	}
	if second.Page.Records != 1 || !second.Page.Complete {
		t.Fatalf("second page records/complete = %d/%v, want 1/true", second.Page.Records, second.Page.Complete)
	}
	if got := len(first.Body.(map[string]any)["queue_urls"].([]string)) + len(second.Body.(map[string]any)["queue_urls"].([]string)); got != 3 {
		t.Fatalf("records reached across pages = %d, want 3", got)
	}
}

// TestOperationDirectReadRefusesSQSNavigationItCannotHonour: SQS has no
// addressable page number, and three of its reads return a single object. Both
// must be refused rather than answered with page one.
func TestOperationDirectReadRefusesSQSNavigationItCannotHonour(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = w.Write([]byte(`<ListQueuesResponse><ListQueuesResult></ListQueuesResult></ListQueuesResponse>`))
	}))
	defer srv.Close()

	c := native.New()
	cfg := testRuntimeConfig(srv.URL)
	for _, tc := range []struct {
		name string
		req  connectors.OperationDirectReadRequest
		want string
	}{
		{name: "page number on a cursor read", req: connectors.OperationDirectReadRequest{Operation: "list_queues", Page: 3}, want: "no addressable page number"},
		{name: "cursor on a single-object read", req: connectors.OperationDirectReadRequest{Operation: "get_queue_url", PageCursor: "abc"}, want: "cannot address a page by cursor"},
		{name: "page number on a single-object read", req: connectors.OperationDirectReadRequest{Operation: "get_queue_url", Page: 2}, want: "cannot address page 2"},
		{name: "cursor alongside next_token", req: connectors.OperationDirectReadRequest{Operation: "list_queues", PageCursor: "abc", Body: map[string]any{"next_token": "xyz"}}, want: "they select different pages"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := hits
			req := tc.req
			req.Config = cfg
			if req.Operation == "get_queue_url" {
				req.Body = map[string]any{"queue_name": "orders"}
			}
			_, err := c.OperationDirectRead(context.Background(), req)
			if err == nil {
				t.Fatal("navigation the operation cannot honour returned no error, want a refusal")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %q, want it to contain %q", err.Error(), tc.want)
			}
			if hits != before {
				t.Fatal("the refused navigation still reached AWS")
			}
		})
	}
}

// TestManifestRequiredFieldsCarryOnlyFieldNames pins the machine contract.
// required_fields is published by `pm connectors inspect --json` as a list of
// FIELD NAMES; the either/or constraint travels in required_any_fields instead,
// so a machine consumer never reads a synthesized sentence as a field.
func TestManifestRequiredFieldsCarryOnlyFieldNames(t *testing.T) {
	manifest := native.New().Manifest()
	var checked int
	for _, action := range manifest.WriteActions {
		for _, field := range action.RequiredFields {
			if strings.ContainsAny(field, " +") {
				t.Fatalf("write action %q required_fields contains %q, which is not a field name", action.Name, field)
			}
		}
		switch action.Name {
		case "set_queue_attributes":
			checked++
			if len(action.RequiredFields) != 0 {
				t.Fatalf("set_queue_attributes required_fields = %v, want empty", action.RequiredFields)
			}
			if fmt.Sprint(action.RequiredAnyFields) != "[[attributes] [attribute_name attribute_value]]" {
				t.Fatalf("set_queue_attributes required_any_fields = %v", action.RequiredAnyFields)
			}
		case "tag_queue":
			checked++
			if fmt.Sprint(action.RequiredAnyFields) != "[[tags] [tag_key tag_value]]" {
				t.Fatalf("tag_queue required_any_fields = %v", action.RequiredAnyFields)
			}
		}
		for _, group := range action.RequiredAnyFields {
			for _, field := range group {
				for _, optional := range action.OptionalFields {
					if field == optional {
						t.Fatalf("write action %q lists %q as both required-any and optional", action.Name, field)
					}
				}
			}
		}
	}
	if checked != 2 {
		t.Fatalf("checked %d either/or actions, want 2", checked)
	}
}
