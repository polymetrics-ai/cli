package testrail_test

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/testrail"
)

func TestReadAuthenticatesAndMapsProjects(t *testing.T) {
	var sawAuth string
	var sawURI string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawURI = r.URL.RequestURI()
		if r.URL.RequestURI() != "/index.php?/api/v2/get_projects" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":1,"name":"Core","announcement":"Fixture","is_completed":false},{"id":2,"name":"Mobile","is_completed":true}]`))
	}))
	defer srv.Close()

	c := testrail.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "username": "user@example.com"}, Secrets: map[string]string{"password": "test_password"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user@example.com:test_password"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if sawURI != "/index.php?/api/v2/get_projects" {
		t.Fatalf("URI = %q", sawURI)
	}
	if len(got) != 2 || got[0]["id"] == nil || got[1]["name"] != "Mobile" {
		t.Fatalf("records not mapped: %+v", got)
	}
}

func TestFixtureModeNoCredentials(t *testing.T) {
	c := testrail.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var count int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		count++
		if rec["id"] == nil {
			t.Fatalf("fixture missing id: %+v", rec)
		}
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if count == 0 {
		t.Fatal("fixture emitted no records")
	}
}

func TestCatalogRegistrationAndReadOnly(t *testing.T) {
	c := testrail.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "testrail" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	if caps := c.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
	if _, ok := connectors.NewRegistry().Get("testrail"); !ok {
		t.Fatal("registry did not resolve testrail")
	}
}
