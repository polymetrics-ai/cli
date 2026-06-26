package finnhub

import "polymetrics.ai/internal/connectors"

// scope describes how a stream is fanned out across the request space.
type scope int

const (
	// scopeOnce issues a single request (e.g. global market news).
	scopeOnce scope = iota
	// scopeExchange issues a single request parameterized by the exchange config.
	scopeExchange
	// scopeSymbol issues one request per configured symbol; Finnhub has no list
	// pagination, so iterating the symbols config is the connector's "pages".
	scopeSymbol
)

// streamEndpoint maps a stream name to the Finnhub API resource path (relative to
// base_url) it reads from, how it is scoped, whether it needs the date window,
// and the record mapper that flattens its items.
type streamEndpoint struct {
	// resource is the Finnhub path segment (e.g. "company-news").
	resource string
	// scope controls request fan-out (once / per-exchange / per-symbol).
	scope scope
	// dateWindow is true when the endpoint requires from/to query params.
	dateWindow bool
	// mapRecord flattens a raw Finnhub item into a connectors.Record. symbol is
	// the symbol the request was scoped to ("" for non-symbol streams).
	mapRecord func(item map[string]any, symbol string) connectors.Record
}

// finnhubStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in finnhubStreams; the read path
// is fully data-driven from this table.
var finnhubStreamEndpoints = map[string]streamEndpoint{
	"stock_symbols":         {resource: "stock/symbol", scope: scopeExchange, mapRecord: stockSymbolRecord},
	"market_news":           {resource: "news", scope: scopeOnce, mapRecord: newsRecord},
	"company_news":          {resource: "company-news", scope: scopeSymbol, dateWindow: true, mapRecord: newsRecord},
	"company_profile":       {resource: "stock/profile2", scope: scopeSymbol, mapRecord: companyProfileRecord},
	"stock_recommendations": {resource: "stock/recommendation", scope: scopeSymbol, mapRecord: recommendationRecord},
}

// finnhubStreams returns the connector's published stream catalog.
func finnhubStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "stock_symbols",
			Description: "Supported stock symbols for the configured exchange.",
			PrimaryKey:  []string{"symbol"},
			Fields:      stockSymbolFields(),
		},
		{
			Name:         "market_news",
			Description:  "Latest market news for the configured category.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"datetime"},
			Fields:       newsFields(),
		},
		{
			Name:         "company_news",
			Description:  "Company-specific news for each configured symbol.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"datetime"},
			Fields:       newsFields(),
		},
		{
			Name:        "company_profile",
			Description: "Company profile for each configured symbol.",
			PrimaryKey:  []string{"ticker"},
			Fields:      companyProfileFields(),
		},
		{
			Name:        "stock_recommendations",
			Description: "Analyst recommendation trends for each configured symbol.",
			PrimaryKey:  []string{"symbol", "period"},
			Fields:      recommendationFields(),
		},
	}
}

func stockSymbolFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "displaySymbol", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "figi", Type: "string"},
		{Name: "mic", Type: "string"},
	}
}

func newsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "symbol", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "datetime", Type: "integer"},
		{Name: "headline", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "related", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "image", Type: "string"},
	}
}

func companyProfileFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ticker", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "exchange", Type: "string"},
		{Name: "ipo", Type: "string"},
		{Name: "marketCapitalization", Type: "number"},
		{Name: "shareOutstanding", Type: "number"},
		{Name: "finnhubIndustry", Type: "string"},
		{Name: "weburl", Type: "string"},
		{Name: "logo", Type: "string"},
		{Name: "phone", Type: "string"},
	}
}

func recommendationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "period", Type: "string"},
		{Name: "buy", Type: "integer"},
		{Name: "hold", Type: "integer"},
		{Name: "sell", Type: "integer"},
		{Name: "strongBuy", Type: "integer"},
		{Name: "strongSell", Type: "integer"},
	}
}

func stockSymbolRecord(item map[string]any, _ string) connectors.Record {
	return connectors.Record{
		"symbol":        item["symbol"],
		"displaySymbol": item["displaySymbol"],
		"description":   item["description"],
		"type":          item["type"],
		"currency":      item["currency"],
		"figi":          item["figi"],
		"mic":           item["mic"],
	}
}

func newsRecord(item map[string]any, symbol string) connectors.Record {
	rec := connectors.Record{
		"id":       item["id"],
		"category": item["category"],
		"datetime": item["datetime"],
		"headline": item["headline"],
		"source":   item["source"],
		"summary":  item["summary"],
		"related":  item["related"],
		"url":      item["url"],
		"image":    item["image"],
	}
	// company_news is scoped to a symbol; market_news is not. Surface the symbol
	// the article was fetched for so downstream joins work.
	if symbol != "" {
		rec["symbol"] = symbol
	} else {
		rec["symbol"] = item["related"]
	}
	return rec
}

func companyProfileRecord(item map[string]any, symbol string) connectors.Record {
	ticker := item["ticker"]
	if ticker == nil || ticker == "" {
		ticker = symbol
	}
	return connectors.Record{
		"ticker":               ticker,
		"name":                 item["name"],
		"country":              item["country"],
		"currency":             item["currency"],
		"exchange":             item["exchange"],
		"ipo":                  item["ipo"],
		"marketCapitalization": item["marketCapitalization"],
		"shareOutstanding":     item["shareOutstanding"],
		"finnhubIndustry":      item["finnhubIndustry"],
		"weburl":               item["weburl"],
		"logo":                 item["logo"],
		"phone":                item["phone"],
	}
}

func recommendationRecord(item map[string]any, symbol string) connectors.Record {
	sym := item["symbol"]
	if sym == nil || sym == "" {
		sym = symbol
	}
	return connectors.Record{
		"symbol":     sym,
		"period":     item["period"],
		"buy":        item["buy"],
		"hold":       item["hold"],
		"sell":       item["sell"],
		"strongBuy":  item["strongBuy"],
		"strongSell": item["strongSell"],
	}
}
