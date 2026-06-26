package finage

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Finage API resource path (relative to
// base_url), the JSON path to its records array, whether it is partitioned by
// the configured symbols, and the record mapper that flattens its objects.
//
// Finage's "market information" list endpoints return a top-level JSON array
// (recordsPath ""), while news returns {symbol, news:[...]} (recordsPath
// "news"). symbolPath endpoints embed a per-symbol segment and are fetched once
// per configured symbol.
type streamEndpoint struct {
	// resource is the path segment relative to base_url. For symbolPath streams
	// it contains a single "%s" placeholder for the symbol.
	resource string
	// recordsPath is the dotted JSON path to the records array ("" = root array).
	recordsPath string
	// symbolPath is true when resource needs a symbol substituted and the stream
	// is fetched once per configured symbol.
	symbolPath bool
	// extraParams are static query parameters Finage requires for the endpoint.
	extraParams map[string]string
	// mapRecord flattens a raw Finage object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// finageStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in finageStreams; the read path
// is fully data-driven from this table.
var finageStreamEndpoints = map[string]streamEndpoint{
	"most_active_us_stocks": {
		resource:    "/fnd/market-information/us/most-actives",
		recordsPath: "",
		mapRecord:   finageMoverRecord,
	},
	"most_gainers": {
		resource:    "/fnd/market-information/us/most-gainers",
		recordsPath: "",
		mapRecord:   finageMoverRecord,
	},
	"most_losers": {
		resource:    "/fnd/market-information/us/most-losers",
		recordsPath: "",
		mapRecord:   finageMoverRecord,
	},
	"sector_performance": {
		resource:    "/fnd/market-information/us/sector-performance",
		recordsPath: "",
		mapRecord:   finageSectorRecord,
	},
	"delisted_companies": {
		resource:    "/fnd/delisted-companies/",
		recordsPath: "",
		extraParams: map[string]string{"limit": "1000", "period": "annual"},
		mapRecord:   finageDelistedRecord,
	},
	"market_news": {
		resource:    "/news/market/%s",
		recordsPath: "news",
		symbolPath:  true,
		extraParams: map[string]string{"limit": "30"},
		mapRecord:   finageNewsRecord,
	},
}

// finageStreams returns the connector's published stream catalog.
func finageStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "most_active_us_stocks",
			Description: "Most active US stocks for the latest trading session.",
			PrimaryKey:  []string{"symbol"},
			Fields:      finageMoverFields(),
		},
		{
			Name:        "most_gainers",
			Description: "Top gaining US stocks for the latest trading session.",
			PrimaryKey:  []string{"symbol"},
			Fields:      finageMoverFields(),
		},
		{
			Name:        "most_losers",
			Description: "Top losing US stocks for the latest trading session.",
			PrimaryKey:  []string{"symbol"},
			Fields:      finageMoverFields(),
		},
		{
			Name:        "sector_performance",
			Description: "US sector performance for the latest trading session.",
			PrimaryKey:  []string{"sector"},
			Fields:      finageSectorFields(),
		},
		{
			Name:        "delisted_companies",
			Description: "Companies delisted from US exchanges.",
			PrimaryKey:  []string{"symbol"},
			Fields:      finageDelistedFields(),
		},
		{
			Name:        "market_news",
			Description: "Market news per configured symbol.",
			PrimaryKey:  []string{"url"},
			Fields:      finageNewsFields(),
		},
	}
}

func finageMoverFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "company_name", Type: "string"},
		{Name: "change", Type: "number"},
		{Name: "change_percentage", Type: "string"},
		{Name: "price", Type: "string"},
	}
}

func finageSectorFields() []connectors.Field {
	return []connectors.Field{
		{Name: "sector", Type: "string"},
		{Name: "change_percentage", Type: "string"},
	}
}

func finageDelistedFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "company_name", Type: "string"},
		{Name: "exchange", Type: "string"},
		{Name: "ipo_date", Type: "string"},
		{Name: "delisted_date", Type: "string"},
	}
}

func finageNewsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "date", Type: "string"},
	}
}

func finageMoverRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"symbol":            item["symbol"],
		"company_name":      item["company_name"],
		"change":            item["change"],
		"change_percentage": item["change_percentage"],
		"price":             item["price"],
	}
}

func finageSectorRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"sector":            item["sector"],
		"change_percentage": item["change_percentage"],
	}
}

func finageDelistedRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"symbol":        item["symbol"],
		"company_name":  item["company_name"],
		"exchange":      item["exchange"],
		"ipo_date":      item["ipo_date"],
		"delisted_date": item["delisted_date"],
	}
}

func finageNewsRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"symbol":      item["symbol"],
		"title":       item["title"],
		"url":         item["url"],
		"source":      item["source"],
		"description": item["description"],
		"date":        item["date"],
	}
}
