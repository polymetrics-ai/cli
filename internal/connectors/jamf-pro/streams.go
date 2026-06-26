package jamfpro

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Jamf Pro API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
// Every modern Jamf Pro API list endpoint returns {totalCount, results:[...]}.
type streamEndpoint struct {
	// resource is the API path segment under base_url (e.g. "v1/buildings").
	resource string
	// mapRecord flattens a raw Jamf Pro object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// jamfStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in jamfStreams; the read path
// is fully data-driven from this table.
var jamfStreamEndpoints = map[string]streamEndpoint{
	"buildings":   {resource: "v1/buildings", mapRecord: jamfBuildingRecord},
	"departments": {resource: "v1/departments", mapRecord: jamfDepartmentRecord},
	"categories":  {resource: "v1/categories", mapRecord: jamfCategoryRecord},
	"scripts":     {resource: "v1/scripts", mapRecord: jamfScriptRecord},
}

// jamfStreams returns the connector's published stream catalog. Jamf Pro API
// objects expose a string id and offer no incremental cursor on these core
// configuration resources, so each stream is full-refresh with primary key
// ["id"] and no cursor fields.
func jamfStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "buildings",
			Description: "Jamf Pro buildings.",
			PrimaryKey:  []string{"id"},
			Fields:      jamfBuildingFields(),
		},
		{
			Name:        "departments",
			Description: "Jamf Pro departments.",
			PrimaryKey:  []string{"id"},
			Fields:      jamfDepartmentFields(),
		},
		{
			Name:        "categories",
			Description: "Jamf Pro categories.",
			PrimaryKey:  []string{"id"},
			Fields:      jamfCategoryFields(),
		},
		{
			Name:        "scripts",
			Description: "Jamf Pro scripts.",
			PrimaryKey:  []string{"id"},
			Fields:      jamfScriptFields(),
		},
	}
}

func jamfBuildingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "streetAddress1", Type: "string"},
		{Name: "streetAddress2", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "stateProvince", Type: "string"},
		{Name: "zipPostalCode", Type: "string"},
		{Name: "country", Type: "string"},
	}
}

func jamfDepartmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

func jamfCategoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "priority", Type: "integer"},
	}
}

func jamfScriptFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "info", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "priority", Type: "string"},
		{Name: "categoryId", Type: "string"},
		{Name: "categoryName", Type: "string"},
		{Name: "osRequirements", Type: "string"},
	}
}

func jamfBuildingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"streetAddress1": item["streetAddress1"],
		"streetAddress2": item["streetAddress2"],
		"city":           item["city"],
		"stateProvince":  item["stateProvince"],
		"zipPostalCode":  item["zipPostalCode"],
		"country":        item["country"],
	}
}

func jamfDepartmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"name": item["name"],
	}
}

func jamfCategoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"name":     item["name"],
		"priority": item["priority"],
	}
}

func jamfScriptRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"info":           item["info"],
		"notes":          item["notes"],
		"priority":       item["priority"],
		"categoryId":     item["categoryId"],
		"categoryName":   item["categoryName"],
		"osRequirements": item["osRequirements"],
	}
}
