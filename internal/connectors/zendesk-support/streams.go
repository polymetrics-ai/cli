package zendesksupport

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to its Zendesk API resource. resource is the
// path segment under /api/v2 (e.g. "tickets"); recordsKey is the JSON object key
// the collection array lives under in the response body (e.g. "tickets"). The
// mapper flattens a raw object into a connectors.Record.
type streamEndpoint struct {
	resource   string
	recordsKey string
	mapRecord  func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Every entry is a Zendesk
// collection endpoint that supports cursor pagination (page[size]/page[after]
// with meta.after_cursor / meta.has_more). Adding a stream means adding one entry
// here plus a Stream definition in streams().
var streamEndpoints = map[string]streamEndpoint{
	"tickets":              {resource: "tickets", recordsKey: "tickets", mapRecord: ticketRecord},
	"users":                {resource: "users", recordsKey: "users", mapRecord: userRecord},
	"organizations":        {resource: "organizations", recordsKey: "organizations", mapRecord: organizationRecord},
	"groups":               {resource: "groups", recordsKey: "groups", mapRecord: groupRecord},
	"satisfaction_ratings": {resource: "satisfaction_ratings", recordsKey: "satisfaction_ratings", mapRecord: satisfactionRatingRecord},
}

// streams returns the connector's published stream catalog. Every Zendesk object
// exposes a numeric id and an updated_at timestamp, so the primary key is ["id"]
// and the incremental cursor field is ["updated_at"] across the board.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "tickets",
			Description:  "Zendesk Support tickets.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       ticketFields(),
		},
		{
			Name:         "users",
			Description:  "Zendesk Support users (agents and end users).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       userFields(),
		},
		{
			Name:         "organizations",
			Description:  "Zendesk Support organizations.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       organizationFields(),
		},
		{
			Name:         "groups",
			Description:  "Zendesk Support agent groups.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       groupFields(),
		},
		{
			Name:         "satisfaction_ratings",
			Description:  "Zendesk Support satisfaction ratings.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       satisfactionRatingFields(),
		},
	}
}

func ticketFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "subject", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "priority", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "requester_id", Type: "integer"},
		{Name: "assignee_id", Type: "integer"},
		{Name: "organization_id", Type: "integer"},
		{Name: "group_id", Type: "integer"},
		{Name: "brand_id", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "verified", Type: "boolean"},
		{Name: "organization_id", Type: "integer"},
		{Name: "time_zone", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func organizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "details", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "group_id", Type: "integer"},
		{Name: "shared_tickets", Type: "boolean"},
		{Name: "shared_comments", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func groupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "default", Type: "boolean"},
		{Name: "deleted", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func satisfactionRatingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "score", Type: "string"},
		{Name: "comment", Type: "string"},
		{Name: "reason", Type: "string"},
		{Name: "ticket_id", Type: "integer"},
		{Name: "assignee_id", Type: "integer"},
		{Name: "requester_id", Type: "integer"},
		{Name: "group_id", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func ticketRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"subject":         item["subject"],
		"description":     item["description"],
		"status":          item["status"],
		"priority":        item["priority"],
		"type":            item["type"],
		"requester_id":    item["requester_id"],
		"assignee_id":     item["assignee_id"],
		"organization_id": item["organization_id"],
		"group_id":        item["group_id"],
		"brand_id":        item["brand_id"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"email":           item["email"],
		"role":            item["role"],
		"phone":           item["phone"],
		"active":          item["active"],
		"verified":        item["verified"],
		"organization_id": item["organization_id"],
		"time_zone":       item["time_zone"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
	}
}

func organizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"details":         item["details"],
		"notes":           item["notes"],
		"group_id":        item["group_id"],
		"shared_tickets":  item["shared_tickets"],
		"shared_comments": item["shared_comments"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
	}
}

func groupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"default":     item["default"],
		"deleted":     item["deleted"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func satisfactionRatingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"score":        item["score"],
		"comment":      item["comment"],
		"reason":       item["reason"],
		"ticket_id":    item["ticket_id"],
		"assignee_id":  item["assignee_id"],
		"requester_id": item["requester_id"],
		"group_id":     item["group_id"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}
