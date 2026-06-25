package intercom

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Intercom API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Intercom list endpoint path segment (e.g. "contacts").
	resource string
	// mapRecord flattens a raw Intercom object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// intercomStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in intercomStreams; the read
// path is fully data-driven from this table.
var intercomStreamEndpoints = map[string]streamEndpoint{
	"contacts":      {resource: "contacts", mapRecord: intercomContactRecord},
	"companies":     {resource: "companies", mapRecord: intercomCompanyRecord},
	"conversations": {resource: "conversations", mapRecord: intercomConversationRecord},
	"admins":        {resource: "admins", mapRecord: intercomAdminRecord},
	"tags":          {resource: "tags", mapRecord: intercomTagRecord},
}

// intercomStreams returns the connector's published stream catalog. Most
// Intercom objects expose a string id and unix created_at/updated_at timestamps,
// so the primary key is ["id"] and the incremental cursor field is
// ["updated_at"] where available (admins/tags have no updated_at, so they are
// full-refresh only).
func intercomStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "Intercom contacts (users and leads).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       intercomContactFields(),
		},
		{
			Name:         "companies",
			Description:  "Intercom companies.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       intercomCompanyFields(),
		},
		{
			Name:         "conversations",
			Description:  "Intercom conversations.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       intercomConversationFields(),
		},
		{
			Name:         "admins",
			Description:  "Intercom workspace admins (teammates).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       intercomAdminFields(),
		},
		{
			Name:         "tags",
			Description:  "Intercom tags.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       intercomTagFields(),
		},
	}
}

func intercomContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "external_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "last_seen_at", Type: "integer"},
		{Name: "signed_up_at", Type: "integer"},
		{Name: "owner_id", Type: "integer"},
		{Name: "unsubscribed_from_emails", Type: "boolean"},
	}
}

func intercomCompanyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "company_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "last_request_at", Type: "integer"},
		{Name: "monthly_spend", Type: "number"},
		{Name: "session_count", Type: "integer"},
		{Name: "user_count", Type: "integer"},
		{Name: "size", Type: "integer"},
		{Name: "website", Type: "string"},
		{Name: "industry", Type: "string"},
	}
}

func intercomConversationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "open", Type: "boolean"},
		{Name: "read", Type: "boolean"},
		{Name: "priority", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "waiting_since", Type: "integer"},
		{Name: "snoozed_until", Type: "integer"},
		{Name: "admin_assignee_id", Type: "integer"},
	}
}

func intercomAdminFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "away_mode_enabled", Type: "boolean"},
		{Name: "away_mode_reassign", Type: "boolean"},
		{Name: "has_inbox_seat", Type: "boolean"},
	}
}

func intercomTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

func intercomContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                       item["id"],
		"type":                     item["type"],
		"role":                     item["role"],
		"external_id":              item["external_id"],
		"email":                    item["email"],
		"phone":                    item["phone"],
		"name":                     item["name"],
		"created_at":               item["created_at"],
		"updated_at":               item["updated_at"],
		"last_seen_at":             item["last_seen_at"],
		"signed_up_at":             item["signed_up_at"],
		"owner_id":                 item["owner_id"],
		"unsubscribed_from_emails": item["unsubscribed_from_emails"],
	}
}

func intercomCompanyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"type":            item["type"],
		"company_id":      item["company_id"],
		"name":            item["name"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
		"last_request_at": item["last_request_at"],
		"monthly_spend":   item["monthly_spend"],
		"session_count":   item["session_count"],
		"user_count":      item["user_count"],
		"size":            item["size"],
		"website":         item["website"],
		"industry":        item["industry"],
	}
}

func intercomConversationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"type":              item["type"],
		"title":             item["title"],
		"state":             item["state"],
		"open":              item["open"],
		"read":              item["read"],
		"priority":          item["priority"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
		"waiting_since":     item["waiting_since"],
		"snoozed_until":     item["snoozed_until"],
		"admin_assignee_id": item["admin_assignee_id"],
	}
}

func intercomAdminRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"type":               item["type"],
		"name":               item["name"],
		"email":              item["email"],
		"job_title":          item["job_title"],
		"away_mode_enabled":  item["away_mode_enabled"],
		"away_mode_reassign": item["away_mode_reassign"],
		"has_inbox_seat":     item["has_inbox_seat"],
	}
}

func intercomTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"type": item["type"],
		"name": item["name"],
	}
}
