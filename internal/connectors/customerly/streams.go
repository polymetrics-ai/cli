package customerly

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Customerly list endpoint path
// (relative to base_url), the dotted JSON path to its records array, and the
// record mapper that flattens its objects. Adding a stream is one entry here
// plus a Stream definition in customerlyStreams; the read path is data-driven
// from this table.
type streamEndpoint struct {
	// resource is the list endpoint path segment (e.g. "users/list").
	resource string
	// recordsPath is the dotted path to the records array (e.g. "data.users").
	recordsPath string
	// mapRecord flattens a raw Customerly object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// customerlyStreamEndpoints is the per-stream routing table. The two streams
// here are confirmed against the Customerly v1 API (users/list and leads/list),
// both returning {data:{<stream>:[...]}} and supporting page-increment paging.
var customerlyStreamEndpoints = map[string]streamEndpoint{
	"users": {resource: "users/list", recordsPath: "data.users", mapRecord: customerlyUserRecord},
	"leads": {resource: "leads/list", recordsPath: "data.leads", mapRecord: customerlyLeadRecord},
}

// customerlyStreams returns the connector's published stream catalog. Both
// streams expose a string/numeric id and a `last_update` timestamp used as the
// incremental cursor.
func customerlyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "users",
			Description:  "Customerly users (identified contacts).",
			PrimaryKey:   []string{"user_id", "email"},
			CursorFields: []string{"last_update"},
			Fields:       customerlyUserFields(),
		},
		{
			Name:         "leads",
			Description:  "Customerly leads (anonymous or pre-identified contacts).",
			PrimaryKey:   []string{"crmhero_user_id"},
			CursorFields: []string{"last_update"},
			Fields:       customerlyLeadFields(),
		},
	}
}

func customerlyUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "user_id", Type: "integer"},
		{Name: "crmhero_user_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "sub_active", Type: "boolean"},
		{Name: "sub_status", Type: "string"},
		{Name: "create_date", Type: "string"},
		{Name: "last_update", Type: "string"},
		{Name: "first_seen_at", Type: "string"},
		{Name: "last_activity", Type: "string"},
	}
}

func customerlyLeadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "crmhero_user_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "sub_active", Type: "boolean"},
		{Name: "sub_status", Type: "string"},
		{Name: "create_date", Type: "string"},
		{Name: "last_update", Type: "string"},
	}
}

func customerlyUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"user_id":         item["user_id"],
		"crmhero_user_id": item["crmhero_user_id"],
		"email":           item["email"],
		"name":            item["name"],
		"username":        item["username"],
		"role":            item["role"],
		"country":         item["country"],
		"city":            item["city"],
		"timezone":        item["timezone"],
		"sub_active":      item["sub_active"],
		"sub_status":      item["sub_status"],
		"create_date":     item["create_date"],
		"last_update":     item["last_update"],
		"first_seen_at":   item["first_seen_at"],
		"last_activity":   item["last_activity"],
	}
}

func customerlyLeadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"crmhero_user_id": item["crmhero_user_id"],
		"email":           item["email"],
		"name":            item["name"],
		"username":        item["username"],
		"role":            item["role"],
		"country":         item["country"],
		"city":            item["city"],
		"timezone":        item["timezone"],
		"sub_active":      item["sub_active"],
		"sub_status":      item["sub_status"],
		"create_date":     item["create_date"],
		"last_update":     item["last_update"],
	}
}
