package flexmail

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to its Flexmail API resource path (relative
// to base_url), the JSON path where the record array lives, the record mapper,
// and whether the endpoint paginates with offset/limit.
type streamEndpoint struct {
	// resource is the Flexmail API path segment (e.g. "contacts" -> /contacts).
	resource string
	// recordsPath is the dotted path to the record array in the response body.
	// Flexmail wraps collections in HAL style: {"_embedded":{"item":[...]}}.
	recordsPath string
	// paginated is true for endpoints that support offset/limit pagination
	// (contacts, sources). Others return their full collection in one response.
	paginated bool
	// mapRecord flattens a raw Flexmail object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// flexmailStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in flexmailStreams; the read
// path is fully data-driven from this table.
var flexmailStreamEndpoints = map[string]streamEndpoint{
	"contacts":      {resource: "contacts", recordsPath: "_embedded.item", paginated: true, mapRecord: flexmailContactRecord},
	"custom_fields": {resource: "custom-fields", recordsPath: "_embedded.item", paginated: false, mapRecord: flexmailCustomFieldRecord},
	"interests":     {resource: "interests", recordsPath: "_embedded.item", paginated: false, mapRecord: flexmailInterestRecord},
	"segments":      {resource: "segments", recordsPath: "_embedded.item", paginated: false, mapRecord: flexmailSegmentRecord},
	"sources":       {resource: "sources", recordsPath: "_embedded.item", paginated: true, mapRecord: flexmailSourceRecord},
}

// flexmailStreams returns the connector's published stream catalog. Every
// Flexmail object exposes an "id" primary key. None of the streams support
// incremental sync (Flexmail is full-refresh only), so CursorFields are empty.
func flexmailStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "contacts",
			Description: "Flexmail contacts (subscribers).",
			PrimaryKey:  []string{"id"},
			Fields:      flexmailContactFields(),
		},
		{
			Name:        "custom_fields",
			Description: "Flexmail custom contact fields.",
			PrimaryKey:  []string{"id"},
			Fields:      flexmailCustomFieldFields(),
		},
		{
			Name:        "interests",
			Description: "Flexmail interests (subscription preferences).",
			PrimaryKey:  []string{"id"},
			Fields:      flexmailInterestFields(),
		},
		{
			Name:        "segments",
			Description: "Flexmail audience segments.",
			PrimaryKey:  []string{"id"},
			Fields:      flexmailSegmentFields(),
		},
		{
			Name:        "sources",
			Description: "Flexmail contact sources.",
			PrimaryKey:  []string{"id"},
			Fields:      flexmailSourceFields(),
		},
	}
}

func flexmailContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "custom_fields", Type: "object"},
	}
}

func flexmailCustomFieldFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "placeholder", Type: "string"},
	}
}

func flexmailInterestFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "label", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "visibility", Type: "string"},
	}
}

func flexmailSegmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "number_of_contacts", Type: "integer"},
	}
}

func flexmailSourceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
	}
}

func flexmailContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"email":         item["email"],
		"name":          item["name"],
		"first_name":    item["first_name"],
		"language":      item["language"],
		"custom_fields": item["custom_fields"],
	}
}

func flexmailCustomFieldRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"type":        item["type"],
		"placeholder": item["placeholder"],
	}
}

func flexmailInterestRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"label":       item["label"],
		"description": item["description"],
		"visibility":  item["visibility"],
	}
}

func flexmailSegmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"number_of_contacts": item["number_of_contacts"],
	}
}

func flexmailSourceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"name": item["name"],
	}
}
