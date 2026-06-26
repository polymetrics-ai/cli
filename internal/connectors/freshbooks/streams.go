package freshbooks

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the FreshBooks accounting API resource
// path segment (relative to /accounting/account/{account_id}/) it reads from,
// the JSON key under response.result that holds the array, and the record mapper
// that flattens its objects.
type streamEndpoint struct {
	// resource is the path under /accounting/account/{account_id}/ for the list
	// endpoint, e.g. "users/clients".
	resource string
	// arrayKey is the key under response.result holding the array of records,
	// e.g. "clients".
	arrayKey string
	// mapRecord flattens a raw FreshBooks object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// freshbooksStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in freshbooksStreams; the
// read path is fully data-driven from this table.
var freshbooksStreamEndpoints = map[string]streamEndpoint{
	"clients":  {resource: "users/clients", arrayKey: "clients", mapRecord: freshbooksClientRecord},
	"invoices": {resource: "invoices/invoices", arrayKey: "invoices", mapRecord: freshbooksInvoiceRecord},
	"expenses": {resource: "expenses/expenses", arrayKey: "expenses", mapRecord: freshbooksExpenseRecord},
	"payments": {resource: "payments/payments", arrayKey: "payments", mapRecord: freshbooksPaymentRecord},
	"items":    {resource: "items/items", arrayKey: "items", mapRecord: freshbooksItemRecord},
}

// freshbooksStreams returns the connector's published stream catalog. Every
// FreshBooks accounting object exposes a numeric id and an "updated" timestamp,
// so the primary key is ["id"] and the incremental cursor field is ["updated"]
// across the board.
func freshbooksStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "clients",
			Description:  "FreshBooks clients (customers).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       freshbooksClientFields(),
		},
		{
			Name:         "invoices",
			Description:  "FreshBooks invoices.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       freshbooksInvoiceFields(),
		},
		{
			Name:         "expenses",
			Description:  "FreshBooks expenses.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       freshbooksExpenseFields(),
		},
		{
			Name:         "payments",
			Description:  "FreshBooks payments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       freshbooksPaymentFields(),
		},
		{
			Name:         "items",
			Description:  "FreshBooks items (catalog products/services).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       freshbooksItemFields(),
		},
	}
}

func freshbooksClientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "userid", Type: "integer"},
		{Name: "organization", Type: "string"},
		{Name: "fname", Type: "string"},
		{Name: "lname", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "vis_state", Type: "integer"},
		{Name: "updated", Type: "string"},
	}
}

func freshbooksInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "invoiceid", Type: "integer"},
		{Name: "invoice_number", Type: "string"},
		{Name: "customerid", Type: "integer"},
		{Name: "status", Type: "integer"},
		{Name: "amount", Type: "object"},
		{Name: "outstanding", Type: "object"},
		{Name: "currency_code", Type: "string"},
		{Name: "create_date", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func freshbooksExpenseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "expenseid", Type: "integer"},
		{Name: "categoryid", Type: "integer"},
		{Name: "staffid", Type: "integer"},
		{Name: "amount", Type: "object"},
		{Name: "vendor", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func freshbooksPaymentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "invoiceid", Type: "integer"},
		{Name: "amount", Type: "object"},
		{Name: "type", Type: "string"},
		{Name: "note", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func freshbooksItemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "itemid", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "qty", Type: "string"},
		{Name: "unit_cost", Type: "object"},
		{Name: "inventory", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func freshbooksClientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"userid":        item["userid"],
		"organization":  item["organization"],
		"fname":         item["fname"],
		"lname":         item["lname"],
		"email":         item["email"],
		"currency_code": item["currency_code"],
		"vis_state":     item["vis_state"],
		"updated":       item["updated"],
	}
}

func freshbooksInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"invoiceid":      item["invoiceid"],
		"invoice_number": item["invoice_number"],
		"customerid":     item["customerid"],
		"status":         item["status"],
		"amount":         item["amount"],
		"outstanding":    item["outstanding"],
		"currency_code":  item["currency_code"],
		"create_date":    item["create_date"],
		"updated":        item["updated"],
	}
}

func freshbooksExpenseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"expenseid":  item["expenseid"],
		"categoryid": item["categoryid"],
		"staffid":    item["staffid"],
		"amount":     item["amount"],
		"vendor":     item["vendor"],
		"notes":      item["notes"],
		"date":       item["date"],
		"updated":    item["updated"],
	}
}

func freshbooksPaymentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"invoiceid": item["invoiceid"],
		"amount":    item["amount"],
		"type":      item["type"],
		"note":      item["note"],
		"date":      item["date"],
		"updated":   item["updated"],
	}
}

func freshbooksItemRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"itemid":      item["itemid"],
		"name":        item["name"],
		"description": item["description"],
		"qty":         item["qty"],
		"unit_cost":   item["unit_cost"],
		"inventory":   item["inventory"],
		"updated":     item["updated"],
	}
}
