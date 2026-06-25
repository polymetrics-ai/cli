package appcues

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Appcues resource segment (appended to
// accounts/{account_id}/) it reads from, and the record mapper that flattens its
// objects into a connectors.Record.
type streamEndpoint struct {
	// resource is the trailing path segment, e.g. "flows".
	resource string
	// mapRecord flattens a raw Appcues object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// appcuesStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in appcuesStreams; the read path
// is fully data-driven from this table.
var appcuesStreamEndpoints = map[string]streamEndpoint{
	"flows":      {resource: "flows", mapRecord: appcuesFlowRecord},
	"segments":   {resource: "segments", mapRecord: appcuesSegmentRecord},
	"tags":       {resource: "tags", mapRecord: appcuesTagRecord},
	"checklists": {resource: "checklists", mapRecord: appcuesChecklistRecord},
	"banners":    {resource: "banners", mapRecord: appcuesBannerRecord},
}

// appcuesStreams returns the connector's published stream catalog. Every Appcues
// resource exposes a string "id" primary key. The list endpoints are full-refresh
// (the upstream connector advertises only full_refresh), and where a resource
// carries an "updatedAt" timestamp it is surfaced as the cursor field.
func appcuesStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "flows",
			Description:  "Appcues flows (in-app guidance experiences).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       appcuesFlowFields(),
		},
		{
			Name:         "segments",
			Description:  "Appcues audience segments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       appcuesSegmentFields(),
		},
		{
			Name:         "tags",
			Description:  "Appcues content tags.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       appcuesTagFields(),
		},
		{
			Name:         "checklists",
			Description:  "Appcues checklists.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       appcuesChecklistFields(),
		},
		{
			Name:         "banners",
			Description:  "Appcues banners.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       appcuesBannerFields(),
		},
	}
}

func appcuesFlowFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "published", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
		{Name: "createdBy", Type: "string"},
		{Name: "updatedBy", Type: "string"},
	}
}

func appcuesSegmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func appcuesTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func appcuesChecklistFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "published", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func appcuesBannerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "published", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func appcuesFlowRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"state":     item["state"],
		"published": item["published"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
		"createdBy": item["createdBy"],
		"updatedBy": item["updatedBy"],
	}
}

func appcuesSegmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"createdAt":   item["createdAt"],
		"updatedAt":   item["updatedAt"],
	}
}

func appcuesTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}

func appcuesChecklistRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"state":     item["state"],
		"published": item["published"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}

func appcuesBannerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"state":     item["state"],
		"published": item["published"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}
