package chargedesk

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the ChargeDesk API resource path (relative
// to base_url) it reads from, the record mapper that flattens its objects, and
// the field that uniquely identifies a record (used by fixture mode).
type streamEndpoint struct {
	// resource is the ChargeDesk list endpoint path segment (e.g. "charges").
	resource string
	// idField is the object's primary-key field name (e.g. "charge_id").
	idField string
	// mapRecord flattens a raw ChargeDesk object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// chargedeskStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in chargedeskStreams; the
// read path is fully data-driven from this table.
var chargedeskStreamEndpoints = map[string]streamEndpoint{
	"charges":       {resource: "charges", idField: "charge_id", mapRecord: chargedeskChargeRecord},
	"customers":     {resource: "customers", idField: "customer_id", mapRecord: chargedeskCustomerRecord},
	"subscriptions": {resource: "subscriptions", idField: "subscription_id", mapRecord: chargedeskSubscriptionRecord},
	"products":      {resource: "products", idField: "product_id", mapRecord: chargedeskProductRecord},
}

// chargedeskStreams returns the connector's published stream catalog. ChargeDesk
// list objects carry a unix `occurred` timestamp for charges/subscriptions; the
// primary key differs per resource. Only full-refresh is published upstream, but
// occurred is exposed as a cursor field where present for incremental clients.
func chargedeskStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "charges",
			Description:  "ChargeDesk charges (individual transactions and payment records).",
			PrimaryKey:   []string{"charge_id"},
			CursorFields: []string{"occurred"},
			Fields:       chargedeskChargeFields(),
		},
		{
			Name:         "customers",
			Description:  "ChargeDesk customers (billing entities).",
			PrimaryKey:   []string{"customer_id"},
			CursorFields: []string{"occurred"},
			Fields:       chargedeskCustomerFields(),
		},
		{
			Name:         "subscriptions",
			Description:  "ChargeDesk subscriptions (recurring billing arrangements).",
			PrimaryKey:   []string{"subscription_id"},
			CursorFields: []string{"occurred"},
			Fields:       chargedeskSubscriptionFields(),
		},
		{
			Name:         "products",
			Description:  "ChargeDesk products (groupings of related charges).",
			PrimaryKey:   []string{"product_id"},
			CursorFields: []string{"occurred"},
			Fields:       chargedeskProductFields(),
		},
	}
}

func chargedeskChargeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "charge_id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "occurred", Type: "integer"},
		{Name: "amount", Type: "string"},
		{Name: "amount_refunded", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "customer_email", Type: "string"},
		{Name: "customer_name", Type: "string"},
		{Name: "transaction_id", Type: "string"},
		{Name: "subscription_id", Type: "string"},
		{Name: "product_id", Type: "string"},
		{Name: "payment_method", Type: "string"},
	}
}

func chargedeskCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "customer_id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "occurred", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "delinquent", Type: "boolean"},
		{Name: "tax_number", Type: "string"},
	}
}

func chargedeskSubscriptionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "subscription_id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "occurred", Type: "integer"},
		{Name: "customer_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "interval", Type: "string"},
		{Name: "current_period_start", Type: "integer"},
		{Name: "current_period_end", Type: "integer"},
		{Name: "product_id", Type: "string"},
	}
}

func chargedeskProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "product_id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "occurred", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "interval", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func chargedeskChargeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"charge_id":       item["charge_id"],
		"object":          item["object"],
		"occurred":        item["occurred"],
		"amount":          item["amount"],
		"amount_refunded": item["amount_refunded"],
		"currency":        item["currency"],
		"status":          item["status"],
		"description":     item["description"],
		"customer_id":     item["customer_id"],
		"customer_email":  item["customer_email"],
		"customer_name":   item["customer_name"],
		"transaction_id":  item["transaction_id"],
		"subscription_id": item["subscription_id"],
		"product_id":      item["product_id"],
		"payment_method":  item["payment_method"],
	}
}

func chargedeskCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"customer_id": item["customer_id"],
		"object":      item["object"],
		"occurred":    item["occurred"],
		"email":       item["email"],
		"name":        item["name"],
		"country":     item["country"],
		"phone":       item["phone"],
		"currency":    item["currency"],
		"delinquent":  item["delinquent"],
		"tax_number":  item["tax_number"],
	}
}

func chargedeskSubscriptionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"subscription_id":      item["subscription_id"],
		"object":               item["object"],
		"occurred":             item["occurred"],
		"customer_id":          item["customer_id"],
		"status":               item["status"],
		"amount":               item["amount"],
		"currency":             item["currency"],
		"interval":             item["interval"],
		"current_period_start": item["current_period_start"],
		"current_period_end":   item["current_period_end"],
		"product_id":           item["product_id"],
	}
}

func chargedeskProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"product_id": item["product_id"],
		"object":     item["object"],
		"occurred":   item["occurred"],
		"name":       item["name"],
		"amount":     item["amount"],
		"currency":   item["currency"],
		"interval":   item["interval"],
		"status":     item["status"],
	}
}
