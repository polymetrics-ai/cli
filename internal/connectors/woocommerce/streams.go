package woocommerce

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the WooCommerce REST resource path
// (relative to base_url, e.g. "orders") and the record mapper that flattens its
// objects. WooCommerce list endpoints return a top-level JSON array.
type streamEndpoint struct {
	// resource is the WooCommerce list endpoint path segment (e.g. "orders").
	resource string
	// mapRecord flattens a raw WooCommerce object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// woocommerceStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in woocommerceStreams; the
// read path is fully data-driven from this table.
var woocommerceStreamEndpoints = map[string]streamEndpoint{
	"orders":    {resource: "orders", mapRecord: woocommerceOrderRecord},
	"products":  {resource: "products", mapRecord: woocommerceProductRecord},
	"customers": {resource: "customers", mapRecord: woocommerceCustomerRecord},
	"coupons":   {resource: "coupons", mapRecord: woocommerceCouponRecord},
}

// woocommerceStreams returns the connector's published stream catalog. Every
// WooCommerce object exposes an integer id; the incremental cursor field is
// date_modified_gmt (filtered with modified_after / date_modified_gmt) where the
// resource supports it.
func woocommerceStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "orders",
			Description:  "WooCommerce orders.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modified_gmt"},
			Fields:       woocommerceOrderFields(),
		},
		{
			Name:         "products",
			Description:  "WooCommerce products in the catalog.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modified_gmt"},
			Fields:       woocommerceProductFields(),
		},
		{
			Name:         "customers",
			Description:  "WooCommerce customers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modified_gmt"},
			Fields:       woocommerceCustomerFields(),
		},
		{
			Name:         "coupons",
			Description:  "WooCommerce coupons.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modified_gmt"},
			Fields:       woocommerceCouponFields(),
		},
	}
}

func woocommerceOrderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "number", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "total", Type: "string"},
		{Name: "total_tax", Type: "string"},
		{Name: "customer_id", Type: "integer"},
		{Name: "payment_method", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_created_gmt", Type: "string"},
		{Name: "date_modified", Type: "string"},
		{Name: "date_modified_gmt", Type: "string"},
		{Name: "date_paid", Type: "string"},
	}
}

func woocommerceProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "sku", Type: "string"},
		{Name: "price", Type: "string"},
		{Name: "regular_price", Type: "string"},
		{Name: "sale_price", Type: "string"},
		{Name: "stock_status", Type: "string"},
		{Name: "stock_quantity", Type: "integer"},
		{Name: "total_sales", Type: "integer"},
		{Name: "date_created_gmt", Type: "string"},
		{Name: "date_modified_gmt", Type: "string"},
	}
}

func woocommerceCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "is_paying_customer", Type: "boolean"},
		{Name: "date_created", Type: "string"},
		{Name: "date_created_gmt", Type: "string"},
		{Name: "date_modified", Type: "string"},
		{Name: "date_modified_gmt", Type: "string"},
	}
}

func woocommerceCouponFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "code", Type: "string"},
		{Name: "discount_type", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "usage_count", Type: "integer"},
		{Name: "usage_limit", Type: "integer"},
		{Name: "date_expires", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_created_gmt", Type: "string"},
		{Name: "date_modified", Type: "string"},
		{Name: "date_modified_gmt", Type: "string"},
	}
}

func woocommerceOrderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"number":            item["number"],
		"status":            item["status"],
		"currency":          item["currency"],
		"total":             item["total"],
		"total_tax":         item["total_tax"],
		"customer_id":       item["customer_id"],
		"payment_method":    item["payment_method"],
		"date_created":      item["date_created"],
		"date_created_gmt":  item["date_created_gmt"],
		"date_modified":     item["date_modified"],
		"date_modified_gmt": item["date_modified_gmt"],
		"date_paid":         item["date_paid"],
	}
}

func woocommerceProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"slug":              item["slug"],
		"type":              item["type"],
		"status":            item["status"],
		"sku":               item["sku"],
		"price":             item["price"],
		"regular_price":     item["regular_price"],
		"sale_price":        item["sale_price"],
		"stock_status":      item["stock_status"],
		"stock_quantity":    item["stock_quantity"],
		"total_sales":       item["total_sales"],
		"date_created_gmt":  item["date_created_gmt"],
		"date_modified_gmt": item["date_modified_gmt"],
	}
}

func woocommerceCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"email":              item["email"],
		"first_name":         item["first_name"],
		"last_name":          item["last_name"],
		"username":           item["username"],
		"role":               item["role"],
		"is_paying_customer": item["is_paying_customer"],
		"date_created":       item["date_created"],
		"date_created_gmt":   item["date_created_gmt"],
		"date_modified":      item["date_modified"],
		"date_modified_gmt":  item["date_modified_gmt"],
	}
}

func woocommerceCouponRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"code":              item["code"],
		"discount_type":     item["discount_type"],
		"amount":            item["amount"],
		"usage_count":       item["usage_count"],
		"usage_limit":       item["usage_limit"],
		"date_expires":      item["date_expires"],
		"date_created":      item["date_created"],
		"date_created_gmt":  item["date_created_gmt"],
		"date_modified":     item["date_modified"],
		"date_modified_gmt": item["date_modified_gmt"],
	}
}
