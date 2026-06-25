package paystack

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Paystack API resource path (relative
// to base_url) it lists from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Paystack list endpoint path segment (e.g. "customer").
	// Paystack list endpoints are singular nouns: /customer, /transaction.
	resource string
	// mapRecord flattens a raw Paystack object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// paystackStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in paystackStreams; the read
// path is fully data-driven from this table.
var paystackStreamEndpoints = map[string]streamEndpoint{
	"customers":     {resource: "customer", mapRecord: paystackCustomerRecord},
	"transactions":  {resource: "transaction", mapRecord: paystackTransactionRecord},
	"subscriptions": {resource: "subscription", mapRecord: paystackSubscriptionRecord},
	"invoices":      {resource: "paymentrequest", mapRecord: paystackInvoiceRecord},
	"disputes":      {resource: "dispute", mapRecord: paystackDisputeRecord},
}

// paystackStreams returns the connector's published stream catalog. Every
// Paystack object exposes an integer id and an ISO-8601 createdAt timestamp, so
// the primary key is ["id"] and the incremental cursor field is ["createdAt"].
func paystackStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "customers",
			Description:  "Paystack customers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       paystackCustomerFields(),
		},
		{
			Name:         "transactions",
			Description:  "Paystack transactions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       paystackTransactionFields(),
		},
		{
			Name:         "subscriptions",
			Description:  "Paystack subscriptions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       paystackSubscriptionFields(),
		},
		{
			Name:         "invoices",
			Description:  "Paystack invoices (payment requests).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       paystackInvoiceFields(),
		},
		{
			Name:         "disputes",
			Description:  "Paystack disputes.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       paystackDisputeFields(),
		},
	}
}

func paystackCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "customer_code", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "domain", Type: "string"},
		{Name: "risk_action", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func paystackTransactionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "reference", Type: "string"},
		{Name: "amount", Type: "integer"},
		{Name: "currency", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "domain", Type: "string"},
		{Name: "gateway_response", Type: "string"},
		{Name: "channel", Type: "string"},
		{Name: "paid_at", Type: "string"},
		{Name: "createdAt", Type: "string"},
	}
}

func paystackSubscriptionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "subscription_code", Type: "string"},
		{Name: "email_token", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "amount", Type: "integer"},
		{Name: "domain", Type: "string"},
		{Name: "next_payment_date", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func paystackInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "request_code", Type: "string"},
		{Name: "amount", Type: "integer"},
		{Name: "currency", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "paid", Type: "boolean"},
		{Name: "domain", Type: "string"},
		{Name: "due_date", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func paystackDisputeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "refund_amount", Type: "integer"},
		{Name: "currency", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "resolution", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "domain", Type: "string"},
		{Name: "due_at", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func paystackCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"customer_code": item["customer_code"],
		"email":         item["email"],
		"first_name":    item["first_name"],
		"last_name":     item["last_name"],
		"phone":         item["phone"],
		"domain":        item["domain"],
		"risk_action":   item["risk_action"],
		"createdAt":     item["createdAt"],
		"updatedAt":     item["updatedAt"],
	}
}

func paystackTransactionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"reference":        item["reference"],
		"amount":           item["amount"],
		"currency":         item["currency"],
		"status":           item["status"],
		"domain":           item["domain"],
		"gateway_response": item["gateway_response"],
		"channel":          item["channel"],
		"paid_at":          item["paid_at"],
		"createdAt":        item["createdAt"],
	}
}

func paystackSubscriptionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"subscription_code": item["subscription_code"],
		"email_token":       item["email_token"],
		"status":            item["status"],
		"amount":            item["amount"],
		"domain":            item["domain"],
		"next_payment_date": item["next_payment_date"],
		"createdAt":         item["createdAt"],
		"updatedAt":         item["updatedAt"],
	}
}

func paystackInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"request_code": item["request_code"],
		"amount":       item["amount"],
		"currency":     item["currency"],
		"status":       item["status"],
		"paid":         item["paid"],
		"domain":       item["domain"],
		"due_date":     item["due_date"],
		"createdAt":    item["createdAt"],
		"updatedAt":    item["updatedAt"],
	}
}

func paystackDisputeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"refund_amount": item["refund_amount"],
		"currency":      item["currency"],
		"status":        item["status"],
		"resolution":    item["resolution"],
		"category":      item["category"],
		"domain":        item["domain"],
		"due_at":        item["due_at"],
		"createdAt":     item["createdAt"],
		"updatedAt":     item["updatedAt"],
	}
}
