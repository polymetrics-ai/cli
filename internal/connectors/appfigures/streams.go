package appfigures

import "polymetrics/internal/connectors"

// shape describes how a stream's records are laid out in the Appfigures JSON
// response so the read loop can extract them generically.
type shape int

const (
	// shapeArrayAtPath: records live in an array at recordsPath (e.g. reviews
	// returns {"reviews":[...]}), and the endpoint is page-paginated.
	shapeArrayAtPath shape = iota
	// shapeKeyedObject: the body is a JSON object keyed by id whose values are
	// the records (e.g. products/mine returns {"<id>":{...}}). A single request
	// returns the whole set; there is no pagination.
	shapeKeyedObject
)

// streamEndpoint maps a stream name to the Appfigures API resource path
// (relative to base_url), the response shape, the array path within the body
// (for shapeArrayAtPath), and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Appfigures path segment (e.g. "reviews", "products/mine").
	resource string
	// shape selects the extraction + pagination strategy.
	shape shape
	// recordsPath is the dotted path to the records array for shapeArrayAtPath
	// (e.g. "reviews"); ignored for shapeKeyedObject.
	recordsPath string
	// mapRecord flattens a raw Appfigures object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// appfiguresStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in appfiguresStreams; the
// read path is fully data-driven from this table.
var appfiguresStreamEndpoints = map[string]streamEndpoint{
	"reviews":    {resource: "reviews", shape: shapeArrayAtPath, recordsPath: "reviews", mapRecord: appfiguresReviewRecord},
	"products":   {resource: "products/mine", shape: shapeKeyedObject, mapRecord: appfiguresProductRecord},
	"sales":      {resource: "reports/sales", shape: shapeKeyedObject, mapRecord: appfiguresSalesRecord},
	"ratings":    {resource: "reports/ratings", shape: shapeKeyedObject, mapRecord: appfiguresRatingsRecord},
	"categories": {resource: "data/categories", shape: shapeKeyedObject, mapRecord: appfiguresCategoryRecord},
}

// appfiguresStreams returns the connector's published stream catalog.
func appfiguresStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "reviews",
			Description: "App Store and Google Play reviews of your products.",
			PrimaryKey:  []string{"id"},
			Fields:      appfiguresReviewFields(),
		},
		{
			Name:        "products",
			Description: "Products (apps) in your Appfigures account.",
			PrimaryKey:  []string{"id"},
			Fields:      appfiguresProductFields(),
		},
		{
			Name:        "sales",
			Description: "Sales report aggregates from the Appfigures reports/sales endpoint.",
			Fields:      appfiguresSalesFields(),
		},
		{
			Name:        "ratings",
			Description: "Ratings report aggregates from the Appfigures reports/ratings endpoint.",
			Fields:      appfiguresRatingsFields(),
		},
		{
			Name:        "categories",
			Description: "Store categories reference data.",
			PrimaryKey:  []string{"id"},
			Fields:      appfiguresCategoryFields(),
		},
	}
}

func appfiguresReviewFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "product", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "review", Type: "string"},
		{Name: "author", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "stars", Type: "number"},
		{Name: "iso", Type: "string"},
		{Name: "has_response", Type: "boolean"},
		{Name: "weight", Type: "integer"},
	}
}

func appfiguresProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "developer", Type: "string"},
		{Name: "vendor_identifier", Type: "string"},
		{Name: "ref_no", Type: "string"},
		{Name: "sku", Type: "string"},
		{Name: "store", Type: "string"},
		{Name: "store_id", Type: "integer"},
		{Name: "added", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func appfiguresSalesFields() []connectors.Field {
	return []connectors.Field{
		{Name: "date", Type: "string"},
		{Name: "downloads", Type: "integer"},
		{Name: "updates", Type: "integer"},
		{Name: "revenue", Type: "string"},
		{Name: "returns", Type: "integer"},
		{Name: "net_downloads", Type: "integer"},
		{Name: "promos", Type: "integer"},
	}
}

func appfiguresRatingsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "date", Type: "string"},
		{Name: "stars", Type: "number"},
		{Name: "breakdown", Type: "string"},
		{Name: "average", Type: "number"},
	}
}

func appfiguresCategoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "store", Type: "string"},
		{Name: "device", Type: "string"},
		{Name: "subtype", Type: "string"},
	}
}

func appfiguresReviewRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"product":        item["product"],
		"title":          item["title"],
		"review":         item["review"],
		"author":         item["author"],
		"version":        item["version"],
		"date":           item["date"],
		"stars":          item["stars"],
		"iso":            item["iso"],
		"has_response":   item["has_response"],
		"weight":         item["weight"],
		"original_title": item["original_title"],
	}
}

func appfiguresProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"developer":         item["developer"],
		"vendor_identifier": item["vendor_identifier"],
		"ref_no":            item["ref_no"],
		"sku":               item["sku"],
		"store":             item["store"],
		"store_id":          item["store_id"],
		"added":             item["added"],
		"updated":           item["updated"],
	}
}

func appfiguresSalesRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"date":          item["date"],
		"downloads":     item["downloads"],
		"updates":       item["updates"],
		"revenue":       item["revenue"],
		"returns":       item["returns"],
		"net_downloads": item["net_downloads"],
		"promos":        item["promos"],
	}
}

func appfiguresRatingsRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"date":      item["date"],
		"stars":     item["stars"],
		"breakdown": item["breakdown"],
		"average":   item["average"],
	}
}

func appfiguresCategoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"name":    item["name"],
		"store":   item["store"],
		"device":  item["device"],
		"subtype": item["subtype"],
	}
}
