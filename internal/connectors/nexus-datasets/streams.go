package nexusdatasets

import "polymetrics.ai/internal/connectors"

// nexusStreams returns the connector's published stream catalog. The Infor Nexus
// Data API (v3.1) exposes a single configurable export dataset whose records are
// returned in a raw_data envelope; the stream name "datasets" mirrors the
// upstream upstream connector. The dataset payload is opaque, so the cursor is the
// record updated_at timestamp and the primary key is the record id.
func nexusStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "datasets",
			Description:  "Records exported from the configured Infor Nexus dataset.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       nexusDatasetFields(),
		},
	}
}

func nexusDatasetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "dataset_name", Type: "string"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "raw_data", Type: "object"},
		{Name: "raw_data_string", Type: "string"},
	}
}

// nexusDatasetRecord flattens a raw Infor Nexus dataset record into a
// connectors.Record. The record payload is preserved verbatim under raw_data
// (and stringified under raw_data_string) so downstream consumers keep the full,
// schema-less dataset row, mirroring the upstream upstream connector contract.
func nexusDatasetRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":         stringField(item, "id"),
		"updated_at": firstNonEmpty(item, "updated_at", "modified_at", "last_modified", "timestamp"),
	}
	if raw, ok := item["raw_data"]; ok {
		rec["raw_data"] = raw
	} else {
		// The dataset record may itself be the payload (no envelope); preserve it.
		rec["raw_data"] = item
	}
	if s, ok := item["raw_data_string"]; ok {
		rec["raw_data_string"] = s
	}
	if id := rec["id"]; id == "" || id == nil {
		// Fall back to any record key the export commonly uses for identity.
		rec["id"] = firstNonEmpty(item, "record_id", "key", "uid")
	}
	return rec
}
