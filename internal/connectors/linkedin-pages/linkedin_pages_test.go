package linkedinpages_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	linkedinpages "polymetrics.ai/internal/connectors/linkedin-pages"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the LinkedIn Pages
// connector: it asserts Bearer auth on the access token, the LinkedIn-Version and
// X-Restli-Protocol-Version headers, the org-scoped organizationalEntity finder,
// start/count offset pagination across two pages of elements[], and record
// mapping. Red until internal/connectors/linkedin-pages exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawVersion, sawRestli, sawEntity, sawFinder string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawVersion = r.Header.Get("LinkedIn-Version")
		sawRestli = r.Header.Get("X-Restli-Protocol-Version")
		sawEntity = r.URL.Query().Get("organizationalEntity")
		sawFinder = r.URL.Query().Get("q")
		if r.URL.Path != "/organizationalEntityFollowerStatistics" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("start") {
		case "", "0":
			// Full first page of count=2 -> there is a next page.
			_, _ = w.Write([]byte(`{"elements":[
				{"organizationalEntity":"urn:li:organization:123","followerCountsByAssociationType":[{"associationType":"SPONSORED","followerCounts":{"organicFollowerGain":5,"paidFollowerGain":1}}]},
				{"organizationalEntity":"urn:li:organization:123","followerCountsBySeniority":[{"seniority":"urn:li:seniority:9","followerCounts":{"organicFollowerGain":3,"paidFollowerGain":0}}]}
			]}`))
		case "2":
			// Short page -> pagination stops after this.
			_, _ = w.Write([]byte(`{"elements":[
				{"organizationalEntity":"urn:li:organization:123","followerCountsByCountry":[{"country":"urn:li:country:us","followerCounts":{"organicFollowerGain":2,"paidFollowerGain":0}}]}
			]}`))
		default:
			t.Errorf("unexpected start=%q", r.URL.Query().Get("start"))
			_, _ = w.Write([]byte(`{"elements":[]}`))
		}
	}))
	defer srv.Close()

	c := linkedinpages.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"credentials.access_token": "li_token_123", "org_id": "123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "follower_statistics", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer li_token_123" {
		t.Fatalf("Authorization = %q, want Bearer li_token_123", sawAuth)
	}
	if sawVersion == "" {
		t.Fatalf("LinkedIn-Version header was not set")
	}
	if sawRestli != "2.0.0" {
		t.Fatalf("X-Restli-Protocol-Version = %q, want 2.0.0", sawRestli)
	}
	if sawFinder != "organizationalEntity" {
		t.Fatalf("q = %q, want organizationalEntity", sawFinder)
	}
	if sawEntity != "urn:li:organization:123" {
		t.Fatalf("organizationalEntity = %q, want urn:li:organization:123", sawEntity)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["organizationalEntity"] == nil {
			t.Fatalf("record missing organizationalEntity: %+v", rec)
		}
		if rec["org_id"] != "123" {
			t.Fatalf("record org_id = %v, want 123", rec["org_id"])
		}
	}
}

// TestReadOrganizationLookup confirms the organizations stream reads the single
// organization object at /organizations/{org_id} and flattens key fields.
func TestReadOrganizationLookup(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{
			"id":123,
			"vanityName":"acme",
			"localizedName":"Acme Corp",
			"localizedWebsite":"https://acme.example",
			"primaryOrganizationType":"NONE",
			"$URN":"urn:li:organization:123"
		}`))
	}))
	defer srv.Close()

	c := linkedinpages.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_token": "tok", "org_id": "123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read organizations: %v", err)
	}
	if sawPath != "/organizations/123" {
		t.Fatalf("path = %q, want /organizations/123", sawPath)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["vanity_name"] != "acme" || got[0]["localized_name"] != "Acme Corp" {
		t.Fatalf("organization record not flattened: %+v", got[0])
	}
}

// TestReadTotalFollowerCount confirms the total_follower_count stream reads the
// networkSizes single object and maps firstDegreeSize.
func TestReadTotalFollowerCount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("edgeType"); got != "COMPANY_FOLLOWED_BY_MEMBER" {
			t.Errorf("edgeType = %q, want COMPANY_FOLLOWED_BY_MEMBER", got)
		}
		_, _ = w.Write([]byte(`{"firstDegreeSize":219145}`))
	}))
	defer srv.Close()

	c := linkedinpages.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_token": "tok", "org_id": "123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "total_follower_count", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read total_follower_count: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["first_degree_size"] == nil {
		t.Fatalf("record missing first_degree_size: %+v", got[0])
	}
	if got[0]["org_id"] != "123" {
		t.Fatalf("record org_id = %v, want 123", got[0]["org_id"])
	}
}

// TestReadRequiresOrgID confirms a non-fixture read fails without org_id.
func TestReadRequiresOrgID(t *testing.T) {
	c := linkedinpages.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://api.linkedin.com/rest"},
		Secrets: map[string]string{"credentials.access_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "follower_statistics", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read without org_id should fail")
	}
}

// TestFixtureModeReadsWithoutNetwork confirms fixture mode emits deterministic
// records with no creds and no network for every core stream.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := linkedinpages.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"organizations", "follower_statistics", "share_statistics", "total_follower_count"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture %s emitted no records", stream)
		}
		if got[0]["org_id"] == nil {
			t.Fatalf("fixture %s record missing org_id: %+v", stream, got[0])
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode with no creds.
func TestCheckFixtureMode(t *testing.T) {
	c := linkedinpages.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCheckRequiresSecret confirms Check fails without an access token when not
// in fixture mode.
func TestCheckRequiresSecret(t *testing.T) {
	c := linkedinpages.New()
	err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{}})
	if err == nil {
		t.Fatal("Check without access token should fail")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := linkedinpages.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"organizations": false, "follower_statistics": false, "share_statistics": false, "total_follower_count": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestBaseURLRejectsBadScheme confirms SSRF guard on base_url override.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := linkedinpages.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"credentials.access_token": "x", "org_id": "1"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "follower_statistics", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = linkedinpages.New() // ensure init ran
	c := linkedinpages.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("linkedin-pages"); !ok {
		t.Fatal("registry did not resolve linkedin-pages (self-registration)")
	}
}
