package fastbill

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to its FastBill SERVICE operation (e.g.
// "customer.get"), the response collection key the records array lives under
// (e.g. "CUSTOMERS" within RESPONSE), and the record mapper that flattens the
// upstream upper-case objects into connectors.Records.
type streamEndpoint struct {
	// service is the FastBill SERVICE value, e.g. "customer.get".
	service string
	// collection is the key under RESPONSE that holds the records array, e.g.
	// "CUSTOMERS" for RESPONSE.CUSTOMERS.
	collection string
	// mapRecord flattens a raw FastBill object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// fastbillStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in fastbillStreams; the read
// path is fully data-driven from this table.
var fastbillStreamEndpoints = map[string]streamEndpoint{
	"customers":          {service: "customer.get", collection: "CUSTOMERS", mapRecord: fastbillCustomerRecord},
	"invoices":           {service: "invoice.get", collection: "INVOICES", mapRecord: fastbillInvoiceRecord},
	"products":           {service: "article.get", collection: "ARTICLES", mapRecord: fastbillProductRecord},
	"recurring_invoices": {service: "recurring.get", collection: "INVOICES", mapRecord: fastbillRecurringInvoiceRecord},
	"revenues":           {service: "revenue.get", collection: "REVENUES", mapRecord: fastbillRevenueRecord},
}

// fastbillStreams returns the connector's published stream catalog. FastBill is
// full-refresh only upstream, so no cursor fields are declared.
func fastbillStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "customers",
			Description: "FastBill customers.",
			PrimaryKey:  []string{"customer_id"},
			Fields:      fastbillCustomerFields(),
		},
		{
			Name:        "invoices",
			Description: "FastBill invoices.",
			PrimaryKey:  []string{"invoice_id"},
			Fields:      fastbillInvoiceFields(),
		},
		{
			Name:        "products",
			Description: "FastBill articles/products.",
			PrimaryKey:  []string{"article_number"},
			Fields:      fastbillProductFields(),
		},
		{
			Name:        "recurring_invoices",
			Description: "FastBill recurring invoice templates.",
			PrimaryKey:  []string{"invoice_id"},
			Fields:      fastbillInvoiceFields(),
		},
		{
			Name:        "revenues",
			Description: "FastBill revenues.",
			PrimaryKey:  []string{"invoice_id"},
			Fields:      fastbillRevenueFields(),
		},
	}
}

func fastbillCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "customer_id", Type: "string"},
		{Name: "customer_number", Type: "string"},
		{Name: "customer_type", Type: "string"},
		{Name: "organization", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "country_code", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "created", Type: "string"},
	}
}

func fastbillInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "invoice_id", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "invoice_number", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "invoice_date", Type: "string"},
		{Name: "due_date", Type: "string"},
		{Name: "total", Type: "string"},
		{Name: "vat_total", Type: "string"},
		{Name: "sub_total", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "is_canceled", Type: "string"},
	}
}

func fastbillProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "article_number", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "unit_price", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "vat_percent", Type: "string"},
		{Name: "is_greedy", Type: "string"},
	}
}

func fastbillRevenueFields() []connectors.Field {
	return []connectors.Field{
		{Name: "invoice_id", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "invoice_number", Type: "string"},
		{Name: "invoice_date", Type: "string"},
		{Name: "total", Type: "string"},
		{Name: "vat_total", Type: "string"},
		{Name: "currency_code", Type: "string"},
	}
}

func fastbillCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"customer_id":     item["CUSTOMER_ID"],
		"customer_number": item["CUSTOMER_NUMBER"],
		"customer_type":   item["CUSTOMER_TYPE"],
		"organization":    item["ORGANIZATION"],
		"first_name":      item["FIRST_NAME"],
		"last_name":       item["LAST_NAME"],
		"email":           item["EMAIL"],
		"phone":           item["PHONE"],
		"country_code":    item["COUNTRY_CODE"],
		"currency_code":   item["CURRENCY_CODE"],
		"created":         item["CREATED"],
	}
}

func fastbillInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"invoice_id":     item["INVOICE_ID"],
		"customer_id":    item["CUSTOMER_ID"],
		"invoice_number": item["INVOICE_NUMBER"],
		"type":           item["TYPE"],
		"invoice_date":   item["INVOICE_DATE"],
		"due_date":       item["DUE_DATE"],
		"total":          item["TOTAL"],
		"vat_total":      item["VAT_TOTAL"],
		"sub_total":      item["SUB_TOTAL"],
		"currency_code":  item["CURRENCY_CODE"],
		"is_canceled":    item["IS_CANCELED"],
	}
}

// fastbillRecurringInvoiceRecord mirrors the invoice shape; recurring.get returns
// invoice-shaped templates under RESPONSE.INVOICES.
func fastbillRecurringInvoiceRecord(item map[string]any) connectors.Record {
	return fastbillInvoiceRecord(item)
}

func fastbillProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"article_number": item["ARTICLE_NUMBER"],
		"title":          item["TITLE"],
		"description":    item["DESCRIPTION"],
		"unit_price":     item["UNIT_PRICE"],
		"currency_code":  item["CURRENCY_CODE"],
		"vat_percent":    item["VAT_PERCENT"],
		"is_greedy":      item["IS_GREEDY"],
	}
}

func fastbillRevenueRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"invoice_id":     item["INVOICE_ID"],
		"customer_id":    item["CUSTOMER_ID"],
		"invoice_number": item["INVOICE_NUMBER"],
		"invoice_date":   item["INVOICE_DATE"],
		"total":          item["TOTAL"],
		"vat_total":      item["VAT_TOTAL"],
		"currency_code":  item["CURRENCY_CODE"],
	}
}
