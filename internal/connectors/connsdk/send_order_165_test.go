package connsdk

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"sync"
	"testing"
)

type sendAdmission165 func(context.Context, RateLimitRequest) error

func (f sendAdmission165) AdmitSend(ctx context.Context, request RateLimitRequest) error {
	return f(ctx, request)
}

type routeAdmission165 func(context.Context, RateLimitRoute) (string, error)

func (f routeAdmission165) AdmitRoute(ctx context.Context, route RateLimitRoute) (string, error) {
	return f(ctx, route)
}
func (f routeAdmission165) ObserveRoute(context.Context, RateLimitRoute, RateLimitObservation) {}

func TestRequesterCallerCapPrecedesProviderAdmission165(t *testing.T) {
	for _, mode := range []string{"json", "form", "multipart", "stream"} {
		for _, cut := range []string{"caller", "provider", "route", "allowed"} {
			t.Run(mode+"/"+cut, func(t *testing.T) {
				const expected = `{"records":[{"id":1},{"id":2}]}`
				var mu sync.Mutex
				var observed []string
				record := func(s string) { mu.Lock(); defer mu.Unlock(); observed = append(observed, s) }
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					record("wire:" + r.Method + ":" + r.URL.Path)
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, expected)
				}))
				defer server.Close()
				callerCause := errors.New("caller hard cap exhausted")
				providerCause := errors.New("provider reservation denied")
				refused := &RateBudgetRefusalError{Code: RateBudgetRefusalReservationDenied, Reason: "fixture_refusal", Err: providerCause}
				requester := &Requester{BaseURL: server.URL, DisableRetries: true,
					SendAdmission: sendAdmission165(func(_ context.Context, r RateLimitRequest) error {
						record("caller")
						if r.Attempt != 1 {
							t.Errorf("initial attempt=%d", r.Attempt)
						}
						if cut == "caller" {
							return callerCause
						}
						return nil
					}),
					Admission: rateLimitAdmissionFunc(func(_ context.Context, r RateLimitRequest) error {
						record("provider")
						if r.Attempt != 1 {
							t.Errorf("initial provider attempt=%d", r.Attempt)
						}
						if cut == "provider" {
							return refused
						}
						return nil
					}),
				}
				requester.RouteRateLimits = routeAdmission165(func(_ context.Context, route RateLimitRoute) (string, error) {
					record("route")
					if route.Path != "/resource" || route.Attempt != 1 {
						t.Errorf("wrong actual declared route: %+v", route)
					}
					if cut == "route" {
						return "", refused
					}
					return "", nil
				})
				var body []byte
				var err error
				switch mode {
				case "stream":
					var response *StreamResponse
					response, err = requester.DoStream(t.Context(), http.MethodGet, "/resource", nil, StreamOptions{})
					if response != nil {
						defer response.Body.Close()
						body, err = io.ReadAll(io.LimitReader(response.Body, 1024))
					}
				default:
					var response *Response
					switch mode {
					case "json":
						response, err = requester.Do(t.Context(), http.MethodPost, "/resource", nil, map[string]string{"name": "fixture"})
					case "form":
						response, err = requester.DoForm(t.Context(), http.MethodPost, "/resource", nil, url.Values{"name": {"fixture"}})
					case "multipart":
						response, err = requester.DoMultipart(t.Context(), http.MethodPost, "/resource", nil, MultipartForm{Fields: map[string]string{"name": "fixture"}})
					}
					if response != nil {
						body = response.Body
					}
				}
				expectedOrder := []string{"caller"}
				switch cut {
				case "caller":
					if !errors.Is(err, callerCause) {
						t.Fatalf("caller cause lost: %v", err)
					}
				case "provider":
					expectedOrder = append(expectedOrder, "provider")
					var typed *RateBudgetRefusalError
					if !errors.As(err, &typed) || typed != refused || !errors.Is(err, providerCause) {
						t.Fatalf("typed provider refusal lost: %v", err)
					}
				case "route":
					expectedOrder = append(expectedOrder, "provider", "route")
					var typed *RateBudgetRefusalError
					if !errors.As(err, &typed) || typed != refused || !errors.Is(err, providerCause) {
						t.Fatalf("typed route refusal lost: %v", err)
					}
				case "allowed":
					method := http.MethodPost
					if mode == "stream" {
						method = http.MethodGet
					}
					expectedOrder = append(expectedOrder, "provider", "route", "wire:"+method+":/resource")
					if err != nil || string(body) != expected {
						t.Fatalf("returned bytes=%q error=%v", body, err)
					}
				}
				mu.Lock()
				actual := slices.Clone(observed)
				mu.Unlock()
				if !slices.Equal(actual, expectedOrder) {
					t.Fatalf("admission/send order=%v want=%v", actual, expectedOrder)
				}
				if cut != "allowed" && len(body) != 0 {
					t.Fatalf("refusal returned body %q", body)
				}
			})
		}
	}
}
