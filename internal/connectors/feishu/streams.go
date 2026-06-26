package feishu

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream to its Bitable resource shape. The Feishu Bitable
// REST API addresses everything under an app (the Base) identified by app_token;
// the records and fields streams further require a table_id. The read path is
// fully data-driven from this table plus feishuStreamEndpoints below.
type streamEndpoint struct {
	// resource is the path suffix appended after the app_token segment. For the
	// records and fields streams it includes the tables/{table_id}/... portion,
	// filled in at request time because table_id is config-driven.
	resourceFor func(tableID string) string
	// needsTable is true when the stream requires a configured table_id.
	needsTable bool
	// mapRecord flattens a raw Bitable item into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// feishuStreamEndpoints is the per-stream routing table. Adding a stream means
// one entry here plus a Stream definition in feishuStreams.
var feishuStreamEndpoints = map[string]streamEndpoint{
	"records": {
		resourceFor: func(tableID string) string { return "tables/" + tableID + "/records" },
		needsTable:  true,
		mapRecord:   feishuRecordRecord,
	},
	"tables": {
		resourceFor: func(string) string { return "tables" },
		needsTable:  false,
		mapRecord:   feishuTableRecord,
	},
	"fields": {
		resourceFor: func(tableID string) string { return "tables/" + tableID + "/fields" },
		needsTable:  true,
		mapRecord:   feishuFieldRecord,
	},
}

// feishuStreams returns the connector's published stream catalog. Feishu Bitable
// is a read source (a hosted spreadsheet/database); the connector is read-only.
func feishuStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "records",
			Description: "Rows of the configured Bitable table; user-defined columns are flattened from the Feishu fields object.",
			PrimaryKey:  []string{"record_id"},
			Fields:      feishuRecordFields(),
		},
		{
			Name:        "tables",
			Description: "Tables (sheets) that make up the Bitable app/Base.",
			PrimaryKey:  []string{"table_id"},
			Fields:      feishuTableFields(),
		},
		{
			Name:        "fields",
			Description: "Column/field definitions of the configured Bitable table.",
			PrimaryKey:  []string{"field_id"},
			Fields:      feishuFieldFields(),
		},
	}
}

func feishuRecordFields() []connectors.Field {
	return []connectors.Field{
		{Name: "record_id", Type: "string"},
		{Name: "fields", Type: "object"},
	}
}

func feishuTableFields() []connectors.Field {
	return []connectors.Field{
		{Name: "table_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "revision", Type: "integer"},
	}
}

func feishuFieldFields() []connectors.Field {
	return []connectors.Field{
		{Name: "field_id", Type: "string"},
		{Name: "field_name", Type: "string"},
		{Name: "type", Type: "integer"},
		{Name: "ui_type", Type: "string"},
		{Name: "is_primary", Type: "boolean"},
		{Name: "is_hidden", Type: "boolean"},
		{Name: "property", Type: "object"},
	}
}

// feishuRecordRecord flattens a Bitable record. The stable record_id is exposed
// (record_id or, as a fallback, id) and the user-defined columns in fields are
// hoisted to the top level so downstream tables expose real column names, while
// the raw fields object is retained for fidelity.
func feishuRecordRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"record_id": firstNonEmpty(item, "record_id", "id"),
		"fields":    item["fields"],
	}
	if fields, ok := item["fields"].(map[string]any); ok {
		for k, v := range fields {
			if _, clash := rec[k]; clash {
				continue
			}
			rec[k] = v
		}
	}
	return rec
}

func feishuTableRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"table_id": item["table_id"],
		"name":     item["name"],
		"revision": item["revision"],
	}
}

func feishuFieldRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"field_id":   firstNonEmpty(item, "field_id", "id"),
		"field_name": item["field_name"],
		"type":       item["type"],
		"ui_type":    item["ui_type"],
		"is_primary": item["is_primary"],
		"is_hidden":  item["is_hidden"],
		"property":   item["property"],
	}
}

// firstNonEmpty returns the first of keys whose value renders to a non-empty
// string, falling back to nil.
func firstNonEmpty(item map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := item[k]; ok && stringField(item, k) != "" {
			return v
		}
	}
	return nil
}
