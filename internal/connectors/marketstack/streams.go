package marketstack

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Marketstack API resource path
// (relative to base_url) it reads from, the record mapper that flattens its
// objects, and whether the endpoint accepts the `symbols` filter.
type streamEndpoint struct {
	// resource is the Marketstack endpoint path segment (e.g. "eod").
	resource string
	// mapRecord flattens a raw Marketstack object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// acceptsSymbols reports whether the endpoint honors the symbols query
	// parameter (eod/splits/dividends do; exchanges/tickers do not).
	acceptsSymbols bool
}

// marketstackStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in marketstackStreams;
// the read path is fully data-driven from this table. Every endpoint returns its
// rows under the top-level "data" array and paginates with offset/limit.
var marketstackStreamEndpoints = map[string]streamEndpoint{
	"exchanges": {resource: "exchanges", mapRecord: exchangeRecord},
	"tickers":   {resource: "tickers", mapRecord: tickerRecord},
	"eod":       {resource: "eod", mapRecord: eodRecord, acceptsSymbols: true},
	"splits":    {resource: "splits", mapRecord: splitRecord, acceptsSymbols: true},
	"dividends": {resource: "dividends", mapRecord: dividendRecord, acceptsSymbols: true},
}

// marketstackStreams returns the connector's published stream catalog.
func marketstackStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "exchanges",
			Description: "Global stock exchanges Marketstack provides data for.",
			PrimaryKey:  []string{"mic"},
			Fields:      exchangeFields(),
		},
		{
			Name:        "tickers",
			Description: "Stock ticker symbols available on Marketstack.",
			PrimaryKey:  []string{"symbol"},
			Fields:      tickerFields(),
		},
		{
			Name:         "eod",
			Description:  "End-of-day historical stock prices.",
			PrimaryKey:   []string{"symbol", "date"},
			CursorFields: []string{"date"},
			Fields:       eodFields(),
		},
		{
			Name:         "splits",
			Description:  "Historical stock split events.",
			PrimaryKey:   []string{"symbol", "date"},
			CursorFields: []string{"date"},
			Fields:       splitFields(),
		},
		{
			Name:         "dividends",
			Description:  "Historical dividend payments.",
			PrimaryKey:   []string{"symbol", "date"},
			CursorFields: []string{"date"},
			Fields:       dividendFields(),
		},
	}
}

func exchangeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "mic", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "acronym", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "country_code", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "currency_name", Type: "string"},
		{Name: "currency_symbol", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "timezone_abbr", Type: "string"},
	}
}

func tickerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "has_eod", Type: "boolean"},
		{Name: "has_intraday", Type: "boolean"},
		{Name: "stock_exchange_mic", Type: "string"},
		{Name: "stock_exchange_name", Type: "string"},
	}
}

func eodFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "exchange", Type: "string"},
		{Name: "open", Type: "number"},
		{Name: "high", Type: "number"},
		{Name: "low", Type: "number"},
		{Name: "close", Type: "number"},
		{Name: "volume", Type: "number"},
		{Name: "adj_open", Type: "number"},
		{Name: "adj_high", Type: "number"},
		{Name: "adj_low", Type: "number"},
		{Name: "adj_close", Type: "number"},
		{Name: "adj_volume", Type: "number"},
		{Name: "split_factor", Type: "number"},
		{Name: "dividend", Type: "number"},
	}
}

func splitFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "split_factor", Type: "number"},
	}
}

func dividendFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "dividend", Type: "number"},
	}
}

func exchangeRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"mic":          item["mic"],
		"name":         item["name"],
		"acronym":      item["acronym"],
		"country":      item["country"],
		"country_code": item["country_code"],
		"city":         item["city"],
		"website":      item["website"],
	}
	if currency, ok := item["currency"].(map[string]any); ok {
		rec["currency_code"] = currency["code"]
		rec["currency_name"] = currency["name"]
		rec["currency_symbol"] = currency["symbol"]
	}
	if tz, ok := item["timezone"].(map[string]any); ok {
		rec["timezone"] = tz["timezone"]
		rec["timezone_abbr"] = tz["abbr"]
	}
	return rec
}

func tickerRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"symbol":       item["symbol"],
		"name":         item["name"],
		"has_eod":      item["has_eod"],
		"has_intraday": item["has_intraday"],
	}
	if ex, ok := item["stock_exchange"].(map[string]any); ok {
		rec["stock_exchange_mic"] = ex["mic"]
		rec["stock_exchange_name"] = ex["name"]
	}
	return rec
}

func eodRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"symbol":       item["symbol"],
		"date":         item["date"],
		"exchange":     item["exchange"],
		"open":         item["open"],
		"high":         item["high"],
		"low":          item["low"],
		"close":        item["close"],
		"volume":       item["volume"],
		"adj_open":     item["adj_open"],
		"adj_high":     item["adj_high"],
		"adj_low":      item["adj_low"],
		"adj_close":    item["adj_close"],
		"adj_volume":   item["adj_volume"],
		"split_factor": item["split_factor"],
		"dividend":     item["dividend"],
	}
}

func splitRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"symbol":       item["symbol"],
		"date":         item["date"],
		"split_factor": item["split_factor"],
	}
}

func dividendRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"symbol":   item["symbol"],
		"date":     item["date"],
		"dividend": item["dividend"],
	}
}
