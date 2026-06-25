package freshsales

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Freshsales resource and the JSON key
// under which its records array lives, plus the record mapper that flattens its
// objects. Freshsales list endpoints are addressed as
// "<resource>/view/<view_id>" and return {"<recordsKey>":[...], "meta":{...}}.
type streamEndpoint struct {
	// resource is the Freshsales resource path segment (e.g. "contacts").
	resource string
	// recordsKey is the JSON key holding the array of records (e.g. "contacts").
	recordsKey string
	// mapRecord flattens a raw Freshsales object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// freshsalesStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in freshsalesStreams; the
// read path is fully data-driven from this table.
var freshsalesStreamEndpoints = map[string]streamEndpoint{
	"contacts":       {resource: "contacts", recordsKey: "contacts", mapRecord: freshsalesContactRecord},
	"sales_accounts": {resource: "sales_accounts", recordsKey: "sales_accounts", mapRecord: freshsalesAccountRecord},
	"deals":          {resource: "deals", recordsKey: "deals", mapRecord: freshsalesDealRecord},
	"leads":          {resource: "leads", recordsKey: "leads", mapRecord: freshsalesLeadRecord},
}

// freshsalesStreams returns the connector's published stream catalog. Every
// Freshsales CRM object exposes a numeric id and an updated_at timestamp, so the
// primary key is ["id"] and the incremental cursor field is ["updated_at"]
// across the board.
func freshsalesStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "Freshsales contacts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshsalesContactFields(),
		},
		{
			Name:         "sales_accounts",
			Description:  "Freshsales sales accounts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshsalesAccountFields(),
		},
		{
			Name:         "deals",
			Description:  "Freshsales deals.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshsalesDealFields(),
		},
		{
			Name:         "leads",
			Description:  "Freshsales leads.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       freshsalesLeadFields(),
		},
	}
}

func freshsalesContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "mobile_number", Type: "string"},
		{Name: "work_number", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "owner_id", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshsalesAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "industry_type_id", Type: "integer"},
		{Name: "number_of_employees", Type: "integer"},
		{Name: "annual_revenue", Type: "number"},
		{Name: "city", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "owner_id", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshsalesDealFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "amount", Type: "number"},
		{Name: "currency_id", Type: "integer"},
		{Name: "deal_stage_id", Type: "integer"},
		{Name: "deal_pipeline_id", Type: "integer"},
		{Name: "sales_account_id", Type: "integer"},
		{Name: "owner_id", Type: "integer"},
		{Name: "probability", Type: "integer"},
		{Name: "expected_close", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshsalesLeadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "company_name", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "lead_stage_id", Type: "integer"},
		{Name: "owner_id", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func freshsalesContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"first_name":    item["first_name"],
		"last_name":     item["last_name"],
		"display_name":  item["display_name"],
		"email":         item["email"],
		"mobile_number": item["mobile_number"],
		"work_number":   item["work_number"],
		"job_title":     item["job_title"],
		"city":          item["city"],
		"country":       item["country"],
		"owner_id":      item["owner_id"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func freshsalesAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"name":                item["name"],
		"website":             item["website"],
		"phone":               item["phone"],
		"industry_type_id":    item["industry_type_id"],
		"number_of_employees": item["number_of_employees"],
		"annual_revenue":      item["annual_revenue"],
		"city":                item["city"],
		"country":             item["country"],
		"owner_id":            item["owner_id"],
		"created_at":          item["created_at"],
		"updated_at":          item["updated_at"],
	}
}

func freshsalesDealRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"amount":           item["amount"],
		"currency_id":      item["currency_id"],
		"deal_stage_id":    item["deal_stage_id"],
		"deal_pipeline_id": item["deal_pipeline_id"],
		"sales_account_id": item["sales_account_id"],
		"owner_id":         item["owner_id"],
		"probability":      item["probability"],
		"expected_close":   item["expected_close"],
		"created_at":       item["created_at"],
		"updated_at":       item["updated_at"],
	}
}

func freshsalesLeadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"first_name":    item["first_name"],
		"last_name":     item["last_name"],
		"display_name":  item["display_name"],
		"email":         item["email"],
		"company_name":  item["company_name"],
		"job_title":     item["job_title"],
		"city":          item["city"],
		"country":       item["country"],
		"lead_stage_id": item["lead_stage_id"],
		"owner_id":      item["owner_id"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}
