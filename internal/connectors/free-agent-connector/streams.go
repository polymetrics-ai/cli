package freeagentconnector

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the FreeAgent API resource path (relative
// to base_url), the top-level JSON key wrapping its array, and the record mapper
// that flattens its objects.
type streamEndpoint struct {
	// resource is the FreeAgent list endpoint path segment (e.g. "contacts").
	resource string
	// recordsKey is the top-level JSON key holding the array (e.g. "contacts").
	recordsKey string
	// mapRecord flattens a raw FreeAgent object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streamDefs; the read path is fully
// data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"contacts": {resource: "contacts", recordsKey: "contacts", mapRecord: contactRecord},
	"invoices": {resource: "invoices", recordsKey: "invoices", mapRecord: invoiceRecord},
	"bills":    {resource: "bills", recordsKey: "bills", mapRecord: billRecord},
	"projects": {resource: "projects", recordsKey: "projects", mapRecord: projectRecord},
	"tasks":    {resource: "tasks", recordsKey: "tasks", mapRecord: taskRecord},
}

// streamDefs returns the connector's published stream catalog. Every FreeAgent
// resource exposes a string `url` identifier and `updated_at`/`created_at`
// timestamps, so the primary key is ["url"] and the incremental cursor field is
// ["updated_at"] across the board.
func streamDefs() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "FreeAgent contacts (customers and suppliers).",
			PrimaryKey:   []string{"url"},
			CursorFields: []string{"updated_at"},
			Fields:       contactFields(),
		},
		{
			Name:         "invoices",
			Description:  "FreeAgent sales invoices.",
			PrimaryKey:   []string{"url"},
			CursorFields: []string{"updated_at"},
			Fields:       invoiceFields(),
		},
		{
			Name:         "bills",
			Description:  "FreeAgent purchase bills.",
			PrimaryKey:   []string{"url"},
			CursorFields: []string{"updated_at"},
			Fields:       billFields(),
		},
		{
			Name:         "projects",
			Description:  "FreeAgent projects.",
			PrimaryKey:   []string{"url"},
			CursorFields: []string{"updated_at"},
			Fields:       projectFields(),
		},
		{
			Name:         "tasks",
			Description:  "FreeAgent project tasks.",
			PrimaryKey:   []string{"url"},
			CursorFields: []string{"updated_at"},
			Fields:       taskFields(),
		},
	}
}

func contactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "url", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "organisation_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone_number", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "account_balance", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func invoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "url", Type: "string"},
		{Name: "reference", Type: "string"},
		{Name: "contact", Type: "string"},
		{Name: "dated_on", Type: "string"},
		{Name: "due_on", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "net_value", Type: "string"},
		{Name: "total_value", Type: "string"},
		{Name: "due_value", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func billFields() []connectors.Field {
	return []connectors.Field{
		{Name: "url", Type: "string"},
		{Name: "reference", Type: "string"},
		{Name: "contact", Type: "string"},
		{Name: "dated_on", Type: "string"},
		{Name: "due_on", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "total_value", Type: "string"},
		{Name: "due_value", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func projectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "url", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "contact", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "budget", Type: "string"},
		{Name: "budget_units", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func taskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "url", Type: "string"},
		{Name: "project", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "is_billable", Type: "boolean"},
		{Name: "billing_rate", Type: "string"},
		{Name: "billing_period", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func contactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"url":               item["url"],
		"first_name":        item["first_name"],
		"last_name":         item["last_name"],
		"organisation_name": item["organisation_name"],
		"email":             item["email"],
		"phone_number":      item["phone_number"],
		"status":            item["status"],
		"account_balance":   item["account_balance"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
	}
}

func invoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"url":         item["url"],
		"reference":   item["reference"],
		"contact":     item["contact"],
		"dated_on":    item["dated_on"],
		"due_on":      item["due_on"],
		"status":      item["status"],
		"currency":    item["currency"],
		"net_value":   item["net_value"],
		"total_value": item["total_value"],
		"due_value":   item["due_value"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func billRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"url":         item["url"],
		"reference":   item["reference"],
		"contact":     item["contact"],
		"dated_on":    item["dated_on"],
		"due_on":      item["due_on"],
		"status":      item["status"],
		"currency":    item["currency"],
		"total_value": item["total_value"],
		"due_value":   item["due_value"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func projectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"url":          item["url"],
		"name":         item["name"],
		"contact":      item["contact"],
		"status":       item["status"],
		"currency":     item["currency"],
		"budget":       item["budget"],
		"budget_units": item["budget_units"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func taskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"url":            item["url"],
		"project":        item["project"],
		"name":           item["name"],
		"status":         item["status"],
		"is_billable":    item["is_billable"],
		"billing_rate":   item["billing_rate"],
		"billing_period": item["billing_period"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}
