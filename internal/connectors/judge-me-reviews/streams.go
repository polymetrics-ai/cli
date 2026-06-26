package judgemereviews

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Judge.me API resource path (relative
// to base_url), the JSON key under which its array of objects lives, and the
// record mapper that flattens those objects into connectors.Records.
type streamEndpoint struct {
	// resource is the list endpoint path segment (e.g. "reviews").
	resource string
	// recordsKey is the top-level JSON key holding the array (e.g. "reviews").
	recordsKey string
	// mapRecord flattens a raw Judge.me object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams(); the read path is fully
// data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"reviews":  {resource: "reviews", recordsKey: "reviews", mapRecord: reviewRecord},
	"products": {resource: "products", recordsKey: "products", mapRecord: productRecord},
	"widgets":  {resource: "widgets", recordsKey: "widgets", mapRecord: widgetRecord},
}

// streams returns the connector's published stream catalog. Judge.me objects
// expose a numeric id and a created_at timestamp, so the primary key is ["id"]
// and the incremental cursor field is ["created_at"].
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "reviews",
			Description:  "Product reviews collected by Judge.me for the shop.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       reviewFields(),
		},
		{
			Name:         "products",
			Description:  "Products tracked by Judge.me, with aggregate review stats.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       productFields(),
		},
		{
			Name:         "widgets",
			Description:  "Judge.me review widgets configured for the shop.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       widgetFields(),
		},
	}
}

func reviewFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "body", Type: "string"},
		{Name: "rating", Type: "integer"},
		{Name: "product_external_id", Type: "string"},
		{Name: "reviewer_id", Type: "integer"},
		{Name: "reviewer_name", Type: "string"},
		{Name: "reviewer_email", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "curated", Type: "string"},
		{Name: "published", Type: "boolean"},
		{Name: "hidden", Type: "boolean"},
		{Name: "verified", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func productFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "external_id", Type: "string"},
		{Name: "handle", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func widgetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "widget_type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func reviewRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":                  item["id"],
		"title":               item["title"],
		"body":                item["body"],
		"rating":              item["rating"],
		"product_external_id": item["product_external_id"],
		"source":              item["source"],
		"curated":             item["curated"],
		"published":           item["published"],
		"hidden":              item["hidden"],
		"verified":            item["verified"],
		"created_at":          item["created_at"],
		"updated_at":          item["updated_at"],
	}
	// Judge.me nests reviewer details under a "reviewer" object; flatten the
	// useful identity fields onto the record so the stream is fully tabular.
	if reviewer, ok := item["reviewer"].(map[string]any); ok {
		rec["reviewer_id"] = reviewer["id"]
		rec["reviewer_name"] = reviewer["name"]
		rec["reviewer_email"] = reviewer["email"]
	}
	return rec
}

func productRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"external_id": item["external_id"],
		"handle":      item["handle"],
		"title":       item["title"],
		"url":         item["url"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func widgetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"widget_type": item["widget_type"],
		"status":      item["status"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}
