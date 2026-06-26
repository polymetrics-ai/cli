package nutshell

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Nutshell REST resource path (relative
// to base_url), the JSON envelope key that holds the records array, the record
// mapper, and whether the endpoint is paginated. Reference endpoints (e.g.
// users) return a single page, so paginated is false there.
type streamEndpoint struct {
	// resource is the REST list path segment (e.g. "contacts").
	resource string
	// recordsKey is the top-level JSON key holding the records array
	// (e.g. {"contacts":[...]}).
	recordsKey string
	// mapRecord flattens a raw Nutshell object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// paginated reports whether the endpoint honors page[page]/page[limit].
	paginated bool
}

// nutshellStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in nutshellStreams; the read
// path is fully data-driven from this table.
//
// Endpoints, envelope keys, and pagination are taken from the Nutshell REST API
// (https://app.nutshell.com/rest/) as exercised by the upstream Airbyte
// source-nutshell manifest.
var nutshellStreamEndpoints = map[string]streamEndpoint{
	"accounts":   {resource: "accounts", recordsKey: "accounts", mapRecord: nutshellAccountRecord, paginated: true},
	"contacts":   {resource: "contacts", recordsKey: "contacts", mapRecord: nutshellContactRecord, paginated: true},
	"leads":      {resource: "leads", recordsKey: "leads", mapRecord: nutshellLeadRecord, paginated: true},
	"activities": {resource: "activities", recordsKey: "activities", mapRecord: nutshellActivityRecord, paginated: true},
	"users":      {resource: "users", recordsKey: "users", mapRecord: nutshellUserRecord, paginated: false},
}

// nutshellStreams returns the connector's published stream catalog. Every
// Nutshell object exposes a numeric id (primary key) and a modifiedTime
// timestamp used as the incremental cursor where present.
func nutshellStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "accounts",
			Description:  "Nutshell accounts (companies).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modifiedTime"},
			Fields:       nutshellAccountFields(),
		},
		{
			Name:         "contacts",
			Description:  "Nutshell contacts (people).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modifiedTime"},
			Fields:       nutshellContactFields(),
		},
		{
			Name:         "leads",
			Description:  "Nutshell leads (deals/opportunities).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modifiedTime"},
			Fields:       nutshellLeadFields(),
		},
		{
			Name:         "activities",
			Description:  "Nutshell activities (logged interactions).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modifiedTime"},
			Fields:       nutshellActivityFields(),
		},
		{
			Name:        "users",
			Description: "Nutshell users (team members).",
			PrimaryKey:  []string{"id"},
			Fields:      nutshellUserFields(),
		},
	}
}

func nutshellAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "entityType", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "industryId", Type: "integer"},
		{Name: "accountTypeId", Type: "integer"},
		{Name: "url", Type: "string"},
		{Name: "isHotLead", Type: "boolean"},
		{Name: "createdTime", Type: "timestamp"},
		{Name: "modifiedTime", Type: "timestamp"},
	}
}

func nutshellContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "entityType", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "htmlUrl", Type: "string"},
		{Name: "createdTime", Type: "timestamp"},
		{Name: "modifiedTime", Type: "timestamp"},
	}
}

func nutshellLeadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "entityType", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "integer"},
		{Name: "confidence", Type: "integer"},
		{Name: "value", Type: "string"},
		{Name: "isOverdue", Type: "boolean"},
		{Name: "createdTime", Type: "timestamp"},
		{Name: "modifiedTime", Type: "timestamp"},
		{Name: "closedTime", Type: "timestamp"},
	}
}

func nutshellActivityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "entityType", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "activityTypeId", Type: "integer"},
		{Name: "status", Type: "integer"},
		{Name: "isFlagged", Type: "boolean"},
		{Name: "logNote", Type: "string"},
		{Name: "createdTime", Type: "timestamp"},
		{Name: "modifiedTime", Type: "timestamp"},
	}
}

func nutshellUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "entityType", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "emails", Type: "string"},
		{Name: "isEnabled", Type: "boolean"},
		{Name: "isAdministrator", Type: "boolean"},
		{Name: "createdTime", Type: "timestamp"},
		{Name: "modifiedTime", Type: "timestamp"},
	}
}

func nutshellAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"entityType":    item["entityType"],
		"name":          item["name"],
		"industryId":    item["industryId"],
		"accountTypeId": item["accountTypeId"],
		"url":           item["url"],
		"isHotLead":     item["isHotLead"],
		"createdTime":   item["createdTime"],
		"modifiedTime":  item["modifiedTime"],
	}
}

func nutshellContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"entityType":   item["entityType"],
		"name":         item["name"],
		"description":  item["description"],
		"htmlUrl":      item["htmlUrl"],
		"createdTime":  item["createdTime"],
		"modifiedTime": item["modifiedTime"],
	}
}

func nutshellLeadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"entityType":   item["entityType"],
		"name":         item["name"],
		"status":       item["status"],
		"confidence":   item["confidence"],
		"value":        item["value"],
		"isOverdue":    item["isOverdue"],
		"createdTime":  item["createdTime"],
		"modifiedTime": item["modifiedTime"],
		"closedTime":   item["closedTime"],
	}
}

func nutshellActivityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"entityType":     item["entityType"],
		"name":           item["name"],
		"description":    item["description"],
		"activityTypeId": item["activityTypeId"],
		"status":         item["status"],
		"isFlagged":      item["isFlagged"],
		"logNote":        item["logNote"],
		"createdTime":    item["createdTime"],
		"modifiedTime":   item["modifiedTime"],
	}
}

func nutshellUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"entityType":      item["entityType"],
		"name":            item["name"],
		"emails":          item["emails"],
		"isEnabled":       item["isEnabled"],
		"isAdministrator": item["isAdministrator"],
		"createdTime":     item["createdTime"],
		"modifiedTime":    item["modifiedTime"],
	}
}
