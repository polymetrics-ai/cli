package invoiced

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Invoiced API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Invoiced list endpoint path segment (e.g. "customers").
	resource string
	// mapRecord flattens a raw Invoiced object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// invoicedStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in invoicedStreams; the read
// path is fully data-driven from this table.
var invoicedStreamEndpoints = map[string]streamEndpoint{
	"customers":     {resource: "customers", mapRecord: invoicedCustomerRecord},
	"invoices":      {resource: "invoices", mapRecord: invoicedInvoiceRecord},
	"payments":      {resource: "payments", mapRecord: invoicedPaymentRecord},
	"subscriptions": {resource: "subscriptions", mapRecord: invoicedSubscriptionRecord},
	"estimates":     {resource: "estimates", mapRecord: invoicedEstimateRecord},
}

// invoicedStreams returns the connector's published stream catalog. Every
// Invoiced object exposes a numeric id and a unix `updated_at` timestamp, so the
// primary key is ["id"] and the cursor field is ["updated_at"] across the board.
func invoicedStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "customers",
			Description:  "Invoiced customers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       invoicedCustomerFields(),
		},
		{
			Name:         "invoices",
			Description:  "Invoiced invoices.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       invoicedInvoiceFields(),
		},
		{
			Name:         "payments",
			Description:  "Invoiced payments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       invoicedPaymentFields(),
		},
		{
			Name:         "subscriptions",
			Description:  "Invoiced subscriptions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       invoicedSubscriptionFields(),
		},
		{
			Name:         "estimates",
			Description:  "Invoiced estimates.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       invoicedEstimateFields(),
		},
	}
}

func invoicedCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "object", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "balance", Type: "number"},
		{Name: "phone", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func invoicedInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "object", Type: "string"},
		{Name: "customer", Type: "integer"},
		{Name: "number", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "total", Type: "number"},
		{Name: "balance", Type: "number"},
		{Name: "paid", Type: "boolean"},
		{Name: "closed", Type: "boolean"},
		{Name: "due_date", Type: "integer"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func invoicedPaymentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "object", Type: "string"},
		{Name: "customer", Type: "integer"},
		{Name: "invoice", Type: "integer"},
		{Name: "amount", Type: "number"},
		{Name: "currency", Type: "string"},
		{Name: "method", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "date", Type: "integer"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func invoicedSubscriptionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "object", Type: "string"},
		{Name: "customer", Type: "integer"},
		{Name: "plan", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "quantity", Type: "integer"},
		{Name: "start_date", Type: "integer"},
		{Name: "period_start", Type: "integer"},
		{Name: "period_end", Type: "integer"},
		{Name: "canceled_at", Type: "integer"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func invoicedEstimateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "object", Type: "string"},
		{Name: "customer", Type: "integer"},
		{Name: "number", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "total", Type: "number"},
		{Name: "approved", Type: "boolean"},
		{Name: "closed", Type: "boolean"},
		{Name: "expiration_date", Type: "integer"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func invoicedCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"object":     item["object"],
		"name":       item["name"],
		"number":     item["number"],
		"email":      item["email"],
		"type":       item["type"],
		"currency":   item["currency"],
		"balance":    item["balance"],
		"phone":      item["phone"],
		"country":    item["country"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func invoicedInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"object":     item["object"],
		"customer":   item["customer"],
		"number":     item["number"],
		"currency":   item["currency"],
		"status":     item["status"],
		"total":      item["total"],
		"balance":    item["balance"],
		"paid":       item["paid"],
		"closed":     item["closed"],
		"due_date":   item["due_date"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func invoicedPaymentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"object":     item["object"],
		"customer":   item["customer"],
		"invoice":    item["invoice"],
		"amount":     item["amount"],
		"currency":   item["currency"],
		"method":     item["method"],
		"status":     item["status"],
		"date":       item["date"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func invoicedSubscriptionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"object":       item["object"],
		"customer":     item["customer"],
		"plan":         item["plan"],
		"status":       item["status"],
		"quantity":     item["quantity"],
		"start_date":   item["start_date"],
		"period_start": item["period_start"],
		"period_end":   item["period_end"],
		"canceled_at":  item["canceled_at"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func invoicedEstimateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"object":          item["object"],
		"customer":        item["customer"],
		"number":          item["number"],
		"currency":        item["currency"],
		"status":          item["status"],
		"total":           item["total"],
		"approved":        item["approved"],
		"closed":          item["closed"],
		"expiration_date": item["expiration_date"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
	}
}
