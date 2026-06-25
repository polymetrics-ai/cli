package chargebee

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Chargebee API resource path (relative
// to base_url) it reads from, the JSON key each list item wraps its object in
// (Chargebee returns {"list":[{"customer":{...}}, ...]}), and the record mapper
// that flattens those objects.
type streamEndpoint struct {
	// resource is the Chargebee list endpoint path segment (e.g. "customers").
	resource string
	// envelope is the per-item wrapper key (e.g. "customer"). Each element of the
	// top-level "list" array is an object with this single key.
	envelope string
	// mapRecord flattens a raw Chargebee resource object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// chargebeeStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in chargebeeStreams; the read
// path is fully data-driven from this table.
var chargebeeStreamEndpoints = map[string]streamEndpoint{
	"customers":     {resource: "customers", envelope: "customer", mapRecord: chargebeeCustomerRecord},
	"subscriptions": {resource: "subscriptions", envelope: "subscription", mapRecord: chargebeeSubscriptionRecord},
	"invoices":      {resource: "invoices", envelope: "invoice", mapRecord: chargebeeInvoiceRecord},
	"plans":         {resource: "plans", envelope: "plan", mapRecord: chargebeePlanRecord},
	"items":         {resource: "items", envelope: "item", mapRecord: chargebeeItemRecord},
}

// chargebeeStreams returns the connector's published stream catalog. Every
// Chargebee object exposes a string id and a unix `updated_at` timestamp, so the
// primary key is ["id"] and the incremental cursor field is ["updated_at"]
// across the board.
func chargebeeStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "customers",
			Description:  "Chargebee customers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chargebeeCustomerFields(),
		},
		{
			Name:         "subscriptions",
			Description:  "Chargebee subscriptions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chargebeeSubscriptionFields(),
		},
		{
			Name:         "invoices",
			Description:  "Chargebee invoices.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chargebeeInvoiceFields(),
		},
		{
			Name:         "plans",
			Description:  "Chargebee plans (product catalog 1.0).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chargebeePlanFields(),
		},
		{
			Name:         "items",
			Description:  "Chargebee items (product catalog 2.0).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chargebeeItemFields(),
		},
	}
}

func chargebeeCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "auto_collection", Type: "string"},
		{Name: "net_term_days", Type: "integer"},
		{Name: "taxability", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "deleted", Type: "boolean"},
	}
}

func chargebeeSubscriptionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "plan_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "plan_quantity", Type: "integer"},
		{Name: "plan_amount", Type: "integer"},
		{Name: "current_term_start", Type: "integer"},
		{Name: "current_term_end", Type: "integer"},
		{Name: "created_at", Type: "integer"},
		{Name: "started_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "deleted", Type: "boolean"},
	}
}

func chargebeeInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "subscription_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "total", Type: "integer"},
		{Name: "amount_paid", Type: "integer"},
		{Name: "amount_due", Type: "integer"},
		{Name: "date", Type: "integer"},
		{Name: "due_date", Type: "integer"},
		{Name: "paid_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "deleted", Type: "boolean"},
	}
}

func chargebeePlanFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "invoice_name", Type: "string"},
		{Name: "price", Type: "integer"},
		{Name: "currency_code", Type: "string"},
		{Name: "period", Type: "integer"},
		{Name: "period_unit", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func chargebeeItemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "item_family_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "is_shippable", Type: "boolean"},
		{Name: "enabled_for_checkout", Type: "boolean"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func chargebeeCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"first_name":      item["first_name"],
		"last_name":       item["last_name"],
		"email":           item["email"],
		"company":         item["company"],
		"phone":           item["phone"],
		"auto_collection": item["auto_collection"],
		"net_term_days":   item["net_term_days"],
		"taxability":      item["taxability"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
		"deleted":         item["deleted"],
	}
}

func chargebeeSubscriptionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"customer_id":        item["customer_id"],
		"plan_id":            item["plan_id"],
		"status":             item["status"],
		"currency_code":      item["currency_code"],
		"plan_quantity":      item["plan_quantity"],
		"plan_amount":        item["plan_amount"],
		"current_term_start": item["current_term_start"],
		"current_term_end":   item["current_term_end"],
		"created_at":         item["created_at"],
		"started_at":         item["started_at"],
		"updated_at":         item["updated_at"],
		"deleted":            item["deleted"],
	}
}

func chargebeeInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"customer_id":     item["customer_id"],
		"subscription_id": item["subscription_id"],
		"status":          item["status"],
		"currency_code":   item["currency_code"],
		"total":           item["total"],
		"amount_paid":     item["amount_paid"],
		"amount_due":      item["amount_due"],
		"date":            item["date"],
		"due_date":        item["due_date"],
		"paid_at":         item["paid_at"],
		"updated_at":      item["updated_at"],
		"deleted":         item["deleted"],
	}
}

func chargebeePlanRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"invoice_name":  item["invoice_name"],
		"price":         item["price"],
		"currency_code": item["currency_code"],
		"period":        item["period"],
		"period_unit":   item["period_unit"],
		"status":        item["status"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func chargebeeItemRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["id"],
		"name":                 item["name"],
		"type":                 item["type"],
		"item_family_id":       item["item_family_id"],
		"status":               item["status"],
		"is_shippable":         item["is_shippable"],
		"enabled_for_checkout": item["enabled_for_checkout"],
		"created_at":           item["created_at"],
		"updated_at":           item["updated_at"],
	}
}
