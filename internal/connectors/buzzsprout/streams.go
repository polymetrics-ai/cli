package buzzsprout

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Buzzsprout API resource and how its
// objects are flattened into connectors.Record values.
type streamEndpoint struct {
	// accountLevel is true for endpoints under /api (not scoped to a podcast),
	// false for endpoints under /api/{podcast_id}.
	accountLevel bool
	// resource is the path segment after the account/podcast prefix, ending in
	// .json (e.g. "episodes.json").
	resource string
	// mapRecord flattens a raw Buzzsprout object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// buzzsproutStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in buzzsproutStreams.
var buzzsproutStreamEndpoints = map[string]streamEndpoint{
	"episodes": {accountLevel: false, resource: "episodes.json", mapRecord: buzzsproutEpisodeRecord},
	"podcasts": {accountLevel: true, resource: "podcasts.json", mapRecord: buzzsproutPodcastRecord},
}

// buzzsproutStreams returns the connector's published stream catalog.
func buzzsproutStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "episodes",
			Description:  "Buzzsprout episodes for the configured podcast.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"published_at"},
			Fields:       buzzsproutEpisodeFields(),
		},
		{
			Name:        "podcasts",
			Description: "Buzzsprout podcasts owned by the account.",
			PrimaryKey:  []string{"id"},
			Fields:      buzzsproutPodcastFields(),
		},
	}
}

func buzzsproutEpisodeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "audio_url", Type: "string"},
		{Name: "artwork_url", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "artist", Type: "string"},
		{Name: "tags", Type: "string"},
		{Name: "published_at", Type: "timestamp"},
		{Name: "duration", Type: "integer"},
		{Name: "hq", Type: "boolean"},
		{Name: "magic_mastering", Type: "boolean"},
		{Name: "guid", Type: "string"},
		{Name: "inactive_at", Type: "string"},
		{Name: "episode_number", Type: "integer"},
		{Name: "season_number", Type: "integer"},
		{Name: "explicit", Type: "boolean"},
		{Name: "private", Type: "boolean"},
		{Name: "total_plays", Type: "integer"},
	}
}

func buzzsproutPodcastFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "author", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "website_address", Type: "string"},
		{Name: "contact_email", Type: "string"},
		{Name: "keywords", Type: "string"},
		{Name: "explicit", Type: "boolean"},
		{Name: "main_category", Type: "string"},
		{Name: "sub_category", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "artwork_url", Type: "string"},
	}
}

func buzzsproutEpisodeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"title":           item["title"],
		"audio_url":       item["audio_url"],
		"artwork_url":     item["artwork_url"],
		"description":     item["description"],
		"summary":         item["summary"],
		"artist":          item["artist"],
		"tags":            item["tags"],
		"published_at":    item["published_at"],
		"duration":        item["duration"],
		"hq":              item["hq"],
		"magic_mastering": item["magic_mastering"],
		"guid":            item["guid"],
		"inactive_at":     item["inactive_at"],
		"episode_number":  item["episode_number"],
		"season_number":   item["season_number"],
		"explicit":        item["explicit"],
		"private":         item["private"],
		"total_plays":     item["total_plays"],
	}
}

func buzzsproutPodcastRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"title":           item["title"],
		"author":          item["author"],
		"description":     item["description"],
		"website_address": item["website_address"],
		"contact_email":   item["contact_email"],
		"keywords":        item["keywords"],
		"explicit":        item["explicit"],
		"main_category":   item["main_category"],
		"sub_category":    item["sub_category"],
		"language":        item["language"],
		"timezone":        item["timezone"],
		"artwork_url":     item["artwork_url"],
	}
}
