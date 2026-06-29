package ezofficeinventory

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the EZOfficeInventory API resource path
// (relative to base_url), the JSON field path holding the records array, and the
// record mapper that flattens its objects. The read path is fully data-driven
// from this table: adding a stream means adding one entry here plus a Stream
// definition in ezoStreams.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "assets.api").
	resource string
	// recordsPath is the JSON field holding the records array (e.g. "assets").
	recordsPath string
	// staticParams are fixed query parameters sent on every request.
	staticParams map[string]string
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// detailParams enriches asset-like responses with image/document/custom-field
// data, mirroring the upstream upstream manifest's request_parameters.
var detailParams = map[string]string{
	"show_image_urls":       "true",
	"show_document_urls":    "true",
	"include_custom_fields": "true",
	"show_document_details": "true",
}

// ezoStreamEndpoints is the per-stream routing table.
var ezoStreamEndpoints = map[string]streamEndpoint{
	"assets":       {resource: "assets.api", recordsPath: "assets", staticParams: detailParams, mapRecord: assetRecord},
	"inventories":  {resource: "inventory.api", recordsPath: "assets", staticParams: detailParams, mapRecord: inventoryRecord},
	"members":      {resource: "members.api", recordsPath: "members", mapRecord: memberRecord},
	"locations":    {resource: "locations/get_line_item_locations.api", recordsPath: "locations", mapRecord: locationRecord},
	"asset_stocks": {resource: "stock_assets.api", recordsPath: "assets", staticParams: detailParams, mapRecord: assetRecord},
}

// ezoStreams returns the connector's published stream catalog. Assets, inventory
// items, and stock assets are keyed by "identifier"; members and locations by
// "id" (matching the upstream manifest primary keys).
func ezoStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "assets",
			Description: "EZOfficeInventory fixed assets.",
			PrimaryKey:  []string{"identifier"},
			Fields:      assetFields(),
		},
		{
			Name:        "inventories",
			Description: "EZOfficeInventory inventory (consumable) items.",
			PrimaryKey:  []string{"identifier"},
			Fields:      assetFields(),
		},
		{
			Name:        "asset_stocks",
			Description: "EZOfficeInventory stock assets.",
			PrimaryKey:  []string{"identifier"},
			Fields:      assetFields(),
		},
		{
			Name:        "members",
			Description: "EZOfficeInventory members (users and contacts).",
			PrimaryKey:  []string{"id"},
			Fields:      memberFields(),
		},
		{
			Name:        "locations",
			Description: "EZOfficeInventory locations.",
			PrimaryKey:  []string{"id"},
			Fields:      locationFields(),
		},
	}
}

func assetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "identifier", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "asset_type", Type: "string"},
		{Name: "group_id", Type: "integer"},
		{Name: "location_id", Type: "integer"},
		{Name: "location_name", Type: "string"},
		{Name: "assigned_to_user_email", Type: "string"},
		{Name: "assigned_to_user_name", Type: "string"},
		{Name: "price", Type: "string"},
		{Name: "purchased_on", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func memberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "full_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "role_id", Type: "integer"},
		{Name: "role_name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "contact_type", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func locationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "parent_id", Type: "integer"},
		{Name: "street1", Type: "string"},
		{Name: "street2", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "zipcode", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func assetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"identifier":             item["identifier"],
		"name":                   item["name"],
		"description":            item["description"],
		"asset_type":             item["asset_type"],
		"group_id":               item["group_id"],
		"location_id":            item["location_id"],
		"location_name":          item["location_name"],
		"assigned_to_user_email": item["assigned_to_user_email"],
		"assigned_to_user_name":  item["assigned_to_user_name"],
		"price":                  item["price"],
		"purchased_on":           item["purchased_on"],
		"created_at":             item["created_at"],
		"updated_at":             item["updated_at"],
	}
}

func inventoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"identifier":    item["identifier"],
		"name":          item["name"],
		"description":   item["description"],
		"asset_type":    item["asset_type"],
		"group_id":      item["group_id"],
		"location_id":   item["location_id"],
		"location_name": item["location_name"],
		"price":         item["price"],
		"net_quantity":  item["net_quantity"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func memberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"full_name":    item["full_name"],
		"email":        item["email"],
		"role_id":      item["role_id"],
		"role_name":    item["role_name"],
		"status":       item["status"],
		"contact_type": item["contact_type"],
		"country":      item["country"],
		"created_at":   item["created_at"],
	}
}

func locationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"parent_id":   item["parent_id"],
		"street1":     item["street1"],
		"street2":     item["street2"],
		"city":        item["city"],
		"state":       item["state"],
		"zipcode":     item["zipcode"],
		"country":     item["country"],
		"status":      item["status"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}
