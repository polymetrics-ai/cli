package credential_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/credential"
)

type credentialTransport210 func(*http.Request) (*http.Response, error)

func (f credentialTransport210) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// The observer examines raw request headers before net/http transport may
// normalize surrounding whitespace; form values are compared after decoding.
func credentialMatches210(r *http.Request, name, value string) bool {
	if name == "safetyculture" {
		return r.Header.Get("Authorization") == "Bearer "+value
	}
	if err := r.ParseForm(); err != nil {
		return false
	}
	return r.PostForm.Get("apiKey") == value && len(r.PostForm["apiKey"]) == 1
}

func testRegisteredCredentialBytes210(t *testing.T, name string) {
	t.Helper()
	key, stream := "api_key", "boards"
	if name == "safetyculture" {
		key, stream = "access_token", "audits"
	}
	for _, tc := range []struct {
		name, value     string
		supplied, valid bool
		empty, unsafe   bool
	}{
		{"leading_space", " fixture", true, true, false, false},
		{"trailing_space", "fixture ", true, true, false, false},
		{"surrounding_spaces", " fixture ", true, true, false, false},
		{"absent", "", false, false, false, false},
		{"explicit_empty", "", true, false, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testCredentialRouteCase210(t, name, key, stream, tc.value, tc.supplied, tc.valid, tc.empty, tc.unsafe)
		})
	}
	if name == "safetyculture" {
		for _, tc := range []struct{ name, value string }{{"CRLF_header", "fixture\r\ninvalid"}, {"NUL_header", "fixture\x00invalid"}} {
			t.Run(tc.name, func(t *testing.T) {
				testCredentialRouteCase210(t, name, key, stream, tc.value, true, false, false, true)
			})
		}
	}
}

func testCredentialRouteCase210(t *testing.T, name, key, stream, value string, supplied, valid, wantEmpty, wantUnsafe bool) {
	t.Helper()
	for _, method := range []string{"Check", "Read"} {
		t.Run(method, func(t *testing.T) {
			registry := bundleregistry.New()
			connector, ok := registry.Get(name)
			if !ok {
				t.Fatal("registered credential route missing")
			}
			previous := http.DefaultTransport
			t.Cleanup(func() { http.DefaultTransport = previous })
			var sends atomic.Int64
			http.DefaultTransport = credentialTransport210(func(r *http.Request) (*http.Response, error) {
				sends.Add(1)
				if !valid {
					t.Error("invalid credential reached transport")
				}
				if !credentialMatches210(r, name, value) {
					t.Error("registered route changed exact credential bytes")
				}
				body := `{"boards":[{"id":"board-1","created":"2026-01-02T00:00:00Z"}],"hasMore":false}`
				if name == "safetyculture" {
					if r.URL.Host != "api.safetyculture.io" || r.URL.Path != "/audits" || r.Method != http.MethodGet {
						t.Error("registered SafetyCulture route changed")
					}
					body = `{"audits":[{"id":"audit-1","name":"Example","modified_at":"2026-01-01T00:00:00Z"}],"links":{"next":""}}`
				} else if r.URL.Host != "canny.io" || r.URL.Path != "/api/v1/boards/list" || r.Method != http.MethodPost || !strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
					t.Error("registered Canny form route changed")
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
			})
			cfg := connectors.RuntimeConfig{Secrets: map[string]string{}}
			if supplied {
				cfg.Secrets[key] = value
			}
			var err error
			var records []connectors.Record
			if method == "Check" {
				err = connector.Check(context.Background(), cfg)
			} else {
				err = connector.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(r connectors.Record) error { records = append(records, r); return nil })
			}
			if valid {
				if err != nil || sends.Load() != 1 {
					t.Fatal("valid registered route did not perform exactly one bounded request")
				}
				if method == "Read" {
					want := "board-1"
					if name == "safetyculture" {
						want = "audit-1"
					}
					if len(records) != 1 || records[0]["id"] != want {
						t.Fatal("registered read did not return independently expected record")
					}
				}
			} else {
				if err == nil || sends.Load() != 0 || len(records) != 0 {
					t.Fatal("invalid credential did not refuse before request/record side effects")
				}
				var empty *credential.EmptySecretError
				var unsafe *credential.InvalidSecretValueError
				if errors.As(err, &empty) != wantEmpty || errors.As(err, &unsafe) != wantUnsafe {
					t.Fatal("absent, empty or unsafe credential classification changed")
				}
			}
			if value != "" && err != nil && strings.Contains(err.Error(), value) {
				t.Fatal("credential value disclosed by diagnostic")
			}
		})
	}
}

func TestRegisteredCredentialObserverRejectsTrimmedBytes210(t *testing.T) {
	const expected = " fixture "
	for _, name := range []string{"safetyculture", "canny"} {
		for _, tc := range []struct {
			name, value string
			want        bool
		}{{"exact", expected, true}, {"trimmed_left", strings.TrimLeft(expected, " "), false}, {"trimmed_right", strings.TrimRight(expected, " "), false}, {"trimmed_both", strings.TrimSpace(expected), false}} {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				body := url.Values{"apiKey": {tc.value}}.Encode()
				r, err := http.NewRequest(http.MethodPost, "https://fixture.invalid", strings.NewReader(body))
				if err != nil {
					t.Fatal("construct observer counterexample")
				}
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				r.Header.Set("Authorization", "Bearer "+tc.value)
				if credentialMatches210(r, name, expected) != tc.want {
					t.Fatal("exact-byte observer accepted a trimming mutation")
				}
			})
		}
	}
}
