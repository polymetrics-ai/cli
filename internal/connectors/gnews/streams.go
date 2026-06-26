package gnews

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the GNews API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its articles.
type streamEndpoint struct {
	// resource is the GNews endpoint path segment (e.g. "search").
	resource string
	// mapRecord flattens a raw GNews article object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// gnewsStreamEndpoints is the per-stream routing table. Both GNews endpoints
// return the same {totalArticles, articles:[...]} envelope and the same article
// shape, so they share a mapper. Adding a stream means adding one entry here
// plus a Stream definition in gnewsStreams.
var gnewsStreamEndpoints = map[string]streamEndpoint{
	"search":        {resource: "search", mapRecord: gnewsArticleRecord},
	"top_headlines": {resource: "top-headlines", mapRecord: gnewsArticleRecord},
}

// gnewsStreams returns the connector's published stream catalog. Every GNews
// article exposes a stable string id and an RFC3339 publishedAt timestamp, so
// the primary key is ["id"] and the incremental cursor field is ["published_at"]
// across the board.
func gnewsStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "search",
			Description:  "Keyword-based GNews article search (requires a query).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"published_at"},
			Fields:       gnewsArticleFields(),
		},
		{
			Name:         "top_headlines",
			Description:  "Trending GNews top headlines, optionally scoped to a topic.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"published_at"},
			Fields:       gnewsArticleFields(),
		},
	}
}

func gnewsArticleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "image", Type: "string"},
		{Name: "published_at", Type: "timestamp"},
		{Name: "lang", Type: "string"},
		{Name: "source_id", Type: "string"},
		{Name: "source_name", Type: "string"},
		{Name: "source_url", Type: "string"},
		{Name: "source_country", Type: "string"},
	}
}

// gnewsArticleRecord flattens a raw GNews article (with a nested "source"
// object) into a flat connectors.Record. The API field "publishedAt" is mapped
// to the snake_case "published_at" cursor field.
func gnewsArticleRecord(item map[string]any) connectors.Record {
	source := mapField(item, "source")
	return connectors.Record{
		"id":             item["id"],
		"title":          item["title"],
		"description":    item["description"],
		"content":        item["content"],
		"url":            item["url"],
		"image":          item["image"],
		"published_at":   item["publishedAt"],
		"lang":           item["lang"],
		"source_id":      source["id"],
		"source_name":    source["name"],
		"source_url":     source["url"],
		"source_country": source["country"],
	}
}

// mapField returns the nested object at key, or an empty map when absent.
func mapField(item map[string]any, key string) map[string]any {
	if nested, ok := item[key].(map[string]any); ok {
		return nested
	}
	return map[string]any{}
}
