package engine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadCapWithKnownContinuationNeverReportsSuccess(t *testing.T) {
	requests := 0
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests != 1 {
			t.Fatalf("requests = %d, want capped read to stop before a second provider request", requests)
		}
		_, _ = w.Write([]byte(`{"items":[{"id":"one"}],"next":"resume-token"}`))
	})
	bundle := newTestBundle(t, srv, StreamSpec{
		Name:    "events",
		Path:    "/events",
		Records: RecordsSpec{Path: "items"},
		Pagination: &PaginationSpec{
			Type:                     "cursor",
			CursorParam:              "cursor",
			TokenPath:                "next",
			PageSize:                 1,
			MaxPages:                 1,
			RequireContinuationOnCap: true,
		},
	})
	emitted := 0
	err := Read(context.Background(), bundle, connectors.ReadRequest{Stream: "events"}, nil, func(connectors.Record) error {
		emitted++
		return nil
	})
	var stopped *connectors.ReadBudgetStoppedError
	if !errors.As(err, &stopped) {
		t.Fatalf("capped continuation error = %T %v, want ReadBudgetStoppedError", err, err)
	}
	if len(stopped.Continuation.Token) == 0 || emitted != 1 || requests != 1 {
		t.Fatalf("capped continuation = %#v emitted=%d requests=%d", stopped.Continuation, emitted, requests)
	}
}

func TestReadContinuationResumesExactProviderCursor(t *testing.T) {
	var cursors []string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("cursor")
		cursors = append(cursors, cursor)
		switch cursor {
		case "":
			_, _ = w.Write([]byte(`{"items":[{"id":"one"}],"next":"page-two"}`))
		case "page-two":
			_, _ = w.Write([]byte(`{"items":[{"id":"two"}],"next":"page-three"}`))
		default:
			t.Fatalf("unexpected resumed cursor %q", cursor)
		}
	})
	bundle := newTestBundle(t, srv, StreamSpec{
		Name:    "events",
		Path:    "/events",
		Records: RecordsSpec{Path: "items"},
		Pagination: &PaginationSpec{
			Type:                     "cursor",
			CursorParam:              "cursor",
			TokenPath:                "next",
			PageSize:                 1,
			RequireContinuationOnCap: true,
		},
	})

	var first []connectors.Record
	err := ReadWithOutcome(context.Background(), bundle, connectors.ReadRequest{Stream: "events", MaxPages: 1}, nil, func(record connectors.Record) error {
		first = append(first, record)
		return nil
	})
	var firstStop *connectors.ReadBudgetStoppedError
	if !errors.As(err, &firstStop) {
		t.Fatalf("first capped read error = %T %v, want ReadBudgetStoppedError", err, err)
	}

	var second []connectors.Record
	err = ReadWithOutcome(context.Background(), bundle, connectors.ReadRequest{Stream: "events", MaxPages: 1, Continuation: firstStop.Continuation.Clone()}, nil, func(record connectors.Record) error {
		second = append(second, record)
		return nil
	})
	var secondStop *connectors.ReadBudgetStoppedError
	if !errors.As(err, &secondStop) {
		t.Fatalf("resumed capped read error = %T %v, want ReadBudgetStoppedError", err, err)
	}
	if len(first) != 1 || first[0]["id"] != "one" || len(second) != 1 || second[0]["id"] != "two" {
		t.Fatalf("resumed records = %#v then %#v, want one then two", first, second)
	}
	if got, want := fmt.Sprint(cursors), "[ page-two]"; got != want {
		t.Fatalf("provider cursors = %s, want %s without replaying an acknowledged page", got, want)
	}
}

func TestReadContinuationResumesExactProviderOffset(t *testing.T) {
	var offsets []string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("offset")
		offsets = append(offsets, offset)
		switch offset {
		case "0":
			_, _ = w.Write([]byte(`{"items":[{"id":"one"}]}`))
		case "1":
			_, _ = w.Write([]byte(`{"items":[{"id":"two"}]}`))
		default:
			t.Fatalf("unexpected resumed offset %q", offset)
		}
	})
	bundle := newTestBundle(t, srv, StreamSpec{
		Name:    "events",
		Path:    "/events",
		Records: RecordsSpec{Path: "items"},
		Pagination: &PaginationSpec{
			Type:                     "offset_limit",
			LimitParam:               "limit",
			OffsetParam:              "offset",
			PageSize:                 1,
			RequireContinuationOnCap: true,
		},
	})

	firstStop := readOneCappedPage(t, bundle, nil)
	secondStop := readOneCappedPage(t, bundle, firstStop.Continuation.Clone())
	if len(secondStop.Continuation.Token) == 0 {
		t.Fatal("resumed offset read did not produce the next continuation")
	}
	if got, want := fmt.Sprint(offsets), "[0 1]"; got != want {
		t.Fatalf("provider offsets = %s, want %s without replaying an acknowledged page", got, want)
	}
}

func TestReadContinuationResumesExactProviderURL(t *testing.T) {
	var requestQueries []string
	var serverURL string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestQueries = append(requestQueries, r.URL.RawQuery)
		switch r.URL.Query().Get("token") {
		case "":
			_, _ = w.Write([]byte(fmt.Sprintf(`{"items":[{"id":"one"}],"next":%q}`, serverURL+"/events?token=two")))
		case "two":
			_, _ = w.Write([]byte(fmt.Sprintf(`{"items":[{"id":"two"}],"next":%q}`, serverURL+"/events?token=three")))
		default:
			t.Fatalf("unexpected resumed URL query %q", r.URL.RawQuery)
		}
	})
	serverURL = srv.URL
	bundle := newTestBundle(t, srv, StreamSpec{
		Name:    "events",
		Path:    "/events",
		Records: RecordsSpec{Path: "items"},
		Pagination: &PaginationSpec{
			Type:                     "next_url",
			NextURLPath:              "next",
			RequireContinuationOnCap: true,
		},
	})

	firstStop := readOneCappedPage(t, bundle, nil)
	secondStop := readOneCappedPage(t, bundle, firstStop.Continuation.Clone())
	if len(secondStop.Continuation.Token) == 0 {
		t.Fatal("resumed URL read did not produce the next continuation")
	}
	if got, want := fmt.Sprint(requestQueries), "[ token=two]"; got != want {
		t.Fatalf("provider URL queries = %s, want %s without replaying an acknowledged page", got, want)
	}
}

func readOneCappedPage(t *testing.T, bundle Bundle, continuation *connectors.ReadContinuation) *connectors.ReadBudgetStoppedError {
	t.Helper()
	err := ReadWithOutcome(context.Background(), bundle, connectors.ReadRequest{Stream: "events", MaxPages: 1, Continuation: continuation}, nil, func(connectors.Record) error {
		return nil
	})
	var stopped *connectors.ReadBudgetStoppedError
	if !errors.As(err, &stopped) {
		t.Fatalf("capped read error = %T %v, want ReadBudgetStoppedError", err, err)
	}
	return stopped
}
