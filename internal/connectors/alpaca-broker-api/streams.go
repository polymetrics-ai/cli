package alpacabrokerapi

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Alpaca Broker API resource path
// (relative to base_url) it reads from, the record mapper that flattens its
// objects, and whether the endpoint returns a single object (clock) rather than
// an array.
type streamEndpoint struct {
	// resource is the Broker API path segment (e.g. "accounts").
	resource string
	// mapRecord flattens a raw Alpaca object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// singleton is true for endpoints that return one object (no pagination),
	// e.g. /clock.
	singleton bool
	// paginates is true for endpoints that support page_token cursor paging
	// over a top-level array (accounts). calendar/assets/country_info return the
	// full list in one response.
	paginates bool
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams; the read path is fully
// data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"accounts":     {resource: "accounts", mapRecord: accountRecord, paginates: true},
	"assets":       {resource: "assets", mapRecord: assetRecord},
	"calendar":     {resource: "calendar", mapRecord: calendarRecord},
	"clock":        {resource: "clock", mapRecord: clockRecord, singleton: true},
	"country_info": {resource: "country_info", mapRecord: countryInfoRecord},
}

// streams returns the connector's published stream catalog.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "accounts",
			Description:  "Alpaca Broker API trading accounts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       accountFields(),
		},
		{
			Name:        "assets",
			Description: "Tradable stock and crypto assets.",
			PrimaryKey:  []string{"id"},
			Fields:      assetFields(),
		},
		{
			Name:        "calendar",
			Description: "Market trading calendar (open/close days and times).",
			PrimaryKey:  []string{"date"},
			Fields:      calendarFields(),
		},
		{
			Name:        "clock",
			Description: "Current market clock status.",
			PrimaryKey:  []string{"timestamp"},
			Fields:      clockFields(),
		},
		{
			Name:        "country_info",
			Description: "Supported countries and their attributes.",
			PrimaryKey:  []string{"country_code"},
			Fields:      countryInfoFields(),
		},
	}
}

func accountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "account_number", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "crypto_status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "account_type", Type: "string"},
		{Name: "enabled_assets", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "kyc_results", Type: "string"},
		{Name: "last_equity", Type: "string"},
	}
}

func assetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "class", Type: "string"},
		{Name: "exchange", Type: "string"},
		{Name: "symbol", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "tradable", Type: "boolean"},
		{Name: "marginable", Type: "boolean"},
		{Name: "shortable", Type: "boolean"},
		{Name: "easy_to_borrow", Type: "boolean"},
		{Name: "fractionable", Type: "boolean"},
	}
}

func calendarFields() []connectors.Field {
	return []connectors.Field{
		{Name: "date", Type: "string"},
		{Name: "open", Type: "string"},
		{Name: "close", Type: "string"},
		{Name: "session_open", Type: "string"},
		{Name: "session_close", Type: "string"},
	}
}

func clockFields() []connectors.Field {
	return []connectors.Field{
		{Name: "timestamp", Type: "string"},
		{Name: "is_open", Type: "boolean"},
		{Name: "next_open", Type: "string"},
		{Name: "next_close", Type: "string"},
	}
}

func countryInfoFields() []connectors.Field {
	return []connectors.Field{
		{Name: "country_code", Type: "string"},
		{Name: "country_name", Type: "string"},
		{Name: "phone_calling_code", Type: "string"},
	}
}

func accountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"account_number": item["account_number"],
		"status":         item["status"],
		"crypto_status":  item["crypto_status"],
		"currency":       item["currency"],
		"account_type":   item["account_type"],
		"enabled_assets": item["enabled_assets"],
		"created_at":     item["created_at"],
		"kyc_results":    item["kyc_results"],
		"last_equity":    item["last_equity"],
	}
}

func assetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"class":          item["class"],
		"exchange":       item["exchange"],
		"symbol":         item["symbol"],
		"name":           item["name"],
		"status":         item["status"],
		"tradable":       item["tradable"],
		"marginable":     item["marginable"],
		"shortable":      item["shortable"],
		"easy_to_borrow": item["easy_to_borrow"],
		"fractionable":   item["fractionable"],
	}
}

func calendarRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"date":          item["date"],
		"open":          item["open"],
		"close":         item["close"],
		"session_open":  item["session_open"],
		"session_close": item["session_close"],
	}
}

func clockRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"timestamp":  item["timestamp"],
		"is_open":    item["is_open"],
		"next_open":  item["next_open"],
		"next_close": item["next_close"],
	}
}

func countryInfoRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"country_code":       item["country_code"],
		"country_name":       item["country_name"],
		"phone_calling_code": item["phone_calling_code"],
	}
}
