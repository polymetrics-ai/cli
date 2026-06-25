package appfollow

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the AppFollow API resource path (relative
// to base_url), the dotted JSON path its records live at, and the mapper that
// flattens each raw object into a connectors.Record.
//
// Several AppFollow endpoints are "per-app" reports addressed by an external app
// id (ext_id). Those streams set perExtID=true, and the read path fans out one
// request per configured ext_id, forwarding it as the ext_id query parameter.
type streamEndpoint struct {
	// resource is the path segment under /api/v2 (e.g. "account/apps").
	resource string
	// recordsPath is the dotted path to the records array/object in the body
	// ("" selects the root). For ratings the rows live under ratings.list and
	// are handled specially in the read loop, so recordsPath is unused there.
	recordsPath string
	// perExtID indicates the endpoint must be called once per configured ext_id.
	perExtID bool
	// mapRecord flattens a raw AppFollow object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// appfollowStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in appfollowStreams.
var appfollowStreamEndpoints = map[string]streamEndpoint{
	"users":           {resource: "account/users", recordsPath: "", mapRecord: userRecord},
	"app_collections": {resource: "account/apps", recordsPath: "apps", mapRecord: appCollectionRecord},
	"app_lists":       {resource: "account/apps/app", recordsPath: "apps_app", mapRecord: appListRecord},
	"ratings":         {resource: "meta/ratings", perExtID: true, mapRecord: ratingRecord},
}

// appfollowStreams returns the connector's published stream catalog.
func appfollowStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "users",
			Description: "Users on the AppFollow account.",
			PrimaryKey:  []string{"id"},
			Fields:      userFields(),
		},
		{
			Name:        "app_collections",
			Description: "App collections configured in the AppFollow account.",
			PrimaryKey:  []string{"id"},
			Fields:      appCollectionFields(),
		},
		{
			Name:        "app_lists",
			Description: "Apps belonging to each collection (one request per collection).",
			PrimaryKey:  []string{"app_id"},
			Fields:      appListFields(),
		},
		{
			Name:        "ratings",
			Description: "Per-day rating breakdown for each tracked app (one request per ext_id).",
			PrimaryKey:  []string{"ext_id", "date", "country"},
			Fields:      ratingFields(),
		},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func appCollectionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "title_normalized", Type: "string"},
		{Name: "count_apps", Type: "integer"},
		{Name: "countries", Type: "string"},
		{Name: "languages", Type: "string"},
		{Name: "created", Type: "string"},
	}
}

func appListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "app_id", Type: "integer"},
		{Name: "app_collection_id", Type: "integer"},
		{Name: "ext_id", Type: "string"},
		{Name: "store", Type: "string"},
		{Name: "count_reviews", Type: "integer"},
		{Name: "count_whatsnew", Type: "integer"},
		{Name: "is_favorite", Type: "integer"},
		{Name: "watch_url", Type: "string"},
		{Name: "created", Type: "string"},
	}
}

func ratingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ext_id", Type: "string"},
		{Name: "store", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "rating", Type: "number"},
		{Name: "stars1", Type: "integer"},
		{Name: "stars2", Type: "integer"},
		{Name: "stars3", Type: "integer"},
		{Name: "stars4", Type: "integer"},
		{Name: "stars5", Type: "integer"},
		{Name: "stars_total", Type: "integer"},
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"email":   item["email"],
		"name":    item["name"],
		"role":    item["role"],
		"status":  item["status"],
		"updated": item["updated"],
	}
}

func appCollectionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"title":            item["title"],
		"title_normalized": item["title_normalized"],
		"count_apps":       item["count_apps"],
		"countries":        item["countries"],
		"languages":        item["languages"],
		"created":          item["created"],
	}
}

func appListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"app_id":            item["app_id"],
		"app_collection_id": item["app_collection_id"],
		"ext_id":            item["ext_id"],
		"store":             item["store"],
		"count_reviews":     item["count_reviews"],
		"count_whatsnew":    item["count_whatsnew"],
		"is_favorite":       item["is_favorite"],
		"watch_url":         item["watch_url"],
		"created":           item["created"],
	}
}

// ratingRecord flattens one row of a ratings.list entry. The enclosing ext_id
// and store are injected by the read loop before this mapper runs.
func ratingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ext_id":      item["ext_id"],
		"store":       item["store"],
		"date":        item["date"],
		"country":     item["country"],
		"version":     item["version"],
		"rating":      item["rating"],
		"stars1":      item["stars1"],
		"stars2":      item["stars2"],
		"stars3":      item["stars3"],
		"stars4":      item["stars4"],
		"stars5":      item["stars5"],
		"stars_total": item["stars_total"],
	}
}
