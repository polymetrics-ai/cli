package engine

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

func TestDeclarativeWriteRetryPolicy(t *testing.T) {
	t.Run("single attempt without provider key", func(t *testing.T) {
		var attempts int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			attempts++
			http.Error(w, "ambiguous", http.StatusInternalServerError)
		}))
		defer srv.Close()

		rt := &Runtime{Requester: &connsdk.Requester{
			BaseURL: srv.URL,
			Sleep:   func(context.Context, time.Duration) error { return nil },
		}}
		action := WriteAction{Name: "create_widget", Method: http.MethodPost, Path: "/widgets"}
		err := executeWriteRecord(context.Background(), newWriteTestBundle(srv, action), action,
			connectors.Record{"name": "fixture"}, 0, connectors.RuntimeConfig{}, rt)
		if err == nil {
			t.Fatal("executeWriteRecord succeeded after an ambiguous mutation response")
		}
		if attempts != 1 {
			t.Fatalf("attempts = %d, want 1", attempts)
		}
	})

	t.Run("explicitly idempotent delete retries without provider key", func(t *testing.T) {
		var attempts int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			attempts++
			if attempts == 1 {
				http.Error(w, "retry", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		rt := &Runtime{Requester: &connsdk.Requester{
			BaseURL: srv.URL,
			Sleep:   func(context.Context, time.Duration) error { return nil },
		}}
		action := WriteAction{
			Name: "delete_widget", Kind: "delete", Method: http.MethodDelete, Path: "/widgets/fixture",
			Delete: &DeleteSpec{Idempotent: true},
		}
		if err := executeWriteRecord(context.Background(), newWriteTestBundle(srv, action), action,
			connectors.Record{}, 0, connectors.RuntimeConfig{}, rt); err != nil {
			t.Fatalf("executeWriteRecord: %v", err)
		}
		if attempts != 2 {
			t.Fatalf("attempts = %d, want 2", attempts)
		}
	})

	t.Run("unmarked delete remains single attempt", func(t *testing.T) {
		var attempts int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			attempts++
			http.Error(w, "ambiguous", http.StatusInternalServerError)
		}))
		defer srv.Close()

		rt := &Runtime{Requester: &connsdk.Requester{
			BaseURL: srv.URL,
			Sleep:   func(context.Context, time.Duration) error { return nil },
		}}
		action := WriteAction{
			Name: "delete_widget", Kind: "delete", Method: http.MethodDelete, Path: "/widgets/fixture",
			Delete: &DeleteSpec{},
		}
		err := executeWriteRecord(context.Background(), newWriteTestBundle(srv, action), action,
			connectors.Record{}, 0, connectors.RuntimeConfig{}, rt)
		if err == nil {
			t.Fatal("executeWriteRecord succeeded after an ambiguous delete response")
		}
		if attempts != 1 {
			t.Fatalf("attempts = %d, want 1", attempts)
		}
	})

	t.Run("stable provider key across retries", func(t *testing.T) {
		var attempts int
		var keys []string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			keys = append(keys, r.Header.Get("Idempotency-Key"))
			if attempts == 1 {
				http.Error(w, "retry", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		rt := &Runtime{Requester: &connsdk.Requester{
			BaseURL:        srv.URL,
			DefaultHeaders: map[string]string{"idempotency-key": "stale"},
			Sleep:          func(context.Context, time.Duration) error { return nil },
		}}
		action := WriteAction{
			Name: "create_widget", Method: http.MethodPost, Path: "/widgets",
			IdempotencyKeyHeader: "Idempotency-Key",
		}
		if err := executeWriteRecord(context.Background(), newWriteTestBundle(srv, action), action,
			connectors.Record{"name": "fixture"}, 0, connectors.RuntimeConfig{}, rt); err != nil {
			t.Fatalf("executeWriteRecord: %v", err)
		}
		if attempts != 2 {
			t.Fatalf("attempts = %d, want 2", attempts)
		}
		if keys[0] == "" || keys[0] == "stale" || keys[0] != keys[1] {
			t.Fatalf("idempotency keys = %#v, want one stable non-empty value", keys)
		}
	})

	t.Run("fresh provider key per record", func(t *testing.T) {
		var keys []string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			keys = append(keys, r.Header.Get("Idempotency-Key"))
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		rt := &Runtime{Requester: &connsdk.Requester{BaseURL: srv.URL}}
		action := WriteAction{
			Name: "create_widget", Method: http.MethodPost, Path: "/widgets",
			IdempotencyKeyHeader: "Idempotency-Key",
		}
		for i := 0; i < 2; i++ {
			if err := executeWriteRecord(context.Background(), newWriteTestBundle(srv, action), action,
				connectors.Record{"name": "fixture"}, i, connectors.RuntimeConfig{}, rt); err != nil {
				t.Fatalf("executeWriteRecord %d: %v", i, err)
			}
		}
		if len(keys) != 2 || keys[0] == "" || keys[1] == "" || keys[0] == keys[1] {
			t.Fatalf("idempotency keys = %#v, want two distinct non-empty values", keys)
		}
	})
}

func TestDeclarativeWriteIdempotencyRetriesOriginalURLOnly(t *testing.T) {
	var attempts int
	var keys []string
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		paths = append(paths, r.URL.RequestURI())
		if attempts == 1 {
			w.Header().Set("X-Request-ID", "first-failure")
			http.Error(w, "retry", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("X-Request-ID", "terminal-success")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	action := WriteAction{
		Name: "create_widget", Method: http.MethodPost, Path: "/widgets",
		IdempotencyKeyHeader: "Idempotency-Key",
	}
	bundle := newWriteTestBundle(srv, action)
	records := []connectors.Record{{"name": "fixture"}}
	req := connectors.WriteRequest{Action: action.Name}
	preview, err := DryRunWrite(context.Background(), bundle, req, records, nil)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	if _, err := Write(context.Background(), bundle, req, records, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}
	wantKey := writeIdempotencyKey(bundle.Name, action, preview.Digest, "", 0)
	if wantKey == "" || len(keys) != 2 || keys[0] != wantKey || keys[1] != wantKey {
		t.Fatalf("idempotency keys = %#v, want preview-bound %q", keys, wantKey)
	}
	if len(paths) != 2 || paths[0] != "/widgets" || paths[1] != "/widgets" {
		t.Fatalf("retry paths = %#v, want original URL only", paths)
	}
}

func TestRequiredJSONWriteSendsEmptyObject(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	action := WriteAction{
		Name: "update_widget", Method: http.MethodPut, Path: "/widgets/fixture",
		BodyType: "json", BodyRequired: true,
	}
	rt := &Runtime{Requester: &connsdk.Requester{BaseURL: srv.URL}}
	if err := executeWriteRecord(context.Background(), newWriteTestBundle(srv, action), action,
		connectors.Record{}, 0, connectors.RuntimeConfig{}, rt); err != nil {
		t.Fatalf("executeWriteRecord: %v", err)
	}
	if string(body) != "{}" {
		t.Fatalf("body = %q, want {}", body)
	}
}
