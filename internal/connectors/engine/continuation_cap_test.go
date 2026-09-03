package engine

import (
	"context"
	"errors"
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
