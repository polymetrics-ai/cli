package cloudbeds_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/cloudbeds"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Cloudbeds
// connector: Bearer auth on Authorization, page-increment pagination over the
// data[] array using pageNumber, and record mapping. It drives two pages: the
// first returns a full page (pageSize records) so a second page is fetched, and
// the second returns a short page so pagination stops.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/getReservations" {
			http.NotFound(w, r)
			return
		}
		size, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
		if size <= 0 {
			t.Errorf("missing pageSize, got query %q", r.URL.RawQuery)
		}
		switch r.URL.Query().Get("pageNumber") {
		case "1":
			// A full page: exactly pageSize records -> connector asks for page 2.
			var b []byte
			b = append(b, []byte(`{"success":true,"data":[`)...)
			for i := 0; i < size; i++ {
				if i > 0 {
					b = append(b, ',')
				}
				b = append(b, []byte(fmt.Sprintf(`{"reservationID":"res_%d","status":"confirmed","propertyID":"p1"}`, i))...)
			}
			b = append(b, []byte(`],"count":`+strconv.Itoa(size)+`}`)...)
			_, _ = w.Write(b)
		case "2":
			// A short page: 1 record -> pagination stops.
			_, _ = w.Write([]byte(`{"success":true,"data":[{"reservationID":"res_last","status":"checked_out","propertyID":"p1"}],"count":1}`))
		default:
			t.Errorf("unexpected pageNumber=%q", r.URL.Query().Get("pageNumber"))
			_, _ = w.Write([]byte(`{"success":true,"data":[],"count":0}`))
		}
	}))
	defer srv.Close()

	c := cloudbeds.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "cbat_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reservations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer cbat_test_123" {
		t.Fatalf("Authorization = %q, want Bearer cbat_test_123", sawAuth)
	}
	// page 1 = 2 records (pageSize=2), page 2 = 1 record => 3 total across 2 pages.
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["reservationID"] == nil || rec["status"] == nil {
			t.Fatalf("record missing reservationID/status: %+v", rec)
		}
	}
}

// TestFixtureModeReadsWithoutNetwork confirms the credential-free fixture path
// emits deterministic records so conformance can run without live creds.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := cloudbeds.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "guests", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture Read emitted no records")
	}
	for _, rec := range got {
		if rec["guestID"] == nil {
			t.Fatalf("fixture record missing guestID: %+v", rec)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog and primary keys.
func TestCatalogStreams(t *testing.T) {
	c := cloudbeds.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]string{
		"guests":       "guestID",
		"hotels":       "",
		"rooms":        "propertyID",
		"reservations": "reservationID",
		"transactions": "transactionID",
	}
	for _, s := range cat.Streams {
		pk, ok := want[s.Name]
		if !ok {
			continue
		}
		delete(want, s.Name)
		if pk == "" {
			continue
		}
		if len(s.PrimaryKey) != 1 || s.PrimaryKey[0] != pk {
			t.Fatalf("stream %s primary key = %v, want [%s]", s.Name, s.PrimaryKey, pk)
		}
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

// TestRegistryResolvesCloudbeds confirms self-registration via init().
func TestRegistryResolvesCloudbeds(t *testing.T) {
	_ = cloudbeds.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("cloudbeds"); !ok {
		t.Fatal("registry did not resolve cloudbeds (self-registration)")
	}
	caps := cloudbeds.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
