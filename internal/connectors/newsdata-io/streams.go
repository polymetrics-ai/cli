package newsdataio

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the NewsData.io API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the NewsData.io endpoint path segment (e.g. "latest").
	resource string
	// mapRecord flattens a raw NewsData.io object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// newsdataStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in newsdataStreams; the read
// path is fully data-driven from this table.
var newsdataStreamEndpoints = map[string]streamEndpoint{
	"latest":  {resource: "latest", mapRecord: newsdataArticleRecord},
	"crypto":  {resource: "crypto", mapRecord: newsdataArticleRecord},
	"archive": {resource: "archive", mapRecord: newsdataArticleRecord},
	"sources": {resource: "sources", mapRecord: newsdataSourceRecord},
}

// newsdataStreams returns the connector's published stream catalog. Article
// streams (latest, crypto, archive) share a primary key of article_id and an
// incremental cursor of pubDate; sources is keyed by source id.
func newsdataStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "latest",
			Description:  "Latest news articles from the NewsData.io latest endpoint.",
			PrimaryKey:   []string{"article_id"},
			CursorFields: []string{"pubDate"},
			Fields:       newsdataArticleFields(),
		},
		{
			Name:         "crypto",
			Description:  "Cryptocurrency news articles from the NewsData.io crypto endpoint.",
			PrimaryKey:   []string{"article_id"},
			CursorFields: []string{"pubDate"},
			Fields:       newsdataArticleFields(),
		},
		{
			Name:         "archive",
			Description:  "Historical news articles from the NewsData.io archive endpoint.",
			PrimaryKey:   []string{"article_id"},
			CursorFields: []string{"pubDate"},
			Fields:       newsdataArticleFields(),
		},
		{
			Name:        "sources",
			Description: "Available news sources from the NewsData.io sources endpoint.",
			PrimaryKey:  []string{"id"},
			Fields:      newsdataSourceFields(),
		},
	}
}

func newsdataArticleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "article_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "link", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "pubDate", Type: "string"},
		{Name: "image_url", Type: "string"},
		{Name: "source_id", Type: "string"},
		{Name: "source_name", Type: "string"},
		{Name: "source_url", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "creator", Type: "array"},
		{Name: "keywords", Type: "array"},
		{Name: "category", Type: "array"},
		{Name: "country", Type: "array"},
		{Name: "duplicate", Type: "boolean"},
	}
}

func newsdataSourceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "icon", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "priority", Type: "integer"},
		{Name: "language", Type: "array"},
		{Name: "category", Type: "array"},
		{Name: "country", Type: "array"},
	}
}

func newsdataArticleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"article_id":  item["article_id"],
		"title":       item["title"],
		"link":        item["link"],
		"description": item["description"],
		"content":     item["content"],
		"pubDate":     item["pubDate"],
		"image_url":   item["image_url"],
		"source_id":   item["source_id"],
		"source_name": item["source_name"],
		"source_url":  item["source_url"],
		"language":    item["language"],
		"creator":     item["creator"],
		"keywords":    item["keywords"],
		"category":    item["category"],
		"country":     item["country"],
		"duplicate":   item["duplicate"],
	}
}

func newsdataSourceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"url":         item["url"],
		"icon":        item["icon"],
		"description": item["description"],
		"priority":    item["priority"],
		"language":    item["language"],
		"category":    item["category"],
		"country":     item["country"],
	}
}
