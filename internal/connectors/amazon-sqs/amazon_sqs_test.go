package amazonsqs_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	amazonsqs "polymetrics.ai/internal/connectors/amazon-sqs"
)

func TestReadMessagesSignsAuthenticatesAndMaps(t *testing.T) {
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

	c := amazonsqs.New()
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

func TestFixtureCatalogRegistryAndWrite(t *testing.T) {
	c := amazonsqs.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var n int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(connectors.Record) error { n++; return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if n == 0 {
		t.Fatal("fixture Read emitted no records")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) == 0 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write err = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("amazon-sqs"); !ok {
		t.Fatal("registry did not resolve amazon-sqs")
	}
}
