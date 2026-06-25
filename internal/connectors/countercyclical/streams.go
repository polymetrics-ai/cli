package countercyclical

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Countercyclical API resource path
// (relative to base_url) it reads from, and the record mapper that selects the
// published subset of fields for its objects.
type streamEndpoint struct {
	// resource is the API path segment, e.g. "investments".
	resource string
	// mapRecord projects a raw API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. The read path is fully
// data-driven from this table; adding a stream means adding one entry here plus
// a Stream definition in streamCatalog.
//
// Source: the official manifest-only Airbyte connector
// (source-countercyclical) defines exactly these three streams, each a GET on
// /<resource> returning a root-level JSON array (DpathExtractor field_path: [])
// with primary key id and no pagination.
var streamEndpoints = map[string]streamEndpoint{
	"investments": {resource: "investments", mapRecord: investmentRecord},
	"valuations":  {resource: "valuations", mapRecord: valuationRecord},
	"memos":       {resource: "memos", mapRecord: memoRecord},
}

// streamCatalog returns the connector's published stream catalog. Every object
// exposes a string id (primary key) and updatedAt/createdAt timestamps. The
// upstream connector advertises full-refresh only (no incremental cursor), so
// CursorFields is left empty.
func streamCatalog() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "investments",
			Description: "Countercyclical investments (companies tracked by the investment team).",
			PrimaryKey:  []string{"id"},
			Fields:      investmentFields(),
		},
		{
			Name:        "valuations",
			Description: "Countercyclical valuation models.",
			PrimaryKey:  []string{"id"},
			Fields:      valuationFields(),
		},
		{
			Name:        "memos",
			Description: "Countercyclical research memos and documents.",
			PrimaryKey:  []string{"id"},
			Fields:      memoFields(),
		},
	}
}

func investmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "editedName", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "tickerSymbol", Type: "string"},
		{Name: "exchange", Type: "string"},
		{Name: "cik", Type: "string"},
		{Name: "figi", Type: "string"},
		{Name: "lei", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "sector", Type: "string"},
		{Name: "industry", Type: "string"},
		{Name: "marketType", Type: "string"},
		{Name: "financingType", Type: "string"},
		{Name: "employees", Type: "integer"},
		{Name: "website", Type: "string"},
		{Name: "isArchived", Type: "boolean"},
		{Name: "isFavorite", Type: "boolean"},
		{Name: "isLocked", Type: "boolean"},
		{Name: "visibility", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func valuationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "delineation", Type: "string"},
		{Name: "discountRate", Type: "number"},
		{Name: "growthMetric", Type: "string"},
		{Name: "growthRate", Type: "number"},
		{Name: "terminalRate", Type: "number"},
		{Name: "terminalPeriod", Type: "string"},
		{Name: "startingQuarter", Type: "integer"},
		{Name: "startingYear", Type: "integer"},
		{Name: "endingQuarter", Type: "integer"},
		{Name: "endingYear", Type: "integer"},
		{Name: "shareToken", Type: "string"},
		{Name: "isFavorite", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func memoFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "body", Type: "string"},
		{Name: "documentType", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "favorited", Type: "boolean"},
		{Name: "locked", Type: "boolean"},
		{Name: "publiclyVisible", Type: "boolean"},
		{Name: "sourcesVisible", Type: "boolean"},
		{Name: "tocVisible", Type: "boolean"},
		{Name: "bannerVisible", Type: "boolean"},
		{Name: "views", Type: "integer"},
		{Name: "emoji", Type: "string"},
		{Name: "backgroundColor", Type: "string"},
		{Name: "foregroundColor", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func investmentRecord(item map[string]any) connectors.Record {
	return projectFields(item, []string{
		"id", "name", "editedName", "description", "tickerSymbol", "exchange",
		"cik", "figi", "lei", "country", "sector", "industry", "marketType",
		"financingType", "employees", "website", "isArchived", "isFavorite",
		"isLocked", "visibility", "createdAt", "updatedAt",
	})
}

func valuationRecord(item map[string]any) connectors.Record {
	return projectFields(item, []string{
		"id", "name", "description", "status", "delineation", "discountRate",
		"growthMetric", "growthRate", "terminalRate", "terminalPeriod",
		"startingQuarter", "startingYear", "endingQuarter", "endingYear",
		"shareToken", "isFavorite", "createdAt", "updatedAt",
	})
}

func memoRecord(item map[string]any) connectors.Record {
	return projectFields(item, []string{
		"id", "title", "body", "documentType", "archived", "favorited", "locked",
		"publiclyVisible", "sourcesVisible", "tocVisible", "bannerVisible",
		"views", "emoji", "backgroundColor", "foregroundColor", "createdAt",
		"updatedAt",
	})
}

// projectFields copies the allow-listed keys from a raw object into a Record,
// preserving the API's camelCase field names. Keys absent from the source are
// emitted as nil so the record shape stays stable across rows.
func projectFields(item map[string]any, keys []string) connectors.Record {
	out := make(connectors.Record, len(keys))
	for _, k := range keys {
		out[k] = item[k]
	}
	return out
}
