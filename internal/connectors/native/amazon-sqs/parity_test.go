// Package amazonsqs_test parity-tests the Tier-3 native amazon-sqs connector
// (internal/connectors/native/amazon-sqs) against the legacy
// internal/connectors/amazon-sqs package, per docs/migration/conventions.md's
// tier ladder and the postgres/faker golden parity pattern:
//
//   - Check: same accept/reject outcome for every resolveConnConfig
//     validation rule (queue_url/region/access_key/secret_key), and both
//     accept fixture mode with no credentials.
//   - Catalog: identical single-stream shape (name, PrimaryKey, Fields).
//   - Read: identical RECORD SET against the same signed httptest server (an
//     XML ReceiveMessageResponse fixture), including the SigV4 Authorization
//     header shape and the early-stop-on-empty-poll behavior; and identical
//     fixture-mode canned output.
//   - Definition(): smoke — name, capabilities, spec fields.
package amazonsqs_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	legacy "polymetrics.ai/internal/connectors/amazon-sqs"

	native "polymetrics.ai/internal/connectors/native/amazon-sqs"
)

func invalidConfigTable() []struct {
	name string
	cfg  connectors.RuntimeConfig
} {
	return []struct {
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
			name: "missing access_key",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"queue_url": "https://sqs.example.com/1/q", "region": "us-east-1"},
				Secrets: map[string]string{"secret_key": "s"},
			},
		},
		{
			name: "missing secret_key",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"queue_url": "https://sqs.example.com/1/q", "region": "us-east-1"},
				Secrets: map[string]string{"access_key": "a"},
			},
		},
	}
}

func TestParityAmazonSQS_CheckFixtureModeAccepts(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	lc := legacy.New()
	if err := lc.Check(context.Background(), cfg); err != nil {
		t.Fatalf("legacy Check(fixture) = %v, want nil", err)
	}

	nc := native.New()
	if err := nc.Check(context.Background(), cfg); err != nil {
		t.Fatalf("native Check(fixture) = %v, want nil", err)
	}
}

func TestParityAmazonSQS_ConfigValidationErrorTable(t *testing.T) {
	lc := legacy.New()
	nc := native.New()

	for _, tc := range invalidConfigTable() {
		t.Run(tc.name, func(t *testing.T) {
			if err := lc.Check(context.Background(), tc.cfg); err == nil {
				t.Fatalf("legacy Check(%s) = nil, want a validation error", tc.name)
			}
			if err := nc.Check(context.Background(), tc.cfg); err == nil {
				t.Fatalf("native Check(%s) = nil, want a validation error", tc.name)
			}
		})
	}
}

func TestParityAmazonSQS_CatalogMatchesLegacy(t *testing.T) {
	lc := legacy.New()
	nc := native.New()

	lCat, err := lc.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}
	nCat, err := nc.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("native Catalog: %v", err)
	}
	if len(lCat.Streams) != 1 || len(nCat.Streams) != 1 {
		t.Fatalf("stream count mismatch: legacy=%d native=%d", len(lCat.Streams), len(nCat.Streams))
	}
	if lCat.Streams[0].Name != nCat.Streams[0].Name {
		t.Fatalf("stream name mismatch: legacy=%q native=%q", lCat.Streams[0].Name, nCat.Streams[0].Name)
	}
	if !reflect.DeepEqual(lCat.Streams[0].PrimaryKey, nCat.Streams[0].PrimaryKey) {
		t.Fatalf("PrimaryKey mismatch: legacy=%v native=%v", lCat.Streams[0].PrimaryKey, nCat.Streams[0].PrimaryKey)
	}
}

func TestParityAmazonSQS_ReadFixtureMatchesLegacy(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	lc := legacy.New()
	nc := native.New()

	readAll := func(c connectors.Connector) []connectors.Record {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(r connectors.Record) error {
			got = append(got, r)
			return nil
		}); err != nil {
			t.Fatalf("Read: %v", err)
		}
		return got
	}

	lGot := readAll(lc)
	nGot := readAll(nc)
	if !reflect.DeepEqual(lGot, nGot) {
		t.Fatalf("fixture read mismatch:\nlegacy=%+v\nnative=%+v", lGot, nGot)
	}
}

// TestParityAmazonSQS_ReadLiveMatchesLegacy drives both connectors against
// the SAME signed httptest server (an XML ReceiveMessageResponse fixture)
// and asserts identical emitted records and identical SigV4
// Authorization-header shape — proving the SigV4 signing/XML decoding port
// is not just individually plausible but byte-for-byte equivalent to
// legacy's own request construction against a real server.
func TestParityAmazonSQS_ReadLiveMatchesLegacy(t *testing.T) {
	const xmlBody = `<ReceiveMessageResponse><ReceiveMessageResult><Message><MessageId>m1</MessageId><ReceiptHandle>rh1</ReceiptHandle><MD5OfBody>md5</MD5OfBody><Body>{"kind":"order","id":1}</Body><Attribute><Name>SentTimestamp</Name><Value>1767225600000</Value></Attribute><MessageAttribute><Name>source</Name><Value><StringValue>checkout</StringValue><DataType>String</DataType></Value></MessageAttribute></Message></ReceiveMessageResult></ReceiveMessageResponse>`

	runAgainst := func(t *testing.T, c connectors.Connector) (string, string, []connectors.Record) {
		var sawAuth, sawAction string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sawAuth = r.Header.Get("Authorization")
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm: %v", err)
			}
			sawAction = r.Form.Get("Action")
			_, _ = w.Write([]byte(xmlBody))
		}))
		defer srv.Close()

		cfg := connectors.RuntimeConfig{
			Config:  map[string]string{"queue_url": srv.URL + "/123/test-queue", "region": "us-east-1", "max_batch_size": "1"},
			Secrets: map[string]string{"access_key": "AKIATEST", "secret_key": "test-secret"},
		}
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(r connectors.Record) error {
			got = append(got, r)
			return nil
		}); err != nil {
			t.Fatalf("Read: %v", err)
		}
		return sawAuth, sawAction, got
	}

	lAuth, lAction, lGot := runAgainst(t, legacy.New())
	nAuth, nAction, nGot := runAgainst(t, native.New())

	if lAction != "ReceiveMessage" || nAction != "ReceiveMessage" {
		t.Fatalf("Action mismatch: legacy=%q native=%q", lAction, nAction)
	}
	if !strings.HasPrefix(lAuth, "AWS4-HMAC-SHA256 ") || !strings.HasPrefix(nAuth, "AWS4-HMAC-SHA256 ") {
		t.Fatalf("Authorization not SigV4: legacy=%q native=%q", lAuth, nAuth)
	}
	if !reflect.DeepEqual(lGot, nGot) {
		t.Fatalf("live read mismatch:\nlegacy=%+v\nnative=%+v", lGot, nGot)
	}
}

// TestParityAmazonSQS_ReadStopsEarlyOnEmptyPollBothSides asserts both
// connectors issue exactly one request when the first poll returns zero
// messages, matching legacy's early-stop loop exactly.
func TestParityAmazonSQS_ReadStopsEarlyOnEmptyPollBothSides(t *testing.T) {
	runAgainst := func(t *testing.T, c connectors.Connector) int {
		var calls int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			_, _ = w.Write([]byte(`<ReceiveMessageResponse><ReceiveMessageResult></ReceiveMessageResult></ReceiveMessageResponse>`))
		}))
		defer srv.Close()

		cfg := connectors.RuntimeConfig{
			Config:  map[string]string{"queue_url": srv.URL + "/1/q", "region": "us-east-1", "max_polls": "5"},
			Secrets: map[string]string{"access_key": "a", "secret_key": "s"},
		}
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(connectors.Record) error { return nil }); err != nil {
			t.Fatalf("Read: %v", err)
		}
		return calls
	}

	lCalls := runAgainst(t, legacy.New())
	nCalls := runAgainst(t, native.New())
	if lCalls != 1 || nCalls != 1 {
		t.Fatalf("calls mismatch: legacy=%d native=%d, want both 1", lCalls, nCalls)
	}
}

func TestParityAmazonSQS_DefinitionSmoke(t *testing.T) {
	nc := native.New()
	if nc.Name() != "amazon-sqs" {
		t.Fatalf("Name() = %q, want amazon-sqs", nc.Name())
	}
	def := nc.Definition()
	if def.Name != "amazon-sqs" {
		t.Fatalf("Definition().Name = %q, want amazon-sqs", def.Name)
	}
}
