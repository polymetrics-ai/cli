package giphy

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Giphy API resource path (relative to
// base_url) it reads from, whether that endpoint requires a search query, the
// config key that supplies the query, and the record mapper for its objects.
type streamEndpoint struct {
	// resource is the Giphy endpoint path (e.g. "gifs/search").
	resource string
	// needsQuery is true for search endpoints that require a `q` parameter.
	needsQuery bool
	// queryConfigKey is the config key whose value is sent as `q`. Falls back to
	// the generic "query" config key when empty.
	queryConfigKey string
	// mapRecord flattens a raw Giphy object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// giphyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in giphyStreams; the read path
// is fully data-driven from this table.
var giphyStreamEndpoints = map[string]streamEndpoint{
	"gif_search":     {resource: "gifs/search", needsQuery: true, queryConfigKey: "query_for_gif", mapRecord: giphyMediaRecord},
	"sticker_search": {resource: "stickers/search", needsQuery: true, queryConfigKey: "query_for_stickers", mapRecord: giphyMediaRecord},
	"clip_search":    {resource: "clips/search", needsQuery: true, queryConfigKey: "query_for_clips", mapRecord: giphyMediaRecord},
	"trending_gifs":  {resource: "gifs/trending", needsQuery: false, mapRecord: giphyMediaRecord},
}

// giphyStreams returns the connector's published stream catalog. Every Giphy
// media object exposes a string id, so the primary key is ["id"]. Giphy search
// and trending endpoints do not expose a server-side updated cursor, so these
// streams are full-refresh (no CursorFields); import_datetime is surfaced for
// downstream filtering.
func giphyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "gif_search",
			Description: "GIFs returned by the Giphy GIF search endpoint for the configured query.",
			PrimaryKey:  []string{"id"},
			Fields:      giphyMediaFields(),
		},
		{
			Name:        "sticker_search",
			Description: "Stickers returned by the Giphy sticker search endpoint for the configured query.",
			PrimaryKey:  []string{"id"},
			Fields:      giphyMediaFields(),
		},
		{
			Name:        "clip_search",
			Description: "Clips returned by the Giphy clip search endpoint for the configured query.",
			PrimaryKey:  []string{"id"},
			Fields:      giphyMediaFields(),
		},
		{
			Name:        "trending_gifs",
			Description: "Currently trending GIFs from the Giphy trending endpoint.",
			PrimaryKey:  []string{"id"},
			Fields:      giphyMediaFields(),
		},
	}
}

func giphyMediaFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "bitly_url", Type: "string"},
		{Name: "embed_url", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "rating", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "source_tld", Type: "string"},
		{Name: "content_url", Type: "string"},
		{Name: "import_datetime", Type: "string"},
		{Name: "trending_datetime", Type: "string"},
	}
}

// giphyMediaRecord flattens a raw Giphy media object (gif/sticker/clip) into a
// connectors.Record. All search and trending endpoints share the GIF object
// shape, so one mapper serves every stream.
func giphyMediaRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"type":              item["type"],
		"slug":              item["slug"],
		"url":               item["url"],
		"bitly_url":         item["bitly_url"],
		"embed_url":         item["embed_url"],
		"title":             item["title"],
		"rating":            item["rating"],
		"username":          item["username"],
		"source":            item["source"],
		"source_tld":        item["source_tld"],
		"content_url":       item["content_url"],
		"import_datetime":   item["import_datetime"],
		"trending_datetime": item["trending_datetime"],
	}
}
