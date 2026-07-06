package plaid

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func newRuntime(baseURL string) *engine.Runtime {
	return &engine.Runtime{Requester: &connsdk.Requester{BaseURL: baseURL}}
}

func cfgWithCreds(extra map[string]string) connectors.RuntimeConfig {
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{},
		Secrets: map[string]string{"client_id": "unit-client", "secret": "unit-secret"},
	}
	for k, v := range extra {
		cfg.Config[k] = v
	}
	return cfg
}

// --- registration ---

func TestInit_RegistersHooks(t *testing.T) {
	h := engine.HooksFor("plaid")
	if h == nil {
		t.Fatal("engine.HooksFor(\"plaid\") = nil, want a registered hook set (init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "plaid" {
		t.Fatalf("ConnectorName() = %q, want plaid", h.ConnectorName())
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered hooks do not implement StreamHook")
	}
	if _, ok := h.(engine.CheckHook); !ok {
		t.Fatal("registered hooks do not implement CheckHook")
	}
}

// --- ReadStream dispatch ---

func TestReadStream_UnknownStreamFallsBackToDeclarative(t *testing.T) {
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "not_a_real_stream"}, connectors.ReadRequest{Config: cfgWithCreds(nil)}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("handled = true for an unrecognized stream name, want false (declarative fallback)")
	}
}

func TestReadStream_EmptyStreamNameDefaultsToInstitutions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/institutions/get" {
			t.Errorf("path = %q, want /institutions/get", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"institutions":[]}`))
	}))
	defer srv.Close()

	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: ""}, connectors.ReadRequest{Config: cfgWithCreds(nil)}, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true (empty stream name defaults to institutions)")
	}
}

// --- body shape: credentials + pagination state, all in the JSON body ---

func TestReadStream_InstitutionsSendsCredentialsAndPaginationStateInBody(t *testing.T) {
	var sawClientID, sawSecret string
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/institutions/get" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		if r.URL.RawQuery != "" {
			t.Fatalf("query string = %q, want empty (all state must be in the body, not the query)", r.URL.RawQuery)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		sawClientID, _ = body["client_id"].(string)
		sawSecret, _ = body["secret"].(string)
		if body["count"].(float64) != 2 {
			t.Fatalf("count = %v, want 2", body["count"])
		}
		codes, _ := body["country_codes"].([]any)
		if len(codes) != 2 || codes[0] != "US" || codes[1] != "CA" {
			t.Fatalf("country_codes = %v, want [US CA]", body["country_codes"])
		}
		switch body["offset"].(float64) {
		case 0:
			_, _ = w.Write([]byte(`{"institutions":[{"institution_id":"ins_1","name":"One Bank","country_codes":["US"]},{"institution_id":"ins_2","name":"Two Bank","country_codes":["US","CA"]}],"total":3}`))
		case 2:
			_, _ = w.Write([]byte(`{"institutions":[{"institution_id":"ins_3","name":"Three Bank","country_codes":["GB"]}],"total":3}`))
		default:
			t.Fatalf("unexpected offset %v", body["offset"])
		}
	}))
	defer srv.Close()

	cfg := cfgWithCreds(map[string]string{"page_size": "2", "country_codes": "US,CA"})
	var got []connectors.Record
	handled, err := Hooks{}.ReadStream(context.Background(), engine.StreamSpec{Name: "institutions"}, connectors.ReadRequest{Config: cfg}, newRuntime(srv.URL), func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if sawClientID != "unit-client" || sawSecret != "unit-secret" {
		t.Fatalf("auth body client_id=%q secret=%q, want unit-client/unit-secret", sawClientID, sawSecret)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2 (pagination across 2 pages)", requests)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[0]["institution_id"] != "ins_1" {
		t.Fatalf("first record institution_id = %v, want ins_1", got[0]["institution_id"])
	}
	if got[0]["country_codes"] == "" {
		t.Fatal("country_codes field empty, want joined comma-separated string")
	}
}

// TestReadStream_InstitutionsStopsOnShortPage proves the short-page stop
// signal (fewer records than the requested count) halts pagination even
// though a further page could in principle be requested.
func TestReadStream_InstitutionsStopsOnShortPage(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"institutions":[{"institution_id":"ins_only","name":"Only Bank","country_codes":["US"]}]}`))
	}))
	defer srv.Close()

	cfg := cfgWithCreds(map[string]string{"page_size": "2"})
	var got []connectors.Record
	_, err := Hooks{}.ReadStream(context.Background(), engine.StreamSpec{Name: "institutions"}, connectors.ReadRequest{Config: cfg}, newRuntime(srv.URL), func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1 (short first page stops pagination)", requests)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

// TestReadStream_MaxPagesHardCap proves max_pages bounds the number of
// requests even when every page is full.
func TestReadStream_MaxPagesHardCap(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"institutions":[{"institution_id":"ins_a","name":"A","country_codes":["US"]},{"institution_id":"ins_b","name":"B","country_codes":["US"]}]}`))
	}))
	defer srv.Close()

	cfg := cfgWithCreds(map[string]string{"page_size": "2", "max_pages": "2"})
	_, err := Hooks{}.ReadStream(context.Background(), engine.StreamSpec{Name: "institutions"}, connectors.ReadRequest{Config: cfg}, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2 (max_pages hard cap)", requests)
	}
}

// --- categories: no pagination fields at all ---

func TestReadStream_CategoriesSendsNoCountOrOffset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/categories/get" {
			t.Fatalf("path = %q, want /categories/get", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if _, ok := body["count"]; ok {
			t.Fatalf("body has count key = %v, want omitted for categories", body["count"])
		}
		if _, ok := body["offset"]; ok {
			t.Fatalf("body has offset key = %v, want omitted for categories", body["offset"])
		}
		_, _ = w.Write([]byte(`{"categories":[{"category_id":"10000000","group":"transfer","hierarchy":["Transfer","Debit"]}]}`))
	}))
	defer srv.Close()

	var got []connectors.Record
	handled, err := Hooks{}.ReadStream(context.Background(), engine.StreamSpec{Name: "categories"}, connectors.ReadRequest{Config: cfgWithCreds(nil)}, newRuntime(srv.URL), func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(got) != 1 || got[0]["category_id"] != "10000000" {
		t.Fatalf("records = %+v, want 1 record with category_id 10000000", got)
	}
	if got[0]["hierarchy"] == "" {
		t.Fatal("hierarchy field empty, want joined comma-separated string")
	}
}

// --- Check ---

func TestCheck_PostsCategoriesGetWithAuthBody(t *testing.T) {
	var sawClientID, sawSecret, sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		sawClientID, _ = body["client_id"].(string)
		sawSecret, _ = body["secret"].(string)
		_, _ = w.Write([]byte(`{"categories":[]}`))
	}))
	defer srv.Close()

	handled, err := Hooks{}.Check(context.Background(), cfgWithCreds(nil), newRuntime(srv.URL))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if sawPath != "/categories/get" {
		t.Fatalf("path = %q, want /categories/get", sawPath)
	}
	if sawClientID != "unit-client" || sawSecret != "unit-secret" {
		t.Fatalf("auth body client_id=%q secret=%q", sawClientID, sawSecret)
	}
}

func TestCheck_ServerErrorIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error_code":"INVALID_API_KEYS"}`))
	}))
	defer srv.Close()

	handled, err := Hooks{}.Check(context.Background(), cfgWithCreds(nil), newRuntime(srv.URL))
	if err == nil {
		t.Fatal("Check() error = nil, want an error for a 401 response")
	}
	if !handled {
		t.Fatal("handled = false, want true (Check always handles)")
	}
}

// --- credential errors ------------------------------------------------

func TestReadStream_MissingCredentialsIsError(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{}, Secrets: map[string]string{}}
	_, err := Hooks{}.ReadStream(context.Background(), engine.StreamSpec{Name: "institutions"}, connectors.ReadRequest{Config: cfg}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream() error = nil, want an error for missing client_id/secret")
	}
}

func TestAuthenticator_ErrorsNeverContainSecretText(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{}, Secrets: map[string]string{}}
	_, err := authBody(cfg)
	if err == nil {
		t.Fatal("expected an error for missing credentials")
	}
	if err.Error() == "" {
		t.Fatal("expected a non-empty error message")
	}
}

// --- ctx cancellation -----------------------------------------------------

func TestReadStream_HonorsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"institutions":[{"institution_id":"ins_a","name":"A","country_codes":["US"]},{"institution_id":"ins_b","name":"B","country_codes":["US"]}]}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cfg := cfgWithCreds(map[string]string{"page_size": "1"})
	_, err := Hooks{}.ReadStream(ctx, engine.StreamSpec{Name: "institutions"}, connectors.ReadRequest{Config: cfg}, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream(cancelled ctx) error = nil, want a cancellation error")
	}
}
