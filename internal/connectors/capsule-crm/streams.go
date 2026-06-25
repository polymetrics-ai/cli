package capsulecrm

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Capsule CRM v2 list endpoint path
// (relative to base_url), the JSON wrapper key holding the array of objects, and
// the record mapper that flattens each object into a connectors.Record.
type streamEndpoint struct {
	// resource is the API list path segment (e.g. "parties").
	resource string
	// recordsKey is the top-level JSON key wrapping the array (e.g. "parties").
	recordsKey string
	// mapRecord flattens a raw Capsule object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// capsuleStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in capsuleStreams; the read path
// is fully data-driven from this table.
var capsuleStreamEndpoints = map[string]streamEndpoint{
	"parties":       {resource: "parties", recordsKey: "parties", mapRecord: capsulePartyRecord},
	"opportunities": {resource: "opportunities", recordsKey: "opportunities", mapRecord: capsuleOpportunityRecord},
	"kases":         {resource: "kases", recordsKey: "kases", mapRecord: capsuleKaseRecord},
	"tasks":         {resource: "tasks", recordsKey: "tasks", mapRecord: capsuleTaskRecord},
	"users":         {resource: "users", recordsKey: "users", mapRecord: capsuleUserRecord},
}

// capsuleStreams returns the connector's published stream catalog. Every Capsule
// object exposes a numeric id and ISO-8601 createdAt/updatedAt timestamps, so the
// primary key is ["id"] and the incremental cursor field is ["updatedAt"].
func capsuleStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "parties",
			Description:  "Capsule CRM parties (people and organisations).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       capsulePartyFields(),
		},
		{
			Name:         "opportunities",
			Description:  "Capsule CRM sales opportunities.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       capsuleOpportunityFields(),
		},
		{
			Name:         "kases",
			Description:  "Capsule CRM cases (projects).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       capsuleKaseFields(),
		},
		{
			Name:         "tasks",
			Description:  "Capsule CRM tasks.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       capsuleTaskFields(),
		},
		{
			Name:         "users",
			Description:  "Capsule CRM users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       capsuleUserFields(),
		},
	}
}

func capsulePartyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "organisation_name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "about", Type: "string"},
		{Name: "owner", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "last_contacted_at", Type: "timestamp"},
	}
}

func capsuleOpportunityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "party_id", Type: "integer"},
		{Name: "milestone_id", Type: "integer"},
		{Name: "milestone_name", Type: "string"},
		{Name: "value_amount", Type: "number"},
		{Name: "value_currency", Type: "string"},
		{Name: "probability", Type: "number"},
		{Name: "expected_close_on", Type: "string"},
		{Name: "closed_on", Type: "string"},
		{Name: "lost_reason", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func capsuleKaseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "party_id", Type: "integer"},
		{Name: "owner", Type: "string"},
		{Name: "closed_on", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func capsuleTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "description", Type: "string"},
		{Name: "detail", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "due_on", Type: "string"},
		{Name: "category_id", Type: "integer"},
		{Name: "party_id", Type: "integer"},
		{Name: "opportunity_id", Type: "integer"},
		{Name: "kase_id", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func capsuleUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "username", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func capsulePartyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"type":              item["type"],
		"first_name":        item["firstName"],
		"last_name":         item["lastName"],
		"organisation_name": item["organisationName"],
		"title":             item["title"],
		"job_title":         item["jobTitle"],
		"about":             item["about"],
		"owner":             item["owner"],
		"created_at":        item["createdAt"],
		"updated_at":        item["updatedAt"],
		"last_contacted_at": item["lastContactedAt"],
	}
}

func capsuleOpportunityRecord(item map[string]any) connectors.Record {
	value := nestedObject(item["value"])
	milestone := nestedObject(item["milestone"])
	party := nestedObject(item["party"])
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"description":       item["description"],
		"party_id":          party["id"],
		"milestone_id":      milestone["id"],
		"milestone_name":    milestone["name"],
		"value_amount":      value["amount"],
		"value_currency":    value["currency"],
		"probability":       item["probability"],
		"expected_close_on": item["expectedCloseOn"],
		"closed_on":         item["closedOn"],
		"lost_reason":       item["lostReason"],
		"created_at":        item["createdAt"],
		"updated_at":        item["updatedAt"],
	}
}

func capsuleKaseRecord(item map[string]any) connectors.Record {
	party := nestedObject(item["party"])
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"status":      item["status"],
		"party_id":    party["id"],
		"owner":       item["owner"],
		"closed_on":   item["closedOn"],
		"created_at":  item["createdAt"],
		"updated_at":  item["updatedAt"],
	}
}

func capsuleTaskRecord(item map[string]any) connectors.Record {
	party := nestedObject(item["party"])
	opportunity := nestedObject(item["opportunity"])
	kase := nestedObject(item["kase"])
	category := nestedObject(item["category"])
	return connectors.Record{
		"id":             item["id"],
		"description":    item["description"],
		"detail":         item["detail"],
		"status":         item["status"],
		"due_on":         item["dueOn"],
		"category_id":    category["id"],
		"party_id":       party["id"],
		"opportunity_id": opportunity["id"],
		"kase_id":        kase["id"],
		"created_at":     item["createdAt"],
		"updated_at":     item["updatedAt"],
	}
}

func capsuleUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"username":   item["username"],
		"name":       item["name"],
		"email":      item["email"],
		"status":     item["status"],
		"created_at": item["createdAt"],
		"updated_at": item["updatedAt"],
	}
}

// nestedObject safely returns a nested JSON object, or an empty map when the
// field is absent or not an object, so mappers can index it without panicking.
func nestedObject(v any) map[string]any {
	if obj, ok := v.(map[string]any); ok {
		return obj
	}
	return map[string]any{}
}
