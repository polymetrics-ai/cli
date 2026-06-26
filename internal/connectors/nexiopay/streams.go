package nexiopay

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Nexio API resource path (relative to
// base_url), the JSON path where its records live, whether it paginates, and the
// record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Nexio endpoint path segment (e.g. "card/v3").
	resource string
	// recordsPath is the dotted JSON path to the record array. "" selects the
	// response root (a top-level array).
	recordsPath string
	// paginated is true when the endpoint supports Nexio offset pagination.
	paginated bool
	// primaryKey is the stream's id field, also used to synthesize fixture ids.
	primaryKey string
	// mapRecord flattens a raw Nexio object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// nexioStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in nexioStreams; the read path
// is fully data-driven from this table.
var nexioStreamEndpoints = map[string]streamEndpoint{
	"card_tokens":   {resource: "card/v3", recordsPath: "rows", paginated: true, primaryKey: "key", mapRecord: cardTokenRecord},
	"recipients":    {resource: "payout/v3/recipient", recordsPath: "rows", paginated: true, primaryKey: "recipientId", mapRecord: recipientRecord},
	"spendbacks":    {resource: "payout/v3/spendback", recordsPath: "rows", paginated: true, primaryKey: "id", mapRecord: spendbackRecord},
	"payment_types": {resource: "transaction/v3/paymentTypes", recordsPath: "", paginated: false, primaryKey: "id", mapRecord: paymentTypeRecord},
	"terminal_list": {resource: "pay/v3/getTerminalList", recordsPath: "", paginated: false, primaryKey: "terminalId", mapRecord: terminalRecord},
	"user":          {resource: "user/v3/account/whoAmI", recordsPath: "", paginated: false, primaryKey: "accountId", mapRecord: userRecord},
}

// nexioStreams returns the connector's published stream catalog.
func nexioStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "card_tokens",
			Description: "Stored card tokens (vaulted cards) for the account.",
			PrimaryKey:  []string{"key"},
			Fields:      cardTokenFields(),
		},
		{
			Name:        "recipients",
			Description: "Payout recipients configured for the account.",
			PrimaryKey:  []string{"recipientId"},
			Fields:      recipientFields(),
		},
		{
			Name:        "spendbacks",
			Description: "Spendback payout records.",
			PrimaryKey:  []string{"id"},
			Fields:      spendbackFields(),
		},
		{
			Name:        "payment_types",
			Description: "Payment type definitions available to the account.",
			PrimaryKey:  []string{"id"},
			Fields:      paymentTypeFields(),
		},
		{
			Name:        "terminal_list",
			Description: "Terminals registered to the account.",
			PrimaryKey:  []string{"terminalId"},
			Fields:      terminalFields(),
		},
		{
			Name:        "user",
			Description: "The authenticated API user / account (whoAmI).",
			PrimaryKey:  []string{"accountId"},
			Fields:      userFields(),
		},
	}
}

func cardTokenFields() []connectors.Field {
	return []connectors.Field{
		{Name: "key", Type: "string"},
		{Name: "cardType", Type: "string"},
		{Name: "lastFour", Type: "string"},
		{Name: "expirationMonth", Type: "string"},
		{Name: "expirationYear", Type: "string"},
		{Name: "cardHolderName", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "createdDate", Type: "string"},
	}
}

func recipientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "recipientId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "createdDate", Type: "string"},
		{Name: "updatedDate", Type: "string"},
	}
}

func spendbackFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "recipientId", Type: "string"},
		{Name: "amount", Type: "number"},
		{Name: "currency", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "createdDate", Type: "string"},
	}
}

func paymentTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "displayName", Type: "string"},
		{Name: "enabled", Type: "boolean"},
	}
}

func terminalFields() []connectors.Field {
	return []connectors.Field{
		{Name: "terminalId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "merchantId", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "accountId", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "merchantId", Type: "string"},
		{Name: "role", Type: "string"},
	}
}

func cardTokenRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"key":             item["key"],
		"cardType":        item["cardType"],
		"lastFour":        item["lastFour"],
		"expirationMonth": item["expirationMonth"],
		"expirationYear":  item["expirationYear"],
		"cardHolderName":  item["cardHolderName"],
		"currency":        item["currency"],
		"createdDate":     item["createdDate"],
	}
}

func recipientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"recipientId": item["recipientId"],
		"name":        item["name"],
		"email":       item["email"],
		"status":      item["status"],
		"currency":    item["currency"],
		"createdDate": item["createdDate"],
		"updatedDate": item["updatedDate"],
	}
}

func spendbackRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"recipientId": item["recipientId"],
		"amount":      item["amount"],
		"currency":    item["currency"],
		"status":      item["status"],
		"createdDate": item["createdDate"],
	}
}

func paymentTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"displayName": item["displayName"],
		"enabled":     item["enabled"],
	}
}

func terminalRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"terminalId": item["terminalId"],
		"name":       item["name"],
		"merchantId": item["merchantId"],
		"status":     item["status"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"accountId":  item["accountId"],
		"username":   item["username"],
		"email":      item["email"],
		"merchantId": item["merchantId"],
		"role":       item["role"],
	}
}
