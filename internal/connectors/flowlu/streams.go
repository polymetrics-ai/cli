package flowlu

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Flowlu API resource path (relative to
// base_url, which already ends in /module) and the record mapper that projects
// its objects into a connectors.Record.
type streamEndpoint struct {
	// resource is the Flowlu list endpoint path (e.g. "crm/account/list").
	resource string
	// mapRecord projects a raw Flowlu item into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// flowluStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in flowluStreams; the read path
// is fully data-driven from this table.
//
// Flowlu groups endpoints by module: the base URL ends in /module, so each
// resource is "<module>/<entity>/list".
var flowluStreamEndpoints = map[string]streamEndpoint{
	"accounts":     {resource: "crm/account/list", mapRecord: flowluAccountRecord},
	"leads":        {resource: "crm/lead/list", mapRecord: flowluLeadRecord},
	"tasks":        {resource: "task/tasks/list", mapRecord: flowluTaskRecord},
	"projects":     {resource: "st/projects/list", mapRecord: flowluProjectRecord},
	"invoices":     {resource: "fin/invoice/list", mapRecord: flowluInvoiceRecord},
	"agile_issues": {resource: "agile/issues/list", mapRecord: flowluAgileIssueRecord},
}

// flowluStreams returns the connector's published stream catalog. Flowlu list
// items expose a numeric id and (for most entities) created/updated timestamps,
// so the primary key is ["id"] across the board.
func flowluStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "accounts",
			Description:  "Flowlu CRM accounts (companies and contacts).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_date"},
			Fields:       flowluAccountFields(),
		},
		{
			Name:         "leads",
			Description:  "Flowlu CRM leads.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_date"},
			Fields:       flowluLeadFields(),
		},
		{
			Name:         "tasks",
			Description:  "Flowlu tasks.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_date"},
			Fields:       flowluTaskFields(),
		},
		{
			Name:         "projects",
			Description:  "Flowlu projects.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_date"},
			Fields:       flowluProjectFields(),
		},
		{
			Name:         "invoices",
			Description:  "Flowlu finance invoices.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_date"},
			Fields:       flowluInvoiceFields(),
		},
		{
			Name:         "agile_issues",
			Description:  "Flowlu agile issues.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_date"},
			Fields:       flowluAgileIssueFields(),
		},
	}
}

func flowluAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "active", Type: "integer"},
		{Name: "owner_id", Type: "integer"},
		{Name: "created_date", Type: "string"},
		{Name: "updated_date", Type: "string"},
	}
}

func flowluLeadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "stage_id", Type: "integer"},
		{Name: "pipeline_id", Type: "integer"},
		{Name: "budget", Type: "string"},
		{Name: "owner_id", Type: "integer"},
		{Name: "active", Type: "integer"},
		{Name: "created_date", Type: "string"},
		{Name: "updated_date", Type: "string"},
	}
}

func flowluTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "priority", Type: "integer"},
		{Name: "workflow_stage_id", Type: "integer"},
		{Name: "responsible_id", Type: "integer"},
		{Name: "owner_id", Type: "integer"},
		{Name: "deadline", Type: "string"},
		{Name: "created_date", Type: "string"},
		{Name: "updated_date", Type: "string"},
	}
}

func flowluProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "stage_id", Type: "integer"},
		{Name: "manager_id", Type: "integer"},
		{Name: "owner_id", Type: "integer"},
		{Name: "active", Type: "integer"},
		{Name: "created_date", Type: "string"},
		{Name: "updated_date", Type: "string"},
	}
}

func flowluInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "invoice_number", Type: "string"},
		{Name: "customer_id", Type: "integer"},
		{Name: "total_amount", Type: "string"},
		{Name: "currency_id", Type: "integer"},
		{Name: "invoice_status", Type: "integer"},
		{Name: "invoice_date", Type: "string"},
		{Name: "created_date", Type: "string"},
		{Name: "updated_date", Type: "string"},
	}
}

func flowluAgileIssueFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "project_id", Type: "integer"},
		{Name: "sprint_id", Type: "integer"},
		{Name: "type", Type: "integer"},
		{Name: "priority", Type: "integer"},
		{Name: "owner_id", Type: "integer"},
		{Name: "created_date", Type: "string"},
		{Name: "updated_date", Type: "string"},
	}
}

func flowluAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"type":         item["type"],
		"name":         item["name"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"email":        item["email"],
		"phone":        item["phone"],
		"active":       item["active"],
		"owner_id":     item["owner_id"],
		"created_date": item["created_date"],
		"updated_date": item["updated_date"],
	}
}

func flowluLeadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"title":        item["title"],
		"stage_id":     item["stage_id"],
		"pipeline_id":  item["pipeline_id"],
		"budget":       item["budget"],
		"owner_id":     item["owner_id"],
		"active":       item["active"],
		"created_date": item["created_date"],
		"updated_date": item["updated_date"],
	}
}

func flowluTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"description":       item["description"],
		"priority":          item["priority"],
		"workflow_stage_id": item["workflow_stage_id"],
		"responsible_id":    item["responsible_id"],
		"owner_id":          item["owner_id"],
		"deadline":          item["deadline"],
		"created_date":      item["created_date"],
		"updated_date":      item["updated_date"],
	}
}

func flowluProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"description":  item["description"],
		"stage_id":     item["stage_id"],
		"manager_id":   item["manager_id"],
		"owner_id":     item["owner_id"],
		"active":       item["active"],
		"created_date": item["created_date"],
		"updated_date": item["updated_date"],
	}
}

func flowluInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"invoice_number": item["invoice_number"],
		"customer_id":    item["customer_id"],
		"total_amount":   item["total_amount"],
		"currency_id":    item["currency_id"],
		"invoice_status": item["invoice_status"],
		"invoice_date":   item["invoice_date"],
		"created_date":   item["created_date"],
		"updated_date":   item["updated_date"],
	}
}

func flowluAgileIssueRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"description":  item["description"],
		"project_id":   item["project_id"],
		"sprint_id":    item["sprint_id"],
		"type":         item["type"],
		"priority":     item["priority"],
		"owner_id":     item["owner_id"],
		"created_date": item["created_date"],
		"updated_date": item["updated_date"],
	}
}
