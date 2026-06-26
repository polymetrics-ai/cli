package lightspeedretail

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Lightspeed Retail (X-Series) API
// resource path (relative to base_url) it reads from, and the record mapper that
// flattens its objects.
type streamEndpoint struct {
	// resource is the X-Series list endpoint path (e.g. "api/2.0/products").
	resource string
	// mapRecord flattens a raw Lightspeed object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// lightspeedStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in lightspeedStreams; the
// read path is fully data-driven from this table. Every endpoint returns a JSON
// object with a "data" array and a "version" object whose "max" drives cursor
// pagination via the "after" parameter.
var lightspeedStreamEndpoints = map[string]streamEndpoint{
	"products":  {resource: "api/2.0/products", mapRecord: lightspeedProductRecord},
	"customers": {resource: "api/2.0/customers", mapRecord: lightspeedCustomerRecord},
	"sales":     {resource: "api/2.0/sales", mapRecord: lightspeedSaleRecord},
	"outlets":   {resource: "api/2.0/outlets", mapRecord: lightspeedOutletRecord},
	"registers": {resource: "api/2.0/registers", mapRecord: lightspeedRegisterRecord},
}

// lightspeedStreams returns the connector's published stream catalog. Every
// X-Series object exposes a string id and a numeric version, so the primary key
// is ["id"] and the incremental cursor field is ["version"] across the board.
func lightspeedStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "products",
			Description:  "Lightspeed Retail products (with variants and pricing).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"version"},
			Fields:       lightspeedProductFields(),
		},
		{
			Name:         "customers",
			Description:  "Lightspeed Retail customers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"version"},
			Fields:       lightspeedCustomerFields(),
		},
		{
			Name:         "sales",
			Description:  "Lightspeed Retail sales (orders/transactions).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"version"},
			Fields:       lightspeedSaleFields(),
		},
		{
			Name:         "outlets",
			Description:  "Lightspeed Retail outlets (store locations).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"version"},
			Fields:       lightspeedOutletFields(),
		},
		{
			Name:         "registers",
			Description:  "Lightspeed Retail registers (point-of-sale terminals).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"version"},
			Fields:       lightspeedRegisterFields(),
		},
	}
}

func lightspeedProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "version", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "handle", Type: "string"},
		{Name: "sku", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "brand_id", Type: "string"},
		{Name: "supplier_id", Type: "string"},
		{Name: "product_category", Type: "string"},
		{Name: "price_including_tax", Type: "number"},
		{Name: "price_excluding_tax", Type: "number"},
		{Name: "supply_price", Type: "number"},
		{Name: "has_variants", Type: "boolean"},
		{Name: "is_active", Type: "boolean"},
		{Name: "is_composite", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func lightspeedCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "version", Type: "integer"},
		{Name: "customer_code", Type: "string"},
		{Name: "customer_group_id", Type: "string"},
		{Name: "balance", Type: "number"},
		{Name: "loyalty_balance", Type: "number"},
		{Name: "year_to_date", Type: "number"},
		{Name: "do_not_email", Type: "boolean"},
		{Name: "enable_loyalty", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func lightspeedSaleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "version", Type: "integer"},
		{Name: "invoice_number", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "register_id", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "total_price", Type: "number"},
		{Name: "total_tax", Type: "number"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "sale_date", Type: "timestamp"},
	}
}

func lightspeedOutletFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "version", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "currency_symbol", Type: "string"},
		{Name: "default_tax_id", Type: "string"},
		{Name: "display_prices", Type: "string"},
		{Name: "time_zone", Type: "string"},
	}
}

func lightspeedRegisterFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "version", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "outlet_id", Type: "string"},
		{Name: "invoice_prefix", Type: "string"},
		{Name: "invoice_sequence", Type: "integer"},
		{Name: "is_open", Type: "boolean"},
		{Name: "email_receipt", Type: "boolean"},
		{Name: "print_receipt", Type: "boolean"},
	}
}

func lightspeedProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"version":             item["version"],
		"name":                item["name"],
		"handle":              item["handle"],
		"sku":                 item["sku"],
		"description":         item["description"],
		"brand_id":            item["brand_id"],
		"supplier_id":         item["supplier_id"],
		"product_category":    item["product_category"],
		"price_including_tax": item["price_including_tax"],
		"price_excluding_tax": item["price_excluding_tax"],
		"supply_price":        item["supply_price"],
		"has_variants":        item["has_variants"],
		"is_active":           item["is_active"],
		"is_composite":        item["is_composite"],
		"created_at":          item["created_at"],
		"updated_at":          item["updated_at"],
	}
}

func lightspeedCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"version":           item["version"],
		"customer_code":     item["customer_code"],
		"customer_group_id": item["customer_group_id"],
		"balance":           item["balance"],
		"loyalty_balance":   item["loyalty_balance"],
		"year_to_date":      item["year_to_date"],
		"do_not_email":      item["do_not_email"],
		"enable_loyalty":    item["enable_loyalty"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
	}
}

func lightspeedSaleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"version":        item["version"],
		"invoice_number": item["invoice_number"],
		"customer_id":    item["customer_id"],
		"status":         item["status"],
		"register_id":    item["register_id"],
		"user_id":        item["user_id"],
		"total_price":    item["total_price"],
		"total_tax":      item["total_tax"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
		"sale_date":      item["sale_date"],
	}
}

func lightspeedOutletRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"version":         item["version"],
		"name":            item["name"],
		"currency":        item["currency"],
		"currency_symbol": item["currency_symbol"],
		"default_tax_id":  item["default_tax_id"],
		"display_prices":  item["display_prices"],
		"time_zone":       item["time_zone"],
	}
}

func lightspeedRegisterRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"version":          item["version"],
		"name":             item["name"],
		"outlet_id":        item["outlet_id"],
		"invoice_prefix":   item["invoice_prefix"],
		"invoice_sequence": item["invoice_sequence"],
		"is_open":          item["is_open"],
		"email_receipt":    item["email_receipt"],
		"print_receipt":    item["print_receipt"],
	}
}
