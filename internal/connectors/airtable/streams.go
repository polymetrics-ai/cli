package airtable

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to how its Airtable list endpoint is built
// and the record mapper that flattens its objects. Airtable has three relevant
// list shapes:
//   - bases:   GET /meta/bases                  -> {"bases":[...],"offset":...}
//   - tables:  GET /meta/bases/{baseId}/tables  -> {"tables":[...]}
//   - records: GET /{baseId}/{tableId}          -> {"records":[...],"offset":...}
//
// recordsPath is the JSON key holding the array. needsBase / needsTable flag
// which config inputs the resource path requires.
type streamEndpoint struct {
	recordsPath string
	needsBase   bool
	needsTable  bool
	mapRecord   func(map[string]any) connectors.Record
}

// airtableStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in airtableStreams.
var airtableStreamEndpoints = map[string]streamEndpoint{
	"bases":   {recordsPath: "bases", mapRecord: airtableBaseRecord},
	"tables":  {recordsPath: "tables", needsBase: true, mapRecord: airtableTableRecord},
	"records": {recordsPath: "records", needsBase: true, needsTable: true, mapRecord: airtableRecordRecord},
}

// airtableStreams returns the connector's published stream catalog.
func airtableStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "bases",
			Description: "Airtable bases accessible to the token (Metadata API /meta/bases).",
			PrimaryKey:  []string{"id"},
			Fields:      airtableBaseFields(),
		},
		{
			Name:        "tables",
			Description: "Tables in the configured base (Metadata API /meta/bases/{baseId}/tables). Requires config base_id.",
			PrimaryKey:  []string{"id"},
			Fields:      airtableTableFields(),
		},
		{
			Name:        "records",
			Description: "Records in the configured table (/{baseId}/{tableId}). Requires config base_id and table_id.",
			PrimaryKey:  []string{"id"},
			Fields:      airtableRecordFields(),
		},
	}
}

func airtableBaseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "permissionLevel", Type: "string"},
	}
}

func airtableTableFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "primaryFieldId", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "fields", Type: "object"},
		{Name: "views", Type: "object"},
	}
}

func airtableRecordFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "createdTime", Type: "string"},
		{Name: "fields", Type: "object"},
	}
}

func airtableBaseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"permissionLevel": item["permissionLevel"],
	}
}

func airtableTableRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"primaryFieldId": item["primaryFieldId"],
		"description":    item["description"],
		"fields":         item["fields"],
		"views":          item["views"],
	}
}

func airtableRecordRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"createdTime": item["createdTime"],
		"fields":      item["fields"],
	}
}
