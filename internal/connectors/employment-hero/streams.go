package employmenthero

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Employment Hero API resource path
// (relative to base_url) it reads from, whether the resource is scoped to an
// organisation, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// orgScoped indicates the path requires an organisation id (a substream of
	// organisations, e.g. organisations/{org}/employees).
	orgScoped bool
	// path returns the resource path. For org-scoped streams orgID is the
	// resolved organisation id; for root streams it is ignored.
	path func(orgID string) string
	// mapRecord flattens a raw Employment Hero object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// employmentHeroStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in employmentHeroStreams;
// the read path is fully data-driven from this table.
var employmentHeroStreamEndpoints = map[string]streamEndpoint{
	"organisations": {
		orgScoped: false,
		path:      func(string) string { return "organisations" },
		mapRecord: organisationRecord,
	},
	"employees": {
		orgScoped: true,
		path:      func(org string) string { return "organisations/" + org + "/employees" },
		mapRecord: employeeRecord,
	},
	"leave_requests": {
		orgScoped: true,
		path:      func(org string) string { return "organisations/" + org + "/leave_requests" },
		mapRecord: leaveRequestRecord,
	},
	"teams": {
		orgScoped: true,
		path:      func(org string) string { return "organisations/" + org + "/teams" },
		mapRecord: teamRecord,
	},
}

// employmentHeroStreams returns the connector's published stream catalog. Every
// Employment Hero object exposes a string id; the API is full-refresh only, so
// no incremental cursor fields are declared.
func employmentHeroStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "organisations",
			Description: "Employment Hero organisations the API token can access. Root stream; supplies organisation ids for the other streams.",
			PrimaryKey:  []string{"id"},
			Fields:      organisationFields(),
		},
		{
			Name:        "employees",
			Description: "Employees within an organisation (organisations/{organization_id}/employees).",
			PrimaryKey:  []string{"id"},
			Fields:      employeeFields(),
		},
		{
			Name:        "leave_requests",
			Description: "Leave requests within an organisation (organisations/{organization_id}/leave_requests).",
			PrimaryKey:  []string{"id"},
			Fields:      leaveRequestFields(),
		},
		{
			Name:        "teams",
			Description: "Teams within an organisation (organisations/{organization_id}/teams).",
			PrimaryKey:  []string{"id"},
			Fields:      teamFields(),
		},
	}
}

func organisationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "logo_url", Type: "string"},
	}
}

func employeeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "middle_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "known_as", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "account_email", Type: "string"},
		{Name: "company_email", Type: "string"},
		{Name: "personal_email", Type: "string"},
		{Name: "company_mobile", Type: "string"},
		{Name: "personal_mobile_number", Type: "string"},
		{Name: "gender", Type: "string"},
		{Name: "date_of_birth", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "location", Type: "string"},
		{Name: "start_date", Type: "string"},
		{Name: "employing_entity", Type: "string"},
		{Name: "primary_manager", Type: "string"},
	}
}

func leaveRequestFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "employee_id", Type: "string"},
		{Name: "leave_category_name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "start_date", Type: "string"},
		{Name: "end_date", Type: "string"},
		{Name: "total_hours", Type: "number"},
		{Name: "leave_balance_amount", Type: "number"},
		{Name: "comment", Type: "string"},
	}
}

func teamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func organisationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"name":     item["name"],
		"country":  item["country"],
		"phone":    item["phone"],
		"logo_url": item["logo_url"],
	}
}

func employeeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                     item["id"],
		"first_name":             item["first_name"],
		"middle_name":            item["middle_name"],
		"last_name":              item["last_name"],
		"known_as":               item["known_as"],
		"title":                  item["title"],
		"job_title":              item["job_title"],
		"role":                   item["role"],
		"account_email":          item["account_email"],
		"company_email":          item["company_email"],
		"personal_email":         item["personal_email"],
		"company_mobile":         item["company_mobile"],
		"personal_mobile_number": item["personal_mobile_number"],
		"gender":                 item["gender"],
		"date_of_birth":          item["date_of_birth"],
		"country":                item["country"],
		"location":               item["location"],
		"start_date":             item["start_date"],
		"employing_entity":       item["employing_entity"],
		"primary_manager":        item["primary_manager"],
	}
}

func leaveRequestRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["id"],
		"employee_id":          item["employee_id"],
		"leave_category_name":  item["leave_category_name"],
		"status":               item["status"],
		"start_date":           item["start_date"],
		"end_date":             item["end_date"],
		"total_hours":          item["total_hours"],
		"leave_balance_amount": item["leave_balance_amount"],
		"comment":              item["comment"],
	}
}

func teamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":     item["id"],
		"name":   item["name"],
		"status": item["status"],
	}
}
