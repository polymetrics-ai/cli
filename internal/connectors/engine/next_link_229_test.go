package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"
)

// Literal CP21 page positions and named identities, independent of composition.
func TestNextLinkPhysicalPages229(t *testing.T) {
	for _, scenario := range []string{"healthy", "page_two_404", "repeated_next", "malformed_next"} {
		t.Run(scenario, func(t *testing.T) {
			var mu sync.Mutex
			var requests []string
			var origin string
			faultReached := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				requests = append(requests, r.RequestURI)
				w.Header().Set("Content-Type", "application/json")
				switch r.RequestURI {
				case "/widgets?page=1&page_size=2":
					_, _ = fmt.Fprintf(w, `{"data":[{"id":"alpha"},{"id":"bravo"}],"next":%q}`, origin+"/widgets?page=2&page_size=2")
				case "/widgets?page=2&page_size=2":
					faultReached = true
					switch scenario {
					case "page_two_404":
						w.WriteHeader(404)
					case "repeated_next":
						_, _ = fmt.Fprintf(w, `{"data":[{"id":"charlie"}],"next":%q}`, origin+"/widgets?page=2&page_size=2")
					case "malformed_next":
						_, _ = fmt.Fprint(w, `{"data":[{"id":"charlie"}],"next":"http://%"}`)
					default:
						_, _ = fmt.Fprint(w, `{"data":[{"id":"charlie"}],"next":null}`)
					}
				default:
					w.WriteHeader(400)
				}
			}))
			defer server.Close()
			origin = server.URL
			b := newTestBundle(t, server, StreamSpec{Query: map[string]QueryParam{"page": {Template: "1"}, "page_size": {Template: "2"}}, Pagination: &PaginationSpec{Type: "next_url", NextURLPath: "next", PageParam: "page", SizeParam: "page_size", PageSize: 2}})
			rows, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
			mu.Lock()
			defer mu.Unlock()
			t.Logf("actual physical requests=%q page_two_reached=%v rows=%v error=%v", requests, faultReached, rows, err)
			if !reflect.DeepEqual(requests, []string{"/widgets?page=1&page_size=2", "/widgets?page=2&page_size=2"}) || !faultReached {
				t.Fatalf("second physical request not reached: %q", requests)
			}
			if scenario != "healthy" {
				if err == nil {
					t.Fatal("downstream fault accepted")
				}
				return
			}
			ids := []string{}
			for _, r := range rows {
				ids = append(ids, fmt.Sprint(r["id"]))
			}
			if err != nil || !reflect.DeepEqual(ids, []string{"alpha", "bravo", "charlie"}) {
				t.Fatalf("complete identities=%v err=%v", ids, err)
			}
		})
	}
}

func TestNextLinkQueryOwnership229(t *testing.T) {
	for _, mode := range []string{"saved", "direct"} {
		for _, tc := range []struct {
			name, next, want string
			policy           *NextURLQuerySpec
		}{
			{"complete", "?page=2&page_size=3", "?page=2&page_size=3", nil},
			{"missing_position", "?page_size=3", "?page_size=3", nil},
			{"retained_missing", "?page=2", "?page=2&filter=initial&page_size=2", &NextURLQuerySpec{Allowed: []string{"filter", "page_size"}, Retain: []string{"filter", "page_size"}}},
			{"link_wins", "?page=2&filter=next&page_size=3", "?page=2&filter=next&page_size=3", &NextURLQuerySpec{Allowed: []string{"filter", "page_size"}, Retain: []string{"filter", "page_size"}}},
			{"signed_ordered_encoded", "?signature=a%2fb+%20&page=2&tag=b&tag=a&tag=b", "?signature=a%2fb+%20&page=2&tag=b&tag=a&tag=b", &NextURLQuerySpec{Allowed: []string{"signature", "tag"}}},
		} {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				var mu sync.Mutex
				var requests []string
				var origin string
				firstURI := "/widgets?filter=initial&page_size=2&private=not-retained"
				if mode == "saved" {
					firstURI = "/widgets?filter=initial&page=1&page_size=2&private=not-retained"
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mu.Lock()
					defer mu.Unlock()
					requests = append(requests, r.RequestURI)
					w.Header().Set("Content-Type", "application/json")
					if len(requests) == 1 && r.RequestURI == firstURI {
						_, _ = fmt.Fprintf(w, `{"data":[{"id":"alpha"},{"id":"bravo"}],"next":%q}`, origin+"/widgets"+tc.next)
						return
					}
					if len(requests) == 2 && r.RequestURI == "/widgets"+tc.want {
						_, _ = fmt.Fprint(w, `{"data":[{"id":"charlie"}],"next":null}`)
						return
					}
					w.WriteHeader(400)
				}))
				defer server.Close()
				origin = server.URL
				spec := &PaginationSpec{Type: "next_url", NextURLPath: "next", PageParam: "page", SizeParam: "page_size", PageSize: 2, NextURLQuery: tc.policy}
				ids := []string{}
				if mode == "saved" {
					b := newTestBundle(t, server, StreamSpec{Pagination: spec, Query: map[string]QueryParam{"page": {Template: "1"}, "page_size": {Template: "2"}, "filter": {Template: "initial"}, "private": {Template: "not-retained"}}})
					rows, err := readAll(t, t.Context(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
					if err != nil {
						t.Fatal(err)
					}
					for _, r := range rows {
						ids = append(ids, fmt.Sprint(r["id"]))
					}
				} else {
					b := paginatedDirectReadBundle(origin, spec, "/widgets")
					cursor := ""
					for i := 0; i < 2; i++ {
						result, err := DirectRead(t.Context(), b, connectors.DirectReadRequest{Method: "GET", Path: "/widgets", Query: map[string]string{"filter": "initial", "private": "not-retained"}, PageCursor: cursor, OutputPolicy: "json_redacted"}, nil)
						if err != nil {
							t.Fatal(err)
						}
						for _, r := range result.Body.(map[string]any)["data"].([]any) {
							ids = append(ids, fmt.Sprint(r.(map[string]any)["id"]))
						}
						mu.Lock()
						sends := len(requests)
						mu.Unlock()
						if sends != i+1 {
							t.Fatal("direct read performed more than one send")
						}
						cursor = result.Page.NextCursor
						if i == 0 && cursor != origin+"/widgets"+tc.want {
							t.Fatalf("reported cursor=%q does not equal effective next request", cursor)
						}
						if i == 1 && tc.name == "complete" && result.Page.Size != 3 {
							t.Fatalf("physical link size3 reported as %d", result.Page.Size)
						}
						if i == 1 && (!result.Page.Complete || cursor != "") {
							t.Fatal("terminal direct page changed")
						}
					}
				}
				mu.Lock()
				defer mu.Unlock()
				t.Logf("physical=%q ids=%q", requests, ids)
				if !reflect.DeepEqual(requests, []string{firstURI, "/widgets" + tc.want}) || !reflect.DeepEqual(ids, []string{"alpha", "bravo", "charlie"}) {
					t.Fatal("wire or exact identities changed")
				}
			})
		}
	}
}

func TestNextLinkEffectiveLoop229(t *testing.T) {
	for _, mode := range []string{"saved", "direct"} {
		for _, scenario := range []string{"initial_self", "query_order", "retained_equivalence", "resumed_self"} {
			t.Run(mode+"/"+scenario, func(t *testing.T) {
				var sends int
				var origin string
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					sends++
					w.Header().Set("Content-Type", "application/json")
					next := "/widgets?filter=initial&page_size=2"
					if scenario == "query_order" {
						next = "/widgets?page_size=2&filter=initial"
					}
					if scenario == "retained_equivalence" {
						next = "/widgets"
					}
					if scenario == "resumed_self" {
						next = "/widgets?page=2&page_size=2&filter=initial"
					}
					_, _ = fmt.Fprintf(w, `{"data":[{"id":"alpha"}],"next":%q}`, origin+next)
				}))
				defer server.Close()
				origin = server.URL
				spec := &PaginationSpec{Type: "next_url", NextURLPath: "next", PageParam: "page", SizeParam: "page_size", PageSize: 2, NextURLQuery: &NextURLQuerySpec{Allowed: []string{"filter", "page_size"}, Retain: []string{"filter", "page_size"}}}
				var err error
				if mode == "saved" {
					b := newTestBundle(t, server, StreamSpec{Pagination: spec, Query: map[string]QueryParam{"page_size": {Template: "2"}, "filter": {Template: "initial"}}})
					req := connectors.ReadRequest{Stream: "widgets"}
					if scenario == "resumed_self" {
						c, e := newReadContinuation(b, b.Streams[0], &connsdk.NextPage{URL: origin + "/widgets?page=2&page_size=2&filter=initial"})
						if e != nil {
							t.Fatal(e)
						}
						req.Continuation = &c
					}
					err = ReadWithOutcome(t.Context(), b, req, nil, func(connectors.Record) error { return nil })
				} else {
					b := paginatedDirectReadBundle(origin, spec, "/widgets")
					cursor := ""
					if scenario == "resumed_self" {
						cursor = origin + "/widgets?page=2&page_size=2&filter=initial"
					}
					_, err = DirectRead(t.Context(), b, connectors.DirectReadRequest{Method: "GET", Path: "/widgets", Query: map[string]string{"filter": "initial"}, PageCursor: cursor, OutputPolicy: "json_redacted"}, nil)
				}
				if err == nil || sends != 1 {
					t.Fatalf("effective loop err=%v sends=%d", err, sends)
				}
				t.Logf("same effective request refused after exactly%d send: %v", sends, err)
			})
		}
	}
}

func TestNextLinkFiveIdentities229(t *testing.T) {
	for _, mode := range []string{"saved", "direct"} {
		for _, fault := range []string{"healthy", "404", "cross_origin", "malformed", "reordered_repeat"} {
			t.Run(mode+"/"+fault, func(t *testing.T) {
				var requests []string
				var origin string
				faultReached := false
				var foreignSends atomic.Int32
				foreign := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { foreignSends.Add(1) }))
				defer foreign.Close()
				first := "/widgets?page_size=2"
				if mode == "saved" {
					first = "/widgets?page=1&page_size=2"
				}
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests = append(requests, r.RequestURI)
					w.Header().Set("Content-Type", "application/json")
					switch r.RequestURI {
					case first:
						_, _ = fmt.Fprintf(w, `{"data":[{"id":"repo-a"},{"id":"repo-b"}],"next":%q}`, origin+"/widgets?page=2&page_size=2")
					case "/widgets?page=2&page_size=2":
						faultReached = true
						next := origin + "/widgets?page=3&page_size=2"
						switch fault {
						case "404":
							w.WriteHeader(404)
							return
						case "cross_origin":
							next = foreign.URL + "/widgets?page=3"
						case "malformed":
							next = "http://%"
						case "reordered_repeat":
							next = origin + "/widgets?page_size=2&page=2"
						}
						_, _ = fmt.Fprintf(w, `{"data":[{"id":"repo-c"},{"id":"repo-d"}],"next":%q}`, next)
					case "/widgets?page=3&page_size=2":
						_, _ = fmt.Fprint(w, `{"data":[{"id":"repo-e"}]}`)
					default:
						w.WriteHeader(400)
					}
				}))
				defer srv.Close()
				origin = srv.URL
				spec := &PaginationSpec{Type: "next_url", NextURLPath: "next", PageParam: "page", SizeParam: "page_size", PageSize: 2}
				ids := []string{}
				var err error
				if mode == "saved" {
					b := newTestBundle(t, srv, StreamSpec{Pagination: spec, Query: map[string]QueryParam{"page": {Template: "1"}, "page_size": {Template: "2"}}})
					rows, e := readAll(t, t.Context(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
					err = e
					for _, r := range rows {
						ids = append(ids, fmt.Sprint(r["id"]))
					}
				} else {
					b := paginatedDirectReadBundle(origin, spec, "/widgets")
					cursor := ""
					for i := 0; i < 3; i++ {
						result, e := DirectRead(t.Context(), b, connectors.DirectReadRequest{Method: "GET", Path: "/widgets", PageCursor: cursor, OutputPolicy: "json_redacted"}, nil)
						err = e
						if e != nil {
							break
						}
						for _, r := range result.Body.(map[string]any)["data"].([]any) {
							ids = append(ids, fmt.Sprint(r.(map[string]any)["id"]))
						}
						cursor = result.Page.NextCursor
						if len(requests) != i+1 {
							t.Fatal("direct exceeded single-page request")
						}
						if cursor == "" {
							break
						}
					}
				}
				want := []string{first, "/widgets?page=2&page_size=2"}
				if fault == "healthy" {
					want = append(want, "/widgets?page=3&page_size=2")
				}
				t.Logf("fault=%s reached_second=%v requests=%q ids=%q err=%v", fault, faultReached, requests, ids, err)
				if !faultReached || !reflect.DeepEqual(requests, want) || foreignSends.Load() != 0 {
					t.Fatal("fault not reached or unauthorized/trailing send")
				}
				if fault == "healthy" {
					if err != nil || !reflect.DeepEqual(ids, []string{"repo-a", "repo-b", "repo-c", "repo-d", "repo-e"}) {
						t.Fatal("five exact identities not returned")
					}
				} else if err == nil {
					t.Fatal("fault accepted")
				}
			})
		}
	}
}

func TestNextLinkAdmissionAndBounds229(t *testing.T) {
	for _, mode := range []string{"saved", "direct"} {
		for _, kind := range []string{"wrong_strategy", "retained_position", "retained_undeclared", "duplicate_allowed", "invalid_name", "canceled", "malformed_cursor", "cross_origin_cursor", "userinfo_cursor", "fragment_cursor"} {
			t.Run(mode+"/"+kind, func(t *testing.T) {
				var sends atomic.Int32
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { sends.Add(1); _, _ = fmt.Fprint(w, `{"data":[]}`) }))
				defer srv.Close()
				spec := &PaginationSpec{Type: "next_url", NextURLPath: "next", PageParam: "position", NextURLQuery: &NextURLQuerySpec{Allowed: []string{"filter"}, Retain: []string{"filter"}}}
				ctx := t.Context()
				cursor := ""
				switch kind {
				case "wrong_strategy":
					spec.Type = "none"
				case "retained_position":
					spec.NextURLQuery = &NextURLQuerySpec{Allowed: []string{"position"}, Retain: []string{"position"}}
				case "retained_undeclared":
					spec.NextURLQuery.Retain = []string{"private"}
				case "duplicate_allowed":
					spec.NextURLQuery.Allowed = []string{"filter", "filter"}
				case "invalid_name":
					spec.NextURLQuery.Allowed = []string{"filter\n"}
				case "canceled":
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					cancel()
				case "malformed_cursor":
					cursor = "http://%"
				case "cross_origin_cursor":
					cursor = "https://foreign.invalid/widgets?position=2"
				case "userinfo_cursor":
					cursor = strings.Replace(srv.URL, "http://", "http://synthetic@", 1) + "/widgets?position=2"
				case "fragment_cursor":
					cursor = srv.URL + "/widgets?position=2#fragment"
				}
				var err error
				if mode == "saved" {
					b := newTestBundle(t, srv, StreamSpec{Pagination: spec})
					req := connectors.ReadRequest{Stream: "widgets"}
					if cursor != "" {
						c, e := newReadContinuation(b, b.Streams[0], &connsdk.NextPage{URL: cursor})
						if e != nil {
							t.Fatal(e)
						}
						req.Continuation = &c
					}
					err = ReadWithOutcome(ctx, b, req, nil, func(connectors.Record) error { return nil })
				} else {
					b := paginatedDirectReadBundle(srv.URL, spec, "/widgets")
					_, err = DirectRead(ctx, b, connectors.DirectReadRequest{Method: "GET", Path: "/widgets", PageCursor: cursor, OutputPolicy: "json_redacted"}, nil)
				}
				if err == nil || sends.Load() != 0 {
					t.Fatalf("source/cursor refusal err=%v sends=%d", err, sends.Load())
				}
			})
		}
	}
	t.Run("direct_response_bound", func(t *testing.T) {
		var sends atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sends.Add(1)
			_, _ = fmt.Fprint(w, `{"data":[{"id":"`+strings.Repeat("x", 1024)+`"}]}`)
		}))
		defer srv.Close()
		b := paginatedDirectReadBundle(srv.URL, &PaginationSpec{Type: "next_url", NextURLPath: "next"}, "/widgets")
		_, err := DirectRead(t.Context(), b, connectors.DirectReadRequest{Method: "GET", Path: "/widgets", MaxBytes: 128, OutputPolicy: "json_redacted"}, nil)
		if err == nil || sends.Load() != 1 {
			t.Fatalf("response bound err=%v sends=%d", err, sends.Load())
		}
	})
}

func TestNextLinkCapsuleAndSharedCapability229(t *testing.T) {
	var requests []string
	var origin string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.RequestURI)
		w.Header().Set("Content-Type", "application/json")
		if len(requests) == 1 {
			_, _ = fmt.Fprintf(w, `{"data":[{"id":"alpha"},{"id":"bravo"}],"next":%q}`, origin+"/widgets?page=2&tag=b&tag=a")
			return
		}
		_, _ = fmt.Fprint(w, `{"data":[{"id":"charlie"}],"next":null}`)
	}))
	defer srv.Close()
	origin = srv.URL
	b := withAllRateLimit(newTestBundle(t, srv, StreamSpec{Query: map[string]QueryParam{"page": {Template: "1"}}, Pagination: &PaginationSpec{Type: "next_url", NextURLPath: "next", PageParam: "page", RequireContinuationOnCap: true, NextURLQuery: &NextURLQuerySpec{Allowed: []string{"tag"}}}}))
	b.RateLimits.Policies[0].Coordination = connsdk.RateLimitCoordinationRequireShared
	cfg := rateLimitTestConfig(t)
	capability := &borrowedRateCoordinator218{}
	cfg.SharedRateLimits = capability
	ids := []string{}
	emit := func(r connectors.Record) error { ids = append(ids, fmt.Sprint(r["id"])); return nil }
	err := ReadWithOutcome(t.Context(), b, connectors.ReadRequest{Stream: "widgets", Config: cfg, MaxPages: 1}, nil, emit)
	var stopped *connectors.ReadBudgetStoppedError
	if !errors.As(err, &stopped) || len(requests) != 1 {
		t.Fatalf("cap err=%v requests=%q", err, requests)
	}
	page, err := readContinuationPage(b, b.Streams[0], stopped.Continuation.Clone())
	if err != nil || page.URL != origin+"/widgets?page=2&tag=b&tag=a" || len(page.Query) != 0 {
		t.Fatalf("capsule lost exact effective URL: %#v %v", page, err)
	}
	changed := b
	changed.Streams = append([]StreamSpec(nil), b.Streams...)
	copySpec := *b.Streams[0].Pagination
	copySpec.NextURLQuery = &NextURLQuerySpec{Allowed: []string{"tag", "filter"}}
	changed.Streams[0].Pagination = &copySpec
	if _, err = readContinuationPage(changed, changed.Streams[0], stopped.Continuation.Clone()); err == nil {
		t.Fatal("changed ownership definition accepted old capsule")
	}
	err = ReadWithOutcome(t.Context(), b, connectors.ReadRequest{Stream: "widgets", Config: cfg, Continuation: stopped.Continuation.Clone()}, nil, emit)
	if err != nil || !reflect.DeepEqual(ids, []string{"alpha", "bravo", "charlie"}) || !reflect.DeepEqual(requests, []string{"/widgets?page=1", "/widgets?page=2&tag=b&tag=a"}) || capability.admissions.Load() != 2 {
		t.Fatalf("resume identities=%q requests=%q admission=%d err=%v", ids, requests, capability.admissions.Load(), err)
	}
	t.Logf("capsule effective URL preserved, changed definition refused; requests=%q identities=%q actual borrowed admissions=%d", requests, ids, capability.admissions.Load())
}

func TestNextLinkOperationAdmission229(t *testing.T) {
	spec := &PaginationSpec{Type: "next_url", NextURLPath: "next", PageParam: "position", SizeParam: "size", NextURLQuery: &NextURLQuerySpec{Allowed: []string{"filter", "signature"}, Retain: []string{"filter"}}}
	parameters := []OperationParameter{{Name: "position", In: "query", Type: "string"}, {Name: "size", In: "query", Type: "integer"}, {Name: "filter", In: "query", Type: "string"}, {Name: "signature", In: "query", Type: "string"}}
	b := operationPaginatedBundle("https://fixture.invalid", spec, parameters, "/widgets")
	if err := validateRESTOperationPagination(b.Operations[0]); err != nil {
		t.Fatal(err)
	}
	b.Operations[0].REST.PaginationParameters = parameters[:3]
	if err := validateRESTOperationPagination(b.Operations[0]); err == nil {
		t.Fatal("unsigned source inventory admitted undeclared signature key")
	}
}

// Ordinary Requester overrides remain unchanged outside next-link composition.
func TestNextLinkOrdinaryRequesterCompatibility229(t *testing.T) {
	var request string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request = r.RequestURI
		_, _ = fmt.Fprint(w, `{"ok":true}`)
	}))
	defer srv.Close()
	r := connsdk.Requester{BaseURL: srv.URL}
	_, err := r.Do(t.Context(), "GET", "/widgets?page=2&tag=b&tag=a", url.Values{"page": {"1"}}, nil)
	if err != nil || request != "/widgets?page=1&tag=b&tag=a" {
		t.Fatalf("ordinary Requester override changed: %q %v", request, err)
	}
}

func TestNextLinkCancellation229(t *testing.T) {
	for _, mode := range []string{"saved", "direct"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var sends atomic.Int32
			var origin string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count := sends.Add(1)
				if count == 2 {
					cancel()
					<-r.Context().Done()
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintf(w, `{"data":[{"id":"alpha"},{"id":"bravo"}],"next":%q}`, origin+"/widgets?page=2")
			}))
			defer srv.Close()
			origin = srv.URL
			spec := &PaginationSpec{Type: "next_url", NextURLPath: "next", PageParam: "page"}
			var err error
			if mode == "saved" {
				b := newTestBundle(t, srv, StreamSpec{Pagination: spec})
				_, err = readAll(t, ctx, b, connectors.ReadRequest{Stream: "widgets"}, nil)
			} else {
				b := paginatedDirectReadBundle(origin, spec, "/widgets")
				first, e := DirectRead(ctx, b, connectors.DirectReadRequest{Method: "GET", Path: "/widgets", OutputPolicy: "json_redacted"}, nil)
				if e != nil {
					t.Fatal(e)
				}
				_, err = DirectRead(ctx, b, connectors.DirectReadRequest{Method: "GET", Path: "/widgets", PageCursor: first.Page.NextCursor, OutputPolicy: "json_redacted"}, nil)
			}
			if !errors.Is(err, context.Canceled) || sends.Load() != 2 {
				t.Fatalf("second-request cancellation err=%v sends=%d", err, sends.Load())
			}
			t.Log("cancellation reached second physical request; no third send")
		})
	}
}

func TestNextLinkEncodedURLBound229(t *testing.T) {
	var sends atomic.Int32
	var origin string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sends.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"data":[{"id":"alpha"}],"next":%q}`, origin+"/widgets?page="+strings.Repeat("x", connectors.MaxDirectReadPageCursorBytes))
	}))
	defer srv.Close()
	origin = srv.URL
	b := newTestBundle(t, srv, StreamSpec{Pagination: &PaginationSpec{Type: "next_url", NextURLPath: "next"}})
	_, err := readAll(t, t.Context(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err == nil || sends.Load() != 1 {
		t.Fatalf("next URL byte bound err=%v sends=%d", err, sends.Load())
	}
}

func TestNextLinkLoadedDescriptor229(t *testing.T) {
	for _, location := range []string{"base", "stream", "operation"} {
		for _, invalid := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/invalid_%v", location, invalid), func(t *testing.T) {
				fsys := fullValidBundleFS("acme")
				policy := map[string]any{"allowed": []string{"filter"}, "retain": []string{"filter"}}
				if invalid {
					policy["arbitrary"] = true
				}
				pagination := map[string]any{"type": "next_url", "next_url_path": "next", "page_param": "position", "next_url_query": policy}
				if location == "operation" {
					operations := map[string]any{"operations": []any{map[string]any{"id": "acme.list", "kind": "rest_read", "summary": "List", "risk": "low", "approval": "none", "output_policy": "json_redacted", "rest": map[string]any{"method": "GET", "path": "/widgets", "max_bytes": 1024, "pagination": pagination, "pagination_parameters": []any{map[string]any{"name": "position", "in": "query", "type": "string"}, map[string]any{"name": "filter", "in": "query", "type": "string"}}}}}}
					raw, err := json.Marshal(operations)
					if err != nil {
						t.Fatal(err)
					}
					fsys["acme/operations.json"] = &fstest.MapFile{Data: raw}
				} else {
					var doc map[string]any
					if err := json.Unmarshal(fsys["acme/streams.json"].Data, &doc); err != nil {
						t.Fatal(err)
					}
					if location == "base" {
						doc["base"].(map[string]any)["pagination"] = pagination
					} else {
						doc["streams"].([]any)[0].(map[string]any)["pagination"] = pagination
					}
					raw, err := json.Marshal(doc)
					if err != nil {
						t.Fatal(err)
					}
					fsys["acme/streams.json"] = &fstest.MapFile{Data: raw}
				}
				b, err := Load(fsys, "acme")
				if invalid {
					if err == nil {
						t.Fatal("unknown query-contract field admitted")
					}
					return
				}
				if err != nil {
					t.Fatalf("closed next-link descriptor rejected by actual loader: %v", err)
				}
				spec := b.HTTP.Pagination
				if location == "stream" {
					spec = b.Streams[0].Pagination
				}
				if location == "operation" {
					spec = b.Operations[0].REST.Pagination
				}
				if spec.NextURLQuery == nil || !reflect.DeepEqual(spec.NextURLQuery.Allowed, []string{"filter"}) || !reflect.DeepEqual(spec.NextURLQuery.Retain, []string{"filter"}) {
					t.Fatal("loaded descriptor lost closed ownership")
				}
			})
		}
	}
}
