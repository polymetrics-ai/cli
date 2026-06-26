package boxdataextract_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	boxdataextract "polymetrics.ai/internal/connectors/box-data-extract"
)

func TestReadFilesPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var tokenCalls int
	var sawBearer, sawSubjectType, sawSubjectID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokenCalls++
			_ = r.ParseForm()
			sawSubjectType = r.Form.Get("box_subject_type")
			sawSubjectID = r.Form.Get("box_subject_id")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"box_tok","expires_in":3600}`))
		case "/2.0/folders/0/items":
			sawBearer = r.Header.Get("Authorization")
			offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
			switch offset {
			case 0:
				_, _ = w.Write([]byte(`{"total_count":3,"entries":[{"id":"file_1","type":"file","name":"one.pdf"},{"id":"file_2","type":"file","name":"two.pdf"}]}`))
			case 2:
				_, _ = w.Write([]byte(`{"total_count":3,"entries":[{"id":"file_3","type":"file","name":"three.pdf"}]}`))
			default:
				t.Fatalf("unexpected offset %d", offset)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := boxdataextract.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/2.0", "token_url": srv.URL + "/oauth2/token", "box_folder_id": "0", "box_subject_type": "enterprise", "box_subject_id": "ent_1", "page_size": "2"},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "files", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenCalls == 0 || sawSubjectType != "enterprise" || sawSubjectID != "ent_1" {
		t.Fatalf("token exchange not wired correctly: calls=%d type=%q id=%q", tokenCalls, sawSubjectType, sawSubjectID)
	}
	if sawBearer != "Bearer box_tok" {
		t.Fatalf("Authorization = %q, want Bearer token", sawBearer)
	}
	if len(got) != 3 || got[0]["id"] != "file_1" || got[0]["name"] != "one.pdf" {
		t.Fatalf("records mapped wrong: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := boxdataextract.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"files", "file_text"} {
		count := 0
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(connectors.Record) error { count++; return nil }); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if count == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "box-data-extract" || len(cat.Streams) != 2 {
		t.Fatalf("Catalog = %+v err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("box-data-extract"); !ok {
		t.Fatal("registry did not resolve box-data-extract")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
