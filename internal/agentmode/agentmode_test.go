package agentmode

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func testRecords(n int) []connectors.Record {
	out := make([]connectors.Record, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, connectors.Record{
			"id":         "cus_" + strconv.Itoa(i),
			"name":       "Customer Number " + strconv.Itoa(i),
			"email":      "customer" + strconv.Itoa(i) + "@example.com",
			"plan":       "enterprise",
			"updated_at": "2026-06-20T10:00:00Z",
			"notes":      "a verbose note field that bloats full-row agent dumps",
		})
	}
	return out
}

func TestFieldsProjection(t *testing.T) {
	t.Parallel()

	records := testRecords(3)
	got := FieldsProjection(records, []string{"id", "email", "missing"})

	if len(got) != 3 {
		t.Fatalf("FieldsProjection() rows = %d, want 3", len(got))
	}
	for i, row := range got {
		if len(row) != 2 {
			t.Fatalf("row %d key count = %d, want 2: %v", i, len(row), row)
		}
		for _, want := range []string{"id", "email"} {
			if _, ok := row[want]; !ok {
				t.Fatalf("row %d missing projected field %q: %v", i, want, row)
			}
		}
		for _, dropped := range []string{"missing", "name", "notes"} {
			if _, ok := row[dropped]; ok {
				t.Fatalf("row %d retained field %q: %v", i, dropped, row)
			}
		}
	}
	if len(records[0]) == 2 {
		t.Fatalf("FieldsProjection mutated source record: %v", records[0])
	}
}

func TestSummarize(t *testing.T) {
	t.Parallel()

	payload, err := Summarize("QueryResult", testRecords(50), 3)
	if err != nil {
		t.Fatalf("Summarize() error = %v", err)
	}
	if bytes.Contains(payload, []byte("\n")) {
		t.Fatalf("summary should be compact JSON without embedded newlines: %s", payload)
	}

	var got struct {
		Kind   string              `json:"kind"`
		Count  int                 `json:"count"`
		Fields []string            `json:"fields"`
		Sample []connectors.Record `json:"sample"`
	}
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("summary is not valid JSON: %v", err)
	}
	if got.Kind != "QueryResult" {
		t.Fatalf("kind = %q, want QueryResult", got.Kind)
	}
	if got.Count != 50 {
		t.Fatalf("count = %d, want 50", got.Count)
	}
	if len(got.Sample) != 3 {
		t.Fatalf("sample rows = %d, want 3", len(got.Sample))
	}
	wantFields := []string{"email", "id", "name", "notes", "plan", "updated_at"}
	if strings.Join(got.Fields, ",") != strings.Join(wantFields, ",") {
		t.Fatalf("fields = %v, want sorted %v", got.Fields, wantFields)
	}
}

func TestSummarizeClampsSample(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		records []connectors.Record
		sampleN int
		want    int
	}{
		{name: "larger than count", records: testRecords(2), sampleN: 10, want: 2},
		{name: "negative", records: testRecords(2), sampleN: -1, want: 0},
		{name: "empty", records: nil, sampleN: 3, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := Summarize("QueryResult", tt.records, tt.sampleN)
			if err != nil {
				t.Fatalf("Summarize() error = %v", err)
			}
			var got struct {
				Count  int                 `json:"count"`
				Fields []string            `json:"fields"`
				Sample []connectors.Record `json:"sample"`
			}
			if err := json.Unmarshal(payload, &got); err != nil {
				t.Fatalf("summary is not valid JSON: %v", err)
			}
			if got.Count != len(tt.records) {
				t.Fatalf("count = %d, want %d", got.Count, len(tt.records))
			}
			if len(got.Sample) != tt.want {
				t.Fatalf("sample rows = %d, want %d", len(got.Sample), tt.want)
			}
		})
	}
}

func TestEncodeStreamNDJSON(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := EncodeStream(&buf, testRecords(4)); err != nil {
		t.Fatalf("EncodeStream() error = %v", err)
	}
	if got := buf.String(); got == "" || got[len(got)-1] != '\n' {
		t.Fatalf("NDJSON stream must end with newline, got %q", got)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("NDJSON line count = %d, want 4", len(lines))
	}
	for i, line := range lines {
		var row connectors.Record
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("line %d is not valid JSON: %v", i, err)
		}
	}
}

func TestTokenReductionAgainstVerboseEnvelope(t *testing.T) {
	t.Parallel()

	records := testRecords(50)
	verbose, err := VerboseEnvelope("QueryResult", records)
	if err != nil {
		t.Fatalf("VerboseEnvelope() error = %v", err)
	}
	summary, err := Summarize("QueryResult", records, 3)
	if err != nil {
		t.Fatalf("Summarize() error = %v", err)
	}

	if EstimateTokens(summary)*2 >= EstimateTokens(verbose) {
		t.Fatalf("summary tokens = %d, verbose tokens = %d; want summary < 50%%",
			EstimateTokens(summary), EstimateTokens(verbose))
	}

	var projected bytes.Buffer
	err = EncodeStream(&projected, FieldsProjection(records, []string{"id", "email"}))
	if err != nil {
		t.Fatalf("EncodeStream(projected) error = %v", err)
	}
	if EstimateTokens(projected.Bytes())*2 >= EstimateTokens(verbose) {
		t.Fatalf("projected tokens = %d, verbose tokens = %d; want projected < 50%%",
			EstimateTokens(projected.Bytes()), EstimateTokens(verbose))
	}
}
