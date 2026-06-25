package googleanalyticsdataapi

import (
	"bytes"
	"encoding/json"
	"fmt"

	"polymetrics/internal/connectors"
)

// reportResponse is the subset of the GA4 runReport response the connector
// consumes. Numbers are decoded with json.Number to preserve integer fidelity
// for rowCount; metric/dimension values arrive as strings from the API.
type reportResponse struct {
	DimensionHeaders []header    `json:"dimensionHeaders"`
	MetricHeaders    []header    `json:"metricHeaders"`
	Rows             []reportRow `json:"rows"`
	RowCount         int         `json:"-"`
	RawRowCount      json.Number `json:"rowCount"`
}

type header struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

type reportRow struct {
	DimensionValues []reportValue `json:"dimensionValues"`
	MetricValues    []reportValue `json:"metricValues"`
}

type reportValue struct {
	Value string `json:"value"`
}

// decodeReport parses a runReport response body, resolving rowCount from the raw
// json.Number into an int.
func decodeReport(body []byte) (*reportResponse, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var out reportResponse
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("decode runReport response: %w", err)
	}
	if out.RawRowCount != "" {
		if n, err := out.RawRowCount.Int64(); err == nil {
			out.RowCount = int(n)
		}
	}
	return &out, nil
}

// mapRow flattens a runReport row into a connectors.Record by projecting the
// report's dimension/metric headers onto the row's positional values. The
// property_id is included so multi-property reads stay distinguishable and the
// composite primary key holds.
func mapRow(property string, report *reportResponse, row reportRow) connectors.Record {
	record := connectors.Record{"property_id": property}
	for i, h := range report.DimensionHeaders {
		if i < len(row.DimensionValues) {
			record[h.Name] = row.DimensionValues[i].Value
		}
	}
	for i, h := range report.MetricHeaders {
		if i < len(row.MetricValues) {
			record[h.Name] = row.MetricValues[i].Value
		}
	}
	return record
}
