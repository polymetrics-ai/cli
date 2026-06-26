package economic

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the e-conomic REST collection path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects into a connectors.Record.
type streamEndpoint struct {
	// resource is the e-conomic collection path segment (e.g. "customers" or
	// "invoices/booked").
	resource string
	// mapRecord flattens a raw e-conomic object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in economicStreams; the read path is
// fully data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"customers": {resource: "customers", mapRecord: customerRecord},
	"products":  {resource: "products", mapRecord: productRecord},
	"suppliers": {resource: "suppliers", mapRecord: supplierRecord},
	"accounts":  {resource: "accounts", mapRecord: accountRecord},
	"invoices":  {resource: "invoices/booked", mapRecord: invoiceRecord},
}

// economicStreams returns the connector's published stream catalog. e-conomic
// collections expose integer business keys (customerNumber, productNumber,
// supplierNumber, accountNumber, bookedInvoiceNumber) which become the primary
// key. The REST API supports full-refresh only, so no incremental cursor is
// declared.
func economicStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "customers",
			Description: "e-conomic customers (debtors).",
			PrimaryKey:  []string{"customer_number"},
			Fields:      customerFields(),
		},
		{
			Name:        "products",
			Description: "e-conomic products.",
			PrimaryKey:  []string{"product_number"},
			Fields:      productFields(),
		},
		{
			Name:        "suppliers",
			Description: "e-conomic suppliers (creditors).",
			PrimaryKey:  []string{"supplier_number"},
			Fields:      supplierFields(),
		},
		{
			Name:        "accounts",
			Description: "e-conomic chart-of-accounts entries.",
			PrimaryKey:  []string{"account_number"},
			Fields:      accountFields(),
		},
		{
			Name:        "invoices",
			Description: "e-conomic booked invoices.",
			PrimaryKey:  []string{"booked_invoice_number"},
			Fields:      invoiceFields(),
		},
	}
}

func customerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "customer_number", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "zip", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "address", Type: "string"},
		{Name: "balance", Type: "number"},
		{Name: "barred", Type: "boolean"},
		{Name: "credit_limit", Type: "number"},
		{Name: "vat_zone_number", Type: "integer"},
		{Name: "customer_group_number", Type: "integer"},
		{Name: "self", Type: "string"},
	}
}

func productFields() []connectors.Field {
	return []connectors.Field{
		{Name: "product_number", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "sales_price", Type: "number"},
		{Name: "cost_price", Type: "number"},
		{Name: "recommended_price", Type: "number"},
		{Name: "barred", Type: "boolean"},
		{Name: "unit_number", Type: "integer"},
		{Name: "product_group_number", Type: "integer"},
		{Name: "self", Type: "string"},
	}
}

func supplierFields() []connectors.Field {
	return []connectors.Field{
		{Name: "supplier_number", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "zip", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "address", Type: "string"},
		{Name: "barred", Type: "boolean"},
		{Name: "supplier_group_number", Type: "integer"},
		{Name: "vat_zone_number", Type: "integer"},
		{Name: "self", Type: "string"},
	}
}

func accountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "account_number", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "account_type", Type: "string"},
		{Name: "balance", Type: "number"},
		{Name: "block_direct_entries", Type: "boolean"},
		{Name: "debit_credit", Type: "string"},
		{Name: "vat_code", Type: "string"},
		{Name: "self", Type: "string"},
	}
}

func invoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "booked_invoice_number", Type: "integer"},
		{Name: "date", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "net_amount", Type: "number"},
		{Name: "gross_amount", Type: "number"},
		{Name: "vat_amount", Type: "number"},
		{Name: "remainder", Type: "number"},
		{Name: "due_date", Type: "string"},
		{Name: "customer_number", Type: "integer"},
		{Name: "payment_terms_number", Type: "integer"},
		{Name: "self", Type: "string"},
	}
}

func customerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"customer_number":       item["customerNumber"],
		"name":                  item["name"],
		"currency":              item["currency"],
		"email":                 item["email"],
		"city":                  item["city"],
		"zip":                   item["zip"],
		"country":               item["country"],
		"address":               item["address"],
		"balance":               item["balance"],
		"barred":                item["barred"],
		"credit_limit":          item["creditLimit"],
		"vat_zone_number":       refNumber(item["vatZone"], "vatZoneNumber"),
		"customer_group_number": refNumber(item["customerGroup"], "customerGroupNumber"),
		"self":                  item["self"],
	}
}

func productRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"product_number":       item["productNumber"],
		"name":                 item["name"],
		"description":          item["description"],
		"sales_price":          item["salesPrice"],
		"cost_price":           item["costPrice"],
		"recommended_price":    item["recommendedPrice"],
		"barred":               item["barred"],
		"unit_number":          refNumber(item["unit"], "unitNumber"),
		"product_group_number": refNumber(item["productGroup"], "productGroupNumber"),
		"self":                 item["self"],
	}
}

func supplierRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"supplier_number":       item["supplierNumber"],
		"name":                  item["name"],
		"currency":              item["currency"],
		"email":                 item["email"],
		"city":                  item["city"],
		"zip":                   item["zip"],
		"country":               item["country"],
		"address":               item["address"],
		"barred":                item["barred"],
		"supplier_group_number": refNumber(item["supplierGroup"], "supplierGroupNumber"),
		"vat_zone_number":       refNumber(item["vatZone"], "vatZoneNumber"),
		"self":                  item["self"],
	}
}

func accountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"account_number":       item["accountNumber"],
		"name":                 item["name"],
		"account_type":         item["accountType"],
		"balance":              item["balance"],
		"block_direct_entries": item["blockDirectEntries"],
		"debit_credit":         item["debitCredit"],
		"vat_code":             item["vatCode"],
		"self":                 item["self"],
	}
}

func invoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"booked_invoice_number": item["bookedInvoiceNumber"],
		"date":                  item["date"],
		"currency":              item["currency"],
		"net_amount":            item["netAmount"],
		"gross_amount":          item["grossAmount"],
		"vat_amount":            item["vatAmount"],
		"remainder":             item["remainder"],
		"due_date":              item["dueDate"],
		"customer_number":       refNumber(item["customer"], "customerNumber"),
		"payment_terms_number":  refNumber(item["paymentTerms"], "paymentTermsNumber"),
		"self":                  item["self"],
	}
}

// refNumber pulls a numeric field out of an e-conomic reference object. Many
// e-conomic fields are nested as {"<thing>Number": N, "self": "..."}; this
// flattens the business key out so downstream consumers get a scalar. A missing
// or non-object reference yields nil.
func refNumber(ref any, key string) any {
	obj, ok := ref.(map[string]any)
	if !ok {
		return nil
	}
	return obj[key]
}
