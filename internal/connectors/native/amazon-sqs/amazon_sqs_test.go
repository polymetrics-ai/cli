package amazonsqs_test

import (
	"context"
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
	if caps.Write {
		t.Fatalf("amazon-sqs source connector must be read-only, got Write=true")
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

func TestWriteUnsupported(t *testing.T) {
	c := native.New()
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write err = %v, want ErrUnsupportedOperation", err)
	}
}
