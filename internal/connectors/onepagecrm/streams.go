package onepagecrm

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to its OnePageCRM API resource path (relative
// to base_url), the JSON path to the records array in the response, the key that
// each array element is wrapped under (OnePageCRM nests every object under a
// singular key, e.g. {"contact": {...}}), and the record mapper.
type streamEndpoint struct {
	// resource is the list endpoint path segment (e.g. "contacts").
	resource string
	// arrayPath is the dotted path under the response root to the records array
	// (e.g. "data.contacts" or "data" for top-level arrays).
	arrayPath string
	// wrapKey is the singular key each element is wrapped under (e.g. "contact").
	// Empty means the element is the record itself.
	wrapKey string
	// mapRecord flattens an unwrapped OnePageCRM object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// onepagecrmStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in onepagecrmStreams; the
// read path is fully data-driven from this table.
var onepagecrmStreamEndpoints = map[string]streamEndpoint{
	"contacts":  {resource: "contacts", arrayPath: "data.contacts", wrapKey: "contact", mapRecord: contactRecord},
	"deals":     {resource: "deals", arrayPath: "data.deals", wrapKey: "deal", mapRecord: dealRecord},
	"actions":   {resource: "actions", arrayPath: "data.actions", wrapKey: "action", mapRecord: actionRecord},
	"companies": {resource: "companies", arrayPath: "data.companies", wrapKey: "company", mapRecord: companyRecord},
	"users":     {resource: "users", arrayPath: "data", wrapKey: "user", mapRecord: userRecord},
}

// onepagecrmStreams returns the connector's published stream catalog. Every
// OnePageCRM object exposes a string id; objects that track modifications expose
// a unix `updated_at`, used as the incremental cursor where available.
func onepagecrmStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "OnePageCRM contacts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       contactFields(),
		},
		{
			Name:         "deals",
			Description:  "OnePageCRM deals (sales opportunities).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       dealFields(),
		},
		{
			Name:         "actions",
			Description:  "OnePageCRM actions (next actions / tasks).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       actionFields(),
		},
		{
			Name:         "companies",
			Description:  "OnePageCRM companies.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       companyFields(),
		},
		{
			Name:        "users",
			Description: "OnePageCRM users in the account.",
			PrimaryKey:  []string{"id"},
			Fields:      userFields(),
		},
	}
}

func contactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "company_name", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "status_id", Type: "string"},
		{Name: "owner_id", Type: "string"},
		{Name: "starred", Type: "boolean"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func dealFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "stage", Type: "string"},
		{Name: "amount", Type: "number"},
		{Name: "currency", Type: "string"},
		{Name: "contact_id", Type: "string"},
		{Name: "owner_id", Type: "string"},
		{Name: "expected_close_date", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func actionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "contact_id", Type: "string"},
		{Name: "text", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "done", Type: "boolean"},
		{Name: "assignee_id", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func companyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func contactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"company_name": item["company_name"],
		"job_title":    item["job_title"],
		"status_id":    item["status_id"],
		"owner_id":     item["owner_id"],
		"starred":      item["starred"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func dealRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"name":                item["name"],
		"status":              item["status"],
		"stage":               item["stage"],
		"amount":              item["amount"],
		"currency":            item["currency"],
		"contact_id":          item["contact_id"],
		"owner_id":            item["owner_id"],
		"expected_close_date": item["expected_close_date"],
		"created_at":          item["created_at"],
		"updated_at":          item["updated_at"],
	}
}

func actionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"contact_id":  item["contact_id"],
		"text":        item["text"],
		"status":      item["status"],
		"date":        item["date"],
		"done":        item["done"],
		"assignee_id": item["assignee_id"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func companyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"phone":       item["phone"],
		"url":         item["url"],
		"description": item["description"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"first_name": item["first_name"],
		"last_name":  item["last_name"],
		"email":      item["email"],
		"role":       item["role"],
		"status":     item["status"],
	}
}
