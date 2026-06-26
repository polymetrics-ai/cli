package newsapi

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the News API resource path (relative to
// base_url), the JSON path holding the record array in the response, the record
// mapper, and whether the endpoint is paginated. The read path is data-driven
// from this table.
type streamEndpoint struct {
	// resource is the path segment under base_url (e.g. "everything").
	resource string
	// recordsPath is the dotted JSON path to the array of records.
	recordsPath string
	// paginated indicates the endpoint supports page/pageSize pagination.
	paginated bool
	// mapRecord flattens a raw News API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams().
var streamEndpoints = map[string]streamEndpoint{
	"everything":    {resource: "everything", recordsPath: "articles", paginated: true, mapRecord: articleRecord},
	"top_headlines": {resource: "top-headlines", recordsPath: "articles", paginated: true, mapRecord: articleRecord},
	"sources":       {resource: "top-headlines/sources", recordsPath: "sources", paginated: false, mapRecord: sourceRecord},
}

// streams returns the connector's published stream catalog. Articles have no
// stable id from the API, so the URL is the primary key and publishedAt is the
// incremental cursor. Sources are keyed by their id and are non-incremental.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "everything",
			Description:  "Articles from the News API /v2/everything search endpoint.",
			PrimaryKey:   []string{"url"},
			CursorFields: []string{"published_at"},
			Fields:       articleFields(),
		},
		{
			Name:         "top_headlines",
			Description:  "Breaking news headlines from the News API /v2/top-headlines endpoint.",
			PrimaryKey:   []string{"url"},
			CursorFields: []string{"published_at"},
			Fields:       articleFields(),
		},
		{
			Name:        "sources",
			Description: "News sources/publishers from the News API /v2/top-headlines/sources endpoint.",
			PrimaryKey:  []string{"id"},
			Fields:      sourceFields(),
		},
	}
}

func articleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "source_id", Type: "string"},
		{Name: "source_name", Type: "string"},
		{Name: "author", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "url_to_image", Type: "string"},
		{Name: "published_at", Type: "timestamp"},
		{Name: "content", Type: "string"},
	}
}

func sourceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "country", Type: "string"},
	}
}

// articleRecord flattens a News API article object. The nested {source:{id,name}}
// is hoisted to source_id/source_name; the API's camelCase keys are normalized to
// snake_case to match the published field schema.
func articleRecord(item map[string]any) connectors.Record {
	sourceID, sourceName := any(nil), any(nil)
	if src, ok := item["source"].(map[string]any); ok {
		sourceID = src["id"]
		sourceName = src["name"]
	}
	return connectors.Record{
		"source_id":    sourceID,
		"source_name":  sourceName,
		"author":       item["author"],
		"title":        item["title"],
		"description":  item["description"],
		"url":          item["url"],
		"url_to_image": item["urlToImage"],
		"published_at": item["publishedAt"],
		"content":      item["content"],
	}
}

func sourceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"url":         item["url"],
		"category":    item["category"],
		"language":    item["language"],
		"country":     item["country"],
	}
}
