package coingeckocoins

import "polymetrics.ai/internal/connectors"

// coingeckoStreams is the set of streams this connector publishes. CoinGecko's
// coin endpoints are not list-paginated; each stream is driven by config
// (coin_id, vs_currency, days, start_date/end_date) and shaped by the read path
// in coingecko_coins.go. The three streams mirror the upstream Airbyte connector
// (market_chart, history) plus a coin metadata snapshot.
func coingeckoStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "market_chart",
			Description:  "Historical price, market cap, and total volume time series for a coin from /coins/{id}/market_chart, one record per timestamp.",
			PrimaryKey:   []string{"coin_id", "vs_currency", "timestamp"},
			CursorFields: []string{"timestamp"},
			Fields:       marketChartFields(),
		},
		{
			Name:         "history",
			Description:  "Daily historical snapshots for a coin from /coins/{id}/history, one record per date between start_date and end_date.",
			PrimaryKey:   []string{"coin_id", "date"},
			CursorFields: []string{"date"},
			Fields:       historyFields(),
		},
		{
			Name:        "coin",
			Description: "Current coin metadata and market data snapshot from /coins/{id}.",
			PrimaryKey:  []string{"id"},
			Fields:      coinFields(),
		},
	}
}

func marketChartFields() []connectors.Field {
	return []connectors.Field{
		{Name: "coin_id", Type: "string"},
		{Name: "vs_currency", Type: "string"},
		{Name: "timestamp", Type: "integer"},
		{Name: "price", Type: "number"},
		{Name: "market_cap", Type: "number"},
		{Name: "total_volume", Type: "number"},
	}
}

func historyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "coin_id", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "id", Type: "string"},
		{Name: "symbol", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "current_price", Type: "object"},
		{Name: "market_cap", Type: "object"},
		{Name: "total_volume", Type: "object"},
	}
}

func coinFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "symbol", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "market_cap_rank", Type: "integer"},
		{Name: "hashing_algorithm", Type: "string"},
		{Name: "categories", Type: "array"},
		{Name: "market_data", Type: "object"},
		{Name: "last_updated", Type: "string"},
	}
}

// marketChartRecord builds one market_chart record from the value at a single
// timestamp index across the parallel prices/market_caps/total_volumes arrays.
func marketChartRecord(coinID, vsCurrency string, timestamp, price, marketCap, totalVolume any) connectors.Record {
	return connectors.Record{
		"coin_id":      coinID,
		"vs_currency":  vsCurrency,
		"timestamp":    timestamp,
		"price":        price,
		"market_cap":   marketCap,
		"total_volume": totalVolume,
	}
}

// historyRecord flattens a /coins/{id}/history snapshot into a record, injecting
// the coin_id and the requested date (which is the natural cursor/primary key
// since the date param is what drives the per-day "pagination").
func historyRecord(coinID, date string, item map[string]any) connectors.Record {
	rec := connectors.Record{
		"coin_id": coinID,
		"date":    date,
		"id":      item["id"],
		"symbol":  item["symbol"],
		"name":    item["name"],
	}
	if md, ok := item["market_data"].(map[string]any); ok {
		rec["current_price"] = md["current_price"]
		rec["market_cap"] = md["market_cap"]
		rec["total_volume"] = md["total_volume"]
	}
	return rec
}

// coinRecord flattens a /coins/{id} metadata snapshot into a record.
func coinRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"symbol":            item["symbol"],
		"name":              item["name"],
		"market_cap_rank":   item["market_cap_rank"],
		"hashing_algorithm": item["hashing_algorithm"],
		"categories":        item["categories"],
		"market_data":       item["market_data"],
		"last_updated":      item["last_updated"],
	}
}
