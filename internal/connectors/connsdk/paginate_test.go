package connsdk

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func collect(t *testing.T, r *Requester, method, path string, base url.Values, p Paginator, recordsPath string) []Record {
	t.Helper()
	var got []Record
	err := Harvest(context.Background(), r, method, path, base, p, recordsPath, 0, func(rec Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Harvest: %v", err)
	}
	return got
}

func TestHarvestPageNumberPaginator(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}
		switch page {
		case 1:
			_, _ = w.Write([]byte(`[{"id":1},{"id":2}]`)) // full page
		case 2:
			_, _ = w.Write([]byte(`[{"id":3}]`)) // short page -> stop
		default:
			t.Errorf("unexpected page %d", page)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	p := &PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize: 2}
	got := collect(t, r, http.MethodGet, "/items", nil, p, "")
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
}

func TestHarvestOffsetPaginator(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		switch offset {
		case 0:
			_, _ = w.Write([]byte(`{"data":[{"id":1},{"id":2}]}`))
		case 2:
			_, _ = w.Write([]byte(`{"data":[{"id":3},{"id":4}]}`))
		case 4:
			_, _ = w.Write([]byte(`{"data":[{"id":5}]}`))
		default:
			t.Errorf("unexpected offset %d", offset)
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	p := &OffsetPaginator{LimitParam: "limit", OffsetParam: "offset", PageSize: 2}
	got := collect(t, r, http.MethodGet, "/items", nil, p, "data")
	if len(got) != 5 {
		t.Fatalf("records = %d, want 5", len(got))
	}
}

func TestHarvestCursorPaginator(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"items":[{"id":1}],"meta":{"next":"c2"}}`))
		case "c2":
			_, _ = w.Write([]byte(`{"items":[{"id":2}],"meta":{"next":"c3"}}`))
		case "c3":
			_, _ = w.Write([]byte(`{"items":[{"id":3}],"meta":{"next":""}}`))
		default:
			t.Errorf("unexpected cursor %q", r.URL.Query().Get("cursor"))
		}
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	p := &CursorPaginator{CursorParam: "cursor", TokenPath: "meta.next"}
	got := collect(t, r, http.MethodGet, "/items", nil, p, "items")
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
}

func TestHarvestLinkHeaderPaginator(t *testing.T) {
	var base string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "", "1":
			w.Header().Set("Link", fmt.Sprintf(`<%s/items?page=2>; rel="next"`, base))
			_, _ = w.Write([]byte(`[{"id":1}]`))
		case "2":
			// no Link header -> last page
			_, _ = w.Write([]byte(`[{"id":2}]`))
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()
	base = srv.URL

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	p := &LinkHeaderPaginator{}
	got := collect(t, r, http.MethodGet, "/items?page=1", nil, p, "")
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
}

func TestParseLinkNext(t *testing.T) {
	header := `<https://api.example.com/x?page=3>; rel="prev", <https://api.example.com/x?page=5>; rel="next"`
	if got := parseLinkNext(header); got != "https://api.example.com/x?page=5" {
		t.Fatalf("parseLinkNext = %q", got)
	}
	if got := parseLinkNext(""); got != "" {
		t.Fatalf("parseLinkNext(empty) = %q", got)
	}
}
