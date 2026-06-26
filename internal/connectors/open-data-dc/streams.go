package opendatadc

import "polymetrics.ai/internal/connectors"

// streamEndpoint describes how to read a single Open Data DC (MAR 2) stream:
// the path template it requests, the dotted JSON path to the records array in
// the response, and the mapper that flattens each raw item into a Record.
type streamEndpoint struct {
	// recordsPath is the dotted path to the records array (e.g. "Result.addresses").
	recordsPath string
	// mapRecord normalizes a single raw item into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Path construction differs per
// stream (locations and units embed a path segment, ssls uses a query param) so
// the path itself is built in Read rather than stored here.
var streamEndpoints = map[string]streamEndpoint{
	"locations": {recordsPath: "Result.addresses", mapRecord: mapLocationRecord},
	"units":     {recordsPath: "Result.units", mapRecord: mapUnitRecord},
	"ssls":      {recordsPath: "Result.ssls", mapRecord: mapSslRecord},
}

// streams returns the connector's published stream catalog. The MAR API is a
// read-only address-lookup service with no incremental cursor, so streams are
// full-refresh only (empty CursorFields).
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "locations",
			Description: "MAR address/place/block lookup results for the configured location query.",
			PrimaryKey:  []string{"MarId"},
			Fields:      locationFields(),
		},
		{
			Name:        "units",
			Description: "Condo/residential units for the configured MAR id (marid).",
			PrimaryKey:  []string{"UnitNum"},
			Fields:      unitFields(),
		},
		{
			Name:        "ssls",
			Description: "Square-Suffix-Lot (SSL) records for the configured MAR id (marid).",
			PrimaryKey:  []string{"SSL"},
			Fields:      sslFields(),
		},
	}
}

func locationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "MarId", Type: "string"},
		{Name: "FullAddress", Type: "string"},
		{Name: "SSL", Type: "string"},
		{Name: "StName", Type: "string"},
		{Name: "AddrNum", Type: "string"},
		{Name: "Quadrant", Type: "string"},
		{Name: "Ward", Type: "string"},
		{Name: "Anc", Type: "string"},
		{Name: "Zipcode", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "ResidenceType", Type: "string"},
		{Name: "CensusTract", Type: "string"},
		{Name: "Latitude", Type: "number"},
		{Name: "Longitude", Type: "number"},
		{Name: "Xcoord", Type: "number"},
		{Name: "Ycoord", Type: "number"},
	}
}

func unitFields() []connectors.Field {
	return []connectors.Field{
		{Name: "UnitNum", Type: "string"},
		{Name: "MarId", Type: "string"},
		{Name: "FullAddress", Type: "string"},
		{Name: "UnitType", Type: "string"},
		{Name: "UnitSSL", Type: "string"},
		{Name: "Status", Type: "string"},
	}
}

func sslFields() []connectors.Field {
	return []connectors.Field{
		{Name: "SSL", Type: "string"},
		{Name: "MarId", Type: "string"},
		{Name: "FullAddress", Type: "string"},
		{Name: "Square", Type: "string"},
		{Name: "Lot", Type: "string"},
		{Name: "Col", Type: "string"},
		{Name: "Lot_type", Type: "string"},
	}
}

// mapLocationRecord flattens a locations item. Each item is shaped like
// {"address":{"properties":{...}}}; the useful fields live under
// address.properties, so they are lifted to the top level of the record.
func mapLocationRecord(item map[string]any) connectors.Record {
	props := locationProperties(item)
	rec := connectors.Record{}
	for _, f := range locationFields() {
		rec[f.Name] = props[f.Name]
	}
	// Preserve the distance score returned by the search ranking when present.
	if d, ok := item["distance"]; ok {
		rec["distance"] = d
	}
	return rec
}

// locationProperties digs out the address.properties object from a locations
// item, tolerating a flat item (already at properties level) as a fallback.
func locationProperties(item map[string]any) map[string]any {
	if addr, ok := item["address"].(map[string]any); ok {
		if props, ok := addr["properties"].(map[string]any); ok {
			return props
		}
	}
	if props, ok := item["properties"].(map[string]any); ok {
		return props
	}
	return item
}

func mapUnitRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{}
	for _, f := range unitFields() {
		rec[f.Name] = item[f.Name]
	}
	return rec
}

func mapSslRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{}
	for _, f := range sslFields() {
		rec[f.Name] = item[f.Name]
	}
	return rec
}
