package gocardless

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the GoCardless API resource it reads from.
// GoCardless list responses key the record array by the resource name (e.g.
// {"payments":[...]}), so resource doubles as the records JSON path.
type streamEndpoint struct {
	// resource is the GoCardless list endpoint path segment and the JSON key
	// under which the record array is returned (e.g. "payments").
	resource string
	// mapRecord flattens a raw GoCardless object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// gocardlessStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in gocardlessStreams; the
// read path is fully data-driven from this table. The core set mirrors the
// upstream upstream source: payments, mandates, payouts, refunds.
var gocardlessStreamEndpoints = map[string]streamEndpoint{
	"payments": {resource: "payments", mapRecord: gocardlessPaymentRecord},
	"mandates": {resource: "mandates", mapRecord: gocardlessMandateRecord},
	"payouts":  {resource: "payouts", mapRecord: gocardlessPayoutRecord},
	"refunds":  {resource: "refunds", mapRecord: gocardlessRefundRecord},
}

// gocardlessStreams returns the connector's published stream catalog. Every
// GoCardless resource exposes a string id and an RFC3339 created_at timestamp,
// so the primary key is ["id"] and the cursor field is ["created_at"].
func gocardlessStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "payments",
			Description:  "GoCardless payments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       gocardlessPaymentFields(),
		},
		{
			Name:         "mandates",
			Description:  "GoCardless mandates.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       gocardlessMandateFields(),
		},
		{
			Name:         "payouts",
			Description:  "GoCardless payouts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       gocardlessPayoutFields(),
		},
		{
			Name:         "refunds",
			Description:  "GoCardless refunds.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       gocardlessRefundFields(),
		},
	}
}

func gocardlessPaymentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "charge_date", Type: "string"},
		{Name: "amount", Type: "integer"},
		{Name: "amount_refunded", Type: "integer"},
		{Name: "currency", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "reference", Type: "string"},
		{Name: "mandate", Type: "string"},
		{Name: "payout", Type: "string"},
	}
}

func gocardlessMandateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "reference", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "scheme", Type: "string"},
		{Name: "next_possible_charge_date", Type: "string"},
		{Name: "payments_require_approval", Type: "boolean"},
		{Name: "customer_bank_account", Type: "string"},
		{Name: "creditor", Type: "string"},
	}
}

func gocardlessPayoutFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "amount", Type: "integer"},
		{Name: "deducted_fees", Type: "integer"},
		{Name: "currency", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "payout_type", Type: "string"},
		{Name: "reference", Type: "string"},
		{Name: "arrival_date", Type: "string"},
		{Name: "creditor", Type: "string"},
		{Name: "creditor_bank_account", Type: "string"},
	}
}

func gocardlessRefundFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "amount", Type: "integer"},
		{Name: "currency", Type: "string"},
		{Name: "reference", Type: "string"},
		{Name: "payment", Type: "string"},
		{Name: "mandate", Type: "string"},
	}
}

func gocardlessPaymentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"created_at":      item["created_at"],
		"charge_date":     item["charge_date"],
		"amount":          item["amount"],
		"amount_refunded": item["amount_refunded"],
		"currency":        item["currency"],
		"status":          item["status"],
		"description":     item["description"],
		"reference":       item["reference"],
		"mandate":         linkField(item, "mandate"),
		"payout":          linkField(item, "payout"),
	}
}

func gocardlessMandateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                        item["id"],
		"created_at":                item["created_at"],
		"reference":                 item["reference"],
		"status":                    item["status"],
		"scheme":                    item["scheme"],
		"next_possible_charge_date": item["next_possible_charge_date"],
		"payments_require_approval": item["payments_require_approval"],
		"customer_bank_account":     linkField(item, "customer_bank_account"),
		"creditor":                  linkField(item, "creditor"),
	}
}

func gocardlessPayoutRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["id"],
		"created_at":            item["created_at"],
		"amount":                item["amount"],
		"deducted_fees":         item["deducted_fees"],
		"currency":              item["currency"],
		"status":                item["status"],
		"payout_type":           item["payout_type"],
		"reference":             item["reference"],
		"arrival_date":          item["arrival_date"],
		"creditor":              linkField(item, "creditor"),
		"creditor_bank_account": linkField(item, "creditor_bank_account"),
	}
}

func gocardlessRefundRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"created_at": item["created_at"],
		"amount":     item["amount"],
		"currency":   item["currency"],
		"reference":  item["reference"],
		"payment":    linkField(item, "payment"),
		"mandate":    linkField(item, "mandate"),
	}
}

// linkField flattens a GoCardless relationship. The API nests related resource
// ids under a "links" object (e.g. {"links":{"mandate":"MD123"}}); when present
// the linked id is preferred, otherwise a top-level value is returned.
func linkField(item map[string]any, key string) any {
	if links, ok := item["links"].(map[string]any); ok {
		if v, ok := links[key]; ok {
			return v
		}
	}
	return item[key]
}
