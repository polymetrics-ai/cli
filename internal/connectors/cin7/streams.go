package cin7

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Cin7 Core API resource path (relative
// to base_url), the JSON record-selector path that holds the array of objects in
// the response, and the record mapper that flattens those objects.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "product").
	resource string
	// recordsPath is the dotted JSON path to the records array in the response
	// envelope (e.g. "Products", "CustomerList").
	recordsPath string
	// params are extra query parameters always sent for this stream (e.g. the
	// Include* expansion flags Cin7 exposes).
	params map[string]string
	// mapRecord flattens a raw Cin7 object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// cin7StreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in cin7Streams; the read path is
// fully data-driven from this table. These are the core Cin7 Core (DEAR) objects.
var cin7StreamEndpoints = map[string]streamEndpoint{
	"products": {
		resource:    "product",
		recordsPath: "Products",
		params: map[string]string{
			"IncludeDeprecated": "true",
		},
		mapRecord: cin7ProductRecord,
	},
	"customers": {
		resource:    "customer",
		recordsPath: "CustomerList",
		params: map[string]string{
			"IncludeDeprecated": "true",
		},
		mapRecord: cin7CustomerRecord,
	},
	"suppliers": {
		resource:    "supplier",
		recordsPath: "SupplierList",
		params: map[string]string{
			"IncludeDeprecated": "true",
		},
		mapRecord: cin7SupplierRecord,
	},
	"sale_list": {
		resource:    "saleList",
		recordsPath: "SaleList",
		mapRecord:   cin7SaleRecord,
	},
	"purchase_list": {
		resource:    "purchaseList",
		recordsPath: "PurchaseList",
		mapRecord:   cin7PurchaseRecord,
	},
}

// cin7Streams returns the connector's published stream catalog. Cin7 Core
// objects are keyed by the string "ID"; the API exposes only full-refresh syncs,
// so no incremental cursor field is published.
func cin7Streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "products",
			Description: "Cin7 Core products (the product master catalog).",
			PrimaryKey:  []string{"id"},
			Fields:      cin7ProductFields(),
		},
		{
			Name:        "customers",
			Description: "Cin7 Core customers.",
			PrimaryKey:  []string{"id"},
			Fields:      cin7CustomerFields(),
		},
		{
			Name:        "suppliers",
			Description: "Cin7 Core suppliers.",
			PrimaryKey:  []string{"id"},
			Fields:      cin7SupplierFields(),
		},
		{
			Name:        "sale_list",
			Description: "Cin7 Core sale order summaries.",
			PrimaryKey:  []string{"id"},
			Fields:      cin7SaleFields(),
		},
		{
			Name:        "purchase_list",
			Description: "Cin7 Core purchase order summaries.",
			PrimaryKey:  []string{"id"},
			Fields:      cin7PurchaseFields(),
		},
	}
}

func cin7ProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "sku", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "brand", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "uom", Type: "string"},
		{Name: "price_tier1", Type: "number"},
		{Name: "cost", Type: "number"},
		{Name: "last_modified", Type: "string"},
	}
}

func cin7CustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "payment_term", Type: "string"},
		{Name: "tax_rule", Type: "string"},
		{Name: "last_modified", Type: "string"},
	}
}

func cin7SupplierFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "payment_term", Type: "string"},
		{Name: "last_modified", Type: "string"},
	}
}

func cin7SaleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "order_number", Type: "string"},
		{Name: "customer", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "order_status", Type: "string"},
		{Name: "invoice_status", Type: "string"},
		{Name: "order_date", Type: "string"},
		{Name: "invoice_amount", Type: "number"},
		{Name: "last_modified", Type: "string"},
	}
}

func cin7PurchaseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "order_number", Type: "string"},
		{Name: "supplier", Type: "string"},
		{Name: "supplier_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "order_status", Type: "string"},
		{Name: "invoice_amount", Type: "number"},
		{Name: "order_date", Type: "string"},
		{Name: "last_modified", Type: "string"},
	}
}

func cin7ProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            firstField(item, "ID", "SKU"),
		"sku":           item["SKU"],
		"name":          item["Name"],
		"category":      item["Category"],
		"brand":         item["Brand"],
		"type":          item["Type"],
		"status":        item["Status"],
		"uom":           item["UOM"],
		"price_tier1":   item["PriceTier1"],
		"cost":          item["AverageCost"],
		"last_modified": item["LastModifiedOn"],
	}
}

func cin7CustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["ID"],
		"name":          item["Name"],
		"email":         item["Email"],
		"phone":         item["Phone"],
		"status":        item["Status"],
		"currency":      item["Currency"],
		"payment_term":  item["PaymentTerm"],
		"tax_rule":      item["TaxRule"],
		"last_modified": item["LastModifiedOn"],
	}
}

func cin7SupplierRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["ID"],
		"name":          item["Name"],
		"email":         item["Email"],
		"phone":         item["Phone"],
		"status":        item["Status"],
		"currency":      item["Currency"],
		"payment_term":  item["PaymentTerm"],
		"last_modified": item["LastModifiedOn"],
	}
}

func cin7SaleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             firstField(item, "SaleID", "ID"),
		"order_number":   item["OrderNumber"],
		"customer":       item["Customer"],
		"customer_id":    item["CustomerID"],
		"status":         item["Status"],
		"order_status":   item["OrderStatus"],
		"invoice_status": item["InvoiceStatus"],
		"order_date":     item["OrderDate"],
		"invoice_amount": item["InvoiceAmount"],
		"last_modified":  item["LastModifiedOn"],
	}
}

func cin7PurchaseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             firstField(item, "ID", "TaskID"),
		"order_number":   item["OrderNumber"],
		"supplier":       item["Supplier"],
		"supplier_id":    item["SupplierID"],
		"status":         item["Status"],
		"order_status":   item["OrderStatus"],
		"invoice_amount": item["InvoiceAmount"],
		"order_date":     item["OrderDate"],
		"last_modified":  item["LastModifiedOn"],
	}
}
