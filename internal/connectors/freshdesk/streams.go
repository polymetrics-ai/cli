package freshdesk

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Freshdesk API resource path (relative
// to base_url, e.g. "tickets") and the record mapper that flattens its objects.
type streamEndpoint struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
}

// freshdeskStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in freshdeskStreams; the read
// path is fully data-driven from this table.
var freshdeskStreamEndpoints = map[string]streamEndpoint{
	"tickets":   {resource: "tickets", mapRecord: freshdeskTicketRecord},
	"contacts":  {resource: "contacts", mapRecord: freshdeskContactRecord},
	"companies": {resource: "companies", mapRecord: freshdeskCompanyRecord},
	"agents":    {resource: "agents", mapRecord: freshdeskAgentRecord},
	"groups":    {resource: "groups", mapRecord: freshdeskGroupRecord},
}

// freshdeskStreams returns the connector's published stream catalog. Every
// Freshdesk object exposes an integer id and RFC3339 created_at/updated_at
// timestamps, so the primary key is ["id"] and the incremental cursor field is
// ["updated_at"] across the board.
func freshdeskStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "tickets",
			Description:  "Freshdesk support tickets.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshdeskTicketFields(),
		},
		{
			Name:         "contacts",
			Description:  "Freshdesk contacts (requesters).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshdeskContactFields(),
		},
		{
			Name:         "companies",
			Description:  "Freshdesk companies.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshdeskCompanyFields(),
		},
		{
			Name:         "agents",
			Description:  "Freshdesk agents.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshdeskAgentFields(),
		},
		{
			Name:         "groups",
			Description:  "Freshdesk agent groups.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshdeskGroupFields(),
		},
	}
}

func freshdeskTicketFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "subject", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "integer"},
		{Name: "priority", Type: "integer"},
		{Name: "source", Type: "integer"},
		{Name: "requester_id", Type: "integer"},
		{Name: "responder_id", Type: "integer"},
		{Name: "group_id", Type: "integer"},
		{Name: "company_id", Type: "integer"},
		{Name: "spam", Type: "boolean"},
		{Name: "due_by", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshdeskContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "mobile", Type: "string"},
		{Name: "company_id", Type: "integer"},
		{Name: "active", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshdeskCompanyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "note", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshdeskAgentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "available", Type: "boolean"},
		{Name: "occasional", Type: "boolean"},
		{Name: "ticket_scope", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshdeskGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshdeskTicketRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"subject":      item["subject"],
		"type":         item["type"],
		"status":       item["status"],
		"priority":     item["priority"],
		"source":       item["source"],
		"requester_id": item["requester_id"],
		"responder_id": item["responder_id"],
		"group_id":     item["group_id"],
		"company_id":   item["company_id"],
		"spam":         item["spam"],
		"due_by":       item["due_by"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func freshdeskContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"email":      item["email"],
		"phone":      item["phone"],
		"mobile":     item["mobile"],
		"company_id": item["company_id"],
		"active":     item["active"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func freshdeskCompanyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"note":        item["note"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func freshdeskAgentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"available":    item["available"],
		"occasional":   item["occasional"],
		"ticket_scope": item["ticket_scope"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func freshdeskGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}
