package chameleon

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Chameleon API resource path (relative
// to base_url), the JSON field that holds its record array, and the mapper that
// flattens raw objects into connectors.Record values.
type streamEndpoint struct {
	// resource is the path segment under the v3 base (e.g. "edit/surveys").
	resource string
	// fieldPath is the top-level JSON key holding the record array (e.g.
	// "surveys"). Chameleon returns {"<fieldPath>":[...]} for list endpoints.
	fieldPath string
	// cursorField is the incremental cursor field for the stream. Most streams
	// use updated_at; changes uses created_at.
	cursorField string
	// mapRecord flattens a raw Chameleon object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// chameleonStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in chameleonStreams; the read
// path is fully data-driven from this table.
var chameleonStreamEndpoints = map[string]streamEndpoint{
	"surveys":   {resource: "edit/surveys", fieldPath: "surveys", cursorField: "updated_at", mapRecord: chameleonExperienceRecord},
	"tours":     {resource: "edit/tours", fieldPath: "tours", cursorField: "updated_at", mapRecord: chameleonExperienceRecord},
	"launchers": {resource: "edit/launchers", fieldPath: "launchers", cursorField: "updated_at", mapRecord: chameleonExperienceRecord},
	"tooltips":  {resource: "edit/tooltips", fieldPath: "tooltips", cursorField: "updated_at", mapRecord: chameleonExperienceRecord},
	"segments":  {resource: "edit/segments", fieldPath: "segments", cursorField: "updated_at", mapRecord: chameleonSegmentRecord},
}

// chameleonStreams returns the connector's published stream catalog. Every
// Chameleon object exposes a string id and an updated_at timestamp, so the
// primary key is ["id"] and the incremental cursor is ["updated_at"] for the
// core experience and segment streams.
func chameleonStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "surveys",
			Description:  "Chameleon Microsurveys (in-product surveys).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chameleonExperienceFields(),
		},
		{
			Name:         "tours",
			Description:  "Chameleon Tours (multi-step product walkthroughs).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chameleonExperienceFields(),
		},
		{
			Name:         "launchers",
			Description:  "Chameleon Launchers (in-app widgets surfacing experiences).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chameleonExperienceFields(),
		},
		{
			Name:         "tooltips",
			Description:  "Chameleon Tooltips (contextual in-product hints).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chameleonExperienceFields(),
		},
		{
			Name:         "segments",
			Description:  "Chameleon Segments (audience definitions).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chameleonSegmentFields(),
		},
	}
}

// chameleonExperienceFields describes the shared shape of Chameleon experiences
// (surveys, tours, launchers, tooltips), which carry the same envelope fields.
func chameleonExperienceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "is_live", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func chameleonSegmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

// chameleonExperienceRecord flattens a surveys/tours/launchers/tooltips object.
func chameleonExperienceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"title":      item["title"],
		"type":       item["type"],
		"state":      item["state"],
		"is_live":    item["is_live"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func chameleonSegmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}
