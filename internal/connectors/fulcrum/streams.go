package fulcrum

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Fulcrum API resource path (relative
// to base_url) it reads from, the top-level JSON key holding the array, and the
// record mapper that flattens its objects.
//
// Fulcrum list endpoints look like GET /api/v2/<resource>.json and return
// {"<resource>":[...],"total_count":N,"current_page":1,"total_pages":M,
// "per_page":100}. Both the path segment and the records key are the plural
// resource name, so a single field captures both.
type streamEndpoint struct {
	// resource is the plural resource name (e.g. "forms"). The request path is
	// "<resource>.json" and the records array lives at body["<resource>"].
	resource string
	// mapRecord flattens a raw Fulcrum object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// fulcrumStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in fulcrumStreams; the read
// path is fully data-driven from this table.
var fulcrumStreamEndpoints = map[string]streamEndpoint{
	"forms":               {resource: "forms", mapRecord: fulcrumFormRecord},
	"records":             {resource: "records", mapRecord: fulcrumRecordRecord},
	"projects":            {resource: "projects", mapRecord: fulcrumProjectRecord},
	"choice_lists":        {resource: "choice_lists", mapRecord: fulcrumChoiceListRecord},
	"classification_sets": {resource: "classification_sets", mapRecord: fulcrumClassificationSetRecord},
}

// fulcrumStreams returns the connector's published stream catalog. Every Fulcrum
// object exposes a string id and an updated_at RFC3339 timestamp, so the primary
// key is ["id"] and the incremental cursor field is ["updated_at"] across the
// board.
func fulcrumStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "forms",
			Description:  "Fulcrum forms (app/data schema definitions).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       fulcrumFormFields(),
		},
		{
			Name:         "records",
			Description:  "Fulcrum records (data collected against forms).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       fulcrumRecordFields(),
		},
		{
			Name:         "projects",
			Description:  "Fulcrum projects used to organize records.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       fulcrumProjectFields(),
		},
		{
			Name:         "choice_lists",
			Description:  "Fulcrum reusable choice lists.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       fulcrumChoiceListFields(),
		},
		{
			Name:         "classification_sets",
			Description:  "Fulcrum classification sets (hierarchical choices).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       fulcrumClassificationSetFields(),
		},
	}
}

func fulcrumFormFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "record_count", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "auto_assign", Type: "boolean"},
	}
}

func fulcrumRecordFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "form_id", Type: "string"},
		{Name: "project_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "latitude", Type: "number"},
		{Name: "longitude", Type: "number"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "created_by", Type: "string"},
		{Name: "updated_by", Type: "string"},
	}
}

func fulcrumProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func fulcrumChoiceListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func fulcrumClassificationSetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func fulcrumFormRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"description":  item["description"],
		"status":       item["status"],
		"record_count": item["record_count"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
		"auto_assign":  item["auto_assign"],
	}
}

func fulcrumRecordRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"form_id":    item["form_id"],
		"project_id": item["project_id"],
		"status":     item["status"],
		"latitude":   item["latitude"],
		"longitude":  item["longitude"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
		"created_by": item["created_by"],
		"updated_by": item["updated_by"],
	}
}

func fulcrumProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func fulcrumChoiceListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func fulcrumClassificationSetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}
