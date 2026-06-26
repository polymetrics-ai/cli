package getlago

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Lago API resource path (relative to
// api_url), the JSON key holding its record array, and the record mapper.
type streamEndpoint struct {
	// resource is the Lago list endpoint path segment (e.g. "customers").
	resource string
	// recordsKey is the top-level JSON key wrapping the array (Lago wraps each
	// list under a resource-named key alongside a "meta" pagination object).
	recordsKey string
	// mapRecord flattens a raw Lago object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// getlagoStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in getlagoStreams; the read path
// is fully data-driven from this table.
var getlagoStreamEndpoints = map[string]streamEndpoint{
	"customers":        {resource: "customers", recordsKey: "customers", mapRecord: customerRecord},
	"invoices":         {resource: "invoices", recordsKey: "invoices", mapRecord: invoiceRecord},
	"subscriptions":    {resource: "subscriptions", recordsKey: "subscriptions", mapRecord: subscriptionRecord},
	"plans":            {resource: "plans", recordsKey: "plans", mapRecord: planRecord},
	"billable_metrics": {resource: "billable_metrics", recordsKey: "billable_metrics", mapRecord: billableMetricRecord},
}

// getlagoStreams returns the connector's published stream catalog. Every Lago
// object exposes a string lago_id (UUID) primary key and a created_at RFC3339
// timestamp, so the cursor field is created_at across the board.
func getlagoStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "customers",
			Description:  "Lago customers.",
			PrimaryKey:   []string{"lago_id"},
			CursorFields: []string{"created_at"},
			Fields:       customerFields(),
		},
		{
			Name:         "invoices",
			Description:  "Lago invoices.",
			PrimaryKey:   []string{"lago_id"},
			CursorFields: []string{"created_at"},
			Fields:       invoiceFields(),
		},
		{
			Name:         "subscriptions",
			Description:  "Lago subscriptions.",
			PrimaryKey:   []string{"lago_id"},
			CursorFields: []string{"created_at"},
			Fields:       subscriptionFields(),
		},
		{
			Name:         "plans",
			Description:  "Lago plans.",
			PrimaryKey:   []string{"lago_id"},
			CursorFields: []string{"created_at"},
			Fields:       planFields(),
		},
		{
			Name:         "billable_metrics",
			Description:  "Lago billable metrics.",
			PrimaryKey:   []string{"lago_id"},
			CursorFields: []string{"created_at"},
			Fields:       billableMetricFields(),
		},
	}
}

func customerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "lago_id", Type: "string"},
		{Name: "sequential_id", Type: "integer"},
		{Name: "external_id", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "customer_type", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func invoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "lago_id", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "issuing_date", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "payment_status", Type: "string"},
		{Name: "invoice_type", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "fees_amount_cents", Type: "integer"},
		{Name: "taxes_amount_cents", Type: "integer"},
		{Name: "total_amount_cents", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func subscriptionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "lago_id", Type: "string"},
		{Name: "external_id", Type: "string"},
		{Name: "lago_customer_id", Type: "string"},
		{Name: "external_customer_id", Type: "string"},
		{Name: "plan_code", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "billing_time", Type: "string"},
		{Name: "started_at", Type: "string"},
		{Name: "terminated_at", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func planFields() []connectors.Field {
	return []connectors.Field{
		{Name: "lago_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "interval", Type: "string"},
		{Name: "amount_cents", Type: "integer"},
		{Name: "amount_currency", Type: "string"},
		{Name: "pay_in_advance", Type: "boolean"},
		{Name: "trial_period", Type: "number"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func billableMetricFields() []connectors.Field {
	return []connectors.Field{
		{Name: "lago_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "aggregation_type", Type: "string"},
		{Name: "field_name", Type: "string"},
		{Name: "recurring", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func customerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"lago_id":       item["lago_id"],
		"sequential_id": item["sequential_id"],
		"external_id":   item["external_id"],
		"slug":          item["slug"],
		"name":          item["name"],
		"email":         item["email"],
		"currency":      item["currency"],
		"country":       item["country"],
		"customer_type": item["customer_type"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func invoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"lago_id":            item["lago_id"],
		"number":             item["number"],
		"issuing_date":       item["issuing_date"],
		"status":             item["status"],
		"payment_status":     item["payment_status"],
		"invoice_type":       item["invoice_type"],
		"currency":           item["currency"],
		"fees_amount_cents":  item["fees_amount_cents"],
		"taxes_amount_cents": item["taxes_amount_cents"],
		"total_amount_cents": item["total_amount_cents"],
		"created_at":         item["created_at"],
		"updated_at":         item["updated_at"],
	}
}

func subscriptionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"lago_id":              item["lago_id"],
		"external_id":          item["external_id"],
		"lago_customer_id":     item["lago_customer_id"],
		"external_customer_id": item["external_customer_id"],
		"plan_code":            item["plan_code"],
		"status":               item["status"],
		"billing_time":         item["billing_time"],
		"started_at":           item["started_at"],
		"terminated_at":        item["terminated_at"],
		"created_at":           item["created_at"],
	}
}

func planRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"lago_id":         item["lago_id"],
		"name":            item["name"],
		"code":            item["code"],
		"interval":        item["interval"],
		"amount_cents":    item["amount_cents"],
		"amount_currency": item["amount_currency"],
		"pay_in_advance":  item["pay_in_advance"],
		"trial_period":    item["trial_period"],
		"created_at":      item["created_at"],
	}
}

func billableMetricRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"lago_id":          item["lago_id"],
		"name":             item["name"],
		"code":             item["code"],
		"aggregation_type": item["aggregation_type"],
		"field_name":       item["field_name"],
		"recurring":        item["recurring"],
		"created_at":       item["created_at"],
	}
}
