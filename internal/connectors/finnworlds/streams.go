package finnworlds

import "polymetrics.ai/internal/connectors"

// partitionKind identifies which config list a stream fans out over: each value
// in the list becomes one HTTP request with that value injected as a query param.
type partitionKind int

const (
	partitionNone partitionKind = iota
	partitionTickers
	partitionCommodities
)

// streamEndpoint maps a stream name to the Finnworlds API resource path, the
// dotted JSON path to its records array, the partition list it fans out over, and
// the record mapper that flattens its objects.
//
// Finnworlds wraps every response as {"result":{"output": ...}}. Most streams
// expose the array directly at result.output; dividends and stock_splits nest it
// one level deeper. There is no page-token pagination — a read fans out across a
// config list (tickers / commodities), one request per value, and each request
// returns the full dataset for that value.
type streamEndpoint struct {
	resource     string
	recordsPath  string
	partition    partitionKind
	partitionKey string // query param name for the partition value (e.g. "ticker")
	stitchField  string // record field the partition value is written onto
	mapRecord    func(map[string]any) connectors.Record
}

// finnworldsStreamEndpoints is the per-stream routing table. Adding a stream is
// one entry here plus a Stream definition in finnworldsStreams; the read path is
// fully data-driven from this table.
var finnworldsStreamEndpoints = map[string]streamEndpoint{
	"dividends": {
		resource:     "dividends",
		recordsPath:  "result.output.dividends",
		partition:    partitionTickers,
		partitionKey: "ticker",
		stitchField:  "ticker",
		mapRecord:    dividendRecord,
	},
	"stock_splits": {
		resource:     "stocksplits",
		recordsPath:  "result.output.stocksplits",
		partition:    partitionTickers,
		partitionKey: "ticker",
		stitchField:  "ticker",
		mapRecord:    stockSplitRecord,
	},
	"historical_candlestick": {
		resource:     "historicalcandlestick",
		recordsPath:  "result.output",
		partition:    partitionTickers,
		partitionKey: "ticker",
		stitchField:  "ticker",
		mapRecord:    candlestickRecord,
	},
	"commodities": {
		resource:     "commodities",
		recordsPath:  "result.output",
		partition:    partitionCommodities,
		partitionKey: "commodity_name",
		stitchField:  "commodity_name",
		mapRecord:    commodityRecord,
	},
}

// finnworldsStreams returns the connector's published stream catalog.
func finnworldsStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "dividends",
			Description:  "Dividend history for the configured tickers.",
			PrimaryKey:   []string{"ticker", "date"},
			CursorFields: []string{"date"},
			Fields:       dividendFields(),
		},
		{
			Name:         "stock_splits",
			Description:  "Stock split history for the configured tickers.",
			PrimaryKey:   []string{"ticker", "date"},
			CursorFields: []string{"date"},
			Fields:       stockSplitFields(),
		},
		{
			Name:         "historical_candlestick",
			Description:  "Historical OHLC candlestick data for the configured tickers.",
			PrimaryKey:   []string{"ticker", "date"},
			CursorFields: []string{"date"},
			Fields:       candlestickFields(),
		},
		{
			Name:         "commodities",
			Description:  "Commodity prices for the configured commodities.",
			PrimaryKey:   []string{"commodity_name", "datetime"},
			CursorFields: []string{"datetime"},
			Fields:       commodityFields(),
		},
	}
}

func dividendFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ticker", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "dividend_rate", Type: "string"},
	}
}

func stockSplitFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ticker", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "stock_split", Type: "string"},
	}
}

func candlestickFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ticker", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "open", Type: "string"},
		{Name: "high", Type: "string"},
		{Name: "low", Type: "string"},
		{Name: "close", Type: "string"},
		{Name: "adjusted_close", Type: "string"},
		{Name: "trade_volume", Type: "string"},
		{Name: "opentime", Type: "number"},
		{Name: "closetime", Type: "number"},
	}
}

func commodityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "commodity_name", Type: "string"},
		{Name: "datetime", Type: "string"},
		{Name: "commodity_price", Type: "string"},
		{Name: "commodity_unit", Type: "string"},
		{Name: "price_change_day", Type: "string"},
		{Name: "percentage_day", Type: "string"},
		{Name: "percentage_week", Type: "string"},
		{Name: "percentage_month", Type: "string"},
		{Name: "percentage_year", Type: "string"},
	}
}

func dividendRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ticker":        item["ticker"],
		"date":          item["date"],
		"dividend_rate": item["dividend_rate"],
	}
}

func stockSplitRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ticker":      item["ticker"],
		"date":        item["date"],
		"stock_split": item["stock_split"],
	}
}

func candlestickRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ticker":         item["ticker"],
		"date":           item["date"],
		"open":           item["open"],
		"high":           item["high"],
		"low":            item["low"],
		"close":          item["close"],
		"adjusted_close": item["adjusted_close"],
		"trade_volume":   item["trade_volume"],
		"opentime":       item["opentime"],
		"closetime":      item["closetime"],
	}
}

func commodityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"commodity_name":   item["commodity_name"],
		"datetime":         item["datetime"],
		"commodity_price":  item["commodity_price"],
		"commodity_unit":   item["commodity_unit"],
		"price_change_day": item["price_change_day"],
		"percentage_day":   item["percentage_day"],
		"percentage_week":  item["percentage_week"],
		"percentage_month": item["percentage_month"],
		"percentage_year":  item["percentage_year"],
	}
}
