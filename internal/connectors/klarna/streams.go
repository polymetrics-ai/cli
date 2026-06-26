package klarna

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Klarna Settlements API resource path
// (relative to base_url) it reads from, the JSON path to the records array in
// the response body, and the record mapper that flattens each object.
type streamEndpoint struct {
	// resource is the API path segment, e.g. "settlements/v1/payouts".
	resource string
	// recordsPath is the dotted JSON path to the array of records in the body
	// (e.g. "payouts").
	recordsPath string
	// mapRecord flattens a raw Klarna object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// klarnaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in klarnaStreams; the read path
// is fully data-driven from this table.
//
// Klarna's Settlements API exposes payouts and the transactions that compose
// them. payout_summary is a convenience read of payouts grouped by reference.
var klarnaStreamEndpoints = map[string]streamEndpoint{
	"payouts":        {resource: "settlements/v1/payouts", recordsPath: "payouts", mapRecord: klarnaPayoutRecord},
	"transactions":   {resource: "settlements/v1/transactions", recordsPath: "transactions", mapRecord: klarnaTransactionRecord},
	"payout_summary": {resource: "settlements/v1/payouts", recordsPath: "payouts", mapRecord: klarnaPayoutSummaryRecord},
}

// klarnaStreams returns the connector's published stream catalog. Klarna's
// Settlements API only supports full-refresh, so no incremental cursor fields
// are published.
func klarnaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "payouts",
			Description: "Klarna settlement payouts.",
			PrimaryKey:  []string{"payout_reference"},
			Fields:      klarnaPayoutFields(),
		},
		{
			Name:        "transactions",
			Description: "Klarna settlement transactions composing payouts.",
			PrimaryKey:  []string{"transaction_id"},
			Fields:      klarnaTransactionFields(),
		},
		{
			Name:        "payout_summary",
			Description: "Klarna payout settlement totals, keyed by payout reference.",
			PrimaryKey:  []string{"payout_reference"},
			Fields:      klarnaPayoutSummaryFields(),
		},
	}
}

func klarnaPayoutFields() []connectors.Field {
	return []connectors.Field{
		{Name: "payout_reference", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "payment_reference", Type: "string"},
		{Name: "merchant_settlement_type", Type: "string"},
		{Name: "settlement_amount", Type: "integer"},
		{Name: "totals", Type: "object"},
	}
}

func klarnaTransactionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "transaction_id", Type: "string"},
		{Name: "payout_reference", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "amount", Type: "integer"},
		{Name: "currency_code", Type: "string"},
		{Name: "capture_id", Type: "string"},
		{Name: "order_id", Type: "string"},
		{Name: "short_order_id", Type: "string"},
		{Name: "merchant_reference1", Type: "string"},
		{Name: "merchant_reference2", Type: "string"},
		{Name: "sale_date", Type: "string"},
		{Name: "capture_date", Type: "string"},
	}
}

func klarnaPayoutSummaryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "payout_reference", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "settlement_amount", Type: "integer"},
		{Name: "sale_amount", Type: "integer"},
		{Name: "return_amount", Type: "integer"},
		{Name: "fee_amount", Type: "integer"},
	}
}

func klarnaPayoutRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"payout_reference":         item["payout_reference"],
		"currency_code":            item["currency_code"],
		"payment_reference":        item["payment_reference"],
		"merchant_settlement_type": item["merchant_settlement_type"],
		"totals":                   item["totals"],
	}
	if totals, ok := item["totals"].(map[string]any); ok {
		rec["settlement_amount"] = totals["settlement_amount"]
	}
	return rec
}

func klarnaTransactionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"transaction_id":      item["transaction_id"],
		"payout_reference":    item["payout_reference"],
		"type":                item["type"],
		"amount":              item["amount"],
		"currency_code":       item["currency_code"],
		"capture_id":          item["capture_id"],
		"order_id":            item["order_id"],
		"short_order_id":      item["short_order_id"],
		"merchant_reference1": item["merchant_reference1"],
		"merchant_reference2": item["merchant_reference2"],
		"sale_date":           item["sale_date"],
		"capture_date":        item["capture_date"],
	}
}

func klarnaPayoutSummaryRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"payout_reference": item["payout_reference"],
		"currency_code":    item["currency_code"],
	}
	if totals, ok := item["totals"].(map[string]any); ok {
		rec["settlement_amount"] = totals["settlement_amount"]
		rec["sale_amount"] = totals["sale_amount"]
		rec["return_amount"] = totals["return_amount"]
		rec["fee_amount"] = totals["fee_amount"]
	}
	return rec
}
