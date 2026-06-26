package getgist

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Gist API resource path (relative to
// base_url) it reads from, the JSON key holding the records array in the list
// response, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Gist list endpoint path segment (e.g. "contacts").
	resource string
	// recordsKey is the JSON key under which the list response nests its array
	// (Gist keys arrays by resource name, e.g. {"contacts":[...]}).
	recordsKey string
	// mapRecord flattens a raw Gist object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// getgistStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in getgistStreams; the read
// path is fully data-driven from this table.
var getgistStreamEndpoints = map[string]streamEndpoint{
	"contacts":  {resource: "contacts", recordsKey: "contacts", mapRecord: getgistContactRecord},
	"tags":      {resource: "tags", recordsKey: "tags", mapRecord: getgistTagRecord},
	"segments":  {resource: "segments", recordsKey: "segments", mapRecord: getgistSegmentRecord},
	"campaigns": {resource: "campaigns", recordsKey: "campaigns", mapRecord: getgistCampaignRecord},
	"forms":     {resource: "forms", recordsKey: "forms", mapRecord: getgistFormRecord},
	"teammates": {resource: "teammates", recordsKey: "teammates", mapRecord: getgistTeammateRecord},
}

// getgistStreams returns the connector's published stream catalog. Every Gist
// object exposes an id, so the primary key is ["id"] across the board. Contacts
// carry an updated_at timestamp usable as an incremental cursor.
func getgistStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "Gist contacts (leads and users).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       getgistContactFields(),
		},
		{
			Name:        "tags",
			Description: "Gist tags applied to contacts.",
			PrimaryKey:  []string{"id"},
			Fields:      getgistTagFields(),
		},
		{
			Name:        "segments",
			Description: "Gist contact segments.",
			PrimaryKey:  []string{"id"},
			Fields:      getgistSegmentFields(),
		},
		{
			Name:        "campaigns",
			Description: "Gist email campaigns.",
			PrimaryKey:  []string{"id"},
			Fields:      getgistCampaignFields(),
		},
		{
			Name:        "forms",
			Description: "Gist forms.",
			PrimaryKey:  []string{"id"},
			Fields:      getgistFormFields(),
		},
		{
			Name:        "teammates",
			Description: "Gist workspace teammates.",
			PrimaryKey:  []string{"id"},
			Fields:      getgistTeammateFields(),
		},
	}
}

func getgistContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "signed_up_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "last_seen_at", Type: "integer"},
		{Name: "last_contacted_at", Type: "integer"},
		{Name: "session_count", Type: "integer"},
		{Name: "unsubscribed_from_emails", Type: "boolean"},
	}
}

func getgistTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

func getgistSegmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "person_type", Type: "string"},
		{Name: "count", Type: "integer"},
	}
}

func getgistCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func getgistFormFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func getgistTeammateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
	}
}

func getgistContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                       item["id"],
		"type":                     item["type"],
		"user_id":                  item["user_id"],
		"email":                    item["email"],
		"name":                     item["name"],
		"phone":                    item["phone"],
		"created_at":               item["created_at"],
		"signed_up_at":             item["signed_up_at"],
		"updated_at":               item["updated_at"],
		"last_seen_at":             item["last_seen_at"],
		"last_contacted_at":        item["last_contacted_at"],
		"session_count":            item["session_count"],
		"unsubscribed_from_emails": item["unsubscribed_from_emails"],
	}
}

func getgistTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"type": item["type"],
		"name": item["name"],
	}
}

func getgistSegmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"type":        item["type"],
		"name":        item["name"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
		"person_type": item["person_type"],
		"count":       item["count"],
	}
}

func getgistCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"type":       item["type"],
		"name":       item["name"],
		"subject":    item["subject"],
		"status":     item["status"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func getgistFormRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"type":       item["type"],
		"name":       item["name"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func getgistTeammateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    item["id"],
		"type":  item["type"],
		"name":  item["name"],
		"email": item["email"],
	}
}
