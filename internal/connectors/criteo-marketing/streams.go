package criteomarketing

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Criteo Marketing Solutions API
// resource path (relative to base_url, including the API version segment) it
// reads from, the dotted JSON path to the records array in the response, and the
// record mapper that flattens its JSONAPI objects.
type streamEndpoint struct {
	// resource is the API path segment, e.g.
	// "2024-01/marketing-solutions/ad-sets/search".
	resource string
	// method is the HTTP method used to list the resource. Criteo's search-style
	// list endpoints accept GET with offset/limit query parameters.
	method string
	// recordsPath is the dotted path to the records array, e.g. "data".
	recordsPath string
	// mapRecord flattens a raw Criteo JSONAPI object {id,type,attributes:{...}}
	// into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// criteoStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in criteoStreams; the read path
// is fully data-driven from this table.
//
// Criteo Marketing Solutions returns JSONAPI-shaped payloads:
// {"data":[{"id":"...","type":"...","attributes":{...}}]}. The mappers lift the
// id alongside the flattened attributes.
var criteoStreamEndpoints = map[string]streamEndpoint{
	"ad_sets":     {resource: "2024-01/marketing-solutions/ad-sets/search", method: "GET", recordsPath: "data", mapRecord: criteoAdSetRecord},
	"advertisers": {resource: "2024-01/marketing-solutions/advertisers", method: "GET", recordsPath: "data", mapRecord: criteoAdvertiserRecord},
	"campaigns":   {resource: "2024-01/marketing-solutions/campaigns/search", method: "GET", recordsPath: "data", mapRecord: criteoCampaignRecord},
	"audiences":   {resource: "2024-01/marketing-solutions/audiences", method: "GET", recordsPath: "data", mapRecord: criteoAudienceRecord},
	"statistics":  {resource: "2024-01/statistics/report", method: "GET", recordsPath: "Rows", mapRecord: criteoStatisticsRecord},
}

// criteoStreams returns the connector's published stream catalog. Every Criteo
// JSONAPI object exposes a string id, so the primary key is ["id"] for the
// resource streams. The statistics report is keyed by the dimensions it groups
// on (advertiser + campaign + day).
func criteoStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "ad_sets",
			Description:  "Criteo ad sets (line items) with their targeting and budget settings.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{},
			Fields:       criteoAdSetFields(),
		},
		{
			Name:         "advertisers",
			Description:  "Criteo advertiser accounts accessible to the API credentials.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{},
			Fields:       criteoAdvertiserFields(),
		},
		{
			Name:         "campaigns",
			Description:  "Criteo marketing campaigns.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{},
			Fields:       criteoCampaignFields(),
		},
		{
			Name:         "audiences",
			Description:  "Criteo audiences available for targeting.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{},
			Fields:       criteoAudienceFields(),
		},
		{
			Name:         "statistics",
			Description:  "Criteo ad spend statistics report grouped by advertiser, campaign, and day.",
			PrimaryKey:   []string{"AdvertiserId", "CampaignId", "Day"},
			CursorFields: []string{"Day"},
			Fields:       criteoStatisticsFields(),
		},
	}
}

func criteoAdSetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "advertiserId", Type: "string"},
		{Name: "datasetId", Type: "string"},
		{Name: "campaignId", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "mediaType", Type: "string"},
		{Name: "destinationEnvironment", Type: "string"},
		{Name: "objective", Type: "string"},
	}
}

func criteoAdvertiserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "timezone", Type: "string"},
	}
}

func criteoCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "advertiserId", Type: "string"},
		{Name: "objective", Type: "string"},
		{Name: "spendLimit", Type: "object"},
		{Name: "goal", Type: "string"},
	}
}

func criteoAudienceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "advertiserId", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "nbActiveUsers", Type: "integer"},
	}
}

func criteoStatisticsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "AdvertiserId", Type: "string"},
		{Name: "CampaignId", Type: "string"},
		{Name: "Day", Type: "string"},
		{Name: "Clicks", Type: "integer"},
		{Name: "Displays", Type: "integer"},
		{Name: "Spend", Type: "number"},
		{Name: "Currency", Type: "string"},
	}
}

// criteoAdSetRecord flattens a JSONAPI ad-set object into a record.
func criteoAdSetRecord(item map[string]any) connectors.Record {
	attrs := attributes(item)
	return connectors.Record{
		"id":                     item["id"],
		"type":                   item["type"],
		"name":                   attrs["name"],
		"advertiserId":           attrs["advertiserId"],
		"datasetId":              attrs["datasetId"],
		"campaignId":             attrs["campaignId"],
		"status":                 attrs["status"],
		"mediaType":              attrs["mediaType"],
		"destinationEnvironment": attrs["destinationEnvironment"],
		"objective":              attrs["objective"],
	}
}

func criteoAdvertiserRecord(item map[string]any) connectors.Record {
	attrs := attributes(item)
	return connectors.Record{
		"id":       item["id"],
		"type":     item["type"],
		"name":     attrs["name"],
		"country":  attrs["country"],
		"currency": attrs["currency"],
		"timezone": attrs["timezone"],
	}
}

func criteoCampaignRecord(item map[string]any) connectors.Record {
	attrs := attributes(item)
	return connectors.Record{
		"id":           item["id"],
		"type":         item["type"],
		"name":         attrs["name"],
		"advertiserId": attrs["advertiserId"],
		"objective":    attrs["objective"],
		"spendLimit":   attrs["spendLimit"],
		"goal":         attrs["goal"],
	}
}

func criteoAudienceRecord(item map[string]any) connectors.Record {
	attrs := attributes(item)
	return connectors.Record{
		"id":            item["id"],
		"type":          item["type"],
		"name":          attrs["name"],
		"advertiserId":  attrs["advertiserId"],
		"description":   attrs["description"],
		"nbActiveUsers": attrs["nbActiveUsers"],
	}
}

// criteoStatisticsRecord maps a statistics report row. Report rows are flat (no
// JSONAPI attributes wrapper), so fields are read directly off the row.
func criteoStatisticsRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"AdvertiserId": item["AdvertiserId"],
		"CampaignId":   item["CampaignId"],
		"Day":          item["Day"],
		"Clicks":       item["Clicks"],
		"Displays":     item["Displays"],
		"Spend":        item["Spend"],
		"Currency":     item["Currency"],
	}
}

// attributes returns the JSONAPI "attributes" sub-object of a Criteo resource, or
// an empty map when the record is flat (e.g. report rows). Falling back to the
// item itself lets a mapper read attribute-named keys whether or not the payload
// is wrapped.
func attributes(item map[string]any) map[string]any {
	if attrs, ok := item["attributes"].(map[string]any); ok {
		return attrs
	}
	return item
}
