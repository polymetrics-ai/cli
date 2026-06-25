package convertkit

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the ConvertKit v3 API resource path it
// reads from, the JSON key holding the records array in the list response, the
// record mapper, and whether the endpoint supports page-based pagination.
type streamEndpoint struct {
	// resource is the v3 list endpoint path segment (e.g. "subscribers").
	resource string
	// arrayKey is the JSON key in the list response holding the records array.
	// ConvertKit nests each resource list under a key named after the resource.
	arrayKey string
	// paginated is true for endpoints that return total_pages and accept a
	// "page" query param (subscribers, broadcasts); false for endpoints that
	// return the full collection in a single array (forms, tags, sequences).
	paginated bool
	// mapRecord flattens a raw ConvertKit object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// convertkitStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in convertkitStreams; the
// read path is fully data-driven from this table.
var convertkitStreamEndpoints = map[string]streamEndpoint{
	"subscribers": {resource: "subscribers", arrayKey: "subscribers", paginated: true, mapRecord: convertkitSubscriberRecord},
	"broadcasts":  {resource: "broadcasts", arrayKey: "broadcasts", paginated: true, mapRecord: convertkitBroadcastRecord},
	"forms":       {resource: "forms", arrayKey: "forms", paginated: false, mapRecord: convertkitFormRecord},
	"tags":        {resource: "tags", arrayKey: "tags", paginated: false, mapRecord: convertkitTagRecord},
	"sequences":   {resource: "sequences", arrayKey: "sequences", paginated: false, mapRecord: convertkitSequenceRecord},
}

// convertkitStreams returns the connector's published stream catalog. Every
// ConvertKit object exposes an integer id and an RFC3339 created_at timestamp, so
// the primary key is ["id"] and the cursor field is ["created_at"] across the
// board (the upstream connector is full-refresh only, but the cursor field is
// declared for downstream incremental dedupe).
func convertkitStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "subscribers",
			Description:  "ConvertKit subscribers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       convertkitSubscriberFields(),
		},
		{
			Name:         "forms",
			Description:  "ConvertKit forms.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       convertkitFormFields(),
		},
		{
			Name:         "sequences",
			Description:  "ConvertKit sequences (email courses).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       convertkitSequenceFields(),
		},
		{
			Name:         "tags",
			Description:  "ConvertKit subscriber tags.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       convertkitTagFields(),
		},
		{
			Name:         "broadcasts",
			Description:  "ConvertKit broadcasts (one-off emails).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       convertkitBroadcastFields(),
		},
	}
}

func convertkitSubscriberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "email_address", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func convertkitFormFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "format", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "created_at", Type: "string"},
	}
}

func convertkitSequenceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "hold", Type: "boolean"},
		{Name: "repeat", Type: "boolean"},
		{Name: "created_at", Type: "string"},
	}
}

func convertkitTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func convertkitBroadcastFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "subject", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "public", Type: "boolean"},
		{Name: "published_at", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func convertkitSubscriberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"first_name":    item["first_name"],
		"email_address": item["email_address"],
		"state":         item["state"],
		"created_at":    item["created_at"],
	}
}

func convertkitFormRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"type":       item["type"],
		"format":     item["format"],
		"archived":   item["archived"],
		"created_at": item["created_at"],
	}
}

func convertkitSequenceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"hold":       item["hold"],
		"repeat":     item["repeat"],
		"created_at": item["created_at"],
	}
}

func convertkitTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"created_at": item["created_at"],
	}
}

func convertkitBroadcastRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"subject":      item["subject"],
		"description":  item["description"],
		"public":       item["public"],
		"published_at": item["published_at"],
		"created_at":   item["created_at"],
	}
}
