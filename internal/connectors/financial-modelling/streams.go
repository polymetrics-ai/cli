package financialmodelling

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Financial Modeling Prep API resource
// path (relative to base_url) it reads from, the record mapper that flattens its
// objects, and whether the endpoint supports limit/offset pagination.
type streamEndpoint struct {
	// resource is the FMP endpoint path segment (e.g. "stock/list").
	resource string
	// mapRecord flattens a raw FMP object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// paginated indicates the endpoint accepts limit/offset query params. Most
	// FMP list endpoints return the entire array in one response and ignore
	// pagination params; the screener honours limit/offset.
	paginated bool
}

// fmStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in fmStreams; the read path is
// fully data-driven from this table. The core set intentionally avoids the
// per-symbol partitioned endpoints (company_profile, historical prices) so each
// stream is a single list endpoint.
var fmStreamEndpoints = map[string]streamEndpoint{
	"stocks":             {resource: "stock/list", mapRecord: fmSymbolRecord},
	"etfs":               {resource: "etf/list", mapRecord: fmSymbolRecord},
	"stock_screener":     {resource: "stock-screener", mapRecord: fmScreenerRecord, paginated: true},
	"delisted_companies": {resource: "delisted-companies", mapRecord: fmDelistedRecord, paginated: true},
}

// fmStreams returns the connector's published stream catalog.
func fmStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "stocks",
			Description: "All tradable stock symbols on Financial Modeling Prep (stock/list).",
			PrimaryKey:  []string{"symbol"},
			Fields:      fmSymbolFields(),
		},
		{
			Name:        "etfs",
			Description: "All tradable ETF symbols on Financial Modeling Prep (etf/list).",
			PrimaryKey:  []string{"symbol"},
			Fields:      fmSymbolFields(),
		},
		{
			Name:        "stock_screener",
			Description: "Stocks matching the configured screener filters (exchange, market cap).",
			PrimaryKey:  []string{"symbol"},
			Fields:      fmScreenerFields(),
		},
		{
			Name:        "delisted_companies",
			Description: "Companies delisted from supported exchanges.",
			PrimaryKey:  []string{"symbol"},
			Fields:      fmDelistedFields(),
		},
	}
}

func fmSymbolFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "price", Type: "number"},
		{Name: "exchange", Type: "string"},
		{Name: "exchange_short_name", Type: "string"},
		{Name: "type", Type: "string"},
	}
}

func fmScreenerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "company_name", Type: "string"},
		{Name: "market_cap", Type: "integer"},
		{Name: "sector", Type: "string"},
		{Name: "industry", Type: "string"},
		{Name: "beta", Type: "number"},
		{Name: "price", Type: "number"},
		{Name: "last_annual_dividend", Type: "number"},
		{Name: "volume", Type: "integer"},
		{Name: "exchange", Type: "string"},
		{Name: "exchange_short_name", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "is_etf", Type: "boolean"},
		{Name: "is_actively_trading", Type: "boolean"},
	}
}

func fmDelistedFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "company_name", Type: "string"},
		{Name: "exchange", Type: "string"},
		{Name: "ipo_date", Type: "string"},
		{Name: "delisted_date", Type: "string"},
	}
}

// fmSymbolRecord maps stock/list and etf/list objects.
func fmSymbolRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"symbol":              item["symbol"],
		"name":                item["name"],
		"price":               item["price"],
		"exchange":            item["exchange"],
		"exchange_short_name": item["exchangeShortName"],
		"type":                item["type"],
	}
}

// fmScreenerRecord maps stock-screener objects.
func fmScreenerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"symbol":               item["symbol"],
		"company_name":         item["companyName"],
		"market_cap":           item["marketCap"],
		"sector":               item["sector"],
		"industry":             item["industry"],
		"beta":                 item["beta"],
		"price":                item["price"],
		"last_annual_dividend": item["lastAnnualDividend"],
		"volume":               item["volume"],
		"exchange":             item["exchange"],
		"exchange_short_name":  item["exchangeShortName"],
		"country":              item["country"],
		"is_etf":               item["isEtf"],
		"is_actively_trading":  item["isActivelyTrading"],
	}
}

// fmDelistedRecord maps delisted-companies objects.
func fmDelistedRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"symbol":        item["symbol"],
		"company_name":  item["companyName"],
		"exchange":      item["exchange"],
		"ipo_date":      item["ipoDate"],
		"delisted_date": item["delistedDate"],
	}
}
