package instagram

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Facebook Graph API edge it reads from
// (relative to the Instagram Business Account node), the comma-separated field
// list requested via the `fields` query param, the record mapper, and whether
// the endpoint returns a single node (e.g. the users profile) rather than a
// paginated data[] edge.
type streamEndpoint struct {
	// edge is the path segment appended after the IG user id. An empty edge
	// reads the node itself (used by the users/profile stream).
	edge string
	// fields is the value of the Graph API `fields` query parameter.
	fields string
	// mapRecord flattens a raw Graph API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// single is true when the endpoint returns one object rather than a
	// paginated {data:[...]} edge.
	single bool
}

// instagramStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in instagramStreams; the
// read path is fully data-driven from this table.
var instagramStreamEndpoints = map[string]streamEndpoint{
	"users":         {edge: "", fields: "id,username,name,biography,website,profile_picture_url,followers_count,follows_count,media_count", mapRecord: instagramUserRecord, single: true},
	"media":         {edge: "media", fields: "id,caption,media_type,media_product_type,media_url,permalink,thumbnail_url,timestamp,username,like_count,comments_count", mapRecord: instagramMediaRecord},
	"stories":       {edge: "stories", fields: "id,caption,media_type,media_product_type,media_url,permalink,thumbnail_url,timestamp,username", mapRecord: instagramStoryRecord},
	"user_insights": {edge: "insights", fields: "", mapRecord: instagramUserInsightRecord},
}

// instagramStreams returns the connector's published stream catalog.
func instagramStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "users",
			Description: "Profile information for the Instagram Business/Creator account.",
			PrimaryKey:  []string{"id"},
			Fields:      instagramUserFields(),
		},
		{
			Name:         "media",
			Description:  "Photos, videos, reels, and carousel albums published by the account.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"timestamp"},
			Fields:       instagramMediaFields(),
		},
		{
			Name:        "stories",
			Description: "Currently published (non-expired) stories for the account.",
			PrimaryKey:  []string{"id"},
			Fields:      instagramStoryFields(),
		},
		{
			Name:         "user_insights",
			Description:  "Daily account-level insight metrics (reach, follower activity).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"end_time"},
			Fields:       instagramUserInsightFields(),
		},
	}
}

func instagramUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "biography", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "profile_picture_url", Type: "string"},
		{Name: "followers_count", Type: "integer"},
		{Name: "follows_count", Type: "integer"},
		{Name: "media_count", Type: "integer"},
	}
}

func instagramMediaFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "caption", Type: "string"},
		{Name: "media_type", Type: "string"},
		{Name: "media_product_type", Type: "string"},
		{Name: "media_url", Type: "string"},
		{Name: "permalink", Type: "string"},
		{Name: "thumbnail_url", Type: "string"},
		{Name: "timestamp", Type: "timestamp"},
		{Name: "username", Type: "string"},
		{Name: "like_count", Type: "integer"},
		{Name: "comments_count", Type: "integer"},
	}
}

func instagramStoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "caption", Type: "string"},
		{Name: "media_type", Type: "string"},
		{Name: "media_product_type", Type: "string"},
		{Name: "media_url", Type: "string"},
		{Name: "permalink", Type: "string"},
		{Name: "thumbnail_url", Type: "string"},
		{Name: "timestamp", Type: "timestamp"},
		{Name: "username", Type: "string"},
	}
}

func instagramUserInsightFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "period", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "end_time", Type: "timestamp"},
		{Name: "value", Type: "integer"},
	}
}

func instagramUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"username":            item["username"],
		"name":                item["name"],
		"biography":           item["biography"],
		"website":             item["website"],
		"profile_picture_url": item["profile_picture_url"],
		"followers_count":     item["followers_count"],
		"follows_count":       item["follows_count"],
		"media_count":         item["media_count"],
	}
}

func instagramMediaRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"caption":            item["caption"],
		"media_type":         item["media_type"],
		"media_product_type": item["media_product_type"],
		"media_url":          item["media_url"],
		"permalink":          item["permalink"],
		"thumbnail_url":      item["thumbnail_url"],
		"timestamp":          item["timestamp"],
		"username":           item["username"],
		"like_count":         item["like_count"],
		"comments_count":     item["comments_count"],
	}
}

func instagramStoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"caption":            item["caption"],
		"media_type":         item["media_type"],
		"media_product_type": item["media_product_type"],
		"media_url":          item["media_url"],
		"permalink":          item["permalink"],
		"thumbnail_url":      item["thumbnail_url"],
		"timestamp":          item["timestamp"],
		"username":           item["username"],
	}
}

// instagramUserInsightRecord flattens an insights metric. The Graph API returns
// insight objects as {name, period, title, description, values:[{value,end_time}]}.
// The mapper hoists the most recent value/end_time onto the record for a flat
// row; the raw values array is preserved under "values".
func instagramUserInsightRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"period":      item["period"],
		"title":       item["title"],
		"description": item["description"],
		"values":      item["values"],
	}
	if values, ok := item["values"].([]any); ok && len(values) > 0 {
		if last, ok := values[len(values)-1].(map[string]any); ok {
			rec["value"] = last["value"]
			rec["end_time"] = last["end_time"]
		}
	}
	// Insight rows have no native id; synthesize a stable one from name+end_time.
	if rec["id"] == nil {
		name := stringField(item, "name")
		end := stringField(map[string]any{"end_time": rec["end_time"]}, "end_time")
		if name != "" {
			rec["id"] = name + "_" + end
		}
	}
	return rec
}
