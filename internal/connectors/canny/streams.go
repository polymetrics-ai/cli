package canny

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Canny API resource path (relative to
// base_url) it reads from, the JSON key holding the records array in the
// response, and the record mapper that flattens its objects.
//
// Canny list endpoints are POST calls that return {<key>:[...], hasMore:bool}.
// The api key travels in the form body, never the URL, so it is never logged.
type streamEndpoint struct {
	// resource is the Canny list endpoint path (e.g. "posts/list").
	resource string
	// recordsKey is the response JSON key holding the array (e.g. "posts").
	recordsKey string
	// paginated is false for endpoints that return everything in one response
	// (boards), so the read loop does not request a second page.
	paginated bool
	// mapRecord flattens a raw Canny object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// cannyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in cannyStreams; the read path
// is fully data-driven from this table.
var cannyStreamEndpoints = map[string]streamEndpoint{
	"boards":     {resource: "boards/list", recordsKey: "boards", paginated: false, mapRecord: cannyBoardRecord},
	"posts":      {resource: "posts/list", recordsKey: "posts", paginated: true, mapRecord: cannyPostRecord},
	"comments":   {resource: "comments/list", recordsKey: "comments", paginated: true, mapRecord: cannyCommentRecord},
	"categories": {resource: "categories/list", recordsKey: "categories", paginated: true, mapRecord: cannyCategoryRecord},
	"companies":  {resource: "companies/list", recordsKey: "companies", paginated: true, mapRecord: cannyCompanyRecord},
}

// cannyStreams returns the connector's published stream catalog. Every Canny
// object exposes a string id and an ISO-8601 `created` timestamp, so the primary
// key is ["id"] and the incremental cursor field is ["created"] where present.
func cannyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "boards",
			Description:  "Canny boards.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       cannyBoardFields(),
		},
		{
			Name:         "posts",
			Description:  "Canny posts (feature requests / feedback).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       cannyPostFields(),
		},
		{
			Name:         "comments",
			Description:  "Canny comments on posts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       cannyCommentFields(),
		},
		{
			Name:         "categories",
			Description:  "Canny board categories.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       cannyCategoryFields(),
		},
		{
			Name:         "companies",
			Description:  "Canny companies.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       cannyCompanyFields(),
		},
	}
}

func cannyBoardFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "postCount", Type: "integer"},
		{Name: "url", Type: "string"},
		{Name: "isPrivate", Type: "boolean"},
	}
}

func cannyPostFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "details", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "score", Type: "integer"},
		{Name: "commentCount", Type: "integer"},
		{Name: "url", Type: "string"},
		{Name: "eta", Type: "string"},
		{Name: "statusChangedAt", Type: "string"},
	}
}

func cannyCommentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "value", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "likeCount", Type: "integer"},
		{Name: "internal", Type: "boolean"},
		{Name: "private", Type: "boolean"},
		{Name: "parentID", Type: "string"},
	}
}

func cannyCategoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "postCount", Type: "integer"},
		{Name: "url", Type: "string"},
		{Name: "parentID", Type: "string"},
	}
}

func cannyCompanyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "memberCount", Type: "integer"},
		{Name: "monthlySpend", Type: "number"},
		{Name: "domain", Type: "string"},
	}
}

func cannyBoardRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"created":   item["created"],
		"postCount": item["postCount"],
		"url":       item["url"],
		"isPrivate": item["isPrivate"],
	}
}

func cannyPostRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"title":           item["title"],
		"details":         item["details"],
		"status":          item["status"],
		"created":         item["created"],
		"score":           item["score"],
		"commentCount":    item["commentCount"],
		"url":             item["url"],
		"eta":             item["eta"],
		"statusChangedAt": item["statusChangedAt"],
	}
}

func cannyCommentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"value":     item["value"],
		"created":   item["created"],
		"likeCount": item["likeCount"],
		"internal":  item["internal"],
		"private":   item["private"],
		"parentID":  item["parentID"],
	}
}

func cannyCategoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"created":   item["created"],
		"postCount": item["postCount"],
		"url":       item["url"],
		"parentID":  item["parentID"],
	}
}

func cannyCompanyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"created":      item["created"],
		"memberCount":  item["memberCount"],
		"monthlySpend": item["monthlySpend"],
		"domain":       item["domain"],
	}
}
