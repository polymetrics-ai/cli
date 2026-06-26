package katana

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Katana API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Katana list endpoint path segment (e.g. "products").
	resource string
	// mapRecord flattens a raw Katana object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// katanaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in katanaStreams; the read path
// is fully data-driven from this table.
var katanaStreamEndpoints = map[string]streamEndpoint{
	"products":     {resource: "products", mapRecord: katanaProductRecord},
	"materials":    {resource: "materials", mapRecord: katanaMaterialRecord},
	"variants":     {resource: "variants", mapRecord: katanaVariantRecord},
	"sales_orders": {resource: "sales_orders", mapRecord: katanaSalesOrderRecord},
	"customers":    {resource: "customers", mapRecord: katanaCustomerRecord},
}

// katanaStreams returns the connector's published stream catalog. Every Katana
// object exposes an integer id and ISO-8601 created_at/updated_at timestamps, so
// the primary key is ["id"] and the incremental cursor field is ["updated_at"]
// across the board.
func katanaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "products",
			Description:  "Katana products (sellable finished goods).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       katanaProductFields(),
		},
		{
			Name:         "materials",
			Description:  "Katana materials (purchasable inputs).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       katanaMaterialFields(),
		},
		{
			Name:         "variants",
			Description:  "Katana product/material variants (SKUs).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       katanaVariantFields(),
		},
		{
			Name:         "sales_orders",
			Description:  "Katana sales orders.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       katanaSalesOrderFields(),
		},
		{
			Name:         "customers",
			Description:  "Katana customers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       katanaCustomerFields(),
		},
	}
}

func katanaProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "uom", Type: "string"},
		{Name: "category_name", Type: "string"},
		{Name: "is_sellable", Type: "boolean"},
		{Name: "is_producible", Type: "boolean"},
		{Name: "is_purchasable", Type: "boolean"},
		{Name: "default_supplier_id", Type: "integer"},
		{Name: "additional_info", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func katanaMaterialFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "uom", Type: "string"},
		{Name: "category_name", Type: "string"},
		{Name: "default_supplier_id", Type: "integer"},
		{Name: "is_sellable", Type: "boolean"},
		{Name: "additional_info", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func katanaVariantFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "sku", Type: "string"},
		{Name: "product_id", Type: "integer"},
		{Name: "material_id", Type: "integer"},
		{Name: "sales_price", Type: "number"},
		{Name: "purchase_price", Type: "number"},
		{Name: "type", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func katanaSalesOrderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "order_no", Type: "string"},
		{Name: "customer_id", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "total", Type: "number"},
		{Name: "total_in_base_currency", Type: "number"},
		{Name: "order_created_date", Type: "timestamp"},
		{Name: "delivery_date", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func katanaCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "reference_id", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func katanaProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"name":                item["name"],
		"uom":                 item["uom"],
		"category_name":       item["category_name"],
		"is_sellable":         item["is_sellable"],
		"is_producible":       item["is_producible"],
		"is_purchasable":      item["is_purchasable"],
		"default_supplier_id": item["default_supplier_id"],
		"additional_info":     item["additional_info"],
		"created_at":          item["created_at"],
		"updated_at":          item["updated_at"],
	}
}

func katanaMaterialRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"name":                item["name"],
		"uom":                 item["uom"],
		"category_name":       item["category_name"],
		"default_supplier_id": item["default_supplier_id"],
		"is_sellable":         item["is_sellable"],
		"additional_info":     item["additional_info"],
		"created_at":          item["created_at"],
		"updated_at":          item["updated_at"],
	}
}

func katanaVariantRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"sku":            item["sku"],
		"product_id":     item["product_id"],
		"material_id":    item["material_id"],
		"sales_price":    item["sales_price"],
		"purchase_price": item["purchase_price"],
		"type":           item["type"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}

func katanaSalesOrderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                     item["id"],
		"order_no":               item["order_no"],
		"customer_id":            item["customer_id"],
		"status":                 item["status"],
		"currency":               item["currency"],
		"total":                  item["total"],
		"total_in_base_currency": item["total_in_base_currency"],
		"order_created_date":     item["order_created_date"],
		"delivery_date":          item["delivery_date"],
		"created_at":             item["created_at"],
		"updated_at":             item["updated_at"],
	}
}

func katanaCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"email":        item["email"],
		"phone":        item["phone"],
		"currency":     item["currency"],
		"reference_id": item["reference_id"],
		"category":     item["category"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}
