package httpsource

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

func TestSourceImplementsConnectorAndStatefulReader(t *testing.T) {
	var _ connectors.Connector = Source{}
	var _ connectors.StatefulReader = Source{}
}

func TestReadHarvestsPagesWithBearerAuth(t *testing.T) {
	var sawAuth []string
	var sawPages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/widgets" {
			http.NotFound(w, r)
			return
		}
		sawAuth = append(sawAuth, r.Header.Get("Authorization"))
		sawPages = append(sawPages, r.URL.Query().Get("page"))
		if got := r.URL.Query().Get("per_page"); got != "2" {
			t.Errorf("per_page = %q, want 2", got)
		}

		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"items":[{"id":"w1","updated_at":"2026-01-01T00:00:00Z"},{"id":"w2","updated_at":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"items":[{"id":"w3","updated_at":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	}))
	defer srv.Close()

	src := New(Spec{
		Name:            "demo",
		DisplayName:     "Demo",
		Description:     "Reads demo widgets.",
		DefaultBaseURL:  "https://api.example.test",
		DefaultPageSize: 100,
		MaxPageSize:     100,
		Auth:            AuthSpec{Type: AuthBearer, SecretName: "access_token"},
		Streams: []StreamSpec{{
			Name:         "widgets",
			Description:  "Widgets.",
			Path:         "widgets",
			RecordsPath:  "items",
			Fields:       []connectors.Field{{Name: "id", Type: "string"}, {Name: "updated_at", Type: "timestamp"}},
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Paginator: func(pageSize int) connsdk.Paginator {
				return &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize: pageSize}
			},
		}},
	})
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"access_token": "token_123"},
	}

	var got []connectors.Record
	err := src.Read(context.Background(), connectors.ReadRequest{Stream: "widgets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if !reflect.DeepEqual(sawPages, []string{"1", "2"}) {
		t.Fatalf("pages = %v, want [1 2]", sawPages)
	}
	for _, auth := range sawAuth {
		if auth != "Bearer token_123" {
			t.Fatalf("Authorization = %q, want bearer token", auth)
		}
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[2]["id"] != "w3" {
		t.Fatalf("last record = %+v, want id w3", got[2])
	}
}

func TestAPIKeyHeaderAndMaxPages(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.Header.Get("X-API-Key"); got != "key_123" {
			t.Errorf("X-API-Key = %q, want key_123", got)
		}
		_, _ = w.Write([]byte(fmt.Sprintf(`{"data":[{"id":"row_%d"}]}`, requests)))
	}))
	defer srv.Close()

	src := New(Spec{
		Name:            "apikeydemo",
		DefaultBaseURL:  srv.URL,
		DefaultPageSize: 1,
		MaxPageSize:     10,
		Auth:            AuthSpec{Type: AuthAPIKeyHeader, Header: "X-API-Key"},
		Streams: []StreamSpec{{
			Name:        "rows",
			Path:        "rows",
			RecordsPath: "data",
			PrimaryKey:  []string{"id"},
			Paginator: func(pageSize int) connsdk.Paginator {
				return &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize: pageSize}
			},
		}},
	})
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"max_pages": "1"},
		Secrets: map[string]string{"api_key": "key_123"},
	}

	var got []connectors.Record
	err := src.Read(context.Background(), connectors.ReadRequest{Stream: "rows", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1 due to max_pages", requests)
	}
	if len(got) != 1 || got[0]["id"] != "row_1" {
		t.Fatalf("records = %+v, want first page only", got)
	}
}

func TestFixtureModeCheckReadStateAndWrite(t *testing.T) {
	src := New(Spec{
		Name: "fixturedemo",
		Auth: AuthSpec{Type: AuthBearer, SecretName: "access_token"},
		Streams: []StreamSpec{{
			Name:         "widgets",
			Path:         "widgets",
			Fields:       []connectors.Field{{Name: "id", Type: "string"}, {Name: "updated_at", Type: "timestamp"}},
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
		}},
	})
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	if err := src.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := src.Read(context.Background(), connectors.ReadRequest{Stream: "widgets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["fixture"] != true || got[0]["connector"] != "fixturedemo" || got[0]["id"] == nil {
		t.Fatalf("fixture record missing stable fields: %+v", got[0])
	}

	state, err := src.InitialState(context.Background(), "widgets", connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("InitialState: %v", err)
	}
	if state["stream"] != "widgets" || state[connsdk.CursorStateKey] != "" {
		t.Fatalf("state = %+v, want stream and empty cursor", state)
	}
	if _, err := src.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}

func TestValidationErrors(t *testing.T) {
	src := New(Spec{
		Name:            "validatedemo",
		DefaultBaseURL:  "https://api.example.test",
		DefaultPageSize: 50,
		MaxPageSize:     100,
		Auth:            AuthSpec{Type: AuthBearer, SecretName: "access_token"},
		Streams:         []StreamSpec{{Name: "widgets", Path: "widgets", RecordsPath: "items"}},
	})

	tests := []struct {
		name string
		cfg  connectors.RuntimeConfig
		want string
	}{
		{name: "bad base_url scheme", cfg: connectors.RuntimeConfig{Config: map[string]string{"base_url": "file:///etc/passwd"}, Secrets: map[string]string{"access_token": "token"}}, want: "base_url"},
		{name: "bad page_size", cfg: connectors.RuntimeConfig{Config: map[string]string{"page_size": "0"}, Secrets: map[string]string{"access_token": "token"}}, want: "page_size"},
		{name: "bad max_pages", cfg: connectors.RuntimeConfig{Config: map[string]string{"max_pages": "sometimes"}, Secrets: map[string]string{"access_token": "token"}}, want: "max_pages"},
		{name: "missing secret", cfg: connectors.RuntimeConfig{}, want: "access_token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := src.Check(context.Background(), tt.cfg)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Check error = %v, want substring %q", err, tt.want)
			}
		})
	}
}
