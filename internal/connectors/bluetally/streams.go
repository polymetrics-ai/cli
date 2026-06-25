package bluetally

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the BlueTally API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the path segment under /api/v1 (e.g. "assets").
	resource string
	// mapRecord flattens a raw BlueTally object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// bluetallyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in bluetallyStreams; the read
// path is fully data-driven from this table.
var bluetallyStreamEndpoints = map[string]streamEndpoint{
	"assets":       {resource: "assets", mapRecord: bluetallyAssetRecord},
	"employees":    {resource: "employees", mapRecord: bluetallyEmployeeRecord},
	"licenses":     {resource: "licenses", mapRecord: bluetallyLicenseRecord},
	"maintenances": {resource: "maintenances", mapRecord: bluetallyMaintenanceRecord},
	"accessories":  {resource: "accessories", mapRecord: bluetallyAccessoryRecord},
}

// bluetallyStreams returns the connector's published stream catalog. Every
// BlueTally object exposes an integer id and an updated_at timestamp, so the
// primary key is ["id"] and the incremental cursor field is ["updated_at"].
func bluetallyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "assets",
			Description:  "BlueTally assets (hardware/equipment inventory).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       bluetallyAssetFields(),
		},
		{
			Name:         "employees",
			Description:  "BlueTally employees that assets and licenses are assigned to.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       bluetallyEmployeeFields(),
		},
		{
			Name:         "licenses",
			Description:  "BlueTally software licenses and their seat allocations.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       bluetallyLicenseFields(),
		},
		{
			Name:         "maintenances",
			Description:  "BlueTally asset maintenance records.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       bluetallyMaintenanceFields(),
		},
		{
			Name:         "accessories",
			Description:  "BlueTally accessories inventory.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       bluetallyAccessoryFields(),
		},
	}
}

func bluetallyAssetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "asset_id", Type: "string"},
		{Name: "asset_name", Type: "string"},
		{Name: "asset_serial", Type: "string"},
		{Name: "status_id", Type: "integer"},
		{Name: "category_id", Type: "integer"},
		{Name: "category_name", Type: "string"},
		{Name: "product_id", Type: "integer"},
		{Name: "product_name", Type: "string"},
		{Name: "location_id", Type: "integer"},
		{Name: "department_id", Type: "integer"},
		{Name: "supplier_id", Type: "integer"},
		{Name: "purchase_date", Type: "string"},
		{Name: "purchase_cost", Type: "number"},
		{Name: "currency", Type: "string"},
		{Name: "warranty_expiration_date", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func bluetallyEmployeeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "location_id", Type: "integer"},
		{Name: "department_id", Type: "integer"},
		{Name: "manager_id", Type: "integer"},
		{Name: "archived", Type: "boolean"},
		{Name: "number_of_assets", Type: "integer"},
		{Name: "number_of_accessories", Type: "integer"},
		{Name: "number_of_consumables", Type: "integer"},
		{Name: "number_of_licenses", Type: "integer"},
		{Name: "notes", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func bluetallyLicenseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "licensed_to_name", Type: "string"},
		{Name: "licensed_to_email", Type: "string"},
		{Name: "license_type", Type: "string"},
		{Name: "order_number", Type: "string"},
		{Name: "purchase_date", Type: "string"},
		{Name: "expiration_date", Type: "string"},
		{Name: "termination_date", Type: "string"},
		{Name: "purchase_cost", Type: "number"},
		{Name: "unit_cost", Type: "number"},
		{Name: "currency", Type: "string"},
		{Name: "minimum_seats", Type: "integer"},
		{Name: "number_of_seats", Type: "integer"},
		{Name: "available", Type: "integer"},
		{Name: "category_id", Type: "integer"},
		{Name: "manufacturer_id", Type: "integer"},
		{Name: "supplier_id", Type: "integer"},
		{Name: "department_id", Type: "integer"},
		{Name: "location_id", Type: "integer"},
		{Name: "notes", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func bluetallyMaintenanceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "asset_id", Type: "integer"},
		{Name: "supplier_id", Type: "integer"},
		{Name: "start_date", Type: "string"},
		{Name: "end_date", Type: "string"},
		{Name: "cost", Type: "number"},
		{Name: "notes", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func bluetallyAccessoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "model_number", Type: "string"},
		{Name: "category_id", Type: "integer"},
		{Name: "manufacturer_id", Type: "integer"},
		{Name: "supplier_id", Type: "integer"},
		{Name: "location_id", Type: "integer"},
		{Name: "department_id", Type: "integer"},
		{Name: "quantity", Type: "integer"},
		{Name: "available", Type: "integer"},
		{Name: "purchase_date", Type: "string"},
		{Name: "purchase_cost", Type: "number"},
		{Name: "currency", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func bluetallyAssetRecord(item map[string]any) connectors.Record {
	return pick(item,
		"id", "asset_id", "asset_name", "asset_serial", "status_id",
		"category_id", "category_name", "product_id", "product_name",
		"location_id", "department_id", "supplier_id", "purchase_date",
		"purchase_cost", "currency", "warranty_expiration_date", "notes",
		"created_at", "updated_at",
	)
}

func bluetallyEmployeeRecord(item map[string]any) connectors.Record {
	return pick(item,
		"id", "name", "email", "title", "location_id", "department_id",
		"manager_id", "archived", "number_of_assets", "number_of_accessories",
		"number_of_consumables", "number_of_licenses", "notes",
		"created_at", "updated_at",
	)
}

func bluetallyLicenseRecord(item map[string]any) connectors.Record {
	return pick(item,
		"id", "name", "licensed_to_name", "licensed_to_email", "license_type",
		"order_number", "purchase_date", "expiration_date", "termination_date",
		"purchase_cost", "unit_cost", "currency", "minimum_seats",
		"number_of_seats", "available", "category_id", "manufacturer_id",
		"supplier_id", "department_id", "location_id", "notes",
		"created_at", "updated_at",
	)
}

func bluetallyMaintenanceRecord(item map[string]any) connectors.Record {
	return pick(item,
		"id", "type", "name", "asset_id", "supplier_id", "start_date",
		"end_date", "cost", "notes", "created_at", "updated_at",
	)
}

func bluetallyAccessoryRecord(item map[string]any) connectors.Record {
	return pick(item,
		"id", "name", "model_number", "category_id", "manufacturer_id",
		"supplier_id", "location_id", "department_id", "quantity", "available",
		"purchase_date", "purchase_cost", "currency", "notes",
		"created_at", "updated_at",
	)
}

// pick projects the named keys from a raw BlueTally object into a Record. Keys
// absent from the source are carried through as nil, keeping the record shape
// stable and matching the declared field set.
func pick(item map[string]any, keys ...string) connectors.Record {
	out := make(connectors.Record, len(keys))
	for _, k := range keys {
		out[k] = item[k]
	}
	return out
}
