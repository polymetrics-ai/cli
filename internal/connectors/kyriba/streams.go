package kyriba

import "polymetrics.ai/internal/connectors"

type streamEndpoint struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"bank_accounts": {resource: "bank-accounts", mapRecord: bankAccountRecord},
	"transactions":  {resource: "transactions", mapRecord: transactionRecord},
	"statements":    {resource: "statements", mapRecord: statementRecord},
	"payments":      {resource: "payments", mapRecord: paymentRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "bank_accounts", Description: "Kyriba bank accounts.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "account_number", Type: "string"}, {Name: "currency", Type: "string"}, {Name: "status", Type: "string"}}},
		{Name: "transactions", Description: "Kyriba transactions.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "account_number", Type: "string"}, {Name: "amount", Type: "number"}, {Name: "currency", Type: "string"}, {Name: "status", Type: "string"}}},
		{Name: "statements", Description: "Kyriba statements.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "account_number", Type: "string"}, {Name: "currency", Type: "string"}, {Name: "status", Type: "string"}}},
		{Name: "payments", Description: "Kyriba payments exposed as read-only API records.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "amount", Type: "number"}, {Name: "currency", Type: "string"}, {Name: "status", Type: "string"}}},
	}
}

func bankAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "account_number": first(item, "accountNumber", "account_number"), "currency": item["currency"], "status": item["status"]}
}
func transactionRecord(item map[string]any) connectors.Record {
	rec := bankAccountRecord(item)
	rec["amount"] = item["amount"]
	return rec
}
func statementRecord(item map[string]any) connectors.Record { return bankAccountRecord(item) }
func paymentRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "amount": item["amount"], "currency": item["currency"], "status": item["status"]}
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if item[key] != nil {
			return item[key]
		}
	}
	return nil
}
