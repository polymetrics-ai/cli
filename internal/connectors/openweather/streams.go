package openweather

import "polymetrics.ai/internal/connectors"

// streamSpec maps a stream name to the top-level key of the One Call API 3.0
// response it reads from and the record mapper that flattens its objects. The
// One Call endpoint returns a single JSON document whose keys (current, hourly,
// daily, alerts, minutely) each back one stream; "current" is a single object
// while the others are arrays.
type streamSpec struct {
	// jsonKey is the top-level response key holding this stream's data.
	jsonKey string
	// single is true when the key holds one object (current) rather than an array.
	single bool
	// mapRecord flattens one raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamSpecs is the per-stream routing table. The read path is fully
// data-driven from this map.
var streamSpecs = map[string]streamSpec{
	"current": {jsonKey: "current", single: true, mapRecord: mapCurrent},
	"hourly":  {jsonKey: "hourly", mapRecord: mapHourly},
	"daily":   {jsonKey: "daily", mapRecord: mapDaily},
	"alerts":  {jsonKey: "alerts", mapRecord: mapAlert},
}

// openweatherStreams returns the connector's published catalog. Time-series
// streams use the unix `dt` timestamp as both an effective identity and the
// incremental cursor; alerts use the alert `start` time.
func openweatherStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "current",
			Description:  "Current weather conditions at the configured location(s).",
			PrimaryKey:   []string{"lat", "lon", "dt"},
			CursorFields: []string{"dt"},
			Fields:       currentFields(),
		},
		{
			Name:         "hourly",
			Description:  "Hourly weather forecast (48 hours) per location.",
			PrimaryKey:   []string{"lat", "lon", "dt"},
			CursorFields: []string{"dt"},
			Fields:       hourlyFields(),
		},
		{
			Name:         "daily",
			Description:  "Daily weather forecast (8 days) per location.",
			PrimaryKey:   []string{"lat", "lon", "dt"},
			CursorFields: []string{"dt"},
			Fields:       dailyFields(),
		},
		{
			Name:         "alerts",
			Description:  "Government weather alerts for the configured location(s).",
			PrimaryKey:   []string{"lat", "lon", "start", "event"},
			CursorFields: []string{"start"},
			Fields:       alertFields(),
		},
	}
}

func currentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "lat", Type: "number"},
		{Name: "lon", Type: "number"},
		{Name: "timezone", Type: "string"},
		{Name: "dt", Type: "integer"},
		{Name: "sunrise", Type: "integer"},
		{Name: "sunset", Type: "integer"},
		{Name: "temp", Type: "number"},
		{Name: "feels_like", Type: "number"},
		{Name: "pressure", Type: "integer"},
		{Name: "humidity", Type: "integer"},
		{Name: "dew_point", Type: "number"},
		{Name: "uvi", Type: "number"},
		{Name: "clouds", Type: "integer"},
		{Name: "visibility", Type: "integer"},
		{Name: "wind_speed", Type: "number"},
		{Name: "wind_deg", Type: "integer"},
		{Name: "wind_gust", Type: "number"},
		{Name: "weather_main", Type: "string"},
		{Name: "weather_description", Type: "string"},
		{Name: "weather_icon", Type: "string"},
	}
}

func hourlyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "lat", Type: "number"},
		{Name: "lon", Type: "number"},
		{Name: "timezone", Type: "string"},
		{Name: "dt", Type: "integer"},
		{Name: "temp", Type: "number"},
		{Name: "feels_like", Type: "number"},
		{Name: "pressure", Type: "integer"},
		{Name: "humidity", Type: "integer"},
		{Name: "dew_point", Type: "number"},
		{Name: "uvi", Type: "number"},
		{Name: "clouds", Type: "integer"},
		{Name: "visibility", Type: "integer"},
		{Name: "wind_speed", Type: "number"},
		{Name: "wind_deg", Type: "integer"},
		{Name: "wind_gust", Type: "number"},
		{Name: "pop", Type: "number"},
		{Name: "weather_main", Type: "string"},
		{Name: "weather_description", Type: "string"},
		{Name: "weather_icon", Type: "string"},
	}
}

func dailyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "lat", Type: "number"},
		{Name: "lon", Type: "number"},
		{Name: "timezone", Type: "string"},
		{Name: "dt", Type: "integer"},
		{Name: "sunrise", Type: "integer"},
		{Name: "sunset", Type: "integer"},
		{Name: "summary", Type: "string"},
		{Name: "temp_day", Type: "number"},
		{Name: "temp_min", Type: "number"},
		{Name: "temp_max", Type: "number"},
		{Name: "pressure", Type: "integer"},
		{Name: "humidity", Type: "integer"},
		{Name: "wind_speed", Type: "number"},
		{Name: "wind_deg", Type: "integer"},
		{Name: "pop", Type: "number"},
		{Name: "uvi", Type: "number"},
		{Name: "weather_main", Type: "string"},
		{Name: "weather_description", Type: "string"},
		{Name: "weather_icon", Type: "string"},
	}
}

func alertFields() []connectors.Field {
	return []connectors.Field{
		{Name: "lat", Type: "number"},
		{Name: "lon", Type: "number"},
		{Name: "timezone", Type: "string"},
		{Name: "sender_name", Type: "string"},
		{Name: "event", Type: "string"},
		{Name: "start", Type: "integer"},
		{Name: "end", Type: "integer"},
		{Name: "description", Type: "string"},
		{Name: "tags", Type: "array"},
	}
}

// mapCurrent flattens the One Call `current` object, lifting the first weather
// element into scalar weather_* columns.
func mapCurrent(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"dt":         item["dt"],
		"sunrise":    item["sunrise"],
		"sunset":     item["sunset"],
		"temp":       item["temp"],
		"feels_like": item["feels_like"],
		"pressure":   item["pressure"],
		"humidity":   item["humidity"],
		"dew_point":  item["dew_point"],
		"uvi":        item["uvi"],
		"clouds":     item["clouds"],
		"visibility": item["visibility"],
		"wind_speed": item["wind_speed"],
		"wind_deg":   item["wind_deg"],
		"wind_gust":  item["wind_gust"],
	}
	addWeather(rec, item["weather"])
	return rec
}

func mapHourly(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"dt":         item["dt"],
		"temp":       item["temp"],
		"feels_like": item["feels_like"],
		"pressure":   item["pressure"],
		"humidity":   item["humidity"],
		"dew_point":  item["dew_point"],
		"uvi":        item["uvi"],
		"clouds":     item["clouds"],
		"visibility": item["visibility"],
		"wind_speed": item["wind_speed"],
		"wind_deg":   item["wind_deg"],
		"wind_gust":  item["wind_gust"],
		"pop":        item["pop"],
	}
	addWeather(rec, item["weather"])
	return rec
}

func mapDaily(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"dt":         item["dt"],
		"sunrise":    item["sunrise"],
		"sunset":     item["sunset"],
		"summary":    item["summary"],
		"pressure":   item["pressure"],
		"humidity":   item["humidity"],
		"wind_speed": item["wind_speed"],
		"wind_deg":   item["wind_deg"],
		"pop":        item["pop"],
		"uvi":        item["uvi"],
	}
	if temp, ok := item["temp"].(map[string]any); ok {
		rec["temp_day"] = temp["day"]
		rec["temp_min"] = temp["min"]
		rec["temp_max"] = temp["max"]
	}
	addWeather(rec, item["weather"])
	return rec
}

func mapAlert(item map[string]any) connectors.Record {
	return connectors.Record{
		"sender_name": item["sender_name"],
		"event":       item["event"],
		"start":       item["start"],
		"end":         item["end"],
		"description": item["description"],
		"tags":        item["tags"],
	}
}

// addWeather lifts the first element of the One Call `weather` array into the
// scalar weather_main/description/icon columns. The array is optional.
func addWeather(rec connectors.Record, raw any) {
	arr, ok := raw.([]any)
	if !ok || len(arr) == 0 {
		return
	}
	first, ok := arr[0].(map[string]any)
	if !ok {
		return
	}
	rec["weather_main"] = first["main"]
	rec["weather_description"] = first["description"]
	rec["weather_icon"] = first["icon"]
}
