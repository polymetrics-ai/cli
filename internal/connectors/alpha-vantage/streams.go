package alphavantage

import "polymetrics/internal/connectors"

// streamSpec describes an Alpha Vantage "function" endpoint and how to extract
// its records. Alpha Vantage does not return JSON arrays: the time-series
// endpoints return a single object keyed by date strings, and GLOBAL_QUOTE
// returns a single flat object. seriesKey names the JSON object that holds the
// dated entries; an empty seriesKey marks a single-object (quote) stream whose
// records live at objectKey.
type streamSpec struct {
	// function is the Alpha Vantage `function` query value (e.g. TIME_SERIES_DAILY).
	function string
	// seriesKey is the response key holding the date-keyed series object, or ""
	// for single-object streams.
	seriesKey string
	// objectKey is the response key holding the single object for quote-style
	// streams (used only when seriesKey == "").
	objectKey string
	// intraday marks streams that require/accept the interval param.
	intraday bool
}

// alphaVantageStreamSpecs is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in alphaVantageStreams.
var alphaVantageStreamSpecs = map[string]streamSpec{
	"time_series_daily":    {function: "TIME_SERIES_DAILY", seriesKey: "Time Series (Daily)"},
	"time_series_weekly":   {function: "TIME_SERIES_WEEKLY", seriesKey: "Weekly Time Series"},
	"time_series_monthly":  {function: "TIME_SERIES_MONTHLY", seriesKey: "Monthly Time Series"},
	"time_series_intraday": {function: "TIME_SERIES_INTRADAY", intraday: true}, // seriesKey resolved at read time from interval
	"global_quote":         {function: "GLOBAL_QUOTE", objectKey: "Global Quote"},
}

// ohlcvFields are shared by every time-series stream: each dated entry exposes
// the same open/high/low/close/volume columns, plus the date and symbol the
// connector injects.
func ohlcvFields() []connectors.Field {
	return []connectors.Field{
		{Name: "date", Type: "string"},
		{Name: "symbol", Type: "string"},
		{Name: "open", Type: "number"},
		{Name: "high", Type: "number"},
		{Name: "low", Type: "number"},
		{Name: "close", Type: "number"},
		{Name: "volume", Type: "integer"},
	}
}

// alphaVantageStreams returns the connector's published stream catalog. Every
// time-series stream is keyed by (symbol, date); global_quote is keyed by
// (symbol, latest_trading_day). The cursor field is the date, which advances
// monotonically as new bars arrive.
func alphaVantageStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "time_series_daily",
			Description:  "Daily open/high/low/close/volume time series for the configured symbol.",
			PrimaryKey:   []string{"symbol", "date"},
			CursorFields: []string{"date"},
			Fields:       ohlcvFields(),
		},
		{
			Name:         "time_series_weekly",
			Description:  "Weekly open/high/low/close/volume time series for the configured symbol.",
			PrimaryKey:   []string{"symbol", "date"},
			CursorFields: []string{"date"},
			Fields:       ohlcvFields(),
		},
		{
			Name:         "time_series_monthly",
			Description:  "Monthly open/high/low/close/volume time series for the configured symbol.",
			PrimaryKey:   []string{"symbol", "date"},
			CursorFields: []string{"date"},
			Fields:       ohlcvFields(),
		},
		{
			Name:         "time_series_intraday",
			Description:  "Intraday open/high/low/close/volume bars for the configured symbol and interval.",
			PrimaryKey:   []string{"symbol", "date"},
			CursorFields: []string{"date"},
			Fields:       ohlcvFields(),
		},
		{
			Name:         "global_quote",
			Description:  "Latest price and volume snapshot for the configured symbol.",
			PrimaryKey:   []string{"symbol", "latest_trading_day"},
			CursorFields: []string{"latest_trading_day"},
			Fields: []connectors.Field{
				{Name: "symbol", Type: "string"},
				{Name: "open", Type: "number"},
				{Name: "high", Type: "number"},
				{Name: "low", Type: "number"},
				{Name: "price", Type: "number"},
				{Name: "volume", Type: "integer"},
				{Name: "latest_trading_day", Type: "string"},
				{Name: "previous_close", Type: "number"},
				{Name: "change", Type: "number"},
				{Name: "change_percent", Type: "string"},
			},
		},
	}
}

// ohlcvRecord flattens one dated time-series entry into a connectors.Record. The
// raw Alpha Vantage keys are numbered ("1. open", "5. volume"); they are mapped
// to stable snake_case column names. The connector injects symbol and date since
// those are not present inside each entry.
func ohlcvRecord(symbol, date string, entry map[string]any) connectors.Record {
	return connectors.Record{
		"symbol": symbol,
		"date":   date,
		"open":   numberedField(entry, "1. open"),
		"high":   numberedField(entry, "2. high"),
		"low":    numberedField(entry, "3. low"),
		"close":  numberedField(entry, "4. close"),
		"volume": numberedField(entry, "5. volume"),
	}
}

// globalQuoteRecord flattens the single Global Quote object into a record with
// stable column names.
func globalQuoteRecord(quote map[string]any) connectors.Record {
	return connectors.Record{
		"symbol":             numberedField(quote, "01. symbol"),
		"open":               numberedField(quote, "02. open"),
		"high":               numberedField(quote, "03. high"),
		"low":                numberedField(quote, "04. low"),
		"price":              numberedField(quote, "05. price"),
		"volume":             numberedField(quote, "06. volume"),
		"latest_trading_day": numberedField(quote, "07. latest trading day"),
		"previous_close":     numberedField(quote, "08. previous close"),
		"change":             numberedField(quote, "09. change"),
		"change_percent":     numberedField(quote, "10. change percent"),
	}
}

// numberedField reads a value from a raw Alpha Vantage entry. Values arrive as
// JSON strings (the API returns numbers as quoted strings); nil missing keys are
// returned as nil so downstream consumers can distinguish absent columns.
func numberedField(entry map[string]any, key string) any {
	if v, ok := entry[key]; ok {
		return v
	}
	return nil
}
