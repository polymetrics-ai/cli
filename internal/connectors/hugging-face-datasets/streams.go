package huggingfacedatasets

import "polymetrics.ai/internal/connectors"

// streamKind classifies how a stream is read from the dataset-viewer API.
type streamKind int

const (
	// kindList reads a single non-paginated JSON object and emits the array at
	// recordsPath (e.g. /splits -> "splits", /size -> "size.splits").
	kindList streamKind = iota
	// kindRows reads the offset-paginated /rows endpoint and flattens each
	// element's nested "row" object.
	kindRows
)

// streamEndpoint maps a stream name to the dataset-viewer resource path, the
// JSON path to its records array, the read strategy, and the record mapper.
type streamEndpoint struct {
	resource    string
	recordsPath string
	kind        streamKind
	mapRecord   func(item map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in datasetStreams; the read path is
// fully data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"splits": {resource: "splits", recordsPath: "splits", kind: kindList, mapRecord: splitRecord},
	"sizes":  {resource: "size", recordsPath: "size.splits", kind: kindList, mapRecord: sizeRecord},
	"rows":   {resource: "rows", recordsPath: "rows", kind: kindRows, mapRecord: rowRecord},
}

// datasetStreams returns the connector's published stream catalog.
func datasetStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "splits",
			Description: "Splits and subsets (configs) available for the configured dataset.",
			PrimaryKey:  []string{"dataset", "config", "split"},
			Fields:      splitFields(),
		},
		{
			Name:        "sizes",
			Description: "Per-split size metrics (row counts and byte sizes) for the configured dataset.",
			PrimaryKey:  []string{"dataset", "config", "split"},
			Fields:      sizeFields(),
		},
		{
			Name:        "rows",
			Description: "Rows of a single (config, split) slice of the configured dataset, paginated by offset.",
			PrimaryKey:  []string{"row_idx"},
			Fields:      rowFields(),
		},
	}
}

func splitFields() []connectors.Field {
	return []connectors.Field{
		{Name: "dataset", Type: "string"},
		{Name: "config", Type: "string"},
		{Name: "split", Type: "string"},
	}
}

func sizeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "dataset", Type: "string"},
		{Name: "config", Type: "string"},
		{Name: "split", Type: "string"},
		{Name: "num_rows", Type: "integer"},
		{Name: "num_columns", Type: "integer"},
		{Name: "num_bytes_parquet_files", Type: "integer"},
		{Name: "num_bytes_memory", Type: "integer"},
	}
}

func rowFields() []connectors.Field {
	return []connectors.Field{
		{Name: "row_idx", Type: "integer"},
		{Name: "dataset", Type: "string"},
		{Name: "config", Type: "string"},
		{Name: "split", Type: "string"},
		{Name: "row", Type: "object"},
		{Name: "truncated_cells", Type: "array"},
	}
}

func splitRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"dataset": item["dataset"],
		"config":  item["config"],
		"split":   item["split"],
	}
}

func sizeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"dataset":                 item["dataset"],
		"config":                  item["config"],
		"split":                   item["split"],
		"num_rows":                item["num_rows"],
		"num_columns":             item["num_columns"],
		"num_bytes_parquet_files": item["num_bytes_parquet_files"],
		"num_bytes_memory":        item["num_bytes_memory"],
	}
}

// rowRecord flattens a /rows element. Each element is shaped
// {"row_idx":N,"row":{...},"truncated_cells":[...]}; the nested "row" columns
// are hoisted to the top level for convenience while the original nested object,
// the index, and the truncation flags are preserved.
func rowRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"row_idx":         item["row_idx"],
		"row":             item["row"],
		"truncated_cells": item["truncated_cells"],
	}
	if inner, ok := item["row"].(map[string]any); ok {
		for k, v := range inner {
			// Do not clobber the structural keys with a same-named column.
			if _, taken := rec[k]; taken {
				continue
			}
			rec[k] = v
		}
	}
	return rec
}
