package openaq

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the OpenAQ v3 resource path (relative to
// base_url) it reads from, plus the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the OpenAQ list endpoint path segment (e.g. "countries").
	resource string
	// mapRecord flattens a raw OpenAQ result object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// openaqStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in openaqStreams; the read path
// is fully data-driven from this table. All OpenAQ v3 list endpoints share the
// {meta:{...,found},results:[...]} envelope and page/limit pagination.
var openaqStreamEndpoints = map[string]streamEndpoint{
	"countries":     {resource: "countries", mapRecord: openaqCountryRecord},
	"parameters":    {resource: "parameters", mapRecord: openaqParameterRecord},
	"locations":     {resource: "locations", mapRecord: openaqLocationRecord},
	"instruments":   {resource: "instruments", mapRecord: openaqInstrumentRecord},
	"manufacturers": {resource: "manufacturers", mapRecord: openaqManufacturerRecord},
}

// openaqStreams returns the connector's published stream catalog. OpenAQ v3
// reference resources are full-refresh (no incremental cursor), keyed by id.
func openaqStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "countries",
			Description: "Countries for which OpenAQ has air quality data.",
			PrimaryKey:  []string{"id"},
			Fields:      openaqCountryFields(),
		},
		{
			Name:        "parameters",
			Description: "Air quality parameters measured by OpenAQ (e.g. pm25, o3).",
			PrimaryKey:  []string{"id"},
			Fields:      openaqParameterFields(),
		},
		{
			Name:        "locations",
			Description: "Monitoring locations / stations reporting to OpenAQ.",
			PrimaryKey:  []string{"id"},
			Fields:      openaqLocationFields(),
		},
		{
			Name:        "instruments",
			Description: "Instruments / devices used by OpenAQ monitoring locations.",
			PrimaryKey:  []string{"id"},
			Fields:      openaqInstrumentFields(),
		},
		{
			Name:        "manufacturers",
			Description: "Manufacturers of OpenAQ monitoring instruments.",
			PrimaryKey:  []string{"id"},
			Fields:      openaqManufacturerFields(),
		},
	}
}

func openaqCountryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "code", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "datetimeFirst", Type: "string"},
		{Name: "datetimeLast", Type: "string"},
		{Name: "parameters", Type: "array"},
	}
}

func openaqParameterFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "units", Type: "string"},
		{Name: "displayName", Type: "string"},
		{Name: "description", Type: "string"},
	}
}

func openaqLocationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "locality", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "country", Type: "object"},
		{Name: "owner", Type: "object"},
		{Name: "provider", Type: "object"},
		{Name: "coordinates", Type: "object"},
		{Name: "sensors", Type: "array"},
		{Name: "isMobile", Type: "boolean"},
		{Name: "isMonitor", Type: "boolean"},
		{Name: "datetimeFirst", Type: "object"},
		{Name: "datetimeLast", Type: "object"},
	}
}

func openaqInstrumentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "isMonitor", Type: "boolean"},
		{Name: "manufacturer", Type: "object"},
	}
}

func openaqManufacturerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "instruments", Type: "array"},
	}
}

func openaqCountryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"code":          item["code"],
		"name":          item["name"],
		"datetimeFirst": item["datetimeFirst"],
		"datetimeLast":  item["datetimeLast"],
		"parameters":    item["parameters"],
	}
}

func openaqParameterRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"units":       item["units"],
		"displayName": item["displayName"],
		"description": item["description"],
	}
}

func openaqLocationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"locality":      item["locality"],
		"timezone":      item["timezone"],
		"country":       item["country"],
		"owner":         item["owner"],
		"provider":      item["provider"],
		"coordinates":   item["coordinates"],
		"sensors":       item["sensors"],
		"isMobile":      item["isMobile"],
		"isMonitor":     item["isMonitor"],
		"datetimeFirst": item["datetimeFirst"],
		"datetimeLast":  item["datetimeLast"],
	}
}

func openaqInstrumentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"isMonitor":    item["isMonitor"],
		"manufacturer": item["manufacturer"],
	}
}

func openaqManufacturerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"instruments": item["instruments"],
	}
}
