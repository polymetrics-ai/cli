package metricool

import "polymetrics.ai/internal/connectors"

// dateStyle selects how the start/end date range is encoded into the request.
type dateStyle int

const (
	// dateNone means the stream takes no date range params (e.g. brands).
	dateNone dateStyle = iota
	// dateLegacy encodes start/end as YYYYMMDD (the /stats/* endpoints).
	dateLegacy
	// dateV2 encodes from/to as YYYY-MM-DDTHH:MM:SS (the /v2/* endpoints).
	dateV2
)

// streamEndpoint maps a stream name to the Metricool API resource path (relative
// to base_url), the JSON path the records live at, the date encoding style, and
// whether the endpoint is partitioned per blog. The Metricool API is not
// paginated; the natural fan-out is one request per blog_id.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "stats/instagram/posts").
	resource string
	// recordsPath is the dotted JSON path to the records array. Empty means the
	// root response is itself the array (the /stats/* shape); "data" is the
	// envelope used by the /v2/* analytics endpoints.
	recordsPath string
	// dates controls how the configured date range is encoded.
	dates dateStyle
	// perBlog indicates the endpoint is scoped by blogId (true for almost every
	// stream). brands is account-wide and ignores blogId.
	perBlog bool
	// mapRecord flattens a raw API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// metricoolStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in metricoolStreams; the read
// path is fully data-driven from this table.
var metricoolStreamEndpoints = map[string]streamEndpoint{
	"brands": {
		resource:    "admin/simpleProfiles",
		recordsPath: "",
		dates:       dateNone,
		perBlog:     false,
		mapRecord:   metricoolBrandRecord,
	},
	"instagram_posts": {
		resource:    "stats/instagram/posts",
		recordsPath: "",
		dates:       dateLegacy,
		perBlog:     true,
		mapRecord:   metricoolInstagramPostRecord,
	},
	"facebook_posts": {
		resource:    "stats/facebook/posts",
		recordsPath: "",
		dates:       dateLegacy,
		perBlog:     true,
		mapRecord:   metricoolFacebookPostRecord,
	},
	"linkedin_posts": {
		resource:    "stats/linkedin/posts",
		recordsPath: "",
		dates:       dateLegacy,
		perBlog:     true,
		mapRecord:   metricoolLinkedInPostRecord,
	},
	"tiktok_posts": {
		resource:    "v2/analytics/posts/tiktok",
		recordsPath: "data",
		dates:       dateV2,
		perBlog:     true,
		mapRecord:   metricoolTikTokPostRecord,
	},
}

// metricoolStreams returns the connector's published stream catalog.
func metricoolStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "brands",
			Description: "Metricool brand profiles (one row per connected brand/blog).",
			PrimaryKey:  []string{"id"},
			Fields:      metricoolBrandFields(),
		},
		{
			Name:        "instagram_posts",
			Description: "Instagram post analytics per brand.",
			PrimaryKey:  []string{"blogId", "postId"},
			Fields:      metricoolInstagramPostFields(),
		},
		{
			Name:        "facebook_posts",
			Description: "Facebook post analytics per brand.",
			PrimaryKey:  []string{"blogId", "postId"},
			Fields:      metricoolFacebookPostFields(),
		},
		{
			Name:        "linkedin_posts",
			Description: "LinkedIn post analytics per brand.",
			PrimaryKey:  []string{"blogId", "postId"},
			Fields:      metricoolLinkedInPostFields(),
		},
		{
			Name:        "tiktok_posts",
			Description: "TikTok post analytics per brand.",
			PrimaryKey:  []string{"blogId", "videoId"},
			Fields:      metricoolTikTokPostFields(),
		},
	}
}

func metricoolBrandFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "label", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "userId", Type: "integer"},
		{Name: "url", Type: "string"},
		{Name: "timezone", Type: "string"},
	}
}

func metricoolInstagramPostFields() []connectors.Field {
	return []connectors.Field{
		{Name: "blogId", Type: "string"},
		{Name: "postId", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "publishDate", Type: "string"},
		{Name: "text", Type: "string"},
		{Name: "interactions", Type: "number"},
		{Name: "impressions", Type: "number"},
		{Name: "reach", Type: "number"},
		{Name: "likes", Type: "number"},
		{Name: "comments", Type: "number"},
		{Name: "saved", Type: "number"},
		{Name: "url", Type: "string"},
	}
}

func metricoolFacebookPostFields() []connectors.Field {
	return []connectors.Field{
		{Name: "blogId", Type: "string"},
		{Name: "postId", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "publishDate", Type: "string"},
		{Name: "text", Type: "string"},
		{Name: "interactions", Type: "number"},
		{Name: "impressions", Type: "number"},
		{Name: "reach", Type: "number"},
		{Name: "likes", Type: "number"},
		{Name: "comments", Type: "number"},
		{Name: "shares", Type: "number"},
		{Name: "url", Type: "string"},
	}
}

func metricoolLinkedInPostFields() []connectors.Field {
	return []connectors.Field{
		{Name: "blogId", Type: "string"},
		{Name: "postId", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "publishDate", Type: "string"},
		{Name: "text", Type: "string"},
		{Name: "interactions", Type: "number"},
		{Name: "impressions", Type: "number"},
		{Name: "likes", Type: "number"},
		{Name: "comments", Type: "number"},
		{Name: "shares", Type: "number"},
		{Name: "clicks", Type: "number"},
		{Name: "url", Type: "string"},
	}
}

func metricoolTikTokPostFields() []connectors.Field {
	return []connectors.Field{
		{Name: "blogId", Type: "string"},
		{Name: "videoId", Type: "string"},
		{Name: "publishDate", Type: "string"},
		{Name: "text", Type: "string"},
		{Name: "views", Type: "number"},
		{Name: "likes", Type: "number"},
		{Name: "comments", Type: "number"},
		{Name: "shares", Type: "number"},
		{Name: "reach", Type: "number"},
		{Name: "engagement", Type: "number"},
		{Name: "url", Type: "string"},
	}
}

func metricoolBrandRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"label":    item["label"],
		"title":    item["title"],
		"userId":   item["userId"],
		"url":      item["url"],
		"timezone": item["timezone"],
	}
}

func metricoolInstagramPostRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"postId":       item["postId"],
		"type":         item["type"],
		"publishDate":  item["publishDate"],
		"text":         item["text"],
		"interactions": item["interactions"],
		"impressions":  item["impressions"],
		"reach":        item["reach"],
		"likes":        item["likes"],
		"comments":     item["comments"],
		"saved":        item["saved"],
		"url":          item["url"],
	}
}

func metricoolFacebookPostRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"postId":       item["postId"],
		"type":         item["type"],
		"publishDate":  item["publishDate"],
		"text":         item["text"],
		"interactions": item["interactions"],
		"impressions":  item["impressions"],
		"reach":        item["reach"],
		"likes":        item["likes"],
		"comments":     item["comments"],
		"shares":       item["shares"],
		"url":          item["url"],
	}
}

func metricoolLinkedInPostRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"postId":       item["postId"],
		"type":         item["type"],
		"publishDate":  item["publishDate"],
		"text":         item["text"],
		"interactions": item["interactions"],
		"impressions":  item["impressions"],
		"likes":        item["likes"],
		"comments":     item["comments"],
		"shares":       item["shares"],
		"clicks":       item["clicks"],
		"url":          item["url"],
	}
}

func metricoolTikTokPostRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"videoId":     item["videoId"],
		"publishDate": item["publishDate"],
		"text":        item["text"],
		"views":       item["views"],
		"likes":       item["likes"],
		"comments":    item["comments"],
		"shares":      item["shares"],
		"reach":       item["reach"],
		"engagement":  item["engagement"],
		"url":         item["url"],
	}
}
