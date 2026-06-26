package factorial

import (
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// streamEndpoint maps a stream name to the Factorial API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "employees/employees").
	resource string
	// mapRecord flattens a raw Factorial object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// incremental is true when the stream exposes an updated_at cursor.
	incremental bool
}

// factorialStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in factorialStreams; the
// read path is fully data-driven from this table. The resources mirror the
// Factorial v2 API (base https://api.factorialhr.com/api/v2/resources/).
var factorialStreamEndpoints = map[string]streamEndpoint{
	"employees":   {resource: "employees/employees", mapRecord: factorialEmployeeRecord, incremental: true},
	"teams":       {resource: "teams/teams", mapRecord: factorialTeamRecord},
	"leaves":      {resource: "timeoff/leaves", mapRecord: factorialLeaveRecord, incremental: true},
	"leave_types": {resource: "timeoff/leave_types", mapRecord: factorialLeaveTypeRecord},
	"locations":   {resource: "locations/locations", mapRecord: factorialLocationRecord},
}

// factorialStreams returns the connector's published stream catalog. Every
// Factorial object exposes a numeric id; incremental streams additionally carry
// an updated_at cursor.
func factorialStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "employees",
			Description:  "Factorial employees.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       factorialEmployeeFields(),
		},
		{
			Name:        "teams",
			Description: "Factorial teams.",
			PrimaryKey:  []string{"id"},
			Fields:      factorialTeamFields(),
		},
		{
			Name:         "leaves",
			Description:  "Factorial time-off leaves.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       factorialLeaveFields(),
		},
		{
			Name:        "leave_types",
			Description: "Factorial time-off leave types.",
			PrimaryKey:  []string{"id"},
			Fields:      factorialLeaveTypeFields(),
		},
		{
			Name:        "locations",
			Description: "Factorial work locations.",
			PrimaryKey:  []string{"id"},
			Fields:      factorialLocationFields(),
		},
	}
}

func factorialEmployeeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company_id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "full_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "manager_id", Type: "integer"},
		{Name: "legal_entity_id", Type: "integer"},
		{Name: "location_id", Type: "integer"},
		{Name: "team_ids", Type: "array"},
		{Name: "gender", Type: "string"},
		{Name: "birthday_on", Type: "string"},
		{Name: "terminated_on", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func factorialTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "avatar", Type: "string"},
		{Name: "employee_ids", Type: "array"},
		{Name: "lead_ids", Type: "array"},
	}
}

func factorialLeaveFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "employee_id", Type: "integer"},
		{Name: "leave_type_id", Type: "integer"},
		{Name: "start_on", Type: "string"},
		{Name: "finish_on", Type: "string"},
		{Name: "half_day", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "approved", Type: "boolean"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func factorialLeaveTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "identifier", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "approval_required", Type: "boolean"},
		{Name: "active", Type: "boolean"},
	}
}

func factorialLocationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "address_line_1", Type: "string"},
		{Name: "postal_code", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "main", Type: "boolean"},
	}
}

func factorialEmployeeRecord(item map[string]any) connectors.Record {
	full := stringField(item, "full_name")
	if full == "" {
		full = strings.TrimSpace(stringField(item, "first_name") + " " + stringField(item, "last_name"))
	}
	return connectors.Record{
		"id":              item["id"],
		"company_id":      item["company_id"],
		"first_name":      item["first_name"],
		"last_name":       item["last_name"],
		"full_name":       full,
		"email":           item["email"],
		"manager_id":      item["manager_id"],
		"legal_entity_id": item["legal_entity_id"],
		"location_id":     item["location_id"],
		"team_ids":        item["team_ids"],
		"gender":          item["gender"],
		"birthday_on":     item["birthday_on"],
		"terminated_on":   item["terminated_on"],
		"active":          item["active"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
	}
}

func factorialTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"company_id":   item["company_id"],
		"name":         item["name"],
		"description":  item["description"],
		"avatar":       item["avatar"],
		"employee_ids": item["employee_ids"],
		"lead_ids":     item["lead_ids"],
	}
}

func factorialLeaveRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"employee_id":   item["employee_id"],
		"leave_type_id": item["leave_type_id"],
		"start_on":      item["start_on"],
		"finish_on":     item["finish_on"],
		"half_day":      item["half_day"],
		"description":   item["description"],
		"approved":      item["approved"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func factorialLeaveTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"company_id":        item["company_id"],
		"name":              item["name"],
		"identifier":        item["identifier"],
		"color":             item["color"],
		"approval_required": item["approval_required"],
		"active":            item["active"],
	}
}

func factorialLocationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"company_id":     item["company_id"],
		"name":           item["name"],
		"country":        item["country"],
		"state":          item["state"],
		"city":           item["city"],
		"address_line_1": item["address_line_1"],
		"postal_code":    item["postal_code"],
		"timezone":       item["timezone"],
		"main":           item["main"],
	}
}

// fixtureItem builds a deterministic per-stream raw object for fixture mode.
func fixtureItem(stream string, i int) map[string]any {
	item := map[string]any{
		"id":            i,
		"company_id":    1,
		"connector":     "factorial",
		"fixture":       true,
		"first_name":    fmt.Sprintf("First%d", i),
		"last_name":     fmt.Sprintf("Last%d", i),
		"full_name":     fmt.Sprintf("First%d Last%d", i, i),
		"email":         fmt.Sprintf("fixture+%d@example.com", i),
		"name":          fmt.Sprintf("%s_fixture_%d", strings.TrimSuffix(stream, "s"), i),
		"active":        true,
		"created_at":    fmt.Sprintf("2026-01-0%dT00:00:00.000Z", i),
		"updated_at":    fmt.Sprintf("2026-01-0%dT00:00:00.000Z", i),
		"employee_id":   i,
		"leave_type_id": 1,
		"start_on":      "2026-01-01",
		"finish_on":     "2026-01-02",
		"approved":      true,
		"identifier":    "holiday",
		"country":       "ES",
		"city":          "Barcelona",
		"main":          i == 1,
	}
	return item
}
