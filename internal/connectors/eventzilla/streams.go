package eventzilla

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Eventzilla API resource it reads.
//
// Eventzilla list endpoints return a JSON object whose records live under a
// named field path (e.g. {"events":[...]}). Some streams are children of an
// event: they live at /events/{event_id}/<child> and must be fanned out across
// every event id. The fields here drive a fully data-driven read path.
type streamEndpoint struct {
	// fieldPath is the JSON key under which the records array lives.
	fieldPath string
	// parentScoped marks a substream read per-event at
	// /events/{event_id}/<child>. When false the resource is a top-level path.
	parentScoped bool
	// resource is the top-level path segment (parentScoped=false) or the child
	// segment appended after /events/{event_id}/ (parentScoped=true).
	resource string
	// mapRecord flattens a raw Eventzilla object into a connectors.Record.
	mapRecord func(item map[string]any) connectors.Record
}

// eventzillaStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in eventzillaStreams.
var eventzillaStreamEndpoints = map[string]streamEndpoint{
	"events":     {resource: "events", fieldPath: "events", mapRecord: eventRecord},
	"categories": {resource: "categories", fieldPath: "categories", mapRecord: categoryRecord},
	"users":      {resource: "users", fieldPath: "users", mapRecord: userRecord},
	"attendees":  {resource: "attendees", fieldPath: "attendees", parentScoped: true, mapRecord: attendeeRecord},
	"tickets":    {resource: "tickets", fieldPath: "tickets", parentScoped: true, mapRecord: ticketRecord},
}

// eventzillaStreams returns the connector's published stream catalog covering
// the core Eventzilla resources. Eventzilla supports only full-refresh sync, so
// no cursor fields are declared.
func eventzillaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "events",
			Description: "Eventzilla events for the authenticated organizer.",
			PrimaryKey:  []string{"id"},
			Fields:      eventFields(),
		},
		{
			Name:        "categories",
			Description: "Eventzilla event categories.",
			PrimaryKey:  []string{"category"},
			Fields:      categoryFields(),
		},
		{
			Name:        "users",
			Description: "Eventzilla user accounts.",
			PrimaryKey:  []string{"id"},
			Fields:      userFields(),
		},
		{
			Name:        "attendees",
			Description: "Eventzilla attendees, per event.",
			PrimaryKey:  []string{"id"},
			Fields:      attendeeFields(),
		},
		{
			Name:        "tickets",
			Description: "Eventzilla ticket types, per event.",
			PrimaryKey:  []string{"id"},
			Fields:      ticketFields(),
		},
	}
}

func eventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "venue", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "start_date", Type: "string"},
		{Name: "start_time", Type: "string"},
		{Name: "end_date", Type: "string"},
		{Name: "end_time", Type: "string"},
		{Name: "time_zone", Type: "string"},
		{Name: "categories", Type: "string"},
		{Name: "tickets_sold", Type: "integer"},
		{Name: "tickets_total", Type: "integer"},
	}
}

func categoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "category", Type: "string"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "username", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "user_type", Type: "string"},
		{Name: "phone_primary", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "last_seen", Type: "string"},
	}
}

func attendeeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "event_id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "ticket_type", Type: "string"},
		{Name: "refno", Type: "string"},
		{Name: "transaction_amount", Type: "number"},
		{Name: "transaction_status", Type: "string"},
		{Name: "transaction_date", Type: "string"},
		{Name: "is_attended", Type: "string"},
	}
}

func ticketFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "event_id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "ticket_type", Type: "string"},
		{Name: "price", Type: "number"},
		{Name: "quantity_total", Type: "integer"},
		{Name: "is_visible", Type: "boolean"},
		{Name: "sales_start_date", Type: "string"},
		{Name: "sales_end_date", Type: "string"},
	}
}

func eventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"title":         item["title"],
		"status":        item["status"],
		"url":           item["url"],
		"venue":         item["venue"],
		"currency":      item["currency"],
		"start_date":    item["start_date"],
		"start_time":    item["start_time"],
		"end_date":      item["end_date"],
		"end_time":      item["end_time"],
		"time_zone":     item["time_zone"],
		"categories":    item["categories"],
		"tickets_sold":  item["tickets_sold"],
		"tickets_total": item["tickets_total"],
	}
}

func categoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"category": item["category"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"username":      item["username"],
		"first_name":    item["first_name"],
		"last_name":     item["last_name"],
		"email":         item["email"],
		"company":       item["company"],
		"user_type":     item["user_type"],
		"phone_primary": item["phone_primary"],
		"timezone":      item["timezone"],
		"last_seen":     item["last_seen"],
	}
}

func attendeeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"event_id":           item["event_id"],
		"first_name":         item["first_name"],
		"last_name":          item["last_name"],
		"email":              item["email"],
		"ticket_type":        item["ticket_type"],
		"refno":              item["refno"],
		"transaction_amount": item["transaction_amount"],
		"transaction_status": item["transaction_status"],
		"transaction_date":   item["transaction_date"],
		"is_attended":        item["is_attended"],
	}
}

func ticketRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"event_id":         item["event_id"],
		"title":            item["title"],
		"ticket_type":      item["ticket_type"],
		"price":            item["price"],
		"quantity_total":   item["quantity_total"],
		"is_visible":       item["is_visible"],
		"sales_start_date": item["sales_start_date"],
		"sales_end_date":   item["sales_end_date"],
	}
}
