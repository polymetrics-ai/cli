package copper

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Copper search resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
// Copper lists are read with POST /<resource>/search and a body that carries
// page_number/page_size.
type streamEndpoint struct {
	// resource is the Copper resource path segment (e.g. "people"); the connector
	// appends "/search" at request time.
	resource string
	// mapRecord flattens a raw Copper object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// copperStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in copperStreams; the read path
// is fully data-driven from this table.
var copperStreamEndpoints = map[string]streamEndpoint{
	"people":        {resource: "people", mapRecord: copperPersonRecord},
	"companies":     {resource: "companies", mapRecord: copperCompanyRecord},
	"opportunities": {resource: "opportunities", mapRecord: copperOpportunityRecord},
	"leads":         {resource: "leads", mapRecord: copperLeadRecord},
	"tasks":         {resource: "tasks", mapRecord: copperTaskRecord},
}

// copperStreams returns the connector's published stream catalog. Every Copper
// object exposes a numeric id and a unix `date_modified` timestamp, so the
// primary key is ["id"] and the incremental cursor field is ["date_modified"]
// across the board.
func copperStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "people",
			Description:  "Copper people (contacts).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modified"},
			Fields:       copperPersonFields(),
		},
		{
			Name:         "companies",
			Description:  "Copper companies.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modified"},
			Fields:       copperCompanyFields(),
		},
		{
			Name:         "opportunities",
			Description:  "Copper opportunities (deals).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modified"},
			Fields:       copperOpportunityFields(),
		},
		{
			Name:         "leads",
			Description:  "Copper leads.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modified"},
			Fields:       copperLeadFields(),
		},
		{
			Name:         "tasks",
			Description:  "Copper tasks.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modified"},
			Fields:       copperTaskFields(),
		},
	}
}

func copperPersonFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "prefix", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "company_id", Type: "integer"},
		{Name: "company_name", Type: "string"},
		{Name: "emails", Type: "array"},
		{Name: "phone_numbers", Type: "array"},
		{Name: "assignee_id", Type: "integer"},
		{Name: "contact_type_id", Type: "integer"},
		{Name: "date_created", Type: "integer"},
		{Name: "date_modified", Type: "integer"},
	}
}

func copperCompanyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "assignee_id", Type: "integer"},
		{Name: "email_domain", Type: "string"},
		{Name: "phone_numbers", Type: "array"},
		{Name: "websites", Type: "array"},
		{Name: "address", Type: "object"},
		{Name: "details", Type: "string"},
		{Name: "date_created", Type: "integer"},
		{Name: "date_modified", Type: "integer"},
	}
}

func copperOpportunityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "company_id", Type: "integer"},
		{Name: "company_name", Type: "string"},
		{Name: "primary_contact_id", Type: "integer"},
		{Name: "assignee_id", Type: "integer"},
		{Name: "pipeline_id", Type: "integer"},
		{Name: "pipeline_stage_id", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "monetary_value", Type: "number"},
		{Name: "win_probability", Type: "number"},
		{Name: "close_date", Type: "string"},
		{Name: "date_created", Type: "integer"},
		{Name: "date_modified", Type: "integer"},
	}
}

func copperLeadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "company_name", Type: "string"},
		{Name: "email", Type: "object"},
		{Name: "phone_numbers", Type: "array"},
		{Name: "assignee_id", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "status_id", Type: "integer"},
		{Name: "monetary_value", Type: "number"},
		{Name: "date_created", Type: "integer"},
		{Name: "date_modified", Type: "integer"},
	}
}

func copperTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "related_resource", Type: "object"},
		{Name: "assignee_id", Type: "integer"},
		{Name: "due_date", Type: "integer"},
		{Name: "reminder_date", Type: "integer"},
		{Name: "completed_date", Type: "integer"},
		{Name: "priority", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "details", Type: "string"},
		{Name: "date_created", Type: "integer"},
		{Name: "date_modified", Type: "integer"},
	}
}

func copperPersonRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"prefix":          item["prefix"],
		"first_name":      item["first_name"],
		"last_name":       item["last_name"],
		"title":           item["title"],
		"company_id":      item["company_id"],
		"company_name":    item["company_name"],
		"emails":          item["emails"],
		"phone_numbers":   item["phone_numbers"],
		"assignee_id":     item["assignee_id"],
		"contact_type_id": item["contact_type_id"],
		"date_created":    item["date_created"],
		"date_modified":   item["date_modified"],
	}
}

func copperCompanyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"assignee_id":   item["assignee_id"],
		"email_domain":  item["email_domain"],
		"phone_numbers": item["phone_numbers"],
		"websites":      item["websites"],
		"address":       item["address"],
		"details":       item["details"],
		"date_created":  item["date_created"],
		"date_modified": item["date_modified"],
	}
}

func copperOpportunityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"company_id":         item["company_id"],
		"company_name":       item["company_name"],
		"primary_contact_id": item["primary_contact_id"],
		"assignee_id":        item["assignee_id"],
		"pipeline_id":        item["pipeline_id"],
		"pipeline_stage_id":  item["pipeline_stage_id"],
		"status":             item["status"],
		"monetary_value":     item["monetary_value"],
		"win_probability":    item["win_probability"],
		"close_date":         item["close_date"],
		"date_created":       item["date_created"],
		"date_modified":      item["date_modified"],
	}
}

func copperLeadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"title":          item["title"],
		"company_name":   item["company_name"],
		"email":          item["email"],
		"phone_numbers":  item["phone_numbers"],
		"assignee_id":    item["assignee_id"],
		"status":         item["status"],
		"status_id":      item["status_id"],
		"monetary_value": item["monetary_value"],
		"date_created":   item["date_created"],
		"date_modified":  item["date_modified"],
	}
}

func copperTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"related_resource": item["related_resource"],
		"assignee_id":      item["assignee_id"],
		"due_date":         item["due_date"],
		"reminder_date":    item["reminder_date"],
		"completed_date":   item["completed_date"],
		"priority":         item["priority"],
		"status":           item["status"],
		"details":          item["details"],
		"date_created":     item["date_created"],
		"date_modified":    item["date_modified"],
	}
}
