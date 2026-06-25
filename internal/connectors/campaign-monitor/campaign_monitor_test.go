package campaignmonitor_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	campaignmonitor "polymetrics.ai/internal/connectors/campaign-monitor"
)

// TestReadPaginatesAndAuthenticates is the red-first test: HTTP Basic auth
// (API key/username as the user, password as the pass), Campaign Monitor's
// page/NumberOfPages pagination over the Results array, and record mapping.
// The campaigns stream is per-client, so it reads /clients/{clientid}/campaigns.json.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/clients/cid_1/campaigns.json" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"Results":[{"CampaignID":"camp_1","Name":"Spring","Subject":"Hi","SentDate":"2026-01-01 10:00:00","TotalRecipients":100},{"CampaignID":"camp_2","Name":"Summer","Subject":"Yo","SentDate":"2026-02-01 10:00:00","TotalRecipients":200}],"PageNumber":1,"PageSize":2,"RecordsOnThisPage":2,"TotalNumberOfRecords":3,"NumberOfPages":2}`))
		case "2":
			_, _ = w.Write([]byte(`{"Results":[{"CampaignID":"camp_3","Name":"Fall","Subject":"Hey","SentDate":"2026-03-01 10:00:00","TotalRecipients":300}],"PageNumber":2,"PageSize":2,"RecordsOnThisPage":1,"TotalNumberOfRecords":3,"NumberOfPages":2}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"Results":[],"PageNumber":99,"NumberOfPages":2}`))
		}
	}))
	defer srv.Close()

	c := campaignmonitor.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "my_api_key", "client_id": "cid_1"},
		Secrets: map[string]string{"password": "x"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("my_api_key:x"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["CampaignID"] == nil || rec["Name"] == nil {
			t.Fatalf("record missing CampaignID/Name: %+v", rec)
		}
	}
	if got[0]["CampaignID"] != "camp_1" || got[2]["CampaignID"] != "camp_3" {
		t.Fatalf("unexpected record order: %v / %v", got[0]["CampaignID"], got[2]["CampaignID"])
	}
}

// TestReadClientsTopLevelArray covers the clients stream, whose response is a
// bare top-level JSON array (no Results wrapper, single page).
func TestReadClientsTopLevelArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clients.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"ClientID":"cid_1","Name":"Acme"},{"ClientID":"cid_2","Name":"Globex"}]`))
	}))
	defer srv.Close()

	c := campaignmonitor.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "my_api_key"},
		Secrets: map[string]string{"password": "x"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read clients: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("clients = %d, want 2", len(got))
	}
	if got[0]["ClientID"] != "cid_1" || got[1]["Name"] != "Globex" {
		t.Fatalf("unexpected clients mapping: %+v", got)
	}
}

// TestFixtureModeNoNetwork verifies credential-free conformance: fixture mode
// emits deterministic records without any HTTP call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := campaignmonitor.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"clients", "campaigns", "lists", "suppressionlist"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read %s returned no records", stream)
		}
	}
	// Check should also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := campaignmonitor.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"clients": false, "campaigns": false, "lists": false, "suppressionlist": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q missing fields", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestBaseURLSSRFValidation(t *testing.T) {
	c := campaignmonitor.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil", "username": "k"},
		Secrets: map[string]string{"password": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = campaignmonitor.New() // ensure init ran
	c := campaignmonitor.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("campaign-monitor is read-only; Write should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("campaign-monitor"); !ok {
		t.Fatal("registry did not resolve campaign-monitor (self-registration)")
	}
}
