package breezometer

import "polymetrics/internal/connectors"

// listStream marks endpoints whose `data` payload is an array of time-series
// points (forecast/history) versus single point-in-time objects (current
// conditions). It controls both extraction shape and pagination.
type streamEndpoint struct {
	// resource is the BreezoMeter API path (relative to base_url).
	resource string
	// list is true when `data` is an array of points; false for a single object.
	list bool
	// mapRecord flattens a raw BreezoMeter data point into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// breezometerStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in breezometerStreams;
// the read path is fully data-driven from this table.
var breezometerStreamEndpoints = map[string]streamEndpoint{
	"air_quality_current":  {resource: "air-quality/v2/current-conditions", list: false, mapRecord: airQualityRecord},
	"air_quality_forecast": {resource: "air-quality/v2/forecast/hourly", list: true, mapRecord: airQualityRecord},
	"air_quality_history":  {resource: "air-quality/v2/historical/hourly", list: true, mapRecord: airQualityRecord},
	"pollen_forecast":      {resource: "pollen/v2/forecast/daily", list: true, mapRecord: pollenRecord},
	"weather_current":      {resource: "weather/v1/current-conditions", list: false, mapRecord: weatherRecord},
}

// breezometerStreams returns the connector's published stream catalog. Every
// BreezoMeter point carries a `datetime` and the injected location, so the
// composite primary key is [datetime, latitude, longitude] and the cursor is
// datetime where a time-series is returned.
func breezometerStreams() []connectors.Stream {
	geo := []connectors.Field{
		{Name: "datetime", Type: "string"},
		{Name: "latitude", Type: "string"},
		{Name: "longitude", Type: "string"},
		{Name: "data_available", Type: "boolean"},
	}
	return []connectors.Stream{
		{
			Name:         "air_quality_current",
			Description:  "BreezoMeter current air-quality conditions for the configured location.",
			PrimaryKey:   []string{"datetime", "latitude", "longitude"},
			CursorFields: []string{"datetime"},
			Fields:       append(geo, airQualityFields()...),
		},
		{
			Name:         "air_quality_forecast",
			Description:  "BreezoMeter hourly air-quality forecast for the configured location.",
			PrimaryKey:   []string{"datetime", "latitude", "longitude"},
			CursorFields: []string{"datetime"},
			Fields:       append(geo, airQualityFields()...),
		},
		{
			Name:         "air_quality_history",
			Description:  "BreezoMeter hourly historical air quality for the configured location.",
			PrimaryKey:   []string{"datetime", "latitude", "longitude"},
			CursorFields: []string{"datetime"},
			Fields:       append(geo, airQualityFields()...),
		},
		{
			Name:         "pollen_forecast",
			Description:  "BreezoMeter daily pollen forecast for the configured location.",
			PrimaryKey:   []string{"datetime", "latitude", "longitude"},
			CursorFields: []string{"datetime"},
			Fields:       append(geo, pollenFields()...),
		},
		{
			Name:         "weather_current",
			Description:  "BreezoMeter current weather conditions for the configured location.",
			PrimaryKey:   []string{"datetime", "latitude", "longitude"},
			CursorFields: []string{"datetime"},
			Fields:       append(geo, weatherFields()...),
		},
	}
}

func airQualityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "indexes", Type: "object"},
		{Name: "pollutants", Type: "object"},
		{Name: "health_recommendations", Type: "object"},
	}
}

func pollenFields() []connectors.Field {
	return []connectors.Field{
		{Name: "date", Type: "string"},
		{Name: "types", Type: "object"},
		{Name: "plants", Type: "object"},
		{Name: "index", Type: "object"},
	}
}

func weatherFields() []connectors.Field {
	return []connectors.Field{
		{Name: "temperature", Type: "object"},
		{Name: "feels_like_temperature", Type: "object"},
		{Name: "relative_humidity", Type: "integer"},
		{Name: "wind", Type: "object"},
		{Name: "precipitation", Type: "object"},
		{Name: "weather_condition", Type: "object"},
	}
}

func airQualityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"datetime":               item["datetime"],
		"indexes":                item["indexes"],
		"pollutants":             item["pollutants"],
		"health_recommendations": item["health_recommendations"],
		"data_available":         item["data_available"],
	}
}

func pollenRecord(item map[string]any) connectors.Record {
	dt := item["datetime"]
	if dt == nil {
		dt = item["date"]
	}
	return connectors.Record{
		"datetime":       dt,
		"date":           item["date"],
		"types":          item["types"],
		"plants":         item["plants"],
		"index":          item["index"],
		"data_available": item["data_available"],
	}
}

func weatherRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"datetime":               item["datetime"],
		"temperature":            item["temperature"],
		"feels_like_temperature": item["feels_like_temperature"],
		"relative_humidity":      item["relative_humidity"],
		"wind":                   item["wind"],
		"precipitation":          item["precipitation"],
		"weather_condition":      item["weather_condition"],
		"data_available":         item["data_available"],
	}
}
