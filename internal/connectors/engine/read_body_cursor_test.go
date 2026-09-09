package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadCursorPaginationMovesTokenIntoDeclaredBody(t *testing.T) {
	var calls int
	server := jsonServer(t, func(w http.ResponseWriter, request *http.Request) {
		calls++
		if request.URL.Query().Get("NextToken") != "" {
			t.Fatal("cursor token leaked into query instead of declared body")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		switch calls {
		case 1:
			if _, found := body["NextToken"]; found {
				t.Fatal("first request unexpectedly carried a cursor")
			}
			_, _ = w.Write([]byte(`{"Events":[{"id":"one"}],"NextToken":"cursor-1"}`))
		case 2:
			if body["NextToken"] != "cursor-1" {
				t.Fatalf("second request cursor = %#v, want cursor-1", body["NextToken"])
			}
			_, _ = w.Write([]byte(`{"Events":[{"id":"two"}]}`))
		default:
			t.Fatal("unexpected extra page")
		}
	})
	bundle := newTestBundle(t, server, StreamSpec{
		Method:  http.MethodPost,
		Path:    "/",
		Body:    map[string]any{"MaxResults": 1},
		Records: RecordsSpec{Path: "Events"},
		Pagination: &PaginationSpec{
			Type:            "cursor",
			CursorParam:     "NextToken",
			TokenPath:       "NextToken",
			BodyCursorField: "NextToken",
		},
	})
	records, err := readAll(t, context.Background(), bundle, connectors.ReadRequest{}, nil)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(records) != 2 || records[0]["id"] != "one" || records[1]["id"] != "two" {
		t.Fatalf("records = %#v, want both cursor pages", records)
	}
}
