package amplitude

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Amplitude API resource path (relative
// to base_url), the JSON path the record array lives at, and the record mapper
// that flattens each object into a connectors.Record.
type streamEndpoint struct {
	// resource is the API path segment, e.g. "api/3/cohorts".
	resource string
	// recordsPath is the dotted JSON path to the array of records in the body.
	recordsPath string
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// amplitudeStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in amplitudeStreams; the read
// path is fully data-driven from this table.
var amplitudeStreamEndpoints = map[string]streamEndpoint{
	"cohorts":     {resource: "api/3/cohorts", recordsPath: "cohorts", mapRecord: amplitudeCohortRecord},
	"annotations": {resource: "api/3/annotations", recordsPath: "data", mapRecord: amplitudeAnnotationRecord},
	"events_list": {resource: "api/2/events/list", recordsPath: "data", mapRecord: amplitudeEventListRecord},
}

// amplitudeStreams returns the connector's published stream catalog. These are
// all full-refresh list endpoints (no incremental cursor), so CursorFields are
// empty.
func amplitudeStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "cohorts",
			Description: "Amplitude behavioral cohorts (GET /api/3/cohorts).",
			PrimaryKey:  []string{"id"},
			Fields:      amplitudeCohortFields(),
		},
		{
			Name:        "annotations",
			Description: "Amplitude chart annotations (GET /api/3/annotations).",
			PrimaryKey:  []string{"id"},
			Fields:      amplitudeAnnotationFields(),
		},
		{
			Name:        "events_list",
			Description: "Amplitude active event types for the project (GET /api/2/events/list).",
			PrimaryKey:  []string{"value"},
			Fields:      amplitudeEventListFields(),
		},
	}
}

func amplitudeCohortFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "size", Type: "integer"},
		{Name: "archived", Type: "boolean"},
		{Name: "published", Type: "boolean"},
		{Name: "owners", Type: "array"},
		{Name: "type", Type: "string"},
		{Name: "lastComputed", Type: "integer"},
		{Name: "lastMod", Type: "integer"},
		{Name: "createdAt", Type: "integer"},
	}
}

func amplitudeAnnotationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "date", Type: "string"},
		{Name: "label", Type: "string"},
		{Name: "details", Type: "string"},
	}
}

func amplitudeEventListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "value", Type: "string"},
		{Name: "display", Type: "string"},
		{Name: "totals", Type: "integer"},
		{Name: "non_active", Type: "boolean"},
		{Name: "deleted", Type: "boolean"},
		{Name: "hidden", Type: "boolean"},
		{Name: "flow_hidden", Type: "boolean"},
	}
}

func amplitudeCohortRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"description":  item["description"],
		"size":         item["size"],
		"archived":     item["archived"],
		"published":    item["published"],
		"owners":       item["owners"],
		"type":         item["type"],
		"lastComputed": item["lastComputed"],
		"lastMod":      item["lastMod"],
		"createdAt":    item["createdAt"],
	}
}

func amplitudeAnnotationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"date":    item["date"],
		"label":   item["label"],
		"details": item["details"],
	}
}

func amplitudeEventListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"value":       item["value"],
		"display":     item["display"],
		"totals":      item["totals"],
		"non_active":  item["non_active"],
		"deleted":     item["deleted"],
		"hidden":      item["hidden"],
		"flow_hidden": item["flow_hidden"],
	}
}
