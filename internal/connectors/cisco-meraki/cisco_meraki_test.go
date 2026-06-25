package ciscomeraki_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	ciscomeraki "polymetrics.ai/internal/connectors/cisco-meraki"
)

// TestReadOrganizationsPaginatesAndAuthenticates is the red-first test: it asserts
// Bearer auth on the Authorization header and RFC5988 Link-header pagination across
// two pages of the organizations stream, plus record mapping.
func TestReadOrganizationsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/organizations" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("startingAfter") {
		case "":
			// First page: signal there is a next page via the Link header.
			next := fmt.Sprintf("<%s/organizations?perPage=2&startingAfter=org_2>; rel=\"next\"", srvURL)
			w.Header().Set("Link", next)
			_, _ = w.Write([]byte(`[{"id":"org_1","name":"Acme"},{"id":"org_2","name":"Globex"}]`))
		case "org_2":
			// Last page: no Link header => pagination stops.
			_, _ = w.Write([]byte(`[{"id":"org_3","name":"Initech"}]`))
		default:
			t.Errorf("unexpected startingAfter=%q", r.URL.Query().Get("startingAfter"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := ciscomeraki.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer key_abc123" {
		t.Fatalf("Authorization = %q, want Bearer key_abc123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages via Link header)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadOrgScopedFansOutPerOrganization verifies that an org-scoped stream
// (organization_devices) first lists organizations and then reads the per-org
// endpoint for each, stamping organizationId onto every record.
func TestReadOrgScopedFansOutPerOrganization(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/organizations":
			_, _ = w.Write([]byte(`[{"id":"org_1","name":"Acme"},{"id":"org_2","name":"Globex"}]`))
		case r.URL.Path == "/organizations/org_1/devices":
			_, _ = w.Write([]byte(`[{"serial":"Q2XX-1111","name":"AP-1","model":"MR46"}]`))
		case r.URL.Path == "/organizations/org_2/devices":
			_, _ = w.Write([]byte(`[{"serial":"Q2XX-2222","name":"SW-1","model":"MS120"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := ciscomeraki.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organization_devices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (one device per org)", len(got))
	}
	orgIDs := map[string]bool{}
	for _, rec := range got {
		if rec["serial"] == nil {
			t.Fatalf("record missing serial: %+v", rec)
		}
		orgID, _ := rec["organizationId"].(string)
		if orgID == "" {
			t.Fatalf("record missing organizationId stamp: %+v", rec)
		}
		orgIDs[orgID] = true
	}
	if !orgIDs["org_1"] || !orgIDs["org_2"] {
		t.Fatalf("expected devices stamped with both org ids, got %v", orgIDs)
	}
}

// TestFixtureModeNeedsNoNetwork confirms the credential-free fixture path the
// conformance harness relies on: deterministic records, no HTTP server.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := ciscomeraki.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"organizations", "organization_networks", "organization_devices", "organization_admins"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) produced no records", stream)
		}
	}

	// Check must also short-circuit in fixture mode with no creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on the base_url override.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := ciscomeraki.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "key_abc123"},
	}
	err := c.Check(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Check with file:// base_url = %v, want base_url scheme error", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := ciscomeraki.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{
		"organizations":         false,
		"organization_networks": false,
		"organization_devices":  false,
		"organization_admins":   false,
	}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("stream %q has no primary key", s.Name)
			}
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration via the process-global
// registry and that the connector advertises read (not write) capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = ciscomeraki.New() // ensure init() ran
	caps := ciscomeraki.New().Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (Meraki connector is read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("cisco-meraki"); !ok {
		t.Fatal("registry did not resolve cisco-meraki (self-registration)")
	}
}
