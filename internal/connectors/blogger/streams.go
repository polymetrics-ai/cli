package blogger

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to how it is read from the Blogger API.
type streamEndpoint struct {
	// path returns the API path (relative to base_url) for the stream given the
	// resolved blog id. Blogger paths are blog-scoped (e.g.
	// blogs/{blogId}/posts).
	path func(blogID string) string
	// single indicates the endpoint returns a single resource object rather than
	// a paginated {items:[...]} list (e.g. the blog itself).
	single bool
	// mapRecord flattens a raw Blogger object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// bloggerStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in bloggerStreams.
var bloggerStreamEndpoints = map[string]streamEndpoint{
	"blogs": {
		path:      func(blogID string) string { return "blogs/" + blogID },
		single:    true,
		mapRecord: bloggerBlogRecord,
	},
	"posts": {
		path:      func(blogID string) string { return "blogs/" + blogID + "/posts" },
		mapRecord: bloggerPostRecord,
	},
	"pages": {
		path:      func(blogID string) string { return "blogs/" + blogID + "/pages" },
		mapRecord: bloggerPageRecord,
	},
	"comments": {
		path:      func(blogID string) string { return "blogs/" + blogID + "/comments" },
		mapRecord: bloggerCommentRecord,
	},
}

// bloggerStreams returns the connector's published stream catalog. Every Blogger
// resource exposes a string id and an updated/published RFC3339 timestamp, so
// the primary key is ["id"] and the incremental cursor field is ["updated"].
func bloggerStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "blogs",
			Description:  "The Blogger blog identified by blog_id.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       bloggerBlogFields(),
		},
		{
			Name:         "posts",
			Description:  "Posts published to the blog.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       bloggerPostFields(),
		},
		{
			Name:         "pages",
			Description:  "Static pages of the blog.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       bloggerPageFields(),
		},
		{
			Name:         "comments",
			Description:  "Comments across the blog's posts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       bloggerCommentFields(),
		},
	}
}

func bloggerBlogFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "kind", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "published", Type: "string"},
		{Name: "updated", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "posts_total", Type: "integer"},
		{Name: "pages_total", Type: "integer"},
	}
}

func bloggerPostFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "kind", Type: "string"},
		{Name: "blog_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "published", Type: "string"},
		{Name: "updated", Type: "string"},
		{Name: "author_id", Type: "string"},
		{Name: "author_display_name", Type: "string"},
		{Name: "replies_total", Type: "integer"},
		{Name: "status", Type: "string"},
	}
}

func bloggerPageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "kind", Type: "string"},
		{Name: "blog_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "published", Type: "string"},
		{Name: "updated", Type: "string"},
		{Name: "author_id", Type: "string"},
		{Name: "author_display_name", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func bloggerCommentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "kind", Type: "string"},
		{Name: "post_id", Type: "string"},
		{Name: "blog_id", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "published", Type: "string"},
		{Name: "updated", Type: "string"},
		{Name: "author_id", Type: "string"},
		{Name: "author_display_name", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func bloggerBlogRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"kind":        item["kind"],
		"name":        item["name"],
		"description": item["description"],
		"published":   item["published"],
		"updated":     item["updated"],
		"url":         item["url"],
		"posts_total": nestedField(item, "posts", "totalItems"),
		"pages_total": nestedField(item, "pages", "totalItems"),
	}
}

func bloggerPostRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"kind":                item["kind"],
		"blog_id":             nestedField(item, "blog", "id"),
		"title":               item["title"],
		"content":             item["content"],
		"url":                 item["url"],
		"published":           item["published"],
		"updated":             item["updated"],
		"author_id":           nestedField(item, "author", "id"),
		"author_display_name": nestedField(item, "author", "displayName"),
		"replies_total":       nestedField(item, "replies", "totalItems"),
		"status":              item["status"],
	}
}

func bloggerPageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"kind":                item["kind"],
		"blog_id":             nestedField(item, "blog", "id"),
		"title":               item["title"],
		"content":             item["content"],
		"url":                 item["url"],
		"published":           item["published"],
		"updated":             item["updated"],
		"author_id":           nestedField(item, "author", "id"),
		"author_display_name": nestedField(item, "author", "displayName"),
		"status":              item["status"],
	}
}

func bloggerCommentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"kind":                item["kind"],
		"post_id":             nestedField(item, "post", "id"),
		"blog_id":             nestedField(item, "blog", "id"),
		"content":             item["content"],
		"published":           item["published"],
		"updated":             item["updated"],
		"author_id":           nestedField(item, "author", "id"),
		"author_display_name": nestedField(item, "author", "displayName"),
		"status":              item["status"],
	}
}

// nestedField reads item[outer][inner], returning nil when any hop is absent.
func nestedField(item map[string]any, outer, inner string) any {
	sub, ok := item[outer].(map[string]any)
	if !ok {
		return nil
	}
	return sub[inner]
}
