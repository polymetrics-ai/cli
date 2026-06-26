package invoiceninja

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Invoice Ninja API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects into a connectors.Record.
type streamEndpoint struct {
	// resource is the list endpoint path segment (e.g. "clients").
	resource string
	// mapRecord flattens a raw Invoice Ninja object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// invoiceNinjaStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in invoiceNinjaStreams;
// the read path is fully data-driven from this table.
var invoiceNinjaStreamEndpoints = map[string]streamEndpoint{
	"clients":  {resource: "clients", mapRecord: clientRecord},
	"invoices": {resource: "invoices", mapRecord: invoiceRecord},
	"products": {resource: "products", mapRecord: productRecord},
	"payments": {resource: "payments", mapRecord: paymentRecord},
	"quotes":   {resource: "quotes", mapRecord: quoteRecord},
}

// invoiceNinjaStreams returns the connector's published stream catalog. Every
// Invoice Ninja object exposes a string `id` primary key. The API only supports
// full-refresh syncs (no incremental cursor), so CursorFields is empty.
func invoiceNinjaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "clients",
			Description: "Invoice Ninja clients (customers).",
			PrimaryKey:  []string{"id"},
			Fields:      clientFields(),
		},
		{
			Name:        "invoices",
			Description: "Invoice Ninja invoices.",
			PrimaryKey:  []string{"id"},
			Fields:      invoiceFields(),
		},
		{
			Name:        "products",
			Description: "Invoice Ninja products.",
			PrimaryKey:  []string{"id"},
			Fields:      productFields(),
		},
		{
			Name:        "payments",
			Description: "Invoice Ninja payments.",
			PrimaryKey:  []string{"id"},
			Fields:      paymentFields(),
		},
		{
			Name:        "quotes",
			Description: "Invoice Ninja quotes.",
			PrimaryKey:  []string{"id"},
			Fields:      quoteFields(),
		},
	}
}

func clientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "balance", Type: "number"},
		{Name: "paid_to_date", Type: "number"},
		{Name: "currency_id", Type: "string"},
		{Name: "vat_number", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "is_deleted", Type: "boolean"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "archived_at", Type: "integer"},
	}
}

func invoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "client_id", Type: "string"},
		{Name: "status_id", Type: "string"},
		{Name: "amount", Type: "number"},
		{Name: "balance", Type: "number"},
		{Name: "paid_to_date", Type: "number"},
		{Name: "date", Type: "string"},
		{Name: "due_date", Type: "string"},
		{Name: "currency_id", Type: "string"},
		{Name: "is_deleted", Type: "boolean"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func productFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "product_key", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "price", Type: "number"},
		{Name: "cost", Type: "number"},
		{Name: "quantity", Type: "number"},
		{Name: "tax_name1", Type: "string"},
		{Name: "tax_rate1", Type: "number"},
		{Name: "is_deleted", Type: "boolean"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func paymentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "client_id", Type: "string"},
		{Name: "amount", Type: "number"},
		{Name: "applied", Type: "number"},
		{Name: "refunded", Type: "number"},
		{Name: "status_id", Type: "string"},
		{Name: "transaction_reference", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "currency_id", Type: "string"},
		{Name: "is_deleted", Type: "boolean"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func quoteFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "client_id", Type: "string"},
		{Name: "status_id", Type: "string"},
		{Name: "amount", Type: "number"},
		{Name: "balance", Type: "number"},
		{Name: "date", Type: "string"},
		{Name: "due_date", Type: "string"},
		{Name: "valid_until", Type: "string"},
		{Name: "currency_id", Type: "string"},
		{Name: "is_deleted", Type: "boolean"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
	}
}

func clientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"display_name": item["display_name"],
		"number":       item["number"],
		"balance":      item["balance"],
		"paid_to_date": item["paid_to_date"],
		"currency_id":  item["currency_id"],
		"vat_number":   item["vat_number"],
		"phone":        item["phone"],
		"website":      item["website"],
		"is_deleted":   item["is_deleted"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
		"archived_at":  item["archived_at"],
	}
}

func invoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"number":       item["number"],
		"client_id":    item["client_id"],
		"status_id":    item["status_id"],
		"amount":       item["amount"],
		"balance":      item["balance"],
		"paid_to_date": item["paid_to_date"],
		"date":         item["date"],
		"due_date":     item["due_date"],
		"currency_id":  item["currency_id"],
		"is_deleted":   item["is_deleted"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func productRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"product_key": item["product_key"],
		"notes":       item["notes"],
		"price":       item["price"],
		"cost":        item["cost"],
		"quantity":    item["quantity"],
		"tax_name1":   item["tax_name1"],
		"tax_rate1":   item["tax_rate1"],
		"is_deleted":  item["is_deleted"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func paymentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["id"],
		"number":                item["number"],
		"client_id":             item["client_id"],
		"amount":                item["amount"],
		"applied":               item["applied"],
		"refunded":              item["refunded"],
		"status_id":             item["status_id"],
		"transaction_reference": item["transaction_reference"],
		"date":                  item["date"],
		"currency_id":           item["currency_id"],
		"is_deleted":            item["is_deleted"],
		"created_at":            item["created_at"],
		"updated_at":            item["updated_at"],
	}
}

func quoteRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"number":      item["number"],
		"client_id":   item["client_id"],
		"status_id":   item["status_id"],
		"amount":      item["amount"],
		"balance":     item["balance"],
		"date":        item["date"],
		"due_date":    item["due_date"],
		"valid_until": item["valid_until"],
		"currency_id": item["currency_id"],
		"is_deleted":  item["is_deleted"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}
