package easypromos_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/easypromos"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Easypromos
// connector: Bearer auth, cursor pagination via paging.next_cursor over the
// items[] array, and record mapping. Red until internal/connectors/easypromos
// exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/promotions" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("next_cursor") {
		case "":
			_, _ = w.Write([]byte(`{"items":[{"id":"p_1","title":"Spring Giveaway"},{"id":"p_2","title":"Summer Contest"}],"paging":{"next_cursor":"abc123"}}`))
		case "abc123":
			_, _ = w.Write([]byte(`{"items":[{"id":"p_3","title":"Fall Sweepstakes"}],"paging":{"next_cursor":null}}`))
		default:
			t.Errorf("unexpected next_cursor=%q", r.URL.Query().Get("next_cursor"))
			_, _ = w.Write([]byte(`{"items":[],"paging":{"next_cursor":null}}`))
		}
	}))
	defer srv.Close()

	c := easypromos.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"bearer_token": "jwt_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "promotions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer jwt_test_123" {
		t.Fatalf("Authorization = %q, want Bearer jwt_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["title"] != "Spring Giveaway" {
		t.Fatalf("record mapping failed: title = %v, want Spring Giveaway", got[0]["title"])
	}
}

// TestReadSubstreamUsesPromotionID verifies that a per-promotion stream targets
// the /{resource}/{promotion_id} path using the promotion_id config.
func TestReadSubstreamUsesPromotionID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/42" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"items":[{"id":"u_1","email":"a@example.com"}],"paging":{"next_cursor":null}}`))
	}))
	defer srv.Close()

	c := easypromos.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "promotion_id": "42"},
		Secrets: map[string]string{"bearer_token": "jwt_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read users: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "u_1" {
		t.Fatalf("records = %+v, want one user u_1", got)
	}
}

// TestSubstreamRequiresPromotionID asserts a per-promotion stream errors when no
// promotion_id is configured (and not in fixture mode).
func TestSubstreamRequiresPromotionID(t *testing.T) {
	c := easypromos.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://api.easypromosapp.com/v2"},
		Secrets: map[string]string{"bearer_token": "jwt_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read users without promotion_id should error")
	}
}

// TestFixtureMode confirms fixture mode emits deterministic records with no
// network access, for every stream, so conformance passes without creds.
func TestFixtureMode(t *testing.T) {
	c := easypromos.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"promotions", "organizing_brands", "stages", "users", "participations", "prizes"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture Read(%s) records = %d, want 2", stream, len(got))
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := easypromos.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := easypromos.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"bearer_token": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "promotions", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected (SSRF guard)")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = easypromos.New() // ensure init ran
	c := easypromos.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only source)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("easypromos"); !ok {
		t.Fatal("registry did not resolve easypromos (self-registration)")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := easypromos.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{
		"promotions": true, "organizing_brands": true, "stages": true,
		"users": true, "participations": true, "prizes": true,
	}
	for _, s := range cat.Streams {
		delete(want, s.Name)
		if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "id" {
			t.Fatalf("stream %q primary key = %v, want [id]", s.Name, s.PrimaryKey)
		}
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}
