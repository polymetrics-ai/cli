package campayn_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/campayn"
)

// TestReadListsAuthenticates is the red-first test for the Campayn connector:
// it asserts the custom TRUEREST apikey Authorization header, that the root JSON
// array is extracted as records, and that records are mapped with their id.
func TestReadListsAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/lists.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"1","list_name":"Newsletter","contact_count":42},{"id":"2","list_name":"Promo","contact_count":7}]`))
	}))
	defer srv.Close()

	c := campayn.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "TRUEREST apikey=key_123" {
		t.Fatalf("Authorization = %q, want TRUEREST apikey=key_123", sawAuth)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "1" || got[0]["list_name"] != "Newsletter" {
		t.Fatalf("record[0] mapped wrong: %+v", got[0])
	}
}

// TestReadContactsTraversesListPartitions asserts the substream fan-out: contacts
// is read per parent list, so the connector first lists then fetches contacts for
// each list id (the multi-request "pagination" equivalent for this API).
func TestReadContactsTraversesListPartitions(t *testing.T) {
	var contactPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/lists.json":
			_, _ = w.Write([]byte(`[{"id":"10"},{"id":"20"}]`))
		case strings.HasSuffix(r.URL.Path, "/contacts.json"):
			contactPaths = append(contactPaths, r.URL.Path)
			switch r.URL.Path {
			case "/lists/10/contacts.json":
				_, _ = w.Write([]byte(`[{"id":"c1","email":"a@example.com","first_name":"Ada"}]`))
			case "/lists/20/contacts.json":
				_, _ = w.Write([]byte(`[{"id":"c2","email":"b@example.com","first_name":"Bob"}]`))
			default:
				http.NotFound(w, r)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := campayn.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(contactPaths) != 2 {
		t.Fatalf("contact requests = %d (%v), want 2 (one per list partition)", len(contactPaths), contactPaths)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (one contact per list)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("contact record missing id: %+v", rec)
		}
		if rec["list_id"] == nil {
			t.Fatalf("contact record missing injected list_id: %+v", rec)
		}
	}
}

// TestFixtureModeReadsWithoutNetwork ensures credential-free conformance works.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := campayn.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"lists", "forms", "contacts", "emails", "reports"} {
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
}

func TestCheckFixtureModeNoNetwork(t *testing.T) {
	c := campayn.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := campayn.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"lists": true, "forms": true, "contacts": true, "emails": true, "reports": true}
	if len(cat.Streams) != len(want) {
		t.Fatalf("streams = %d, want %d", len(cat.Streams), len(want))
	}
	for _, s := range cat.Streams {
		if !want[s.Name] {
			t.Fatalf("unexpected stream %q", s.Name)
		}
		if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "id" {
			t.Fatalf("stream %q primary key = %v, want [id]", s.Name, s.PrimaryKey)
		}
	}
}

func TestReadOnlyCapability(t *testing.T) {
	caps := campayn.New().Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read+Check+Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("campayn API is read-only; Write should be false")
	}
}

func TestRegistryResolution(t *testing.T) {
	_ = campayn.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("campayn"); !ok {
		t.Fatal("registry did not resolve campayn (self-registration)")
	}
}

// TestSubDomainTemplatesHost confirms a valid sub_domain produces a real read
// against <sub>.campayn.com path while an injection-style sub_domain is rejected.
func TestSubDomainRejectsInjection(t *testing.T) {
	c := campayn.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"sub_domain": "evil.com/x"},
		Secrets: map[string]string{"api_key": "key_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with injection sub_domain should error")
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := campayn.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "key_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should error")
	}
}
