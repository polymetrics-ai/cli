package adobecommercemagento

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Magento REST resource path (relative
// to <base>/rest/<api_version>) it reads from, and the record mapper that
// flattens its objects.
type streamEndpoint struct {
	// resource is the Magento list endpoint path segment after the version
	// prefix (e.g. "products", "customers/search", "categories/list").
	resource string
	// mapRecord flattens a raw Magento object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// magentoStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in magentoStreams; the read path
// is fully data-driven from this table.
//
// Magento exposes several search-backed list endpoints. products and orders are
// listed directly under their resource; customers and categories are listed via
// their dedicated search/list endpoints. All accept searchCriteria query params
// and return {"items":[...],"total_count":N,"search_criteria":{...}}.
var magentoStreamEndpoints = map[string]streamEndpoint{
	"products":   {resource: "products", mapRecord: magentoProductRecord},
	"orders":     {resource: "orders", mapRecord: magentoOrderRecord},
	"customers":  {resource: "customers/search", mapRecord: magentoCustomerRecord},
	"categories": {resource: "categories/list", mapRecord: magentoCategoryRecord},
	"invoices":   {resource: "invoices", mapRecord: magentoInvoiceRecord},
}

// magentoStreams returns the connector's published stream catalog. Magento
// entities carry an integer id (entity_id) and most carry an "updated_at"
// timestamp used as the incremental cursor.
func magentoStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "products",
			Description:  "Adobe Commerce (Magento) catalog products.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       magentoProductFields(),
		},
		{
			Name:         "orders",
			Description:  "Adobe Commerce (Magento) sales orders.",
			PrimaryKey:   []string{"entity_id"},
			CursorFields: []string{"updated_at"},
			Fields:       magentoOrderFields(),
		},
		{
			Name:         "customers",
			Description:  "Adobe Commerce (Magento) customers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       magentoCustomerFields(),
		},
		{
			Name:         "categories",
			Description:  "Adobe Commerce (Magento) catalog categories.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       magentoCategoryFields(),
		},
		{
			Name:         "invoices",
			Description:  "Adobe Commerce (Magento) sales invoices.",
			PrimaryKey:   []string{"entity_id"},
			CursorFields: []string{"updated_at"},
			Fields:       magentoInvoiceFields(),
		},
	}
}

func magentoProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "sku", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "attribute_set_id", Type: "integer"},
		{Name: "price", Type: "number"},
		{Name: "status", Type: "integer"},
		{Name: "visibility", Type: "integer"},
		{Name: "type_id", Type: "string"},
		{Name: "weight", Type: "number"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func magentoOrderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "entity_id", Type: "integer"},
		{Name: "increment_id", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "customer_id", Type: "integer"},
		{Name: "customer_email", Type: "string"},
		{Name: "grand_total", Type: "number"},
		{Name: "base_grand_total", Type: "number"},
		{Name: "order_currency_code", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func magentoCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "group_id", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "firstname", Type: "string"},
		{Name: "lastname", Type: "string"},
		{Name: "store_id", Type: "integer"},
		{Name: "website_id", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func magentoCategoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "parent_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "position", Type: "integer"},
		{Name: "level", Type: "integer"},
		{Name: "product_count", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func magentoInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "entity_id", Type: "integer"},
		{Name: "increment_id", Type: "string"},
		{Name: "order_id", Type: "integer"},
		{Name: "state", Type: "integer"},
		{Name: "grand_total", Type: "number"},
		{Name: "base_grand_total", Type: "number"},
		{Name: "store_id", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func magentoProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"sku":              item["sku"],
		"name":             item["name"],
		"attribute_set_id": item["attribute_set_id"],
		"price":            item["price"],
		"status":           item["status"],
		"visibility":       item["visibility"],
		"type_id":          item["type_id"],
		"weight":           item["weight"],
		"created_at":       item["created_at"],
		"updated_at":       item["updated_at"],
	}
}

func magentoOrderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"entity_id":           item["entity_id"],
		"increment_id":        item["increment_id"],
		"state":               item["state"],
		"status":              item["status"],
		"customer_id":         item["customer_id"],
		"customer_email":      item["customer_email"],
		"grand_total":         item["grand_total"],
		"base_grand_total":    item["base_grand_total"],
		"order_currency_code": item["order_currency_code"],
		"created_at":          item["created_at"],
		"updated_at":          item["updated_at"],
	}
}

func magentoCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"group_id":   item["group_id"],
		"email":      item["email"],
		"firstname":  item["firstname"],
		"lastname":   item["lastname"],
		"store_id":   item["store_id"],
		"website_id": item["website_id"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func magentoCategoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"parent_id":     item["parent_id"],
		"name":          item["name"],
		"is_active":     item["is_active"],
		"position":      item["position"],
		"level":         item["level"],
		"product_count": item["product_count"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func magentoInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"entity_id":        item["entity_id"],
		"increment_id":     item["increment_id"],
		"order_id":         item["order_id"],
		"state":            item["state"],
		"grand_total":      item["grand_total"],
		"base_grand_total": item["base_grand_total"],
		"store_id":         item["store_id"],
		"created_at":       item["created_at"],
		"updated_at":       item["updated_at"],
	}
}
