package newsdata

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the NewsData.io API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects. The newsRecord mapper is shared by the article-shaped streams.
type streamEndpoint struct {
	// resource is the NewsData.io endpoint path segment (e.g. "latest").
	resource string
	// paginated is true for endpoints that return a nextPage cursor token. The
	// sources endpoint returns a single unpaginated list.
	paginated bool
	// mapRecord flattens a raw NewsData.io object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// newsdataStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in newsdataStreams; the read
// path is fully data-driven from this table.
var newsdataStreamEndpoints = map[string]streamEndpoint{
	"latest":  {resource: "latest", paginated: true, mapRecord: newsRecord},
	"crypto":  {resource: "crypto", paginated: true, mapRecord: newsRecord},
	"sources": {resource: "sources", paginated: false, mapRecord: sourceRecord},
}

// newsdataStreams returns the connector's published stream catalog. Article
// streams (latest, crypto) key on article_id and use pubDate as their cursor.
// The sources stream keys on the source id.
func newsdataStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "latest",
			Description:  "Latest news articles from NewsData.io.",
			PrimaryKey:   []string{"article_id"},
			CursorFields: []string{"pubDate"},
			Fields:       newsFields(),
		},
		{
			Name:         "crypto",
			Description:  "Cryptocurrency-related news articles from NewsData.io.",
			PrimaryKey:   []string{"article_id"},
			CursorFields: []string{"pubDate"},
			Fields:       newsFields(),
		},
		{
			Name:        "sources",
			Description: "News sources available in NewsData.io.",
			PrimaryKey:  []string{"id"},
			Fields:      sourceFields(),
		},
	}
}

func newsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "article_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "link", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "pubDate", Type: "string"},
		{Name: "image_url", Type: "string"},
		{Name: "source_id", Type: "string"},
		{Name: "source_priority", Type: "integer"},
		{Name: "language", Type: "string"},
		{Name: "creator", Type: "array"},
		{Name: "keywords", Type: "array"},
		{Name: "category", Type: "array"},
		{Name: "country", Type: "array"},
	}
}

func sourceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "icon", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "category", Type: "array"},
		{Name: "language", Type: "array"},
		{Name: "country", Type: "array"},
	}
}

func newsRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"article_id":      item["article_id"],
		"title":           item["title"],
		"link":            item["link"],
		"description":     item["description"],
		"content":         item["content"],
		"pubDate":         item["pubDate"],
		"image_url":       item["image_url"],
		"source_id":       item["source_id"],
		"source_priority": item["source_priority"],
		"language":        item["language"],
		"creator":         item["creator"],
		"keywords":        item["keywords"],
		"category":        item["category"],
		"country":         item["country"],
	}
}

func sourceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"url":         item["url"],
		"icon":        item["icon"],
		"description": item["description"],
		"category":    item["category"],
		"language":    item["language"],
		"country":     item["country"],
	}
}
