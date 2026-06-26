package huggingfacedatasets_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	huggingfacedatasets "polymetrics.ai/internal/connectors/hugging-face-datasets"
)

// TestReadRowsPaginatesAndAuthenticates is the red-first test: it asserts the
// optional Bearer auth header is sent, that the /rows endpoint is offset
// paginated across two pages, and that each emitted record carries the flattened
// row fields plus the row index. Red until the package exists.
func TestReadRowsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawDataset, sawConfig, sawSplit string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/rows" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		sawDataset = q.Get("dataset")
		sawConfig = q.Get("config")
		sawSplit = q.Get("split")
		offset := q.Get("offset")
		switch offset {
		case "0":
			// length=2 page that is full -> there must be a next page.
			_, _ = w.Write([]byte(`{"features":[],"rows":[{"row_idx":0,"row":{"text":"a","label":1}},{"row_idx":1,"row":{"text":"b","label":0}}],"num_rows_total":3}`))
		case "2":
			_, _ = w.Write([]byte(`{"features":[],"rows":[{"row_idx":2,"row":{"text":"c","label":1}}],"num_rows_total":3}`))
		default:
			t.Errorf("unexpected offset=%q", offset)
			_, _ = w.Write([]byte(`{"features":[],"rows":[],"num_rows_total":3}`))
		}
	}))
	defer srv.Close()

	c := huggingfacedatasets.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":     srv.URL,
			"dataset_name": "ibm/duorc",
			"config":       "SelfRC",
			"split":        "train",
			"page_size":    "2",
		},
		Secrets: map[string]string{"api_token": "hf_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "rows", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer hf_test_123" {
		t.Fatalf("Authorization = %q, want Bearer hf_test_123", sawAuth)
	}
	if sawDataset != "ibm/duorc" || sawConfig != "SelfRC" || sawSplit != "train" {
		t.Fatalf("query params dataset=%q config=%q split=%q", sawDataset, sawConfig, sawSplit)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for i, rec := range got {
		if rec["row_idx"] == nil {
			t.Fatalf("record %d missing row_idx: %+v", i, rec)
		}
		if rec["text"] == nil {
			t.Fatalf("record %d missing flattened row field text: %+v", i, rec)
		}
		if rec["dataset"] != "ibm/duorc" {
			t.Fatalf("record %d dataset = %v, want ibm/duorc", i, rec["dataset"])
		}
	}
	if got[0]["text"] != "a" || got[2]["text"] != "c" {
		t.Fatalf("row order/mapping wrong: %v", got)
	}
}

// TestReadSplitsNoAuthHeaderWhenTokenAbsent verifies the splits stream maps the
// {"splits":[...]} payload and that no Authorization header is sent when no token
// is configured (public datasets are read without credentials).
func TestReadSplitsNoAuthHeaderWhenTokenAbsent(t *testing.T) {
	var hadAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hadAuth = r.Header["Authorization"]
		if r.URL.Path != "/splits" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"splits":[{"dataset":"ibm/duorc","config":"SelfRC","split":"train"},{"dataset":"ibm/duorc","config":"SelfRC","split":"test"}],"pending":[],"failed":[]}`))
	}))
	defer srv.Close()

	c := huggingfacedatasets.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL, "dataset_name": "ibm/duorc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "splits", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read splits: %v", err)
	}
	if hadAuth {
		t.Fatal("Authorization header sent without a configured token")
	}
	if len(got) != 2 {
		t.Fatalf("splits = %d, want 2", len(got))
	}
	if got[0]["split"] != "train" || got[0]["config"] != "SelfRC" {
		t.Fatalf("split mapping wrong: %+v", got[0])
	}
}

// TestFixtureModeRead exercises the credential-free fixture path used by the
// conformance harness.
func TestFixtureModeRead(t *testing.T) {
	c := huggingfacedatasets.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "dataset_name": "fixture/dataset"}}
	for _, stream := range []string{"splits", "sizes", "rows"} {
		var n int
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			n++
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if n == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
}

func TestCatalogStreams(t *testing.T) {
	c := huggingfacedatasets.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "hugging-face-datasets" {
		t.Fatalf("catalog connector = %q", cat.Connector)
	}
	want := map[string]bool{"splits": false, "sizes": false, "rows": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	c := huggingfacedatasets.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("hugging-face-datasets"); !ok {
		t.Fatal("registry did not resolve hugging-face-datasets (self-registration)")
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs to bound SSRF risk.
func TestBaseURLValidation(t *testing.T) {
	c := huggingfacedatasets.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": "file:///etc/passwd", "dataset_name": "x/y"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "splits", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url")
	}
}

// guard against an accidental import drop of strconv in future edits.
var _ = strconv.Itoa
