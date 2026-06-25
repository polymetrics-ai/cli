package coinapi

import "polymetrics.ai/internal/connectors"

// streamKind classifies how a stream's endpoint is shaped and paginated.
type streamKind int

const (
	// kindMetadata is a static reference list (symbols, exchanges, assets):
	// GET /v1/<resource> returning a top-level JSON array, no time pagination.
	kindMetadata streamKind = iota
	// kindTimeseries is a symbol-scoped historical series (ohlcv/trades):
	// GET /v1/<resource>/<symbol_id>/history with limit + time_start/time_end,
	// paginated by advancing time_start past the last record's cursor field.
	kindTimeseries
)

// streamEndpoint maps a stream name to the CoinAPI resource it reads from, the
// record mapper, and (for time series) the cursor field used both as the
// incremental cursor and the pagination key.
type streamEndpoint struct {
	kind     streamKind
	resource string
	// cursorField is the record field holding the time cursor (time series only).
	cursorField string
	mapRecord   func(map[string]any) connectors.Record
}

// coinAPIStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in coinAPIStreams; the read
// path is fully data-driven from this table.
var coinAPIStreamEndpoints = map[string]streamEndpoint{
	"symbols":                {kind: kindMetadata, resource: "symbols", mapRecord: coinAPISymbolRecord},
	"exchanges":              {kind: kindMetadata, resource: "exchanges", mapRecord: coinAPIExchangeRecord},
	"assets":                 {kind: kindMetadata, resource: "assets", mapRecord: coinAPIAssetRecord},
	"ohlcv_historical_data":  {kind: kindTimeseries, resource: "ohlcv", cursorField: "time_period_start", mapRecord: coinAPIOHLCVRecord},
	"trades_historical_data": {kind: kindTimeseries, resource: "trades", cursorField: "time_exchange", mapRecord: coinAPITradeRecord},
}

// coinAPIStreams returns the connector's published stream catalog. The metadata
// streams are full-refresh reference lists; the historical series carry an
// incremental time cursor.
func coinAPIStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "symbols",
			Description: "All CoinAPI symbols (instrument identifiers across exchanges).",
			PrimaryKey:  []string{"symbol_id"},
			Fields:      coinAPISymbolFields(),
		},
		{
			Name:        "exchanges",
			Description: "All exchanges tracked by CoinAPI.",
			PrimaryKey:  []string{"exchange_id"},
			Fields:      coinAPIExchangeFields(),
		},
		{
			Name:        "assets",
			Description: "All assets (crypto and fiat) tracked by CoinAPI.",
			PrimaryKey:  []string{"asset_id"},
			Fields:      coinAPIAssetFields(),
		},
		{
			Name:         "ohlcv_historical_data",
			Description:  "Historical OHLCV (open/high/low/close/volume) candles for the configured symbol_id and period.",
			PrimaryKey:   []string{"symbol_id", "time_period_start"},
			CursorFields: []string{"time_period_start"},
			Fields:       coinAPIOHLCVFields(),
		},
		{
			Name:         "trades_historical_data",
			Description:  "Historical executed trades for the configured symbol_id.",
			PrimaryKey:   []string{"symbol_id", "uuid"},
			CursorFields: []string{"time_exchange"},
			Fields:       coinAPITradeFields(),
		},
	}
}

func coinAPISymbolFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol_id", Type: "string"},
		{Name: "exchange_id", Type: "string"},
		{Name: "symbol_type", Type: "string"},
		{Name: "asset_id_base", Type: "string"},
		{Name: "asset_id_quote", Type: "string"},
		{Name: "data_start", Type: "string"},
		{Name: "data_end", Type: "string"},
	}
}

func coinAPIExchangeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "exchange_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "data_quote_start", Type: "string"},
		{Name: "data_quote_end", Type: "string"},
		{Name: "data_symbols_count", Type: "integer"},
	}
}

func coinAPIAssetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "asset_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type_is_crypto", Type: "integer"},
		{Name: "data_start", Type: "string"},
		{Name: "data_end", Type: "string"},
		{Name: "price_usd", Type: "number"},
	}
}

func coinAPIOHLCVFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol_id", Type: "string"},
		{Name: "period_id", Type: "string"},
		{Name: "time_period_start", Type: "string"},
		{Name: "time_period_end", Type: "string"},
		{Name: "time_open", Type: "string"},
		{Name: "time_close", Type: "string"},
		{Name: "price_open", Type: "number"},
		{Name: "price_high", Type: "number"},
		{Name: "price_low", Type: "number"},
		{Name: "price_close", Type: "number"},
		{Name: "volume_traded", Type: "number"},
		{Name: "trades_count", Type: "integer"},
	}
}

func coinAPITradeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "symbol_id", Type: "string"},
		{Name: "uuid", Type: "string"},
		{Name: "time_exchange", Type: "string"},
		{Name: "time_coinapi", Type: "string"},
		{Name: "price", Type: "number"},
		{Name: "size", Type: "number"},
		{Name: "taker_side", Type: "string"},
	}
}

func coinAPISymbolRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"symbol_id":      item["symbol_id"],
		"exchange_id":    item["exchange_id"],
		"symbol_type":    item["symbol_type"],
		"asset_id_base":  item["asset_id_base"],
		"asset_id_quote": item["asset_id_quote"],
		"data_start":     item["data_start"],
		"data_end":       item["data_end"],
	}
}

func coinAPIExchangeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"exchange_id":        item["exchange_id"],
		"name":               item["name"],
		"website":            item["website"],
		"data_quote_start":   item["data_quote_start"],
		"data_quote_end":     item["data_quote_end"],
		"data_symbols_count": item["data_symbols_count"],
	}
}

func coinAPIAssetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"asset_id":       item["asset_id"],
		"name":           item["name"],
		"type_is_crypto": item["type_is_crypto"],
		"data_start":     item["data_start"],
		"data_end":       item["data_end"],
		"price_usd":      item["price_usd"],
	}
}

func coinAPIOHLCVRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"time_period_start": item["time_period_start"],
		"time_period_end":   item["time_period_end"],
		"time_open":         item["time_open"],
		"time_close":        item["time_close"],
		"price_open":        item["price_open"],
		"price_high":        item["price_high"],
		"price_low":         item["price_low"],
		"price_close":       item["price_close"],
		"volume_traded":     item["volume_traded"],
		"trades_count":      item["trades_count"],
	}
}

func coinAPITradeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uuid":          item["uuid"],
		"time_exchange": item["time_exchange"],
		"time_coinapi":  item["time_coinapi"],
		"price":         item["price"],
		"size":          item["size"],
		"taker_side":    item["taker_side"],
	}
}
