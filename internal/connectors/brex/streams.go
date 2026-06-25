package brex

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Brex API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// path is the Brex list endpoint path (e.g. "/v2/users").
	path string
	// mapRecord flattens a raw Brex object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// brexStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in brexStreams; the read path
// is fully data-driven from this table. Paths mirror the Brex platform API
// (https://platform.brexapis.com) as exposed by the Airbyte source-brex
// manifest.
var brexStreamEndpoints = map[string]streamEndpoint{
	"transactions": {path: "/v2/transactions/card/primary", mapRecord: brexTransactionRecord},
	"users":        {path: "/v2/users", mapRecord: brexUserRecord},
	"expenses":     {path: "/v1/expenses/card", mapRecord: brexExpenseRecord},
	"vendors":      {path: "/v1/vendors", mapRecord: brexVendorRecord},
	"budgets":      {path: "/v2/budgets", mapRecord: brexBudgetRecord},
}

// brexStreams returns the connector's published stream catalog. Brex list
// responses wrap records in {"items":[...],"next_cursor":...}; pagination is
// cursor-based via the `cursor` query parameter.
func brexStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "transactions",
			Description:  "Brex primary card transactions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"posted_at_date"},
			Fields:       brexTransactionFields(),
		},
		{
			Name:        "users",
			Description: "Brex users.",
			PrimaryKey:  []string{"id"},
			Fields:      brexUserFields(),
		},
		{
			Name:         "expenses",
			Description:  "Brex card expenses.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"purchased_at"},
			Fields:       brexExpenseFields(),
		},
		{
			Name:        "vendors",
			Description: "Brex vendors.",
			PrimaryKey:  []string{"id"},
			Fields:      brexVendorFields(),
		},
		{
			Name:        "budgets",
			Description: "Brex budgets.",
			PrimaryKey:  []string{"budget_id"},
			Fields:      brexBudgetFields(),
		},
	}
}

func brexTransactionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "card_id", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "amount", Type: "object"},
		{Name: "initiated_at_date", Type: "string"},
		{Name: "posted_at_date", Type: "string"},
		{Name: "type", Type: "string"},
	}
}

func brexUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "manager_id", Type: "string"},
		{Name: "department_id", Type: "string"},
	}
}

func brexExpenseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "memo", Type: "string"},
		{Name: "location_id", Type: "string"},
		{Name: "department_id", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "merchant_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "purchased_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "original_amount", Type: "object"},
	}
}

func brexVendorFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "company_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "payment_accounts", Type: "array"},
	}
}

func brexBudgetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "budget_id", Type: "string"},
		{Name: "parent_budget_id", Type: "string"},
		{Name: "account_id", Type: "string"},
		{Name: "creator_user_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "limit", Type: "object"},
		{Name: "period_type", Type: "string"},
	}
}

func brexTransactionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"card_id":           item["card_id"],
		"description":       item["description"],
		"amount":            item["amount"],
		"initiated_at_date": item["initiated_at_date"],
		"posted_at_date":    item["posted_at_date"],
		"type":              item["type"],
	}
}

func brexUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"first_name":    item["first_name"],
		"last_name":     item["last_name"],
		"email":         item["email"],
		"status":        item["status"],
		"manager_id":    item["manager_id"],
		"department_id": item["department_id"],
	}
}

func brexExpenseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"memo":            item["memo"],
		"location_id":     item["location_id"],
		"department_id":   item["department_id"],
		"user_id":         item["user_id"],
		"merchant_id":     item["merchant_id"],
		"status":          item["status"],
		"category":        item["category"],
		"purchased_at":    item["purchased_at"],
		"updated_at":      item["updated_at"],
		"original_amount": item["original_amount"],
	}
}

func brexVendorRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"company_name":     item["company_name"],
		"email":            item["email"],
		"phone":            item["phone"],
		"payment_accounts": item["payment_accounts"],
	}
}

func brexBudgetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"budget_id":        item["budget_id"],
		"parent_budget_id": item["parent_budget_id"],
		"account_id":       item["account_id"],
		"creator_user_id":  item["creator_user_id"],
		"name":             item["name"],
		"description":      item["description"],
		"status":           item["status"],
		"limit":            item["limit"],
		"period_type":      item["period_type"],
	}
}
