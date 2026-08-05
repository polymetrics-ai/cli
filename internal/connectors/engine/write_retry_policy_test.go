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
