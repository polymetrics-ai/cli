package engine

import (
	"testing"

	"polymetrics.ai/internal/connectors/connsdk"
)

func TestArrayZipProjectionCombinesDeclaredArrays(t *testing.T) {
	records, err := applyArrayZipProjection([]connsdk.Record{{
		"meta":      map[string]any{"symbol": "AAPL", "currency": "USD"},
		"timestamp": []any{float64(1), float64(2)},
		"indicators": map[string]any{
			"quote": []any{map[string]any{"open": []any{float64(10), float64(11)}, "close": []any{float64(12), float64(13)}}},
		},
	}}, &ArrayZipProjectionSpec{
		StaticFields: []ArrayZipFieldSpec{
			{Field: "symbol", Path: "meta.symbol"},
			{Field: "currency", Path: "meta.currency"},
		},
		ArrayFields: []ArrayZipFieldSpec{
			{Field: "timestamp", Path: "timestamp"},
			{Field: "open", Path: "indicators.quote.0.open"},
			{Field: "close", Path: "indicators.quote.0.close"},
		},
	})
	if err != nil {
		t.Fatalf("applyArrayZipProjection: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("records = %#v, want two zipped rows", records)
	}
	if records[0]["symbol"] != "AAPL" || records[0]["currency"] != "USD" || records[0]["timestamp"] != float64(1) || records[0]["open"] != float64(10) || records[0]["close"] != float64(12) {
		t.Fatalf("first zipped record = %#v", records[0])
	}
	if records[1]["timestamp"] != float64(2) || records[1]["open"] != float64(11) || records[1]["close"] != float64(13) {
		t.Fatalf("second zipped record = %#v", records[1])
	}
}

func TestArrayZipProjectionRejectsMismatchedArrays(t *testing.T) {
	_, err := applyArrayZipProjection([]connsdk.Record{{
		"timestamps": []any{float64(1), float64(2)},
		"open":       []any{float64(10)},
	}}, &ArrayZipProjectionSpec{
		ArrayFields: []ArrayZipFieldSpec{
			{Field: "timestamp", Path: "timestamps"},
			{Field: "open", Path: "open"},
		},
	})
	if err == nil {
		t.Fatal("mismatched declared arrays were accepted")
	}
}
