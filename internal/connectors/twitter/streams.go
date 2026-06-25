package twitter

import "polymetrics/internal/connectors"

// streamEndpoint describes how a published stream maps onto the Twitter API v2
// recent-search endpoint. Both streams read from the same endpoint
// (tweets/search/recent); they differ only in which JSON path the records live
// at (data[] for tweets, includes.users[] for authors) and the record mapper.
type streamEndpoint struct {
	// resource is the API path segment relative to base_url.
	resource string
	// recordsPath is the dotted JSON path to the record array for this stream.
	recordsPath string
	// mapRecord flattens a raw Twitter object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// twitterStreamEndpoints is the per-stream routing table. The read path is fully
// data-driven from this table.
var twitterStreamEndpoints = map[string]streamEndpoint{
	"tweets":  {resource: "tweets/search/recent", recordsPath: "data", mapRecord: twitterTweetRecord},
	"authors": {resource: "tweets/search/recent", recordsPath: "includes.users", mapRecord: twitterAuthorRecord},
}

// twitterStreams returns the connector's published stream catalog. It mirrors the
// upstream Airbyte source-twitter connector, which exposes Tweets and Authors,
// both derived from the recent-search endpoint. Twitter v2 supports only
// full_refresh for recent search, but every tweet carries created_at, which we
// surface as a cursor field for downstream incremental dedup.
func twitterStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "tweets",
			Description:  "Tweets matching the configured search query from the Twitter API v2 recent search endpoint.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       twitterTweetFields(),
		},
		{
			Name:         "authors",
			Description:  "Authors (users) of the matching tweets, expanded from the recent search endpoint.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       twitterAuthorFields(),
		},
	}
}

func twitterTweetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "text", Type: "string"},
		{Name: "author_id", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "conversation_id", Type: "string"},
		{Name: "lang", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "in_reply_to_user_id", Type: "string"},
		{Name: "possibly_sensitive", Type: "boolean"},
		{Name: "public_metrics", Type: "object"},
	}
}

func twitterAuthorFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "location", Type: "string"},
		{Name: "verified", Type: "boolean"},
		{Name: "protected", Type: "boolean"},
		{Name: "url", Type: "string"},
		{Name: "public_metrics", Type: "object"},
	}
}

func twitterTweetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"text":                item["text"],
		"author_id":           item["author_id"],
		"created_at":          item["created_at"],
		"conversation_id":     item["conversation_id"],
		"lang":                item["lang"],
		"source":              item["source"],
		"in_reply_to_user_id": item["in_reply_to_user_id"],
		"possibly_sensitive":  item["possibly_sensitive"],
		"public_metrics":      item["public_metrics"],
	}
}

func twitterAuthorRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"username":       item["username"],
		"created_at":     item["created_at"],
		"description":    item["description"],
		"location":       item["location"],
		"verified":       item["verified"],
		"protected":      item["protected"],
		"url":            item["url"],
		"public_metrics": item["public_metrics"],
	}
}
