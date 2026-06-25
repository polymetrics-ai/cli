package cimis

import "polymetrics/internal/connectors"

// scope distinguishes the two flavours of CIMIS /api/data reads.
type scope string

const (
	scopeDaily  scope = "daily"
	scopeHourly scope = "hourly"
)

// streamDef describes how a published stream maps onto a CIMIS endpoint.
//
// The "daily" and "hourly" streams both read GET /api/data; they differ only in
// which configured data-item set is sent and (for the response) which Scope is
// expected. The "stations" stream reads GET /api/station, which CIMIS serves
// without an appKey and which carries no incremental cursor.
type streamDef struct {
	// resource is the endpoint path segment relative to base_url (e.g. "api/data").
	resource string
	// recordsPath is the dotted JSON path to the records array. For /api/station
	// the array is at the root key "Stations"; for /api/data the records live
	// under Data.Providers[].Records[] and are flattened in code (providersPath).
	recordsPath string
	// providersPath, when set, marks a Data.Providers[].Records[] response that
	// must be flattened across providers (the /api/data shape).
	providersPath bool
	// scope selects which data-item config key feeds the request and tags records.
	scope scope
	// mapRecord lifts a raw CIMIS object into a connectors.Record, flattening the
	// nested {Value,Qc,Unit} data-item objects.
	mapRecord func(map[string]any) connectors.Record
}

// cimisStreamDefs is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in cimisStreams.
var cimisStreamDefs = map[string]streamDef{
	"daily":    {resource: "api/data", providersPath: true, scope: scopeDaily, mapRecord: cimisDataRecord},
	"hourly":   {resource: "api/data", providersPath: true, scope: scopeHourly, mapRecord: cimisDataRecord},
	"stations": {resource: "api/station", recordsPath: "Stations", mapRecord: cimisStationRecord},
}

// cimisStreams returns the connector's published stream catalog.
func cimisStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "daily",
			Description:  "CIMIS daily weather/ET observations for the configured targets and date range.",
			PrimaryKey:   []string{"Station", "Date"},
			CursorFields: []string{"Date"},
			Fields:       cimisDataFields(),
		},
		{
			Name:         "hourly",
			Description:  "CIMIS hourly weather/ET observations for the configured targets and date range.",
			PrimaryKey:   []string{"Station", "Date", "Hour"},
			CursorFields: []string{"Date"},
			Fields:       cimisDataFields(),
		},
		{
			Name:        "stations",
			Description: "CIMIS weather station metadata (no appKey required).",
			PrimaryKey:  []string{"StationNbr"},
			Fields:      cimisStationFields(),
		},
	}
}

func cimisDataFields() []connectors.Field {
	return []connectors.Field{
		{Name: "Date", Type: "string"},
		{Name: "Julian", Type: "string"},
		{Name: "Station", Type: "string"},
		{Name: "Standard", Type: "string"},
		{Name: "ZipCodes", Type: "array"},
		{Name: "Scope", Type: "string"},
		{Name: "Hour", Type: "string"},
		// Data items arrive as nested {Value,Qc,Unit}; the mapper flattens each
		// item X into X_Value / X_Qc / X_Unit string fields alongside the raw item.
	}
}

func cimisStationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "StationNbr", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "City", Type: "string"},
		{Name: "RegionalOffice", Type: "string"},
		{Name: "County", Type: "string"},
		{Name: "ConnectDate", Type: "string"},
		{Name: "DisconnectDate", Type: "string"},
		{Name: "IsActive", Type: "string"},
		{Name: "IsEtoStation", Type: "string"},
		{Name: "Elevation", Type: "string"},
		{Name: "GroundCover", Type: "string"},
		{Name: "HmsLatitude", Type: "string"},
		{Name: "HmsLongitude", Type: "string"},
		{Name: "ZipCodes", Type: "array"},
		{Name: "SitingDesc", Type: "string"},
	}
}

// cimisDataRecord copies the scalar record fields through and flattens every
// nested data-item object {Value,Qc,Unit} into <Item>_Value/_Qc/_Unit string
// fields, while preserving the original nested object too.
func cimisDataRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{}
	for key, value := range item {
		rec[key] = value
		if obj, ok := value.(map[string]any); ok {
			if _, hasValue := obj["Value"]; hasValue {
				rec[key+"_Value"] = stringField(obj, "Value")
				rec[key+"_Qc"] = stringField(obj, "Qc")
				rec[key+"_Unit"] = stringField(obj, "Unit")
			}
		}
	}
	return rec
}

// cimisStationRecord passes station metadata through unchanged.
func cimisStationRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{}
	for key, value := range item {
		rec[key] = value
	}
	return rec
}
