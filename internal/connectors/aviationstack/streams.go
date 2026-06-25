package aviationstack

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the aviationstack API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects. The read path is fully data-driven from this table.
type streamEndpoint struct {
	// resource is the aviationstack list endpoint path segment (e.g. "airlines").
	resource string
	// mapRecord flattens a raw aviationstack object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in aviationStreams.
var streamEndpoints = map[string]streamEndpoint{
	"flights":   {resource: "flights", mapRecord: flightRecord},
	"airlines":  {resource: "airlines", mapRecord: airlineRecord},
	"airports":  {resource: "airports", mapRecord: airportRecord},
	"airplanes": {resource: "airplanes", mapRecord: airplaneRecord},
	"countries": {resource: "countries", mapRecord: countryRecord},
}

// aviationStreams returns the connector's published stream catalog. The reference
// data streams (airlines, airports, airplanes, countries) carry a stable string
// "id"; the live flights stream is keyed on its flight_date + flight number and
// is cursored on flight_date.
func aviationStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "flights",
			Description:  "Real-time and historical flight records.",
			PrimaryKey:   []string{"flight_date", "flight_iata"},
			CursorFields: []string{"flight_date"},
			Fields:       flightFields(),
		},
		{
			Name:        "airlines",
			Description: "Airline reference data (names, IATA/ICAO codes, country).",
			PrimaryKey:  []string{"id"},
			Fields:      airlineFields(),
		},
		{
			Name:        "airports",
			Description: "Airport reference data (names, codes, geo coordinates, timezone).",
			PrimaryKey:  []string{"id"},
			Fields:      airportFields(),
		},
		{
			Name:        "airplanes",
			Description: "Airplane reference data (registration, model, production line).",
			PrimaryKey:  []string{"id"},
			Fields:      airplaneFields(),
		},
		{
			Name:        "countries",
			Description: "Country reference data (names, codes, capital, population).",
			PrimaryKey:  []string{"id"},
			Fields:      countryFields(),
		},
	}
}

func flightFields() []connectors.Field {
	return []connectors.Field{
		{Name: "flight_date", Type: "string"},
		{Name: "flight_status", Type: "string"},
		{Name: "flight_iata", Type: "string"},
		{Name: "flight_icao", Type: "string"},
		{Name: "flight_number", Type: "string"},
		{Name: "departure_airport", Type: "string"},
		{Name: "departure_iata", Type: "string"},
		{Name: "departure_scheduled", Type: "string"},
		{Name: "arrival_airport", Type: "string"},
		{Name: "arrival_iata", Type: "string"},
		{Name: "arrival_scheduled", Type: "string"},
		{Name: "airline_name", Type: "string"},
		{Name: "airline_iata", Type: "string"},
	}
}

func airlineFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "airline_name", Type: "string"},
		{Name: "iata_code", Type: "string"},
		{Name: "icao_code", Type: "string"},
		{Name: "callsign", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "fleet_size", Type: "string"},
		{Name: "country_name", Type: "string"},
		{Name: "country_iso2", Type: "string"},
		{Name: "date_founded", Type: "string"},
	}
}

func airportFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "airport_name", Type: "string"},
		{Name: "iata_code", Type: "string"},
		{Name: "icao_code", Type: "string"},
		{Name: "latitude", Type: "string"},
		{Name: "longitude", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "gmt", Type: "string"},
		{Name: "country_name", Type: "string"},
		{Name: "country_iso2", Type: "string"},
		{Name: "city_iata_code", Type: "string"},
	}
}

func airplaneFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "registration_number", Type: "string"},
		{Name: "production_line", Type: "string"},
		{Name: "iata_type", Type: "string"},
		{Name: "model_name", Type: "string"},
		{Name: "model_code", Type: "string"},
		{Name: "icao_code_hex", Type: "string"},
		{Name: "plane_owner", Type: "string"},
		{Name: "airline_iata_code", Type: "string"},
		{Name: "plane_status", Type: "string"},
		{Name: "first_flight_date", Type: "string"},
	}
}

func countryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "country_name", Type: "string"},
		{Name: "country_iso2", Type: "string"},
		{Name: "country_iso3", Type: "string"},
		{Name: "country_iso_numeric", Type: "string"},
		{Name: "capital", Type: "string"},
		{Name: "continent", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "population", Type: "string"},
		{Name: "phone_prefix", Type: "string"},
	}
}

func flightRecord(item map[string]any) connectors.Record {
	departure, _ := item["departure"].(map[string]any)
	arrival, _ := item["arrival"].(map[string]any)
	airline, _ := item["airline"].(map[string]any)
	flight, _ := item["flight"].(map[string]any)
	return connectors.Record{
		"flight_date":         item["flight_date"],
		"flight_status":       item["flight_status"],
		"flight_iata":         nestedField(flight, "iata"),
		"flight_icao":         nestedField(flight, "icao"),
		"flight_number":       nestedField(flight, "number"),
		"departure_airport":   nestedField(departure, "airport"),
		"departure_iata":      nestedField(departure, "iata"),
		"departure_scheduled": nestedField(departure, "scheduled"),
		"arrival_airport":     nestedField(arrival, "airport"),
		"arrival_iata":        nestedField(arrival, "iata"),
		"arrival_scheduled":   nestedField(arrival, "scheduled"),
		"airline_name":        nestedField(airline, "name"),
		"airline_iata":        nestedField(airline, "iata"),
	}
}

func airlineRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"airline_name": item["airline_name"],
		"iata_code":    item["iata_code"],
		"icao_code":    item["icao_code"],
		"callsign":     item["callsign"],
		"type":         item["type"],
		"status":       item["status"],
		"fleet_size":   item["fleet_size"],
		"country_name": item["country_name"],
		"country_iso2": item["country_iso2"],
		"date_founded": item["date_founded"],
	}
}

func airportRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"airport_name":   item["airport_name"],
		"iata_code":      item["iata_code"],
		"icao_code":      item["icao_code"],
		"latitude":       item["latitude"],
		"longitude":      item["longitude"],
		"timezone":       item["timezone"],
		"gmt":            item["gmt"],
		"country_name":   item["country_name"],
		"country_iso2":   item["country_iso2"],
		"city_iata_code": item["city_iata_code"],
	}
}

func airplaneRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"registration_number": item["registration_number"],
		"production_line":     item["production_line"],
		"iata_type":           item["iata_type"],
		"model_name":          item["model_name"],
		"model_code":          item["model_code"],
		"icao_code_hex":       item["icao_code_hex"],
		"plane_owner":         item["plane_owner"],
		"airline_iata_code":   item["airline_iata_code"],
		"plane_status":        item["plane_status"],
		"first_flight_date":   item["first_flight_date"],
	}
}

func countryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"country_name":        item["country_name"],
		"country_iso2":        item["country_iso2"],
		"country_iso3":        item["country_iso3"],
		"country_iso_numeric": item["country_iso_numeric"],
		"capital":             item["capital"],
		"continent":           item["continent"],
		"currency_code":       item["currency_code"],
		"population":          item["population"],
		"phone_prefix":        item["phone_prefix"],
	}
}

// nestedField reads a key from a nested object map, returning nil when the parent
// is absent. aviationstack flights nest departure/arrival/airline/flight objects.
func nestedField(obj map[string]any, key string) any {
	if obj == nil {
		return nil
	}
	return obj[key]
}
