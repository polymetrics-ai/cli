package kisi_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/kisi"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Kisi
// connector: it asserts the Authorization: KISI-LOGIN header, offset/limit
// pagination across two pages of a top-level JSON array, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	const pageSize = 2
	var sawAuth string
	var sawLimits []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/members" {
			http.NotFound(w, r)
			return
		}
		sawLimits = append(sawLimits, r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		switch offset {
		case 0:
			// Full page of pageSize records -> paginator continues.
			_, _ = w.Write([]byte(`[{"id":1,"name":"Ada","email":"ada@example.com"},{"id":2,"name":"Grace","email":"grace@example.com"}]`))
		case pageSize:
			// Short page (1 < pageSize) -> paginator stops.
			_, _ = w.Write([]byte(`[{"id":3,"name":"Katherine","email":"katherine@example.com"}]`))
		default:
			t.Errorf("unexpected offset=%d", offset)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := kisi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": strconv.Itoa(pageSize)},
		Secrets: map[string]string{"api_key": "kisi_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "members", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "KISI-LOGIN kisi_test_key" {
		t.Fatalf("Authorization = %q, want KISI-LOGIN kisi_test_key", sawAuth)
	}
	if len(sawLimits) != 2 {
		t.Fatalf("requested %d pages, want 2", len(sawLimits))
	}
	for _, l := range sawLimits {
		if l != strconv.Itoa(pageSize) {
			t.Fatalf("limit param = %q, want %d", l, pageSize)
		}
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestSecretMissingFails ensures a real (non-fixture) read without an api_key
// fails fast rather than issuing an unauthenticated request.
func TestSecretMissingFails(t *testing.T) {
	c := kisi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://api.kisi.io"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "members", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read without api_key should fail")
	}
}

// TestBaseURLValidation rejects non-http(s) base_url overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := kisi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "members", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestFixtureModeNoNetwork exercises the credential-free fixture path used by
// the conformance harness: deterministic records, no network, no secret.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := kisi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"members", "locks", "groups", "users", "logins"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams verifies the published catalog carries the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := kisi.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "kisi" {
		t.Fatalf("catalog connector = %q, want kisi", cat.Connector)
	}
	want := map[string]bool{"members": false, "locks": false, "groups": false, "users": false, "logins": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegistryResolvesAndReadOnly confirms self-registration and the read-only
// capability surface.
func TestRegistryResolvesAndReadOnly(t *testing.T) {
	_ = kisi.New() // ensure init ran
	c := kisi.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (Kisi is read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("kisi"); !ok {
		t.Fatal("registry did not resolve kisi (self-registration)")
	}
}

// TestUnknownStream guards routing.
func TestUnknownStream(t *testing.T) {
	c := kisi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "does_not_exist", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read of unknown stream should fail")
	}
	_ = fmt.Sprint(err)
}
