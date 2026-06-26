package freshservice

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Freshservice API resource path
// (relative to base_url, e.g. "tickets"), the JSON wrapper key that holds the
// record array in a list response (e.g. {"tickets":[...]}), and the record
// mapper that flattens its objects.
type streamEndpoint struct {
	resource  string
	listKey   string
	mapRecord func(map[string]any) connectors.Record
}

// freshserviceStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in freshserviceStreams;
// the read path is fully data-driven from this table.
var freshserviceStreamEndpoints = map[string]streamEndpoint{
	"tickets":    {resource: "tickets", listKey: "tickets", mapRecord: freshserviceTicketRecord},
	"agents":     {resource: "agents", listKey: "agents", mapRecord: freshserviceAgentRecord},
	"requesters": {resource: "requesters", listKey: "requesters", mapRecord: freshserviceRequesterRecord},
	"assets":     {resource: "assets", listKey: "assets", mapRecord: freshserviceAssetRecord},
	"problems":   {resource: "problems", listKey: "problems", mapRecord: freshserviceProblemRecord},
}

// freshserviceStreams returns the connector's published stream catalog. Every
// Freshservice object exposes an integer id and RFC3339 created_at/updated_at
// timestamps, so the primary key is ["id"] and the incremental cursor field is
// ["updated_at"] across the board.
func freshserviceStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "tickets",
			Description:  "Freshservice service desk tickets.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshserviceTicketFields(),
		},
		{
			Name:         "agents",
			Description:  "Freshservice agents.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshserviceAgentFields(),
		},
		{
			Name:         "requesters",
			Description:  "Freshservice requesters (end users).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshserviceRequesterFields(),
		},
		{
			Name:         "assets",
			Description:  "Freshservice CMDB assets.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshserviceAssetFields(),
		},
		{
			Name:         "problems",
			Description:  "Freshservice problems.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshserviceProblemFields(),
		},
	}
}

func freshserviceTicketFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "subject", Type: "string"},
		{Name: "description_text", Type: "string"},
		{Name: "status", Type: "integer"},
		{Name: "priority", Type: "integer"},
		{Name: "source", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "requester_id", Type: "integer"},
		{Name: "responder_id", Type: "integer"},
		{Name: "group_id", Type: "integer"},
		{Name: "department_id", Type: "integer"},
		{Name: "spam", Type: "boolean"},
		{Name: "due_by", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshserviceAgentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "occasional", Type: "boolean"},
		{Name: "job_title", Type: "string"},
		{Name: "department_ids", Type: "array"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshserviceRequesterFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "primary_email", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "department_ids", Type: "array"},
		{Name: "active", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshserviceAssetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "display_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "asset_type_id", Type: "integer"},
		{Name: "asset_tag", Type: "string"},
		{Name: "impact", Type: "string"},
		{Name: "user_id", Type: "integer"},
		{Name: "department_id", Type: "integer"},
		{Name: "agent_id", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshserviceProblemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "subject", Type: "string"},
		{Name: "status", Type: "integer"},
		{Name: "priority", Type: "integer"},
		{Name: "impact", Type: "integer"},
		{Name: "requester_id", Type: "integer"},
		{Name: "agent_id", Type: "integer"},
		{Name: "group_id", Type: "integer"},
		{Name: "department_id", Type: "integer"},
		{Name: "due_by", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshserviceTicketRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"subject":          item["subject"],
		"description_text": item["description_text"],
		"status":           item["status"],
		"priority":         item["priority"],
		"source":           item["source"],
		"type":             item["type"],
		"requester_id":     item["requester_id"],
		"responder_id":     item["responder_id"],
		"group_id":         item["group_id"],
		"department_id":    item["department_id"],
		"spam":             item["spam"],
		"due_by":           item["due_by"],
		"created_at":       item["created_at"],
		"updated_at":       item["updated_at"],
	}
}

func freshserviceAgentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"first_name":     item["first_name"],
		"last_name":      item["last_name"],
		"email":          item["email"],
		"active":         item["active"],
		"occasional":     item["occasional"],
		"job_title":      item["job_title"],
		"department_ids": item["department_ids"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}

func freshserviceRequesterRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"first_name":     item["first_name"],
		"last_name":      item["last_name"],
		"primary_email":  item["primary_email"],
		"job_title":      item["job_title"],
		"department_ids": item["department_ids"],
		"active":         item["active"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}

func freshserviceAssetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"display_id":    item["display_id"],
		"name":          item["name"],
		"asset_type_id": item["asset_type_id"],
		"asset_tag":     item["asset_tag"],
		"impact":        item["impact"],
		"user_id":       item["user_id"],
		"department_id": item["department_id"],
		"agent_id":      item["agent_id"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func freshserviceProblemRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"subject":       item["subject"],
		"status":        item["status"],
		"priority":      item["priority"],
		"impact":        item["impact"],
		"requester_id":  item["requester_id"],
		"agent_id":      item["agent_id"],
		"group_id":      item["group_id"],
		"department_id": item["department_id"],
		"due_by":        item["due_by"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}
