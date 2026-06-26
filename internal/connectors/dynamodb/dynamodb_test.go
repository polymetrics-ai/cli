package dynamodb_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/dynamodb"
)

func TestReadItemsSignsPaginatesAndMapsRecords(t *testing.T) {
	var calls int
	var sawAuth, sawTarget string
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		sawAuth = r.Header.Get("Authorization")
		sawTarget = r.Header.Get("X-Amz-Target")
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		if r.Method != http.MethodPost || r.URL.Path != "/" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		switch calls {
		case 1:
			_, _ = w.Write([]byte(`{"Items":[{"pk":{"S":"user#1"},"name":{"S":"Ada"}},{"pk":{"S":"user#2"},"score":{"N":"42"}}],"LastEvaluatedKey":{"pk":{"S":"user#2"}}}`))
		case 2:
			if !strings.Contains(string(body), "ExclusiveStartKey") {
				t.Fatalf("second request missing ExclusiveStartKey body=%s", body)
			}
			_, _ = w.Write([]byte(`{"Items":[{"pk":{"S":"user#3"},"active":{"BOOL":true}}]}`))
		default:
			t.Fatalf("unexpected call %d", calls)
		}
	}))
	defer srv.Close()

	c := dynamodb.Connector{Client: srv.Client(), Now: func() time.Time { return time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC) }}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"endpoint": srv.URL, "region": "us-east-1", "table_name": "users", "page_size": "2"}, Secrets: map[string]string{"access_key_id": "AKID", "secret_access_key": "SECRET"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawTarget != "DynamoDB_20120810.Scan" {
		t.Fatalf("X-Amz-Target = %q", sawTarget)
	}
	if !strings.Contains(sawAuth, "AWS4-HMAC-SHA256") || !strings.Contains(sawAuth, "Credential=AKID/20260626/us-east-1/dynamodb/aws4_request") || strings.Contains(sawAuth, "SECRET") {
		t.Fatalf("Authorization header not signed as expected: %q", sawAuth)
	}
	if !strings.Contains(bodies[0], `"TableName":"users"`) {
		t.Fatalf("first request body missing table name: %s", bodies[0])
	}
	if len(got) != 3 || got[0]["name"] != "Ada" || got[1]["score"] != "42" || got[2]["active"] != true {
		t.Fatalf("records mapped wrong: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := dynamodb.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	count := 0
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: cfg}, func(connectors.Record) error { count++; return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if count == 0 {
		t.Fatal("fixture Read emitted no records")
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "dynamodb" || len(cat.Streams) != 1 {
		t.Fatalf("Catalog = %+v err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("dynamodb"); !ok {
		t.Fatal("registry did not resolve dynamodb")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
