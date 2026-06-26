package ebayfinance

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the eBay Finances API resource path
// (relative to base_url), the JSON field that holds the records array, the
// stream's primary key field, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the path segment under the Finances v1 base, e.g. "transaction".
	resource string
	// recordsField is the JSON key whose value is the records array, e.g.
	// "transactions". Empty means the response body itself is a single object.
	recordsField string
	// pkField is the object's primary-key field name.
	pkField string
	// singleObject is true for endpoints that return one object (no array, no
	// pagination), e.g. seller_funds_summary.
	singleObject bool
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// ebayStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in ebayStreams; the read path
// is fully data-driven from this table.
var ebayStreamEndpoints = map[string]streamEndpoint{
	"transactions":         {resource: "transaction", recordsField: "transactions", pkField: "transactionId", mapRecord: ebayTransactionRecord},
	"payouts":              {resource: "payout", recordsField: "payouts", pkField: "payoutId", mapRecord: ebayPayoutRecord},
	"transfers":            {resource: "transfer", recordsField: "transfers", pkField: "transferId", mapRecord: ebayTransferRecord},
	"seller_funds_summary": {resource: "seller_funds_summary", recordsField: "", pkField: "", singleObject: true, mapRecord: ebaySellerFundsSummaryRecord},
}

// ebayStreams returns the connector's published stream catalog.
func ebayStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "transactions",
			Description:  "eBay monetary transactions (sales, refunds, credits, fees, etc.) from the Finances API.",
			PrimaryKey:   []string{"transactionId"},
			CursorFields: []string{"transactionDate"},
			Fields:       ebayTransactionFields(),
		},
		{
			Name:         "payouts",
			Description:  "Seller payouts settled to the seller's bank account.",
			PrimaryKey:   []string{"payoutId"},
			CursorFields: []string{"payoutDate"},
			Fields:       ebayPayoutFields(),
		},
		{
			Name:         "transfers",
			Description:  "Monies owed by the seller to eBay that were charged against the seller balance.",
			PrimaryKey:   []string{"transferId"},
			CursorFields: []string{"transferDate"},
			Fields:       ebayTransferFields(),
		},
		{
			Name:        "seller_funds_summary",
			Description: "Summary of funds that are processing, on hold, or available for payout.",
			PrimaryKey:  []string{},
			Fields:      ebaySellerFundsSummaryFields(),
		},
	}
}

func ebayTransactionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "transactionId", Type: "string"},
		{Name: "transactionType", Type: "string"},
		{Name: "transactionStatus", Type: "string"},
		{Name: "transactionDate", Type: "timestamp"},
		{Name: "amount_value", Type: "string"},
		{Name: "amount_currency", Type: "string"},
		{Name: "bookingEntry", Type: "string"},
		{Name: "orderId", Type: "string"},
		{Name: "salesRecordReference", Type: "string"},
		{Name: "payoutId", Type: "string"},
		{Name: "transactionMemo", Type: "string"},
		{Name: "feeType", Type: "string"},
	}
}

func ebayPayoutFields() []connectors.Field {
	return []connectors.Field{
		{Name: "payoutId", Type: "string"},
		{Name: "payoutStatus", Type: "string"},
		{Name: "payoutStatusDescription", Type: "string"},
		{Name: "payoutDate", Type: "timestamp"},
		{Name: "amount_value", Type: "string"},
		{Name: "amount_currency", Type: "string"},
		{Name: "transactionCount", Type: "integer"},
		{Name: "payoutInstrument_nickname", Type: "string"},
		{Name: "payoutInstrument_accountLastFourDigits", Type: "string"},
	}
}

func ebayTransferFields() []connectors.Field {
	return []connectors.Field{
		{Name: "transferId", Type: "string"},
		{Name: "transferStatus", Type: "string"},
		{Name: "transferType", Type: "string"},
		{Name: "transferDate", Type: "timestamp"},
		{Name: "amount_value", Type: "string"},
		{Name: "amount_currency", Type: "string"},
		{Name: "reason", Type: "string"},
	}
}

func ebaySellerFundsSummaryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "totalFunds_value", Type: "string"},
		{Name: "totalFunds_currency", Type: "string"},
		{Name: "availableFunds_value", Type: "string"},
		{Name: "availableFunds_currency", Type: "string"},
		{Name: "fundsOnHold_value", Type: "string"},
		{Name: "fundsOnHold_currency", Type: "string"},
		{Name: "processingFunds_value", Type: "string"},
		{Name: "processingFunds_currency", Type: "string"},
	}
}

func ebayTransactionRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"transactionId":        item["transactionId"],
		"transactionType":      item["transactionType"],
		"transactionStatus":    item["transactionStatus"],
		"transactionDate":      item["transactionDate"],
		"bookingEntry":         item["bookingEntry"],
		"orderId":              item["orderId"],
		"salesRecordReference": item["salesRecordReference"],
		"payoutId":             item["payoutId"],
		"transactionMemo":      item["transactionMemo"],
		"feeType":              item["feeType"],
	}
	flattenAmount(rec, "amount", item["amount"])
	return rec
}

func ebayPayoutRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"payoutId":                item["payoutId"],
		"payoutStatus":            item["payoutStatus"],
		"payoutStatusDescription": item["payoutStatusDescription"],
		"payoutDate":              item["payoutDate"],
		"transactionCount":        item["transactionCount"],
	}
	flattenAmount(rec, "amount", item["amount"])
	if pi, ok := item["payoutInstrument"].(map[string]any); ok {
		rec["payoutInstrument_nickname"] = pi["nickname"]
		rec["payoutInstrument_accountLastFourDigits"] = pi["accountLastFourDigits"]
	}
	return rec
}

func ebayTransferRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"transferId":     item["transferId"],
		"transferStatus": item["transferStatus"],
		"transferType":   item["transferType"],
		"transferDate":   item["transferDate"],
		"reason":         item["reason"],
	}
	flattenAmount(rec, "amount", item["amount"])
	return rec
}

func ebaySellerFundsSummaryRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{}
	for _, key := range []string{"totalFunds", "availableFunds", "fundsOnHold", "processingFunds"} {
		flattenAmount(rec, key, item[key])
	}
	return rec
}

// flattenAmount lifts an eBay {value, currency} money object into two
// underscore-prefixed scalar fields, e.g. amount -> amount_value, amount_currency.
func flattenAmount(rec connectors.Record, prefix string, raw any) {
	obj, ok := raw.(map[string]any)
	if !ok {
		return
	}
	if v, ok := obj["value"]; ok {
		rec[prefix+"_value"] = v
	}
	if v, ok := obj["currency"]; ok {
		rec[prefix+"_currency"] = v
	}
}
