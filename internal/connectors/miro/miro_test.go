package miro_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/miro"
)

// TestReadBoardsPaginatesAndAuthenticates is the red-first test: Bearer auth,
// Miro offset/limit pagination over data[], and record mapping for the boards
// stream.
func TestReadBoardsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v2/boards" {
			http.NotFound(w, r)
			return
		}
		offset := r.URL.Query().Get("offset")
		if r.URL.Query().Get("limit") == "" {
			t.Errorf("expected limit query param, got none")
		}
		switch offset {
		case "", "0":
			// First page: a full page (size 2 here) so the reader asks for more.
			_, _ = w.Write([]byte(`{"data":[{"id":"b1","name":"Board One","type":"board"},{"id":"b2","name":"Board Two","type":"board"}],"total":3,"size":2,"offset":0,"limit":2}`))
		case "2":
			// Second (short) page: stops pagination.
			_, _ = w.Write([]byte(`{"data":[{"id":"b3","name":"Board Three","type":"board"}],"total":3,"size":1,"offset":2,"limit":2}`))
		default:
			t.Errorf("unexpected offset=%q", offset)
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := miro.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "miro_secret_token"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "boards", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer miro_secret_token" {
		t.Fatalf("Authorization = %q, want Bearer miro_secret_token", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadBoardItemsUsesBoardID exercises a board-scoped sub-stream: the
// configured board_id is interpolated into the path and the data[] array is
// mapped, carrying the board_id onto each record.
func TestReadBoardItemsUsesBoardID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/boards/board_42/items" {
			t.Errorf("unexpected path %q", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"i1","type":"sticky_note"},{"id":"i2","type":"shape"}],"total":2,"size":2,"offset":0}`))
	}))
	defer srv.Close()

	c := miro.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "board_id": "board_42", "page_size": "50"},
		Secrets: map[string]string{"api_key": "tok"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "board_items", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["board_id"] != "board_42" {
		t.Fatalf("board_id = %v, want board_42", got[0]["board_id"])
	}
}

func TestReadBoardItemsRequiresBoardID(t *testing.T) {
	c := miro.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://api.miro.com"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "board_items", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("board_items without board_id should error")
	}
}

func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := miro.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"boards", "board_users", "board_items", "board_tags", "board_connectors"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCheckFixtureModeNoNetwork(t *testing.T) {
	c := miro.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestUnknownStreamRejected(t *testing.T) {
	c := miro.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nope", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("unknown stream should error")
	}
}

func TestBadBaseURLRejected(t *testing.T) {
	c := miro.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("non-http base_url should be rejected")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := miro.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "miro" {
		t.Fatalf("catalog connector = %q, want miro", cat.Connector)
	}
	want := map[string]bool{"boards": false, "board_users": false, "board_items": false, "board_tags": false, "board_connectors": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("catalog missing stream %q", name)
		}
	}
}

func TestMetadataReadOnly(t *testing.T) {
	c := miro.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("miro is read-only, Write should be false")
	}
}

func TestRegisteredInRegistry(t *testing.T) {
	_ = miro.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("miro"); !ok {
		t.Fatal("registry did not resolve miro (self-registration)")
	}
}
