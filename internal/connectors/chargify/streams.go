package chargify

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Chargify API resource path (relative
// to base_url) it reads from, the wrapper key Chargify nests each list element
// under (e.g. "customer" for /customers.json), and the record mapper.
//
// Chargify list endpoints return a top-level JSON array of single-key objects:
//
//	[{"customer": {...}}, {"customer": {...}}]
//
// so reading a page means selecting the root array and unwrapping each element
// by its wrapKey before mapping.
type streamEndpoint struct {
	// resource is the path segment including the .json suffix (e.g.
	// "customers.json").
	resource string
	// wrapKey is the singular key each list element nests the object under.
	wrapKey string
	// mapRecord flattens a raw (unwrapped) Chargify object into a Record.
	mapRecord func(map[string]any) connectors.Record
}

// chargifyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in chargifyStreams; the read
// path is fully data-driven from this table.
var chargifyStreamEndpoints = map[string]streamEndpoint{
	"customers":     {resource: "customers.json", wrapKey: "customer", mapRecord: chargifyCustomerRecord},
	"subscriptions": {resource: "subscriptions.json", wrapKey: "subscription", mapRecord: chargifySubscriptionRecord},
	"products":      {resource: "products.json", wrapKey: "product", mapRecord: chargifyProductRecord},
	"coupons":       {resource: "coupons.json", wrapKey: "coupon", mapRecord: chargifyCouponRecord},
	"transactions":  {resource: "transactions.json", wrapKey: "transaction", mapRecord: chargifyTransactionRecord},
}

// chargifyStreams returns the connector's published stream catalog. Every
// Chargify object exposes a numeric id and an updated_at timestamp, so the
// primary key is ["id"] and the incremental cursor field is ["updated_at"].
func chargifyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "customers",
			Description:  "Chargify customers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chargifyCustomerFields(),
		},
		{
			Name:         "subscriptions",
			Description:  "Chargify subscriptions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chargifySubscriptionFields(),
		},
		{
			Name:         "products",
			Description:  "Chargify products.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chargifyProductFields(),
		},
		{
			Name:         "coupons",
			Description:  "Chargify coupons.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       chargifyCouponFields(),
		},
		{
			Name:         "transactions",
			Description:  "Chargify transactions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       chargifyTransactionFields(),
		},
	}
}

func chargifyCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "organization", Type: "string"},
		{Name: "reference", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func chargifySubscriptionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "state", Type: "string"},
		{Name: "customer_id", Type: "integer"},
		{Name: "product_id", Type: "integer"},
		{Name: "balance_in_cents", Type: "integer"},
		{Name: "total_revenue_in_cents", Type: "integer"},
		{Name: "current_period_started_at", Type: "string"},
		{Name: "current_period_ends_at", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func chargifyProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "handle", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "price_in_cents", Type: "integer"},
		{Name: "interval", Type: "integer"},
		{Name: "interval_unit", Type: "string"},
		{Name: "product_family_id", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func chargifyCouponFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "percentage", Type: "string"},
		{Name: "amount_in_cents", Type: "integer"},
		{Name: "product_family_id", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func chargifyTransactionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "transaction_type", Type: "string"},
		{Name: "amount_in_cents", Type: "integer"},
		{Name: "subscription_id", Type: "integer"},
		{Name: "customer_id", Type: "integer"},
		{Name: "product_id", Type: "integer"},
		{Name: "success", Type: "boolean"},
		{Name: "kind", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func chargifyCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"email":        item["email"],
		"organization": item["organization"],
		"reference":    item["reference"],
		"phone":        item["phone"],
		"country":      item["country"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func chargifySubscriptionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                        item["id"],
		"state":                     item["state"],
		"customer_id":               item["customer_id"],
		"product_id":                item["product_id"],
		"balance_in_cents":          item["balance_in_cents"],
		"total_revenue_in_cents":    item["total_revenue_in_cents"],
		"current_period_started_at": item["current_period_started_at"],
		"current_period_ends_at":    item["current_period_ends_at"],
		"created_at":                item["created_at"],
		"updated_at":                item["updated_at"],
	}
}

func chargifyProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"handle":            item["handle"],
		"description":       item["description"],
		"price_in_cents":    item["price_in_cents"],
		"interval":          item["interval"],
		"interval_unit":     item["interval_unit"],
		"product_family_id": item["product_family_id"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
	}
}

func chargifyCouponRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"code":              item["code"],
		"description":       item["description"],
		"percentage":        item["percentage"],
		"amount_in_cents":   item["amount_in_cents"],
		"product_family_id": item["product_family_id"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
	}
}

func chargifyTransactionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"transaction_type": item["transaction_type"],
		"amount_in_cents":  item["amount_in_cents"],
		"subscription_id":  item["subscription_id"],
		"customer_id":      item["customer_id"],
		"product_id":       item["product_id"],
		"success":          item["success"],
		"kind":             item["kind"],
		"created_at":       item["created_at"],
	}
}
