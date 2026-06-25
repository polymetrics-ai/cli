package stripe

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Stripe API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Stripe list endpoint path segment (e.g. "customers").
	resource string
	// mapRecord flattens a raw Stripe object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// stripeStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in stripeStreams; the read path
// is fully data-driven from this table.
var stripeStreamEndpoints = map[string]streamEndpoint{
	"customers":     {resource: "customers", mapRecord: stripeCustomerRecord},
	"charges":       {resource: "charges", mapRecord: stripeChargeRecord},
	"invoices":      {resource: "invoices", mapRecord: stripeInvoiceRecord},
	"subscriptions": {resource: "subscriptions", mapRecord: stripeSubscriptionRecord},
	"products":      {resource: "products", mapRecord: stripeProductRecord},
}

// stripeStreams returns the connector's published stream catalog. Every Stripe
// object exposes a string id and a unix `created` timestamp, so the primary key
// is ["id"] and the incremental cursor field is ["created"] across the board.
func stripeStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "customers",
			Description:  "Stripe customers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       stripeCustomerFields(),
		},
		{
			Name:         "charges",
			Description:  "Stripe charges.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       stripeChargeFields(),
		},
		{
			Name:         "invoices",
			Description:  "Stripe invoices.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       stripeInvoiceFields(),
		},
		{
			Name:         "subscriptions",
			Description:  "Stripe subscriptions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       stripeSubscriptionFields(),
		},
		{
			Name:         "products",
			Description:  "Stripe products.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       stripeProductFields(),
		},
	}
}

func stripeCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "created", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "balance", Type: "integer"},
		{Name: "delinquent", Type: "boolean"},
		{Name: "livemode", Type: "boolean"},
	}
}

func stripeChargeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "created", Type: "integer"},
		{Name: "amount", Type: "integer"},
		{Name: "amount_captured", Type: "integer"},
		{Name: "amount_refunded", Type: "integer"},
		{Name: "currency", Type: "string"},
		{Name: "customer", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "paid", Type: "boolean"},
		{Name: "refunded", Type: "boolean"},
		{Name: "livemode", Type: "boolean"},
	}
}

func stripeInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "created", Type: "integer"},
		{Name: "customer", Type: "string"},
		{Name: "subscription", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "amount_due", Type: "integer"},
		{Name: "amount_paid", Type: "integer"},
		{Name: "amount_remaining", Type: "integer"},
		{Name: "total", Type: "integer"},
		{Name: "paid", Type: "boolean"},
		{Name: "livemode", Type: "boolean"},
	}
}

func stripeSubscriptionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "created", Type: "integer"},
		{Name: "customer", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "current_period_start", Type: "integer"},
		{Name: "current_period_end", Type: "integer"},
		{Name: "cancel_at_period_end", Type: "boolean"},
		{Name: "canceled_at", Type: "integer"},
		{Name: "livemode", Type: "boolean"},
	}
}

func stripeProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "created", Type: "integer"},
		{Name: "updated", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "type", Type: "string"},
		{Name: "livemode", Type: "boolean"},
	}
}

func stripeCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"object":      item["object"],
		"created":     item["created"],
		"email":       item["email"],
		"name":        item["name"],
		"description": item["description"],
		"phone":       item["phone"],
		"currency":    item["currency"],
		"balance":     item["balance"],
		"delinquent":  item["delinquent"],
		"livemode":    item["livemode"],
	}
}

func stripeChargeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"object":          item["object"],
		"created":         item["created"],
		"amount":          item["amount"],
		"amount_captured": item["amount_captured"],
		"amount_refunded": item["amount_refunded"],
		"currency":        item["currency"],
		"customer":        item["customer"],
		"status":          item["status"],
		"paid":            item["paid"],
		"refunded":        item["refunded"],
		"livemode":        item["livemode"],
	}
}

func stripeInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"object":           item["object"],
		"created":          item["created"],
		"customer":         item["customer"],
		"subscription":     item["subscription"],
		"status":           item["status"],
		"currency":         item["currency"],
		"amount_due":       item["amount_due"],
		"amount_paid":      item["amount_paid"],
		"amount_remaining": item["amount_remaining"],
		"total":            item["total"],
		"paid":             item["paid"],
		"livemode":         item["livemode"],
	}
}

func stripeSubscriptionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["id"],
		"object":               item["object"],
		"created":              item["created"],
		"customer":             item["customer"],
		"status":               item["status"],
		"currency":             item["currency"],
		"current_period_start": item["current_period_start"],
		"current_period_end":   item["current_period_end"],
		"cancel_at_period_end": item["cancel_at_period_end"],
		"canceled_at":          item["canceled_at"],
		"livemode":             item["livemode"],
	}
}

func stripeProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"object":      item["object"],
		"created":     item["created"],
		"updated":     item["updated"],
		"name":        item["name"],
		"description": item["description"],
		"active":      item["active"],
		"type":        item["type"],
		"livemode":    item["livemode"],
	}
}
