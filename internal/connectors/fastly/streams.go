package fastly

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Fastly API resource path (relative to
// base_url) it reads from, the record mapper that flattens its objects, and
// whether the endpoint returns a paginated collection (top-level array with
// page/per_page) or a single object.
type streamEndpoint struct {
	// resource is the Fastly endpoint path segment (e.g. "service").
	resource string
	// paginated is true for list endpoints that return a top-level JSON array and
	// accept page/per_page; false for single-object endpoints like /current_user.
	paginated bool
	// mapRecord flattens a raw Fastly object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// fastlyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in fastlyStreams; the read path
// is fully data-driven from this table.
var fastlyStreamEndpoints = map[string]streamEndpoint{
	"services":         {resource: "service", paginated: true, mapRecord: fastlyServiceRecord},
	"current_user":     {resource: "current_user", paginated: false, mapRecord: fastlyUserRecord},
	"current_customer": {resource: "current_customer", paginated: false, mapRecord: fastlyCustomerRecord},
	"datacenters":      {resource: "datacenters", paginated: true, mapRecord: fastlyDatacenterRecord},
}

// fastlyStreams returns the connector's published stream catalog.
func fastlyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "services",
			Description:  "Fastly services for the account.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       fastlyServiceFields(),
		},
		{
			Name:        "current_user",
			Description: "The authenticated Fastly user for the supplied API token.",
			PrimaryKey:  []string{"id"},
			Fields:      fastlyUserFields(),
		},
		{
			Name:        "current_customer",
			Description: "The Fastly customer (account) the token belongs to.",
			PrimaryKey:  []string{"id"},
			Fields:      fastlyCustomerFields(),
		},
		{
			Name:        "datacenters",
			Description: "Fastly points of presence (datacenters).",
			PrimaryKey:  []string{"code"},
			Fields:      fastlyDatacenterFields(),
		},
	}
}

func fastlyServiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "comment", Type: "string"},
		{Name: "version", Type: "integer"},
		{Name: "paused", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "deleted_at", Type: "timestamp"},
	}
}

func fastlyUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "login", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email_hash", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "locked", Type: "boolean"},
		{Name: "two_factor_auth_enabled", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func fastlyCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "owner_id", Type: "string"},
		{Name: "billing_contact_id", Type: "string"},
		{Name: "can_stream_syslog", Type: "boolean"},
		{Name: "has_account_panel", Type: "boolean"},
		{Name: "pricing_plan", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func fastlyDatacenterFields() []connectors.Field {
	return []connectors.Field{
		{Name: "code", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "group", Type: "string"},
		{Name: "shield", Type: "string"},
		{Name: "coordinates", Type: "object"},
	}
}

func fastlyServiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"customer_id": item["customer_id"],
		"type":        item["type"],
		"comment":     item["comment"],
		"version":     item["version"],
		"paused":      item["paused"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
		"deleted_at":  item["deleted_at"],
	}
}

func fastlyUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                      item["id"],
		"login":                   item["login"],
		"name":                    item["name"],
		"email_hash":              item["email_hash"],
		"role":                    item["role"],
		"customer_id":             item["customer_id"],
		"locked":                  item["locked"],
		"two_factor_auth_enabled": item["two_factor_auth_enabled"],
		"created_at":              item["created_at"],
		"updated_at":              item["updated_at"],
	}
}

func fastlyCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"owner_id":           item["owner_id"],
		"billing_contact_id": item["billing_contact_id"],
		"can_stream_syslog":  item["can_stream_syslog"],
		"has_account_panel":  item["has_account_panel"],
		"pricing_plan":       item["pricing_plan"],
		"created_at":         item["created_at"],
		"updated_at":         item["updated_at"],
	}
}

func fastlyDatacenterRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"code":        item["code"],
		"name":        item["name"],
		"group":       item["group"],
		"shield":      item["shield"],
		"coordinates": item["coordinates"],
	}
}
