package gorgias

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Gorgias API resource path (relative to
// base_url, e.g. "api/tickets") it reads from, and the record mapper that
// flattens its objects.
type streamEndpoint struct {
	// resource is the Gorgias list endpoint path segment (e.g. "api/tickets").
	resource string
	// mapRecord flattens a raw Gorgias object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// gorgiasStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in gorgiasStreams; the read path
// is fully data-driven from this table.
var gorgiasStreamEndpoints = map[string]streamEndpoint{
	"tickets":              {resource: "api/tickets", mapRecord: gorgiasTicketRecord},
	"customers":            {resource: "api/customers", mapRecord: gorgiasCustomerRecord},
	"messages":             {resource: "api/messages", mapRecord: gorgiasMessageRecord},
	"satisfaction-surveys": {resource: "api/satisfaction-surveys", mapRecord: gorgiasSurveyRecord},
}

// gorgiasStreams returns the connector's published stream catalog. Every Gorgias
// object exposes an integer id and most expose created_datetime / updated_datetime
// RFC3339 timestamps, so the primary key is ["id"] and the incremental cursor
// field is ["updated_datetime"] where available.
func gorgiasStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "tickets",
			Description:  "Gorgias support tickets.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_datetime"},
			Fields:       gorgiasTicketFields(),
		},
		{
			Name:         "customers",
			Description:  "Gorgias customers (end users who contact support).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_datetime"},
			Fields:       gorgiasCustomerFields(),
		},
		{
			Name:         "messages",
			Description:  "Gorgias ticket messages (inbound and outbound).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_datetime"},
			Fields:       gorgiasMessageFields(),
		},
		{
			Name:         "satisfaction-surveys",
			Description:  "Gorgias customer satisfaction (CSAT) surveys.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_datetime"},
			Fields:       gorgiasSurveyFields(),
		},
	}
}

func gorgiasTicketFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "subject", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "channel", Type: "string"},
		{Name: "via", Type: "string"},
		{Name: "priority", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "is_unread", Type: "boolean"},
		{Name: "spam", Type: "boolean"},
		{Name: "trashed_datetime", Type: "string"},
		{Name: "created_datetime", Type: "string"},
		{Name: "updated_datetime", Type: "string"},
		{Name: "opened_datetime", Type: "string"},
		{Name: "closed_datetime", Type: "string"},
	}
}

func gorgiasCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "external_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "firstname", Type: "string"},
		{Name: "lastname", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "channel", Type: "string"},
		{Name: "created_datetime", Type: "string"},
		{Name: "updated_datetime", Type: "string"},
	}
}

func gorgiasMessageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "ticket_id", Type: "integer"},
		{Name: "channel", Type: "string"},
		{Name: "via", Type: "string"},
		{Name: "from_agent", Type: "boolean"},
		{Name: "subject", Type: "string"},
		{Name: "body_text", Type: "string"},
		{Name: "stripped_text", Type: "string"},
		{Name: "public", Type: "boolean"},
		{Name: "created_datetime", Type: "string"},
		{Name: "sent_datetime", Type: "string"},
	}
}

func gorgiasSurveyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "ticket_id", Type: "integer"},
		{Name: "customer_id", Type: "integer"},
		{Name: "score", Type: "integer"},
		{Name: "scale_range", Type: "integer"},
		{Name: "body_text", Type: "string"},
		{Name: "created_datetime", Type: "string"},
		{Name: "sent_datetime", Type: "string"},
		{Name: "scored_datetime", Type: "string"},
	}
}

func gorgiasTicketRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"subject":          item["subject"],
		"status":           item["status"],
		"channel":          item["channel"],
		"via":              item["via"],
		"priority":         item["priority"],
		"language":         item["language"],
		"is_unread":        item["is_unread"],
		"spam":             item["spam"],
		"trashed_datetime": item["trashed_datetime"],
		"created_datetime": item["created_datetime"],
		"updated_datetime": item["updated_datetime"],
		"opened_datetime":  item["opened_datetime"],
		"closed_datetime":  item["closed_datetime"],
	}
}

func gorgiasCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"external_id":      item["external_id"],
		"email":            item["email"],
		"name":             item["name"],
		"firstname":        item["firstname"],
		"lastname":         item["lastname"],
		"language":         item["language"],
		"timezone":         item["timezone"],
		"channel":          item["channel"],
		"created_datetime": item["created_datetime"],
		"updated_datetime": item["updated_datetime"],
	}
}

func gorgiasMessageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"ticket_id":        item["ticket_id"],
		"channel":          item["channel"],
		"via":              item["via"],
		"from_agent":       item["from_agent"],
		"subject":          item["subject"],
		"body_text":        item["body_text"],
		"stripped_text":    item["stripped_text"],
		"public":           item["public"],
		"created_datetime": item["created_datetime"],
		"sent_datetime":    item["sent_datetime"],
	}
}

func gorgiasSurveyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"ticket_id":        item["ticket_id"],
		"customer_id":      item["customer_id"],
		"score":            item["score"],
		"scale_range":      item["scale_range"],
		"body_text":        item["body_text"],
		"created_datetime": item["created_datetime"],
		"sent_datetime":    item["sent_datetime"],
		"scored_datetime":  item["scored_datetime"],
	}
}
