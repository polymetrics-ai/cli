package openexchangerates

import "polymetrics.ai/internal/connectors"

// Open Exchange Rates returns rate payloads shaped as
// {timestamp, base, rates:{CCY: number, ...}}. The rates object is a map, not an
// array, so each stream's mapper flattens one record per currency code. The
// currencies stream flattens the {CCY: "Name"} map, and usage returns a single
// account-status record.

// oerStreams returns the connector's published stream catalog.
func oerStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "latest",
			Description:  "Latest exchange rates, one row per quote currency.",
			PrimaryKey:   []string{"base", "currency"},
			CursorFields: []string{"timestamp"},
			Fields:       oerRateFields(),
		},
		{
			Name:         "historical",
			Description:  "Historical end-of-day exchange rates per date, one row per quote currency.",
			PrimaryKey:   []string{"date", "base", "currency"},
			CursorFields: []string{"date"},
			Fields:       oerHistoricalFields(),
		},
		{
			Name:        "currencies",
			Description: "Map of supported ISO currency codes to their display names.",
			PrimaryKey:  []string{"currency"},
			Fields: []connectors.Field{
				{Name: "currency", Type: "string"},
				{Name: "name", Type: "string"},
			},
		},
		{
			Name:        "usage",
			Description: "App ID usage and plan status for the account.",
			PrimaryKey:  []string{"app_id"},
			Fields: []connectors.Field{
				{Name: "app_id", Type: "string"},
				{Name: "status", Type: "integer"},
				{Name: "plan", Type: "string"},
				{Name: "requests", Type: "integer"},
				{Name: "requests_quota", Type: "integer"},
				{Name: "requests_remaining", Type: "integer"},
				{Name: "days_elapsed", Type: "integer"},
				{Name: "days_remaining", Type: "integer"},
				{Name: "daily_average", Type: "integer"},
			},
		},
	}
}

func oerRateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "base", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "rate", Type: "number"},
		{Name: "timestamp", Type: "integer"},
	}
}

func oerHistoricalFields() []connectors.Field {
	return []connectors.Field{
		{Name: "date", Type: "string"},
		{Name: "base", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "rate", Type: "number"},
		{Name: "timestamp", Type: "integer"},
	}
}

// rateRecords flattens a {timestamp, base, rates:{...}} payload into one record
// per currency. When date is non-empty (historical stream) it is added to each
// record.
func rateRecords(payload map[string]any, date string, emit func(connectors.Record) error) error {
	base := stringField(payload, "base")
	if base == "" {
		base = "USD"
	}
	timestamp := payload["timestamp"]
	rates, _ := payload["rates"].(map[string]any)
	for _, currency := range sortedKeys(rates) {
		rec := connectors.Record{
			"base":      base,
			"currency":  currency,
			"rate":      rates[currency],
			"timestamp": timestamp,
		}
		if date != "" {
			rec["date"] = date
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// currencyRecords flattens the {CCY: "Name"} currencies map.
func currencyRecords(payload map[string]any, emit func(connectors.Record) error) error {
	for _, currency := range sortedKeys(payload) {
		if err := emit(connectors.Record{
			"currency": currency,
			"name":     payload[currency],
		}); err != nil {
			return err
		}
	}
	return nil
}

// usageRecord flattens the usage.json payload into a single record. The shape is
// {status, data:{app_id, status, plan:{...}, usage:{...}}}.
func usageRecord(payload map[string]any) connectors.Record {
	data, _ := payload["data"].(map[string]any)
	if data == nil {
		data = payload
	}
	usage, _ := data["usage"].(map[string]any)
	plan, _ := data["plan"].(map[string]any)
	rec := connectors.Record{
		"app_id": data["app_id"],
		"status": data["status"],
	}
	if plan != nil {
		rec["plan"] = plan["name"]
	}
	if usage != nil {
		rec["requests"] = usage["requests"]
		rec["requests_quota"] = usage["requests_quota"]
		rec["requests_remaining"] = usage["requests_remaining"]
		rec["days_elapsed"] = usage["days_elapsed"]
		rec["days_remaining"] = usage["days_remaining"]
		rec["daily_average"] = usage["daily_average"]
	}
	return rec
}
