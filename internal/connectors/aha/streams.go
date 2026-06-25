package aha

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Aha! API resource path segment (under
// /api/v1) it reads from, the JSON envelope key that holds its records array, and
// the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Aha! list endpoint path segment (e.g. "features"). The full
	// path is api/v1/<resource>.
	resource string
	// recordsKey is the top-level array key in the response envelope. Aha! keys
	// the array by resource name (e.g. {"features":[...]}); usually identical to
	// resource.
	recordsKey string
	// mapRecord flattens a raw Aha! object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// ahaStreamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in ahaStreams; the read path is fully
// data-driven from this table.
var ahaStreamEndpoints = map[string]streamEndpoint{
	"features":    {resource: "features", recordsKey: "features", mapRecord: ahaFeatureRecord},
	"products":    {resource: "products", recordsKey: "products", mapRecord: ahaProductRecord},
	"ideas":       {resource: "ideas", recordsKey: "ideas", mapRecord: ahaIdeaRecord},
	"releases":    {resource: "releases", recordsKey: "releases", mapRecord: ahaReleaseRecord},
	"initiatives": {resource: "initiatives", recordsKey: "initiatives", mapRecord: ahaInitiativeRecord},
	"goals":       {resource: "goals", recordsKey: "goals", mapRecord: ahaGoalRecord},
}

// ahaStreams returns the connector's published stream catalog. Every Aha! object
// exposes a string id and ISO-8601 created_at/updated_at timestamps, so the
// primary key is ["id"] and the incremental cursor field is ["updated_at"]
// throughout.
func ahaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "features",
			Description:  "Aha! features.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       ahaFeatureFields(),
		},
		{
			Name:         "products",
			Description:  "Aha! products (workspaces).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       ahaProductFields(),
		},
		{
			Name:         "ideas",
			Description:  "Aha! ideas.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       ahaIdeaFields(),
		},
		{
			Name:         "releases",
			Description:  "Aha! releases.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       ahaReleaseFields(),
		},
		{
			Name:         "initiatives",
			Description:  "Aha! initiatives.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       ahaInitiativeFields(),
		},
		{
			Name:         "goals",
			Description:  "Aha! goals.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       ahaGoalFields(),
		},
	}
}

func ahaFeatureFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "reference_num", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "start_date", Type: "string"},
		{Name: "due_date", Type: "string"},
		{Name: "workflow_status", Type: "object"},
		{Name: "url", Type: "string"},
		{Name: "resource", Type: "string"},
	}
}

func ahaProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "reference_prefix", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "product_line", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "url", Type: "string"},
		{Name: "resource", Type: "string"},
	}
}

func ahaIdeaFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "reference_num", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "score", Type: "number"},
		{Name: "votes", Type: "integer"},
		{Name: "workflow_status", Type: "object"},
		{Name: "url", Type: "string"},
		{Name: "resource", Type: "string"},
	}
}

func ahaReleaseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "reference_num", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "start_date", Type: "string"},
		{Name: "release_date", Type: "string"},
		{Name: "released", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "url", Type: "string"},
		{Name: "resource", Type: "string"},
	}
}

func ahaInitiativeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "reference_num", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "workflow_status", Type: "object"},
		{Name: "url", Type: "string"},
		{Name: "resource", Type: "string"},
	}
}

func ahaGoalFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "reference_num", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "workflow_status", Type: "object"},
		{Name: "url", Type: "string"},
		{Name: "resource", Type: "string"},
	}
}

func ahaFeatureRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"reference_num":   item["reference_num"],
		"name":            item["name"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
		"start_date":      item["start_date"],
		"due_date":        item["due_date"],
		"workflow_status": item["workflow_status"],
		"url":             item["url"],
		"resource":        "feature",
	}
}

func ahaProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"reference_prefix": item["reference_prefix"],
		"name":             item["name"],
		"product_line":     item["product_line"],
		"created_at":       item["created_at"],
		"updated_at":       item["updated_at"],
		"url":              item["url"],
		"resource":         "product",
	}
}

func ahaIdeaRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"reference_num":   item["reference_num"],
		"name":            item["name"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
		"score":           item["score"],
		"votes":           item["votes"],
		"workflow_status": item["workflow_status"],
		"url":             item["url"],
		"resource":        "idea",
	}
}

func ahaReleaseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"reference_num": item["reference_num"],
		"name":          item["name"],
		"start_date":    item["start_date"],
		"release_date":  item["release_date"],
		"released":      item["released"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
		"url":           item["url"],
		"resource":      "release",
	}
}

func ahaInitiativeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"reference_num":   item["reference_num"],
		"name":            item["name"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
		"workflow_status": item["workflow_status"],
		"url":             item["url"],
		"resource":        "initiative",
	}
}

func ahaGoalRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"reference_num":   item["reference_num"],
		"name":            item["name"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
		"workflow_status": item["workflow_status"],
		"url":             item["url"],
		"resource":        "goal",
	}
}
