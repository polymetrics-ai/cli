package beamer

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Beamer API resource path (relative to
// the /v0 base) it reads from, the incremental cursor field, the cursor's
// request-parameter name (Beamer filters incremental streams by dateFrom), and
// the record mapper that flattens its objects.
//
// Beamer list endpoints return a bare JSON array of objects (the Airbyte
// manifest's record selector is the root: field_path: []), so recordsKey is ""
// for the root path across the board.
type streamEndpoint struct {
	// resource is the Beamer endpoint path segment (e.g. "nps", "posts").
	resource string
	// recordsKey is the dotted path to the array of records. Beamer returns a
	// root array, so this is "".
	recordsKey string
	// cursorParam is the request parameter used to filter records at/after the
	// incremental cursor. Empty means the stream is full-refresh only.
	cursorParam string
	// mapRecord flattens a raw Beamer object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// beamerStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in beamerStreams; the read path
// is fully data-driven from this table.
//
// The nps stream is the one verified by the upstream Airbyte source-beamer
// manifest (cursor field "date", filtered via dateFrom/dateTo). posts,
// feature-requests, and comments are the other core Beamer REST resources, all
// served as root arrays under the same /v0 base with the same page/maxResults
// pagination.
var beamerStreamEndpoints = map[string]streamEndpoint{
	"nps":              {resource: "nps", cursorParam: "dateFrom", mapRecord: beamerNPSRecord},
	"posts":            {resource: "posts", mapRecord: beamerPostRecord},
	"feature_requests": {resource: "feature-requests", mapRecord: beamerFeatureRequestRecord},
	"comments":         {resource: "comments", mapRecord: beamerCommentRecord},
}

// beamerStreams returns the connector's published stream catalog.
func beamerStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "nps",
			Description:  "Beamer Net Promoter Score survey responses.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date"},
			Fields:       beamerNPSFields(),
		},
		{
			Name:         "posts",
			Description:  "Beamer announcement posts (changelog/news entries).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date"},
			Fields:       beamerPostFields(),
		},
		{
			Name:         "feature_requests",
			Description:  "Beamer feature requests submitted by users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date"},
			Fields:       beamerFeatureRequestFields(),
		},
		{
			Name:         "comments",
			Description:  "Beamer comments on posts and feature requests.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date"},
			Fields:       beamerCommentFields(),
		},
	}
}

// beamerNPSFields mirrors the schema published by the upstream Airbyte
// source-beamer manifest.
func beamerNPSFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "date", Type: "timestamp"},
		{Name: "score", Type: "number"},
		{Name: "feedback", Type: "string"},
		{Name: "userId", Type: "string"},
		{Name: "userEmail", Type: "string"},
		{Name: "userFirstName", Type: "string"},
		{Name: "userLastName", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "os", Type: "string"},
		{Name: "browser", Type: "string"},
		{Name: "origin", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "refUrl", Type: "string"},
		{Name: "filter", Type: "string"},
	}
}

func beamerPostFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "date", Type: "timestamp"},
		{Name: "category", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "feedbackEnabled", Type: "boolean"},
		{Name: "reactionsEnabled", Type: "boolean"},
		{Name: "published", Type: "boolean"},
		{Name: "clicks", Type: "integer"},
		{Name: "views", Type: "integer"},
	}
}

func beamerFeatureRequestFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "date", Type: "timestamp"},
		{Name: "title", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "votesCount", Type: "integer"},
		{Name: "commentsCount", Type: "integer"},
		{Name: "userId", Type: "string"},
		{Name: "userEmail", Type: "string"},
		{Name: "url", Type: "string"},
	}
}

func beamerCommentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "date", Type: "timestamp"},
		{Name: "content", Type: "string"},
		{Name: "postId", Type: "string"},
		{Name: "featureRequestId", Type: "string"},
		{Name: "userId", Type: "string"},
		{Name: "userEmail", Type: "string"},
		{Name: "userFirstName", Type: "string"},
		{Name: "userLastName", Type: "string"},
	}
}

func beamerNPSRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"date":          item["date"],
		"score":         item["score"],
		"feedback":      item["feedback"],
		"userId":        item["userId"],
		"userEmail":     item["userEmail"],
		"userFirstName": item["userFirstName"],
		"userLastName":  item["userLastName"],
		"country":       item["country"],
		"city":          item["city"],
		"language":      item["language"],
		"os":            item["os"],
		"browser":       item["browser"],
		"origin":        item["origin"],
		"url":           item["url"],
		"refUrl":        item["refUrl"],
		"filter":        item["filter"],
	}
}

func beamerPostRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"date":             item["date"],
		"category":         item["category"],
		"title":            item["title"],
		"content":          item["content"],
		"url":              item["url"],
		"feedbackEnabled":  item["feedbackEnabled"],
		"reactionsEnabled": item["reactionsEnabled"],
		"published":        item["published"],
		"clicks":           item["clicks"],
		"views":            item["views"],
	}
}

func beamerFeatureRequestRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"date":          item["date"],
		"title":         item["title"],
		"content":       item["content"],
		"status":        item["status"],
		"votesCount":    item["votesCount"],
		"commentsCount": item["commentsCount"],
		"userId":        item["userId"],
		"userEmail":     item["userEmail"],
		"url":           item["url"],
	}
}

func beamerCommentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"date":             item["date"],
		"content":          item["content"],
		"postId":           item["postId"],
		"featureRequestId": item["featureRequestId"],
		"userId":           item["userId"],
		"userEmail":        item["userEmail"],
		"userFirstName":    item["userFirstName"],
		"userLastName":     item["userLastName"],
	}
}
