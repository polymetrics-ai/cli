package engine

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadDateWindowFanOutRejectsCalendarRangeBeforeIO(t *testing.T) {
	srv := jsonServer(t, func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("calendar-overflow range reached provider I/O")
	})
	bundle := newTestBundle(t, srv, StreamSpec{
		Name:    "windows",
		Path:    "/windows",
		Records: RecordsSpec{Path: "items"},
		DateWindowFanOut: &DateWindowFanOutSpec{
			StartDateConfigKey: "start",
			EndDateConfigKey:   "end",
			BatchSizeConfigKey: "days",
			DateFromQueryParam: "from",
			DateToQueryParam:   "to",
			MaxBatchDays:       365,
			MaxWindows:         600,
		},
	})
	_, err := readAll(t, context.Background(), bundle, connectors.ReadRequest{Stream: "windows", Config: connectors.RuntimeConfig{Config: map[string]string{"start": "0001-01-01", "end": "9999-12-31", "days": "365"}}}, nil)
	if err == nil || !strings.Contains(err.Error(), "windows exceed declared maximum") {
		t.Fatalf("calendar-safe date budget error = %v, want pre-I/O window refusal", err)
	}
}
