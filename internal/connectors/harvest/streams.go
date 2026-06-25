package harvest

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Harvest API resource path (relative to
// base_url) it reads from, the JSON key holding the records array in the list
// response, and the record mapper that flattens its objects.
//
// In the Harvest v2 API the records-array key in a list response is always the
// resource name (e.g. GET /clients -> {"clients":[...]}); resource and recordKey
// therefore coincide for every core stream, but both are kept explicit so an
// endpoint with a differing path or envelope can be added without surprises.
type streamEndpoint struct {
	resource  string
	recordKey string
	mapRecord func(map[string]any) connectors.Record
}

// harvestStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in harvestStreams; the read path
// is fully data-driven from this table.
var harvestStreamEndpoints = map[string]streamEndpoint{
	"clients":      {resource: "clients", recordKey: "clients", mapRecord: harvestClientRecord},
	"projects":     {resource: "projects", recordKey: "projects", mapRecord: harvestProjectRecord},
	"tasks":        {resource: "tasks", recordKey: "tasks", mapRecord: harvestTaskRecord},
	"users":        {resource: "users", recordKey: "users", mapRecord: harvestUserRecord},
	"time_entries": {resource: "time_entries", recordKey: "time_entries", mapRecord: harvestTimeEntryRecord},
}

// harvestStreams returns the connector's published stream catalog. Every Harvest
// object exposes an integer id and updated_at/created_at ISO timestamps, so the
// primary key is ["id"] and the incremental cursor field is ["updated_at"]
// across the board.
func harvestStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "clients",
			Description:  "Harvest clients.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       harvestClientFields(),
		},
		{
			Name:         "projects",
			Description:  "Harvest projects.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       harvestProjectFields(),
		},
		{
			Name:         "tasks",
			Description:  "Harvest tasks.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       harvestTaskFields(),
		},
		{
			Name:         "users",
			Description:  "Harvest users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       harvestUserFields(),
		},
		{
			Name:         "time_entries",
			Description:  "Harvest time entries.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       harvestTimeEntryFields(),
		},
	}
}

func harvestClientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "address", Type: "string"},
		{Name: "statement_key", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func harvestProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "is_billable", Type: "boolean"},
		{Name: "client_id", Type: "integer"},
		{Name: "client_name", Type: "string"},
		{Name: "budget", Type: "number"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func harvestTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "billable_by_default", Type: "boolean"},
		{Name: "default_hourly_rate", Type: "number"},
		{Name: "is_default", Type: "boolean"},
		{Name: "is_active", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func harvestUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "is_admin", Type: "boolean"},
		{Name: "timezone", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func harvestTimeEntryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "spent_date", Type: "string"},
		{Name: "user_id", Type: "integer"},
		{Name: "client_id", Type: "integer"},
		{Name: "project_id", Type: "integer"},
		{Name: "task_id", Type: "integer"},
		{Name: "hours", Type: "number"},
		{Name: "notes", Type: "string"},
		{Name: "is_billed", Type: "boolean"},
		{Name: "is_running", Type: "boolean"},
		{Name: "billable", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func harvestClientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"is_active":     item["is_active"],
		"address":       item["address"],
		"statement_key": item["statement_key"],
		"currency":      item["currency"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func harvestProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"code":        item["code"],
		"is_active":   item["is_active"],
		"is_billable": item["is_billable"],
		"client_id":   nestedField(item, "client", "id"),
		"client_name": nestedField(item, "client", "name"),
		"budget":      item["budget"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func harvestTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"name":                item["name"],
		"billable_by_default": item["billable_by_default"],
		"default_hourly_rate": item["default_hourly_rate"],
		"is_default":          item["is_default"],
		"is_active":           item["is_active"],
		"created_at":          item["created_at"],
		"updated_at":          item["updated_at"],
	}
}

func harvestUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"first_name": item["first_name"],
		"last_name":  item["last_name"],
		"email":      item["email"],
		"is_active":  item["is_active"],
		"is_admin":   item["is_admin"],
		"timezone":   item["timezone"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func harvestTimeEntryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"spent_date": item["spent_date"],
		"user_id":    nestedField(item, "user", "id"),
		"client_id":  nestedField(item, "client", "id"),
		"project_id": nestedField(item, "project", "id"),
		"task_id":    nestedField(item, "task", "id"),
		"hours":      item["hours"],
		"notes":      item["notes"],
		"is_billed":  item["is_billed"],
		"is_running": item["is_running"],
		"billable":   item["billable"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

// nestedField returns item[outer][inner] when the outer value is a JSON object,
// otherwise nil. Harvest nests related resources (client, project, user, task) as
// {"id":..,"name":..} objects on list records.
func nestedField(item map[string]any, outer, inner string) any {
	if obj, ok := item[outer].(map[string]any); ok {
		return obj[inner]
	}
	return nil
}
