package sevenshifts

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the 7shifts API resource path (relative
// to base_url) it reads from, the record mapper that flattens its objects, and
// whether the path is scoped to a company (i.e. requires the company_id config).
type streamEndpoint struct {
	// pathFor builds the request path for the stream. companyID is empty for
	// top-level resources (companies) and otherwise the configured company.
	pathFor func(companyID string) string
	// companyScoped marks streams nested under /v2/company/{company_id}/...
	companyScoped bool
	// mapRecord flattens a raw 7shifts object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams(); the read path is fully
// data-driven from this table. Paths and pagination mirror the 7shifts v2 API
// (https://developers.7shifts.com/reference).
var streamEndpoints = map[string]streamEndpoint{
	"companies": {
		pathFor:       func(string) string { return "/v2/companies" },
		companyScoped: false,
		mapRecord:     companyRecord,
	},
	"locations": {
		pathFor:       func(c string) string { return "/v2/company/" + c + "/locations" },
		companyScoped: true,
		mapRecord:     locationRecord,
	},
	"departments": {
		pathFor:       func(c string) string { return "/v2/company/" + c + "/departments" },
		companyScoped: true,
		mapRecord:     departmentRecord,
	},
	"roles": {
		pathFor:       func(c string) string { return "/v2/company/" + c + "/roles" },
		companyScoped: true,
		mapRecord:     roleRecord,
	},
	"users": {
		pathFor:       func(c string) string { return "/v2/company/" + c + "/users" },
		companyScoped: true,
		mapRecord:     userRecord,
	},
	"shifts": {
		pathFor:       func(c string) string { return "/v2/company/" + c + "/shifts" },
		companyScoped: true,
		mapRecord:     shiftRecord,
	},
	"time_punches": {
		pathFor:       func(c string) string { return "/v2/company/" + c + "/time_punches" },
		companyScoped: true,
		mapRecord:     timePunchRecord,
	},
}

// streams returns the connector's published stream catalog. Every 7shifts object
// exposes an integer id and (for the paginated/incremental streams) a `modified`
// RFC3339 timestamp, so the primary key is ["id"] and the incremental cursor
// field is ["modified"].
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "companies",
			Description:  "7shifts companies the access token can see.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       companyFields(),
		},
		{
			Name:         "locations",
			Description:  "Locations within the configured company.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       locationFields(),
		},
		{
			Name:         "departments",
			Description:  "Departments within the configured company.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       departmentFields(),
		},
		{
			Name:         "roles",
			Description:  "Roles within the configured company.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       roleFields(),
		},
		{
			Name:         "users",
			Description:  "Employees/users within the configured company.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       userFields(),
		},
		{
			Name:         "shifts",
			Description:  "Scheduled shifts within the configured company.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       shiftFields(),
		},
		{
			Name:         "time_punches",
			Description:  "Time clock punches within the configured company.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       timePunchFields(),
		},
	}
}

func companyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "country_code", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func locationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "address", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func departmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company_id", Type: "integer"},
		{Name: "location_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func roleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company_id", Type: "integer"},
		{Name: "location_id", Type: "integer"},
		{Name: "department_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company_id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "hire_date", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func shiftFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company_id", Type: "integer"},
		{Name: "location_id", Type: "integer"},
		{Name: "department_id", Type: "integer"},
		{Name: "role_id", Type: "integer"},
		{Name: "user_id", Type: "integer"},
		{Name: "start", Type: "string"},
		{Name: "end", Type: "string"},
		{Name: "open", Type: "boolean"},
		{Name: "deleted", Type: "boolean"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func timePunchFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company_id", Type: "integer"},
		{Name: "location_id", Type: "integer"},
		{Name: "department_id", Type: "integer"},
		{Name: "role_id", Type: "integer"},
		{Name: "user_id", Type: "integer"},
		{Name: "clocked_in", Type: "string"},
		{Name: "clocked_out", Type: "string"},
		{Name: "approved", Type: "boolean"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func companyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"country_code": item["country_code"],
		"currency":     item["currency"],
		"timezone":     item["timezone"],
		"active":       item["active"],
		"created":      item["created"],
		"modified":     item["modified"],
	}
}

func locationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"company_id": item["company_id"],
		"name":       item["name"],
		"address":    item["address"],
		"city":       item["city"],
		"state":      item["state"],
		"country":    item["country"],
		"timezone":   item["timezone"],
		"created":    item["created"],
		"modified":   item["modified"],
	}
}

func departmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"company_id":  item["company_id"],
		"location_id": item["location_id"],
		"name":        item["name"],
		"created":     item["created"],
		"modified":    item["modified"],
	}
}

func roleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"company_id":    item["company_id"],
		"location_id":   item["location_id"],
		"department_id": item["department_id"],
		"name":          item["name"],
		"color":         item["color"],
		"created":       item["created"],
		"modified":      item["modified"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"company_id": item["company_id"],
		"first_name": item["first_name"],
		"last_name":  item["last_name"],
		"email":      item["email"],
		"type":       item["type"],
		"active":     item["active"],
		"hire_date":  item["hire_date"],
		"created":    item["created"],
		"modified":   item["modified"],
	}
}

func shiftRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"company_id":    item["company_id"],
		"location_id":   item["location_id"],
		"department_id": item["department_id"],
		"role_id":       item["role_id"],
		"user_id":       item["user_id"],
		"start":         item["start"],
		"end":           item["end"],
		"open":          item["open"],
		"deleted":       item["deleted"],
		"created":       item["created"],
		"modified":      item["modified"],
	}
}

func timePunchRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"company_id":    item["company_id"],
		"location_id":   item["location_id"],
		"department_id": item["department_id"],
		"role_id":       item["role_id"],
		"user_id":       item["user_id"],
		"clocked_in":    item["clocked_in"],
		"clocked_out":   item["clocked_out"],
		"approved":      item["approved"],
		"created":       item["created"],
		"modified":      item["modified"],
	}
}
