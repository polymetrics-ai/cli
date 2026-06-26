package lob

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Lob API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Lob list endpoint path segment (e.g. "postcards").
	resource string
	// mapRecord flattens a raw Lob object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// lobStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in lobStreams; the read path is
// fully data-driven from this table.
var lobStreamEndpoints = map[string]streamEndpoint{
	"addresses":     {resource: "addresses", mapRecord: lobAddressRecord},
	"postcards":     {resource: "postcards", mapRecord: lobMailpieceRecord},
	"letters":       {resource: "letters", mapRecord: lobMailpieceRecord},
	"checks":        {resource: "checks", mapRecord: lobMailpieceRecord},
	"bank_accounts": {resource: "bank_accounts", mapRecord: lobBankAccountRecord},
}

// lobStreams returns the connector's published stream catalog. Every Lob object
// exposes a string id and an ISO-8601 date_created timestamp, so the primary key
// is ["id"] and the incremental cursor field is ["date_created"] across the board.
func lobStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "addresses",
			Description:  "Lob address book entries.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_created"},
			Fields:       lobAddressFields(),
		},
		{
			Name:         "postcards",
			Description:  "Lob postcards.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_created"},
			Fields:       lobMailpieceFields(),
		},
		{
			Name:         "letters",
			Description:  "Lob letters.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_created"},
			Fields:       lobMailpieceFields(),
		},
		{
			Name:         "checks",
			Description:  "Lob checks.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_created"},
			Fields:       lobMailpieceFields(),
		},
		{
			Name:         "bank_accounts",
			Description:  "Lob bank accounts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_created"},
			Fields:       lobBankAccountFields(),
		},
	}
}

func lobAddressFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "address_line1", Type: "string"},
		{Name: "address_line2", Type: "string"},
		{Name: "address_city", Type: "string"},
		{Name: "address_state", Type: "string"},
		{Name: "address_zip", Type: "string"},
		{Name: "address_country", Type: "string"},
		{Name: "date_created", Type: "timestamp"},
		{Name: "date_modified", Type: "timestamp"},
		{Name: "deleted", Type: "boolean"},
	}
}

func lobMailpieceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "carrier", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "send_date", Type: "timestamp"},
		{Name: "expected_delivery_date", Type: "string"},
		{Name: "date_created", Type: "timestamp"},
		{Name: "date_modified", Type: "timestamp"},
		{Name: "deleted", Type: "boolean"},
	}
}

func lobBankAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "routing_number", Type: "string"},
		{Name: "account_number", Type: "string"},
		{Name: "account_type", Type: "string"},
		{Name: "signatory", Type: "string"},
		{Name: "bank_name", Type: "string"},
		{Name: "verified", Type: "boolean"},
		{Name: "date_created", Type: "timestamp"},
		{Name: "date_modified", Type: "timestamp"},
		{Name: "deleted", Type: "boolean"},
	}
}

func lobAddressRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"object":          item["object"],
		"description":     item["description"],
		"name":            item["name"],
		"company":         item["company"],
		"email":           item["email"],
		"phone":           item["phone"],
		"address_line1":   item["address_line1"],
		"address_line2":   item["address_line2"],
		"address_city":    item["address_city"],
		"address_state":   item["address_state"],
		"address_zip":     item["address_zip"],
		"address_country": item["address_country"],
		"date_created":    item["date_created"],
		"date_modified":   item["date_modified"],
		"deleted":         item["deleted"],
	}
}

func lobMailpieceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                     item["id"],
		"object":                 item["object"],
		"description":            item["description"],
		"url":                    item["url"],
		"carrier":                item["carrier"],
		"status":                 item["status"],
		"send_date":              item["send_date"],
		"expected_delivery_date": item["expected_delivery_date"],
		"date_created":           item["date_created"],
		"date_modified":          item["date_modified"],
		"deleted":                item["deleted"],
	}
}

func lobBankAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"object":         item["object"],
		"description":    item["description"],
		"routing_number": item["routing_number"],
		"account_number": item["account_number"],
		"account_type":   item["account_type"],
		"signatory":      item["signatory"],
		"bank_name":      item["bank_name"],
		"verified":       item["verified"],
		"date_created":   item["date_created"],
		"date_modified":  item["date_modified"],
		"deleted":        item["deleted"],
	}
}
