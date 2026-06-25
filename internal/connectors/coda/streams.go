package coda

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Coda API resource it reads from.
//
// Coda has two shapes of list endpoint: workspace-level lists (e.g. /docs) and
// doc-scoped lists (e.g. /docs/{docId}/tables). docScoped marks the latter; for
// those the resolved path is /docs/<doc_id>/<resource> where doc_id comes from
// config. mapRecord flattens a raw Coda object into a connectors.Record.
type streamEndpoint struct {
	// resource is the path segment. For workspace lists it is the full relative
	// path (e.g. "docs"); for doc-scoped lists it is the suffix after the doc id
	// (e.g. "tables", giving "docs/<doc_id>/tables").
	resource  string
	docScoped bool
	mapRecord func(map[string]any) connectors.Record
}

// codaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in codaStreams; the read path
// is fully data-driven from this table.
var codaStreamEndpoints = map[string]streamEndpoint{
	"docs":     {resource: "docs", mapRecord: codaDocRecord},
	"tables":   {resource: "tables", docScoped: true, mapRecord: codaTableRecord},
	"pages":    {resource: "pages", docScoped: true, mapRecord: codaPageRecord},
	"formulas": {resource: "formulas", docScoped: true, mapRecord: codaFormulaRecord},
	"controls": {resource: "controls", docScoped: true, mapRecord: codaControlRecord},
}

// codaStreams returns the connector's published stream catalog. Coda objects
// are keyed by a string id and have no incremental cursor on these list
// endpoints (full refresh only), so PrimaryKey is ["id"] and CursorFields is
// empty across the board.
func codaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "docs",
			Description: "Coda docs the API token can access.",
			PrimaryKey:  []string{"id"},
			Fields:      codaDocFields(),
		},
		{
			Name:        "tables",
			Description: "Tables and views within a Coda doc (requires doc_id config).",
			PrimaryKey:  []string{"id"},
			Fields:      codaTableFields(),
		},
		{
			Name:        "pages",
			Description: "Pages within a Coda doc (requires doc_id config).",
			PrimaryKey:  []string{"id"},
			Fields:      codaPageFields(),
		},
		{
			Name:        "formulas",
			Description: "Named formulas within a Coda doc (requires doc_id config).",
			PrimaryKey:  []string{"id"},
			Fields:      codaFormulaFields(),
		},
		{
			Name:        "controls",
			Description: "Controls (buttons, sliders, etc.) within a Coda doc (requires doc_id config).",
			PrimaryKey:  []string{"id"},
			Fields:      codaControlFields(),
		},
	}
}

func codaDocFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "href", Type: "string"},
		{Name: "browserLink", Type: "string"},
		{Name: "owner", Type: "string"},
		{Name: "ownerName", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
		{Name: "workspaceId", Type: "string"},
		{Name: "folderId", Type: "string"},
	}
}

func codaTableFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "tableType", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "href", Type: "string"},
		{Name: "browserLink", Type: "string"},
		{Name: "rowCount", Type: "integer"},
		{Name: "doc_id", Type: "string"},
	}
}

func codaPageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "subtitle", Type: "string"},
		{Name: "href", Type: "string"},
		{Name: "browserLink", Type: "string"},
		{Name: "contentType", Type: "string"},
		{Name: "doc_id", Type: "string"},
	}
}

func codaFormulaFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "href", Type: "string"},
		{Name: "doc_id", Type: "string"},
	}
}

func codaControlFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "controlType", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "href", Type: "string"},
		{Name: "doc_id", Type: "string"},
	}
}

func codaDocRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"type":        item["type"],
		"name":        item["name"],
		"href":        item["href"],
		"browserLink": item["browserLink"],
		"owner":       item["owner"],
		"ownerName":   item["ownerName"],
		"createdAt":   item["createdAt"],
		"updatedAt":   item["updatedAt"],
		"workspaceId": item["workspaceId"],
		"folderId":    item["folderId"],
	}
}

func codaTableRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"type":        item["type"],
		"tableType":   item["tableType"],
		"name":        item["name"],
		"href":        item["href"],
		"browserLink": item["browserLink"],
		"rowCount":    item["rowCount"],
	}
}

func codaPageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"type":        item["type"],
		"name":        item["name"],
		"subtitle":    item["subtitle"],
		"href":        item["href"],
		"browserLink": item["browserLink"],
		"contentType": item["contentType"],
	}
}

func codaFormulaRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"type": item["type"],
		"name": item["name"],
		"href": item["href"],
	}
}

func codaControlRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"type":        item["type"],
		"controlType": item["controlType"],
		"name":        item["name"],
		"href":        item["href"],
	}
}
