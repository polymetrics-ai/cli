package inflowinventory

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the inFlow API resource path segment
// (appended after /<companyid>/) and the record mapper for its objects. inFlow
// resources return a top-level JSON array of objects.
type streamEndpoint struct {
	// resource is the path segment under /<companyid>/ (e.g. "products").
	resource string
	// primaryKey is the object's id field; it doubles as the `after` cursor for
	// inFlow's count/after pagination.
	primaryKey string
	// mapRecord flattens a raw inFlow object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// inflowStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in inflowStreams; the read path
// is fully data-driven from this table.
var inflowStreamEndpoints = map[string]streamEndpoint{
	"products":     {resource: "products", primaryKey: "productId", mapRecord: inflowProductRecord},
	"customers":    {resource: "customers", primaryKey: "customerId", mapRecord: inflowCustomerRecord},
	"vendors":      {resource: "vendors", primaryKey: "vendorId", mapRecord: inflowVendorRecord},
	"sales_orders": {resource: "sales-orders", primaryKey: "salesOrderId", mapRecord: inflowSalesOrderRecord},
	"categories":   {resource: "categories", primaryKey: "categoryId", mapRecord: inflowCategoryRecord},
}

// inflowStreams returns the connector's published stream catalog. inFlow objects
// carry no shared incremental cursor across every resource; products/customers/
// vendors expose lastModifiedById/timestamp but the API does not filter by it, so
// streams are published as full-refresh (no CursorFields).
func inflowStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "products",
			Description:  "inFlow products (inventory items).",
			PrimaryKey:   []string{"productId"},
			CursorFields: []string{"lastModifiedDateTime"},
			Fields:       inflowProductFields(),
		},
		{
			Name:        "customers",
			Description: "inFlow customers.",
			PrimaryKey:  []string{"customerId"},
			Fields:      inflowCustomerFields(),
		},
		{
			Name:        "vendors",
			Description: "inFlow vendors (suppliers).",
			PrimaryKey:  []string{"vendorId"},
			Fields:      inflowVendorFields(),
		},
		{
			Name:        "sales_orders",
			Description: "inFlow sales orders.",
			PrimaryKey:  []string{"salesOrderId"},
			Fields:      inflowSalesOrderFields(),
		},
		{
			Name:        "categories",
			Description: "inFlow product categories.",
			PrimaryKey:  []string{"categoryId"},
			Fields:      inflowCategoryFields(),
		},
	}
}

func inflowProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "productId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "sku", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "categoryId", Type: "string"},
		{Name: "itemType", Type: "string"},
		{Name: "isActive", Type: "boolean"},
		{Name: "isManufacturable", Type: "boolean"},
		{Name: "trackSerials", Type: "boolean"},
		{Name: "lastModifiedDateTime", Type: "string"},
		{Name: "timestamp", Type: "string"},
	}
}

func inflowCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "customerId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "contactName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "fax", Type: "string"},
		{Name: "isActive", Type: "boolean"},
		{Name: "remarks", Type: "string"},
		{Name: "pricingSchemeId", Type: "string"},
		{Name: "taxingSchemeId", Type: "string"},
		{Name: "timestamp", Type: "string"},
	}
}

func inflowVendorFields() []connectors.Field {
	return []connectors.Field{
		{Name: "vendorId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "contactName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "fax", Type: "string"},
		{Name: "isActive", Type: "boolean"},
		{Name: "leadTimeDays", Type: "integer"},
		{Name: "currencyId", Type: "string"},
		{Name: "taxingSchemeId", Type: "string"},
		{Name: "timestamp", Type: "string"},
	}
}

func inflowSalesOrderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "salesOrderId", Type: "string"},
		{Name: "customerId", Type: "string"},
		{Name: "contactName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "currencyId", Type: "string"},
		{Name: "amountPaid", Type: "string"},
		{Name: "balance", Type: "string"},
		{Name: "dueDate", Type: "string"},
		{Name: "invoicedDate", Type: "string"},
		{Name: "inventoryStatus", Type: "string"},
		{Name: "isCompleted", Type: "boolean"},
		{Name: "isInvoiced", Type: "boolean"},
		{Name: "isCancelled", Type: "boolean"},
		{Name: "isQuote", Type: "boolean"},
	}
}

func inflowCategoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "categoryId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "parentCategoryId", Type: "string"},
		{Name: "isDefault", Type: "boolean"},
		{Name: "timestamp", Type: "string"},
	}
}

func inflowProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"productId":            item["productId"],
		"name":                 item["name"],
		"sku":                  item["sku"],
		"description":          item["description"],
		"categoryId":           item["categoryId"],
		"itemType":             item["itemType"],
		"isActive":             item["isActive"],
		"isManufacturable":     item["isManufacturable"],
		"trackSerials":         item["trackSerials"],
		"lastModifiedDateTime": item["lastModifiedDateTime"],
		"timestamp":            item["timestamp"],
	}
}

func inflowCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"customerId":      item["customerId"],
		"name":            item["name"],
		"contactName":     item["contactName"],
		"email":           item["email"],
		"phone":           item["phone"],
		"fax":             item["fax"],
		"isActive":        item["isActive"],
		"remarks":         item["remarks"],
		"pricingSchemeId": item["pricingSchemeId"],
		"taxingSchemeId":  item["taxingSchemeId"],
		"timestamp":       item["timestamp"],
	}
}

func inflowVendorRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"vendorId":       item["vendorId"],
		"name":           item["name"],
		"contactName":    item["contactName"],
		"email":          item["email"],
		"phone":          item["phone"],
		"fax":            item["fax"],
		"isActive":       item["isActive"],
		"leadTimeDays":   item["leadTimeDays"],
		"currencyId":     item["currencyId"],
		"taxingSchemeId": item["taxingSchemeId"],
		"timestamp":      item["timestamp"],
	}
}

func inflowSalesOrderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"salesOrderId":    item["salesOrderId"],
		"customerId":      item["customerId"],
		"contactName":     item["contactName"],
		"email":           item["email"],
		"currencyId":      item["currencyId"],
		"amountPaid":      item["amountPaid"],
		"balance":         item["balance"],
		"dueDate":         item["dueDate"],
		"invoicedDate":    item["invoicedDate"],
		"inventoryStatus": item["inventoryStatus"],
		"isCompleted":     item["isCompleted"],
		"isInvoiced":      item["isInvoiced"],
		"isCancelled":     item["isCancelled"],
		"isQuote":         item["isQuote"],
	}
}

func inflowCategoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"categoryId":       item["categoryId"],
		"name":             item["name"],
		"parentCategoryId": item["parentCategoryId"],
		"isDefault":        item["isDefault"],
		"timestamp":        item["timestamp"],
	}
}
