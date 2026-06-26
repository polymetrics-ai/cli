package searxng

import (
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// searxngStreamSet is the set of streams the connector serves. Both stream the
// same /search endpoint; "reddit" differs only in how the query is scoped (see
// searxngQuery). The result URL is the stable primary key.
var searxngStreamSet = map[string]struct{}{
	"search": {},
	"reddit": {},
}

// searxngStreams returns the connector's published stream catalog.
func searxngStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "search",
			Description:  "General SearXNG metasearch results for config query across the instance's engines.",
			PrimaryKey:   []string{"url"},
			CursorFields: []string{"published_date"},
			Fields:       searxngResultFields(),
		},
		{
			Name:         "reddit",
			Description:  "Reddit-scoped search (site:reddit.com, optional subreddit) returning Reddit threads/comments via any general engine.",
			PrimaryKey:   []string{"url"},
			CursorFields: []string{"published_date"},
			Fields:       searxngResultFields(),
		},
	}
}

func searxngResultFields() []connectors.Field {
	return []connectors.Field{
		{Name: "url", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "engine", Type: "string"},
		{Name: "engines", Type: "string"},
		{Name: "score", Type: "number"},
		{Name: "category", Type: "string"},
		{Name: "published_date", Type: "string"},
		{Name: "thumbnail", Type: "string"},
		{Name: "stream", Type: "string"},
	}
}

// searxngResultRecord flattens a raw SearXNG result object into a flat
// connectors.Record. The "publishedDate" API field maps to the snake_case
// "published_date" cursor field, and the []engines array is joined into a
// comma-separated string for warehouse friendliness.
func searxngResultRecord(stream string, item map[string]any) connectors.Record {
	return connectors.Record{
		"url":            item["url"],
		"title":          item["title"],
		"content":        item["content"],
		"engine":         item["engine"],
		"engines":        joinAny(item["engines"]),
		"score":          item["score"],
		"category":       item["category"],
		"published_date": item["publishedDate"],
		"thumbnail":      item["thumbnail"],
		"stream":         stream,
	}
}

// joinAny renders a []any of scalars (SearXNG's "engines" list) as a
// comma-joined string. Non-slice values are returned as their string form; nil
// becomes "".
func joinAny(v any) any {
	switch t := v.(type) {
	case nil:
		return ""
	case []any:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, fmt.Sprintf("%v", e))
		}
		return strings.Join(parts, ",")
	case []string:
		return strings.Join(t, ",")
	default:
		return fmt.Sprintf("%v", v)
	}
}
