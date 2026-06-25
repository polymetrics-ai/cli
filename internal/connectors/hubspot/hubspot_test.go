package hubspot_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/hubspot"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the HubSpot
// connector: Bearer auth on the private-app access token, HubSpot CRM v3
// after-cursor pagination over results[], stopping when paging.next.after is
// absent, and record mapping (id + flattened properties).
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/crm/v3/objects/contacts" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`{
				"results":[
					{"id":"101","properties":{"email":"a@example.com","firstname":"Ada","createdate":"2026-01-01T00:00:00Z","lastmodifieddate":"2026-02-01T00:00:00Z"},"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-02-01T00:00:00Z","archived":false},
					{"id":"102","properties":{"email":"b@example.com","firstname":"Grace","createdate":"2026-01-02T00:00:00Z","lastmodifieddate":"2026-02-02T00:00:00Z"},"createdAt":"2026-01-02T00:00:00Z","updatedAt":"2026-02-02T00:00:00Z","archived":false}
				],
				"paging":{"next":{"after":"102","link":"https://api.hubapi.com/crm/v3/objects/contacts?after=102"}}
			}`))
		case "102":
			_, _ = w.Write([]byte(`{
				"results":[
					{"id":"103","properties":{"email":"c@example.com","firstname":"Katherine","createdate":"2026-01-03T00:00:00Z","lastmodifieddate":"2026-02-03T00:00:00Z"},"createdAt":"2026-01-03T00:00:00Z","updatedAt":"2026-02-03T00:00:00Z","archived":false}
				],
				"paging":{}
			}`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`{"results":[]}`))
		}
	}))
	defer srv.Close()

	c := hubspot.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_token": "pat-na1-secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer pat-na1-secret" {
		t.Fatalf("Authorization = %q, want Bearer pat-na1-secret", sawAuth)
	}
	if len(sawPaths) != 2 {
		t.Fatalf("requests = %d (%v), want 2 pages", len(sawPaths), sawPaths)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	// Record mapping: id is preserved and properties are flattened to the top.
	first := got[0]
	if first["id"] != "101" {
		t.Fatalf("record id = %v, want 101", first["id"])
	}
	if first["email"] != "a@example.com" {
		t.Fatalf("record email = %v, want a@example.com (flattened from properties)", first["email"])
	}
	if first["firstname"] != "Ada" {
		t.Fatalf("record firstname = %v, want Ada", first["firstname"])
	}
	if first["updatedAt"] != "2026-02-01T00:00:00Z" {
		t.Fatalf("record updatedAt = %v, want 2026-02-01T00:00:00Z", first["updatedAt"])
	}
}

// TestReadStreamsRouting confirms each core stream targets its CRM v3 object path.
func TestReadStreamsRouting(t *testing.T) {
	want := map[string]string{
		"contacts":  "/crm/v3/objects/contacts",
		"companies": "/crm/v3/objects/companies",
		"deals":     "/crm/v3/objects/deals",
		"tickets":   "/crm/v3/objects/tickets",
	}
	for stream, path := range want {
		var sawPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sawPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"results":[{"id":"1","properties":{}}],"paging":{}}`))
		}))
		cfg := connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": srv.URL},
			Secrets: map[string]string{"credentials.access_token": "tok"},
		}
		err := hubspot.New().Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(connectors.Record) error { return nil })
		srv.Close()
		if err != nil {
			t.Fatalf("Read(%s): %v", stream, err)
		}
		if sawPath != path {
			t.Fatalf("stream %s hit %q, want %q", stream, sawPath, path)
		}
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := hubspot.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Fixture Check must not require a secret or hit the network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresSecret(t *testing.T) {
	c := hubspot.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check without access token should fail")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := hubspot.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"credentials.access_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("base_url with non-http scheme should be rejected (SSRF guard)")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := hubspot.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "hubspot" {
		t.Fatalf("catalog connector = %q, want hubspot", cat.Connector)
	}
	names := map[string]bool{}
	for _, s := range cat.Streams {
		names[s.Name] = true
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	for _, want := range []string{"contacts", "companies", "deals", "tickets"} {
		if !names[want] {
			t.Fatalf("catalog missing stream %q", want)
		}
	}
}

func TestWriteValidateAllowList(t *testing.T) {
	c := hubspot.New()
	wv, ok := c.(connectors.WriteValidator)
	if !ok {
		t.Fatal("hubspot connector must implement WriteValidator")
	}
	cfg := connectors.RuntimeConfig{Secrets: map[string]string{"credentials.access_token": "tok"}}
	if err := wv.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "create_contact", Config: cfg}, []connectors.Record{{"email": "a@example.com"}}); err != nil {
		t.Fatalf("ValidateWrite(create_contact) = %v, want nil", err)
	}
	if err := wv.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "update_contact", Config: cfg}, []connectors.Record{{"id": "101", "email": "x@example.com"}}); err != nil {
		t.Fatalf("ValidateWrite(update_contact) = %v, want nil", err)
	}
	// update_contact requires an id.
	if err := wv.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "update_contact", Config: cfg}, []connectors.Record{{"email": "x@example.com"}}); err == nil {
		t.Fatal("update_contact without id should be rejected")
	}
	if err := wv.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "delete_everything", Config: cfg}, []connectors.Record{{}}); err == nil {
		t.Fatal("ValidateWrite(unknown action) should be rejected")
	}
}

func TestWriteCreateContact(t *testing.T) {
	var body map[string]any
	var sawMethod, sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"999","properties":{"email":"new@example.com"}}`))
	}))
	defer srv.Close()

	c := hubspot.New().(connectors.WriteValidator).(interface {
		Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error)
	})
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_token": "tok"},
	}
	res, err := c.Write(context.Background(), connectors.WriteRequest{Action: "create_contact", Config: cfg}, []connectors.Record{{"email": "new@example.com", "firstname": "New"}})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if res.RecordsWritten != 1 {
		t.Fatalf("RecordsWritten = %d, want 1", res.RecordsWritten)
	}
	if sawMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", sawMethod)
	}
	if sawPath != "/crm/v3/objects/contacts" {
		t.Fatalf("path = %q, want /crm/v3/objects/contacts", sawPath)
	}
	props, ok := body["properties"].(map[string]any)
	if !ok {
		t.Fatalf("body missing properties object: %+v", body)
	}
	if props["email"] != "new@example.com" {
		t.Fatalf("properties.email = %v, want new@example.com", props["email"])
	}
	if props["firstname"] != "New" {
		t.Fatalf("properties.firstname = %v, want New", props["firstname"])
	}
}

func TestRegisteredWithCapabilities(t *testing.T) {
	_ = hubspot.New() // ensure init ran
	c := hubspot.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Write || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Check/Catalog/Read/Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("hubspot"); !ok {
		t.Fatal("registry did not resolve hubspot (self-registration)")
	}
}

func TestDryRunWrite(t *testing.T) {
	c := hubspot.New().(connectors.DryRunWriter)
	cfg := connectors.RuntimeConfig{Secrets: map[string]string{"credentials.access_token": "tok"}}
	prev, err := c.DryRunWrite(context.Background(), connectors.WriteRequest{Action: "create_contact", Config: cfg}, []connectors.Record{{"email": "a@example.com"}})
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	if prev.RecordsStaged != 1 {
		t.Fatalf("RecordsStaged = %d, want 1", prev.RecordsStaged)
	}
	if prev.Action != "create_contact" {
		t.Fatalf("Action = %q, want create_contact", prev.Action)
	}
}
