package paypaltransaction

import "polymetrics.ai/internal/connectors"

// pagination identifies how a PayPal endpoint exposes additional pages so the
// read loop knows which paginator to drive.
type pagination int

const (
	// paginationNone is a single-shot endpoint (e.g. balances).
	paginationNone pagination = iota
	// paginationPageIncrement walks page=1..total_pages (transactions, products).
	paginationPageIncrement
	// paginationPageToken follows a links[rel=next] href (disputes).
	paginationPageToken
)

// streamEndpoint maps a stream name to the PayPal API resource path (relative to
// base_url), the JSON path to its records array, the record mapper that flattens
// each item, and the pagination style.
type streamEndpoint struct {
	resource    string
	recordsPath string
	pagination  pagination
	// dateRange is true when the endpoint requires start_date/end_date query
	// params (the reporting endpoints do; catalog/dispute listings do not).
	dateRange bool
	mapRecord func(map[string]any) connectors.Record
}

// paypalStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in paypalStreams; the read path
// is fully data-driven from this table.
var paypalStreamEndpoints = map[string]streamEndpoint{
	"transactions": {
		resource:    "v1/reporting/transactions",
		recordsPath: "transaction_details",
		pagination:  paginationPageIncrement,
		dateRange:   true,
		mapRecord:   transactionRecord,
	},
	"balances": {
		resource:    "v1/reporting/balances",
		recordsPath: "balances",
		pagination:  paginationNone,
		dateRange:   false,
		mapRecord:   balanceRecord,
	},
	"products": {
		resource:    "v1/catalogs/products",
		recordsPath: "products",
		pagination:  paginationPageIncrement,
		dateRange:   false,
		mapRecord:   productRecord,
	},
	"disputes": {
		resource:    "v1/customer/disputes",
		recordsPath: "items",
		pagination:  paginationPageToken,
		dateRange:   false,
		mapRecord:   disputeRecord,
	},
}

// paypalStreams returns the connector's published stream catalog.
func paypalStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "transactions",
			Description:  "PayPal transaction history from the reporting transactions endpoint.",
			PrimaryKey:   []string{"transaction_id"},
			CursorFields: []string{"transaction_initiation_date"},
			Fields:       transactionFields(),
		},
		{
			Name:         "balances",
			Description:  "PayPal account balances by currency.",
			PrimaryKey:   []string{"currency"},
			CursorFields: []string{"as_of_time"},
			Fields:       balanceFields(),
		},
		{
			Name:        "products",
			Description: "PayPal catalog products.",
			PrimaryKey:  []string{"id"},
			Fields:      productFields(),
		},
		{
			Name:         "disputes",
			Description:  "PayPal customer disputes.",
			PrimaryKey:   []string{"dispute_id"},
			CursorFields: []string{"update_time"},
			Fields:       disputeFields(),
		},
	}
}

func transactionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "transaction_id", Type: "string"},
		{Name: "transaction_status", Type: "string"},
		{Name: "transaction_event_code", Type: "string"},
		{Name: "transaction_initiation_date", Type: "string"},
		{Name: "transaction_updated_date", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "fee_amount", Type: "string"},
		{Name: "paypal_account_id", Type: "string"},
	}
}

func balanceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "currency", Type: "string"},
		{Name: "primary", Type: "boolean"},
		{Name: "total_currency_code", Type: "string"},
		{Name: "total_value", Type: "string"},
		{Name: "available_value", Type: "string"},
		{Name: "withheld_value", Type: "string"},
	}
}

func productFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "create_time", Type: "string"},
	}
}

func disputeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "dispute_id", Type: "string"},
		{Name: "reason", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "dispute_state", Type: "string"},
		{Name: "dispute_amount_currency_code", Type: "string"},
		{Name: "dispute_amount_value", Type: "string"},
		{Name: "create_time", Type: "string"},
		{Name: "update_time", Type: "string"},
	}
}

// transactionRecord flattens a {transaction_info:{...}} item. PayPal nests the
// monetary value under transaction_amount; the rest are scalar fields.
func transactionRecord(item map[string]any) connectors.Record {
	info := nestedObject(item, "transaction_info")
	amount := nestedObject(info, "transaction_amount")
	fee := nestedObject(info, "fee_amount")
	return connectors.Record{
		"transaction_id":              info["transaction_id"],
		"transaction_status":          info["transaction_status"],
		"transaction_event_code":      info["transaction_event_code"],
		"transaction_initiation_date": info["transaction_initiation_date"],
		"transaction_updated_date":    info["transaction_updated_date"],
		"currency_code":               amount["currency_code"],
		"amount":                      amount["value"],
		"fee_amount":                  fee["value"],
		"paypal_account_id":           info["paypal_account_id"],
	}
}

func balanceRecord(item map[string]any) connectors.Record {
	total := nestedObject(item, "total_balance")
	available := nestedObject(item, "available_balance")
	withheld := nestedObject(item, "withheld_balance")
	return connectors.Record{
		"currency":            item["currency"],
		"primary":             item["primary"],
		"total_currency_code": total["currency_code"],
		"total_value":         total["value"],
		"available_value":     available["value"],
		"withheld_value":      withheld["value"],
	}
}

func productRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"type":        item["type"],
		"category":    item["category"],
		"create_time": item["create_time"],
	}
}

func disputeRecord(item map[string]any) connectors.Record {
	amount := nestedObject(item, "dispute_amount")
	return connectors.Record{
		"dispute_id":                   item["dispute_id"],
		"reason":                       item["reason"],
		"status":                       item["status"],
		"dispute_state":                item["dispute_state"],
		"dispute_amount_currency_code": amount["currency_code"],
		"dispute_amount_value":         amount["value"],
		"create_time":                  item["create_time"],
		"update_time":                  item["update_time"],
	}
}

// nestedObject returns item[key] as a map, or an empty (non-nil) map when the
// key is absent or not an object, so field lookups stay nil-safe.
func nestedObject(item map[string]any, key string) map[string]any {
	if item == nil {
		return map[string]any{}
	}
	if obj, ok := item[key].(map[string]any); ok {
		return obj
	}
	return map[string]any{}
}
