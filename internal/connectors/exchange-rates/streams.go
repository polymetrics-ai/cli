package exchangerates

import "polymetrics.ai/internal/connectors"

// Stream names exposed by the connector.
const (
	streamExchangeRates = "exchange_rates"
	streamLatest        = "latest"
	streamSymbols       = "symbols"
)

// exchangeRatesStreams returns the connector's published stream catalog.
//
//   - exchange_rates: daily historical rates, one record per ISO date. The
//     connector iterates date-by-date from start_date to end_date (the API's
//     natural unit of progress), so the primary key and incremental cursor are
//     both ["date"].
//   - latest: the most recent rates, a single record keyed on its date.
//   - symbols: the list of supported currency codes, one record per code.
func exchangeRatesStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         streamExchangeRates,
			Description:  "Daily historical foreign-exchange rates, one record per ISO date.",
			PrimaryKey:   []string{"date"},
			CursorFields: []string{"date"},
			Fields:       rateFields(),
		},
		{
			Name:        streamLatest,
			Description: "The latest available foreign-exchange rates.",
			PrimaryKey:  []string{"date"},
			Fields:      rateFields(),
		},
		{
			Name:        streamSymbols,
			Description: "Supported currency codes and their display names.",
			PrimaryKey:  []string{"code"},
			Fields: []connectors.Field{
				{Name: "code", Type: "string"},
				{Name: "name", Type: "string"},
			},
		},
	}
}

func rateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "date", Type: "string"},
		{Name: "base", Type: "string"},
		{Name: "timestamp", Type: "integer"},
		{Name: "historical", Type: "boolean"},
		{Name: "success", Type: "boolean"},
		{Name: "rates", Type: "object"},
	}
}

// rateRecord flattens a single API rates payload (latest or historical) into a
// connectors.Record. The nested rates object is preserved as a map so downstream
// consumers can fan it out per currency without losing the daily grouping.
func rateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"date":       item["date"],
		"base":       item["base"],
		"timestamp":  item["timestamp"],
		"historical": item["historical"],
		"success":    item["success"],
		"rates":      normalizeRates(item["rates"]),
	}
}

// normalizeRates coerces the decoded rates value into a map[string]any so the
// mapped record always exposes a stable type regardless of the JSON decoder used.
func normalizeRates(v any) map[string]any {
	switch r := v.(type) {
	case map[string]any:
		return r
	default:
		return map[string]any{}
	}
}

// symbolRecord maps a single (code, name) pair from the symbols endpoint.
func symbolRecord(code string, name any) connectors.Record {
	return connectors.Record{
		"code": code,
		"name": name,
	}
}
