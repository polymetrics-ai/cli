package tempo

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Tempo API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Tempo v4 list endpoint path segment (e.g. "worklogs").
	resource string
	// mapRecord flattens a raw Tempo object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// tempoStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in tempoStreams; the read path
// is fully data-driven from this table. The core set mirrors the upstream
// Airbyte source-tempo streams.
var tempoStreamEndpoints = map[string]streamEndpoint{
	"accounts":         {resource: "accounts", mapRecord: tempoAccountRecord},
	"customers":        {resource: "customers", mapRecord: tempoCustomerRecord},
	"worklogs":         {resource: "worklogs", mapRecord: tempoWorklogRecord},
	"workload-schemes": {resource: "workload-schemes", mapRecord: tempoWorkloadSchemeRecord},
}

// tempoStreams returns the connector's published stream catalog.
func tempoStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "accounts",
			Description: "Tempo accounts.",
			PrimaryKey:  []string{"id"},
			Fields:      tempoAccountFields(),
		},
		{
			Name:        "customers",
			Description: "Tempo customers.",
			PrimaryKey:  []string{"id"},
			Fields:      tempoCustomerFields(),
		},
		{
			Name:         "worklogs",
			Description:  "Tempo worklogs.",
			PrimaryKey:   []string{"tempo_worklog_id"},
			CursorFields: []string{"updated_at"},
			Fields:       tempoWorklogFields(),
		},
		{
			Name:        "workload-schemes",
			Description: "Tempo workload schemes.",
			PrimaryKey:  []string{"id"},
			Fields:      tempoWorkloadSchemeFields(),
		},
	}
}

func tempoAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "key", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "global", Type: "boolean"},
		{Name: "monthly_budget", Type: "number"},
	}
}

func tempoCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "key", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

func tempoWorklogFields() []connectors.Field {
	return []connectors.Field{
		{Name: "tempo_worklog_id", Type: "integer"},
		{Name: "jira_worklog_id", Type: "integer"},
		{Name: "issue_id", Type: "integer"},
		{Name: "time_spent_seconds", Type: "integer"},
		{Name: "billable_seconds", Type: "integer"},
		{Name: "start_date", Type: "string"},
		{Name: "start_time", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func tempoWorkloadSchemeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "default_scheme", Type: "boolean"},
	}
}

func tempoAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"key":            item["key"],
		"name":           item["name"],
		"status":         item["status"],
		"global":         item["global"],
		"monthly_budget": item["monthlyBudget"],
	}
}

func tempoCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"key":  item["key"],
		"name": item["name"],
	}
}

func tempoWorklogRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"tempo_worklog_id":   item["tempoWorklogId"],
		"jira_worklog_id":    item["jiraWorklogId"],
		"issue_id":           nestedInt(item, "issue", "id"),
		"time_spent_seconds": item["timeSpentSeconds"],
		"billable_seconds":   item["billableSeconds"],
		"start_date":         item["startDate"],
		"start_time":         item["startTime"],
		"description":        item["description"],
		"created_at":         item["createdAt"],
		"updated_at":         item["updatedAt"],
	}
}

func tempoWorkloadSchemeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"description":    item["description"],
		"default_scheme": item["defaultScheme"],
	}
}

// nestedInt safely extracts a nested object field (e.g. worklog.issue.id),
// returning nil when the parent object or field is absent.
func nestedInt(item map[string]any, parent, key string) any {
	obj, ok := item[parent].(map[string]any)
	if !ok {
		return nil
	}
	return obj[key]
}
