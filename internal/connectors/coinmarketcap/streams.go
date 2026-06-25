package coinmarketcap

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the CoinMarketCap Pro API resource path
// (relative to base_url) it reads from, whether that path paginates, and the
// record mapper that flattens its objects. CoinMarketCap wraps every payload in
// {status:{...}, data:...}; `data` is an array for list endpoints and a single
// object for global-metrics, both of which connsdk.RecordsAt handles.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "v1/cryptocurrency/map").
	resource string
	// paginated marks endpoints that accept start/limit pagination.
	paginated bool
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// coinmarketcapStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in coinmarketcapStreams;
// the read path is fully data-driven from this table.
var coinmarketcapStreamEndpoints = map[string]streamEndpoint{
	"map":             {resource: "v1/cryptocurrency/map", paginated: true, mapRecord: cryptoMapRecord},
	"listings_latest": {resource: "v1/cryptocurrency/listings/latest", paginated: true, mapRecord: listingRecord},
	"categories":      {resource: "v1/cryptocurrency/categories", paginated: true, mapRecord: categoryRecord},
	"fiat":            {resource: "v1/fiat/map", paginated: true, mapRecord: fiatRecord},
	"global_metrics":  {resource: "v1/global-metrics/quotes/latest", paginated: false, mapRecord: globalMetricsRecord},
}

// coinmarketcapStreams returns the connector's published stream catalog.
func coinmarketcapStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "map",
			Description: "CoinMarketCap ID map of all active cryptocurrencies.",
			PrimaryKey:  []string{"id"},
			Fields:      cryptoMapFields(),
		},
		{
			Name:        "listings_latest",
			Description: "Latest market ticker listings for active cryptocurrencies, ranked by cmc_rank.",
			PrimaryKey:  []string{"id"},
			Fields:      listingFields(),
		},
		{
			Name:        "categories",
			Description: "Cryptocurrency categories tracked by CoinMarketCap.",
			PrimaryKey:  []string{"id"},
			Fields:      categoryFields(),
		},
		{
			Name:        "fiat",
			Description: "Supported fiat currencies and precious metals.",
			PrimaryKey:  []string{"id"},
			Fields:      fiatFields(),
		},
		{
			Name:        "global_metrics",
			Description: "Latest aggregate global cryptocurrency market metrics (single record).",
			PrimaryKey:  []string{"active_cryptocurrencies"},
			Fields:      globalMetricsFields(),
		},
	}
}

func cryptoMapFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "symbol", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "rank", Type: "integer"},
		{Name: "is_active", Type: "integer"},
		{Name: "first_historical_data", Type: "string"},
		{Name: "last_historical_data", Type: "string"},
	}
}

func listingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "symbol", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "cmc_rank", Type: "integer"},
		{Name: "num_market_pairs", Type: "integer"},
		{Name: "circulating_supply", Type: "number"},
		{Name: "total_supply", Type: "number"},
		{Name: "max_supply", Type: "number"},
		{Name: "last_updated", Type: "string"},
		{Name: "quote", Type: "object"},
	}
}

func categoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "num_tokens", Type: "integer"},
		{Name: "market_cap", Type: "number"},
		{Name: "volume", Type: "number"},
		{Name: "last_updated", Type: "string"},
	}
}

func fiatFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "sign", Type: "string"},
		{Name: "symbol", Type: "string"},
	}
}

func globalMetricsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "active_cryptocurrencies", Type: "integer"},
		{Name: "total_cryptocurrencies", Type: "integer"},
		{Name: "active_market_pairs", Type: "integer"},
		{Name: "active_exchanges", Type: "integer"},
		{Name: "total_exchanges", Type: "integer"},
		{Name: "btc_dominance", Type: "number"},
		{Name: "eth_dominance", Type: "number"},
		{Name: "last_updated", Type: "string"},
		{Name: "quote", Type: "object"},
	}
}

func cryptoMapRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["id"],
		"name":                  item["name"],
		"symbol":                item["symbol"],
		"slug":                  item["slug"],
		"rank":                  item["rank"],
		"is_active":             item["is_active"],
		"first_historical_data": item["first_historical_data"],
		"last_historical_data":  item["last_historical_data"],
	}
}

func listingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"symbol":             item["symbol"],
		"slug":               item["slug"],
		"cmc_rank":           item["cmc_rank"],
		"num_market_pairs":   item["num_market_pairs"],
		"circulating_supply": item["circulating_supply"],
		"total_supply":       item["total_supply"],
		"max_supply":         item["max_supply"],
		"last_updated":       item["last_updated"],
		"quote":              item["quote"],
	}
}

func categoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"title":        item["title"],
		"description":  item["description"],
		"num_tokens":   item["num_tokens"],
		"market_cap":   item["market_cap"],
		"volume":       item["volume"],
		"last_updated": item["last_updated"],
	}
}

func fiatRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":     item["id"],
		"name":   item["name"],
		"sign":   item["sign"],
		"symbol": item["symbol"],
	}
}

func globalMetricsRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"active_cryptocurrencies": item["active_cryptocurrencies"],
		"total_cryptocurrencies":  item["total_cryptocurrencies"],
		"active_market_pairs":     item["active_market_pairs"],
		"active_exchanges":        item["active_exchanges"],
		"total_exchanges":         item["total_exchanges"],
		"btc_dominance":           item["btc_dominance"],
		"eth_dominance":           item["eth_dominance"],
		"last_updated":            item["last_updated"],
		"quote":                   item["quote"],
	}
}
