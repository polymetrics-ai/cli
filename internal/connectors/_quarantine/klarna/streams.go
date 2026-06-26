package klarna

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to its Klarna Settlements API path, the JSON
// field that holds the records array in the response, and the record mapper.
type streamEndpoint struct {
	// path is the request path relative to the resolved base URL.
	path string
	// recordsPath is the dotted JSON path to the records array in the body.
	recordsPath string
	// mapRecord flattens a raw Klarna object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// klarnaStreamEndpoints is the per-stream routing table. The Klarna Settlements
// API exposes exactly two list resources: payouts and the transactions that make
// them up. Adding a stream means adding one entry here plus a Stream definition.
var klarnaStreamEndpoints = map[string]streamEndpoint{
	"payouts":      {path: "/settlements/v1/payouts", recordsPath: "payouts", mapRecord: klarnaPayoutRecord},
	"transactions": {path: "/settlements/v1/transactions", recordsPath: "transactions", mapRecord: klarnaTransactionRecord},
}

// klarnaStreams returns the connector's published stream catalog.
//
// Payouts are keyed by payment_reference (the payout id) and carry an ISO-8601
// payout_date suitable as an incremental cursor. Transactions are keyed by
// capture_id and carry sale_date / capture_date timestamps.
func klarnaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "payouts",
			Description:  "Klarna Settlements payouts: money Klarna has paid out to the merchant.",
			PrimaryKey:   []string{"payment_reference"},
			CursorFields: []string{"payout_date"},
			Fields:       klarnaPayoutFields(),
		},
		{
			Name:         "transactions",
			Description:  "Klarna Settlements transactions: the individual sales, fees, returns, and corrections that make up payouts.",
			PrimaryKey:   []string{"capture_id"},
			CursorFields: []string{"capture_date"},
			Fields:       klarnaTransactionFields(),
		},
	}
}

func klarnaPayoutFields() []connectors.Field {
	return []connectors.Field{
		{Name: "payment_reference", Type: "string"},
		{Name: "payout_date", Type: "timestamp"},
		{Name: "currency_code", Type: "string"},
		{Name: "currency_code_of_registration_country", Type: "string"},
		{Name: "merchant_settlement_type", Type: "string"},
		{Name: "merchant_id", Type: "string"},
		{Name: "transactions", Type: "string"},
		{Name: "totals", Type: "object"},
	}
}

func klarnaTransactionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "capture_id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "detailed_type", Type: "string"},
		{Name: "amount", Type: "integer"},
		{Name: "currency_code", Type: "string"},
		{Name: "payment_reference", Type: "string"},
		{Name: "order_id", Type: "string"},
		{Name: "short_order_id", Type: "string"},
		{Name: "refund_id", Type: "string"},
		{Name: "merchant_id", Type: "string"},
		{Name: "merchant_reference1", Type: "string"},
		{Name: "merchant_reference2", Type: "string"},
		{Name: "sale_date", Type: "timestamp"},
		{Name: "capture_date", Type: "timestamp"},
		{Name: "purchase_country", Type: "string"},
		{Name: "shipping_country", Type: "string"},
		{Name: "vat_rate", Type: "integer"},
		{Name: "vat_amount", Type: "integer"},
		{Name: "payout", Type: "string"},
	}
}

func klarnaPayoutRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"payment_reference":                     item["payment_reference"],
		"payout_date":                           item["payout_date"],
		"currency_code":                         item["currency_code"],
		"currency_code_of_registration_country": item["currency_code_of_registration_country"],
		"merchant_settlement_type":              item["merchant_settlement_type"],
		"merchant_id":                           item["merchant_id"],
		"transactions":                          item["transactions"],
		"totals":                                item["totals"],
	}
}

func klarnaTransactionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"capture_id":          item["capture_id"],
		"type":                item["type"],
		"detailed_type":       item["detailed_type"],
		"amount":              item["amount"],
		"currency_code":       item["currency_code"],
		"payment_reference":   item["payment_reference"],
		"order_id":            item["order_id"],
		"short_order_id":      item["short_order_id"],
		"refund_id":           item["refund_id"],
		"merchant_id":         item["merchant_id"],
		"merchant_reference1": item["merchant_reference1"],
		"merchant_reference2": item["merchant_reference2"],
		"sale_date":           item["sale_date"],
		"capture_date":        item["capture_date"],
		"purchase_country":    item["purchase_country"],
		"shipping_country":    item["shipping_country"],
		"vat_rate":            item["vat_rate"],
		"vat_amount":          item["vat_amount"],
		"payout":              item["payout"],
	}
}
