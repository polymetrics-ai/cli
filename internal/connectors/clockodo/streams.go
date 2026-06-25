package clockodo

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Clockodo API resource path (relative
// to base_url) it reads from, the JSON wrapper key that holds the records array,
// and the mapper that flattens each object into a connectors.Record.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "v2/customers", "absences").
	resource string
	// recordsKey is the top-level JSON key wrapping the array (e.g. "customers").
	recordsKey string
	// paginated reports whether the endpoint supports the `page` parameter and a
	// `paging` object. Non-paginated endpoints are read in a single request.
	paginated bool
	// mapRecord flattens a raw Clockodo object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// clockodoStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in clockodoStreams; the read
// path is fully data-driven from this table.
var clockodoStreamEndpoints = map[string]streamEndpoint{
	"customers": {resource: "v2/customers", recordsKey: "customers", paginated: true, mapRecord: clockodoCustomerRecord},
	"projects":  {resource: "v2/projects", recordsKey: "projects", paginated: true, mapRecord: clockodoProjectRecord},
	"services":  {resource: "v2/services", recordsKey: "services", paginated: false, mapRecord: clockodoServiceRecord},
	"users":     {resource: "v2/users", recordsKey: "users", paginated: false, mapRecord: clockodoUserRecord},
}

// clockodoStreams returns the connector's published stream catalog. Clockodo
// administrative resources are keyed by an integer `id` and have no natural
// incremental cursor, so all are full-refresh with PrimaryKey ["id"].
func clockodoStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "customers",
			Description: "Clockodo customers (clients).",
			PrimaryKey:  []string{"id"},
			Fields:      clockodoCustomerFields(),
		},
		{
			Name:        "projects",
			Description: "Clockodo projects belonging to customers.",
			PrimaryKey:  []string{"id"},
			Fields:      clockodoProjectFields(),
		},
		{
			Name:        "services",
			Description: "Clockodo services (activity types).",
			PrimaryKey:  []string{"id"},
			Fields:      clockodoServiceFields(),
		},
		{
			Name:        "users",
			Description: "Clockodo co-workers (users).",
			PrimaryKey:  []string{"id"},
			Fields:      clockodoUserFields(),
		},
	}
}

func clockodoCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "billable_default", Type: "boolean"},
		{Name: "note", Type: "string"},
		{Name: "color", Type: "integer"},
	}
}

func clockodoProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "customers_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "billable_default", Type: "boolean"},
		{Name: "note", Type: "string"},
		{Name: "budget_money", Type: "number"},
		{Name: "budget_is_hours", Type: "boolean"},
		{Name: "completed", Type: "boolean"},
		{Name: "deadline", Type: "string"},
	}
}

func clockodoServiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "note", Type: "string"},
	}
}

func clockodoUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "number", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "teams_id", Type: "integer"},
	}
}

func clockodoCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"number":           item["number"],
		"active":           item["active"],
		"billable_default": item["billable_default"],
		"note":             item["note"],
		"color":            item["color"],
	}
}

func clockodoProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"customers_id":     item["customers_id"],
		"name":             item["name"],
		"number":           item["number"],
		"active":           item["active"],
		"billable_default": item["billable_default"],
		"note":             item["note"],
		"budget_money":     item["budget_money"],
		"budget_is_hours":  item["budget_is_hours"],
		"completed":        item["completed"],
		"deadline":         item["deadline"],
	}
}

func clockodoServiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":     item["id"],
		"name":   item["name"],
		"number": item["number"],
		"active": item["active"],
		"note":   item["note"],
	}
}

func clockodoUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"name":     item["name"],
		"email":    item["email"],
		"role":     item["role"],
		"active":   item["active"],
		"number":   item["number"],
		"language": item["language"],
		"timezone": item["timezone"],
		"teams_id": item["teams_id"],
	}
}
