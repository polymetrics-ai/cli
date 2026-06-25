package carequalitycommission

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the CQC Syndication API resource path
// (relative to base_url), the JSON field path that holds the records array in
// the response envelope, whether the endpoint is page-incremented, and the
// record mapper that flattens its objects.
//
// Adding a stream means adding one entry to cqcStreamEndpoints plus a Stream
// definition in cqcStreams; the read path is fully data-driven from this table.
type streamEndpoint struct {
	// resource is the path segment under /public/v1 (e.g. "locations").
	resource string
	// recordsPath is the dotted JSON path to the records array in the body
	// (e.g. "locations", "providers", "inspectionAreas").
	recordsPath string
	// paginated reports whether the endpoint supports page/perPage paging. The
	// CQC list endpoints (locations, providers) page; inspection-areas does not.
	paginated bool
	// mapRecord flattens a raw CQC object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// cqcStreamEndpoints is the per-stream routing table for the CORE top-level
// streams. The substream/detail endpoints (provider_locations,
// locations_detailed, providers_detailed) require fanning out over parent IDs
// and are intentionally out of scope for this read-only core connector.
var cqcStreamEndpoints = map[string]streamEndpoint{
	"locations":        {resource: "locations", recordsPath: "locations", paginated: true, mapRecord: cqcLocationRecord},
	"providers":        {resource: "providers", recordsPath: "providers", paginated: true, mapRecord: cqcProviderRecord},
	"inspection_areas": {resource: "inspection-areas", recordsPath: "inspectionAreas", paginated: false, mapRecord: cqcInspectionAreaRecord},
}

// cqcStreams returns the connector's published stream catalog. The CQC API only
// supports full-refresh, so no incremental cursor fields are exposed.
func cqcStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "locations",
			Description:  "CQC registered locations (care homes, hospitals, GP practices, etc.). List view: id, name, postcode.",
			PrimaryKey:   []string{"locationId"},
			CursorFields: nil,
			Fields:       cqcLocationFields(),
		},
		{
			Name:         "providers",
			Description:  "CQC registered providers (the organisations that operate locations). List view: id, name.",
			PrimaryKey:   []string{"providerId"},
			CursorFields: nil,
			Fields:       cqcProviderFields(),
		},
		{
			Name:         "inspection_areas",
			Description:  "CQC inspection areas and their categories (the key questions inspections assess).",
			PrimaryKey:   []string{"inspectionAreaId"},
			CursorFields: nil,
			Fields:       cqcInspectionAreaFields(),
		},
	}
}

func cqcLocationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "locationId", Type: "string"},
		{Name: "locationName", Type: "string"},
		{Name: "postalCode", Type: "string"},
	}
}

func cqcProviderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "providerId", Type: "string"},
		{Name: "providerName", Type: "string"},
	}
}

func cqcInspectionAreaFields() []connectors.Field {
	return []connectors.Field{
		{Name: "inspectionAreaId", Type: "string"},
		{Name: "inspectionAreaName", Type: "string"},
		{Name: "inspectionAreaType", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "endDate", Type: "string"},
		{Name: "supersededBy", Type: "string"},
		{Name: "orgInspectionAreaRetirementDate", Type: "string"},
		{Name: "inspectionCategories", Type: "array"},
	}
}

func cqcLocationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"locationId":   item["locationId"],
		"locationName": item["locationName"],
		"postalCode":   item["postalCode"],
	}
}

func cqcProviderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"providerId":   item["providerId"],
		"providerName": item["providerName"],
	}
}

func cqcInspectionAreaRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"inspectionAreaId":                item["inspectionAreaId"],
		"inspectionAreaName":              item["inspectionAreaName"],
		"inspectionAreaType":              item["inspectionAreaType"],
		"status":                          item["status"],
		"endDate":                         item["endDate"],
		"supersededBy":                    item["supersededBy"],
		"orgInspectionAreaRetirementDate": item["orgInspectionAreaRetirementDate"],
		"inspectionCategories":            item["inspectionCategories"],
	}
}
