package engine

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
)

func TestSelectedMalformedBundlePhysicalBoundary168(t *testing.T) {
	var mu sync.Mutex
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":1,"updated_at":"2026-09-08"},{"id":2,"updated_at":"2026-09-08"}]}`))
	}))
	defer server.Close()
	run := func(files fs.FS) ([]connectors.Record, error) {
		bundle, err := Load(files, "acme")
		if err != nil {
			return nil, err
		}
		var rows []connectors.Record
		err = New(bundle, nil).Read(t.Context(), connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}, Secrets: map[string]string{"token": "fixture-168"}}, MaxRequests: 1}, func(row connectors.Record) error { rows = append(rows, row); return nil })
		return rows, err
	}
	fixture := func() fstest.MapFS {
		files := fullValidBundleFS("acme")
		files["acme/streams.json"] = &fstest.MapFile{Data: []byte(strings.Replace(validStreams, `, "when": "{{ cursor }}"`, "", 1))}
		return files
	}
	good := fixture()
	rows, err := run(good)
	if err != nil || len(rows) != 2 {
		t.Fatalf("healthy same executor did not return two actual records: rows=%v error=%v", rows, err)
	}
	encoded, encodeErr := json.Marshal(rows)
	if encodeErr != nil || string(encoded) != `[{"id":1,"updated_at":"2026-09-08"},{"id":2,"updated_at":"2026-09-08"}]` {
		t.Fatalf("wrong returned row identities: %v", rows)
	}
	mu.Lock()
	initial := append([]string(nil), requests...)
	mu.Unlock()
	if !reflect.DeepEqual(initial, []string{"GET /widgets"}) {
		t.Fatalf("wrong physical request: %v", initial)
	}
	malformed := fixture()
	malformed["acme/rate_limits.json"] = &fstest.MapFile{Data: []byte(`{"state":}`)}
	rows, err = run(malformed)
	var d *BundleDiagnosticError
	if !errors.As(err, &d) || d.File != "rate_limits.json" || d.ReasonCode != "invalid_json" || len(rows) != 0 {
		t.Fatalf("selected failure bypassed loading: rows=%v error=%v", rows, err)
	}
	mu.Lock()
	final := append([]string(nil), requests...)
	mu.Unlock()
	if !reflect.DeepEqual(final, initial) {
		t.Fatalf("malformed selection reached physical provider fixture: %v", final)
	}
}
